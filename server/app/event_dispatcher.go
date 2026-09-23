package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/log"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/service"
)

type Event = config.Event

const defaultEventDispatchTimeout = 10 * time.Second

const (
	defaultEventWorkerCount = 4
	defaultEventQueueSize   = 256
)

var defaultWebhookUserAgent = appinfo.Name + "/" + appinfo.Version

// eventJob 是投递到异步执行池的事件任务，data 已在此前序列化为快照。
type eventJob struct {
	event Event
	data  json.RawMessage
}

// eventPayload 是分发给 webhook / command 的事件载荷信封。
type eventPayload struct {
	Name string `json:"name"`
	On   string `json:"on"`
	Data any    `json:"data"`
}

// articleDeletePayload 是文章删除事件的载荷。
type articleDeletePayload struct {
	ArticleID string `json:"articleID"`
	IsHard    bool   `json:"isHard"`
}

// loginFailedPayload 是登录失败事件的载荷。
type loginFailedPayload struct {
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

type eventDispatcher struct {
	appHooks     *appHooks
	articleHooks *service.ArticleHooks
	authHooks    *service.AuthHooks

	// mu 保护 events 与 timeout，使 ReloadConfig 重建时与 dispatch 并发安全。
	mu      sync.RWMutex
	events  map[string][]Event
	timeout time.Duration

	// jobs 为有界异步执行队列；wg 跟踪在途异步任务，供关闭时排空。
	// closeMu 保护 closed 标志与向 jobs 发送的过程，避免向已关闭通道发送。
	jobs    chan eventJob
	wg      sync.WaitGroup
	closeMu sync.Mutex
	closed  bool

	// workers / queueSize 记录启动时的执行池规模，供配置重载时比对提示。
	workers   int
	queueSize int
}

// shouldRunAsync 判断事件是否异步执行。
// 生命周期事件（app.*）强制同步；其余事件默认异步，可通过 async: false 强制同步。
func (e *eventDispatcher) shouldRunAsync(event Event) bool {
	if strings.HasPrefix(event.On, "app.") {
		return false
	}
	if event.Async == nil {
		return true
	}
	return *event.Async
}

func (e *eventDispatcher) dispatch(eventOn string, data any) {
	e.mu.RLock()
	events := e.events[eventOn]
	e.mu.RUnlock()

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		if !e.shouldRunAsync(event) {
			e.dispatchOne(event, data)
			continue
		}

		// 异步执行前先冻结载荷：请求返回后原始对象可能被继续修改，
		// 直接传指针会导致异步任务读到脏数据甚至数据竞争。
		raw, err := json.Marshal(data)
		if err != nil {
			log.With("type", "trigger_event").With("on", eventOn).
				With("event_name", event.Name).
				Errorf("marshal event payload for async dispatch: %v", err)
			continue
		}

		e.enqueue(eventJob{event: event, data: raw})
	}
}

// dispatchOne 执行单个事件（webhook / command）并记录结果。
func (e *eventDispatcher) dispatchOne(event Event, data any) {
	logTpl := log.With("type", "trigger_event").With("on", event.On)

	var err error = nil

	switch event.Type {
	case "webhook":
		err = e.dispatchWebhook(event, data)
	case "command":
		err = e.dispatchCommand(event, data)
	default:
		logTpl.Errorf("unknown event type: %s", event.Type)
		return
	}

	l := logTpl.With("type", event.Type).
		With("event_name", event.Name)

	if err != nil {
		l.Errorf("failed to dispatch event: %v", err)
		return
	}

	l.Info("event dispatched successfully")
}

// enqueue 将任务投递到异步执行池；队列满或分发器已关闭时丢弃并告警。
// 发送为非阻塞，且全程持锁，保证不会向已关闭的通道发送。
func (e *eventDispatcher) enqueue(job eventJob) {
	e.closeMu.Lock()
	defer e.closeMu.Unlock()

	l := log.With("type", "trigger_event").With("on", job.event.On).
		With("event_name", job.event.Name)

	if e.closed {
		l.Warnf("event dispatcher is closed, dropping event")
		return
	}

	e.wg.Add(1)

	select {
	case e.jobs <- job:
	default:
		e.wg.Done()
		l.Warnf("event queue is full, dropping event")
	}
}

// Close 关闭异步队列并等待在途事件完成，ctx 控制等待上限。
// 可重复调用。
func (e *eventDispatcher) Close(ctx context.Context) error {
	e.closeMu.Lock()
	if !e.closed {
		e.closed = true
		close(e.jobs)
	}
	e.closeMu.Unlock()

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *eventDispatcher) startWorkers(workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for job := range e.jobs {
				e.dispatchOne(job.event, job.data)
				e.wg.Done()
			}
		}()
	}
}

func (e *eventDispatcher) dispatchWebhook(event Event, data any) error {
	if event.Entry == "" {
		return fmt.Errorf("webhook entry is empty")
	}

	body, err := json.Marshal(eventPayload{
		Name: event.Name,
		On:   event.On,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, event.Entry, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", defaultWebhookUserAgent)

	for name, value := range event.Headers {
		req.Header.Set(name, value)

		// Host 无法通过 Header 生效，需同时设置 req.Host
		if strings.EqualFold(name, "Host") {
			req.Host = value
		}
	}

	client := &http.Client{Timeout: e.effectiveTimeout(event)}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (e *eventDispatcher) dispatchCommand(event Event, data any) error {
	if event.Entry == "" {
		return fmt.Errorf("command entry is empty")
	}

	body, err := json.Marshal(eventPayload{
		Name: event.Name,
		On:   event.On,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("marshal command payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.effectiveTimeout(event))
	defer cancel()

	name, args := shellInvocation(runtime.GOOS, event.Entry)
	cmd := exec.CommandContext(ctx, name, args...)

	cmd.Env = append(os.Environ(),
		"EVENT_NAME="+event.Name,
		"EVENT_ON="+event.On,
	)

	cmd.Stdin = bytes.NewReader(body)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command: %w: %s", err, string(output))
	}

	return nil
}

// shellInvocation 返回在指定平台上执行 entry 的 shell 命令与参数。
func shellInvocation(goos, entry string) (string, []string) {
	if goos == "windows" {
		return "cmd", []string{"/c", entry}
	}
	return "sh", []string{"-c", entry}
}

func (e *eventDispatcher) InjectAppHook(appHooks *appHooks) {
	e.appHooks = appHooks

	// 因为事件配置是配置文件中定义的，所以说 appHooks.OnAfterConfig 不可用
	// appHooks.OnAfterConfig = append(appHooks.OnAfterConfig, func(conf *config.Config) error {
	// 	e.dispatch("app.after_config", conf)
	// 	return nil
	// })

	appHooks.OnAfterDB = append(appHooks.OnAfterDB, e.trigger("app.after_db"))
	appHooks.OnAfterCache = append(appHooks.OnAfterCache, e.trigger("app.after_cache"))
	appHooks.OnAfterStorage = append(appHooks.OnAfterStorage, e.trigger("app.after_storage"))
	appHooks.OnMigration = append(appHooks.OnMigration, e.trigger("app.migration"))
	appHooks.OnBeforeServe = append(appHooks.OnBeforeServe, e.trigger("app.before_serve"))
	appHooks.OnShutdown = append(appHooks.OnShutdown, e.trigger("app.shutdown"))

	// 因为事件配置是配置文件中定义的，所以说 appHooks.OnBeforeRouter 不可用
	// appHooks.OnBeforeRouter = append(appHooks.OnBeforeRouter, func(r chi.Router) error {
	// 	e.dispatch("app.before_router", nil)
	// 	return nil
	// })

	// appHooks.OnAfterRouter = append(appHooks.OnAfterRouter, func(r chi.Router) error {
	// 	e.dispatch("app.after_router", nil)
	// 	return nil
	// })
}

func (e *eventDispatcher) InjectArticleHook(articleHooks *service.ArticleHooks) {
	e.articleHooks = articleHooks

	articleHooks.BeforeCreate = append(articleHooks.BeforeCreate, e.triggerArticleFields("article.creating"))
	articleHooks.AfterCreate = append(articleHooks.AfterCreate, e.triggerArticle("article.created"))
	articleHooks.BeforePublish = append(articleHooks.BeforePublish, e.triggerArticle("article.publishing"))
	articleHooks.AfterPublish = append(articleHooks.AfterPublish, e.triggerArticle("article.published"))
	articleHooks.BeforeDelete = append(articleHooks.BeforeDelete, e.triggerArticleDelete("article.deleting"))
	articleHooks.AfterDelete = append(articleHooks.AfterDelete, e.triggerArticleDelete("article.deleted"))
	articleHooks.BeforeUpdateInfo = append(articleHooks.BeforeUpdateInfo, e.triggerArticle("article.updating"))
	articleHooks.AfterUpdateInfo = append(articleHooks.AfterUpdateInfo, e.triggerArticle("article.updated"))
}

func (e *eventDispatcher) InjectAuthHook(authHooks *service.AuthHooks) {
	e.authHooks = authHooks

	authHooks.OnLoginSuccess = append(authHooks.OnLoginSuccess, e.triggerAdmin("auth.login_success"))
	authHooks.OnLoginFailed = append(authHooks.OnLoginFailed, e.triggerLoginFailed("auth.login_failed"))
	authHooks.OnLogout = append(authHooks.OnLogout, e.triggerClaims("auth.logout"))
}

// 以下 trigger* 方法把事件名适配为各 Hook 的函数签名，
// 使 Inject* 中只需一行即可注册一个事件回调。
func (e *eventDispatcher) trigger(on string) func() error {
	return func() error {
		e.dispatch(on, nil)
		return nil
	}
}

func (e *eventDispatcher) triggerArticle(on string) func(context.Context, *data.Article) error {
	return func(_ context.Context, article *data.Article) error {
		e.dispatch(on, article)
		return nil
	}
}

func (e *eventDispatcher) triggerArticleFields(on string) func(context.Context, *service.ArticleUpdateFields) error {
	return func(_ context.Context, fields *service.ArticleUpdateFields) error {
		e.dispatch(on, fields)
		return nil
	}
}

func (e *eventDispatcher) triggerArticleDelete(on string) func(context.Context, string, bool) error {
	return func(_ context.Context, articleID string, isHard bool) error {
		e.dispatch(on, articleDeletePayload{ArticleID: articleID, IsHard: isHard})
		return nil
	}
}

func (e *eventDispatcher) triggerAdmin(on string) func(context.Context, *data.Admin) error {
	return func(_ context.Context, admin *data.Admin) error {
		e.dispatch(on, admin)
		return nil
	}
}

func (e *eventDispatcher) triggerLoginFailed(on string) func(context.Context, string, string) error {
	return func(_ context.Context, username, reason string) error {
		e.dispatch(on, loginFailedPayload{Username: username, Reason: reason})
		return nil
	}
}

func (e *eventDispatcher) triggerClaims(on string) func(context.Context, *authtoken.TokenClaims) error {
	return func(_ context.Context, claims *authtoken.TokenClaims) error {
		e.dispatch(on, claims)
		return nil
	}
}

// MakeEventsMap 按事件触发时机（On）重建事件索引，可用于初始化与配置热重载。
func (e *eventDispatcher) MakeEventsMap(events []Event) {
	eventsMap := make(map[string][]Event)

	for _, event := range events {
		eventsMap[event.On] = append(eventsMap[event.On], event)
	}

	e.mu.Lock()
	e.events = eventsMap
	e.mu.Unlock()
}

// setTimeout 更新分发超时；非正值回退到默认超时。
func (e *eventDispatcher) setTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = defaultEventDispatchTimeout
	}

	e.mu.Lock()
	e.timeout = timeout
	e.mu.Unlock()
}

// effectiveTimeout 返回事件的生效超时：事件级 timeout 优先，否则使用分发器超时。
func (e *eventDispatcher) effectiveTimeout(event Event) time.Duration {
	if event.Timeout > 0 {
		return event.Timeout.Duration()
	}
	return e.timeoutOr()
}

// poolConfig 返回启动时的执行池规模，供配置重载时比对提示。
func (e *eventDispatcher) poolConfig() (workers, queueSize int) {
	return e.workers, e.queueSize
}

// timeoutOr 返回当前分发超时，未设置时返回默认超时。
func (e *eventDispatcher) timeoutOr() time.Duration {
	e.mu.RLock()
	timeout := e.timeout
	e.mu.RUnlock()

	if timeout <= 0 {
		return defaultEventDispatchTimeout
	}
	return timeout
}

func newEventDispatcher(events []Event, timeout time.Duration, workers, queueSize int) *eventDispatcher {
	if workers <= 0 {
		workers = defaultEventWorkerCount
	}
	if queueSize <= 0 {
		queueSize = defaultEventQueueSize
	}

	e := &eventDispatcher{
		jobs:      make(chan eventJob, queueSize),
		workers:   workers,
		queueSize: queueSize,
	}
	e.setTimeout(timeout)
	e.MakeEventsMap(events)
	e.startWorkers(workers)
	return e
}

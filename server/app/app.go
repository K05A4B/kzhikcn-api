package app

import (
	"context"
	"fmt"
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/log"
	"kzhikcn/server/service"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

// AppState 表示 App 当前所处的生命周期阶段
type AppState int

const (
	StateNew          AppState = iota // 未启动
	StateBootstrapped                 // 配置加载完成 (Bootstrap)
	StateInitialized                  // 基础设施就绪 (Initialize)
	StateMigrated                     // 数据库迁移完成 (Migrate)
	StateReady                        // 路由构建完成 (Ready)
	StateServing                      // 正在监听端口 (Serve)
	StateStopped                      // 已停止 (Shutdown)
)

// appHooks 聚合了所有生命周期阶段的 Hook 回调
type appHooks struct {
	OnAfterConfig  []func(*config.Config) error
	OnAfterDB      []func() error
	OnAfterCache   []func() error
	OnAfterStorage []func() error
	OnMigration    []func() error
	OnBeforeRouter []func(chi.Router) error
	OnAfterRouter  []func(chi.Router) error
	OnBeforeServe  []func() error
	OnShutdown     []func() error
}

// App 是服务启动生命周期的编排者。
// 每个实例沿状态机单向推进：New → Bootstrapped → Initialized → Migrated → Ready → Serving → Stopped。
type App struct {
	State   AppState
	Router  chi.Router
	Context *AppContext
	Config  *config.Config

	hooks appHooks
	srv   *http.Server

	eventDispatcher *eventDispatcher

	confMutex sync.RWMutex

	configFile string
}

// New 创建一个处于 StateNew 状态的 App 实例。
func New() *App {
	return &App{
		State: StateNew,
	}
}

func (a *App) loadConfig() error {
	a.confMutex.Lock()
	defer a.confMutex.Unlock()

	conf, err := config.LoadConfigFromFile(a.configFile)

	if err != nil {
		return err
	}

	if a.Config == nil {
		a.Config = conf
	} else {
		(*a.Config) = (*conf)
	}

	return nil
}

func (a *App) ReloadConfig() error {
	if err := a.loadConfig(); err != nil {
		return err
	}

	if a.eventDispatcher != nil {
		a.eventDispatcher.setTimeout(a.Config.EventDispatcher.Timeout.Duration())
		a.eventDispatcher.MakeEventsMap(a.Config.Events)

		// 执行池规模仅启动时生效，配置变更需重启服务
		workers, queueSize := a.eventDispatcher.poolConfig()
		if cfg := a.Config.EventDispatcher.Workers; cfg > 0 && cfg != workers {
			log.Warnf("event_dispatcher.workers 变更需重启服务后生效: %d -> %d", workers, cfg)
		}
		if cfg := a.Config.EventDispatcher.QueueSize; cfg > 0 && cfg != queueSize {
			log.Warnf("event_dispatcher.queue_size 变更需重启服务后生效: %d -> %d", queueSize, cfg)
		}
	}

	return log.Configure(logOptions(a.Config, true))
}

// logOptions 将日志配置翻译为 pkg/log 的构建参数。
func logOptions(conf *config.Config, consoleOutput bool) log.Options {
	opts := log.Options{
		Level:   conf.Log.LogLevel,
		Console: consoleOutput,
	}

	if conf.Log.Enable && conf.Log.Lumberjack != nil {
		opts.Writers = append(opts.Writers, conf.Log.Lumberjack)
	}

	return opts
}

func (a *App) GetConfig() config.Config {
	a.confMutex.RLock()
	conf := *a.Config
	defer a.confMutex.RUnlock()
	return conf
}

// Bootstrap 加载配置文件，完成解析并触发 OnAfterConfig 回调。
// configFile: YAML 配置文件的路径（必须存在）。
func (a *App) Bootstrap(configFile string, consoleOutput bool) error {
	if a.State != StateNew {
		return fmt.Errorf("cannot bootstrap from state %d", a.State)
	}

	a.configFile = configFile
	err := a.loadConfig()
	if err != nil {
		return err
	}

	// 日志必须在其余基础设施（数据库/缓存）开始输出之前完成配置。
	if err := log.Configure(logOptions(a.Config, consoleOutput)); err != nil {
		return err
	}

	// 事件分发器依赖已加载的配置，且必须在 OnAfterConfig 之前构建。
	a.eventDispatcher = newEventDispatcher(
		a.Config.Events,
		a.Config.EventDispatcher.Timeout.Duration(),
		a.Config.EventDispatcher.Workers,
		a.Config.EventDispatcher.QueueSize,
	)
	a.eventDispatcher.InjectAppHook(&a.hooks)

	for _, fn := range a.hooks.OnAfterConfig {
		if err := fn(a.Config); err != nil {
			return err
		}
	}

	a.State = StateBootstrapped
	return nil
}

// BootstrapOrCreate 同 Bootstrap，但如果 configFile 不存在则自动写入默认配置后重试。
func (a *App) BootstrapOrCreate(configFile string, consoleOutput bool) error {
	_, err := os.Stat(configFile)
	if err == nil {
		return a.Bootstrap(configFile, consoleOutput)
	}
	if os.IsNotExist(err) {
		if err := assets.ExportDefaultConfig(configFile); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		return a.Bootstrap(configFile, consoleOutput)
	}
	return fmt.Errorf("failed to stat config file: %w", err)
}

// Initialize 初始化全部基础设施：数据库连接、缓存、存储后端。
// 前置：StateBootstrapped。
func (a *App) Initialize() error {
	return a.initialize(true)
}

// InitializeWithoutCache 与 Initialize 相同，但跳过缓存初始化。
// 供只访问数据库的 CLI 命令使用，避免与运行中的服务争抢缓存目录锁。
// 前置：StateBootstrapped。
func (a *App) InitializeWithoutCache() error {
	return a.initialize(false)
}

func (a *App) initialize(withCache bool) error {
	if a.State != StateBootstrapped {
		return fmt.Errorf("cannot initialize from state %d", a.State)
	}

	conf := a.Config

	// 1. 数据库连接
	if err := data.ConnectDatabase(conf.Database.Driver, conf.Database.Dsn); err != nil {
		return err
	}
	for _, fn := range a.hooks.OnAfterDB {
		if err := fn(); err != nil {
			return err
		}
	}

	// 2. 缓存
	if withCache {
		if err := cache.InitCache(conf); err != nil {
			return err
		}
		for _, fn := range a.hooks.OnAfterCache {
			if err := fn(); err != nil {
				return err
			}
		}
	}

	// 3. 存储
	articlesRepo := assets.Init(conf)
	for _, fn := range a.hooks.OnAfterStorage {
		if err := fn(); err != nil {
			return err
		}
	}

	// 4. 初始化上下文
	a.Context = &AppContext{
		Context: context.Background(),
		app:     a,
	}

	articleHooks := &service.ArticleHooks{}
	authHooks := &service.AuthHooks{}

	// 注入article.*事件钩子
	a.eventDispatcher.InjectArticleHook(articleHooks)
	// 注入auth.*事件钩子
	a.eventDispatcher.InjectAuthHook(authHooks)

	serviceCtx := service.NewServiceContext(a.Config, articlesRepo, data.DB())

	a.Context.ArticleSvc = service.NewArticleService(serviceCtx, articleHooks)
	a.Context.AuthSvc = service.NewAuthService(serviceCtx, authHooks)
	a.Context.AdminSvc = service.NewAdminService(serviceCtx)
	a.Context.TopicSvc = service.NewTopicService(serviceCtx)

	a.State = StateInitialized
	return nil
}

// Migrate 执行数据库迁移（建表 / 增量迁移）并创建默认管理员（首次启动时）。
// 前置：StateInitialized。
func (a *App) Migrate() error {
	if a.State != StateInitialized {
		return fmt.Errorf("cannot migrate from state %d", a.State)
	}

	if !data.ExistSchemaState() {
		// 全新数据库：先建 SchemaState 表，再做完整迁移 + 默认管理员
		if err := data.DB().AutoMigrate(&data.SchemaState{}); err != nil {
			return err
		}
		if err := data.InitDatabase(); err != nil {
			return err
		}
	} else {
		// 已有数据：增量迁移
		if err := data.AutoMigrates(); err != nil {
			return err
		}
	}

	for _, fn := range a.hooks.OnMigration {
		if err := fn(); err != nil {
			return err
		}
	}

	a.State = StateMigrated
	return nil
}

// Ready 构建 HTTP 路由树。在路由构建前后分别触发 OnBeforeRouter / OnAfterRouter，
// 插件可借此注入自己的路由和中间件。
// 前置：StateMigrated。
func (a *App) Ready(router chi.Router) error {
	if a.State != StateMigrated {
		return fmt.Errorf("cannot ready from state %d", a.State)
	}

	r := router

	for _, fn := range a.hooks.OnBeforeRouter {
		if err := fn(r); err != nil {
			return err
		}
	}

	for _, fn := range a.hooks.OnAfterRouter {
		if err := fn(r); err != nil {
			return err
		}
	}

	a.Router = r
	a.State = StateReady
	return nil
}

// Serve 启动 HTTP 服务器（根据配置自动选择 HTTPS 或 HTTP），阻塞直到出错或 Shutdown。
// 前置：StateReady。
func (a *App) Serve(addr string) error {
	if a.State != StateReady {
		return fmt.Errorf("cannot serve from state %d", a.State)
	}

	for _, fn := range a.hooks.OnBeforeServe {
		if err := fn(); err != nil {
			return err
		}
	}

	conf := a.Config
	cert := conf.CertFile
	key := conf.KeyFile

	a.srv = &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	fmt.Println("Version:", appinfo.CurrentInfo.Version)
	fmt.Println("Author:", appinfo.CurrentInfo.Author)
	fmt.Println("Copyright:", appinfo.CurrentInfo.Copyright)

	a.State = StateServing

	if cert != "" && key != "" {
		log.Info("Starting HTTPS server on ", addr)
		return a.srv.ListenAndServeTLS(cert, key)
	}

	log.Info("Starting non-HTTPS server on ", addr)
	return a.srv.ListenAndServe()
}

// Shutdown 关闭服务：停止 HTTP → 关闭缓存 → 断开数据库 → 触发 OnShutdown 回调。
// 从其他 goroutine 调用会触发 Serve() 返回 http.ErrServerClosed。
func (a *App) Shutdown(ctx context.Context) error {
	if a.State != StateServing {
		return fmt.Errorf("cannot shutdown from state %d", a.State)
	}

	var errs []error

	// 1. 优雅停止 HTTP：不处理新请求，等待进行中的请求完成
	if a.srv != nil {
		if err := a.srv.Shutdown(ctx); err != nil {
			errs = append(errs, errors.Wrap(err, "http shutdown"))
		}
	}

	// 2. 排空在途的异步事件：HTTP 已停止，不会再产生新的请求事件
	if a.eventDispatcher != nil {
		if err := a.eventDispatcher.Close(ctx); err != nil {
			errs = append(errs, errors.Wrap(err, "event dispatch drain"))
		}
	}

	// 3. 关闭缓存 (Badger/Redis)
	if err := cache.CloseCache(); err != nil {
		errs = append(errs, errors.Wrap(err, "cache close"))
	}

	// 4. 断开数据库连接
	if err := data.CloseDB(); err != nil {
		errs = append(errs, errors.Wrap(err, "db close"))
	}

	// 5. 用户注册的清理 Hook
	for _, fn := range a.hooks.OnShutdown {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}

	a.State = StateStopped

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// Hook 注册方法

func (a *App) HookAfterConfig(fn func(*config.Config) error) {
	a.hooks.OnAfterConfig = append(a.hooks.OnAfterConfig, fn)
}

func (a *App) HookAfterDB(fn func() error) {
	a.hooks.OnAfterDB = append(a.hooks.OnAfterDB, fn)
}

func (a *App) HookAfterCache(fn func() error) {
	a.hooks.OnAfterCache = append(a.hooks.OnAfterCache, fn)
}

func (a *App) HookAfterStorage(fn func() error) {
	a.hooks.OnAfterStorage = append(a.hooks.OnAfterStorage, fn)
}

func (a *App) HookMigration(fn func() error) {
	a.hooks.OnMigration = append(a.hooks.OnMigration, fn)
}

func (a *App) HookBeforeRouter(fn func(chi.Router) error) {
	a.hooks.OnBeforeRouter = append(a.hooks.OnBeforeRouter, fn)
}

func (a *App) HookAfterRouter(fn func(chi.Router) error) {
	a.hooks.OnAfterRouter = append(a.hooks.OnAfterRouter, fn)
}

func (a *App) HookBeforeServe(fn func() error) {
	a.hooks.OnBeforeServe = append(a.hooks.OnBeforeServe, fn)
}

func (a *App) HookShutdown(fn func() error) {
	a.hooks.OnShutdown = append(a.hooks.OnShutdown, fn)
}

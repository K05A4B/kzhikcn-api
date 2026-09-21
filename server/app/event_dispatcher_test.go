package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"kzhikcn/pkg/data"
	"kzhikcn/server/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rawEnvelope 用于反序列化 webhook / command 收到的事件载荷。
type rawEnvelope struct {
	Name string          `json:"name"`
	On   string          `json:"on"`
	Data json.RawMessage `json:"data"`
}

type capturedRequest struct {
	method      string
	contentType string
	env         rawEnvelope
	raw         []byte
}

// newCaptureServer 返回记录所有收到请求的 httptest.Server。
func newCaptureServer(t *testing.T, status int) (*httptest.Server, *[]capturedRequest) {
	t.Helper()

	got := &[]capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)

		req := capturedRequest{
			method:      r.Method,
			contentType: r.Header.Get("Content-Type"),
			raw:         raw,
		}
		_ = json.Unmarshal(raw, &req.env)

		*got = append(*got, req)
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)

	return srv, got
}

func TestMakeEventsMap_AssignsAndIndexes(t *testing.T) {
	d := &eventDispatcher{}
	d.MakeEventsMap([]Event{
		{Name: "a", On: "article.created", Type: "webhook", Entry: "http://example.com/1"},
		{Name: "b", On: "article.created", Type: "command", Entry: "echo 1"},
		{Name: "c", On: "article.updated", Type: "webhook", Entry: "http://example.com/2"},
	})

	require.Len(t, d.events, 2)
	require.Len(t, d.events["article.created"], 2)
	assert.Equal(t, "a", d.events["article.created"][0].Name)
	assert.Equal(t, "b", d.events["article.created"][1].Name)
	assert.Len(t, d.events["article.updated"], 1)
}

func TestMakeEventsMap_Reload(t *testing.T) {
	d := newEventDispatcher([]Event{{Name: "old", On: "old.event"}}, time.Second)

	d.MakeEventsMap([]Event{{Name: "new", On: "new.event"}})

	assert.Empty(t, d.events["old.event"])
	require.Len(t, d.events["new.event"], 1)
	assert.Equal(t, "new", d.events["new.event"][0].Name)
}

func TestNewEventDispatcher_Timeout(t *testing.T) {
	fallback := newEventDispatcher(nil, 0)
	assert.Equal(t, defaultEventDispatchTimeout, fallback.timeoutOr())

	custom := newEventDispatcher(nil, 50*time.Millisecond)
	assert.Equal(t, 50*time.Millisecond, custom.timeoutOr())
}

func TestDispatch_NoEvents(t *testing.T) {
	d := newEventDispatcher(nil, time.Second)

	assert.NotPanics(t, func() {
		d.dispatch("not.configured", "data")
	})
}

func TestDispatch_UnknownTypeContinues(t *testing.T) {
	srv, got := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher([]Event{
		{Name: "bad", On: "e", Type: "email", Entry: "x"},
		{Name: "ok", On: "e", Type: "webhook", Entry: srv.URL},
	}, time.Second)

	d.dispatch("e", map[string]any{"k": "v"})

	require.Len(t, *got, 1)
	assert.Equal(t, "ok", (*got)[0].env.Name)
}

func TestDispatch_MultipleSameOnInOrder(t *testing.T) {
	srv, got := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher([]Event{
		{Name: "first", On: "e", Type: "webhook", Entry: srv.URL},
		{Name: "second", On: "e", Type: "webhook", Entry: srv.URL},
	}, time.Second)

	d.dispatch("e", nil)

	require.Len(t, *got, 2)
	assert.Equal(t, "first", (*got)[0].env.Name)
	assert.Equal(t, "second", (*got)[1].env.Name)
}

func TestDispatch_BestEffortOnFailure(t *testing.T) {
	failing, failed := newCaptureServer(t, http.StatusInternalServerError)
	ok, succeeded := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher([]Event{
		{Name: "bad", On: "e", Type: "webhook", Entry: failing.URL},
		{Name: "ok", On: "e", Type: "webhook", Entry: ok.URL},
	}, time.Second)

	d.dispatch("e", nil)

	require.Len(t, *failed, 1)
	require.Len(t, *succeeded, 1)
}

func TestDispatchWebhook_Success(t *testing.T) {
	srv, got := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher(nil, time.Second)
	err := d.dispatchWebhook(Event{Name: "hook", On: "article.created", Type: "webhook", Entry: srv.URL}, map[string]any{"k": "v"})
	require.NoError(t, err)

	require.Len(t, *got, 1)
	assert.Equal(t, http.MethodPost, (*got)[0].method)
	assert.Equal(t, "application/json", (*got)[0].contentType)
	assert.Equal(t, "hook", (*got)[0].env.Name)
	assert.Equal(t, "article.created", (*got)[0].env.On)
	assert.JSONEq(t, `{"k":"v"}`, string((*got)[0].env.Data))
}

func TestDispatchWebhook_EmptyEntry(t *testing.T) {
	d := newEventDispatcher(nil, time.Second)

	err := d.dispatchWebhook(Event{Name: "hook", On: "e", Type: "webhook", Entry: ""}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestDispatchWebhook_InvalidURL(t *testing.T) {
	d := newEventDispatcher(nil, time.Second)

	err := d.dispatchWebhook(Event{Name: "hook", On: "e", Type: "webhook", Entry: "://bad"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "build webhook request")
}

func TestDispatchWebhook_MarshalError(t *testing.T) {
	srv, _ := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher(nil, time.Second)
	err := d.dispatchWebhook(Event{Name: "hook", On: "e", Type: "webhook", Entry: srv.URL}, make(chan int))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal")
}

func TestDispatchWebhook_Non2xx(t *testing.T) {
	srv, _ := newCaptureServer(t, http.StatusInternalServerError)

	d := newEventDispatcher(nil, time.Second)
	err := d.dispatchWebhook(Event{Name: "hook", On: "e", Type: "webhook", Entry: srv.URL}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestDispatchWebhook_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := newEventDispatcher(nil, 50*time.Millisecond)
	err := d.dispatchWebhook(Event{Name: "hook", On: "e", Type: "webhook", Entry: srv.URL}, nil)
	require.Error(t, err)
}

func TestShellInvocation(t *testing.T) {
	cases := []struct {
		goos      string
		entry     string
		wantName  string
		wantFirst string
	}{
		{goos: "windows", entry: "echo hi", wantName: "cmd", wantFirst: "/c"},
		{goos: "linux", entry: "echo hi", wantName: "sh", wantFirst: "-c"},
		{goos: "darwin", entry: "echo hi", wantName: "sh", wantFirst: "-c"},
	}

	for _, tc := range cases {
		t.Run(tc.goos, func(t *testing.T) {
			name, args := shellInvocation(tc.goos, tc.entry)
			assert.Equal(t, tc.wantName, name)
			require.Len(t, args, 2)
			assert.Equal(t, tc.wantFirst, args[0])
			assert.Equal(t, tc.entry, args[1])
		})
	}
}

func TestDispatchCommand_ReceivesStdinAndEnv(t *testing.T) {
	dir := t.TempDir()
	outJSON := filepath.Join(dir, "stdin.json")
	envFile := filepath.Join(dir, "env.txt")

	// Windows 下 entry 位于 cmd /c 之后，且 test 进程会转义其中的双引号，
	// 因此这里只使用 shell 内置命令并避免在 entry 中出现引号。
	var entry string
	if runtime.GOOS == "windows" {
		if strings.ContainsAny(dir, " ") {
			t.Skip("临时目录含空格，Windows 下 shell 引号会被 cmd 误解析")
		}
		entry = "(echo %EVENT_NAME%& echo %EVENT_ON%) > " + envFile + " & more > " + outJSON
	} else {
		entry = `printf '%s\n%s' "$EVENT_NAME" "$EVENT_ON" > ` + envFile + ` && cat > ` + outJSON
	}

	d := newEventDispatcher(nil, 10*time.Second)
	err := d.dispatchCommand(Event{Name: "My Event", On: "article.created", Type: "command", Entry: entry}, map[string]any{"id": "42"})
	require.NoError(t, err)

	stdinJSON, err := os.ReadFile(outJSON)
	require.NoError(t, err)

	var env rawEnvelope
	require.NoError(t, json.Unmarshal(stdinJSON, &env))
	assert.Equal(t, "My Event", env.Name)
	assert.Equal(t, "article.created", env.On)
	assert.JSONEq(t, `{"id":"42"}`, string(env.Data))

	envRaw, err := os.ReadFile(envFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(string(envRaw), "\r\n", "\n"), "\n"), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "My Event", lines[0])
	assert.Equal(t, "article.created", lines[1])
}

func TestDispatchCommand_EmptyEntry(t *testing.T) {
	d := newEventDispatcher(nil, time.Second)

	err := d.dispatchCommand(Event{Name: "cmd", On: "e", Type: "command", Entry: ""}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestDispatchCommand_NonZeroExit(t *testing.T) {
	d := newEventDispatcher(nil, 5*time.Second)

	err := d.dispatchCommand(Event{Name: "cmd", On: "e", Type: "command", Entry: "exit 1"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run command")
}

func TestDispatchCommand_Timeout(t *testing.T) {
	entry := "sleep 1"
	if runtime.GOOS == "windows" {
		entry = "ping -n 3 127.0.0.1 > nul"
	}

	d := newEventDispatcher(nil, 50*time.Millisecond)
	err := d.dispatchCommand(Event{Name: "cmd", On: "e", Type: "command", Entry: entry}, nil)
	require.Error(t, err)
}

func TestInjectAppHook_RegistersEnabledHooks(t *testing.T) {
	d := newEventDispatcher(nil, time.Second)
	hooks := &appHooks{}

	d.InjectAppHook(hooks)

	assert.Len(t, hooks.OnAfterDB, 1)
	assert.Len(t, hooks.OnAfterCache, 1)
	assert.Len(t, hooks.OnAfterStorage, 1)
	assert.Len(t, hooks.OnMigration, 1)
	assert.Len(t, hooks.OnBeforeServe, 1)
	assert.Len(t, hooks.OnShutdown, 1)

	// 按设计不启用
	assert.Empty(t, hooks.OnAfterConfig)
	assert.Empty(t, hooks.OnBeforeRouter)
	assert.Empty(t, hooks.OnAfterRouter)
}

func TestInjectArticleHook_RegistersAndDispatches(t *testing.T) {
	ctx := context.Background()
	srv, got := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher([]Event{
		{Name: "created", On: "article.created", Type: "webhook", Entry: srv.URL},
		{Name: "deleting", On: "article.deleting", Type: "webhook", Entry: srv.URL},
	}, time.Second)

	hooks := &service.ArticleHooks{}
	d.InjectArticleHook(hooks)

	assert.Len(t, hooks.BeforeCreate, 1)
	assert.Len(t, hooks.AfterCreate, 1)
	assert.Len(t, hooks.BeforePublish, 1)
	assert.Len(t, hooks.AfterPublish, 1)
	assert.Len(t, hooks.BeforeDelete, 1)
	assert.Len(t, hooks.AfterDelete, 1)
	assert.Len(t, hooks.BeforeUpdateInfo, 1)
	assert.Len(t, hooks.AfterUpdateInfo, 1)

	require.NoError(t, hooks.AfterCreate[0](ctx, &data.Article{Title: "hello"}))
	require.NoError(t, hooks.BeforeDelete[0](ctx, "article-1", true))

	require.Len(t, *got, 2)
	assert.Equal(t, "article.created", (*got)[0].env.On)
	assert.Equal(t, "article.deleting", (*got)[1].env.On)

	var payload articleDeletePayload
	require.NoError(t, json.Unmarshal((*got)[1].env.Data, &payload))
	assert.Equal(t, "article-1", payload.ArticleID)
	assert.True(t, payload.IsHard)
}

func TestInjectAuthHook_RegistersAndSanitizesAdmin(t *testing.T) {
	ctx := context.Background()
	srv, got := newCaptureServer(t, http.StatusOK)

	d := newEventDispatcher([]Event{
		{Name: "login", On: "auth.login_success", Type: "webhook", Entry: srv.URL},
	}, time.Second)

	hooks := &service.AuthHooks{}
	d.InjectAuthHook(hooks)

	assert.Len(t, hooks.OnLoginSuccess, 1)
	assert.Len(t, hooks.OnLoginFailed, 1)
	assert.Len(t, hooks.OnLogout, 1)

	admin := &data.Admin{
		ID:         1,
		Username:   "admin",
		Password:   []byte("$2a$hashedpassword"),
		TotpSecret: []byte("TOTPSECRET"),
	}
	require.NoError(t, hooks.OnLoginSuccess[0](ctx, admin))

	require.Len(t, *got, 1)
	assert.Equal(t, "auth.login_success", (*got)[0].env.On)
	assert.Contains(t, string((*got)[0].raw), `"username":"admin"`)
	assert.NotContains(t, string((*got)[0].raw), "hashedpassword")
	assert.NotContains(t, string((*got)[0].raw), "TOTPSECRET")
}

func TestEventDispatcher_ConcurrentReload(t *testing.T) {
	d := newEventDispatcher([]Event{{Name: "a", On: "e", Type: "email"}}, time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			d.dispatch("e", i)
		}(i)
	}

	for i := 0; i < 10; i++ {
		d.MakeEventsMap([]Event{{Name: "a", On: "e", Type: "email"}})
	}

	wg.Wait()
}

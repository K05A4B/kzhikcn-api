package server

import (
	"context"
	"fmt"
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/log"
	"net/http"
	"os"

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
	State  AppState
	Router chi.Router

	conf  *config.Config
	hooks appHooks
	srv   *http.Server
}

// New 创建一个处于 StateNew 状态的 App 实例。
func New() *App {
	return &App{
		State: StateNew,
	}
}

// Bootstrap 加载配置文件，完成解析并触发 OnAfterConfig 回调。
// configFile: YAML 配置文件的路径（必须存在）。
func (a *App) Bootstrap(configFile string) error {
	if a.State != StateNew {
		return fmt.Errorf("cannot bootstrap from state %d", a.State)
	}

	conf, err := config.LoadConfigFromFile(configFile)
	if err != nil {
		return err
	}
	a.conf = conf

	for _, fn := range a.hooks.OnAfterConfig {
		if err := fn(conf); err != nil {
			return err
		}
	}

	a.State = StateBootstrapped
	return nil
}

// BootstrapOrCreate 同 Bootstrap，但如果 configFile 不存在则自动写入默认配置后重试。
func (a *App) BootstrapOrCreate(configFile string) error {
	_, err := os.Stat(configFile)
	if err == nil {
		return a.Bootstrap(configFile)
	}
	if os.IsNotExist(err) {
		if err := assets.ExportDefaultConfig(configFile); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		return a.Bootstrap(configFile)
	}
	return fmt.Errorf("failed to stat config file: %w", err)
}

// Initialize 初始化基础设施：数据库连接、缓存、存储后端。
// 前置：StateBootstrapped。
func (a *App) Initialize() error {
	if a.State != StateBootstrapped {
		return fmt.Errorf("cannot initialize from state %d", a.State)
	}

	conf := a.conf

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
	if err := cache.InitCache(conf); err != nil {
		return err
	}
	for _, fn := range a.hooks.OnAfterCache {
		if err := fn(); err != nil {
			return err
		}
	}

	// 3. 存储
	assets.Init(conf)
	for _, fn := range a.hooks.OnAfterStorage {
		if err := fn(); err != nil {
			return err
		}
	}

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
		if _, err := data.InitDatabase(); err != nil {
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
func (a *App) Ready() error {
	if a.State != StateMigrated {
		return fmt.Errorf("cannot ready from state %d", a.State)
	}

	r := NewRouter()

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

	conf := a.conf
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

// Shutdown 优雅关闭服务：停止 HTTP → 关闭缓存 → 断开数据库 → 触发 OnShutdown 回调。
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

	// 2. 关闭缓存 (Badger/Redis)
	if err := cache.CloseCache(); err != nil {
		errs = append(errs, errors.Wrap(err, "cache close"))
	}

	// 3. 断开数据库连接
	if err := data.CloseDB(); err != nil {
		errs = append(errs, errors.Wrap(err, "db close"))
	}

	// 4. 用户注册的清理 Hook
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

// ── Hook 注册方法 ────────────────────────────────────────────

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

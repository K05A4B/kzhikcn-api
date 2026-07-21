package cmdserve

import (
	"context"
	"kzhikcn/pkg/log"
	"kzhikcn/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
)

// Serve 返回 serve 子命令的 Action。
// 生命周期：BootstrapOrCreate → Initialize → Migrate → Ready → Serve（goroutine）→ 信号捕获 → Shutdown。
func Serve() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		app := server.New()

		configFile := ctx.String("config")
		if err := app.BootstrapOrCreate(configFile); err != nil {
			return err
		}
		if err := app.Initialize(); err != nil {
			return err
		}
		if err := app.Migrate(); err != nil {
			return err
		}
		if err := app.Ready(); err != nil {
			return err
		}

		// 在 goroutine 中启动 HTTP 服务
		errCh := make(chan error, 1)
		go func() {
			errCh <- app.Serve(ctx.String("address"))
		}()

		// 监听系统信号
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-errCh:
			// Serve 返回（启动失败 或 Shutdown 触发的 http.ErrServerClosed）
			if err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil

		case sig := <-sigCh:
			log.Info("收到信号 ", sig, "，开始优雅关闭...")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			return app.Shutdown(shutdownCtx)
		}
	}
}

package cmdmigrate

import (
	"fmt"

	"kzhikcn/internal/cli/runtime"
	"kzhikcn/server/app"

	"github.com/urfave/cli/v2"
)

func Migrate() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return runtime.WithApp(ctx, func(_ *app.App) error {
			fmt.Fprintf(ctx.App.Writer, "数据库迁移完成\n")
			return nil
		})
	}
}

var Command = &cli.Command{
	Name:     "migrate",
	Usage:    "执行数据库迁移并退出",
	Category: runtime.CategoryConfig,
	Action:   Migrate(),
}

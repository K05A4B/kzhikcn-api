package cli

import (
	"fmt"
	"kzhikcn/internal/appinfo"
	cmdadmin "kzhikcn/internal/cli/admin"
	cmdmigrate "kzhikcn/internal/cli/migrate"
	"kzhikcn/internal/cli/runtime"
	cmdserve "kzhikcn/internal/cli/serve"

	"github.com/urfave/cli/v2"
)

var AppCli = cli.App{
	Name:                 fmt.Sprintf("%s-cli", appinfo.CurrentInfo.Name),
	Usage:                "kzhikcn 无头 CMS 命令行工具",
	Version:              appinfo.CurrentInfo.Version,
	EnableBashCompletion: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Usage:   "配置文件路径",
			Aliases: []string{"c"},
			Value:   "config.yml",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "输出格式 (text|json)",
			Value: string(runtime.FormatText),
		},
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "覆盖日志级别 (debug|info|warn|error)，仅本次运行生效",
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Usage:   "静默模式，等价于 --log-level error",
			Aliases: []string{"q"},
		},
	},

	Commands: []*cli.Command{
		{
			Name:     "gen-config",
			Usage:    "生成配置文件",
			Category: runtime.CategoryBasic,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "default",
					Usage:   "输出默认配置到配置文件",
					Aliases: []string{"d"},
				},
			},
			Action: genConfig,
		},
		{
			Name:     "serve",
			Usage:    "启动服务",
			Category: runtime.CategoryBasic,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Aliases: []string{"a"},
					Name:    "address",
					Usage:   "服务地址",
					Value:   "0.0.0.0:5083",
				},
			},
			Action: func(ctx *cli.Context) error {
				return cmdserve.Serve()(runtime.WithLogOutput(ctx))
			},
		},
		ConfigCommands,
		cmdmigrate.Command,
		cmdadmin.AdminCommands,
	},
}

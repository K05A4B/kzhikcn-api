package cli

import (
	"fmt"
	"kzhikcn/internal/appinfo"
	cmdadmin "kzhikcn/internal/cli/admin"
	"kzhikcn/internal/cli/cliutils"
	cmdserve "kzhikcn/internal/cli/serve"

	"github.com/urfave/cli/v2"
)

var AppCli = cli.App{
	Name: fmt.Sprintf("%s-cli", appinfo.CurrentInfo.Name),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Usage:   "配置文件路径",
			Aliases: []string{"c"},
			Value:   "config.yml",
		},
	},

	Commands: []*cli.Command{
		{
			Name:   "gen-config",
			Usage:  "生成配置文件",
			Action: genConfig,
		},
		{
			Name:  "serve",
			Usage: "启动服务",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Aliases: []string{"a"},
					Name:    "address",
					Usage:   "服务地址",
					Value:   "0.0.0.0:5083",
				},
			},
			Action: cliutils.ConnectDatabase(cmdserve.Serve()),
		},
		cmdadmin.AdminCommands,
	},
}

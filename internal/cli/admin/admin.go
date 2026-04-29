package cmdadmin

import (
	"kzhikcn/internal/cli/cliutils"

	"github.com/urfave/cli/v2"
)

var AdminCommands = &cli.Command{
	Name:  "admin",
	Usage: "管理员相关命令",
	Subcommands: []*cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "添加管理员",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "管理员名称",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "password",
					Aliases:  []string{"p"},
					Usage:    "密码",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "email",
					Aliases: []string{"e"},
					Usage:   "电子邮件",
				},
			},
			Action: cliutils.ConnectDatabase(addAdmin),
		},
		{
			Name:      "modify",
			Aliases:   []string{"m"},
			Usage:     "修改管理员信息",
			UsageText: "使用 -n <用户名> 或 -i <ID> 指定被修改的管理员，如果使用 -i 选项指定管理员后 -n 则表示修改用户名",
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:       "id",
					Aliases:    []string{"i"},
					Usage:      "通过id选定被修改信息的管理员",
					HasBeenSet: false,
				},
				&cli.BoolFlag{
					Name:     "mfa",
					Usage:    "设置管理员MFA状态 (true: 启用 / false: 禁用)",
					Required: false,
				},
				&cli.StringFlag{
					Name:  "totp-secret",
					Usage: "设置TOTP secret",
				},
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "通过用户名选定被修改信息的管理员 / 要修改的用户名",
				},
				&cli.StringFlag{
					Name:  "email",
					Usage: "设置电子邮件地址",
				},
				&cli.StringFlag{
					Name:  "avatar",
					Usage: "设置头像图片地址",
				},
			},
			Action: cliutils.ConnectDatabase(modifyAdmin),
		},
		{
			Name:    "passwd",
			Aliases: []string{"pwd"},
			Usage:   "修改管理员密码",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "管理员账户名",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "password",
					Aliases: []string{"p"},
					Usage:   "新密码",
				},
			},
			Action: cliutils.ConnectDatabase(changePassword),
		},
		{
			Name:    "find",
			Aliases: []string{"f"},
			Usage:   "查询管理员信息",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "管理员账户名",
					Required: true,
				},
			},
			Action: cliutils.ConnectDatabase(findAdminByName),
		},
	},
}

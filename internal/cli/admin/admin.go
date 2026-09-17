package cmdadmin

import (
	"kzhikcn/internal/cli/runtime"
	"kzhikcn/server/app"

	"github.com/urfave/cli/v2"
)

// withApp 在完成数据库迁移后执行 action，结束后自动释放资源。
func withApp(action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return runtime.WithApp(ctx, func(_ *app.App) error {
			return action(ctx)
		})
	}
}

var AdminCommands = &cli.Command{
	Name:     "admin",
	Usage:    "管理员相关命令",
	Category: runtime.CategoryAdmin,
	Subcommands: []*cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "添加管理员",
			UsageText: "使用 -n 指定管理员名称；-p 省略时将交互式输入密码。\n" +
				"示例: kzhikcn-cli admin add -n alice -e alice@example.com",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "管理员名称",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "password",
					Aliases: []string{"p"},
					Usage:   "密码（省略时交互式输入）",
				},
				&cli.StringFlag{
					Name:    "email",
					Aliases: []string{"e"},
					Usage:   "电子邮件",
				},
			},
			Action: withApp(addAdmin),
		},
		{
			Name:      "modify",
			Aliases:   []string{"m"},
			Usage:     "修改管理员信息",
			UsageText: "使用 -n <用户名> 或 -i <ID> 指定被修改的管理员，如果使用 -i 选项指定管理员后 -n 则表示修改用户名",
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:    "id",
					Aliases: []string{"i"},
					Usage:   "通过id选定被修改信息的管理员",
				},
				&cli.BoolFlag{
					Name:  "mfa",
					Usage: "设置管理员MFA状态 (true: 启用 / false: 禁用)",
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
			Action: withApp(modifyAdmin),
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
					Usage:   "新密码（省略时交互式输入）",
				},
			},
			Action: withApp(changePassword),
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
				runtime.FormatFlag(),
			},
			Action: withApp(findAdminByName),
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "列出所有管理员",
			Flags: []cli.Flag{
				runtime.FormatFlag(),
			},
			Action: withApp(listAdmins),
		},
		{
			Name:      "delete",
			Aliases:   []string{"rm"},
			Usage:     "删除管理员",
			UsageText: "使用 -i <ID> 或 -n <用户名> 指定管理员，删除前需确认（非交互环境使用 --yes）",
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:    "id",
					Aliases: []string{"i"},
					Usage:   "通过 ID 指定管理员",
				},
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "通过用户名指定管理员",
				},
				&cli.BoolFlag{
					Name:    "yes",
					Aliases: []string{"y"},
					Usage:   "跳过删除确认",
				},
			},
			Action: withApp(deleteAdmin),
		},
		{
			Name:      "mfa",
			Usage:     "生成或重置管理员的 TOTP 密钥",
			UsageText: "使用 -i <ID> 或 -n <用户名> 指定管理员；已有密钥时需 --force 才会覆盖",
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:    "id",
					Aliases: []string{"i"},
					Usage:   "通过 ID 指定管理员",
				},
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "通过用户名指定管理员",
				},
				&cli.BoolFlag{
					Name:  "enable",
					Usage: "生成后直接启用 MFA",
				},
				&cli.BoolFlag{
					Name:  "force",
					Usage: "覆盖已存在的 TOTP 密钥",
				},
			},
			Action: withApp(generateMFA),
		},
	},
}

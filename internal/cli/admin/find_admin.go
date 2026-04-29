package cmdadmin

import (
	"fmt"
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func findAdminByName(ctx *cli.Context) error {
	username := ctx.String("name")
	admin, err := data.GetAdminByName(username)
	if err != nil {
		return errors.Wrap(err, "查询失败")
	}

	fmt.Fprintf(ctx.App.Writer, "ID: %d\n", admin.ID)
	fmt.Fprintf(ctx.App.Writer, "用户名: %s\n", admin.Username)
	fmt.Fprintf(ctx.App.Writer, "是否启用2FA: %s\n", map[bool]string{true: "是", false: "否"}[admin.EnableMFA])
	fmt.Fprintf(ctx.App.Writer, "电子邮件: %s\n", admin.Email)
	fmt.Fprintf(ctx.App.Writer, "头像链接: %s\n", admin.Avatar)

	return nil
}

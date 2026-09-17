package cmdadmin

import (
	"fmt"
	"io"
	"strings"

	"kzhikcn/internal/cli/runtime"
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

func findAdminByName(ctx *cli.Context) error {
	username := strings.TrimSpace(ctx.String("name"))

	admin, err := data.GetAdminByName(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Errorf("没有找到管理员 %s", username)
	}
	if err != nil {
		return errors.Wrap(err, "查询失败")
	}

	return runtime.PrintResult(ctx, newAdminView(admin), func(w io.Writer) error {
		fmt.Fprintf(w, "ID: %d\n", admin.ID)
		fmt.Fprintf(w, "用户名: %s\n", admin.Username)
		fmt.Fprintf(w, "是否启用2FA: %s\n", mfaText(admin.EnableMFA))
		fmt.Fprintf(w, "电子邮件: %s\n", admin.Email)
		fmt.Fprintf(w, "头像链接: %s\n", admin.Avatar)
		return nil
	})
}

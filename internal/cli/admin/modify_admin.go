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

func modifyAdmin(ctx *cli.Context) error {
	admin, err := resolveAdminByFlags(ctx)
	if err != nil {
		return err
	}

	selected := []string{}
	update := data.Admin{
		ID:     admin.ID,
		Avatar: ctx.String("avatar"),
		Email:  ctx.String("email"),
	}

	if ctx.IsSet("id") && ctx.IsSet("name") {
		username := strings.TrimSpace(ctx.String("name"))
		if username == "" {
			return errors.New("用户名不能为空")
		}
		update.Username = username
		selected = append(selected, "username")
	}

	if ctx.IsSet("mfa") {
		update.EnableMFA = ctx.Bool("mfa")
		selected = append(selected, "enable_mfa")
	}

	if ctx.IsSet("totp-secret") {
		secret := strings.TrimSpace(ctx.String("totp-secret"))
		if secret == "" {
			return errors.New("TOTP secret 不能为空")
		}
		update.TotpSecret = []byte(secret)
		selected = append(selected, "totp_secret")
	}

	for _, field := range []string{"avatar", "email"} {
		if ctx.IsSet(field) {
			selected = append(selected, field)
		}
	}

	if len(selected) == 0 {
		return errors.New("没有需要修改的字段")
	}

	err = data.UpdateAdminByID(admin.ID, &update, func(tx *gorm.DB) *gorm.DB {
		return tx.Select(selected)
	})
	if err != nil {
		return errors.Wrap(err, "更新管理员信息失败")
	}

	return runtime.PrintResult(ctx, map[string]any{
		"id":      admin.ID,
		"updated": selected,
	}, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, "管理员 %s 信息已更新\n", admin.Username)
		return err
	})
}

package cmdadmin

import (
	"kzhikcn/pkg/data"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

func modifyAdmin(ctx *cli.Context) error {
	selected := []string{}

	id := ctx.Uint("id")
	if !ctx.IsSet("id") {
		username := strings.TrimSpace(ctx.String("name"))
		if username == "" {
			return errors.New("需要提供被修改信息的管理员的ID或用户名")
		}

		var err error
		id, err = data.GetAdminIDByName(username)
		if err == gorm.ErrRecordNotFound {
			return errors.New("没有找到管理员" + username)
		}

		if err != nil {
			return errors.Wrap(err, "查找管理员失败")
		}
	}

	admin := data.Admin{
		ID:        id,
		Avatar:    ctx.String("avatar"),
		Email:     ctx.String("email"),
		EnableMFA: ctx.Bool("mfa"),
	}

	if ctx.IsSet("id") && ctx.IsSet("name") {
		admin.Username = ctx.String("name")
		selected = append(selected, "username")
	}

	if ctx.IsSet("mfa") {
		selected = append(selected, "enable_mfa")
	}

	if ctx.IsSet("totp-secret") {
		secret := ctx.String("secret")
		admin.TotpSecret = []byte(secret)

		selected = append(selected, "totp_secret")
	}

	fields := []string{"avatar", "email"}

	for _, v := range fields {
		if ctx.IsSet(v) {
			selected = append(selected, v)
		}
	}

	err := data.UpdateAdminByID(admin.ID, &admin, func(tx *gorm.DB) *gorm.DB {
		return tx.Select(selected)
	})

	if err != nil {
		return errors.Wrap(err, "更新管理员信息失败")
	}

	return nil
}

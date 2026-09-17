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

func changePassword(ctx *cli.Context) error {
	username := strings.TrimSpace(ctx.String("name"))

	admin, err := data.GetAdminByName(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Errorf("没有找到管理员 %s", username)
	}
	if err != nil {
		return errors.Wrap(err, "查找管理员失败")
	}

	password, err := resolvePassword(ctx, "请输入新密码")
	if err != nil {
		return err
	}
	if len(password) < minPasswordLength {
		return errors.Errorf("密码长度不能少于 %d 位", minPasswordLength)
	}

	err = data.UpdateAdminByID(admin.ID, &data.Admin{Password: []byte(password)}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("password")
	})
	if err != nil {
		return errors.Wrap(err, "修改失败")
	}

	return runtime.PrintResult(ctx, map[string]any{"id": admin.ID, "username": admin.Username}, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, "管理员 %s 的密码已修改\n", admin.Username)
		return err
	})
}

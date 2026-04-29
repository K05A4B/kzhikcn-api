package cmdadmin

import (
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

func changePassword(ctx *cli.Context) error {
	username := ctx.String("name")
	password := ctx.String("password")

	id, err := data.GetAdminIDByName(username)
	if err == gorm.ErrRecordNotFound {
		return errors.New("没有找到用户" + username)
	}

	if err != nil {
		return errors.Wrap(err, "查找用户失败")
	}

	err = data.UpdateAdminByID(id, &data.Admin{Password: []byte(password)}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("password")
	})

	if err != nil {
		return errors.Wrap(err, "修改失败")
	}

	return nil
}

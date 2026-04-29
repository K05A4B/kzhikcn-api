package cmdadmin

import (
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func addAdmin(ctx *cli.Context) error {
	name := ctx.String("name")
	password := ctx.String("password")
	email := ctx.String("email")

	err := data.AddAdmin(&data.Admin{
		Username: name,
		Password: []byte(password),
		Email:    email,
	})

	if err != nil {
		return errors.Wrap(err, "添加失败")
	}

	return nil
}

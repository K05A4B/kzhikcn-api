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

func addAdmin(ctx *cli.Context) error {
	name := strings.TrimSpace(ctx.String("name"))
	if name == "" {
		return errors.New("管理员名称不能为空")
	}

	password, err := resolvePassword(ctx, "请输入密码")
	if err != nil {
		return err
	}
	if len(password) < minPasswordLength {
		return errors.Errorf("密码长度不能少于 %d 位", minPasswordLength)
	}

	if _, err := data.GetAdminByName(name); err == nil {
		return errors.Errorf("管理员 %s 已存在", name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "检查管理员是否存在失败")
	}

	err = data.AddAdmin(&data.Admin{
		Username: name,
		Password: []byte(password),
		Email:    strings.TrimSpace(ctx.String("email")),
	})
	if err != nil {
		return errors.Wrap(err, "添加失败")
	}

	return runtime.PrintResult(ctx, map[string]any{"username": name}, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, "管理员 %s 添加成功\n", name)
		return err
	})
}

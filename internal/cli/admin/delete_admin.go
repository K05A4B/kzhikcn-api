package cmdadmin

import (
	"fmt"
	"io"

	"kzhikcn/internal/cli/runtime"
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func deleteAdmin(ctx *cli.Context) error {
	admin, err := resolveAdminByFlags(ctx)
	if err != nil {
		return err
	}

	ok, err := runtime.Confirm(ctx, fmt.Sprintf("确认删除管理员 %s (ID: %d)?", admin.Username, admin.ID))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(ctx.App.Writer, "已取消")
		return nil
	}

	if err := data.DeleteAdminByID(admin.ID); err != nil {
		return errors.Wrap(err, "删除管理员失败")
	}

	return runtime.PrintResult(ctx, map[string]any{
		"id":       admin.ID,
		"username": admin.Username,
	}, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, "管理员 %s 已删除\n", admin.Username)
		return err
	})
}

package cmdadmin

import (
	"fmt"
	"io"
	"text/tabwriter"

	"kzhikcn/internal/cli/runtime"
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func listAdmins(ctx *cli.Context) error {
	admins, err := data.GetAdmin()
	if err != nil {
		return errors.Wrap(err, "查询管理员列表失败")
	}

	views := make([]adminView, 0, len(admins))
	for i := range admins {
		views = append(views, newAdminView(&admins[i]))
	}

	return runtime.PrintResult(ctx, views, func(w io.Writer) error {
		if len(views) == 0 {
			_, err := fmt.Fprintln(w, "暂无管理员")
			return err
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\t用户名\tMFA\t电子邮件\t头像")
		for _, v := range views {
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", v.ID, v.Username, mfaText(v.EnableMFA), v.Email, v.Avatar)
		}
		return tw.Flush()
	})
}

package cmdadmin

import (
	"strings"

	"kzhikcn/internal/cli/runtime"
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

const minPasswordLength = 6

// adminView 是面向 CLI 输出的管理员视图，避免泄露密码等敏感字段。
type adminView struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	EnableMFA bool   `json:"enableMFA"`
}

func newAdminView(admin *data.Admin) adminView {
	return adminView{
		ID:        admin.ID,
		Username:  admin.Username,
		Email:     admin.Email,
		Avatar:    admin.Avatar,
		EnableMFA: admin.EnableMFA,
	}
}

func mfaText(enabled bool) string {
	if enabled {
		return "是"
	}
	return "否"
}

// resolveAdminByFlags 根据 --id 优先、否则 --name 定位管理员。
func resolveAdminByFlags(ctx *cli.Context) (*data.Admin, error) {
	if ctx.IsSet("id") {
		id := ctx.Uint("id")
		admin, err := data.GetAdminById(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Errorf("没有找到 ID 为 %d 的管理员", id)
		}
		if err != nil {
			return nil, errors.Wrap(err, "查找管理员失败")
		}
		return admin, nil
	}

	name := strings.TrimSpace(ctx.String("name"))
	if name == "" {
		return nil, errors.New("需要提供管理员的 ID 或用户名")
	}

	admin, err := data.GetAdminByName(name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Errorf("没有找到管理员 %s", name)
	}
	if err != nil {
		return nil, errors.Wrap(err, "查找管理员失败")
	}

	return admin, nil
}

// resolvePassword 优先使用 --password，未提供时从终端隐藏读取。
func resolvePassword(ctx *cli.Context, label string) (string, error) {
	if password := ctx.String("password"); password != "" {
		return password, nil
	}

	return runtime.PromptPassword(ctx, label)
}

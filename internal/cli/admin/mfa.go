package cmdadmin

import (
	"fmt"
	"io"

	"kzhikcn/internal/appinfo"
	"kzhikcn/internal/cli/runtime"
	"kzhikcn/pkg/data"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

// generateMFA 为管理员生成 TOTP 密钥并输出 otpauth URL。
// 已有密钥时需显式 --force 才会覆盖，避免破坏已绑定的验证器。
func generateMFA(ctx *cli.Context) error {
	admin, err := resolveAdminByFlags(ctx)
	if err != nil {
		return err
	}

	if len(admin.TotpSecret) > 0 && !ctx.Bool("force") {
		return errors.New("该管理员已配置 TOTP，如需重新生成请使用 --force")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appinfo.Name,
		AccountName: admin.Username,
	})
	if err != nil {
		return errors.Wrap(err, "生成 TOTP 密钥失败")
	}

	err = data.UpdateAdminByID(admin.ID, &data.Admin{TotpSecret: []byte(key.Secret())}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("totp_secret")
	})
	if err != nil {
		return errors.Wrap(err, "保存 TOTP 密钥失败")
	}

	enable := ctx.Bool("enable")
	if enable {
		err = data.UpdateAdminByID(admin.ID, &data.Admin{EnableMFA: true}, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("enable_mfa")
		})
		if err != nil {
			return errors.Wrap(err, "启用 MFA 失败")
		}
	}

	result := map[string]any{
		"id":        admin.ID,
		"username":  admin.Username,
		"secret":    key.Secret(),
		"url":       key.URL(),
		"enableMFA": enable,
	}

	return runtime.PrintResult(ctx, result, func(w io.Writer) error {
		fmt.Fprintf(w, "管理员: %s\n", admin.Username)
		fmt.Fprintf(w, "TOTP Secret: %s\n", key.Secret())
		fmt.Fprintf(w, "otpauth URL: %s\n", key.URL())
		if enable {
			fmt.Fprintln(w, "MFA 已启用")
		}
		return nil
	})
}

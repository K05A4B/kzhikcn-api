package admin

import (
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/secutils"
	"net/http"

	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type GenerateTOTPSecretRequest struct {
	Password string `json:"password"`
}

type GenerateTOTPSecretResponse struct {
	Secret      string `json:"secret"`
	AccountName string `json:"accountName"`
	Issuer      string `json:"issuer"`
	URL         string `json:"url"`
}

var GenerateTOTPSecretHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload GenerateTOTPSecretRequest) error {
		claims := authtoken.GetClaims(r.Context())
		admin, err := data.GetAdminById(claims.AdminId, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("password", "id")
		})

		if err == gorm.ErrRecordNotFound {
			return ErrAdminNotFound
		}

		if err != nil {
			return ErrFindAdminFailed.Wrap(err)
		}

		ok, _ := secutils.ComparePassword(admin.Password, payload.Password)
		if !ok {
			return ErrAdminComparePasswordFailed
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      appinfo.CurrentInfo.Name,
			AccountName: claims.Subject,
		})
		if err != nil {
			return ErrAdminGenerateTOTPFailed.Wrap(err)
		}

		admin.TotpSecret = []byte(key.Secret())

		err = data.UpdateAdminByID(admin.ID, admin, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("totp_secret")
		})
		if err != nil {
			return ErrAdminUpdateTOTPSecretFailed.Wrap(err)
		}

		resp.Data = GenerateTOTPSecretResponse{
			Secret:      key.Secret(),
			AccountName: key.AccountName(),
			Issuer:      key.Issuer(),
			URL:         key.URL(),
		}

		return nil
	},

	hdl.MissingFields(func(payload GenerateTOTPSecretRequest) []string {
		missing := []string{}

		if utils.IsEmptyString(payload.Password) {
			missing = append(missing, "password")
		}
		return missing
	}),
)

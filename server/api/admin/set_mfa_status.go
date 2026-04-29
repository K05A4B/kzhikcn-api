package admin

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/secutils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type SetMFAStatusRequest struct {
	Password string `json:"password"`
	Otp      string `json:"otp"`
}

var SetMFAStatusHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload SetMFAStatusRequest) error {
		claims := authtoken.GetClaims(r.Context())
		admin, err := data.GetAdminById(claims.AdminId, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("totp_secret", "password", "id")
		})

		if err == gorm.ErrRecordNotFound {
			return ErrAdminNotFound
		}

		if err != nil {
			return ErrFindAdminFailed.Wrap(err)
		}

		ok, _ := secutils.ComparePassword(admin.Password, payload.Password)
		if !ok {
			return ErrAdminComparePasswordFailed.Wrap(err)
		}

		if !totp.Validate(payload.Otp, string(admin.TotpSecret)) {
			return ErrAdminInvalidOTP
		}

		switch chi.URLParam(r, "action") {
		case "disable":
			admin.EnableMFA = false
		case "enable":
			admin.EnableMFA = true

		default:
			return nil

		}

		err = data.UpdateAdminByID(admin.ID, admin, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("enable_mfa")
		})

		if err != nil {
			return ErrAdminUpdateMFAFailed.Wrap(err)
		}

		return nil
	},

	hdl.MissingFields(func(payload SetMFAStatusRequest) []string {
		missing := []string{}

		if utils.IsEmptyString(payload.Password) {
			missing = append(missing, "password")
		}

		if utils.IsEmptyString(payload.Otp) {
			missing = append(missing, "otp")
		}

		return missing
	}),
)

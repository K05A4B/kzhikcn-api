package admin

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/service"
	"net/http"

	"github.com/pkg/errors"
)

type GenerateTOTPSecretRequest struct {
	Password string `json:"password"`
}

func GenerateTOTPSecret(appCtx *app.AppContext) hdl.Handler[GenerateTOTPSecretRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload GenerateTOTPSecretRequest) error {
			claims := authtoken.GetClaims(r.Context())

			result, err := appCtx.AdminSvc.GenerateTOTP(r.Context(), claims.AdminId, payload.Password, claims.Subject)
			if errors.Is(err, service.ErrAdminValidateFailed) {
				return ErrAdminComparePasswordFailed
			}
			if err != nil {
				return ErrAdminGenerateTOTPFailed.Wrap(err)
			}

			resp.Data = *result
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
}

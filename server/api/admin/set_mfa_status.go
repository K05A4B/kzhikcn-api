package admin

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

type SetMFAStatusRequest struct {
	Password string `json:"password"`
	Otp      string `json:"otp"`
}

func SetMFAStatus(appCtx *app.AppContext) hdl.Handler[SetMFAStatusRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload SetMFAStatusRequest) error {
			claims := authtoken.GetClaims(r.Context())

			enable := chi.URLParam(r, "action") == "enable"

			err := appCtx.AdminSvc.SetMFA(r.Context(), claims.AdminId, enable, payload.Password, payload.Otp)
			if errors.Is(err, service.ErrAdminValidateFailed) {
				return ErrAdminComparePasswordFailed.Wrap(err)
			}
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
}

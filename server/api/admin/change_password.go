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

type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
}

func ChangePassword(appCtx *app.AppContext) hdl.Handler[ChangePasswordRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload ChangePasswordRequest) error {
			claims := authtoken.GetClaims(r.Context())

			err := appCtx.AdminSvc.ChangePassword(r.Context(), claims.AdminId, payload.OldPassword, payload.NewPassword)
			if errors.Is(err, service.ErrAdminValidateFailed) {
				return ErrAdminValidateFailed
			}
			if errors.Is(err, service.ErrAdminComparePasswordFail) {
				return ErrAdminComparePasswordFailed.Wrap(err)
			}
			if err != nil {
				return ErrChangePasswordFailed.Wrap(err)
			}

			return nil
		},

		hdl.MissingFields(func(payload ChangePasswordRequest) []string {
			fields := []string{}
			if utils.IsEmptyString(payload.OldPassword) {
				fields = append(fields, "oldPassword")
			}
			if utils.IsEmptyString(payload.NewPassword) {
				fields = append(fields, "newPassword")
			}
			return fields
		}),
	)
}

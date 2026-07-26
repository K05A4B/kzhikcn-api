package admin

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/service"
	"net/http"
)

type UpdateAdminInfoRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func UpdateAdminInfo(appCtx *app.AppContext) hdl.Handler[UpdateAdminInfoRequest] {
	return hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload UpdateAdminInfoRequest) error {
		claims := authtoken.GetClaims(r.Context())

		fields := service.AdminUpdateFields{}
		if !utils.IsEmptyString(payload.Avatar) {
			fields.Avatar = &payload.Avatar
		}
		if !utils.IsEmptyString(payload.Username) {
			fields.Username = &payload.Username
		}
		if !utils.IsEmptyString(payload.Email) {
			fields.Email = &payload.Email
		}

		err := appCtx.AdminSvc.UpdateInfo(r.Context(), claims.AdminId, fields)
		if err != nil {
			return ErrUpdateAdminInfoFailed.Wrap(err)
		}

		return nil
	})
}

package admin

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"net/http"
)

func AdminInfo(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		claims := authtoken.GetClaims(r.Context())

		admin, err := appCtx.AdminSvc.GetProfile(r.Context(), claims.AdminId)
		if err != nil {
			return ErrFindAdminFailed.Wrap(err)
		}

		resp.Data = admin
		return nil
	})
}

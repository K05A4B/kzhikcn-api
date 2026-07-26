package auth

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"net/http"
)

func Logout(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		err := appCtx.AuthSvc.Logout(r.Context(), authtoken.GetClaims(r.Context()))

		if err != nil {
			return ErrRevokeTokenFailed.Wrap(err)
		}

		return nil
	})
}

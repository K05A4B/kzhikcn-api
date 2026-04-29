package auth

import (
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"net/http"
)

// 注销接口
// POST /api/v1/auth/logout
//
// 认证要求:
//   - 无需认证
//
// 请求类型: 无
//
// 请求参数: 无

// 响应数据: 无

// 错误码:
//   - auth.logout.revoke_token_failed: 撤销token失败
var LogoutHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	claims := authtoken.GetClaims(r.Context())
	err := authtoken.RevokeToken(r.Context(), claims)
	if err != nil {
		return ErrRevokeTokenFailed.Wrap(err)
	}
	return nil
})

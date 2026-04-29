package admin

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

// 查询当前登录用户的信息
// GET /api/v1/users/admin/me
//
// 认证要求:
//   - 需要提供Token
//
// 请求类型: 无
//
// 请求参数: 无
//
// 响应数据:
//
//	data:
//		- id uint 用户ID
//		- username string 用户名
//		- email string 邮件地址
//		- avatar string 头像
//		- enable2FA bool 是否启用MFA
//
// 错误码:
var AdminInfoHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	claims := authtoken.GetClaims(r.Context())

	admin, err := data.GetAdminById(claims.AdminId, func(tx *gorm.DB) *gorm.DB {
		return tx.Omit("password", "tow_fa_secret")
	})

	if err != nil {
		return ErrFindAdminFailed.Wrap(err)
	}

	resp.Data = admin

	return nil
})

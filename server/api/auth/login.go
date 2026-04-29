package auth

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/secutils"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type LoginResponse struct {
	// status 表示认证状态：
	// - authorized: 登录成功
	// - needMFA: 需要进行多因素认证
	Status string `json:"status"`

	// token 为 Access Token（仅 status=authorized 时返回）
	Token string `json:"token,omitempty"`

	// challengeId 为 MFA 挑战 ID（仅 status=needMFA 时返回）
	// 有效期为 120 秒，过期后需重新发起登录流程获取
	ChallengeId string `json:"challengeId,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 登录接口
// POST /api/v1/auth/login
//
// 认证要求:
//   - 无需认证
//
// 请求类型:
//   - Content-Type: application/json
//
// 请求参数:
//   - username (string, 必填): 用户名
//     示例: "admin"
//   - password (string, 必填): 密码
//     示例: "123456"
//
// 响应数据:
//
//	data:
//	  - status (string): 认证状态
//	    - authorized: 登录成功
//	    - needMFA: 需要进行多因素认证
//
//	  - token (string, 可选): Access Token
//	    - 返回条件: status = authorized
//
//	  - challengeId (string, 可选): MFA 挑战 ID
//	    - 返回条件: status = needMFA
//	    - 有效期: 120 秒
//			- 允许重试次数：5
//	    - 说明: 过期后需重新登录获取
//
// 错误码:
//   - auth.login.authentication_failed: 用户名或密码错误
//   - auth.login.compare_password_failed: 密码校验异常
//   - database.query_error: 数据库查询异常
//   - auth.generate_token_failed: 生成Token失败
var LoginHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload LoginRequest) error {

		admin, err := data.GetAdminByName(payload.Username, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("password", "username", "id", "enable_mfa")
		})
		if err == gorm.ErrRecordNotFound {
			return ErrAuthenticationFailed
		}
		if err != nil {
			return ErrFindAdminFailed.Wrap(err)
		}

		ok, err := secutils.ComparePassword(admin.Password, payload.Password)
		if err != nil {
			return ErrValidatePasswordFailed.Wrap(err)
		}

		if !ok {
			return ErrAuthenticationFailed
		}

		respData := &LoginResponse{}
		defer func() { resp.Data = respData }()

		if admin.EnableMFA {
			respData.Status = "needMFA"
			respData.ChallengeId = utils.RandomString(24)

			err = cache.SetJson(r.Context(), cache.Keys(mfaChallengesKey, respData.ChallengeId), &challenge{
				Username:    admin.Username,
				UserId:      admin.ID,
				MaxAttempts: 5,
				Expire:      time.Now().Add(2 * time.Minute),
			}, 2*time.Minute)

			if err != nil {
				return ErrCreateChallengeFailed.Wrap(err)
			}

			return nil
		}

		respData.Status = "authorized"
		resp.Message = "认证成功（建议启用MFA）"

		token, err := permitLogin(r, admin)
		if err != nil {
			return ErrGenerateTokenFailed.Wrap(err)
		}

		respData.Token = token

		return nil
	},

	hdl.MissingFields(func(payload LoginRequest) []string {
		missing := []string{}

		if utils.IsEmptyString(payload.Username) {
			missing = append(missing, "username")
		}

		if utils.IsEmptyString(payload.Password) {
			missing = append(missing, "password")
		}

		return missing
	}),
)

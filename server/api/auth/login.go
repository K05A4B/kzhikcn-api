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

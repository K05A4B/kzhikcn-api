package auth

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/hdl"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type VerifyTOTPRequest struct {
	OTP         string `json:"otp"`
	ChallengeID string `json:"challengeId"`
}

type VerifyTOTPResponse struct {
	Token string `json:"token"`
}

// 验证TOTP接口
// POST /api/v1/auth/mfa/totp
//
// 认证要求:
//   - 无需认证
//
// 请求类型:
//   - Content-Type: application/json
//
// 请求参数:
//   - challengeId (string, 必填): 挑战ID
//     示例: "sR0NQ00AtZGy2HWmi99q5cfLq"
//   - otp (string, 必填): 验证码
//     示例: "123456"
//
// 响应数据:
//
//	data:
//	  - token (string, 可选): Refresh Token
//
// 错误码:
//   - auth.mfa.invalid_challenge: 无效挑战
//   - auth.mfa.decrypt_totp_secret_failed: totp secret 解密失败
//   - auth.mfa.validation_failed: 验证码验证失败
//   - database.query_error: 数据库查询异常
//   - auth.generate_token_failed: 生成刷新令牌失败
var VerifyTOTPHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload VerifyTOTPRequest) error {
		challengeId := payload.ChallengeID
		key := cache.Keys(mfaChallengesKey, challengeId)

		ok, err := cache.Exists(r.Context(), key)
		if err != nil {
			return ErrGetChallengeFailed.Wrap(err)
		}
		if !ok {
			return ErrInvalidChallenge
		}

		cha := &challenge{}

		err = cache.GetJson(r.Context(), key, cha)
		if cha.MaxAttempts <= 0 {
			return ErrInvalidChallenge
		}

		cha.MaxAttempts--

		err = cache.SetJson(r.Context(), key, cha, time.Until(cha.Expire))
		if err != nil {
			log.Error(err)
			cache.Delete(r.Context(), key)
			return ErrInvalidChallenge
		}

		admin, err := data.GetAdminById(cha.UserId, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("totp_secret", "username", "id")
		})

		if err != nil {
			return ErrFindAdminFailed.Wrap(err)
		}

		if !totp.Validate(payload.OTP, string(admin.TotpSecret)) {
			return ErrAuthenticationFailed
		}

		err = cache.Delete(r.Context(), key)
		if err != nil {
			return ErrCleanChallengeFailed.Wrap(err)
		}

		token, err := permitLogin(r, admin)
		if err != nil {
			return ErrGenerateTokenFailed.Wrap(err)
		}

		resp.Data = VerifyTOTPResponse{
			Token: token,
		}

		return nil
	},

	hdl.MissingFields(func(payload VerifyTOTPRequest) []string {
		missing := []string{}

		if utils.IsEmptyString(payload.ChallengeID) {
			missing = append(missing, "challengeId")
		}

		if utils.IsEmptyString(payload.OTP) {
			missing = append(missing, "otp")
		}

		return missing
	}),
)

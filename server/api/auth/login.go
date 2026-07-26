package auth

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/app"
	"kzhikcn/server/service"
	"net/http"

	"github.com/pkg/errors"
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

func Login(appCtx *app.AppContext) hdl.Handler[LoginRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload LoginRequest) error {
			return loginHandler(r, resp, payload, appCtx)
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
		},
		))
}

func loginHandler(r *http.Request, resp *hdl.Response, payload LoginRequest, appCtx *app.AppContext) error {
	adminSvc := appCtx.AdminSvc
	authSvc := appCtx.AuthSvc

	username := payload.Username
	password := payload.Password

	admin, err := adminSvc.GetAdminByName(r.Context(), username)
	if err == gorm.ErrRecordNotFound {
		return ErrAuthenticationFailed
	}

	if err != nil {
		return ErrFindAdminFailed.Wrap(err)
	}

	token, err := authSvc.AdminLogin(r.Context(), admin, password)

	if errors.Is(err, service.ErrValidatePasswordFailed) {
		return ErrValidatePasswordFailed.Wrap(err)
	}

	if errors.Is(err, service.ErrGenerateTokenFailed) {
		return ErrGenerateTokenFailed.Wrap(err)
	}

	if errors.Is(err, service.ErrAuthenticationFailed) {
		return ErrAuthenticationFailed
	}

	result := &LoginResponse{}
	defer func() { resp.Data = result }()

	if errors.Is(err, service.ErrMFARequired) {
		result.Status = "needMFA"
		challenge, err := authSvc.CreateChallenge(r.Context(), username, admin.ID)
		if err != nil {
			return ErrCreateChallengeFailed.Wrap(err)
		}

		result.ChallengeId = challenge.ID

		return nil
	}

	result.Status = "authorized"
	result.Token = token
	resp.Message = "认证成功（建议启用MFA）"

	return nil
}

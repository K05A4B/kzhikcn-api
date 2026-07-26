package auth

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/service"
	"net/http"

	"github.com/pkg/errors"
)

type VerifyTOTPRequest struct {
	OTP         string `json:"otp"`
	ChallengeID string `json:"challengeId"`
}

type VerifyTOTPResponse struct {
	Token string `json:"token"`
}

func VerifyTOTP(appCtx *app.AppContext) hdl.Handler[VerifyTOTPRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload VerifyTOTPRequest) error {
			challenge, err := appCtx.AuthSvc.GetChallenge(r.Context(), payload.ChallengeID)
			if err != nil || challenge.IsExpired() {
				return ErrInvalidChallenge.Wrap(err)
			}

			admin, err := appCtx.AdminSvc.GetAdminById(r.Context(), challenge.UserID)
			if err != nil {
				return ErrFindAdminFailed.Wrap(err)
			}

			err = appCtx.AuthSvc.TOTPValidate(r.Context(), string(admin.TotpSecret), challenge, payload.OTP)
			if errors.Is(err, service.ErrAuthenticationFailed) {
				return ErrAuthenticationFailed
			}
			if err != nil {
				return ErrInvalidChallenge
			}

			challenge.Delete(r.Context())

			token, err := authtoken.IssueToken(admin.ID, admin.Username, true)
			if err != nil {
				return ErrGenerateTokenFailed.Wrap(err)
			}

			resp.Data = VerifyTOTPResponse{Token: token}
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
}

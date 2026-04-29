package admin

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

type UpdateAdminInfoRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

var UpdateAdminInfoHandler = hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload UpdateAdminInfoRequest) error {
	claims := authtoken.GetClaims(r.Context())
	selectedFields := []string{}

	if !utils.IsEmptyString(payload.Avatar) {
		selectedFields = append(selectedFields, "avatar")
	}

	if !utils.IsEmptyString(payload.Username) {
		selectedFields = append(selectedFields, "username")
	}

	if !utils.IsEmptyString(payload.Email) {
		selectedFields = append(selectedFields, "email")
	}

	err := data.UpdateAdminByID(claims.AdminId, &data.Admin{
		Avatar:   payload.Avatar,
		Username: payload.Username,
		Email:    payload.Email,
	}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select(selectedFields)
	})

	if err != nil {
		return ErrUpdateAdminInfoFailed.Wrap(err)
	}

	return nil
})

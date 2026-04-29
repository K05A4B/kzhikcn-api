package admin

import (
	"errors"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/secutils"
	"net/http"

	"gorm.io/gorm"
)

type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
}

var ChangePasswordHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload ChangePasswordRequest) error {
		claims := authtoken.GetClaims(r.Context())
		admin, err := data.GetAdminById(claims.AdminId, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "password")
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAdminNotFound
		}

		if err != nil {
			return ErrFindAdminFailed.Wrap(err)
		}

		ok, err := secutils.ComparePassword(admin.Password, payload.OldPassword)
		if err != nil {
			return ErrAdminComparePasswordFailed.Wrap(err)
		}

		if !ok {
			return ErrAdminValidateFailed
		}

		admin.Password = []byte(payload.NewPassword)

		err = data.UpdateAdminByID(claims.AdminId, admin, func(tx *gorm.DB) *gorm.DB {
			return tx.Select("password")
		})

		if err != nil {
			return ErrChangePasswordFailed.Wrap(err)
		}

		return nil
	},

	hdl.MissingFields(func(payload ChangePasswordRequest) []string {
		fields := []string{}

		if utils.IsEmptyString(payload.OldPassword) {
			fields = append(fields, "oldPassword")
		}

		if utils.IsEmptyString(payload.NewPassword) {
			fields = append(fields, "newPassword")
		}

		return fields
	}),
)

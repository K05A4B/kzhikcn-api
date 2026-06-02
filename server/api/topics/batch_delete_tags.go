package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

type BatchDeleteTopicsRequest struct {
	IDs []uint `json:"ids"`
}

var BatchDeleteTagsHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload BatchDeleteTopicsRequest) error {
		ids := payload.IDs
		err := data.DeleteTag(func(tx *gorm.DB) *gorm.DB {
			return tx.Where("id IN ?", ids)
		})

		if err != nil {
			return ErrTagDeleteFailed
		}

		return nil
	},

	hdl.When(func(payload BatchDeleteTopicsRequest) bool {
		return len(payload.IDs) == 0
	}, ErrTagDeleteIdIsRequired),
)

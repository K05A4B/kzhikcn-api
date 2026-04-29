package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

type BatchDeleteTagsRequest struct {
	IDs []uint `json:"ids"`
}

var BatchDeleteTagsHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload BatchDeleteTagsRequest) error {
		ids := payload.IDs
		err := data.DeleteTag(func(tx *gorm.DB) *gorm.DB {
			return tx.Where("id IN ?", ids)
		})

		if err != nil {
			return ErrTagDeleteFailed
		}

		return nil
	},

	hdl.When(func(payload BatchDeleteTagsRequest) bool {
		return len(payload.IDs) == 0
	}, ErrTagDeleteIdIsRequired),
)

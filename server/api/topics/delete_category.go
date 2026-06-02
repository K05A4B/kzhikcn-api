package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var DeleteCategoryHandler = hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload BatchDeleteTopicsRequest) error {
	err := data.DeleteCategories(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id IN ?", payload.IDs)
	})

	if err != nil {
		return ErrCategoryDeleteFailed.Wrap(err)
	}

	return nil
},
	hdl.When(func(payload BatchDeleteTopicsRequest) bool {
		return len(payload.IDs) == 0
	}, ErrCategoryDeleteIdIsRequired),
)

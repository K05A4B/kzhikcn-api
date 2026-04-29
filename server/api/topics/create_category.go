package topics

import (
	"errors"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var CreateCategoryHandler = hdl.NewHandler[data.EditableCategory](
	func(r *http.Request, resp *hdl.Response, payload data.EditableCategory) error {
		category, err := data.CreateCategory(payload)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrCategoryIsExist
		}

		if err != nil {
			return ErrCategoryCreateFailed
		}

		resp.Data = category

		return nil
	},

	hdl.When(func(payload data.EditableCategory) bool {
		return utils.IsEmptyString(payload.CategoryName)
	}, ErrCategoryCreateCategoryNameIsRequired),
)

package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var DeleteCategoryHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	category, err := getCategoryBase(r)
	if err != nil {
		return err
	}

	err = data.DeleteCategories(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id=?", category.ID)
	})

	if err != nil {
		return ErrCategoryDeleteFailed.Wrap(err)
	}

	return nil
})

package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var UpdateCategoryHandler = hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload data.EditableCategory) error {
	category, err := getCategoryBase(r)
	if err != nil {
		return err
	}

	err = category.Update(payload)
	if err != nil {
		return ErrCategoryUpdateFailed.Wrap(err)
	}

	return nil
})

func getCategoryBase(r *http.Request, modifiers ...data.QueryModifier) (*data.Category, error) {
	name := chi.URLParam(r, "category")
	unescapeName, err := url.QueryUnescape(name)
	if err == nil {
		name = unescapeName
	}

	categories, err := data.GetCategories(func(tx *gorm.DB) *gorm.DB {
		id, err := strconv.ParseUint(name, 10, 64)
		if err == nil {
			tx = tx.Where("id=?", id)
		} else {
			tx = tx.Where("category_name=?", name)
		}

		tx = tx.Limit(1).Select("id")

		return data.ApplyQueryModifier(tx, modifiers...)
	})

	if len(categories) == 0 {
		return nil, ErrCategoryNotFound
	}

	if err != nil {
		return nil, ErrCategoriesFindFailed.Wrap(err)
	}

	category := categories[0]
	return &category, nil
}

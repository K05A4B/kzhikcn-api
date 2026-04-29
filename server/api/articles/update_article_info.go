package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var UpdateArticleInfoHandler = hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload data.EditableArticle) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id")
	})
	if err != nil {
		return err
	}

	err = article.Update(payload)
	if err == data.ErrCategoryNotFound {
		return ErrCategoryNotFound
	}
	if err != nil {
		return ErrUpdateArticleInfoFailed.Wrap(err)
	}

	return nil
})

package articles

import (
	"io"
	"kzhikcn/pkg/assets"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var UpdateArticleRawContentHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Limit(1)
	})

	if err != nil {
		return err
	}

	writer, err := assets.ArticlesRepo.ContentWriter(article.ID.String())
	if err != nil {
		return ErrContentWriteFailed.Wrap(err)
	}

	defer writer.Close()
	_, err = io.Copy(writer, r.Body)
	if err != nil {
		return ErrContentWriteFailed.Wrap(err)
	}

	return nil
})

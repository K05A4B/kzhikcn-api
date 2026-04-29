package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var IncrementArticleLikesHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return err
	}

	likes, err := article.IncrementLikes()
	if err != nil {
		return ErrUpdateArticleLikesFailed.Wrap(err)
	}

	resp.Data = map[string]any{
		"likes": likes,
	}

	return nil
})

var IncrementArticleViewsHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return err
	}

	views, err := article.IncrementViews()
	if err != nil {
		return ErrUpdateArticleViewsFailed.Wrap(err)
	}

	resp.Data = map[string]any{
		"views": views,
	}

	return nil
})

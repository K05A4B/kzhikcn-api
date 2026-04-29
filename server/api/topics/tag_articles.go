package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetArticlesByTagHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	applyExpr, err := httputil.UseExpression(r, queryfilter.WhiteList{
		"id":             nil,
		"title":          nil,
		"views":          nil,
		"likes":          nil,
		"description":    nil,
		"enable_comment": nil,
		"custom_id":      nil,
		"created_at":     queryfilter.TimeValueParser(),
		"update_at":      queryfilter.TimeValueParser(),
	})
	if err != nil {
		return httputil.InvalidExpression(err)
	}

	tag, err := getTagBase(r, func(tx *gorm.DB) *gorm.DB {
		tx = tx.Select("id", "tag_name")
		return tx.Preload("Articles", func(db *gorm.DB) *gorm.DB {

			db = httputil.ApplyPagination(r, 20, 100, db)
			db = db.Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
			db = db.Scopes(data.Adapter(applyExpr))

			return db.Preload("Category")
		})
	})

	if err != nil {
		return err
	}

	resp.Data = tag
	resp.Meta["count"] = len(tag.Articles)

	httputil.SetTotal(resp, data.Article{}, func(tx *gorm.DB) *gorm.DB {
		return tx.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", tag.ID).
			Where("articles.status = ?", data.ARTICLE_STATUS_PUBLISHED).
			Scopes(data.Adapter(applyExpr))
	})

	return nil
})

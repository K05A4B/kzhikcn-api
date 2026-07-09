package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetArticlesByTagHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	claims := authtoken.GetClaims(r.Context())
	wt := queryfilter.WhiteList{
		"id":             nil,
		"title":          nil,
		"views":          nil,
		"likes":          nil,
		"description":    nil,
		"enable_comment": nil,
		"custom_id":      nil,
		"created_at":     queryfilter.TimeValueParser(),
		"updated_at":     queryfilter.TimeValueParser(),
	}

	if claims != nil {
		wt.Add("status")
	}

	applyExpr, err := httputil.UseExpression(r, wt)
	if err != nil {
		return httputil.InvalidExpression(err)
	}

	tag, err := getTagBase(r, func(tx *gorm.DB) *gorm.DB {
		tx = tx.Select("id", "tag_name")
		return tx.Preload("Articles", func(db *gorm.DB) *gorm.DB {

			db = httputil.ApplyPagination(r, 20, 100, db)

			if claims == nil {
				db = db.Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
			}

			db = db.Scopes(data.Adapter(applyExpr))

			return db.Preload("Category").Preload("Tags")
		})
	})

	if err != nil {
		return err
	}

	resp.Data = tag
	resp.Meta["count"] = len(tag.Articles)

	httputil.SetTotal(resp, data.Article{}, func(tx *gorm.DB) *gorm.DB {
		tx = tx.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", tag.ID)
		if claims == nil {
			tx = tx.Where("articles.status = ?", data.ARTICLE_STATUS_PUBLISHED)
		}
		return tx.Scopes(data.Adapter(applyExpr))
	})

	return nil
})

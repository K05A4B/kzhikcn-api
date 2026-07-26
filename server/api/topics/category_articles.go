package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func GetArticlesByCategory(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		claims := authtoken.GetClaims(r.Context())
		wt := queryfilter.WhiteList{
			"id": nil, "title": nil, "views": nil, "likes": nil,
			"description": nil, "enable_comment": nil, "custom_id": nil,
			"created_at": queryfilter.TimeValueParser(),
			"updated_at": queryfilter.TimeValueParser(),
		}
		if claims != nil {
			wt.Add("status")
		}

		applyExpr, err := httputil.UseExpression(r, wt)
		if err != nil {
			return httputil.InvalidExpression(err)
		}

		category, err := appCtx.TopicSvc.GetCategoryByNameOrID(r.Context(), chi.URLParam(r, "category"))
		if err != nil {
			return ErrCategoryNotFound
		}

		categories, err := data.GetCategories(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("id=?", category.ID).Limit(1).
				Select("id", "category_name", "description").
				Preload("Articles", func(db *gorm.DB) *gorm.DB {
					db = httputil.ApplyPagination(r, 20, 100, db)
					db = db.Scopes(data.Adapter(applyExpr))
					if claims == nil {
						db = db.Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
					}
					return db.Preload("Category").Preload("Tags")
				})
			return tx
		})
		if err != nil {
			return ErrCategoriesFindFailed.Wrap(err)
		}
		if len(categories) == 0 {
			return ErrCategoryNotFound
		}

		c := categories[0]
		resp.Data = c
		resp.Meta["count"] = len(c.Articles)

		httputil.SetTotal(resp, data.Article{}, func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("category_id=?", c.ID)
			if claims == nil {
				tx = tx.Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
			}
			return tx
		}, applyExpr)

		return nil
	})
}

package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetArticlesHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	claims := authtoken.GetClaims(r.Context())

	if claims == nil || !claims.IsAdmin {
		return queryArticles(r, resp, articleExprWhiteList(), func(tx *gorm.DB) *gorm.DB {
			// 只允许查询状态为公开的文章信息
			return tx.Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
		})
	}

	return queryArticles(r, resp, articleExprWhiteList().Add("status"), func(tx *gorm.DB) *gorm.DB {
		return tx
	})
})

func queryArticles(r *http.Request, resp *hdl.Response, wt queryfilter.WhiteList, modifier data.QueryModifier) error {
	applyExpr, err := httputil.UseExpression(r, wt)
	if err != nil {
		return httputil.InvalidExpression(err)
	}

	articles, err := data.GetArticles(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("Category").Preload("Tags")
		tx = httputil.ApplyPagination(r, 20, 100, tx)

		return tx
	}, applyExpr, applyArticleOrderBy(r), modifier)

	if err != nil {
		return ErrFindArticleFailed.Wrap(err)
	}

	resp.Data = articles
	resp.Meta["count"] = len(articles)

	httputil.SetTotal(resp, data.Article{}, applyExpr, modifier)

	return nil
}

package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetDeletedArticlesHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	applyExpr, err := httputil.UseExpression(r, articleExprWhiteList().Add("status"))
	if err != nil {
		return err
	}

	onlyDeleted := func(tx *gorm.DB) *gorm.DB {
		tx = tx.Unscoped()

		return tx.Not("deleted_at IS ?", nil)
	}

	articles, err := data.GetArticles(onlyDeleted, applyExpr, applyArticleOrderBy(r), func(tx *gorm.DB) *gorm.DB {
		return httputil.ApplyPagination(r, 20, 100, tx)
	})

	if err != nil {
		return ErrFindArticleFailed.Wrap(err)
	}

	resp.Data = articles
	resp.Meta["count"] = len(articles)

	httputil.SetTotal(resp, data.Article{}, applyExpr, onlyDeleted)
	return nil
})

package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/httputil"
	"kzhikcn/server/service"
	"net/http"
)

func GetDeletedArticles(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		page, limit := httputil.Pagination(r, 20, 100)
		orderBy := httputil.QueryString(r, "orderBy", "publishedAt:desc")

		wt := articleExprWhiteList().Add("status")
		applyExpr, err := httputil.UseExpression(r, wt)
		if err != nil {
			return httputil.InvalidExpression(err)
		}

		articles, total, err := appCtx.ArticleSvc.GetDeletedArticles(r.Context(), service.GetArticlesOptions{
			Page:    page,
			Limit:   limit,
			OrderBy: orderBy,
		}, applyExpr)
		if err != nil {
			return ErrFindArticleFailed.Wrap(err)
		}

		resp.Data = articles
		resp.Meta["total"] = total
		resp.Meta["count"] = len(articles)

		return nil
	})
}

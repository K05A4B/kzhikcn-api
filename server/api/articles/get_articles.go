package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/httputil"
	"kzhikcn/server/service"
	"net/http"
)

func GetArticles(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		claims := authtoken.GetClaims(r.Context())
		page, limit := httputil.Pagination(r, 20, 100)
		orderBy := httputil.QueryString(r, "orderBy", "publishedAt:desc")

		omitStatus := []data.ArticleStatus{}

		wt := articleExprWhiteList()

		// 非管理员用户只能使用此接口查询公开的文章信息
		if claims == nil || !claims.IsAdmin {
			omitStatus = append(omitStatus, data.ARTICLE_STATUS_DRAFT, data.ARTICLE_STATUS_HIDDEN)
		} else {
			// 管理员用户可以在表达式中使用status字段
			wt.Add("status")
		}

		applyExpr, err := httputil.UseExpression(r, wt)
		if err != nil {
			return httputil.InvalidExpression(err)
		}

		articles, total, err := appCtx.ArticleSvc.GetArticles(r.Context(), service.GetArticlesOptions{
			Page:       page,
			Limit:      limit,
			OrderBy:    orderBy,
			OmitStatus: omitStatus,
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

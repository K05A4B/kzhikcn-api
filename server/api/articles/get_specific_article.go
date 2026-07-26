package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// 精确查询文章信息
// GET /api/v1/articles/{article_id}
func SpecificArticle(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		claims := authtoken.GetClaims(r.Context())
		omitStatus := []data.ArticleStatus{}

		if claims == nil || !claims.IsAdmin {
			omitStatus = append(omitStatus, data.ARTICLE_STATUS_DRAFT)
		}

		article, err := appCtx.ArticleSvc.GetArticle(r.Context(), chi.URLParam(r, "article_id"))

		if err == service.ErrArticleNotFound {
			return ErrArticleNotFound
		}

		if err != nil {
			return ErrFindArticleFailed.Wrap(err)
		}

		resp.Data = article

		return nil
	})
}

package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func UpdateArticleInfo(appCtx *app.AppContext) hdl.Handler[service.ArticleUpdateFields] {
	return hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload service.ArticleUpdateFields) error {
		articleID := chi.URLParam(r, "article_id")

		article, err := appCtx.ArticleSvc.Update(r.Context(), articleID, payload)
		if err == service.ErrCategoryNotFound {
			return ErrCategoryNotFound
		}
		if err != nil {
			return ErrUpdateArticleInfoFailed.Wrap(err)
		}

		resp.Data = article
		return nil
	})
}

package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ListArticleAssets(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		list, err := appCtx.ArticleSvc.ListAssets(r.Context(), articleID)
		if err != nil {
			return ErrAssetsListFailed.Wrap(err)
		}

		resp.Data = list
		return nil
	})
}

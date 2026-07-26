package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func DeleteArticleAsset(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")
		assetID := chi.URLParam(r, "asset_id")

		err := appCtx.ArticleSvc.DeleteAsset(r.Context(), articleID, assetID)
		if err != nil {
			return ErrArticleDeleteAssetsFailed.Wrap(err)
		}

		return nil
	})
}

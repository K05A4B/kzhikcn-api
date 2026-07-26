package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/httputil"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetArticleAsset(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")
		assetID := chi.URLParam(r, "asset_id")

		has, err := appCtx.ArticleSvc.HasAsset(r.Context(), articleID, assetID)
		if err != nil {
			return ErrAssetsCheckStatFailed.Wrap(err)
		}
		if !has {
			return ErrAssetsNotFound
		}

		file, err := appCtx.ArticleSvc.OpenAsset(r.Context(), articleID, assetID)
		if err != nil {
			return ErrAssetsOpenFailed.Wrap(err)
		}
		defer file.Close()

		httputil.ApplyCacheControl(hdl.ResponseWriter(r), httputil.CachePublic(), httputil.CacheMaxAge(30*24*time.Hour))

		return hdl.WriteRaw(r, file, "")
	})
}

package articles

import (
	"kzhikcn/pkg/assets"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var GetArticleAssetHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	id := chi.URLParam(r, "article_id")
	assetId := chi.URLParam(r, "asset_id")

	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Limit(1)
	})
	if err != nil {
		return err
	}

	id = article.ID.String()

	hasAssets, err := assets.ArticlesRepo.HasAsset(id, assetId)
	if err != nil {
		return ErrAssetsCheckStatFailed.Wrap(err)
	}

	if !hasAssets {
		return ErrAssetsNotFound
	}

	file, err := assets.ArticlesRepo.OpenAsset(id, assetId)
	if err != nil {
		return ErrAssetsOpenFailed.Wrap(err)
	}

	defer file.Close()

	httputil.ApplyCacheControl(hdl.ResponseWriter(r), httputil.CachePublic(), httputil.CacheMaxAge(30*24*time.Hour))

	return hdl.WriteRaw(r, file, "")
})

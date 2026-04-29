package articles

import (
	"kzhikcn/pkg/assets"
	"kzhikcn/server/common/hdl"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var DeleteArticleAssetHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	assetId := chi.URLParam(r, "asset_id")
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Limit(1)
	})
	if err != nil {
		return err
	}

	err = assets.ArticlesRepo.RemoveAsset(article.ID.String(), assetId)
	if err != nil {
		return ErrArticleDeleteAssetsFailed.Wrap(err)
	}

	return nil
})

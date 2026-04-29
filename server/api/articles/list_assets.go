package articles

import (
	"kzhikcn/pkg/assets"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var ListArticleAssetsHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Limit(1)
	})
	if err != nil {
		return err
	}

	list, err := assets.ArticlesRepo.ListAssets(article.ID.String())
	if err == assets.ErrAssetsDirNotFound {
		resp.Data = []string{}
		return nil
	}

	if err != nil {
		return ErrAssetsListFailed.Wrap(err)
	}

	resp.Data = list

	return nil
})

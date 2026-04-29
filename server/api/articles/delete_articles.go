package articles

import (
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

type BatchDeleteArticlesRequest struct {
	IDs        []string `json:"ids"`
	HardDelete bool     `json:"hardDelete"`
}

var BatchDeleteArticlesHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload BatchDeleteArticlesRequest) error {
		err := data.DeleteArticle(payload.HardDelete, func(tx *gorm.DB) *gorm.DB {
			return tx.Where("id IN ?", payload.IDs)
		})

		if err != nil {
			return ErrDeleteArticleFailed.Wrap(err)
		}

		if payload.HardDelete {
			for _, id := range payload.IDs {
				err = assets.ArticlesRepo.Remove(id)
				if err != nil {
					return ErrArticleCleanAssetsFailed.Wrap(err)
				}
			}
		}

		return nil
	},

	hdl.When(func(payload BatchDeleteArticlesRequest) bool {
		return len(payload.IDs) == 0
	}, hdl.Error(400, "请提供要删除文章的id (ids)", nil, "articles.delete.missing_ids")),
)

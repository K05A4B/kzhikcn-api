package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

type RestoreArticleRequest struct {
	IDs []string `json:"ids"`
}

var RestoreArticleHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload RestoreArticleRequest) error {
		err := data.RestoreArticle(func(tx *gorm.DB) *gorm.DB {
			return tx.Where("id IN ?", payload.IDs)
		})

		if err != nil {
			return ErrRestoreArticleFailed.Wrap(err)
		}
		return nil
	},

	hdl.When(func(payload RestoreArticleRequest) bool {
		return len(payload.IDs) == 0
	}, hdl.Error(400, "请提供要恢复文章的id (ids)", nil, "articles.restore.ids_required")),
)

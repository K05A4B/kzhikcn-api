package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BatchDeleteArticlesRequest struct {
	IDs        []string `json:"ids"`
	HardDelete bool     `json:"hardDelete"`
}

func BatchDeleteArticles(appCtx *app.AppContext) hdl.Handler[BatchDeleteArticlesRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload BatchDeleteArticlesRequest) error {
			for _, id := range payload.IDs {
				if err := appCtx.ArticleSvc.Delete(r.Context(), id, payload.HardDelete); err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						continue
					}
					return ErrDeleteArticleFailed.Wrap(err)
				}
			}
			return nil
		},

		hdl.When(func(payload BatchDeleteArticlesRequest) bool {
			return len(payload.IDs) == 0
		}, hdl.Error(400, "请提供要删除文章的id (ids)", nil, "articles.delete.missing_ids")),
	)
}

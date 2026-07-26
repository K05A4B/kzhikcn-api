package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"
)

type RestoreArticleRequest struct {
	IDs []string `json:"ids"`
}

func RestoreArticle(appCtx *app.AppContext) hdl.Handler[RestoreArticleRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload RestoreArticleRequest) error {
			err := appCtx.ArticleSvc.Restore(r.Context(), payload.IDs)
			if err != nil {
				return ErrRestoreArticleFailed.Wrap(err)
			}
			return nil
		},

		hdl.When(func(payload RestoreArticleRequest) bool {
			return len(payload.IDs) == 0
		}, hdl.Error(400, "请提供要恢复文章的id (ids)", nil, "articles.restore.ids_required")),
	)
}

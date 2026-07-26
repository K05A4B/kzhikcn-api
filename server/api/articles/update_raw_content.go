package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func UpdateArticleRawContent(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		err := appCtx.ArticleSvc.UpdateContent(r.Context(), articleID, r.Body)
		if err != nil {
			return ErrContentWriteFailed.Wrap(err)
		}

		return nil
	})
}

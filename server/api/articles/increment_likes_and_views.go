package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func IncrementArticleLikes(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		likes, err := appCtx.ArticleSvc.IncrementLikes(r.Context(), articleID)
		if err != nil {
			return ErrUpdateArticleLikesFailed.Wrap(err)
		}

		resp.Data = map[string]any{"likes": likes}
		return nil
	})
}

func IncrementArticleViews(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		views, err := appCtx.ArticleSvc.IncrementViews(r.Context(), articleID)
		if err != nil {
			return ErrUpdateArticleViewsFailed.Wrap(err)
		}

		resp.Data = map[string]any{"views": views}
		return nil
	})
}

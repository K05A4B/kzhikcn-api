package articles

import (
	"io"
	"kzhikcn/pkg/assets/article"
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/common/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func GetArticleContent(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		reader, err := appCtx.ArticleSvc.GetContent(r.Context(), articleID)
		if errors.Is(err, article.ErrContentNotFound) {
			return ErrContentNotFound
		}
		if err != nil {
			return ErrContentLoadFailed.Wrap(err)
		}
		defer reader.Close()

		if httputil.Accepts(r, "text/markdown", "text/html", "text/plain") {
			return hdl.WriteRaw(r, reader, "text/plain; charset=utf-8")
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return ErrContentLoadFailed.Wrap(err)
		}

		resp.Data = string(data)
		return nil
	})
}

func GetArticleRenderedContent(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		content, err := appCtx.ArticleSvc.GetRenderedContent(r.Context(), articleID)
		if errors.Is(err, article.ErrContentNotFound) {
			return ErrContentNotFound
		}
		if err != nil {
			return ErrContentRenderFailed.Wrap(err)
		}

		if httputil.Accepts(r, "text/html") {
			return hdl.WriteRawData(r, content, "text/html; charset=utf-8")
		}

		resp.Data = string(content)
		return nil
	})
}

package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"kzhikcn/server/service"
	"net/http"
)

func CreateArticle(appCtx *app.AppContext) hdl.Handler[service.ArticleUpdateFields] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload service.ArticleUpdateFields) error {
			article, err := appCtx.ArticleSvc.Create(r.Context(), &payload)
			if err == service.ErrCategoryNotFound {
				return ErrCategoryNotFound
			}
			if err != nil {
				return ErrCreateArticleFailed.Wrap(err)
			}

			resp.Data = article
			return nil
		},

		hdl.MissingFields(func(payload service.ArticleUpdateFields) []string {
			missing := []string{}
			if payload.Title == nil || *payload.Title == "" {
				missing = append(missing, "title")
			}
			return missing
		}),
	)
}

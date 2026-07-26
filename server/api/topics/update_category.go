package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func UpdateCategory(appCtx *app.AppContext) hdl.Handler[data.EditableCategory] {
	return hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload data.EditableCategory) error {
		name := chi.URLParam(r, "category")

		category, err := appCtx.TopicSvc.GetCategoryByNameOrID(r.Context(), name)
		if err != nil {
			return ErrCategoryNotFound
		}

		err = appCtx.TopicSvc.UpdateCategory(r.Context(), category.ID, payload)
		if err != nil {
			return ErrCategoryUpdateFailed.Wrap(err)
		}

		return nil
	})
}

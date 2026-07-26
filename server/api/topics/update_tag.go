package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func UpdateTag(appCtx *app.AppContext) hdl.Handler[data.EditableTag] {
	return hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload data.EditableTag) error {
		name := chi.URLParam(r, "tag")

		tag, err := appCtx.TopicSvc.GetTagByNameOrID(r.Context(), name)
		if err != nil {
			return ErrTagNotFound
		}

		err = appCtx.TopicSvc.UpdateTag(r.Context(), tag.ID, payload)
		if err != nil {
			return ErrTagUpdateFailed.Wrap(err)
		}

		return nil
	})
}

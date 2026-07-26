package topics

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"
)

func DeleteCategory(appCtx *app.AppContext) hdl.Handler[BatchDeleteTopicsRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload BatchDeleteTopicsRequest) error {
			err := appCtx.TopicSvc.DeleteCategories(r.Context(), payload.IDs)
			if err != nil {
				return ErrCategoryDeleteFailed.Wrap(err)
			}
			return nil
		},

		hdl.When(func(payload BatchDeleteTopicsRequest) bool {
			return len(payload.IDs) == 0
		}, ErrCategoryDeleteIdIsRequired),
	)
}

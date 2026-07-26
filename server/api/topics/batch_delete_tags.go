package topics

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"
)

type BatchDeleteTopicsRequest struct {
	IDs []uint `json:"ids"`
}

func BatchDeleteTags(appCtx *app.AppContext) hdl.Handler[BatchDeleteTopicsRequest] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload BatchDeleteTopicsRequest) error {
			err := appCtx.TopicSvc.DeleteTags(r.Context(), payload.IDs)
			if err != nil {
				return ErrTagDeleteFailed
			}
			return nil
		},

		hdl.When(func(payload BatchDeleteTopicsRequest) bool {
			return len(payload.IDs) == 0
		}, ErrTagDeleteIdIsRequired),
	)
}

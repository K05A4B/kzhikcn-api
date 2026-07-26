package topics

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/app"
	"kzhikcn/server/common/httputil"
	"net/http"
)

func GetTags(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		applyExpr, err := httputil.UseExpression(r, queryfilter.NewWhiteList().Add("id", "tag_name"))
		if err != nil {
			return httputil.InvalidExpression(err)
		}

		page, limit := httputil.Pagination(r, 50, 200)

		tags, total, err := appCtx.TopicSvc.GetTags(r.Context(), page, limit, applyExpr)
		if err != nil {
			return ErrTagsFindFailed.Wrap(err)
		}

		resp.Data = tags
		resp.Meta["count"] = len(tags)
		resp.Meta["total"] = total
		return nil
	})
}

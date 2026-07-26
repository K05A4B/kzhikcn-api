package topics

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/app"
	"kzhikcn/server/common/httputil"
	"net/http"
)

func GetCategories(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		applyExpr, err := httputil.UseExpression(r, queryfilter.NewWhiteList().Add("id", "category_name", "description"))
		if err != nil {
			return httputil.InvalidExpression(err)
		}

		page, limit := httputil.Pagination(r, 20, 200)

		categories, total, err := appCtx.TopicSvc.GetCategories(r.Context(), page, limit, applyExpr)
		if err != nil {
			return ErrCategoriesFindFailed.Wrap(err)
		}

		resp.Meta["count"] = len(categories)
		resp.Meta["total"] = total
		resp.Data = categories
		return nil
	})
}

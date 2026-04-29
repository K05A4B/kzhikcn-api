package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetCategoriesHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	applyExpr, err := httputil.UseExpression(r, queryfilter.NewWhiteList().Add("id", "category_name", "description"))
	if err != nil {
		return httputil.InvalidExpression(err)
	}

	categories, err := data.GetCategories(func(tx *gorm.DB) *gorm.DB {
		tx = httputil.ApplyPagination(r, 20, 200, tx)
		return tx
	}, applyExpr)

	if err != nil {
		return ErrCategoriesFindFailed.Wrap(err)
	}

	resp.Meta["count"] = len(categories)
	resp.Data = categories

	httputil.SetTotal(resp, data.Category{}, applyExpr)
	return nil
})

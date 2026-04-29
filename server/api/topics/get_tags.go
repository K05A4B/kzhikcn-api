package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetTagsHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	applyExpr, err := httputil.UseExpression(r, queryfilter.NewWhiteList().Add("id", "tag_name"))
	if err != nil {
		return httputil.InvalidExpression(err)
	}

	tags, err := data.GetTags(func(tx *gorm.DB) *gorm.DB {
		tx = httputil.ApplyPagination(r, 50, 200, tx)
		return tx
	}, applyExpr)

	if err != nil {
		return ErrTagsFindFailed.Wrap(err)
	}

	resp.Data = tags
	resp.Meta["count"] = len(tags)

	httputil.SetTotal(resp, data.Tag{}, applyExpr)

	return nil
})

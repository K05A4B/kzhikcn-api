package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/app"
	"kzhikcn/server/service"
	"net/http"

	"github.com/pkg/errors"
)

func CreateCategory(appCtx *app.AppContext) hdl.Handler[data.EditableCategory] {
	return hdl.NewHandler(
		func(r *http.Request, resp *hdl.Response, payload data.EditableCategory) error {
			category, err := appCtx.TopicSvc.CreateCategory(r.Context(), payload)
			if errors.Is(err, service.ErrCategoryExist) {
				return ErrCategoryIsExist
			}
			if err != nil {
				return ErrCategoryCreateFailed
			}

			resp.Data = category
			return nil
		},

		hdl.When(func(payload data.EditableCategory) bool {
			return utils.IsEmptyString(payload.CategoryName)
		}, ErrCategoryCreateCategoryNameIsRequired),
	)
}

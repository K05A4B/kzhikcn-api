package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

var CreateArticleHandler = hdl.NewHandler(
	func(r *http.Request, resp *hdl.Response, payload data.EditableArticle) error {
		omittedFields := []string{}

		if payload.EnableComment == nil {
			omittedFields = append(omittedFields, "enable_comment")
		}

		if utils.IsEmptyString(payload.Description) {
			omittedFields = append(omittedFields, "description")
		}

		if utils.IsEmptyString(payload.CoverImage) {
			omittedFields = append(omittedFields, "cover_image")
		}

		article, err := data.CreateArticle(payload, func(tx *gorm.DB) *gorm.DB {
			return tx.Omit(omittedFields...)
		})

		if err == data.ErrCategoryNotFound {
			return ErrCategoryNotFound
		}

		if err != nil {
			return ErrCreateArticleFailed.Wrap(err)
		}

		resp.Data = article

		return nil
	},

	hdl.MissingFields(func(payload data.EditableArticle) []string {
		missing := []string{}

		if utils.IsEmptyString(payload.Title) {
			missing = append(missing, "title")
		}

		return missing
	}),
)

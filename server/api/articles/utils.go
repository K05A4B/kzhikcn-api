package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

func articleExprWhiteList() queryfilter.WhiteList {
	return queryfilter.WhiteList{
		"id":             nil,
		"title":          nil,
		"views":          nil,
		"likes":          nil,
		"description":    nil,
		"enable_comment": nil,
		"custom_id":      nil,
		"created_at":     queryfilter.TimeValueParser(),
		"update_at":      queryfilter.TimeValueParser(),
		"published_at":   queryfilter.TimeValueParser(),
	}
}

func applyArticleOrderBy(r *http.Request) data.QueryModifier {
	return func(tx *gorm.DB) *gorm.DB {
		return httputil.ApplyOrderBy(r, "publishedAt:desc", tx, httputil.ExtendOrderByWithDesc(map[string]string{
			"publishedAt": "published_at",
			"createdAt":   "created_at",
			"updatedAt":   "updated_at",
			"likes":       "likes",
			"views":       "views",
		}))
	}
}

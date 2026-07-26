package articles

import (
	"kzhikcn/pkg/queryfilter"
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

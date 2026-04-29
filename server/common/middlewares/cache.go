package middlewares

import (
	"kzhikcn/server/common/httputil"
	"net/http"
)

func CacheControl(cds ...httputil.CacheDirective) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httputil.ApplyCacheControl(w, cds...)

			next.ServeHTTP(w, r)
		})
	}
}

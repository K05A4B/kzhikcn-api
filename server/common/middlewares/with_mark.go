package middlewares

import (
	"kzhikcn/server/common/httputil"
	"net/http"
)

func WithMarks(marks ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, httputil.WithMarks(r, marks...))
		})
	}
}

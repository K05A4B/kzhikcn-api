package middlewares

import (
	"net/http"
)

func SetHeader(fn func(http.Header)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w.Header())

			next.ServeHTTP(w, r)
		})
	}
}

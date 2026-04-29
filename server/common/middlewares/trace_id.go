package middlewares

import (
	"kzhikcn/pkg/traceid"
	"net/http"
)

func WithTraceID(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := traceid.New(8)

		handler.ServeHTTP(w, r.WithContext(traceid.WithTraceID(r.Context(), *traceID)))
	})
}

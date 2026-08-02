package middlewares

import (
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/traceid"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := traceid.GetTraceID(r.Context())
		reqLogger := log.WithTraceID(traceID.String()).With("type", "access_log")
		reqTime := time.Now()

		w = &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
		}

		reqLogger.
			With("ip", r.RemoteAddr).
			Infof("[RequestLog] %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		duration := time.Since(reqTime)

		reqLogger.
			With("process_duration", duration.String()).
			Infof("[ResponseLog] %s %s => %d", r.Method, r.URL.Path, w.(*responseWriter).statusCode)
	})
}

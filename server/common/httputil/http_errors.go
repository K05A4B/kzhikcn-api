package httputil

import (
	"fmt"
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/traceid"
	"net/http"
	"runtime"
)

var (
	errInternalServerError = "internet server error"
	errDatabaseOperation   = "database operation failed"
)

func HttpError(status int, err error, w http.ResponseWriter, r *http.Request, skip int, extends ...string) {
	text := http.StatusText(status)
	traceID := traceid.GetTraceID(r.Context()).String()
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		log.WithTraceID(traceID).
			With("file", file).
			With("line", line).
			With("status", status).
			Error(err)
	}

	fmt.Fprintf(w, "<h1 align=center>%d %s</h1><p align=center>TraceID: %s</p>", status, text, traceID)
	for _, item := range extends {
		fmt.Fprintf(w, "<p align=center>%s</p>", item)
	}
	fmt.Fprintf(w, "<hr/><p align=center>powered by %s | %s</p>", appinfo.CurrentInfo.Name, appinfo.CurrentInfo.Copyright)
}

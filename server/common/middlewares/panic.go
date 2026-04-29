package middlewares

import (
	"fmt"
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/traceid"
	"kzhikcn/server/common/httputil"
	"net/http"
	"runtime"

	"github.com/pkg/errors"
)

func Recover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recoverErr := recover()
			if recoverErr != nil {
				traceID := traceid.GetTraceID(r.Context()).String()
				stackBuf := make([]byte, 1024*8)
				n := runtime.Stack(stackBuf, false)

				err := errors.Wrapf(errors.New(fmt.Sprint(recoverErr)), "[PanicCapturer] Panic recovered")
				log.WithTraceID(traceID).Errorf("%s\nstacks:\n%s", err, stackBuf[:n])

				httputil.HttpError(500, err, w, r, 3)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

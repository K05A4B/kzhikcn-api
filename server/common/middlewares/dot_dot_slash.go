package middlewares

import (
	"kzhikcn/pkg/log"
	"kzhikcn/server/common/hdl"
	"net/http"
	"net/url"
	"strings"
)

func DotDotSlash(h http.Handler) http.Handler {
	return hdl.Middleware(func(w http.ResponseWriter, r *http.Request, meta map[string]any) error {
		path := r.URL.Path
		decodedPath, err := url.PathUnescape(path)
		if err != nil {
			return hdl.Error(400, "Bad Request", err, "system.bad_request")
		}

		if containsDotDot(decodedPath) {
			meta["ip"] = r.RemoteAddr
			meta["tips"] = "Your behavior has been logged."
			log.Warnf("`%s` Attempted path traversal", r.RemoteAddr)
			return hdl.Error(403, "Forbidden", nil, "system.forbidden")
		}

		h.ServeHTTP(w, r)
		return nil
	})
}

func containsDotDot(path string) bool {
	decoded, err := url.PathUnescape(path)
	if err != nil {
		return true
	}

	if strings.Contains(decoded, "../") {
		return true
	}

	if decoded != path {
		return containsDotDot(decoded)
	}

	return false
}

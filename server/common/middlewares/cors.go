package middlewares

import (
	"kzhikcn/pkg/config"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func CORS(conf config.CORSConf) func(http.Handler) http.Handler {
	if !conf.Enable {
		return func(next http.Handler) http.Handler { return next }
	}

	allowedMethods := conf.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}

	allowedHeaders := conf.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type", "Authorization", "X-Requested-With"}
	}

	allowAllOrigins := slices.Contains(conf.AllowedOrigins, "*")

	maxAge := ""
	if conf.MaxAge > 0 {
		maxAge = strconv.Itoa(int(time.Duration(conf.MaxAge).Seconds()))
	}

	methodsStr := strings.Join(allowedMethods, ", ")
	headersStr := strings.Join(allowedHeaders, ", ")
	exposeStr := strings.Join(conf.ExposedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if slices.Contains(conf.AllowedOrigins, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			if conf.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", methodsStr)
				w.Header().Set("Access-Control-Allow-Headers", headersStr)
				if exposeStr != "" {
					w.Header().Set("Access-Control-Expose-Headers", exposeStr)
				}
				if maxAge != "" {
					w.Header().Set("Access-Control-Max-Age", maxAge)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if exposeStr != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposeStr)
			}

			next.ServeHTTP(w, r)
		})
	}
}

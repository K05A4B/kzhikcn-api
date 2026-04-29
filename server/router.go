package server

import (
	"kzhikcn/server/api"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"kzhikcn/server/common/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httputil.HttpError(404, nil, w, r, 0)
	})

	// 从header中获取真实IP
	// 从前到后依次是：
	// 1. True-Client-IP
	// 2. X-Real-IP
	// 3. X-Forwarded-For
	r.Use(chiMiddleware.RealIP)

	// 为每个请求添加traceID
	r.Use(middlewares.WithTraceID)

	r.Use(middlewares.Recover)
	r.Use(middlewares.DotDotSlash)

	r.Mount("/api/", api.Router())

	r.Group(func(r chi.Router) {
		r.Use(middlewares.BearerParse)
		r.Use(middlewares.HttpRate())

		r.Get("/sitemap.xml", hdl.New(api.SitemapDotXMLHandler))
		r.Get("/rss.xml", hdl.New(api.RssDotXMLHandler))
	})

	return r
}

package server

import (
	"kzhikcn/server/api"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.NotFound(hdl.New(NotFoundHandler))

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

var NotFoundHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	return hdl.Error(http.StatusNotFound, "not found", nil, "system.not_found")
})

package router

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/api"
	"kzhikcn/server/app"
	"kzhikcn/server/common/httputil"
	"kzhikcn/server/common/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func apiRouter(app *app.App) chi.Router {
	router := chi.NewRouter()

	// 解析Token(如果存在Token)并将Claims放入Context中
	router.Use(middlewares.BearerParse)
	router.Use(middlewares.HttpRate(app.Config))
	router.Use(middlewares.CacheControl(httputil.NoCache()))

	router.Mount("/v1", routerVersion1(app))

	return router
}

func NewRouter(app *app.App) chi.Router {
	r := chi.NewRouter()

	r.NotFound(hdl.New(NotFoundHandler))
	r.MethodNotAllowed(hdl.New(MethodNotAllowedHandler))

	// 为每个请求添加traceID
	r.Use(middlewares.WithTraceID)

	r.Use(middlewares.CORS(app.Config.CORS))

	// 从header中获取真实IP
	// 从前到后依次是：
	// 1. True-Client-IP
	// 2. X-Real-IP
	// 3. X-Forwarded-For
	r.Use(chiMiddleware.RealIP)

	r.Use(middlewares.Recover)
	r.Use(middlewares.DotDotSlash)
	r.Use(middlewares.AccessLog)

	r.Mount("/api/", apiRouter(app))

	r.Group(func(r chi.Router) {
		r.Use(middlewares.BearerParse)
		r.Use(middlewares.HttpRate(app.Config))
		r.Get("/sitemap.xml", hdl.New(api.SitemapDotXMLHandler))
		r.Get("/rss.xml", hdl.New(api.RssDotXMLHandler))
	})

	return r
}

var NotFoundHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	return hdl.Error(http.StatusNotFound, "not found", nil, "system.not_found")
})

var MethodNotAllowedHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	return hdl.Error(http.StatusMethodNotAllowed, "method not allowed", nil, "system.method_not_allowed")
})

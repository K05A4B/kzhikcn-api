package api

import (
	"kzhikcn/server/common/httputil"
	"kzhikcn/server/common/middlewares"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	router := chi.NewRouter()

	// 解析Token(如果存在Token)并将Claims放入Context中
	router.Use(middlewares.BearerParse)
	router.Use(middlewares.HttpRate())
	router.Use(middlewares.CacheControl(httputil.NoCache()))

	router.Mount("/v1", Version1())

	return router
}

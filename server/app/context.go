package app

import (
	"context"
	"kzhikcn/pkg/config"
	"kzhikcn/server/service"
)

type AppContext struct {
	context.Context

	app *App

	ArticleSvc *service.ArticleService
	AuthSvc    *service.AuthService
	TopicSvc   *service.TopicService
	AdminSvc   *service.AdminService
}

func (aCtx *AppContext) Config() config.Config {
	return aCtx.app.GetConfig()
}

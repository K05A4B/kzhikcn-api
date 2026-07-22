package service

import (
	"context"
	"kzhikcn/pkg/assets/article"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/authtoken"
)

type ServiceContext struct {
	Conf *config.Config
	Repo article.Repository
}

type ArticleHooks struct {
	BeforeCreate  []func(ctx context.Context, article *data.Article) error
	AfterCreate   []func(ctx context.Context, article *data.Article) error
	BeforePublish []func(ctx context.Context, article *data.Article) error
	AfterPublish  []func(ctx context.Context, article *data.Article) error
	BeforeDelete  []func(ctx context.Context, articleID string, isHard bool) error
	AfterDelete   []func(ctx context.Context, articleID string, isHard bool) error
}

type AuthHooks struct {
	OnLoginSuccess []func(ctx context.Context, admin *data.Admin) error
	OnLoginFailed  []func(ctx context.Context, username, reason string) error
	OnMFARequired  []func(ctx context.Context, admin *data.Admin) error
	OnLogout       []func(ctx context.Context, claims *authtoken.TokenClaims) error
}

func NewServiceContext(conf *config.Config, repo article.Repository) *ServiceContext {
	return &ServiceContext{Conf: conf, Repo: repo}
}

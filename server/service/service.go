package service

import (
	"kzhikcn/pkg/assets/article"
	"kzhikcn/pkg/config"

	"gorm.io/gorm"
)

var ArticleOrderByMapping = map[string]string{
	"publishedAt":      "published_at",
	"publishedAt:desc": "published_at DESC",
	"createdAt":        "created_at",
	"createdAt:desc":   "created_at DESC",
	"updatedAt":        "updated_at",
	"updatedAt:desc":   "updated_at DESC",
	"likes":            "likes",
	"likes:desc":       "likes DESC",
	"views":            "views",
	"views:desc":       "views DESC",
}

type ServiceContext struct {
	Conf *config.Config
	Repo article.Repository
	DB   *gorm.DB
}

func NewServiceContext(conf *config.Config, repo article.Repository, db *gorm.DB) *ServiceContext {
	return &ServiceContext{Conf: conf, Repo: repo, DB: db}
}

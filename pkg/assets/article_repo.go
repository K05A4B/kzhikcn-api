package assets

import (
	"kzhikcn/pkg/assets/article"
	"kzhikcn/pkg/config"
	"log"
)

var (
	ArticlesRepo article.Repository

	ErrContentNotFound   = article.ErrContentNotFound
	ErrAssetsDirNotFound = article.ErrAssetsDirNotFound
)

// Init 初始化文章存储。由 App.Initialize() 显式调用，替代原先的隐式 init()。
func Init(c *config.Config) article.Repository {
	if c.Storage.Provider != "local" {
		log.Fatal("unsupported storage provider: ", c.Storage.Provider)
	}

	ArticlesRepo = &article.LocalRepository{
		AssetsDir: "assets",
		BasePath:  c.Storage.Articles.BasePath,
	}

	return ArticlesRepo
}

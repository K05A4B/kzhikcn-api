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

func init() {
	config.HookLoaded(func(c *config.Config) {
		if c.Storage.Provider != "local" {
			log.Fatal("unsupported storage provider: ", c.Storage.Provider)
		}

		ArticlesRepo = &article.LocalRepository{
			AssetsDir: "assets",
			BasePath:  c.Storage.Articles.BasePath,
		}
	})
}

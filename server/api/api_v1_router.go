package api

import (
	"kzhikcn/server/api/admin"
	"kzhikcn/server/api/articles"
	"kzhikcn/server/api/auth"
	"kzhikcn/server/api/topics"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/middlewares"

	"github.com/go-chi/chi/v5"
)

func Version1() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/auth/login", hdl.New(auth.LoginHandler))
		r.Post("/auth/mfa/totp", hdl.New(auth.VerifyTOTPHandler))

		r.Get("/articles", hdl.New(articles.GetArticlesHandler))                                    // 查询文章列表
		r.Get("/articles/{article_id}", hdl.New(articles.SpecificArticleHandler))                   // 查询单篇文章
		r.Post("/articles/{article_id}/view", hdl.New(articles.IncrementArticleViewsHandler))       // 增加文章浏览量
		r.Post("/articles/{article_id}/like", hdl.New(articles.IncrementArticleLikesHandler))       // 增加文章点赞量
		r.Get("/articles/{article_id}/content", hdl.New(articles.GetArticleRenderedContentHandler)) // 获取渲染后的内容

		r.Get("/articles/{article_id}/assets/{asset_id}", hdl.New(articles.GetArticleAssetHandler)) // 获取资源

		r.Get("/categories/{category}/articles", hdl.New(topics.GetArticlesByCategoryHandler)) // 查询某个分类的相关文章
		r.Get("/categories", hdl.New(topics.GetCategoriesHandler))                             // 获取分类标签列表

		r.Get("/tags/{tag}/articles", hdl.New(topics.GetArticlesByTagHandler)) // 查询拥有某个标签的文章
		r.Get("/tags", hdl.New(topics.GetTagsHandler))                         // 获取标签列表
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.BearerAuth)
		r.Post("/auth/logout", hdl.New(auth.LogoutHandler))

		r.Get("/users/me", hdl.New(admin.AdminInfoHandler))
		r.Put("/users/me/mfa/{action:(en|dis)able}", hdl.New(admin.SetMFAStatusHandler))
		r.Post("/users/me/mfa/totp-secret", hdl.New(admin.GenerateTOTPSecretHandler)) // 为当前用户生成totp的密钥
		r.Patch("/users/me", hdl.New(admin.UpdateAdminInfoHandler))
		r.Put("/users/me/password", hdl.New(admin.ChangePasswordHandler))
	})

	// 管理员仅管理员调用的接口
	r.Mount("/admin", v1Admin())

	return r
}

func v1Admin() chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.BearerAuth)

	// 文章管理
	r.Get("/articles", hdl.New(articles.AdminGetArticlesHandler))                  // 查询文章列表
	r.Get("/articles/trash-bin", hdl.New(articles.GetDeletedArticlesHandler))      // 查找所有被软删除的文章
	r.Post("/articles/trash-bin/restore", hdl.New(articles.RestoreArticleHandler)) // 批量恢复软删除的文章

	r.Post("/articles", hdl.New(articles.CreateArticleHandler))                      // 创建文章
	r.Delete("/articles/batch-delete", hdl.New(articles.BatchDeleteArticlesHandler)) // 批量删除文章
	r.Patch("/articles/{article_id}", hdl.New(articles.UpdateArticleInfoHandler))    // 更新文章信息

	// 文章内容（子资源）
	r.Get("/articles/{article_id}/raw-content", hdl.New(articles.GetArticleContentHandler))       // 获取文章原始内容
	r.Put("/articles/{article_id}/raw-content", hdl.New(articles.UpdateArticleRawContentHandler)) // 更新更新原始内容

	// 文章资源（子资源）
	r.Get("/articles/{article_id}/assets", hdl.New(articles.ListArticleAssetsHandler))                // 获取已上传的资源列表
	r.Post("/articles/{article_id}/assets", hdl.New(articles.UploadArticleAssetHandler))              // 上传资源
	r.Delete("/articles/{article_id}/assets/{asset_id}", hdl.New(articles.DeleteArticleAssetHandler)) // 删除资源

	// 话题管理
	r.Post("/tags/batch-delete", hdl.New(topics.BatchDeleteTagsHandler)) //批量删除标签
	r.Patch("/tags/{tag}", hdl.New(topics.UpdateTagHandler))
	r.Post("/categories", hdl.New(topics.CreateCategoryHandler))
	r.Patch("/categories/{category}", hdl.New(topics.UpdateCategoryHandler))
	r.Delete("/categories/{category}", hdl.New(topics.DeleteCategoryHandler))
	return r
}

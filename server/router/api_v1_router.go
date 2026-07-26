package router

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/api"
	"kzhikcn/server/api/admin"
	"kzhikcn/server/api/articles"
	"kzhikcn/server/api/auth"
	"kzhikcn/server/api/topics"
	"kzhikcn/server/app"
	"kzhikcn/server/common/middlewares"

	"github.com/go-chi/chi/v5"
)

func routerVersion1(app *app.App) chi.Router {
	r := chi.NewRouter()

	// 公共接口
	r.Group(func(r chi.Router) {
		r.Handle("/ping", hdl.New(api.PingHandler))

		r.Post("/auth/login", hdl.New(auth.Login(app.Context)))
		r.Post("/auth/mfa/totp", hdl.New(auth.VerifyTOTP(app.Context)))

		r.Get("/articles", hdl.New(articles.GetArticles(app.Context)))                                    // 查询文章列表
		r.Get("/articles/{article_id}", hdl.New(articles.SpecificArticle(app.Context)))                   // 查询单篇文章
		r.Post("/articles/{article_id}/view", hdl.New(articles.IncrementArticleViews(app.Context)))       // 增加文章浏览量
		r.Post("/articles/{article_id}/like", hdl.New(articles.IncrementArticleLikes(app.Context)))       // 增加文章点赞量
		r.Get("/articles/{article_id}/content", hdl.New(articles.GetArticleRenderedContent(app.Context))) // 获取渲染后的内容
		r.Get("/articles/{article_id}/assets/{asset_id}", hdl.New(articles.GetArticleAsset(app.Context))) // 获取资源

		r.Get("/categories/{category}/articles", hdl.New(topics.GetArticlesByCategory(app.Context))) // 查询某个分类的相关文章
		r.Get("/categories", hdl.New(topics.GetCategories(app.Context)))                             // 获取分类标签列表

		r.Get("/tags/{tag}/articles", hdl.New(topics.GetArticlesByTag(app.Context))) // 查询拥有某个标签的文章
		r.Get("/tags", hdl.New(topics.GetTags(app.Context)))                         // 获取标签列表
	})

	// 需要认证的接口
	r.Group(func(r chi.Router) {
		r.Use(middlewares.BearerAuth)
		r.Post("/auth/logout", hdl.New(auth.Logout(app.Context)))

		r.Get("/users/me", hdl.New(admin.AdminInfo(app.Context)))
		r.Put("/users/me/mfa/{action:(en|dis)able}", hdl.New(admin.SetMFAStatus(app.Context)))
		r.Post("/users/me/mfa/totp-secret", hdl.New(admin.GenerateTOTPSecret(app.Context)))
		r.Patch("/users/me", hdl.New(admin.UpdateAdminInfo(app.Context)))
		r.Put("/users/me/password", hdl.New(admin.ChangePassword(app.Context)))

		// 话题管理
		r.Delete("/tags/batch-delete", hdl.New(topics.BatchDeleteTags(app.Context))) //批量删除标签
		r.Patch("/tags/{tag}", hdl.New(topics.UpdateTag(app.Context)))
		r.Post("/categories", hdl.New(topics.CreateCategory(app.Context)))
		r.Patch("/categories/{category}", hdl.New(topics.UpdateCategory(app.Context)))
		r.Delete("/categories/batch-delete", hdl.New(topics.DeleteCategory(app.Context)))

		// 文章管理
		r.Get("/articles/trash-bin", hdl.New(articles.GetDeletedArticles(app.Context)))      // 查找所有被软删除的文章
		r.Post("/articles/trash-bin/restore", hdl.New(articles.RestoreArticle(app.Context))) // 批量恢复软删除的文章

		r.Post("/articles", hdl.New(articles.CreateArticle(app.Context)))                      // 创建文章
		r.Delete("/articles/batch-delete", hdl.New(articles.BatchDeleteArticles(app.Context))) // 批量删除文章
		r.Patch("/articles/{article_id}", hdl.New(articles.UpdateArticleInfo(app.Context)))    // 更新文章信息

		// 文章内容（子资源）
		r.Get("/articles/{article_id}/raw-content", hdl.New(articles.GetArticleContent(app.Context)))       // 获取文章原始内容
		r.Put("/articles/{article_id}/raw-content", hdl.New(articles.UpdateArticleRawContent(app.Context))) // 更新文章原始内容

		// 文章资源（子资源）
		r.Get("/articles/{article_id}/assets", hdl.New(articles.ListArticleAssets(app.Context)))                // 获取已上传的资源列表
		r.Post("/articles/{article_id}/assets", hdl.New(articles.UploadArticleAsset(app.Context)))              // 上传资源
		r.Delete("/articles/{article_id}/assets/{asset_id}", hdl.New(articles.DeleteArticleAsset(app.Context))) // 删除资源
	})

	return r
}

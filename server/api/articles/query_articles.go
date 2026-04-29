package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

// 获取文章列表
// GET /api/v1/articles
//
// 认证要求:
//   - 无需认证
//
// 请求类型:
//   - Content-Type: application/json
//
// 请求参数: 无

// 查询参数：
//
//	page     uint    页码，最小值 1，默认值 1
//	limit    uint    每页条数，范围 1-100，默认值 20
//	orderBy  string	 排序字段，支持 :desc 后缀表示倒序
//	  -可选值：publishedAt, createdAt, updatedAt, likes, views
//	expr     string 过滤表达式
//		- 过滤表达式允许字段：
//			id, title, views, likes, description, enable_comment, custom_id, created_at, updated_at
//
// 响应数据:
//
//	data: [
//		- id (string):  文章ID
//		-	createdAt (string): 文章创建时间 (ISO8601)
//		-	updatedAt (string): 文章更新时间 (ISO8601)
//		-	publishedAt (string): 文章发布时间 (ISO8601)
//		-	customID (string): 文章自定义ID
//		-	title (string): 文章标题
//		-	views (int): 文章浏览量
//		-	likes (int): 文章点赞量
//		-	categoryID (int | null): 分类ID (没有则为null)
//		-	category (object): 分类详细信息
//				-	id (int): 分类ID (没有则为0)
//				- categoryName (string): 分类名称
//		-	tags: [
//				- id (int): 标签ID
//				- tagName (string): 标签名
//			],
//		-	status (string): 文章状态
//			- 取值范围: published, hidden, draft
//		-	description (string): 文章描述
//		-	enableComment (bool): 是否启用评论
//	]
//
//	meta:
//		- count (number): 查询出多少条文章信息
//		- total (number): 符合条件的文章总数
var GetArticlesHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	return queryArticles(r, resp, articleExprWhiteList(), func(tx *gorm.DB) *gorm.DB {
		// 只允许查询状态为公开的文章信息
		return tx.Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
	})
})

// 获取文章列表
// GET /api/v1/admin/articles
//
// 认证要求:
//   - 需要提供Token
//
// 请求类型:
//   - Content-Type: application/json
//
// 请求参数: 无

// 查询参数：
//
//	page     uint    页码，最小值 1，默认值 1
//	limit    uint    每页条数，范围 1-100，默认值 20
//	orderBy  string	 排序字段，支持 :desc 后缀表示倒序
//	  -可选值：publishedAt, createdAt, updatedAt, likes, views
//	expr     string 过滤表达式
//		- 过滤表达式允许字段：
//			id, title, views, likes, description, enable_comment, custom_id, created_at, updated_at, status
//	onlyDeleted bool	只展示被标记为删除的文章
//
// 响应数据:
//
//	data: [
//		- id (string):  文章ID
//		-	createdAt (string): 文章创建时间 (ISO8601)
//		-	updatedAt (string): 文章更新时间 (ISO8601)
//		-	publishedAt (string): 文章发布时间 (ISO8601)
//		-	customID (string): 文章自定义ID
//		-	title (string): 文章标题
//		-	views (int): 文章浏览量
//		-	likes (int): 文章点赞量
//		-	categoryID (int | null): 分类ID (没有则为null)
//		-	category (object): 分类详细信息
//				-	id (int): 分类ID (没有则为0)
//				- categoryName (string): 分类名称
//		-	tags: [
//				- id (int): 标签ID
//				- tagName (string): 标签名
//			],
//		-	status (string): 文章状态
//			- 取值范围: published, hidden, draft
//		-	description (string): 文章描述
//		-	enableComment (bool): 是否启用评论
//	]
var AdminGetArticlesHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	return queryArticles(r, resp, articleExprWhiteList().Add("status"), func(tx *gorm.DB) *gorm.DB {
		if httputil.QueryBool(r, "onlyDeleted", false) {
			tx = tx.Unscoped().Where("deleted_at IS NOT ?", nil)
		}
		return tx
	})
})

func queryArticles(r *http.Request, resp *hdl.Response, wt queryfilter.WhiteList, modifier data.QueryModifier) error {
	applyExpr, err := httputil.UseExpression(r, wt)
	if err != nil {
		return httputil.InvalidExpression(err)
	}

	articles, err := data.GetArticles(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("Category").Preload("Tags")
		tx = httputil.ApplyPagination(r, 20, 100, tx)

		return tx
	}, applyExpr, applyArticleOrderBy(r), modifier)

	if err != nil {
		return ErrFindArticleFailed.Wrap(err)
	}

	resp.Data = articles
	resp.Meta["count"] = len(articles)

	httputil.SetTotal(resp, data.Article{}, applyExpr, modifier)

	return nil
}

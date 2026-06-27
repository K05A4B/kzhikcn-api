package articles

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"

	"gorm.io/gorm"
)

// 获取文章列表
// GET /api/v1/articles/{article_id}
//
// 认证要求:
//   - 无需认证
//
// 请求类型:
//   - Content-Type: application/json
//
// 请求参数: 无

// 查询参数：无
//
// 响应数据:
//
//	data:
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
var SpecificArticleHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		tx = tx.Limit(1).Preload("Category").Preload("Tags")
		return tx.Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})

	if err != nil {
		return err
	}

	resp.Data = article

	return nil
})

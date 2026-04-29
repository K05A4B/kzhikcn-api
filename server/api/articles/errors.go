package articles

import (
	"kzhikcn/server/common/hdl"
)

var (
	ErrArticleNotFound           = hdl.DefineError(404, "没有找到文章", "articles.not_found")
	ErrArticleCleanAssetsFailed  = hdl.DefineError(500, "清除文章资源失败", "articles.clean_assets_failed")
	ErrArticleDeleteAssetsFailed = hdl.DefineError(500, "删除文章资源失败", "articles.delete_assets_failed")
	ErrCategoryNotFound          = hdl.DefineError(404, "没有找到分类", "articles.category_not_found")
	ErrFindArticleFailed         = hdl.DefineError(500, "查询文章失败", "articles.find_failed")
	ErrCreateArticleFailed       = hdl.DefineError(500, "创建文章失败", "articles.create_failed")
	ErrDeleteArticleFailed       = hdl.DefineError(500, "删除文章失败", "articles.delete_failed")
	ErrUpdateArticleLikesFailed  = hdl.DefineError(500, "更新点赞量失败", "articles.likes.update_failed")
	ErrUpdateArticleViewsFailed  = hdl.DefineError(500, "更新浏览量失败", "articles.views.update_failed")
	ErrRestoreArticleFailed      = hdl.DefineError(500, "恢复文章失败", "articles.restore_failed")
	ErrUpdateArticleInfoFailed   = hdl.DefineError(500, "更新文章信息失败", "articles.update_failed")

	ErrContentNotFound     = hdl.DefineError(404, "没有找到文章正文", "articles.content.not_found")
	ErrContentLoadFailed   = hdl.DefineError(500, "加载文章正文失败", "articles.content.load_failed")
	ErrContentRenderFailed = hdl.DefineError(500, "渲染文章正文失败", "articles.content.render_failed")
	ErrContentWriteFailed  = hdl.DefineError(500, "更新文章正文失败", "articles.content.write_failed")

	ErrAssetsDeleteFailed       = hdl.DefineError(500, "资源删除失败", "articles.assets.delete_failed")
	ErrAssetsCheckStatFailed    = hdl.DefineError(500, "检查资源状态失败", "articles.assets.check_asset_failed")
	ErrAssetsNotFound           = hdl.DefineError(404, "没有找到资源", "articles.assets.not_found")
	ErrAssetsOpenFailed         = hdl.DefineError(500, "加载资源失败", "articles.assets.load_failed")
	ErrAssetsListFailed         = hdl.DefineError(500, "列出资源列表失败", "articles.assets.list_failed")
	ErrAssetsFileMissing        = hdl.DefineError(400, "上传的负载中没有找到文件", "articles.assets.file_missing")
	ErrAssetsInvalidFilename    = hdl.DefineError(400, "资源ID（文件名）不合法", "articles.assets.invalid_filename")
	ErrAssetsFilenameIsRequired = hdl.DefineError(400, "资源ID（文件名）是必须提供的", "articles.assets.filename_is_required")
	ErrAssetsUploadFailed       = hdl.DefineError(500, "上传资源失败", "articles.assets.upload_failed")
)

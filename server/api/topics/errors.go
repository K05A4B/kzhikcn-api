package topics

import "kzhikcn/server/common/hdl"

var (
	ErrTagDeleteIdIsRequired                = hdl.DefineError(400, "请提提供要删除的tag的id (ids)", "topics.tags.delete.ids_is_empty")
	ErrCategoryCreateCategoryNameIsRequired = hdl.DefineError(400, "主题名字为必填项 (categoryName)", "topics.categories.create.category_name_is_required")

	ErrTagDeleteFailed      = hdl.DefineError(500, "删除标签失败", "topics.tags.delete_failed")
	ErrTagsFindFailed       = hdl.DefineError(500, "查询标签信息失败", "topics.tags.find_failed")
	ErrCategoryUpdateFailed = hdl.DefineError(500, "更新分类信息失败", "topics.categories.update_failed")
	ErrTagUpdateFailed      = hdl.DefineError(500, "更新标签信息失败", "topics.tags.update_failed")
	ErrTagNotFound          = hdl.DefineError(404, "没有找到标签信息", "topics.tags.not_found_tag")

	ErrCategoryNotFound     = hdl.DefineError(404, "没有找到分类信息", "topics.categories.not_found_category")
	ErrCategoryIsExist      = hdl.DefineError(400, "分类已存在", "topics.categories.is_exist")
	ErrCategoryCreateFailed = hdl.DefineError(500, "创建分类失败", "topics.categories.create_failed")
	ErrCategoryDeleteFailed = hdl.DefineError(500, "删除分类失败", "topics.categories.delete_failed")
	ErrCategoriesFindFailed = hdl.DefineError(500, "查询分类信息失败", "topics.categories.find_failed")
)

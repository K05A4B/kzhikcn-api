package article

import (
	"errors"
	"io"
)

var ErrContentNotFound = errors.New("content not found")
var ErrAssetsDirNotFound = errors.New("assets not found")

type Repository interface {
	// 打开文章附属资源
	OpenAsset(articleId, name string) (file io.ReadWriteCloser, err error)

	// 删除文章附属资源
	RemoveAsset(articleId, name string) error

	// 列出所有文章附属资源
	ListAssets(articleId string) (list []string, err error)

	// 检查是否有某个资源
	HasAsset(articleId, name string) (bool, error)

	// 获取文章正文内容的reader
	ContentReader(articleId string) (reader io.ReadCloser, err error)

	// 获取文章正文内容的writer
	// 在ContentReader打开时不允许打开ContentWriter 且只能打开一个ContentWriter
	ContentWriter(articleId string) (reader io.WriteCloser, err error)

	// 删除文章所有资源
	Remove(articleId string) error
}

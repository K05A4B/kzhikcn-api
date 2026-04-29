package articles

import (
	"io"
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/hdl"

	"net/http"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

var UploadArticleAssetHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Limit(1)
	})
	if err != nil {
		return err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return ErrAssetsFileMissing.Wrap(err)
	}

	filename := header.Filename

	filename = filepath.Base(filename)
	if strings.HasPrefix(filename, "..") || strings.HasPrefix(filename, "/") {
		return ErrAssetsInvalidFilename
	}

	if utils.IsEmptyString(filename) {
		return ErrAssetsFilenameIsRequired
	}

	wr, err := assets.ArticlesRepo.OpenAsset(article.ID.String(), filename)
	if err != nil {
		return ErrAssetsOpenFailed.Wrap(err)
	}

	defer wr.Close()

	_, err = io.Copy(wr, file)
	if err != nil {
		return ErrAssetsUploadFailed.Wrap(err)
	}
	resp.Data = filename

	return nil
})

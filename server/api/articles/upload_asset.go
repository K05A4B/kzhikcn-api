package articles

import (
	"kzhikcn/pkg/hdl"
	"kzhikcn/server/app"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

func UploadArticleAsset(appCtx *app.AppContext) hdl.Handler[any] {
	return hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
		articleID := chi.URLParam(r, "article_id")

		file, header, err := r.FormFile("file")
		if err != nil {
			return ErrAssetsFileMissing.Wrap(err)
		}

		filename := filepath.Base(header.Filename)
		if strings.HasPrefix(filename, "..") || strings.HasPrefix(filename, "/") {
			return ErrAssetsInvalidFilename
		}

		err = appCtx.ArticleSvc.UploadAsset(r.Context(), articleID, filename, file)
		if err != nil {
			return ErrAssetsUploadFailed.Wrap(err)
		}

		resp.Data = filename
		return nil
	})
}

package articles

import (
	"bytes"
	"io"
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/articlemd"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/httputil"
	"net/http"

	"gorm.io/gorm"
)

var GetArticleContentHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Limit(1)
	})

	if err != nil {
		return err
	}

	reader, err := assets.ArticlesRepo.ContentReader(article.ID.String())
	if err == assets.ErrContentNotFound {
		return ErrContentNotFound
	}

	if err != nil {
		return ErrContentLoadFailed.Wrap(err)
	}

	defer reader.Close()

	if httputil.Accepts(r, "text/markdown", "text/html", "text/plain") {
		return hdl.WriteRaw(r, reader, "text/plain; charset=utf-8")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return ErrContentLoadFailed.Wrap(err)
	}

	resp.Data = string(data)

	return nil
})

var GetArticleRenderedContentHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	article, err := getArticleBase(r, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id", "custom_id").
			Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})

	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = articleRenderedContent(article.ID.String(), &buf)
	if err == assets.ErrContentNotFound {
		return ErrContentNotFound
	}

	if err != nil {
		return ErrContentRenderFailed.Wrap(err)
	}

	if httputil.Accepts(r, "text/html") {
		return hdl.WriteRawData(r, buf.Bytes(), "text/html; charset=utf-8")
	}

	resp.Data = buf.String()

	return nil
})

func articleRenderedContent(articleId string, w io.Writer) error {
	reader, err := assets.ArticlesRepo.ContentReader(articleId)
	if err != nil {
		return err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return articlemd.ParseDocument(content, w)
}

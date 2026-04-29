package articles

import (
	"errors"
	"kzhikcn/pkg/data"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func getArticleBase(r *http.Request, modifier func(tx *gorm.DB) *gorm.DB) (*data.Article, error) {
	id := chi.URLParam(r, "article_id")

	article, err := data.GetArticleByAnyID(id, modifier)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}

	if err != nil {
		return nil, ErrFindArticleFailed.Wrap(err)
	}

	return article, nil
}

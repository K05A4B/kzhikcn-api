package topics

import (
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var UpdateTagHandler = hdl.NewHandler(func(r *http.Request, resp *hdl.Response, payload data.EditableTag) error {
	tag, err := getTagBase(r)
	if err != nil {
		return err
	}

	err = tag.Update(payload)
	if err != nil {
		return ErrTagUpdateFailed.Wrap(err)
	}

	return nil
})

func getTagBase(r *http.Request, modifiers ...data.QueryModifier) (*data.Tag, error) {
	name := chi.URLParam(r, "tag")
	unescapeName, err := url.QueryUnescape(name)
	if err == nil {
		name = unescapeName
	}

	tags, err := data.GetTags(func(tx *gorm.DB) *gorm.DB {
		id, err := strconv.ParseUint(name, 10, 64)
		if err == nil {
			tx = tx.Where("id=?", id)
		} else {
			tx = tx.Where("tag_name=?", name)
		}

		tx = tx.Limit(1).Select("id")
		tx = data.ApplyQueryModifier(tx, modifiers...)

		return tx
	})

	if len(tags) == 0 {
		return nil, ErrTagNotFound
	}

	if err != nil {
		return nil, ErrTagsFindFailed.Wrap(err)
	}

	tag := tags[0]
	return &tag, nil
}

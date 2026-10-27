package service

import (
	"context"
	"kzhikcn/pkg/data"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrTagNotFound   = errors.New("tag not found")
	ErrCategoryExist = errors.New("category already exists")
)

type TopicService struct {
	ctx *ServiceContext
}

func NewTopicService(ctx *ServiceContext) *TopicService {
	return &TopicService{ctx: ctx}
}

func (s *TopicService) GetCategories(ctx context.Context, page, limit int, modifier data.QueryModifier) ([]data.Category, int64, error) {
	categories, err := data.GetCategories(ctx, modifier, data.Pagination(page, limit))
	if err != nil {
		return nil, 0, err
	}

	total, err := data.Total(ctx, data.Category{}, modifier)
	if err != nil {
		return nil, 0, err
	}
	return categories, total, nil
}

func (s *TopicService) CreateCategory(ctx context.Context, input data.EditableCategory) (*data.Category, error) {
	category, err := data.CreateCategory(ctx, input)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrCategoryExist
		}
		return nil, err
	}
	return category, nil
}

func (s *TopicService) UpdateCategory(ctx context.Context, categoryID uint, input data.EditableCategory) error {
	categories, err := data.GetCategories(ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id=?", categoryID).Limit(1)
	})
	if err != nil || len(categories) == 0 {
		return ErrCategoryNotFound
	}

	category := categories[0]
	return category.Update(ctx, input)
}

func (s *TopicService) GetCategoryByNameOrID(ctx context.Context, name string) (*data.Category, error) {
	categories, err := data.GetCategories(ctx, func(tx *gorm.DB) *gorm.DB {
		if isNumeric(name) {
			tx = tx.Where("id=?", name)
		} else {
			tx = tx.Where("category_name=?", name)
		}
		return tx.Limit(1)
	})
	if err != nil || len(categories) == 0 {
		return nil, ErrCategoryNotFound
	}
	c := categories[0]
	return &c, nil
}

func (s *TopicService) DeleteCategories(ctx context.Context, ids []uint) error {
	return data.DeleteCategories(ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id IN ?", ids)
	})
}

func (s *TopicService) GetTags(ctx context.Context, page, limit int, modifier data.QueryModifier) ([]data.Tag, int64, error) {
	tags, err := data.GetTags(ctx, modifier, data.Pagination(page, limit))
	if err != nil {
		return nil, 0, err
	}

	total, err := data.Total(ctx, data.Tag{}, modifier)
	if err != nil {
		return nil, 0, err
	}
	return tags, total, nil
}

func (s *TopicService) UpdateTag(ctx context.Context, tagID uint, input data.EditableTag) error {
	tags, err := data.GetTags(ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id=?", tagID).Limit(1)
	})
	if err != nil || len(tags) == 0 {
		return ErrTagNotFound
	}

	tag := tags[0]
	return tag.Update(ctx, input)
}

func (s *TopicService) GetTagByNameOrID(ctx context.Context, name string) (*data.Tag, error) {
	tags, err := data.GetTags(ctx, func(tx *gorm.DB) *gorm.DB {
		if isNumeric(name) {
			tx = tx.Where("id=?", name)
		} else {
			tx = tx.Where("tag_name=?", name)
		}
		return tx.Limit(1)
	})
	if err != nil || len(tags) == 0 {
		return nil, ErrTagNotFound
	}
	t := tags[0]
	return &t, nil
}

func (s *TopicService) DeleteTags(ctx context.Context, ids []uint) error {
	return data.DeleteTag(ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id IN ?", ids)
	})
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

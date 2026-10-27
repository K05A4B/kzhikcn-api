package data

import (
	"context"
	"kzhikcn/pkg/utils"
)

type EditableCategory struct {
	CategoryName string `json:"categoryName"`
	Description  string `json:"description"`
}

type Category struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CategoryName string    `gorm:"type:VARCHAR(255);uniqueIndex" json:"categoryName"`
	Description  string    `json:"description"`
	Articles     []Article `gorm:"foreignKey:CategoryID" json:"articles,omitzero"`
}

func (c *Category) Update(ctx context.Context, ec EditableCategory) error {
	category := Category{
		CategoryName: ec.CategoryName,
		Description:  ec.Description,
	}

	selectedField := []string{}

	if !utils.IsEmptyString(ec.CategoryName) {
		selectedField = append(selectedField, "category_name")
	}

	if !utils.IsEmptyString(ec.Description) {
		selectedField = append(selectedField, "description")
	}

	tx := withContext(ctx).Model(c)

	return tx.Select(selectedField).Updates(category).Error
}

func GetCategories(ctx context.Context, modifiers ...QueryModifier) ([]Category, error) {
	tx := withContext(ctx).Model(&Category{})
	tx = ApplyQueryModifier(tx, modifiers...)

	categories := []Category{}
	return categories, tx.Find(&categories).Error
}

func CreateCategory(ctx context.Context, ec EditableCategory) (*Category, error) {
	selectedFields := []string{"id"}

	category := &Category{
		CategoryName: ec.CategoryName,
		Description:  ec.Description,
	}

	if !utils.IsEmptyString(ec.CategoryName) {
		selectedFields = append(selectedFields, "category_name")
	}

	if !utils.IsEmptyString(ec.Description) {
		selectedFields = append(selectedFields, "description")
	}

	return category, withContext(ctx).Select(selectedFields).Create(category).Error
}

func DeleteCategories(ctx context.Context, modifiers ...QueryModifier) error {
	tx := withContext(ctx).Model(Category{})

	tx = ApplyQueryModifier(tx, modifiers...)

	return tx.Delete(Category{}).Error
}

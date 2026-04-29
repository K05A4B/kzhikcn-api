package data

import (
	"kzhikcn/pkg/utils"

	"gorm.io/gorm"
)

type EditableTag struct {
	TagName string `json:"tagName"`
}

type Tag struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	TagName  string    `gorm:"uniqueIndex;type:VARCHAR(255)" json:"tagName"`
	Articles []Article `gorm:"many2many:article_tags" json:"articles,omitzero"`
}

func (a *Tag) Update(et EditableTag) error {
	tag := &Tag{
		TagName: et.TagName,
	}

	selectedFields := []string{}
	tx := db.Model(a)

	if !utils.IsEmptyString(et.TagName) {
		selectedFields = append(selectedFields, "tag_name")
	}

	return tx.Select(selectedFields).Updates(tag).Error
}

func FindOrCreateTagsByName(tagNames ...string) ([]*Tag, error) {
	tags := []*Tag{}

	return tags, db.Transaction(func(tx *gorm.DB) error {
		for _, tagName := range tagNames {
			if tagName == "" {
				continue
			}

			tag := &Tag{
				TagName: tagName,
			}

			err := tx.Where("tag_name=?", tagName).FirstOrCreate(tag).Error
			if err != nil {
				return err
			}

			tags = append(tags, tag)
		}

		return nil
	})

}

func DeleteTag(modifiers ...QueryModifier) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tx = ApplyQueryModifier(tx, modifiers...)
		return tx.Delete(&Tag{}).Error
	})
}

package data

import (
	"gorm.io/gorm"
)

type QueryModifier func(tx *gorm.DB) *gorm.DB

func ApplyQueryModifier(tx *gorm.DB, qms ...QueryModifier) *gorm.DB {
	for _, qm := range qms {
		tx = qm(tx)
	}

	return tx
}

func Adapter(modifiers ...QueryModifier) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return ApplyQueryModifier(tx, modifiers...)
	}
}

func LimitQueryModifier(limit int) QueryModifier {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Limit(limit)
	}
}

func OffsetQueryModifier(offset int) QueryModifier {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Offset(offset)
	}
}

func WhereQueryModifier(query any, cond ...any) QueryModifier {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(query, cond...)
	}
}

func Total(v any, mods ...QueryModifier) (int64, error) {
	var res int64

	tx := db

	switch v := v.(type) {
	case string:
		tx = tx.Table(v)

	default:
		tx = tx.Model(v)
	}

	tx = ApplyQueryModifier(tx, mods...)

	err := tx.Count(&res).Error
	if err != nil {
		return 0, err
	}

	return res, nil
}

func OnlyID(tx *gorm.DB) *gorm.DB {
	return tx.Select("id")
}

func OnlyCustomID(tx *gorm.DB) *gorm.DB {
	return tx.Select("custom_id")
}

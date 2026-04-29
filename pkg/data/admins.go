package data

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Admin struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string `gorm:"type:VARCHAR(255);uniqueIndex" json:"username"`
	Password   []byte `gorm:"type:blob" json:"-"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	EnableMFA  bool   `gorm:"default:0;" json:"enableMFA"`
	TotpSecret []byte `gorm:"default:null;type:blob" json:"-"`
}

func (a *Admin) setPassword() (err error) {
	if len(a.Password) != 0 {
		a.Password, err = bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
		if err != nil {
			return
		}
	}

	return
}

func (a *Admin) BeforeCreate(tx *gorm.DB) (err error) {
	err = a.setPassword()
	return
}

func (a *Admin) BeforeUpdate(tx *gorm.DB) (err error) {
	err = a.setPassword()
	return
}

func GetAdmin(modifiers ...QueryModifier) ([]Admin, error) {
	admins := []Admin{}
	tx := db.Model(&Admin{})

	tx = ApplyQueryModifier(tx, modifiers...)

	tx = tx.Find(&admins)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return admins, nil
}

func GetAdminIDByName(username string) (uint, error) {
	admin, err := GetAdminByName(username, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id")
	})

	if err != nil {
		return 0, err
	}

	return admin.ID, err
}

func GetAdminByName(username string, modifiers ...QueryModifier) (*Admin, error) {
	admins, err := GetAdmin(func(tx *gorm.DB) *gorm.DB {
		tx = ApplyQueryModifier(tx, modifiers...)
		tx = tx.Limit(1)
		tx = tx.Where("username=?", username)
		return tx
	})

	if err != nil {
		return nil, err
	}

	if len(admins) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &admins[0], nil
}

func GetAdminById(id uint, modifiers ...QueryModifier) (*Admin, error) {
	admins, err := GetAdmin(func(tx *gorm.DB) *gorm.DB {
		tx = ApplyQueryModifier(tx, modifiers...)
		tx = tx.Limit(1)
		tx = tx.Where("id=?", id)
		return tx
	})

	if err != nil {
		return nil, err
	}

	if len(admins) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &admins[0], nil
}

func AddAdmin(admin *Admin) error {
	return db.Create(admin).Error
}

func DeleteAdminByID(id uint) error {
	return db.Delete(&Admin{}, "id=?", id).Error
}

func UpdateAdminByID(id uint, admin *Admin, modifiers ...QueryModifier) error {
	tx := ApplyQueryModifier(db, modifiers...)
	return tx.Where("id=?", id).Updates(admin).Error
}

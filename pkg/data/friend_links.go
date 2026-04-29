package data

import (
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FriendLinkAuditStatus string

const (
	FriendLinkStatusPending  FriendLinkAuditStatus = "pending"
	FriendLinkStatusApproved FriendLinkAuditStatus = "approved"
	FriendLinkStatusRejected FriendLinkAuditStatus = "rejected"
	FriendLinkStatusIgnored  FriendLinkAuditStatus = "ignored"
)

type FriendLink struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Url         string `gorm:"not null" json:"url"`
	Priority    int    `gorm:"default:0" json:"priority"`
	CoverUrl    string `gorm:"default:'auto'" json:"coverUrl"`
	Group       string `gorm:"default:'normal'" json:"group"`
	Email       string `gorm:"null" json:"email"`
	IsEnabled   *bool  `gorm:"default:1" json:"isEnabled"`
}

type FriendLinkAudit struct {
	ID          uuid.UUID             `gorm:"primaryKey" json:"id"`
	Name        string                `gorm:"not null" json:"name"`
	CreatedAt   time.Time             `json:"createdAt"`
	Description string                `json:"description"`
	Url         string                `gorm:"not null" json:"url"`
	CoverUrl    string                `gorm:"default:'auto'" json:"coverUrl"`
	Email       string                `gorm:"not null" json:"email"`
	Status      FriendLinkAuditStatus `gorm:"not null;default:'pending'" json:"status"`
	Ip          string                `gorm:"not null" json:"ip"`
}

func (f *FriendLinkAudit) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}

	return nil
}

func (f *FriendLinkAudit) SetStatus(status FriendLinkAuditStatus) error {
	if f.ID == uuid.Nil {
		return errors.New("FriendLinkAudit: id cannot be nil")
	}

	return db.Model(f).UpdateColumn("status", status).Error
}

func FindFriendLinkAudit(modifiers ...QueryModifier) ([]FriendLinkAudit, error) {
	tx := ApplyQueryModifier(db, modifiers...)
	links := []FriendLinkAudit{}

	err := tx.Find(&links).Error
	if err != nil {
		return nil, err
	}

	return links, nil
}

func FindFriendLinks(modifiers ...QueryModifier) ([]FriendLink, error) {
	tx := ApplyQueryModifier(db, modifiers...)
	links := []FriendLink{}

	err := tx.Find(&links).Error
	if err != nil {
		return nil, err
	}

	return links, nil
}

func ToFriendLinkAuditStatus(s string) FriendLinkAuditStatus {
	switch s {
	case "approved", "rejected", "ignored":
		return FriendLinkAuditStatus(s)
	default:
		return FriendLinkStatusPending
	}
}

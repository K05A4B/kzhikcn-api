package data

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArticleStatus string

func (a *ArticleStatus) String() string {
	return string(*a)
}

const (
	ARTICLE_STATUS_PUBLISHED ArticleStatus = "published"
	ARTICLE_STATUS_DRAFT     ArticleStatus = "draft"
	ARTICLE_STATUS_HIDDEN    ArticleStatus = "hidden"
)

type Article struct {
	ID            uuid.UUID      `gorm:"primaryKey;type:CHAR(36);" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	PublishedAt   *time.Time     `json:"publishedAt"`
	CustomID      string         `gorm:"uniqueIndex;type:VARCHAR(255)" json:"customID"`
	Title         string         `gorm:"not null" json:"title"`
	Views         int            `gorm:"default:0" json:"views"`
	Likes         int            `gorm:"default:0" json:"likes"`
	CategoryID    *uint          `gorm:"default:null;constraint:OnDelete:SET NULL;" json:"categoryID"`
	Category      Category       `gorm:"foreignKey:CategoryID" json:"category"`
	Tags          []Tag          `gorm:"many2many:article_tags;" json:"tags"`
	Status        ArticleStatus  `gorm:"not null;default:draft" json:"status"`
	Description   string         `json:"description"`
	CoverImage    string         `gorm:"default:null" json:"coverImage"`
	EnableComment bool           `gorm:"default:false" json:"enableComment"`
}

type ArticleTag struct {
	ArticleID uuid.UUID `gorm:"primaryKey"`
	TagID     uint      `gorm:"primaryKey"`

	Article Article `gorm:"constraint:OnDelete:CASCADE;"`
	Tag     Tag     `gorm:"constraint:OnDelete:CASCADE;"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	if a.CustomID == "" {
		a.CustomID = a.ID.String()
	}

	if a.Status == ARTICLE_STATUS_PUBLISHED && a.PublishedAt == nil {
		now := time.Now()
		a.PublishedAt = &now
	}

	return nil
}

func GetArticleByAnyID(id string, modifiers ...QueryModifier) (*Article, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return GetArticleByCustomID(id, modifiers...)
	}

	return GetArticleById(id, modifiers...)
}

func GetArticles(m ...QueryModifier) ([]Article, error) {
	tx := db.Model(&Article{})
	tx = ApplyQueryModifier(tx, m...)

	articles := []Article{}

	return articles, tx.Find(&articles).Error
}

func getFirstArticle(modifiers ...QueryModifier) (*Article, error) {
	articles, err := GetArticles(modifiers...)

	if err != nil {
		return nil, err
	}

	if len(articles) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &articles[0], nil
}

func GetArticleById(id string, modifiers ...QueryModifier) (*Article, error) {
	return getFirstArticle(func(tx *gorm.DB) *gorm.DB {
		tx = ApplyQueryModifier(tx, modifiers...)
		return tx.Where("id=?", id)
	})
}

func GetArticleByCustomID(id string, modifiers ...QueryModifier) (*Article, error) {
	return getFirstArticle(func(tx *gorm.DB) *gorm.DB {
		tx = ApplyQueryModifier(tx, modifiers...)
		return tx.Where("custom_id=?", id)
	})
}

func GetTags(modifiers ...QueryModifier) ([]Tag, error) {
	tx := db.Model(&Tag{})
	tx = ApplyQueryModifier(tx, modifiers...)

	tags := []Tag{}
	return tags, tx.Find(&tags).Error
}

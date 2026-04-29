package data

import (
	"database/sql"
	"kzhikcn/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrArticleIDIsEmpty = errors.New("article id is empty")
	ErrCategoryNotFound = errors.New("category not found")
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

type EditableArticle struct {
	Title         string        `json:"title"`
	CustomID      string        `json:"customId"`
	Category      string        `json:"category"`
	Tags          []string      `json:"tags"`
	Status        ArticleStatus `json:"status"`
	Description   string        `json:"description"`
	CoverImage    string        `json:"coverImage"`
	EnableComment *bool         `json:"enableComment"`
}

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

func (a *Article) BeforeUpdate(tx *gorm.DB) (err error) {
	if a.Status == ARTICLE_STATUS_PUBLISHED {
		var publishedAt sql.NullTime
		err = tx.Model(&Article{}).Select("published_at").Where("id=?", a.ID).Scan(&publishedAt).Error
		if err != nil {
			return err
		}

		if !publishedAt.Valid {
			now := time.Now()
			a.PublishedAt = &now
			tx.Statement.SetColumn("published_at", now)
		}
	}

	return nil
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

func (a *Article) Update(ea EditableArticle) error {
	if a.ID == uuid.Nil {
		return ErrArticleIDIsEmpty
	}

	return db.Transaction(func(tx *gorm.DB) error {
		selectedFields := []string{}

		article := &Article{
			Title:       ea.Title,
			CustomID:    ea.CustomID,
			Description: ea.Description,
			CoverImage:  ea.CoverImage,
		}

		if !utils.IsEmptyString(ea.Title) {
			selectedFields = append(selectedFields, "title")
		}

		if !utils.IsEmptyString(ea.CustomID) {
			selectedFields = append(selectedFields, "custom_id")
		}

		if !utils.IsEmptyString(ea.Description) {
			selectedFields = append(selectedFields, "description")
		}

		if ea.EnableComment != nil {
			article.EnableComment = *ea.EnableComment
			selectedFields = append(selectedFields, "enable_comment")
		}

		if !utils.IsEmptyString(ea.CoverImage) {
			selectedFields = append(selectedFields, "cover_image")
		}

		if !utils.IsEmptyString(ea.Status.String()) {
			article.Status = ToArticleStatus(ea.Status.String())
			selectedFields = append(selectedFields, "status")
		}

		err := tx.Model(a).Select(selectedFields).Updates(article).Error
		if err != nil {
			return err
		}

		if ea.Tags != nil {
			tags := []Tag{}
			for _, tagName := range ea.Tags {
				var tag Tag
				if err := tx.Where("tag_name=?", tagName).FirstOrCreate(&tag, Tag{TagName: tagName}).Error; err != nil {
					return err
				}
				tags = append(tags, tag)
			}

			err = tx.Model(a).Association("Tags").Replace(tags)
			if err != nil {
				return err
			}
		}

		if !utils.IsEmptyString(ea.Category) {
			category := Category{CategoryName: ea.Category}

			err := tx.Where("category_name=?", category.CategoryName).First(&category).Error
			if err == gorm.ErrRecordNotFound {
				return ErrCategoryNotFound
			}
			if err != nil {
				return err
			}

			err = tx.Model(a).Association("Category").Replace(&category)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (a *Article) SetViews(views int) error {
	return db.Model(a).UpdateColumn("views", views).Error
}

func (a *Article) IncrementViews() (int, error) {
	err := db.Model(a).UpdateColumn("views", gorm.Expr("views+?", 1)).Error
	if err != nil {
		return 0, err
	}

	err = db.Model(a).Select("views", "id").Find(a).Error
	if err != nil {
		return 0, err
	}

	return a.Views, nil
}

func (a *Article) SetLikes(views int) error {
	return db.Model(a).UpdateColumn("likes", views).Error
}

func (a *Article) IncrementLikes() (int, error) {
	err := db.Model(a).UpdateColumn("likes", gorm.Expr("likes+?", 1)).Error
	if err != nil {
		return 0, err
	}

	err = db.Model(a).Select("likes", "id").Find(a).Error
	if err != nil {
		return 0, err
	}

	return a.Likes, nil
}

func (a *Article) TouchUpdatedAt() error {
	return db.Model(a).Where("id=?", a.ID).Select("updated_at").UpdateColumn("updated_at", time.Now()).Error
}

func CreateArticle(ea EditableArticle, modifiers ...QueryModifier) (*Article, error) {
	var ar *Article

	err := db.Transaction(func(tx *gorm.DB) error {
		var category Category
		if strings.TrimSpace(ea.Category) != "" {
			if err := tx.Where("category_name = ?", ea.Category).First(&category).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return ErrCategoryNotFound
				}
				return err
			}
		}

		var tags []Tag
		for _, tagName := range ea.Tags {
			tag := Tag{}

			if err := tx.Where("tag_name = ?", tagName).FirstOrCreate(&tag, Tag{
				TagName: tagName,
			}).Error; err != nil {
				return err
			}

			tags = append(tags, tag)
		}

		enableComment := false
		if ea.EnableComment != nil {
			enableComment = *ea.EnableComment
		}

		ar = &Article{
			Title:         ea.Title,
			CustomID:      ea.CustomID,
			Tags:          tags,
			Status:        ToArticleStatus(ea.Status.String()),
			Description:   ea.Description,
			EnableComment: enableComment,
			CoverImage:    ea.CoverImage,
		}

		if strings.TrimSpace(ea.Category) != "" {
			ar.Category = category
		}

		tx = ApplyQueryModifier(tx, modifiers...)
		if err := tx.Create(ar).Error; err != nil {
			return err
		}

		return nil
	})

	return ar, err
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

func DeleteArticle(isHard bool, modifiers ...QueryModifier) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&Article{})
		tx = ApplyQueryModifier(tx, modifiers...)

		if tx == nil {
			return nil
		}

		if isHard {
			tx = tx.Unscoped()
		}

		return tx.Delete(&Article{}).Error
	})
}

func RestoreArticle(modifiers ...QueryModifier) error {
	tx := db.Model(&Article{})
	tx = tx.Unscoped()
	tx = ApplyQueryModifier(tx, modifiers...)

	return tx.UpdateColumn("deleted_at", nil).Error
}

func GetTags(modifiers ...QueryModifier) ([]Tag, error) {
	tx := db.Model(&Tag{})
	tx = ApplyQueryModifier(tx, modifiers...)

	tags := []Tag{}
	return tags, tx.Find(&tags).Error
}

func PruneTags(modifiers ...QueryModifier) error {
	tx := db.Model(&Tag{})
	tx = ApplyQueryModifier(tx, modifiers...)

	noEmpties := []uint{}

	err := db.Table("article_tags").Distinct("tag_id").Find(&noEmpties).Error
	if err != nil {
		return err
	}

	if len(noEmpties) == 0 {
		return nil
	}

	return tx.Where("id NOT IN (?)", noEmpties).Delete(&Tag{}).Error
}

func PruneCategories(modifiers ...QueryModifier) error {
	ids := []uint{}
	tx := db.Model(&Category{}).Distinct("categories.id").
		Joins("INNER JOIN articles ON categories.id = articles.category_id")

	tx = ApplyQueryModifier(tx, modifiers...)

	err := tx.Find(&ids).Error

	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	return db.Model(&Category{}).Where("id NOT IN (?)", ids).Delete(&Category{}).Error
}

func ToArticleStatus(s string) ArticleStatus {
	switch ArticleStatus(strings.ToLower(s)) {
	case ARTICLE_STATUS_PUBLISHED:
		return ARTICLE_STATUS_PUBLISHED

	case ARTICLE_STATUS_HIDDEN:
		return ARTICLE_STATUS_HIDDEN

	default:
		return ARTICLE_STATUS_DRAFT
	}
}

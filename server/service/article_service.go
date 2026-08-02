package service

import (
	"bytes"
	"context"
	"io"
	"kzhikcn/pkg/assets/article"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/articlemd"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrArticleNotFound  = errors.New("article not found")
	ErrCategoryNotFound = errors.New("category not found")
)

type ArticleHooks struct {
	BeforeCreate     []func(ctx context.Context, article *ArticleUpdateFields) error
	AfterCreate      []func(ctx context.Context, article *data.Article) error
	BeforePublish    []func(ctx context.Context, article *data.Article) error
	AfterPublish     []func(ctx context.Context, article *data.Article) error
	BeforeDelete     []func(ctx context.Context, articleID string, isHard bool) error
	AfterDelete      []func(ctx context.Context, articleID string, isHard bool) error
	BeforeUpdateInfo []func(ctx context.Context, article *data.Article) error
	AfterUpdateInfo  []func(ctx context.Context, article *data.Article) error
}

type ArticleService struct {
	ctx   *ServiceContext
	hooks *ArticleHooks
}

type GetArticlesOptions struct {
	Page       int
	Limit      int
	OrderBy    string
	OmitStatus []data.ArticleStatus
}

func NewArticleService(ctx *ServiceContext, hooks *ArticleHooks) *ArticleService {
	if hooks == nil {
		hooks = &ArticleHooks{}
	}
	return &ArticleService{ctx: ctx, hooks: hooks}
}

type ArticleUpdateFields struct {
	Title         *string             `json:"title"`
	CustomID      *string             `json:"customID"`
	Category      *string             `json:"category"`
	Tags          []string            `json:"tags"`
	Status        *data.ArticleStatus `json:"status"`
	Description   *string             `json:"description"`
	CoverImage    *string             `json:"coverImage"`
	EnableComment *bool               `json:"enableComment"`
}

func getArticleByID(id string, mods ...data.QueryModifier) (*data.Article, error) {
	article, err := data.GetArticleByAnyID(id, append([]data.QueryModifier{data.LimitQueryModifier(1)}, mods...)...)
	if article == nil || err == gorm.ErrRecordNotFound {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return article, err
}

func (svc *ArticleService) Update(ctx context.Context, id string, ea ArticleUpdateFields) (*data.Article, error) {
	article, err := getArticleByID(id)
	if err != nil {
		return nil, err
	}

	for _, hook := range svc.hooks.BeforeUpdateInfo {
		if err := hook(ctx, article); err != nil {
			return nil, err
		}
	}

	err = svc.ctx.DB.Transaction(func(tx *gorm.DB) error {
		selectedFields := []string{}

		if ea.Title != nil {
			article.Title = *ea.Title
			selectedFields = append(selectedFields, "title")
		}
		if ea.CustomID != nil {
			article.CustomID = *ea.CustomID
			selectedFields = append(selectedFields, "custom_id")
		}
		if ea.Description != nil {
			article.Description = *ea.Description
			selectedFields = append(selectedFields, "description")
		}
		if ea.EnableComment != nil {
			article.EnableComment = *ea.EnableComment
			selectedFields = append(selectedFields, "enable_comment")
		}
		if ea.CoverImage != nil {
			article.CoverImage = *ea.CoverImage
			selectedFields = append(selectedFields, "cover_image")
		}
		if ea.Status != nil {
			newStatus := *ea.Status
			article.Status = newStatus
			selectedFields = append(selectedFields, "status")
			if newStatus == data.ARTICLE_STATUS_PUBLISHED && article.PublishedAt == nil {
				now := time.Now()
				article.PublishedAt = &now
				selectedFields = append(selectedFields, "published_at")
			}
		}

		err := tx.Model(data.Article{}).Select(selectedFields).Where("id=?", article.ID).Updates(article).Error
		if err != nil {
			return err
		}

		if ea.Tags != nil {
			tags := []data.Tag{}
			for _, tagName := range ea.Tags {
				var tag data.Tag
				if err := tx.Where("tag_name=?", tagName).FirstOrCreate(&tag, data.Tag{TagName: tagName}).Error; err != nil {
					return err
				}
				tags = append(tags, tag)
			}
			err = tx.Model(article).Association("Tags").Replace(tags)
			if err != nil {
				return err
			}
		}

		if ea.Category != nil {
			categoryString := *ea.Category
			category := data.Category{}
			if categoryString != "" && categoryString[0] == '#' {
				id, err := strconv.ParseUint(categoryString[1:], 10, 64)
				if err == nil {
					category.ID = uint(id)
				}
			}
			if category.ID != 0 {
				err = tx.Where("id=?", category.ID).Limit(1).First(&category).Error
			} else {
				err = tx.Where("category_name=?", categoryString).Limit(1).First(&category).Error
			}
			if err == gorm.ErrRecordNotFound {
				return ErrCategoryNotFound
			}
			if err != nil {
				return err
			}
			err = tx.Model(article).Association("Category").Replace(&category)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if ea.Status != nil && *ea.Status == data.ARTICLE_STATUS_PUBLISHED {
		for _, hook := range svc.hooks.AfterPublish {
			if err := hook(ctx, article); err != nil {
				return nil, err
			}
		}
	}

	for _, hook := range svc.hooks.AfterUpdateInfo {
		if err := hook(ctx, article); err != nil {
			return nil, err
		}
	}

	fresh, err := getArticleByID(id,
		func(tx *gorm.DB) *gorm.DB { return tx.Preload("Tags").Preload("Category") })
	if err != nil {
		return nil, err
	}

	return fresh, nil
}

func (svc *ArticleService) IncrementViews(ctx context.Context, id string) (int, error) {
	article, err := getArticleByID(id, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("views", "id").
			Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return 0, err
	}

	err = svc.ctx.DB.Model(article).UpdateColumn("views", gorm.Expr("views+?", 1)).Error
	if err != nil {
		return 0, err
	}
	err = svc.ctx.DB.Model(article).Select("views", "id").Find(article).Error
	if err != nil {
		return 0, err
	}

	return article.Views, nil
}

func (svc *ArticleService) IncrementLikes(ctx context.Context, id string) (int, error) {
	article, err := getArticleByID(id, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("likes", "id").
			Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return 0, err
	}

	err = svc.ctx.DB.Model(article).UpdateColumn("likes", gorm.Expr("likes+?", 1)).Error
	if err != nil {
		return 0, err
	}
	err = svc.ctx.DB.Model(article).Select("likes", "id").Find(article).Error
	if err != nil {
		return 0, err
	}

	return article.Likes, nil
}

func (svc *ArticleService) Create(ctx context.Context, ea *ArticleUpdateFields) (*data.Article, error) {
	for _, hook := range svc.hooks.BeforeCreate {
		if err := hook(ctx, ea); err != nil {
			return nil, err
		}
	}

	var ar *data.Article

	err := svc.ctx.DB.Transaction(func(tx *gorm.DB) error {
		var category data.Category
		if ea.Category != nil {
			err := tx.Where("category_name = ?", ea.Category).First(&category).Error
			if err == gorm.ErrRecordNotFound {
				return ErrCategoryNotFound
			}
			if err != nil {
				return err
			}
		}

		var tags []data.Tag
		for _, tagName := range ea.Tags {
			tag := data.Tag{}
			if err := tx.Where("tag_name = ?", tagName).FirstOrCreate(&tag, data.Tag{TagName: tagName}).Error; err != nil {
				return err
			}
			tags = append(tags, tag)
		}

		ar = &data.Article{
			Title:         utils.NilToValue(ea.Title, ""),
			CustomID:      utils.NilToValue(ea.CustomID, ""),
			Tags:          tags,
			Status:        utils.NilToValue(ea.Status, data.ARTICLE_STATUS_DRAFT),
			Description:   utils.NilToValue(ea.Description, ""),
			EnableComment: utils.NilToValue(ea.EnableComment, false),
			CoverImage:    utils.NilToValue(ea.CoverImage, ""),
		}

		if ea.Category != nil {
			ar.Category = category
		}

		if err := tx.Create(ar).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, hook := range svc.hooks.AfterCreate {
		if err := hook(ctx, ar); err != nil {
			return nil, err
		}
	}

	return ar, nil
}

func (svc *ArticleService) Delete(ctx context.Context, id string, isHard bool) error {
	article, err := getArticleByID(id, data.OnlyID, func(tx *gorm.DB) *gorm.DB {
		if isHard {
			// 修复无法删除标记为已删除的文章的bug
			return tx.Unscoped()
		}

		return tx
	})
	if err != nil {
		return err
	}

	for _, hook := range svc.hooks.BeforeDelete {
		if err := hook(ctx, article.ID.String(), isHard); err != nil {
			return err
		}
	}

	err = svc.ctx.DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&data.Article{}).Where("id=?", article.ID)
		if tx == nil {
			return nil
		}
		if isHard {
			tx = tx.Unscoped()
		}
		if err := tx.Delete(article).Error; err != nil {
			return err
		}

		return svc.ctx.Repo.Remove(article.ID.String())
	})

	if err != nil {
		return err
	}

	for _, hook := range svc.hooks.AfterDelete {
		if err := hook(ctx, article.ID.String(), isHard); err != nil {
			return err
		}
	}

	return nil
}

func (svc *ArticleService) Restore(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx := svc.ctx.DB.Model(&data.Article{})
	tx = tx.Unscoped()
	tx = tx.Where("id IN (?)", ids)

	return tx.UpdateColumn("deleted_at", nil).Error
}

func (svc *ArticleService) GetArticles(ctx context.Context, options GetArticlesOptions, modifier data.QueryModifier) (articles []data.Article, total int64, err error) {
	orderBy := ArticleOrderByMapping[options.OrderBy]
	if orderBy == "" {
		orderBy = "published_at DESC"
	}

	modifiers := []data.QueryModifier{
		modifier,
		func(tx *gorm.DB) *gorm.DB {
			if len(options.OmitStatus) == 0 {
				return tx
			}
			return tx.Where("status NOT IN (?)", options.OmitStatus)
		},
	}

	articles, err = data.GetArticles(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("Category").Preload("Tags")
		tx = tx.Order(orderBy)
		tx = data.ApplyQueryModifier(tx, modifiers...)
		return tx
	}, data.Pagination(options.Page, options.Limit))

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, ErrArticleNotFound
	}

	total, _ = data.Total(articles, modifiers...)

	return
}

func (svc *ArticleService) GetDeletedArticles(ctx context.Context, options GetArticlesOptions, modifier data.QueryModifier) (articles []data.Article, total int64, err error) {
	orderBy := ArticleOrderByMapping[options.OrderBy]
	if orderBy == "" {
		orderBy = "published_at DESC"
	}

	modifiers := []data.QueryModifier{
		modifier,
		func(tx *gorm.DB) *gorm.DB {
			return tx.Unscoped().Not("deleted_at IS ?", nil)
		},
	}

	articles, err = data.GetArticles(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("Category").Preload("Tags")
		tx = tx.Order(orderBy)
		tx = data.ApplyQueryModifier(tx, modifiers...)
		return tx
	}, data.Pagination(options.Page, options.Limit))

	total, _ = data.Total(articles, modifiers...)

	return
}

func (svc *ArticleService) GetArticle(ctx context.Context, id string) (*data.Article, error) {
	article, err := getArticleByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (svc *ArticleService) GetContent(ctx context.Context, articleID string) (io.ReadCloser, error) {
	_, err := getArticleByID(articleID, data.OnlyID)
	if err != nil {
		return nil, err
	}
	return svc.ctx.Repo.ContentReader(articleID)
}

func (svc *ArticleService) UpdateContent(ctx context.Context, articleID string, reader io.Reader) error {
	_, err := getArticleByID(articleID, data.OnlyID)
	if err != nil {
		return err
	}

	writer, err := svc.ctx.Repo.ContentWriter(articleID)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	return err
}

func (svc *ArticleService) GetRenderedContent(ctx context.Context, articleID string) ([]byte, error) {
	_, err := getArticleByID(articleID, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Where("status IN ?",
			[]data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return nil, err
	}

	reader, err := svc.ctx.Repo.ContentReader(articleID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := articlemd.ParseDocument(content, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (svc *ArticleService) UploadAsset(ctx context.Context, articleID, filename string, file io.Reader) error {
	_, err := getArticleByID(articleID, data.OnlyID)
	if err != nil {
		return err
	}

	filename = filepath.Base(filename)
	if strings.HasPrefix(filename, "..") || strings.HasPrefix(filename, "/") {
		return errors.New("invalid filename")
	}
	if utils.IsEmptyString(filename) {
		return errors.New("filename is required")
	}

	wr, err := svc.ctx.Repo.OpenAsset(articleID, filename)
	if err != nil {
		return err
	}
	defer wr.Close()

	_, err = io.Copy(wr, file)
	return err
}

func (svc *ArticleService) DeleteAsset(ctx context.Context, articleID, filename string) error {
	_, err := getArticleByID(articleID, data.OnlyID)
	if err != nil {
		return err
	}
	return svc.ctx.Repo.RemoveAsset(articleID, filename)
}

func (svc *ArticleService) ListAssets(ctx context.Context, articleID string) ([]string, error) {
	_, err := getArticleByID(articleID, data.OnlyID)
	if err != nil {
		return nil, err
	}

	list, err := svc.ctx.Repo.ListAssets(articleID)
	if errors.Is(err, article.ErrAssetsDirNotFound) {
		return []string{}, nil
	}
	return list, err
}

func (svc *ArticleService) HasAsset(ctx context.Context, articleID, filename string) (bool, error) {
	return svc.ctx.Repo.HasAsset(articleID, filename)
}

func (svc *ArticleService) OpenAsset(ctx context.Context, articleID string, assetName string) (io.ReadWriteCloser, error) {
	return svc.ctx.Repo.OpenAsset(articleID, assetName)
}

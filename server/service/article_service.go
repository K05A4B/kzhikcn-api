package service

import (
	"bytes"
	"context"
	"io"
	"kzhikcn/pkg/assets/article"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/log"
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

// isPublishing 判断更新字段是否将文章置为发布状态。
func isPublishing(status *data.ArticleStatus) bool {
	return status != nil && *status == data.ARTICLE_STATUS_PUBLISHED
}

func getArticleByID(ctx context.Context, id string, mods ...data.QueryModifier) (*data.Article, error) {
	article, err := data.GetArticleByAnyID(ctx, id, append([]data.QueryModifier{data.LimitQueryModifier(1)}, mods...)...)
	if article == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return article, nil
}

// resolveCategory 解析分类引用，支持 "#<id>" 与分类名称两种写法。
func resolveCategory(tx *gorm.DB, ref string) (*data.Category, error) {
	category := data.Category{}
	query := tx

	if strings.HasPrefix(ref, "#") {
		id, err := strconv.ParseUint(ref[1:], 10, 64)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		query = query.Where("id=?", id)
	} else {
		query = query.Where("category_name=?", ref)
	}

	err := query.Limit(1).First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	return &category, nil
}

// runArticleHooks 执行事务提交后的钩子。
// 数据已经落库，钩子失败只记录日志，避免调用方因收到错误而重试产生重复副作用。
func (svc *ArticleService) runArticleHooks(stage string, ctx context.Context, article *data.Article,
	hooks []func(ctx context.Context, article *data.Article) error) {
	for _, hook := range hooks {
		if err := hook(ctx, article); err != nil {
			log.Errorf("%s hook failed: articleID=%s, err=%v", stage, article.ID, err)
		}
	}
}

func (svc *ArticleService) Update(ctx context.Context, id string, ea ArticleUpdateFields) (*data.Article, error) {
	article, err := getArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, hook := range svc.hooks.BeforeUpdateInfo {
		if err := hook(ctx, article); err != nil {
			return nil, err
		}
	}

	if isPublishing(ea.Status) {
		for _, hook := range svc.hooks.BeforePublish {
			if err := hook(ctx, article); err != nil {
				return nil, err
			}
		}
	}

	err = svc.ctx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			if *ea.Category == "" {
				// 空字符串表示解除分类关联
				if err := tx.Model(article).Update("category_id", nil).Error; err != nil {
					return err
				}
				article.CategoryID = nil
				article.Category = data.Category{}
			} else {
				category, err := resolveCategory(tx, *ea.Category)
				if err != nil {
					return err
				}

				if err = tx.Model(article).Association("Category").Replace(category); err != nil {
					return err
				}
				article.CategoryID = &category.ID
				article.Category = *category
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if isPublishing(ea.Status) {
		svc.runArticleHooks("after publish", ctx, article, svc.hooks.AfterPublish)
	}

	svc.runArticleHooks("after update info", ctx, article, svc.hooks.AfterUpdateInfo)

	fresh, err := getArticleByID(ctx, id,
		func(tx *gorm.DB) *gorm.DB { return tx.Preload("Tags").Preload("Category") })
	if err != nil {
		return nil, err
	}

	return fresh, nil
}

func (svc *ArticleService) IncrementViews(ctx context.Context, id string) (int, error) {
	article, err := getArticleByID(ctx, id, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("views", "id").
			Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return 0, err
	}

	err = svc.ctx.DB.WithContext(ctx).Model(article).UpdateColumn("views", gorm.Expr("views+?", 1)).Error
	if err != nil {
		return 0, err
	}
	err = svc.ctx.DB.WithContext(ctx).Model(article).Select("views", "id").Find(article).Error
	if err != nil {
		return 0, err
	}

	return article.Views, nil
}

func (svc *ArticleService) IncrementLikes(ctx context.Context, id string) (int, error) {
	article, err := getArticleByID(ctx, id, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("likes", "id").
			Where("status IN ?", []data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return 0, err
	}

	err = svc.ctx.DB.WithContext(ctx).Model(article).UpdateColumn("likes", gorm.Expr("likes+?", 1)).Error
	if err != nil {
		return 0, err
	}

	err = svc.ctx.DB.WithContext(ctx).Model(article).Select("likes", "id").Find(article).Error
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

	err := svc.ctx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var category *data.Category
		if ea.Category != nil && *ea.Category != "" {
			c, err := resolveCategory(tx, *ea.Category)
			if err != nil {
				return err
			}
			category = c
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

		if category != nil {
			ar.Category = *category
		}

		if err := tx.Create(ar).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	svc.runArticleHooks("after create", ctx, ar, svc.hooks.AfterCreate)

	return ar, nil
}

func (svc *ArticleService) Delete(ctx context.Context, id string, isHard bool) error {
	article, err := getArticleByID(ctx, id, data.OnlyID, func(tx *gorm.DB) *gorm.DB {
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

	err = svc.ctx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&data.Article{}).Where("id=?", article.ID)

		if isHard {
			tx = tx.Unscoped()
		}

		return tx.Delete(article).Error
	})

	if err != nil {
		return err
	}

	// 文件系统操作不可回滚，放在事务提交之后执行；
	// 仅硬删除清理文件，软删除保留内容以便 Restore 能恢复。
	// 失败只记录日志，避免"数据已删除但接口返回失败"导致调用方重试。
	if isHard {
		if err := svc.ctx.Repo.Remove(article.ID.String()); err != nil {
			log.Errorf("remove article files failed: articleID=%s, err=%v", article.ID, err)
		}
	}

	for _, hook := range svc.hooks.AfterDelete {
		if err := hook(ctx, article.ID.String(), isHard); err != nil {
			log.Errorf("after delete hook failed: articleID=%s, err=%v", article.ID, err)
		}
	}

	return nil
}

func (svc *ArticleService) Restore(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx := svc.ctx.DB.WithContext(ctx).Model(&data.Article{})
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

	articles, err = data.GetArticles(ctx, func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("Category").Preload("Tags")
		tx = tx.Order(orderBy)
		tx = data.ApplyQueryModifier(tx, modifiers...)
		return tx
	}, data.Pagination(options.Page, options.Limit))

	if err != nil {
		return nil, 0, err
	}

	total, err = data.Total(ctx, data.Article{}, modifiers...)
	if err != nil {
		return nil, 0, err
	}

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

	articles, err = data.GetArticles(ctx, func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("Category").Preload("Tags")
		tx = tx.Order(orderBy)
		tx = data.ApplyQueryModifier(tx, modifiers...)
		return tx
	}, data.Pagination(options.Page, options.Limit))

	if err != nil {
		return nil, 0, err
	}

	total, err = data.Total(ctx, data.Article{}, modifiers...)
	if err != nil {
		return nil, 0, err
	}

	return
}

func (svc *ArticleService) GetArticle(ctx context.Context, id string) (*data.Article, error) {
	return getArticleByID(ctx, id)
}

func (svc *ArticleService) GetContent(ctx context.Context, articleID string) (io.ReadCloser, error) {
	article, err := getArticleByID(ctx, articleID, data.OnlyID)
	if err != nil {
		return nil, err
	}
	return svc.ctx.Repo.ContentReader(article.ID.String())
}

func (svc *ArticleService) UpdateContent(ctx context.Context, articleID string, reader io.Reader) error {
	article, err := getArticleByID(ctx, articleID, data.OnlyID)
	if err != nil {
		return err
	}

	writer, err := svc.ctx.Repo.ContentWriter(article.ID.String())
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	return err
}

func (svc *ArticleService) GetRenderedContent(ctx context.Context, articleID string) ([]byte, error) {
	article, err := getArticleByID(ctx, articleID, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id").Where("status IN ?",
			[]data.ArticleStatus{data.ARTICLE_STATUS_PUBLISHED, data.ARTICLE_STATUS_HIDDEN})
	})
	if err != nil {
		return nil, err
	}

	// fix: 修复自定义ID导致无法读取到文章内容的bug
	reader, err := svc.ctx.Repo.ContentReader(article.ID.String())
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
	article, err := getArticleByID(ctx, articleID, data.OnlyID)
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

	wr, err := svc.ctx.Repo.OpenAsset(article.ID.String(), filename)
	if err != nil {
		return err
	}
	defer wr.Close()

	_, err = io.Copy(wr, file)
	return err
}

func (svc *ArticleService) DeleteAsset(ctx context.Context, articleID, filename string) error {
	article, err := getArticleByID(ctx, articleID, data.OnlyID)
	if err != nil {
		return err
	}
	return svc.ctx.Repo.RemoveAsset(article.ID.String(), filename)
}

func (svc *ArticleService) ListAssets(ctx context.Context, articleID string) ([]string, error) {
	articleEntity, err := getArticleByID(ctx, articleID, data.OnlyID)
	if err != nil {
		return nil, err
	}

	list, err := svc.ctx.Repo.ListAssets(articleEntity.ID.String())
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

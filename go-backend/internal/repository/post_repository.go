package repository

import (
	"tanzanite/internal/domain/post"
	"time"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create 创建文章
func (r *PostRepository) Create(p *post.Post) error {
	return r.db.Create(p).Error
}

// FindByID 根据ID查找文章
func (r *PostRepository) FindByID(id uint) (*post.Post, error) {
	var p post.Post
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySlug 根据slug和语言查找文章
func (r *PostRepository) FindBySlug(slug, locale string) (*post.Post, error) {
	var p post.Post
	err := r.db.Where("slug = ? AND locale = ?", slug, locale).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update 更新文章
func (r *PostRepository) Update(p *post.Post) error {
	return r.db.Save(p).Error
}

// Delete 删除文章（软删除）
func (r *PostRepository) Delete(id uint) error {
	return r.db.Delete(&post.Post{}, id).Error
}

// List 获取文章列表
func (r *PostRepository) List(locale, status string, offset, limit int) ([]post.Post, int64, error) {
	var posts []post.Post
	var total int64

	query := r.db.Model(&post.Post{})

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error
	return posts, total, err
}

// IncrementViewCount 增加浏览次数
func (r *PostRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&post.Post{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// FindTranslations 查找文章的所有翻译版本（已废弃，使用 FindByTranslationGroup）
func (r *PostRepository) FindTranslations(parentID uint) ([]post.Post, error) {
	var posts []post.Post
	err := r.db.Where("parent_id = ? OR id = ?", parentID, parentID).Find(&posts).Error
	return posts, err
}

// FindByTranslationGroup 根据翻译组ID查找所有翻译版本
func (r *PostRepository) FindByTranslationGroup(groupID uint) ([]post.Post, error) {
	var posts []post.Post
	err := r.db.Where("translation_group_id = ?", groupID).Order("locale ASC").Find(&posts).Error
	return posts, err
}

// FindPublished 查找所有已发布的文章
func (r *PostRepository) FindPublished() ([]post.Post, error) {
	var posts []post.Post
	err := r.db.Where("status = ?", "published").Order("published_at DESC").Limit(1000).Find(&posts).Error
	return posts, err
}

// FindPublishedByLocale 查找指定语言的已发布文章
func (r *PostRepository) FindPublishedByLocale(locale string) ([]post.Post, error) {
	var posts []post.Post
	err := r.db.Where("status = ? AND locale = ?", "published", locale).
		Order("published_at DESC").
		Limit(1000).
		Find(&posts).Error
	return posts, err
}

// GetTranslationGroupID 获取文章的翻译组ID
func (r *PostRepository) GetTranslationGroupID(postID uint) (*uint, error) {
	var p post.Post
	err := r.db.Select("translation_group_id").First(&p, postID).Error
	if err != nil {
		return nil, err
	}
	return p.TranslationGroupID, nil
}

// FindAllWithFilters 根据筛选条件获取文章列表
func (r *PostRepository) FindAllWithFilters(page, pageSize int, status, locale, search, authorID string) ([]post.Post, int64, error) {
	var posts []post.Post
	var total int64

	query := r.db.Model(&post.Post{})

	// 应用筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if search != "" {
		query = query.Where("title LIKE ? OR content LIKE ? OR excerpt LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&posts).Error

	return posts, total, err
}

// UpdateStatus 更新文章状态
func (r *PostRepository) UpdateStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 如果状态是已发布，设置发布时间
	if status == "published" {
		var p post.Post
		if err := r.db.First(&p, id).Error; err != nil {
			return err
		}
		// 只有首次发布时设置发布时间
		if p.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = &now
		}
	}

	return r.db.Model(&post.Post{}).Where("id = ?", id).Updates(updates).Error
}

// GetStats 获取文章统计
func (r *PostRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文章数
	var total int64
	if err := r.db.Model(&post.Post{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&post.Post{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range statusStats {
		stats[stat.Status] = stat.Count
	}

	// 按语言统计
	var localeStats []struct {
		Locale string
		Count  int64
	}
	if err := r.db.Model(&post.Post{}).Select("locale, COUNT(*) as count").Group("locale").Scan(&localeStats).Error; err != nil {
		return nil, err
	}

	locales := make(map[string]int64)
	for _, stat := range localeStats {
		locales[stat.Locale] = stat.Count
	}
	stats["locales"] = locales

	// 总浏览量
	var totalViews int64
	if err := r.db.Model(&post.Post{}).Select("COALESCE(SUM(view_count), 0)").Scan(&totalViews).Error; err != nil {
		return nil, err
	}
	stats["total_views"] = totalViews

	return stats, nil
}

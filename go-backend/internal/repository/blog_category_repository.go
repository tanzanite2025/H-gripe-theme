package repository

import (
	"commerce-platform/internal/domain/post"

	"gorm.io/gorm"
)

type BlogCategoryRepository struct {
	db *gorm.DB
}

func NewBlogCategoryRepository(db *gorm.DB) *BlogCategoryRepository {
	return &BlogCategoryRepository{db: db}
}

func (r *BlogCategoryRepository) List(locale string) ([]post.Category, error) {
	var categories []post.Category
	query := r.db.Model(&post.Category{})
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if err := query.Order("sort_order ASC").Order("name ASC").Order("id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *BlogCategoryRepository) FindByID(id uint) (*post.Category, error) {
	var category post.Category
	if err := r.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *BlogCategoryRepository) FindByIDs(ids []uint) ([]post.Category, error) {
	if len(ids) == 0 {
		return []post.Category{}, nil
	}
	var categories []post.Category
	if err := r.db.Where("id IN ?", ids).Order("sort_order ASC").Order("name ASC").Order("id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *BlogCategoryRepository) FindBySlugLocale(slug, locale string) (*post.Category, error) {
	var category post.Category
	if err := r.db.Where("slug = ? AND locale = ?", slug, locale).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *BlogCategoryRepository) ExistsBySlugLocale(slug, locale string, excludeID uint) (bool, error) {
	query := r.db.Model(&post.Category{}).Where("slug = ? AND locale = ?", slug, locale)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BlogCategoryRepository) CountPosts(categoryID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&post.PostCategory{}).Where("category_id = ?", categoryID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BlogCategoryRepository) Create(category *post.Category) error {
	return r.db.Create(category).Error
}

func (r *BlogCategoryRepository) Update(category *post.Category) error {
	return r.db.Model(&post.Category{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
		"name":        category.Name,
		"slug":        category.Slug,
		"description": category.Description,
		"locale":      category.Locale,
		"sort_order":  category.SortOrder,
	}).Error
}

func (r *BlogCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&post.Category{}, id).Error
}

func (r *BlogCategoryRepository) ReplacePostCategories(postID uint, categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", postID).Delete(&post.PostCategory{}).Error; err != nil {
			return err
		}
		if len(categoryIDs) == 0 {
			return nil
		}
		links := make([]post.PostCategory, 0, len(categoryIDs))
		for _, categoryID := range categoryIDs {
			links = append(links, post.PostCategory{
				PostID:     postID,
				CategoryID: categoryID,
			})
		}
		return tx.Create(&links).Error
	})
}

package repository

import (
	"fmt"
	"tanzanite/internal/domain/faq"

	"gorm.io/gorm"
)

// GetCategories 获取所有分类
func (r *FAQRepository) GetCategories(locale string) ([]string, error) {
	var categories []string
	query := r.db.Model(&faq.FAQCategory{}).
		Where("status = ?", "active").
		Where("deleted_at IS NULL")
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	err := query.Distinct("name").Order("name ASC").Pluck("name", &categories).Error
	return categories, err
}

func (r *FAQRepository) ListPages(locale string, includeHidden bool) ([]faq.FAQPage, error) {
	var pages []faq.FAQPage
	query := r.db.Model(&faq.FAQPage{})
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if !includeHidden {
		query = query.Where("status = ?", "active")
	}
	err := query.
		Order("sort_order ASC").
		Order("page_id ASC").
		Find(&pages).Error
	return pages, err
}

func (r *FAQRepository) FindPageByPageIDLocale(pageID, locale string) (*faq.FAQPage, error) {
	var page faq.FAQPage
	err := r.db.Where("page_id = ? AND locale = ?", pageID, locale).First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *FAQRepository) FindPageByRoutePathLocale(routePath, locale string) (*faq.FAQPage, error) {
	var page faq.FAQPage
	err := r.db.
		Where("route_path = ? AND locale = ?", routePath, locale).
		First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *FAQRepository) SavePage(page *faq.FAQPage) error {
	return r.db.Save(page).Error
}

func (r *FAQRepository) CreatePage(page *faq.FAQPage) error {
	return r.db.Create(page).Error
}

func (r *FAQRepository) ListCategories(locale, pageID string, includeHidden bool) ([]faq.FAQCategory, error) {
	var categories []faq.FAQCategory
	query := r.db.Model(&faq.FAQCategory{})
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if pageID != "" {
		query = query.Where("page_id = ?", pageID)
	}
	if !includeHidden {
		query = query.Where("status = ?", "active")
	}
	err := query.
		Order("page_id ASC").
		Order("sort_order ASC").
		Order("category_key ASC").
		Find(&categories).Error
	return categories, err
}

func (r *FAQRepository) FindCategoryByID(id uint) (*faq.FAQCategory, error) {
	var category faq.FAQCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *FAQRepository) FindCategoryByPageKeyLocale(pageID, categoryKey, locale string) (*faq.FAQCategory, error) {
	var category faq.FAQCategory
	err := r.db.
		Where("page_id = ? AND category_key = ? AND locale = ?", pageID, categoryKey, locale).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *FAQRepository) CreateCategory(category *faq.FAQCategory) error {
	return r.db.Create(category).Error
}

func (r *FAQRepository) UpdateCategory(category *faq.FAQCategory, oldPageID, oldCategoryKey, oldLocale string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(category).Error; err != nil {
			return err
		}

		if oldPageID != category.PageID || oldCategoryKey != category.CategoryKey || oldLocale != category.Locale {
			if err := tx.Model(&faq.FAQ{}).
				Where("page_id = ? AND category = ? AND locale = ?", oldPageID, oldCategoryKey, oldLocale).
				Updates(map[string]any{
					"page_id":  category.PageID,
					"category": category.CategoryKey,
					"locale":   category.Locale,
				}).Error; err != nil {
				return fmt.Errorf("sync faqs after category update: %w", err)
			}
		}

		return nil
	})
}

func (r *FAQRepository) DeleteCategory(id uint) error {
	return r.db.Delete(&faq.FAQCategory{}, id).Error
}

func (r *FAQRepository) CountFAQsByCategory(pageID, categoryKey, locale string) (int64, error) {
	var count int64
	err := r.db.Model(&faq.FAQ{}).
		Where("page_id = ? AND category = ? AND locale = ?", pageID, categoryKey, locale).
		Count(&count).Error
	return count, err
}

func (r *FAQRepository) ListFAQCounts(locale string) (map[string]int64, error) {
	type row struct {
		PageID   string
		Category string
		Locale   string
		Count    int64
	}

	var rows []row
	query := r.db.Model(&faq.FAQ{}).
		Select("page_id, category, locale, COUNT(*) AS count").
		Where("deleted_at IS NULL").
		Group("page_id, category, locale")
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, item := range rows {
		counts[faqCountKey(item.PageID, item.Category, item.Locale)] = item.Count
	}
	return counts, nil
}

func faqCountKey(pageID, categoryKey, locale string) string {
	return pageID + "\x00" + categoryKey + "\x00" + locale
}

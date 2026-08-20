package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/visualshowcase"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VisualShowcaseRepository struct {
	db *gorm.DB
}

func NewVisualShowcaseRepository(db *gorm.DB) *VisualShowcaseRepository {
	return &VisualShowcaseRepository{db: db}
}

func (r *VisualShowcaseRepository) ListItems(showcaseKey, locale string, publishedOnly bool) ([]visualshowcase.Item, error) {
	var items []visualshowcase.Item
	query := r.db.Where("showcase_key = ? AND locale = ?", strings.TrimSpace(showcaseKey), strings.TrimSpace(locale))

	if publishedOnly {
		now := time.Now()
		query = query.
			Where("is_published = ?", true).
			Where("(published_from IS NULL OR published_from <= ?)", now).
			Where("(published_until IS NULL OR published_until >= ?)", now)
	}

	err := query.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "desktop_order"}}).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (r *VisualShowcaseRepository) CountItems(showcaseKey, locale string, publishedOnly bool) (int64, error) {
	query := r.db.Model(&visualshowcase.Item{}).
		Where("showcase_key = ? AND locale = ?", strings.TrimSpace(showcaseKey), strings.TrimSpace(locale))

	if publishedOnly {
		now := time.Now()
		query = query.
			Where("is_published = ?", true).
			Where("(published_from IS NULL OR published_from <= ?)", now).
			Where("(published_until IS NULL OR published_until >= ?)", now)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (r *VisualShowcaseRepository) ReplaceItems(showcaseKey, locale string, items []visualshowcase.Item) error {
	key := strings.TrimSpace(showcaseKey)
	normalizedLocale := strings.TrimSpace(locale)

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("showcase_key = ? AND locale = ?", key, normalizedLocale).
			Delete(&visualshowcase.Item{}).Error; err != nil {
			return err
		}

		for index := range items {
			items[index].ID = 0
			items[index].ShowcaseKey = key
			items[index].Locale = normalizedLocale
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *VisualShowcaseRepository) CountItemsByStorageKey(storageKey string) (int64, error) {
	var count int64
	err := r.db.Model(&visualshowcase.Item{}).
		Where("storage_key = ?", strings.TrimSpace(storageKey)).
		Count(&count).Error
	return count, err
}

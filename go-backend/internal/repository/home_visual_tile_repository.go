package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/homevisualtile"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HomeVisualTileRepository struct {
	db *gorm.DB
}

func NewHomeVisualTileRepository(db *gorm.DB) *HomeVisualTileRepository {
	return &HomeVisualTileRepository{db: db}
}

func (r *HomeVisualTileRepository) ListItems(tileSetKey, locale string, publishedOnly bool) ([]homevisualtile.Tile, error) {
	var items []homevisualtile.Tile
	query := r.db.Where("showcase_key = ? AND locale = ?", strings.TrimSpace(tileSetKey), strings.TrimSpace(locale))

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

func (r *HomeVisualTileRepository) CountItems(tileSetKey, locale string, publishedOnly bool) (int64, error) {
	query := r.db.Model(&homevisualtile.Tile{}).
		Where("showcase_key = ? AND locale = ?", strings.TrimSpace(tileSetKey), strings.TrimSpace(locale))

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

func (r *HomeVisualTileRepository) ReplaceItems(tileSetKey, locale string, items []homevisualtile.Tile) error {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("showcase_key = ? AND locale = ?", key, normalizedLocale).
			Delete(&homevisualtile.Tile{}).Error; err != nil {
			return err
		}

		for index := range items {
			items[index].ID = 0
			items[index].TileSetKey = key
			items[index].Locale = normalizedLocale
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *HomeVisualTileRepository) CountItemsByStorageKey(storageKey string) (int64, error) {
	var count int64
	err := r.db.Model(&homevisualtile.Tile{}).
		Where("storage_key = ?", strings.TrimSpace(storageKey)).
		Count(&count).Error
	return count, err
}

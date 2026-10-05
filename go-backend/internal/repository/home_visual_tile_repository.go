package repository

import (
	"errors"
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

func (r *HomeVisualTileRepository) ReplaceItems(
	tileSetKey string,
	locale string,
	items []homevisualtile.Tile,
	afterReplace func(tx *gorm.DB, previous []homevisualtile.Tile) error,
) error {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)

	return r.db.Transaction(func(tx *gorm.DB) error {
		var previous []homevisualtile.Tile
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("showcase_key = ? AND locale = ?", key, normalizedLocale).
			Find(&previous).Error; err != nil {
			return err
		}
		if err := tx.Where("showcase_key = ? AND locale = ?", key, normalizedLocale).
			Delete(&homevisualtile.Tile{}).Error; err != nil {
			return err
		}

		for index := range items {
			items[index].ID = 0
			items[index].TileSetKey = key
			items[index].Locale = normalizedLocale
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		if afterReplace != nil {
			return afterReplace(tx, previous)
		}
		return nil
	})
}

// SaveSingleVisualShowcaseItem updates one active slot without replacing the other slots in the same showcase.
func (r *HomeVisualTileRepository) SaveSingleVisualShowcaseItem(
	tile *homevisualtile.Tile,
	afterSave func(tx *gorm.DB, previous *homevisualtile.Tile) error,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if tile == nil {
		return gorm.ErrInvalidData
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var current homevisualtile.Tile
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("showcase_key = ? AND locale = ? AND desktop_order = ?", tile.TileSetKey, tile.Locale, tile.DesktopOrder).
			First(&current).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var previous *homevisualtile.Tile
		if err == nil {
			previousCopy := current
			previous = &previousCopy
			tile.ID = current.ID
			tile.CreatedAt = current.CreatedAt
		}

		if err := tx.Save(tile).Error; err != nil {
			return err
		}
		if afterSave != nil {
			return afterSave(tx, previous)
		}
		return nil
	})
}

func (r *HomeVisualTileRepository) CountItemsByStorageKey(storageKey string) (int64, error) {
	var count int64
	err := r.db.Model(&homevisualtile.Tile{}).
		Where("storage_key = ?", strings.TrimSpace(storageKey)).
		Count(&count).Error
	return count, err
}

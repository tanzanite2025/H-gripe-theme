package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/homevisualtile"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HomeVisualTileRepository struct {
	db                    *gorm.DB
	validateImageBindings func(*gorm.DB, []homevisualtile.Tile, []homevisualtile.Tile) error
}

var ErrVisualShowcaseTransactionalImageBindingValidationUnavailable = errors.New("transactional visual showcase image binding validation is unavailable")

func NewHomeVisualTileRepository(db *gorm.DB) *HomeVisualTileRepository {
	return &HomeVisualTileRepository{db: db}
}

func (r *HomeVisualTileRepository) WithTx(tx *gorm.DB) *HomeVisualTileRepository {
	clonedValue := *r
	clonedValue.db = tx
	return &clonedValue
}

func (r *HomeVisualTileRepository) ConfigureTransactionalVisualShowcaseImageBindingValidation(validate func(*gorm.DB, []homevisualtile.Tile, []homevisualtile.Tile) error) {
	r.validateImageBindings = validate
}

func (r *HomeVisualTileRepository) WithContext(ctx context.Context) *HomeVisualTileRepository {
	clonedValue := *r
	clonedValue.db = r.db.WithContext(ctx)
	return &clonedValue
}

func (r *HomeVisualTileRepository) WithLockedVisualShowcaseReferencesForObjectCleanup(ctx context.Context, operation func(*HomeVisualTileRepository, *MediaRepository) error) error {
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, dedicatedHomeVisualObjectNamespace, func(tx *gorm.DB) error {
		return operation(NewHomeVisualTileRepository(tx), NewMediaRepository(tx))
	})
}

// VisitActiveVisualShowcaseImageReferences includes drafts for retention. Public
// visibility follows the existing hero contract and scheduled publication rules.
func (r *HomeVisualTileRepository) VisitActiveVisualShowcaseImageReferences(publicOnly bool, visit func(homevisualtile.Tile) bool) error {
	for afterID := uint(0); ; {
		var items []homevisualtile.Tile
		query := r.db.Where("id > ?", afterID)
		if publicOnly {
			now := time.Now()
			query = query.Where("showcase_key = ? OR (is_published = ? AND (published_from IS NULL OR published_from <= ?) AND (published_until IS NULL OR published_until >= ?))", "home-hero", true, now, now)
		}
		if err := query.Order("id ASC").Limit(256).Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			if !visit(item) {
				return nil
			}
		}
		if len(items) < 256 {
			return nil
		}
		afterID = items[len(items)-1].ID
	}
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

	return runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedHomeVisualObjectNamespace, func(tx *gorm.DB) error {
		var previous []homevisualtile.Tile
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("showcase_key = ? AND locale = ?", key, normalizedLocale).
			Order("id ASC").Find(&previous).Error; err != nil {
			return err
		}
		for index := range items {
			items[index].ID = 0
			items[index].TileSetKey = key
			items[index].Locale = normalizedLocale
			items[index].DeletedAt = gorm.DeletedAt{}
		}
		if r.validateImageBindings != nil {
			if err := r.validateImageBindings(tx, items, previous); err != nil {
				return err
			}
		} else if len(items) > 0 {
			return ErrVisualShowcaseTransactionalImageBindingValidationUnavailable
		}
		if err := tx.Where("showcase_key = ? AND locale = ?", key, normalizedLocale).
			Delete(&homevisualtile.Tile{}).Error; err != nil {
			return err
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

	return runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedHomeVisualObjectNamespace, func(tx *gorm.DB) error {
		var current homevisualtile.Tile
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("showcase_key = ? AND locale = ? AND desktop_order = ?", tile.TileSetKey, tile.Locale, tile.DesktopOrder).
			First(&current).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var previous *homevisualtile.Tile
		var previousItems []homevisualtile.Tile
		if err == nil {
			previousCopy := current
			previous = &previousCopy
			tile.ID = current.ID
			tile.CreatedAt = current.CreatedAt
			previousItems = append(previousItems, current)
		} else {
			tile.ID = 0
			tile.CreatedAt = time.Time{}
		}

		tile.DeletedAt = gorm.DeletedAt{}
		if r.validateImageBindings != nil {
			if err := r.validateImageBindings(tx, []homevisualtile.Tile{*tile}, previousItems); err != nil {
				return err
			}
		} else {
			return ErrVisualShowcaseTransactionalImageBindingValidationUnavailable
		}
		if previous == nil {
			if err := tx.Create(tile).Error; err != nil {
				return err
			}
		} else {
			result := tx.Model(&homevisualtile.Tile{}).Where("id = ?", current.ID).
				Select("ImageURL", "ThumbnailURL", "StorageKey", "Title", "Caption", "AltText", "DesktopOrder", "MobilePairIndex", "TargetURL", "TargetLabel", "LayoutVariant", "IsPublished", "PublishedFrom", "PublishedUntil", "Width", "Height").Updates(tile)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
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

package repository

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
)

type YanwenProductCatalogSyncStats struct {
	Scanned int
	Added   int
	Updated int
}

type YanwenProductCatalogRepository struct{ db *gorm.DB }

func NewYanwenProductCatalogRepository(db *gorm.DB) *YanwenProductCatalogRepository {
	return &YanwenProductCatalogRepository{db: db}
}

func (r *YanwenProductCatalogRepository) FindYanwenProductCatalogEntriesByEnvironment(environment string) ([]shipping.YanwenProductCatalogEntry, error) {
	var entries []shipping.YanwenProductCatalogEntry
	err := r.db.Where("environment = ?", environment).
		Order("name_ch ASC").
		Order("name_en ASC").
		Order("product_id ASC").
		Find(&entries).Error
	return entries, err
}

func (r *YanwenProductCatalogRepository) UpsertYanwenProductCatalogEntries(
	environment string,
	entries []shipping.YanwenProductCatalogEntry,
	syncedAt time.Time,
) (YanwenProductCatalogSyncStats, error) {
	stats := YanwenProductCatalogSyncStats{Scanned: len(entries)}
	if r == nil || r.db == nil {
		return stats, errors.New("Yanwen product catalog repository is not configured")
	}
	if syncedAt.IsZero() {
		syncedAt = time.Now().UTC()
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for i := range entries {
			entry := entries[i]
			entry.Environment = environment
			entry.LastSyncedAt = syncedAt.UTC()
			if err := entry.Validate(); err != nil {
				return err
			}
			var existing shipping.YanwenProductCatalogEntry
			err := tx.Where("environment = ? AND product_id = ?", environment, entry.ProductID).First(&existing).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				if err := tx.Create(&entry).Error; err != nil {
					return err
				}
				stats.Added++
			case err != nil:
				return err
			default:
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"name_ch":        entry.NameChinese,
					"name_en":        entry.NameEnglish,
					"last_synced_at": entry.LastSyncedAt,
					"updated_at":     time.Now().UTC(),
				}).Error; err != nil {
					return err
				}
				stats.Updated++
			}
		}
		staleProductsQuery := tx.Where("environment = ?", environment)
		if len(entries) > 0 {
			staleProductsQuery = staleProductsQuery.Where("last_synced_at < ?", syncedAt.UTC())
		}
		if err := staleProductsQuery.Delete(&shipping.YanwenProductCatalogEntry{}).Error; err != nil {
			return err
		}
		return nil
	})
	return stats, err
}

package repository

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
)

type YanwenWarehouseCatalogSyncStats struct {
	Scanned int
	Added   int
	Updated int
}

type YanwenWarehouseCatalogRepository struct{ db *gorm.DB }

func NewYanwenWarehouseCatalogRepository(db *gorm.DB) *YanwenWarehouseCatalogRepository {
	return &YanwenWarehouseCatalogRepository{db: db}
}

func (r *YanwenWarehouseCatalogRepository) FindYanwenWarehouseCatalogEntriesByEnvironment(environment string) ([]shipping.YanwenWarehouseCatalogEntry, error) {
	var entries []shipping.YanwenWarehouseCatalogEntry
	err := r.db.Where("environment = ?", environment).
		Order("area ASC").
		Order("name ASC").
		Order("warehouse_code ASC").
		Find(&entries).Error
	return entries, err
}

func (r *YanwenWarehouseCatalogRepository) UpsertYanwenWarehouseCatalogEntries(
	environment string,
	entries []shipping.YanwenWarehouseCatalogEntry,
	syncedAt time.Time,
) (YanwenWarehouseCatalogSyncStats, error) {
	stats := YanwenWarehouseCatalogSyncStats{Scanned: len(entries)}
	if r == nil || r.db == nil {
		return stats, errors.New("Yanwen warehouse catalog repository is not configured")
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
			var existing shipping.YanwenWarehouseCatalogEntry
			err := tx.Where("environment = ? AND warehouse_code = ?", environment, entry.WarehouseCode).First(&existing).Error
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
					"name":           entry.Name,
					"area":           entry.Area,
					"last_synced_at": entry.LastSyncedAt,
					"updated_at":     time.Now().UTC(),
				}).Error; err != nil {
					return err
				}
				stats.Updated++
			}
		}
		staleWarehousesQuery := tx.Where("environment = ?", environment)
		if len(entries) > 0 {
			staleWarehousesQuery = staleWarehousesQuery.Where("last_synced_at < ?", syncedAt.UTC())
		}
		if err := staleWarehousesQuery.Delete(&shipping.YanwenWarehouseCatalogEntry{}).Error; err != nil {
			return err
		}
		return nil
	})
	return stats, err
}

package repository

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
)

type YanwenCountryCatalogSyncStats struct {
	Scanned int
	Added   int
	Updated int
}

type YanwenCountryCatalogRepository struct{ db *gorm.DB }

func NewYanwenCountryCatalogRepository(db *gorm.DB) *YanwenCountryCatalogRepository {
	return &YanwenCountryCatalogRepository{db: db}
}

func (r *YanwenCountryCatalogRepository) FindYanwenCountryCatalogEntriesByEnvironment(environment string) ([]shipping.YanwenCountryCatalogEntry, error) {
	var entries []shipping.YanwenCountryCatalogEntry
	err := r.db.Where("environment = ?", environment).
		Order("name_ch ASC").
		Order("name_en ASC").
		Order("country_code ASC").
		Order("country_id ASC").
		Find(&entries).Error
	return entries, err
}

func (r *YanwenCountryCatalogRepository) UpsertYanwenCountryCatalogEntries(
	environment string,
	entries []shipping.YanwenCountryCatalogEntry,
	syncedAt time.Time,
) (YanwenCountryCatalogSyncStats, error) {
	stats := YanwenCountryCatalogSyncStats{Scanned: len(entries)}
	if r == nil || r.db == nil {
		return stats, errors.New("Yanwen country catalog repository is not configured")
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
			var existing shipping.YanwenCountryCatalogEntry
			err := tx.Where("environment = ? AND country_id = ?", environment, entry.CountryID).First(&existing).Error
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
					"country_code":   entry.CountryCode,
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
		staleCountriesQuery := tx.Where("environment = ?", environment)
		if len(entries) > 0 {
			staleCountriesQuery = staleCountriesQuery.Where("last_synced_at < ?", syncedAt.UTC())
		}
		if err := staleCountriesQuery.Delete(&shipping.YanwenCountryCatalogEntry{}).Error; err != nil {
			return err
		}
		return nil
	})
	return stats, err
}

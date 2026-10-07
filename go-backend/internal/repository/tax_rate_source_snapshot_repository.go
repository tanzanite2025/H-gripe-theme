package repository

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/taxrate"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TaxRateSourceSnapshotRepository stores provider configuration and immutable
// reference snapshots without reading or writing the checkout tax_rates table.
type TaxRateSourceSnapshotRepository struct {
	db *gorm.DB
}

func NewTaxRateSourceSnapshotRepository(db *gorm.DB) *TaxRateSourceSnapshotRepository {
	return &TaxRateSourceSnapshotRepository{db: db}
}

func (r *TaxRateSourceSnapshotRepository) GetOrCreateSourceConfig() (*taxrate.TaxRateSourceConfig, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("tax rate source snapshot repository is not configured")
	}
	defaultConfig := taxrate.TaxRateSourceConfig{
		ID:                   1,
		ProviderCode:         taxrate.DefaultTaxRateSourceProviderCode,
		RefreshIntervalHours: taxrate.DefaultTaxRateSourceRefreshIntervalHours,
	}
	if err := r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&defaultConfig).Error; err != nil {
		return nil, err
	}
	var config taxrate.TaxRateSourceConfig
	if err := r.db.First(&config, 1).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *TaxRateSourceSnapshotRepository) UpdateSourceConfig(enabled bool, refreshIntervalHours int) (*taxrate.TaxRateSourceConfig, error) {
	if _, err := r.GetOrCreateSourceConfig(); err != nil {
		return nil, err
	}
	if err := r.db.Model(&taxrate.TaxRateSourceConfig{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"enabled":                enabled,
		"refresh_interval_hours": refreshIntervalHours,
		"updated_at":             time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}
	return r.GetOrCreateSourceConfig()
}

func (r *TaxRateSourceSnapshotRepository) RecordSyncStarted(at time.Time) error {
	return r.updateSourceSyncStatus(map[string]interface{}{
		"last_checked_at": at.UTC(),
		"last_sync_error": "",
		"updated_at":      at.UTC(),
	})
}

func (r *TaxRateSourceSnapshotRepository) RecordSyncFailure(at time.Time, syncError string) error {
	return r.updateSourceSyncStatus(map[string]interface{}{
		"last_checked_at": at.UTC(),
		"last_sync_error": strings.TrimSpace(syncError),
		"updated_at":      at.UTC(),
	})
}

func (r *TaxRateSourceSnapshotRepository) RecordSyncSuccess(at time.Time) error {
	return r.updateSourceSyncStatus(map[string]interface{}{
		"last_checked_at":         at.UTC(),
		"last_successful_sync_at": at.UTC(),
		"last_sync_error":         "",
		"updated_at":              at.UTC(),
	})
}

func (r *TaxRateSourceSnapshotRepository) updateSourceSyncStatus(updates map[string]interface{}) error {
	if r == nil || r.db == nil {
		return errors.New("tax rate source snapshot repository is not configured")
	}
	if _, err := r.GetOrCreateSourceConfig(); err != nil {
		return err
	}
	result := r.db.Model(&taxrate.TaxRateSourceConfig{}).Where("id = ?", 1).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("tax rate source configuration row is missing")
	}
	return nil
}

func (r *TaxRateSourceSnapshotRepository) FindLatestSnapshot(providerCode string) (*taxrate.TaxRateSourceSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("tax rate source snapshot repository is not configured")
	}
	var snapshot taxrate.TaxRateSourceSnapshot
	err := r.db.Where("provider_code = ?", strings.TrimSpace(providerCode)).
		Order("captured_at DESC, id DESC").
		First(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *TaxRateSourceSnapshotRepository) ListSnapshotEntries(snapshotID uint) ([]taxrate.TaxRateSourceSnapshotEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("tax rate source snapshot repository is not configured")
	}
	var entries []taxrate.TaxRateSourceSnapshotEntry
	err := r.db.Where("snapshot_id = ?", snapshotID).
		Order("country_code ASC, rate_type ASC, rate_category ASC, rate_decimal ASC").
		Find(&entries).Error
	return entries, err
}

// CreateSnapshotWhenContentChanges serializes publication against the
// singleton configuration row and avoids adding a new version when the latest
// normalized provider data has not changed.
func (r *TaxRateSourceSnapshotRepository) CreateSnapshotWhenContentChanges(
	snapshot *taxrate.TaxRateSourceSnapshot,
	entries []taxrate.TaxRateSourceSnapshotEntry,
) (*taxrate.TaxRateSourceSnapshot, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, errors.New("tax rate source snapshot repository is not configured")
	}
	if snapshot == nil || len(entries) == 0 {
		return nil, false, errors.New("tax rate source snapshot and entries are required")
	}
	var published taxrate.TaxRateSourceSnapshot
	created := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var config taxrate.TaxRateSourceConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&config, 1).Error; err != nil {
			return err
		}

		var latest taxrate.TaxRateSourceSnapshot
		latestErr := tx.Where("provider_code = ?", snapshot.ProviderCode).
			Order("captured_at DESC, id DESC").
			First(&latest).Error
		if latestErr != nil && !errors.Is(latestErr, gorm.ErrRecordNotFound) {
			return latestErr
		}
		if latestErr == nil && latest.ContentSHA256 == snapshot.ContentSHA256 {
			published = latest
			return nil
		}

		if err := tx.Create(snapshot).Error; err != nil {
			return err
		}
		for index := range entries {
			entries[index].SnapshotID = snapshot.ID
		}
		if err := tx.CreateInBatches(&entries, 200).Error; err != nil {
			return err
		}
		published = *snapshot
		created = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return &published, created, nil
}

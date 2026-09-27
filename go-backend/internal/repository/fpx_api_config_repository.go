package repository

import (
	"commerce-platform/internal/domain/shipping"
	"errors"
	"gorm.io/gorm"
	"time"
)

type FpxAPIConfigRepository struct{ db *gorm.DB }

func NewFpxAPIConfigRepository(db *gorm.DB) *FpxAPIConfigRepository {
	return &FpxAPIConfigRepository{db: db}
}
func (r *FpxAPIConfigRepository) FindByEnvironment(environment string) (*shipping.FpxAPIConfig, error) {
	var config shipping.FpxAPIConfig
	if err := r.db.Where("environment = ?", environment).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
func (r *FpxAPIConfigRepository) Save(config *shipping.FpxAPIConfig) error {
	var existing shipping.FpxAPIConfig
	err := r.db.Where("environment = ?", config.Environment).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(config).Error
	}
	if err != nil {
		return err
	}
	config.ID = existing.ID
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"endpoint":                    config.Endpoint,
		"app_key_encrypted":           config.AppKeyEncrypted,
		"app_secret_encrypted":        config.AppSecretEncrypted,
		"access_token_encrypted":      config.AccessTokenEncrypted,
		"enabled":                     config.Enabled,
		"last_sync_status":            config.LastSyncStatus,
		"last_synced_at":              config.LastSyncedAt,
		"last_error":                  config.LastError,
		"last_sync_scanned":           config.LastSyncScanned,
		"last_sync_added":             config.LastSyncAdded,
		"last_sync_updated":           config.LastSyncUpdated,
		"last_sync_preserved_enabled": config.LastSyncPreservedEnabled,
		"updated_at":                  time.Now().UTC(),
	}).Error
}

func (r *FpxAPIConfigRepository) RecordSyncSuccess(environment string, syncedAt time.Time, stats FpxChannelUpsertStats) error {
	return r.db.Model(&shipping.FpxAPIConfig{}).
		Where("environment = ?", environment).
		Updates(map[string]interface{}{
			"last_sync_status":            "success",
			"last_synced_at":              syncedAt,
			"last_error":                  "",
			"last_sync_scanned":           stats.Scanned,
			"last_sync_added":             stats.Added,
			"last_sync_updated":           stats.Updated,
			"last_sync_preserved_enabled": stats.PreservedEnabled,
			"updated_at":                  time.Now().UTC(),
		}).Error
}

func (r *FpxAPIConfigRepository) RecordSyncFailure(environment string, message string) error {
	return r.db.Model(&shipping.FpxAPIConfig{}).
		Where("environment = ?", environment).
		Updates(map[string]interface{}{
			"last_sync_status": "failed",
			"last_error":       message,
			"updated_at":       time.Now().UTC(),
		}).Error
}

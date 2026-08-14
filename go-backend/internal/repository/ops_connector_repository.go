package repository

import (
	"commerce-platform/internal/domain/ops"
	"time"

	"gorm.io/gorm"
)

type OpsConnectorRepository struct {
	db *gorm.DB
}

func NewOpsConnectorRepository(db *gorm.DB) *OpsConnectorRepository {
	return &OpsConnectorRepository{db: db}
}

func (r *OpsConnectorRepository) List() ([]ops.Connector, error) {
	var records []ops.Connector
	err := r.db.
		Order("enabled DESC").
		Order("provider ASC").
		Order("environment ASC").
		Order("name ASC").
		Find(&records).Error
	return records, err
}

func (r *OpsConnectorRepository) FindByID(id uint) (*ops.Connector, error) {
	var record ops.Connector
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsConnectorRepository) FindByName(name string) (*ops.Connector, error) {
	var record ops.Connector
	if err := r.db.Where("name = ?", name).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsConnectorRepository) Create(record *ops.Connector) error {
	return r.db.Create(record).Error
}

func (r *OpsConnectorRepository) Update(record *ops.Connector) error {
	updates := map[string]interface{}{
		"name":                  record.Name,
		"provider":              record.Provider,
		"environment":           record.Environment,
		"endpoint":              record.Endpoint,
		"auth_type":             record.AuthType,
		"credential_ref":        record.CredentialRef,
		"credentials_encrypted": record.CredentialsEncrypted,
		"credential_fields":     record.CredentialFields,
		"scopes":                record.Scopes,
		"status":                record.Status,
		"enabled":               record.Enabled,
		"last_test_status":      record.LastTestStatus,
		"last_tested_at":        record.LastTestedAt,
		"last_error":            record.LastError,
		"notes":                 record.Notes,
	}
	return r.db.Model(&ops.Connector{}).Where("id = ?", record.ID).Updates(updates).Error
}

func (r *OpsConnectorRepository) UpdateEnabled(id uint, enabled bool, status string) error {
	return r.db.Model(&ops.Connector{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled": enabled,
		"status":  status,
	}).Error
}

func (r *OpsConnectorRepository) UpdateTestState(id uint, status string, testedAt *time.Time, connectorStatus, lastError string) error {
	return r.db.Model(&ops.Connector{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_test_status": status,
		"last_tested_at":   testedAt,
		"status":           connectorStatus,
		"last_error":       lastError,
	}).Error
}

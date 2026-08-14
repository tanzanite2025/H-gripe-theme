package repository

import (
	"commerce-platform/internal/domain/ops"
	"time"

	"gorm.io/gorm"
)

type OpsVPSBindingRepository struct {
	db *gorm.DB
}

func NewOpsVPSBindingRepository(db *gorm.DB) *OpsVPSBindingRepository {
	return &OpsVPSBindingRepository{db: db}
}

func (r *OpsVPSBindingRepository) List() ([]ops.VPSBinding, error) {
	var records []ops.VPSBinding
	err := r.db.
		Order("enabled DESC").
		Order("environment ASC").
		Order("name ASC").
		Find(&records).Error
	return records, err
}

func (r *OpsVPSBindingRepository) FindByID(id uint) (*ops.VPSBinding, error) {
	var record ops.VPSBinding
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsVPSBindingRepository) FindByName(name string) (*ops.VPSBinding, error) {
	var record ops.VPSBinding
	if err := r.db.Where("name = ?", name).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsVPSBindingRepository) Create(record *ops.VPSBinding) error {
	return r.db.Create(record).Error
}

func (r *OpsVPSBindingRepository) Update(record *ops.VPSBinding) error {
	updates := map[string]interface{}{
		"name":                 record.Name,
		"provider":             record.Provider,
		"environment":          record.Environment,
		"connector_id":         record.ConnectorID,
		"provider_resource_id": record.ProviderResourceID,
		"hostname":             record.Hostname,
		"ipv4":                 record.IPv4,
		"region":               record.Region,
		"operating_system":     record.OperatingSystem,
		"status":               record.Status,
		"enabled":              record.Enabled,
		"notes":                record.Notes,
	}
	return r.db.Model(&ops.VPSBinding{}).Where("id = ?", record.ID).Updates(updates).Error
}

func (r *OpsVPSBindingRepository) UpdateEnabled(id uint, enabled bool, status string) error {
	return r.db.Model(&ops.VPSBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled": enabled,
		"status":  status,
	}).Error
}

func (r *OpsVPSBindingRepository) UpdateObservedState(
	id uint,
	observedStatus string,
	observedState string,
	observedSource string,
	observedHostname string,
	observedIPv4 string,
	observedOperatingSystem string,
	observedPlan string,
	observedRegion string,
	observedAt time.Time,
	lastError string,
) error {
	return r.db.Model(&ops.VPSBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"observed_status":           observedStatus,
		"observed_state":            observedState,
		"observed_source":           observedSource,
		"observed_hostname":         observedHostname,
		"observed_ipv4":             observedIPv4,
		"observed_operating_system": observedOperatingSystem,
		"observed_plan":             observedPlan,
		"observed_region":           observedRegion,
		"last_observed_at":          observedAt,
		"last_error":                lastError,
	}).Error
}

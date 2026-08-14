package repository

import (
	"commerce-platform/internal/domain/ops"
	"time"

	"gorm.io/gorm"
)

type OpsDomainBindingRepository struct {
	db *gorm.DB
}

func NewOpsDomainBindingRepository(db *gorm.DB) *OpsDomainBindingRepository {
	return &OpsDomainBindingRepository{db: db}
}

func (r *OpsDomainBindingRepository) List() ([]ops.DomainBinding, error) {
	var records []ops.DomainBinding
	err := r.db.Order("enabled DESC").Order("environment ASC").Order("role ASC").Order("domain ASC").Find(&records).Error
	return records, err
}

func (r *OpsDomainBindingRepository) FindByID(id uint) (*ops.DomainBinding, error) {
	var record ops.DomainBinding
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsDomainBindingRepository) FindByDomain(domain string) (*ops.DomainBinding, error) {
	var record ops.DomainBinding
	if err := r.db.Where("domain = ?", domain).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsDomainBindingRepository) ListByProjectID(projectID uint) ([]ops.DomainBinding, error) {
	var records []ops.DomainBinding
	query := r.db.Order("domain ASC")
	if projectID > 0 {
		query = query.Where("project_binding_id = ?", projectID)
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *OpsDomainBindingRepository) Create(record *ops.DomainBinding) error {
	return r.db.Create(record).Error
}

func (r *OpsDomainBindingRepository) Update(record *ops.DomainBinding) error {
	updates := map[string]interface{}{
		"domain":             record.Domain,
		"connector_id":       record.ConnectorID,
		"project_binding_id": record.ProjectBindingID,
		"role":               record.Role,
		"environment":        record.Environment,
		"provider":           record.Provider,
		"zone":               record.Zone,
		"target":             record.Target,
		"proxy_mode":         record.ProxyMode,
		"tls_mode":           record.TLSMode,
		"redirect_target":    record.RedirectTarget,
		"status":             record.Status,
		"enabled":            record.Enabled,
		"notes":              record.Notes,
	}
	return r.db.Model(&ops.DomainBinding{}).Where("id = ?", record.ID).Updates(updates).Error
}

func (r *OpsDomainBindingRepository) UpdateObservedState(
	id uint,
	observedStatus string,
	observedTarget string,
	observedProxyMode string,
	observedTLSMode string,
	observedSource string,
	observedAt time.Time,
	observedError string,
) error {
	return r.db.Model(&ops.DomainBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"observed_status":     observedStatus,
		"observed_target":     observedTarget,
		"observed_proxy_mode": observedProxyMode,
		"observed_tls_mode":   observedTLSMode,
		"observed_source":     observedSource,
		"last_observed_at":    observedAt,
		"observed_error":      observedError,
	}).Error
}

func (r *OpsDomainBindingRepository) UpdateEnabled(id uint, enabled bool, status string) error {
	return r.db.Model(&ops.DomainBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled": enabled,
		"status":  status,
	}).Error
}

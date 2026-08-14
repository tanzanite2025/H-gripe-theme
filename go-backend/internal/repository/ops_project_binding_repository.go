package repository

import (
	"commerce-platform/internal/domain/ops"
	"time"

	"gorm.io/gorm"
)

type OpsProjectBindingRepository struct {
	db *gorm.DB
}

func NewOpsProjectBindingRepository(db *gorm.DB) *OpsProjectBindingRepository {
	return &OpsProjectBindingRepository{db: db}
}

func (r *OpsProjectBindingRepository) List() ([]ops.ProjectBindingView, error) {
	var records []ops.ProjectBindingView
	err := r.db.
		Table("ops_project_bindings AS project").
		Select("project.*, vps.name AS vps_name, vps.provider AS vps_provider, vps.hostname AS vps_hostname, vps.ipv4 AS vps_ipv4").
		Joins("LEFT JOIN ops_vps_bindings AS vps ON vps.id = project.vps_binding_id AND vps.deleted_at IS NULL").
		Where("project.deleted_at IS NULL").
		Order("project.enabled DESC").
		Order("project.environment ASC").
		Order("project.name ASC").
		Scan(&records).Error
	return records, err
}

func (r *OpsProjectBindingRepository) FindByID(id uint) (*ops.ProjectBindingView, error) {
	var record ops.ProjectBindingView
	err := r.db.
		Table("ops_project_bindings AS project").
		Select("project.*, vps.name AS vps_name, vps.provider AS vps_provider, vps.hostname AS vps_hostname, vps.ipv4 AS vps_ipv4").
		Joins("LEFT JOIN ops_vps_bindings AS vps ON vps.id = project.vps_binding_id AND vps.deleted_at IS NULL").
		Where("project.id = ? AND project.deleted_at IS NULL", id).
		Scan(&record).Error
	if err != nil {
		return nil, err
	}
	if record.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &record, nil
}

func (r *OpsProjectBindingRepository) FindByName(name string) (*ops.ProjectBindingView, error) {
	var record ops.ProjectBindingView
	err := r.db.
		Table("ops_project_bindings AS project").
		Select("project.*, vps.name AS vps_name, vps.provider AS vps_provider, vps.hostname AS vps_hostname, vps.ipv4 AS vps_ipv4").
		Joins("LEFT JOIN ops_vps_bindings AS vps ON vps.id = project.vps_binding_id AND vps.deleted_at IS NULL").
		Where("project.name = ? AND project.deleted_at IS NULL", name).
		Scan(&record).Error
	if err != nil {
		return nil, err
	}
	if record.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &record, nil
}

func (r *OpsProjectBindingRepository) Create(record *ops.ProjectBinding) error {
	return r.db.Create(record).Error
}

func (r *OpsProjectBindingRepository) Update(record *ops.ProjectBinding) error {
	updates := map[string]interface{}{
		"name":                        record.Name,
		"vps_binding_id":              record.VPSBindingID,
		"connector_id":                record.ConnectorID,
		"provider_resource_id":        record.ProviderResourceID,
		"environment":                 record.Environment,
		"compose_source":              record.ComposeSource,
		"compose_project_name":        record.ComposeProjectName,
		"gateway_network":             record.GatewayNetwork,
		"gateway_alias":               record.GatewayAlias,
		"services":                    record.Services,
		"networks":                    record.Networks,
		"volumes":                     record.Volumes,
		"current_image_tag":           record.CurrentImageTag,
		"current_commit_sha":          record.CurrentCommitSHA,
		"status":                      record.Status,
		"enabled":                     record.Enabled,
		"last_deployment_at":          record.LastDeploymentAt,
		"backup_policy":               record.BackupPolicy,
		"restore_notes":               record.RestoreNotes,
		"quick_buy_rate_limit_policy": record.QuickBuyRateLimitPolicy,
		"notes":                       record.Notes,
	}
	return r.db.Model(&ops.ProjectBinding{}).Where("id = ?", record.ID).Updates(updates).Error
}

func (r *OpsProjectBindingRepository) UpdateEnabled(id uint, enabled bool, status string) error {
	return r.db.Model(&ops.ProjectBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled": enabled,
		"status":  status,
	}).Error
}

func (r *OpsProjectBindingRepository) UpdateObservedState(
	id uint,
	healthStatus string,
	observedState string,
	observedSource string,
	containerCount int,
	runningContainerCount int,
	healthyContainerCount int,
	checkedAt time.Time,
	lastError string,
) error {
	return r.db.Model(&ops.ProjectBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"health_status":                    healthStatus,
		"observed_state":                   observedState,
		"observed_source":                  observedSource,
		"observed_container_count":         containerCount,
		"observed_running_container_count": runningContainerCount,
		"observed_healthy_container_count": healthyContainerCount,
		"last_checked_at":                  checkedAt,
		"last_error":                       lastError,
	}).Error
}

func (r *OpsProjectBindingRepository) RecordDeployment(id uint, deployedAt time.Time) error {
	return r.db.Model(&ops.ProjectBinding{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_deployment_at": deployedAt,
	}).Error
}

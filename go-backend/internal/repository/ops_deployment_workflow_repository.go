package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"

	"gorm.io/gorm"
)

var ErrOpsDeploymentWorkflowLockHeld = errors.New("operations deployment workflow resource is locked")

type OpsDeploymentWorkflowRepository struct {
	db *gorm.DB
}

func NewOpsDeploymentWorkflowRepository(db *gorm.DB) *OpsDeploymentWorkflowRepository {
	return &OpsDeploymentWorkflowRepository{db: db}
}

func (r *OpsDeploymentWorkflowRepository) List(projectID uint, limit int) ([]ops.DeploymentWorkflowRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	query := r.db.Order("created_at DESC").Limit(limit)
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	var records []ops.DeploymentWorkflowRun
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	for index := range records {
		if err := r.hydrate(&records[index]); err != nil {
			return nil, err
		}
	}
	return records, nil
}

func (r *OpsDeploymentWorkflowRepository) FindByID(id uint) (*ops.DeploymentWorkflowRun, error) {
	var record ops.DeploymentWorkflowRun
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	if err := r.hydrate(&record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OpsDeploymentWorkflowRepository) Create(record *ops.DeploymentWorkflowRun, steps []ops.DeploymentWorkflowStep) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		for index := range steps {
			steps[index].WorkflowRunID = record.ID
		}
		if len(steps) > 0 {
			if err := tx.Create(&steps).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OpsDeploymentWorkflowRepository) UpdateRun(id uint, updates map[string]interface{}) error {
	return r.db.Model(&ops.DeploymentWorkflowRun{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OpsDeploymentWorkflowRepository) UpdateStep(id uint, updates map[string]interface{}) error {
	return r.db.Model(&ops.DeploymentWorkflowStep{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OpsDeploymentWorkflowRepository) AcquireProjectLock(projectID, workflowID uint, ttl time.Duration) error {
	if projectID == 0 || workflowID == 0 {
		return errors.New("project and workflow ids are required for deployment lock")
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	now := time.Now().UTC()
	lock := &ops.DeploymentWorkflowLock{
		ResourceKey:   fmt.Sprintf("project:%d", projectID),
		WorkflowRunID: workflowID,
		AcquiredAt:    now,
		ExpiresAt:     now.Add(ttl),
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at <= ?", now).Delete(&ops.DeploymentWorkflowLock{}).Error; err != nil {
			return err
		}
		return tx.Create(lock).Error
	})
	if err == nil {
		return nil
	}
	if isUniqueConstraintError(err) {
		return ErrOpsDeploymentWorkflowLockHeld
	}
	return err
}

func (r *OpsDeploymentWorkflowRepository) ReleaseWorkflowLocks(workflowID uint) error {
	if workflowID == 0 {
		return nil
	}
	return r.db.Where("workflow_run_id = ?", workflowID).Delete(&ops.DeploymentWorkflowLock{}).Error
}

func (r *OpsDeploymentWorkflowRepository) hydrate(record *ops.DeploymentWorkflowRun) error {
	if record == nil {
		return nil
	}
	var steps []ops.DeploymentWorkflowStep
	if err := r.db.Where("workflow_run_id = ?", record.ID).Order("sequence ASC").Find(&steps).Error; err != nil {
		return err
	}
	record.Steps = steps
	if record.PreflightSnapshot != "" {
		var preflight ops.DeploymentPreflight
		if err := json.Unmarshal([]byte(record.PreflightSnapshot), &preflight); err != nil {
			return err
		}
		record.Preflight = &preflight
	}
	return nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "already exists")
}

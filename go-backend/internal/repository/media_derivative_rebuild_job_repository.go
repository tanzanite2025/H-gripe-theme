package repository

import (
	"errors"
	"strings"
	"time"

	mediadomain "commerce-platform/internal/domain/media"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrMediaDerivativeRebuildLeaseLost = errors.New("media derivative rebuild job lease was lost")

type MediaDerivativeRebuildJobRepository struct {
	db *gorm.DB
}

func NewMediaDerivativeRebuildJobRepository(db *gorm.DB) *MediaDerivativeRebuildJobRepository {
	return &MediaDerivativeRebuildJobRepository{db: db}
}

func (r *MediaDerivativeRebuildJobRepository) WithTx(tx *gorm.DB) *MediaDerivativeRebuildJobRepository {
	return &MediaDerivativeRebuildJobRepository{db: tx}
}

// Enqueue coalesces pending configuration changes while allowing a currently
// running full pass to finish. A later pending pass then revisits earlier
// assets with the newest active preset contract.
func (r *MediaDerivativeRebuildJobRepository) Enqueue(
	reason string,
	now time.Time,
) (*mediadomain.MediaDerivativeRebuildJob, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, gorm.ErrInvalidDB
	}
	var job *mediadomain.MediaDerivativeRebuildJob
	var created bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		job, created, err = r.WithTx(tx).EnqueueInTx(reason, now)
		return err
	})
	return job, created, err
}

func (r *MediaDerivativeRebuildJobRepository) EnqueueInTx(
	reason string,
	now time.Time,
) (*mediadomain.MediaDerivativeRebuildJob, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "preset_changed"
	}

	var existing mediadomain.MediaDerivativeRebuildJob
	err := r.db.
		Where("status = ?", mediadomain.MediaDerivativeRebuildJobStatusPending).
		Order("id ASC").
		First(&existing).Error
	if err == nil {
		if err := r.db.Model(&existing).Updates(map[string]interface{}{
			"reason":     reason,
			"updated_at": now,
		}).Error; err != nil {
			return nil, false, err
		}
		existing.Reason = reason
		existing.UpdatedAt = now
		return &existing, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	created := mediadomain.MediaDerivativeRebuildJob{
		Reason: reason,
		Status: mediadomain.MediaDerivativeRebuildJobStatusPending,
	}
	if err := r.db.Create(&created).Error; err != nil {
		if !isMediaDerivativeRebuildPendingConflict(err) {
			return nil, false, err
		}
		if err := r.db.
			Where("status = ?", mediadomain.MediaDerivativeRebuildJobStatusPending).
			Order("id ASC").
			First(&existing).Error; err != nil {
			return nil, false, err
		}
		return &existing, false, nil
	}
	return &created, true, nil
}

func (r *MediaDerivativeRebuildJobRepository) ListRecent(limit int) ([]mediadomain.MediaDerivativeRebuildJob, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	items := make([]mediadomain.MediaDerivativeRebuildJob, 0, limit)
	err := r.db.Order("id DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *MediaDerivativeRebuildJobRepository) ClaimNext(
	now time.Time,
	workerID string,
	leaseTimeout time.Duration,
) (*mediadomain.MediaDerivativeRebuildJob, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, errors.New("media derivative rebuild worker ID is required")
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 15 * time.Minute
	}

	var claimed *mediadomain.MediaDerivativeRebuildJob
	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.
			Where(
				"status = ? OR (status = ? AND (locked_at IS NULL OR lease_expires_at IS NULL OR lease_expires_at <= ?))",
				mediadomain.MediaDerivativeRebuildJobStatusPending,
				mediadomain.MediaDerivativeRebuildJobStatusRunning,
				now,
			).
			Order("CASE status WHEN 'running' THEN 1 ELSE 2 END").
			Order("id ASC")
		if isSkipLockedSupported(r.db) {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		} else {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}

		var candidate mediadomain.MediaDerivativeRebuildJob
		if err := query.First(&candidate).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else if err != nil {
			return err
		}

		leaseExpiresAt := now.Add(leaseTimeout)
		result := tx.Model(&mediadomain.MediaDerivativeRebuildJob{}).
			Where(
				"id = ? AND (status = ? OR (status = ? AND (locked_at IS NULL OR lease_expires_at IS NULL OR lease_expires_at <= ?)))",
				candidate.ID,
				mediadomain.MediaDerivativeRebuildJobStatusPending,
				mediadomain.MediaDerivativeRebuildJobStatusRunning,
				now,
			).
			Updates(map[string]interface{}{
				"status":           mediadomain.MediaDerivativeRebuildJobStatusRunning,
				"locked_at":        now,
				"locked_by":        workerID,
				"lease_generation": gorm.Expr("lease_generation + 1"),
				"lease_expires_at": leaseExpiresAt,
				"started_at":       gorm.Expr("COALESCE(started_at, ?)", now),
				"updated_at":       now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		candidate.Status = mediadomain.MediaDerivativeRebuildJobStatusRunning
		candidate.LockedAt = &now
		candidate.LockedBy = workerID
		candidate.LeaseGeneration++
		candidate.LeaseExpiresAt = &leaseExpiresAt
		if candidate.StartedAt == nil {
			candidate.StartedAt = &now
		}
		claimed = &candidate
		return nil
	})
	return claimed, err
}

func (r *MediaDerivativeRebuildJobRepository) CompleteBatch(
	job *mediadomain.MediaDerivativeRebuildJob,
	workerID string,
	result MediaDerivativeRebuildBatchUpdate,
	finished bool,
	now time.Time,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if job == nil || job.ID == 0 {
		return errors.New("media derivative rebuild job is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return errors.New("media derivative rebuild worker ID is required")
	}

	updates := map[string]interface{}{
		"cursor_asset_id":            result.CursorAssetID,
		"scanned_assets":             gorm.Expr("scanned_assets + ?", result.ScannedAssets),
		"generated_assets":           gorm.Expr("generated_assets + ?", result.GeneratedAssets),
		"generated_derivatives":      gorm.Expr("generated_derivatives + ?", result.GeneratedDerivatives),
		"failed_assets":              gorm.Expr("failed_assets + ?", result.FailedAssets),
		"updated_product_media_rows": gorm.Expr("updated_product_media_rows + ?", result.UpdatedProductMediaRows),
		"last_error": gorm.Expr(
			"CASE WHEN ? <> '' THEN ? ELSE last_error END",
			strings.TrimSpace(result.LastError),
			strings.TrimSpace(result.LastError),
		),
		"locked_at":        nil,
		"locked_by":        "",
		"lease_expires_at": nil,
		"updated_at":       now,
	}
	if finished {
		updates["status"] = mediadomain.MediaDerivativeRebuildJobStatusSucceeded
		updates["finished_at"] = now
	} else {
		updates["status"] = mediadomain.MediaDerivativeRebuildJobStatusRunning
	}

	dbResult := r.db.Model(&mediadomain.MediaDerivativeRebuildJob{}).
		Where(
			"id = ? AND status = ? AND locked_by = ? AND lease_generation = ? AND lease_expires_at > ?",
			job.ID,
			mediadomain.MediaDerivativeRebuildJobStatusRunning,
			workerID,
			job.LeaseGeneration,
			now,
		).
		Updates(updates)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrMediaDerivativeRebuildLeaseLost
	}
	return nil
}

type MediaDerivativeRebuildBatchUpdate struct {
	CursorAssetID           uint
	ScannedAssets           int
	GeneratedAssets         int
	GeneratedDerivatives    int
	FailedAssets            int
	UpdatedProductMediaRows int64
	LastError               string
}

func isMediaDerivativeRebuildPendingConflict(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique") || strings.Contains(message, "duplicate")
}

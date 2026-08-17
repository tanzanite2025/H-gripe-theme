package repository

import (
	"errors"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteQualityJobRepository struct {
	db *gorm.DB
}

var ErrSiteQualityLeaseLost = errors.New("SiteQuality quality job lease was lost")

func NewSiteQualityJobRepository(db *gorm.DB) *SiteQualityJobRepository {
	return &SiteQualityJobRepository{db: db}
}

func (r *SiteQualityJobRepository) WithTx(tx *gorm.DB) *SiteQualityJobRepository {
	return &SiteQualityJobRepository{db: tx}
}

func (r *SiteQualityJobRepository) Transaction(fn func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality job repository is unavailable")
	}
	return r.db.Transaction(fn)
}

func (r *SiteQualityJobRepository) Enqueue(
	job sitequalitydomain.SiteQualityJob,
) (*sitequalitydomain.SiteQualityJob, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, errors.New("SiteQuality quality job repository is unavailable")
	}
	if job.TargetID == 0 || job.Strategy == "" || job.Kind == "" || job.IdempotencyKey == "" {
		return nil, false, errors.New("SiteQuality quality job is incomplete")
	}
	switch job.Kind {
	case sitequalitydomain.SiteQualityJobKindScheduled,
		sitequalitydomain.SiteQualityJobKindManual,
		sitequalitydomain.SiteQualityJobKindRecheck:
	default:
		return nil, false, errors.New("SiteQuality quality job kind is invalid")
	}
	if job.Status == "" {
		job.Status = sitequalitydomain.SiteQualityJobStatusQueued
	}
	if job.AvailableAt.IsZero() {
		job.AvailableAt = time.Now().UTC()
	} else {
		job.AvailableAt = job.AvailableAt.UTC()
	}
	if job.SampleCount <= 0 || job.RequiredConfirmations <= 0 || job.RequiredConfirmations > job.SampleCount {
		return nil, false, errors.New("SiteQuality quality job sampling policy is invalid")
	}
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 4
	}

	result := job
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoNothing: true,
	}).Create(&result).Error
	if err != nil {
		return nil, false, err
	}
	if result.ID != 0 {
		return &result, true, nil
	}
	if err := r.db.Where("idempotency_key = ?", job.IdempotencyKey).First(&result).Error; err != nil {
		return nil, false, err
	}
	return &result, false, nil
}

func (r *SiteQualityJobRepository) FindByID(id uint) (*sitequalitydomain.SiteQualityJob, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality job repository is unavailable")
	}
	if id == 0 {
		return nil, errors.New("SiteQuality quality job ID is required")
	}
	var job sitequalitydomain.SiteQualityJob
	if err := r.db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *SiteQualityJobRepository) ClaimReady(
	now time.Time,
	workerID string,
	limit int,
	leaseTimeout time.Duration,
) ([]sitequalitydomain.SiteQualityJob, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality job repository is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if workerID == "" {
		return nil, errors.New("SiteQuality quality worker ID is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 10 * time.Minute
	}

	var claimed []sitequalitydomain.SiteQualityJob
	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&sitequalitydomain.SiteQualityJob{}).
			Where(
				"kind IN ? AND EXISTS (SELECT 1 FROM site_quality_targets WHERE site_quality_targets.id = site_quality_jobs.target_id AND site_quality_targets.enabled = ?) AND ((status IN ? AND available_at <= ? AND attempts < max_attempts) OR (status = ? AND ((lease_expires_at IS NOT NULL AND lease_expires_at <= ?) OR (lease_expires_at IS NULL AND locked_at IS NOT NULL AND locked_at <= ?)) AND attempts < max_attempts))",
				[]string{
					sitequalitydomain.SiteQualityJobKindScheduled,
					sitequalitydomain.SiteQualityJobKindManual,
					sitequalitydomain.SiteQualityJobKindRecheck,
				},
				true,
				[]string{sitequalitydomain.SiteQualityJobStatusQueued, sitequalitydomain.SiteQualityJobStatusFailed},
				now,
				sitequalitydomain.SiteQualityJobStatusProcessing,
				now,
				now.Add(-leaseTimeout),
			).
			Order("available_at ASC").
			Order("CASE kind WHEN 'recheck' THEN 1 WHEN 'manual' THEN 2 ELSE 3 END").
			Order("id ASC").
			Limit(limit)
		if isSkipLockedSupported(r.db) {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		} else {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.Find(&claimed).Error; err != nil {
			return err
		}
		if len(claimed) == 0 {
			claimed = []sitequalitydomain.SiteQualityJob{}
			return nil
		}

		for index := range claimed {
			leaseExpiresAt := now.Add(leaseTimeout)
			if err := tx.Model(&sitequalitydomain.SiteQualityJob{}).
				Where("id = ?", claimed[index].ID).
				Updates(map[string]interface{}{
					"status":           sitequalitydomain.SiteQualityJobStatusProcessing,
					"locked_at":        now,
					"locked_by":        workerID,
					"lease_generation": gorm.Expr("lease_generation + 1"),
					"lease_expires_at": leaseExpiresAt,
					"heartbeat_at":     now,
					"attempts":         gorm.Expr("attempts + 1"),
					"started_at":       gorm.Expr("COALESCE(started_at, ?)", now),
					"updated_at":       now,
				}).Error; err != nil {
				return err
			}
			claimed[index].Status = sitequalitydomain.SiteQualityJobStatusProcessing
			claimed[index].LockedAt = &now
			claimed[index].LockedBy = workerID
			claimed[index].LeaseGeneration++
			claimed[index].LeaseExpiresAt = &leaseExpiresAt
			claimed[index].HeartbeatAt = &now
			if claimed[index].StartedAt == nil {
				claimed[index].StartedAt = &now
			}
			claimed[index].Attempts++
		}
		return nil
	})
	return claimed, err
}

func (r *SiteQualityJobRepository) MarkSucceeded(
	id uint,
	workerID string,
	leaseGeneration int64,
	finishedAt time.Time,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality job repository is unavailable")
	}
	if id == 0 {
		return errors.New("SiteQuality quality job ID is required")
	}
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	} else {
		finishedAt = finishedAt.UTC()
	}
	result := r.db.Model(&sitequalitydomain.SiteQualityJob{}).
		Where("id = ? AND status = ? AND locked_by = ? AND lease_generation = ? AND (lease_expires_at IS NULL OR lease_expires_at > ?)",
			id,
			sitequalitydomain.SiteQualityJobStatusProcessing,
			workerID,
			leaseGeneration,
			finishedAt,
		).
		Updates(map[string]interface{}{
			"status":           sitequalitydomain.SiteQualityJobStatusSucceeded,
			"locked_at":        nil,
			"locked_by":        "",
			"lease_expires_at": nil,
			"heartbeat_at":     finishedAt,
			"finished_at":      finishedAt,
			"last_error":       "",
			"updated_at":       finishedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSiteQualityLeaseLost
	}
	return nil
}

func (r *SiteQualityJobRepository) MarkFailed(
	job sitequalitydomain.SiteQualityJob,
	workerID string,
	errorMessage string,
	nextAvailableAt time.Time,
	failedAt time.Time,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality job repository is unavailable")
	}
	if job.ID == 0 {
		return errors.New("SiteQuality quality job ID is required")
	}
	if failedAt.IsZero() {
		failedAt = time.Now().UTC()
	} else {
		failedAt = failedAt.UTC()
	}
	if nextAvailableAt.IsZero() {
		nextAvailableAt = failedAt
	} else {
		nextAvailableAt = nextAvailableAt.UTC()
	}
	status := sitequalitydomain.SiteQualityJobStatusFailed
	finishedAt := interface{}(nil)
	if job.Attempts >= job.MaxAttempts {
		status = sitequalitydomain.SiteQualityJobStatusDeadLetter
		finishedAt = failedAt
	}
	result := r.db.Model(&sitequalitydomain.SiteQualityJob{}).
		Where("id = ? AND status = ? AND locked_by = ? AND lease_generation = ? AND (lease_expires_at IS NULL OR lease_expires_at > ?)",
			job.ID,
			sitequalitydomain.SiteQualityJobStatusProcessing,
			workerID,
			job.LeaseGeneration,
			failedAt,
		).
		Updates(map[string]interface{}{
			"status":           status,
			"available_at":     nextAvailableAt,
			"locked_at":        nil,
			"locked_by":        "",
			"lease_expires_at": nil,
			"heartbeat_at":     failedAt,
			"finished_at":      finishedAt,
			"last_error":       strings.TrimSpace(errorMessage),
			"updated_at":       failedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSiteQualityLeaseLost
	}
	return nil
}

func (r *SiteQualityJobRepository) AssertLease(
	tx *gorm.DB,
	jobID uint,
	workerID string,
	leaseGeneration int64,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality job repository is unavailable")
	}
	if tx == nil || jobID == 0 || strings.TrimSpace(workerID) == "" || leaseGeneration <= 0 {
		return errors.New("SiteQuality quality job lease input is incomplete")
	}
	var job sitequalitydomain.SiteQualityJob
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"id = ? AND status = ? AND locked_by = ? AND lease_generation = ? AND (lease_expires_at IS NULL OR lease_expires_at > ?)",
			jobID,
			sitequalitydomain.SiteQualityJobStatusProcessing,
			workerID,
			leaseGeneration,
			time.Now().UTC(),
		).
		First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSiteQualityLeaseLost
		}
		return err
	}
	return nil
}

func (r *SiteQualityJobRepository) Heartbeat(
	jobID uint,
	workerID string,
	leaseGeneration int64,
	now time.Time,
	leaseTimeout time.Duration,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality job repository is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 10 * time.Minute
	}
	expiresAt := now.Add(leaseTimeout)
	result := r.db.Model(&sitequalitydomain.SiteQualityJob{}).
		Where(
			"id = ? AND status = ? AND locked_by = ? AND lease_generation = ? AND (lease_expires_at IS NULL OR lease_expires_at > ?)",
			jobID,
			sitequalitydomain.SiteQualityJobStatusProcessing,
			workerID,
			leaseGeneration,
			now,
		).
		Updates(map[string]interface{}{
			"heartbeat_at":     now,
			"lease_expires_at": expiresAt,
			"updated_at":       now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSiteQualityLeaseLost
	}
	return nil
}

func (r *SiteQualityJobRepository) CreateEvaluationForLease(
	tx *gorm.DB,
	evaluation *sitequalitydomain.SiteQualityEvaluation,
	workerID string,
	leaseGeneration int64,
) error {
	if evaluation == nil || evaluation.JobID == 0 || evaluation.TargetID == 0 {
		return errors.New("SiteQuality quality evaluation is incomplete")
	}
	if err := r.AssertLease(tx, evaluation.JobID, workerID, leaseGeneration); err != nil {
		return err
	}
	if err := tx.Create(evaluation).Error; err != nil {
		return err
	}
	return nil
}

func (r *SiteQualityJobRepository) AcquireProviderSlot(
	provider string,
	maxSlots int,
	workerID string,
	now time.Time,
	leaseTimeout time.Duration,
) (*sitequalitydomain.SiteQualityProviderSlot, *time.Time, error) {
	if r == nil || r.db == nil {
		return nil, nil, errors.New("SiteQuality quality job repository is unavailable")
	}
	provider = strings.TrimSpace(provider)
	if provider == "" || workerID == "" {
		return nil, nil, errors.New("SiteQuality provider slot input is incomplete")
	}
	if maxSlots <= 0 {
		maxSlots = 1
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 2 * time.Minute
	}

	var claimed *sitequalitydomain.SiteQualityProviderSlot
	var retryAt *time.Time
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for slot := 1; slot <= maxSlots; slot++ {
			row := sitequalitydomain.SiteQualityProviderSlot{
				Provider:    provider,
				Slot:        slot,
				AvailableAt: now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "provider"}, {Name: "slot"}},
				DoNothing: true,
			}).Create(&row).Error; err != nil {
				return err
			}
		}

		query := tx.Model(&sitequalitydomain.SiteQualityProviderSlot{}).
			Where(
				"provider = ? AND available_at <= ? AND (locked_at IS NULL OR locked_at <= ?)",
				provider,
				now,
				now.Add(-leaseTimeout),
			).
			Order("available_at ASC, slot ASC").
			Limit(1)
		if isSkipLockedSupported(r.db) {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		} else {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}

		var slot sitequalitydomain.SiteQualityProviderSlot
		err := query.First(&slot).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var next sitequalitydomain.SiteQualityProviderSlot
			if lookupErr := tx.
				Where("provider = ?", provider).
				Order("available_at ASC, slot ASC").
				First(&next).Error; lookupErr != nil {
				return lookupErr
			}
			nextAt := next.AvailableAt
			if next.LockedAt != nil {
				leaseExpiry := next.LockedAt.Add(leaseTimeout)
				if leaseExpiry.After(nextAt) {
					nextAt = leaseExpiry
				}
			}
			retryAt = &nextAt
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&sitequalitydomain.SiteQualityProviderSlot{}).
			Where("provider = ? AND slot = ?", provider, slot.Slot).
			Updates(map[string]interface{}{
				"locked_at":  now,
				"locked_by":  workerID,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		slot.LockedAt = &now
		slot.LockedBy = workerID
		claimed = &slot
		return nil
	})
	return claimed, retryAt, err
}

func (r *SiteQualityJobRepository) ReleaseProviderSlot(
	provider string,
	slot int,
	workerID string,
	now time.Time,
	minimumSpacing time.Duration,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality job repository is unavailable")
	}
	if strings.TrimSpace(provider) == "" || slot <= 0 || workerID == "" {
		return errors.New("SiteQuality provider slot input is incomplete")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if minimumSpacing < 0 {
		minimumSpacing = 0
	}
	result := r.db.Model(&sitequalitydomain.SiteQualityProviderSlot{}).
		Where("provider = ? AND slot = ? AND locked_by = ?", provider, slot, workerID).
		Updates(map[string]interface{}{
			"available_at": now.Add(minimumSpacing),
			"locked_at":    nil,
			"locked_by":    "",
			"updated_at":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("SiteQuality provider slot lease was lost")
	}
	return nil
}

func isSkipLockedSupported(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	switch db.Dialector.Name() {
	case "postgres", "mysql":
		return true
	default:
		return false
	}
}

func (r *SiteQualityJobRepository) Stats(now time.Time, leaseTimeout time.Duration) (SiteQualityJobStats, error) {
	if r == nil || r.db == nil {
		return SiteQualityJobStats{}, errors.New("SiteQuality quality job repository is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 10 * time.Minute
	}

	var stats SiteQualityJobStats
	base := func() *gorm.DB {
		return r.db.Model(&sitequalitydomain.SiteQualityJob{}).Where(
			"kind IN ?",
			[]string{
				sitequalitydomain.SiteQualityJobKindScheduled,
				sitequalitydomain.SiteQualityJobKindManual,
				sitequalitydomain.SiteQualityJobKindRecheck,
			},
		)
	}
	if err := base().Count(&stats.Total).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusQueued).Count(&stats.Queued).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusProcessing).Count(&stats.Processing).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusSucceeded).Count(&stats.Succeeded).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusFailed).Count(&stats.Failed).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusDeadLetter).Count(&stats.DeadLetter).Error; err != nil {
		return stats, err
	}

	claimableQuery := base().Where(
		"EXISTS (SELECT 1 FROM site_quality_targets WHERE site_quality_targets.id = site_quality_jobs.target_id AND site_quality_targets.enabled = ?) AND ((status IN ? AND available_at <= ? AND attempts < max_attempts) OR (status = ? AND ((lease_expires_at IS NOT NULL AND lease_expires_at <= ?) OR (lease_expires_at IS NULL AND locked_at IS NOT NULL AND locked_at <= ?)) AND attempts < max_attempts))",
		true,
		[]string{sitequalitydomain.SiteQualityJobStatusQueued, sitequalitydomain.SiteQualityJobStatusFailed},
		now,
		sitequalitydomain.SiteQualityJobStatusProcessing,
		now,
		now.Add(-leaseTimeout),
	)
	if err := claimableQuery.Count(&stats.Claimable).Error; err != nil {
		return stats, err
	}
	staleQuery := base().Where(
		"status = ? AND ((lease_expires_at IS NOT NULL AND lease_expires_at <= ?) OR (lease_expires_at IS NULL AND locked_at IS NOT NULL AND locked_at <= ?)) AND attempts < max_attempts",
		sitequalitydomain.SiteQualityJobStatusProcessing,
		now,
		now.Add(-leaseTimeout),
	)
	if err := staleQuery.Count(&stats.StaleLeases).Error; err != nil {
		return stats, err
	}

	if timestamp, err := firstSiteQualityJobTimestamp(base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusQueued), "available_at ASC, id ASC", func(job sitequalitydomain.SiteQualityJob) time.Time {
		return job.AvailableAt
	}); err != nil {
		return stats, err
	} else {
		stats.OldestQueuedAt = timestamp
	}
	if timestamp, err := firstSiteQualityJobTimestamp(base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusProcessing), "locked_at ASC, id ASC", func(job sitequalitydomain.SiteQualityJob) time.Time {
		if job.LockedAt != nil {
			return *job.LockedAt
		}
		return time.Time{}
	}); err != nil {
		return stats, err
	} else {
		stats.OldestProcessingAt = timestamp
	}
	if timestamp, err := firstSiteQualityJobTimestamp(base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusSucceeded), "finished_at DESC, id DESC", func(job sitequalitydomain.SiteQualityJob) time.Time {
		if job.FinishedAt != nil {
			return *job.FinishedAt
		}
		return time.Time{}
	}); err != nil {
		return stats, err
	} else {
		stats.LatestSuccessAt = timestamp
	}
	if timestamp, err := firstSiteQualityJobTimestamp(base().Where("status IN ?", []string{
		sitequalitydomain.SiteQualityJobStatusFailed,
		sitequalitydomain.SiteQualityJobStatusDeadLetter,
	}), "updated_at DESC, id DESC", func(job sitequalitydomain.SiteQualityJob) time.Time {
		return job.UpdatedAt
	}); err != nil {
		return stats, err
	} else {
		stats.LatestFailureAt = timestamp
	}
	if timestamp, err := firstSiteQualityJobTimestamp(base().Where("status = ?", sitequalitydomain.SiteQualityJobStatusDeadLetter), "finished_at DESC, id DESC", func(job sitequalitydomain.SiteQualityJob) time.Time {
		if job.FinishedAt != nil {
			return *job.FinishedAt
		}
		return time.Time{}
	}); err != nil {
		return stats, err
	} else {
		stats.LatestDeadLetterAt = timestamp
	}

	return stats, nil
}

func (r *SiteQualityJobRepository) ProviderSlotStats(
	provider string,
	configured int,
	now time.Time,
	leaseTimeout time.Duration,
) (SiteQualityProviderSlotStats, error) {
	if r == nil || r.db == nil {
		return SiteQualityProviderSlotStats{}, errors.New("SiteQuality quality job repository is unavailable")
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return SiteQualityProviderSlotStats{}, errors.New("SiteQuality provider name is required")
	}
	if configured <= 0 {
		configured = 1
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 2 * time.Minute
	}

	var slots []sitequalitydomain.SiteQualityProviderSlot
	if err := r.db.Where("provider = ?", provider).Order("slot ASC").Find(&slots).Error; err != nil {
		return SiteQualityProviderSlotStats{}, err
	}

	stats := SiteQualityProviderSlotStats{
		Provider:   provider,
		Configured: configured,
		Total:      int64(len(slots)),
	}
	staleBefore := now.Add(-leaseTimeout)
	var nextAvailableAt time.Time
	hasNextAvailableAt := false
	availableNow := int64(0)
	for _, slot := range slots {
		slotAvailableAt := slot.AvailableAt.UTC()
		if slot.LockedAt != nil {
			stats.Locked++
			leaseExpiry := slot.LockedAt.UTC().Add(leaseTimeout)
			if leaseExpiry.After(slotAvailableAt) {
				slotAvailableAt = leaseExpiry
			}
			if slot.LockedAt.UTC().Before(staleBefore) || slot.LockedAt.UTC().Equal(staleBefore) {
				stats.StaleLocked++
				availableNow++
				slotAvailableAt = now
			}
		}
		if slot.LockedAt == nil && !slot.AvailableAt.After(now) {
			availableNow++
		}
		if !hasNextAvailableAt || slotAvailableAt.Before(nextAvailableAt) {
			nextAvailableAt = slotAvailableAt
			hasNextAvailableAt = true
		}
	}

	if configured > len(slots) {
		availableNow += int64(configured - len(slots))
		if !hasNextAvailableAt || now.Before(nextAvailableAt) {
			nextAvailableAt = now
			hasNextAvailableAt = true
		}
	}
	stats.Available = availableNow
	if availableNow > 0 {
		nextAvailableAt = now
		hasNextAvailableAt = true
	}
	if hasNextAvailableAt {
		candidate := nextAvailableAt.UTC()
		stats.NextAvailableAt = &candidate
	}
	return stats, nil
}

func firstSiteQualityJobTimestamp(
	query *gorm.DB,
	order string,
	extract func(sitequalitydomain.SiteQualityJob) time.Time,
) (*time.Time, error) {
	if query == nil {
		return nil, errors.New("SiteQuality quality job query is unavailable")
	}
	var job sitequalitydomain.SiteQualityJob
	if err := query.Order(order).Limit(1).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if extract == nil {
		return nil, nil
	}
	value := extract(job).UTC()
	if value.IsZero() {
		return nil, nil
	}
	copy := value
	return &copy, nil
}

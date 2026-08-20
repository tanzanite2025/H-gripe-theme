package repository

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrOutboxOwnershipLost   = errors.New("outbox event ownership lost")
	ErrOutboxUnknownNotFound = errors.New("unknown outbox event not found")
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) WithTx(tx *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: tx}
}

func (r *OutboxRepository) lockForClaim(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql":
		return query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
	case "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

func (r *OutboxRepository) CreateEvent(event *outbox.Event) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_key"}},
		DoNothing: true,
	}).Create(event).Error
}

func (r *OutboxRepository) FindEventByKey(eventKey string) (*outbox.Event, error) {
	var event outbox.Event
	if err := r.db.Where("event_key = ?", eventKey).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// CountEventsByStatus returns operational counts for one event type. It is
// intentionally aggregate-only so monitoring never needs to read Outbox
// payloads, which can contain business data for unrelated consumers.
func (r *OutboxRepository) CountEventsByStatus(eventType string) (map[string]int64, error) {
	counts := make(map[string]int64)
	if r == nil || r.db == nil || strings.TrimSpace(eventType) == "" {
		return counts, nil
	}

	type statusCount struct {
		Status string
		Count  int64
	}
	var rows []statusCount
	if err := r.db.Model(&outbox.Event{}).
		Select("status, COUNT(*) AS count").
		Where("event_type = ?", eventType).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.Status] = row.Count
	}
	return counts, nil
}

func (r *OutboxRepository) ClaimReadyEvents(now time.Time, workerID string, limit int, lockTimeout time.Duration) ([]outbox.Event, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if lockTimeout <= 0 {
		lockTimeout = 5 * time.Minute
	}
	if workerID == "" {
		workerID = "outbox-worker"
	}

	staleLockCutoff := now.Add(-lockTimeout)
	claimableStatuses := []string{outbox.EventStatusPending, outbox.EventStatusFailed}

	var claimed []outbox.Event
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var events []outbox.Event
		query := tx.Model(&outbox.Event{}).
			Where(
				"(status IN ? AND available_at <= ? AND attempts < max_attempts) OR (status = ? AND locked_at IS NOT NULL AND locked_at <= ? AND attempts < max_attempts)",
				claimableStatuses,
				now,
				outbox.EventStatusProcessing,
				staleLockCutoff,
			).
			Order("available_at ASC, id ASC").
			Limit(limit)
		if err := r.lockForClaim(query).Find(&events).Error; err != nil {
			return err
		}
		if len(events) == 0 {
			claimed = []outbox.Event{}
			return nil
		}

		ids := make([]uint, 0, len(events))
		for index := range events {
			ids = append(ids, events[index].ID)
		}
		if err := tx.Model(&outbox.Event{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"status":          outbox.EventStatusProcessing,
				"locked_at":       now,
				"locked_by":       workerID,
				"last_attempt_at": now,
				"attempts":        gorm.Expr("attempts + 1"),
				"updated_at":      now,
			}).Error; err != nil {
			return err
		}

		for index := range events {
			events[index].Status = outbox.EventStatusProcessing
			events[index].LockedAt = &now
			events[index].LockedBy = workerID
			events[index].Attempts++
		}
		claimed = events
		return nil
	})
	return claimed, err
}

func (r *OutboxRepository) MarkProcessed(id uint, processedAt time.Time) error {
	return r.MarkProcessedByWorker(id, "", processedAt)
}

func (r *OutboxRepository) MarkProcessedByWorker(id uint, workerID string, processedAt time.Time) error {
	if processedAt.IsZero() {
		processedAt = time.Now().UTC()
	} else {
		processedAt = processedAt.UTC()
	}
	query := r.db.Model(&outbox.Event{}).Where("id = ?", id)
	if strings.TrimSpace(workerID) != "" {
		query = query.Where("status = ? AND locked_by = ?", outbox.EventStatusProcessing, workerID)
	}
	result := query.
		Updates(map[string]interface{}{
			"status":          outbox.EventStatusProcessed,
			"processed_at":    processedAt,
			"locked_at":       gorm.Expr("NULL"),
			"locked_by":       "",
			"last_error":      "",
			"uncertain_at":    gorm.Expr("NULL"),
			"reconcile_after": gorm.Expr("NULL"),
			"updated_at":      processedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if strings.TrimSpace(workerID) != "" && result.RowsAffected == 0 {
		return ErrOutboxOwnershipLost
	}
	return nil
}

func (r *OutboxRepository) MarkFailed(id uint, status, errorMessage string, nextAvailableAt, failedAt time.Time) error {
	return r.MarkFailedByWorker(id, "", status, errorMessage, nextAvailableAt, failedAt)
}

func (r *OutboxRepository) MarkFailedByWorker(
	id uint,
	workerID string,
	status,
	errorMessage string,
	nextAvailableAt,
	failedAt time.Time,
) error {
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
	if status == "" {
		status = outbox.EventStatusFailed
	}

	query := r.db.Model(&outbox.Event{}).Where("id = ?", id)
	if strings.TrimSpace(workerID) != "" {
		query = query.Where("status = ? AND locked_by = ?", outbox.EventStatusProcessing, workerID)
	}
	result := query.
		Updates(map[string]interface{}{
			"status":          status,
			"available_at":    nextAvailableAt,
			"locked_at":       gorm.Expr("NULL"),
			"locked_by":       "",
			"last_error":      errorMessage,
			"uncertain_at":    gorm.Expr("NULL"),
			"reconcile_after": gorm.Expr("NULL"),
			"updated_at":      failedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if strings.TrimSpace(workerID) != "" && result.RowsAffected == 0 {
		return ErrOutboxOwnershipLost
	}
	return nil
}

func (r *OutboxRepository) RefreshProcessingLockByWorker(id uint, workerID string, lockedAt time.Time) error {
	if lockedAt.IsZero() {
		lockedAt = time.Now().UTC()
	} else {
		lockedAt = lockedAt.UTC()
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return ErrOutboxOwnershipLost
	}
	result := r.db.Model(&outbox.Event{}).
		Where("id = ? AND status = ? AND locked_by = ?", id, outbox.EventStatusProcessing, workerID).
		Updates(map[string]interface{}{
			"locked_at":  lockedAt,
			"updated_at": lockedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOutboxOwnershipLost
	}
	return nil
}

func (r *OutboxRepository) MarkUnknownByWorker(
	id uint,
	workerID string,
	errorMessage string,
	reconcileAfter,
	uncertainAt time.Time,
) error {
	if uncertainAt.IsZero() {
		uncertainAt = time.Now().UTC()
	} else {
		uncertainAt = uncertainAt.UTC()
	}
	if reconcileAfter.IsZero() {
		reconcileAfter = uncertainAt
	} else {
		reconcileAfter = reconcileAfter.UTC()
	}

	result := r.db.Model(&outbox.Event{}).
		Where("id = ? AND status = ? AND locked_by = ?", id, outbox.EventStatusProcessing, workerID).
		Updates(map[string]interface{}{
			"status":          outbox.EventStatusUnknown,
			"available_at":    reconcileAfter,
			"locked_at":       nil,
			"locked_by":       "",
			"uncertain_at":    uncertainAt,
			"reconcile_after": reconcileAfter,
			"last_error":      errorMessage,
			"updated_at":      uncertainAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOutboxOwnershipLost
	}
	return nil
}

func (r *OutboxRepository) FindUnknownEvents(limit int) ([]outbox.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var events []outbox.Event
	err := r.db.Where("status = ?", outbox.EventStatusUnknown).
		Order("reconcile_after ASC, id ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *OutboxRepository) ResumeUnknownEvent(id uint, nextAvailableAt time.Time, note string, resumedAt time.Time) error {
	if resumedAt.IsZero() {
		resumedAt = time.Now().UTC()
	} else {
		resumedAt = resumedAt.UTC()
	}
	if nextAvailableAt.IsZero() {
		nextAvailableAt = resumedAt
	} else {
		nextAvailableAt = nextAvailableAt.UTC()
	}
	result := r.db.Model(&outbox.Event{}).
		Where("id = ? AND status = ?", id, outbox.EventStatusUnknown).
		Updates(map[string]interface{}{
			"status":          outbox.EventStatusFailed,
			"available_at":    nextAvailableAt,
			"uncertain_at":    gorm.Expr("NULL"),
			"reconcile_after": gorm.Expr("NULL"),
			"last_error":      note,
			"updated_at":      resumedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOutboxUnknownNotFound
	}
	return nil
}

func (r *OutboxRepository) MarkUnknownProcessed(id uint, note string, processedAt time.Time) error {
	if processedAt.IsZero() {
		processedAt = time.Now().UTC()
	} else {
		processedAt = processedAt.UTC()
	}
	result := r.db.Model(&outbox.Event{}).
		Where("id = ? AND status = ?", id, outbox.EventStatusUnknown).
		Updates(map[string]interface{}{
			"status":          outbox.EventStatusProcessed,
			"processed_at":    processedAt,
			"uncertain_at":    gorm.Expr("NULL"),
			"reconcile_after": gorm.Expr("NULL"),
			"last_error":      note,
			"updated_at":      processedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOutboxUnknownNotFound
	}
	return nil
}

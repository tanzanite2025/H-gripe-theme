package repository

import (
	"time"

	"commerce-platform/internal/domain/outbox"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
				"status":     outbox.EventStatusProcessing,
				"locked_at":  now,
				"locked_by":  workerID,
				"attempts":   gorm.Expr("attempts + 1"),
				"updated_at": now,
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
	if processedAt.IsZero() {
		processedAt = time.Now().UTC()
	} else {
		processedAt = processedAt.UTC()
	}
	return r.db.Model(&outbox.Event{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       outbox.EventStatusProcessed,
			"processed_at": processedAt,
			"locked_at":    nil,
			"locked_by":    "",
			"last_error":   "",
			"updated_at":   processedAt,
		}).Error
}

func (r *OutboxRepository) MarkFailed(id uint, status, errorMessage string, nextAvailableAt, failedAt time.Time) error {
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

	return r.db.Model(&outbox.Event{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"available_at": nextAvailableAt,
			"locked_at":    nil,
			"locked_by":    "",
			"last_error":   errorMessage,
			"updated_at":   failedAt,
		}).Error
}

package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/outbox"
)

type DashboardOutboxBusinessReadScope struct {
	CanViewAllEventTypes          bool
	AllowedEventTypes             []string
	AssignedAgentUserID           *uint
	AllowedBusinessAggregateTypes []string
}

func applyDashboardOutboxBusinessReadScope(query *gorm.DB, scope DashboardOutboxBusinessReadScope) *gorm.DB {
	if !scope.CanViewAllEventTypes {
		if len(scope.AllowedEventTypes) == 0 {
			return query.Where("1 = 0")
		}
		query = query.Where("event_type IN ?", scope.AllowedEventTypes)
	}
	if !scope.CanViewAllEventTypes || scope.AssignedAgentUserID != nil {
		if scope.AssignedAgentUserID == nil || *scope.AssignedAgentUserID == 0 {
			// Domains without conversation access must not depend on the tickets
			// table. Exclude both identifiers to reject mismatched event metadata.
			return query.Where("event_type <> ? AND aggregate_type <> ?", outbox.EventTypeCustomerServiceRealtime, outbox.AggregateTypeCustomerServiceConversation)
		}
		assignedAgentUserID := *scope.AssignedAgentUserID
		query = query.Where("((event_type <> ? AND aggregate_type <> ?) OR (aggregate_type = ? AND ? > 0 AND aggregate_id IN (SELECT CAST(id AS TEXT) FROM tickets WHERE assigned_to = ? AND deleted_at IS NULL)))", outbox.EventTypeCustomerServiceRealtime, outbox.AggregateTypeCustomerServiceConversation, outbox.AggregateTypeCustomerServiceConversation, assignedAgentUserID, assignedAgentUserID)
	}
	return query
}

// DashboardOutboxFailureSummaryItem is internal delivery metadata. Raw errors and
// aggregate identifiers are excluded from JSON and translated into safe explanations.
type DashboardOutboxFailureSummaryItem struct {
	ID            uint       `json:"id"`
	EventType     string     `json:"event_type"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	MaxAttempts   int        `json:"max_attempts"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedAt     time.Time  `json:"created_at"`
	AvailableAt   time.Time  `json:"-"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
	AggregateType string     `json:"-"`
	AggregateID   string     `json:"-"`
	LastError     string     `json:"-"`
}

type DashboardOutboxFailureSnapshot struct {
	TotalCount           int64
	FailedCount          int64
	DeadLetterCount      int64
	ManualAttentionCount int64
	Items                []DashboardOutboxFailureSummaryItem
}

type DashboardOutboxBusinessFollowUpReadRepository struct {
	db *gorm.DB
}

func NewDashboardOutboxBusinessFollowUpReadRepository(db *gorm.DB) *DashboardOutboxBusinessFollowUpReadRepository {
	return &DashboardOutboxBusinessFollowUpReadRepository{db: db}
}

// GetDashboardOutboxFailureSnapshot counts the entire actionable queue and limits
// only its preview. Window aggregates keep counts and preview in one SQL snapshot.
// Dead letters come first, followed by the oldest unresolved failures.
func (r *DashboardOutboxBusinessFollowUpReadRepository) GetDashboardOutboxFailureSnapshot(ctx context.Context, limit int, scope DashboardOutboxBusinessReadScope) (*DashboardOutboxFailureSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("dashboard outbox repository is unavailable")
	}
	if limit <= 0 || limit > 100 {
		limit = 5
	}

	var rows []struct {
		DashboardOutboxFailureSummaryItem `gorm:"embedded"`
		TotalCount                        int64
		FailedCount                       int64
		DeadLetterCount                   int64
		ManualAttentionCount              int64
	}
	query := r.db.WithContext(ctx).Model(&outbox.Event{}).
		Select(`id, event_type, status, attempts, max_attempts, updated_at,
			created_at, available_at, last_attempt_at, aggregate_type, aggregate_id, last_error,
			COUNT(*) OVER () AS total_count,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) OVER () AS failed_count,
			SUM(CASE WHEN status = 'dead_letter' THEN 1 ELSE 0 END) OVER () AS dead_letter_count,
			SUM(CASE WHEN status = 'dead_letter' OR attempts >= max_attempts THEN 1 ELSE 0 END) OVER () AS manual_attention_count`).
		Where("status IN ?", []string{outbox.EventStatusFailed, outbox.EventStatusDeadLetter})
	query = applyDashboardOutboxBusinessReadScope(query, scope)
	err := query.
		Order("CASE WHEN status = 'dead_letter' OR attempts >= max_attempts THEN 0 ELSE 1 END, updated_at ASC, id ASC").
		Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	snapshot := &DashboardOutboxFailureSnapshot{Items: make([]DashboardOutboxFailureSummaryItem, 0, len(rows))}
	if len(rows) > 0 {
		snapshot.TotalCount = rows[0].TotalCount
		snapshot.FailedCount = rows[0].FailedCount
		snapshot.DeadLetterCount = rows[0].DeadLetterCount
		snapshot.ManualAttentionCount = rows[0].ManualAttentionCount
	}
	for _, row := range rows {
		snapshot.Items = append(snapshot.Items, row.DashboardOutboxFailureSummaryItem)
	}
	return snapshot, nil
}

func (r *DashboardOutboxBusinessFollowUpReadRepository) FindDashboardBusinessFollowUpFailureByID(ctx context.Context, eventID uint, scope DashboardOutboxBusinessReadScope) (*DashboardOutboxFailureSummaryItem, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("dashboard outbox repository is unavailable")
	}
	var item DashboardOutboxFailureSummaryItem
	query := r.db.WithContext(ctx).Model(&outbox.Event{}).
		Select("id, event_type, status, attempts, max_attempts, created_at, updated_at, available_at, last_attempt_at, aggregate_type, aggregate_id, last_error").
		Where("id = ? AND status IN ?", eventID, []string{outbox.EventStatusFailed, outbox.EventStatusDeadLetter})
	query = applyDashboardOutboxBusinessReadScope(query, scope)
	result := query.Limit(1).Scan(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrOutboxEventNotFound
	}
	return &item, nil
}

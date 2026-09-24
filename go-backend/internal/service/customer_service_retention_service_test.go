package service

import (
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCustomerServiceRetentionEligibilityRequiresClosedMinimumAge(t *testing.T) {
	now := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)
	retention := NewCustomerServiceRetentionService(nil, nil, nil, nil, nil, nil)
	retention.now = func() time.Time { return now }
	retention.ConfigureWindows(24*time.Hour, 30*24*time.Hour)

	open := retention.evaluateRecord(ticket.Ticket{ID: 1, Status: "open", UpdatedAt: now.Add(-48 * time.Hour)}, now)
	require.False(t, open.Eligible)
	require.Equal(t, "conversation is not resolved or closed", open.Reason)

	closed := retention.evaluateRecord(ticket.Ticket{ID: 2, Status: "closed", UpdatedAt: now.Add(-48 * time.Hour)}, now)
	require.True(t, closed.Eligible)
	require.Equal(t, "soft_delete", closed.Action)
}

func TestCustomerServiceRetentionPurgeHonorsRecoveryWindow(t *testing.T) {
	now := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)
	retention := NewCustomerServiceRetentionService(nil, nil, nil, nil, nil, nil)
	retention.ConfigureWindows(24*time.Hour, 30*24*time.Hour)
	deletedAt := now.Add(-24 * time.Hour)
	item := retention.evaluateRecord(ticket.Ticket{ID: 3, Status: "closed", UpdatedAt: now.Add(-90 * 24 * time.Hour), DeletedAt: gorm.DeletedAt{Time: deletedAt, Valid: true}}, now)
	require.False(t, item.Eligible)
	require.Contains(t, item.Reason, "recovery window")
}

func TestCustomerServiceRetentionSoftDeleteAndPurgeRemovesTicketOwnedRows(t *testing.T) {
	db := newCustomerServiceRetentionTestDB(t)
	ticketRepo := repository.NewTicketRepository(db)
	auditService := NewAuditService(repository.NewAuditRepository(db))
	retention := NewCustomerServiceRetentionService(ticketRepo, nil, nil, nil, nil, auditService)
	now := time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)
	retention.now = func() time.Time { return now }
	retention.ConfigureWindows(24*time.Hour, 30*24*time.Hour)

	conversation := ticket.Ticket{
		TicketNumber: "RETENTION-1",
		UserID:       42,
		Subject:      "retention test",
		Category:     "customer_service",
		Status:       "closed",
		UpdatedAt:    now.Add(-90 * 24 * time.Hour),
		ClosedAt:     retentionTestTimePtr(now.Add(-90 * 24 * time.Hour)),
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID:    conversation.ID,
		Content:     "message body",
		Attachments: `["uploads/customer-service/retention-1.txt"]`,
	}).Error)
	require.NoError(t, db.Create(&ticket.CustomerServiceInboxState{
		TicketID:          conversation.ID,
		RecipientUserID:   7,
		AssignmentVersion: 1,
	}).Error)

	softDeleted, err := retention.SoftDelete([]uint{conversation.ID}, 99, "approved retention policy")
	require.NoError(t, err)
	require.Len(t, softDeleted, 1)
	require.Equal(t, "soft_deleted", softDeleted[0].Action)
	require.NotNil(t, softDeleted[0].SoftDeletedAt)

	var tombstone ticket.Ticket
	require.NoError(t, db.Unscoped().First(&tombstone, conversation.ID).Error)
	require.True(t, tombstone.DeletedAt.Valid)
	var auditCount int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Where("resource_id = ?", conversation.ID).Count(&auditCount).Error)
	require.EqualValues(t, 1, auditCount)

	retention.now = func() time.Time { return now.Add(31 * 24 * time.Hour) }
	purged, err := retention.Purge([]uint{conversation.ID}, 99, "approved retention policy")
	require.NoError(t, err)
	require.Len(t, purged, 1)
	require.Equal(t, "purge", purged[0].Action)

	var remaining int64
	require.NoError(t, db.Unscoped().Model(&ticket.Ticket{}).Where("id = ?", conversation.ID).Count(&remaining).Error)
	require.Zero(t, remaining)
	require.NoError(t, db.Unscoped().Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&remaining).Error)
	require.Zero(t, remaining)
	require.NoError(t, db.Unscoped().Model(&ticket.CustomerServiceInboxState{}).Where("ticket_id = ?", conversation.ID).Count(&remaining).Error)
	require.Zero(t, remaining)
	require.NoError(t, db.Model(&audit.AuditLog{}).Where("resource_id = ?", conversation.ID).Count(&auditCount).Error)
	require.EqualValues(t, 2, auditCount)
}

func TestCustomerServiceRetentionRejectsNonCustomerServiceConversation(t *testing.T) {
	db := newCustomerServiceRetentionTestDB(t)
	ticketRepo := repository.NewTicketRepository(db)
	retention := NewCustomerServiceRetentionService(ticketRepo, nil, nil, nil, nil, nil)

	conversation := ticket.Ticket{
		TicketNumber: "RETENTION-OTHER",
		UserID:       42,
		Subject:      "other ticket",
		Category:     "order",
		Status:       "closed",
		UpdatedAt:    time.Now().UTC().Add(-90 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(&conversation).Error)

	eligibility, err := retention.Evaluate([]uint{conversation.ID})
	require.NoError(t, err)
	require.Len(t, eligibility, 1)
	require.False(t, eligibility[0].Eligible)
	require.Contains(t, eligibility[0].Reason, "not found")
}

func TestCustomerServiceRetentionBatchPrevalidatesBeforeSoftDelete(t *testing.T) {
	db := newCustomerServiceRetentionTestDB(t)
	ticketRepo := repository.NewTicketRepository(db)
	retention := NewCustomerServiceRetentionService(ticketRepo, nil, nil, nil, nil, NewAuditService(repository.NewAuditRepository(db)))
	now := time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)
	retention.now = func() time.Time { return now }
	retention.ConfigureWindows(24*time.Hour, 30*24*time.Hour)

	eligible := ticket.Ticket{
		TicketNumber: "RETENTION-BATCH-ELIGIBLE",
		UserID:       42,
		Category:     "customer_service",
		Status:       "closed",
		UpdatedAt:    now.Add(-90 * 24 * time.Hour),
	}
	ineligible := ticket.Ticket{
		TicketNumber: "RETENTION-BATCH-INELIGIBLE",
		UserID:       42,
		Category:     "customer_service",
		Status:       "open",
		UpdatedAt:    now.Add(-90 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(&eligible).Error)
	require.NoError(t, db.Create(&ineligible).Error)

	_, err := retention.SoftDelete([]uint{eligible.ID, ineligible.ID}, 99, "batch prevalidation")
	require.ErrorIs(t, err, ErrCustomerServiceRetentionIneligible)

	var persisted ticket.Ticket
	require.NoError(t, db.Unscoped().First(&persisted, eligible.ID).Error)
	require.False(t, persisted.DeletedAt.Valid)
}

func TestCustomerServiceRetentionPurgeQueuesDurableCleanupEvent(t *testing.T) {
	db := newCustomerServiceRetentionTestDB(t)
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	ticketRepo := repository.NewTicketRepository(db)
	retention := NewCustomerServiceRetentionService(
		ticketRepo,
		nil,
		nil,
		nil,
		nil,
		NewAuditService(repository.NewAuditRepository(db)),
	)
	retention.ConfigureCleanupOutbox(repository.NewOutboxRepository(db))
	now := time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)
	retention.now = func() time.Time { return now }
	retention.ConfigureWindows(24*time.Hour, 30*24*time.Hour)

	conversation := ticket.Ticket{
		TicketNumber: "RETENTION-OUTBOX",
		UserID:       42,
		Subject:      "outbox test",
		Category:     "customer_service",
		Status:       "closed",
		UpdatedAt:    now.Add(-90 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID:    conversation.ID,
		Content:     "message body",
		Attachments: `["uploads/customer-service/outbox.txt"]`,
	}).Error)

	_, err := retention.SoftDelete([]uint{conversation.ID}, 99, "approved retention policy")
	require.NoError(t, err)
	retention.now = func() time.Time { return now.Add(31 * 24 * time.Hour) }
	_, err = retention.Purge([]uint{conversation.ID}, 99, "approved retention policy")
	require.NoError(t, err)

	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceRetentionCleanup).First(&event).Error)
	require.Equal(t, outbox.EventStatusPending, event.Status)
	var payload outbox.CustomerServiceRetentionCleanupPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, conversation.ID, payload.TicketID)
	require.Equal(t, []string{"uploads/customer-service/outbox.txt"}, payload.AttachmentReferences)
	require.NoError(t, retention.HandleCustomerServiceRetentionCleanup(nil, event))
}

func newCustomerServiceRetentionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&ticket.Ticket{},
		&ticket.TicketMessage{},
		&ticket.CustomerServiceInboxState{},
		&audit.AuditLog{},
	))
	return db
}

func retentionTestTimePtr(value time.Time) *time.Time {
	return &value
}

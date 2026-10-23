package repository

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
)

func TestFindCustomerServiceConversationAttachmentReferenceScanDeduplicatesValidPayloads(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&ticket.TicketMessage{}))

	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID:    17,
		Content:     "valid attachment payload",
		Attachments: `["uploads/customer-service/a.txt", "uploads/customer-service/b.txt"]`,
	}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID:    17,
		Content:     "duplicate and malformed payloads",
		Attachments: `["uploads/customer-service/b.txt", "uploads/customer-service/c.txt"]`,
	}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID:    17,
		Content:     "legacy malformed payload",
		Attachments: `not-json`,
	}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID:    18,
		Content:     "different conversation",
		Attachments: `["uploads/customer-service/other.txt"]`,
	}).Error)

	scan, err := NewTicketRepository(db).FindCustomerServiceConversationAttachmentReferenceScan(17)
	require.NoError(t, err)
	require.Equal(t, []string{
		"uploads/customer-service/a.txt",
		"uploads/customer-service/b.txt",
		"uploads/customer-service/c.txt",
	}, scan.References)
}

func TestFindCustomerServiceConversationAttachmentReferenceScanReportsMalformedPayloads(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&ticket.TicketMessage{}))

	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 19, Attachments: `not-json`}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 19, Attachments: ` `}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 19, Attachments: `["uploads/customer-service/valid.txt"]`}).Error)

	scan, err := NewTicketRepository(db).FindCustomerServiceConversationAttachmentReferenceScan(19)
	require.NoError(t, err)
	require.Equal(t, 1, scan.MalformedPayloads)
	require.Equal(t, []string{"uploads/customer-service/valid.txt"}, scan.References)
}

func TestFindCustomerServiceRetentionRecordsIncludesSoftDeletedHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&ticket.Ticket{}, &ticket.TicketMessage{}))

	deletedAt := time.Now().UTC().Add(-24 * time.Hour)
	deleted := ticket.Ticket{
		TicketNumber: "RETENTION-HISTORY-DELETED",
		UserID:       1,
		Subject:      "retained deleted history",
		Category:     "customer_service",
		Status:       "closed",
		DeletedAt:    gorm.DeletedAt{Time: deletedAt, Valid: true},
	}
	active := ticket.Ticket{
		TicketNumber: "RETENTION-HISTORY-ACTIVE",
		UserID:       1,
		Subject:      "retained active history",
		Category:     "customer_service",
		Status:       "closed",
	}
	require.NoError(t, db.Create(&deleted).Error)
	require.NoError(t, db.Create(&active).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: deleted.ID, Content: "deleted history"}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: active.ID, Content: "active history"}).Error)

	records, total, err := NewTicketRepository(db).FindCustomerServiceRetentionRecords(1, 20, CustomerServiceConversationFilters{
		View:   "all",
		Status: "closed",
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, records, 2)
	seenDeleted := false
	for _, record := range records {
		if record.ID == deleted.ID {
			seenDeleted = record.DeletedAt.Valid
		}
	}
	require.True(t, seenDeleted)
}

func TestFindCustomerServiceRetentionRecordsSearchesAllMessagesAndExactNumericIDsWhileLoadingOnlyLatestPreview(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&ticket.Ticket{}, &ticket.TicketMessage{}))
	require.NoError(t, db.Exec(`CREATE TABLE visitor_profiles (
		id INTEGER PRIMARY KEY, customer_service_visitor_hash TEXT, deleted_at DATETIME,
		email TEXT, cart_session_id TEXT
	)`).Error)
	account := user.User{Email: "retention-buyer@example.test", Username: "retention-buyer", Password: "unused"}
	require.NoError(t, db.Create(&account).Error)
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	conversation := ticket.Ticket{ID: 501, TicketNumber: "HISTORY-SEARCH-TARGET", UserID: account.ID,
		CustomerUserID: &account.ID, Subject: "Retained question", Category: "customer_service", Status: "closed", UpdatedAt: now}
	require.NoError(t, db.Create(&conversation).Error)
	for _, message := range []ticket.TicketMessage{
		{ID: 101, TicketID: conversation.ID, Content: "older searchable question", CreatedAt: now.Add(-time.Hour)},
		{ID: 102, TicketID: conversation.ID, Content: "earlier preview at same timestamp", CreatedAt: now},
		{ID: 103, TicketID: conversation.ID, Content: "latest preview", CreatedAt: now},
		{ID: 104, TicketID: conversation.ID, Content: "", CreatedAt: now.Add(time.Hour)},
	} {
		require.NoError(t, db.Create(&message).Error)
	}
	other := ticket.Ticket{ID: 502, TicketNumber: "HISTORY-501-TEXT-MATCH", UserID: account.ID,
		Subject: "Different conversation", Category: "customer_service", Status: "closed", UpdatedAt: now}
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: other.ID, Content: "other history", CreatedAt: now}).Error)
	for _, search := range []string{"501", "older searchable", account.Email, "HISTORY-SEARCH-TARGET"} {
		t.Run(search, func(t *testing.T) {
			records, total, err := NewTicketRepository(db).FindCustomerServiceRetentionRecords(1, 20,
				CustomerServiceConversationFilters{View: "all", Search: search, Status: "closed"})
			require.NoError(t, err)
			require.EqualValues(t, 1, total)
			require.Len(t, records, 1)
			require.Equal(t, conversation.ID, records[0].ID)
			require.Len(t, records[0].Messages, 1)
			require.Equal(t, "latest preview", records[0].Messages[0].Content)
		})
	}
	firstPage, total, err := NewTicketRepository(db).FindCustomerServiceRetentionRecords(1, 1, CustomerServiceConversationFilters{View: "all"})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, firstPage, 1)
	require.Equal(t, other.ID, firstPage[0].ID)
	secondPage, _, err := NewTicketRepository(db).FindCustomerServiceRetentionRecords(2, 1, CustomerServiceConversationFilters{View: "all"})
	require.NoError(t, err)
	require.Len(t, secondPage, 1)
	require.Equal(t, conversation.ID, secondPage[0].ID)
}

func TestPurgeCustomerServiceConversationWithAuditRollsBackOnAuditFailure(t *testing.T) {
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

	conversation := ticket.Ticket{
		TicketNumber: "RETENTION-ROLLBACK",
		UserID:       42,
		Subject:      "rollback test",
		Category:     "customer_service",
		Status:       "closed",
		DeletedAt:    gorm.DeletedAt{Valid: true},
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: conversation.ID, Content: "message"}).Error)
	require.NoError(t, db.Create(&ticket.CustomerServiceInboxState{TicketID: conversation.ID, RecipientUserID: 7}).Error)

	existingAudit := audit.AuditLog{Action: "delete", Resource: "existing", ResourceID: 999}
	require.NoError(t, db.Create(&existingAudit).Error)
	conflictingAudit := audit.AuditLog{
		ID:         existingAudit.ID,
		Action:     "delete",
		Resource:   "customer_service_retention_purge",
		ResourceID: conversation.ID,
	}
	err = NewTicketRepository(db).PurgeCustomerServiceConversationWithAudit(conversation.ID, &conflictingAudit)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Unscoped().Model(&ticket.Ticket{}).Where("id = ?", conversation.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, db.Unscoped().Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, db.Unscoped().Model(&ticket.CustomerServiceInboxState{}).Where("ticket_id = ?", conversation.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

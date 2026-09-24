package repository

import (
	"testing"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/ticket"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

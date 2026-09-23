package scheduler

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCustomerServiceRetentionSchedulerRunOncePurgesExpiredTombstone(t *testing.T) {
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

	now := time.Now().UTC()
	closedAt := now.Add(-90 * 24 * time.Hour)
	conversation := ticket.Ticket{
		TicketNumber: "RETENTION-SCHEDULER-1",
		UserID:       42,
		Subject:      "scheduler drill",
		Category:     "customer_service",
		Status:       "closed",
		UpdatedAt:    closedAt,
		ClosedAt:     &closedAt,
		DeletedAt:    gorm.DeletedAt{Time: now.Add(-31 * 24 * time.Hour), Valid: true},
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{
		TicketID: conversation.ID,
		Content:  "scheduler message",
	}).Error)

	retention := service.NewCustomerServiceRetentionService(
		repository.NewTicketRepository(db),
		nil,
		nil,
		nil,
		nil,
		service.NewAuditService(repository.NewAuditRepository(db)),
	)
	retention.ConfigureWindows(24*time.Hour, 30*24*time.Hour)
	scheduler := NewCustomerServiceRetentionScheduler(retention, config.WorkerConfig{
		CustomerServiceRetentionEnabled:    true,
		CustomerServiceRetentionBatchLimit: 1,
	})

	result, err := scheduler.RunOnce()
	require.NoError(t, err)
	require.Equal(t, 1, result.Scanned)
	require.Equal(t, 1, result.Purged)
	require.Zero(t, result.Blocked)
	require.Zero(t, result.Errors)

	var remaining int64
	require.NoError(t, db.Unscoped().Model(&ticket.Ticket{}).Where("id = ?", conversation.ID).Count(&remaining).Error)
	require.Zero(t, remaining)
	var auditCount int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Where("resource_id = ? AND resource = ?", conversation.ID, "customer_service_retention_purge").Count(&auditCount).Error)
	require.EqualValues(t, 1, auditCount)
}

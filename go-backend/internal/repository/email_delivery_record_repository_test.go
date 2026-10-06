package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/notification"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEmailDeliveryRecordRepositoryUpdatesOneRecordAcrossRetries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailDeliveryRecord{}))

	repo := NewEmailDeliveryRecordRepository(db)
	firstAttempt := time.Date(2026, 10, 6, 1, 2, 3, 0, time.UTC)
	require.NoError(t, repo.UpsertEmailDeliveryRecordForSending(&notification.EmailDeliveryRecord{
		OutboxEventID:   8,
		EventKey:        "notification:payment-8:order_confirmation",
		EventType:       "order.payment_succeeded",
		TemplateCode:    "order_confirmation",
		TemplateLocale:  "en",
		TemplateVersion: 2,
		RecipientEmail:  "buyer@example.com",
		Subject:         "Order ORD-8",
		ReferenceType:   "order",
		ReferenceNumber: "ORD-8",
		Status:          notification.EmailDeliveryStatusSending,
		AttemptCount:    1,
		FirstAttemptAt:  firstAttempt,
		LastAttemptAt:   firstAttempt,
	}))

	secondAttempt := firstAttempt.Add(time.Minute)
	require.NoError(t, repo.UpsertEmailDeliveryRecordForSending(&notification.EmailDeliveryRecord{
		OutboxEventID:   8,
		EventKey:        "notification:payment-8:order_confirmation",
		EventType:       "order.payment_succeeded",
		TemplateCode:    "order_confirmation",
		TemplateLocale:  "en",
		TemplateVersion: 2,
		RecipientEmail:  "buyer@example.com",
		Subject:         "Order ORD-8",
		ReferenceType:   "order",
		ReferenceNumber: "ORD-8",
		Status:          notification.EmailDeliveryStatusSending,
		AttemptCount:    2,
		FirstAttemptAt:  secondAttempt,
		LastAttemptAt:   secondAttempt,
	}))
	require.NoError(t, repo.MarkEmailDeliveryRecordSent("notification:payment-8:order_confirmation", secondAttempt.Add(time.Second)))

	records, total, err := repo.ListEmailDeliveryRecords(EmailDeliveryRecordListFilters{Search: "ord-8", Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, records, 1)
	assert.Equal(t, notification.EmailDeliveryStatusSent, records[0].Status)
	assert.Equal(t, 2, records[0].AttemptCount)
	assert.Equal(t, firstAttempt, records[0].FirstAttemptAt)
	assert.Equal(t, "b***r@example.com", records[0].ToAdminEmailDeliveryRecordView().RecipientEmail)
}

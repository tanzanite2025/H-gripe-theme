package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerServiceReplyMetricsPairsLatestCustomerTurn(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	messages := []ticket.TicketMessage{
		{CreatedAt: start.Add(1 * time.Minute)},
		{CreatedAt: start.Add(2 * time.Minute)},
		{IsStaff: true, CreatedAt: start.Add(5 * time.Minute)},
		{CreatedAt: start.Add(10 * time.Minute)},
		{IsInternal: true, CreatedAt: start.Add(11 * time.Minute)},
		{IsStaff: true, CreatedAt: end},
	}

	totalSeconds, replyCount, unansweredTurns := customerServiceReplyMetrics(messages, start, end)

	assert.Equal(t, 180.0, totalSeconds)
	assert.Equal(t, 1, replyCount)
	assert.Equal(t, 1, unansweredTurns)
}

func TestCustomerServiceReplyMetricsIncludesPendingTurnBeforeWindow(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	messages := []ticket.TicketMessage{
		{CreatedAt: start.Add(-30 * time.Minute)},
		{IsStaff: true, CreatedAt: start.Add(30 * time.Minute)},
	}

	totalSeconds, replyCount, unansweredTurns := customerServiceReplyMetrics(messages, start, end)

	assert.Equal(t, 3600.0, totalSeconds)
	assert.Equal(t, 1, replyCount)
	assert.Equal(t, 0, unansweredTurns)
}

func TestCustomerServiceContextProductImageUsesCanonicalPublicURL(t *testing.T) {
	resolver := NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	item := &product.Product{
		Media: []product.ProductMedia{{
			MediaType: "image",
			IsVisible: true,
			URL:       "http://media.internal:8080/uploads/products/wheel.jpg",
		}},
	}

	require.Equal(t, "https://shop.example.test/uploads/products/wheel.jpg", firstProductImage(item, resolver))
	require.Equal(t, "http://media.internal:8080/uploads/products/wheel.jpg", item.Media[0].URL)
}

func TestCustomerServiceContextAppliesVisitorTimezoneBeforeAccountTimezone(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&visitor.Profile{}))

	userID := uint(42)
	require.NoError(t, db.Create(&visitor.Profile{
		UserID:     &userID,
		Timezone:   "Europe/Berlin",
		LastSeenAt: time.Now().UTC(),
	}).Error)
	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "visitor-hash",
		UserID:                     &userID,
		Timezone:                   "America/Los_Angeles",
		LastSeenAt:                 time.Now().UTC(),
	}).Error)

	contextService := &CustomerServiceContextService{
		visitorProfileService: NewVisitorProfileService(repository.NewVisitorProfileRepository(db)),
	}
	context := &CustomerServiceContext{
		Contact: CustomerServiceContextContact{TimezoneSource: "not_captured"},
	}

	contextService.applyVisitorTimezone(context, "visitor-hash", userID)

	assert.Equal(t, "America/Los_Angeles", context.Contact.Timezone)
	assert.Equal(t, "visitor_profile", context.Contact.TimezoneSource)
}

func TestCustomerServiceContextFallsBackToAccountTimezoneWhenVisitorProfileMissing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&visitor.Profile{}))

	userID := uint(84)
	require.NoError(t, db.Create(&visitor.Profile{
		UserID:     &userID,
		Timezone:   "Asia/Tokyo",
		LastSeenAt: time.Now().UTC(),
	}).Error)

	contextService := &CustomerServiceContextService{
		visitorProfileService: NewVisitorProfileService(repository.NewVisitorProfileRepository(db)),
	}
	context := &CustomerServiceContext{
		Contact: CustomerServiceContextContact{TimezoneSource: "not_captured"},
	}

	contextService.applyVisitorTimezone(context, "missing-visitor-profile", userID)

	assert.Equal(t, "Asia/Tokyo", context.Contact.Timezone)
	assert.Equal(t, "visitor_profile", context.Contact.TimezoneSource)
}

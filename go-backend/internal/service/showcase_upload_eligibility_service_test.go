package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShowcaseUploadEligibilityListsOnlyOwnedOrdersAndMarksCompletedOrders(t *testing.T) {
	db := newShowcaseUploadEligibilityTestDB(t)
	repo := repository.NewOrderRepository(db)
	service := NewShowcaseUploadEligibilityService(repo)
	completedAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)

	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          1,
		UserID:      7,
		OrderNumber: "ORDER-1",
		Status:      "completed",
		CompletedAt: &completedAt,
	})
	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          2,
		UserID:      7,
		OrderNumber: "ORDER-2",
		Status:      "shipped",
	})
	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          3,
		UserID:      7,
		OrderNumber: "ORDER-3",
		Status:      "completed",
	})
	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          4,
		UserID:      8,
		OrderNumber: "ORDER-4",
		Status:      "completed",
		CompletedAt: &completedAt,
	})

	options, err := service.ListOrderOptions(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, options, 3)

	byID := make(map[uint]ShowcaseUploadOrderOption, len(options))
	for _, option := range options {
		byID[option.ID] = option
	}

	assert.True(t, byID[1].Eligible)
	assert.False(t, byID[2].Eligible)
	assert.False(t, byID[3].Eligible)
	_, otherUserOrderReturned := byID[4]
	assert.False(t, otherUserOrderReturned)
}

func TestShowcaseUploadEligibilityRequiresOwnedCompletedOrder(t *testing.T) {
	db := newShowcaseUploadEligibilityTestDB(t)
	repo := repository.NewOrderRepository(db)
	service := NewShowcaseUploadEligibilityService(repo)
	completedAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)

	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          11,
		UserID:      7,
		OrderNumber: "ORDER-11",
		Status:      "completed",
		CompletedAt: &completedAt,
	})
	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          12,
		UserID:      7,
		OrderNumber: "ORDER-12",
		Status:      "processing",
	})
	createShowcaseEligibilityTestOrder(t, db, order.Order{
		ID:          13,
		UserID:      8,
		OrderNumber: "ORDER-13",
		Status:      "completed",
		CompletedAt: &completedAt,
	})

	eligible, err := service.RequireEligibleOrder(context.Background(), 7, 11)
	require.NoError(t, err)
	assert.Equal(t, uint(11), eligible.ID)

	_, err = service.RequireEligibleOrder(context.Background(), 7, 12)
	require.ErrorIs(t, err, ErrShowcaseUploadOrderNotEligible)

	_, err = service.RequireEligibleOrder(context.Background(), 7, 13)
	require.ErrorIs(t, err, ErrShowcaseUploadOrderNotEligible)

	_, err = service.RequireEligibleOrder(context.Background(), 7, 0)
	require.ErrorIs(t, err, ErrShowcaseUploadOrderRequired)
}

func newShowcaseUploadEligibilityTestDB(t *testing.T) *gorm.DB {
	t.Helper()

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

	require.NoError(t, db.AutoMigrate(&order.Order{}))
	return db
}

func createShowcaseEligibilityTestOrder(t *testing.T, db *gorm.DB, item order.Order) {
	t.Helper()
	item.TotalAmount = 100
	item.Currency = "USD"
	require.NoError(t, db.Create(&item).Error)
}

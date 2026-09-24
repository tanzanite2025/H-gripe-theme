package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/order"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderStatsAggregateExactMinorRevenueByCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}))

	now := time.Now()
	orders := []order.Order{
		{OrderNumber: "ORD-STATS-USD-1", Status: "paid", Currency: "USD", TotalAmountMinor: 101},
		{OrderNumber: "ORD-STATS-USD-2", Status: "completed", Currency: "USD", TotalAmountMinor: 202, CreatedAt: now},
		{OrderNumber: "ORD-STATS-JPY", Status: "paid", Currency: "JPY", TotalAmountMinor: 303},
		{OrderNumber: "ORD-STATS-CANCELLED", Status: "cancelled", Currency: "USD", TotalAmountMinor: 100000},
	}
	for i := range orders {
		if orders[i].CreatedAt.IsZero() {
			orders[i].CreatedAt = now
		}
		require.NoError(t, db.Create(&orders[i]).Error)
	}

	repo := NewOrderRepository(db)
	stats, err := repo.GetStats()
	require.NoError(t, err)

	totalRevenue, ok := stats["total_revenue_by_currency"].([]RevenueByCurrency)
	require.True(t, ok)
	require.Equal(t, []RevenueByCurrency{
		{Currency: "JPY", AmountMinor: 303},
		{Currency: "USD", AmountMinor: 303},
	}, totalRevenue)

	todayRevenue, ok := stats["today_revenue_by_currency"].([]RevenueByCurrency)
	require.True(t, ok)
	require.Equal(t, totalRevenue, todayRevenue)
}

func TestSalesByDateRangeKeepsCurrencyRevenueSeparate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}))

	day := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	for _, record := range []order.Order{
		{OrderNumber: "ORD-CHART-USD", Status: "paid", Currency: "USD", TotalAmountMinor: 101, CreatedAt: day},
		{OrderNumber: "ORD-CHART-JPY", Status: "paid", Currency: "JPY", TotalAmountMinor: 202, CreatedAt: day.Add(time.Hour)},
		{OrderNumber: "ORD-CHART-CANCELLED", Status: "cancelled", Currency: "USD", TotalAmountMinor: 999999, CreatedAt: day},
	} {
		require.NoError(t, db.Create(&record).Error)
	}

	data, err := NewOrderRepository(db).GetSalesByDateRange(day.Add(-time.Hour), day.Add(2*time.Hour))
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.Equal(t, int64(2), data[0]["count"])
	require.Equal(t, []RevenueByCurrency{
		{Currency: "JPY", AmountMinor: 202},
		{Currency: "USD", AmountMinor: 101},
	}, data[0]["revenue_by_currency"])
}

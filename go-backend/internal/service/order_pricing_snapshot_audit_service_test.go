package service

import (
	"context"
	"fmt"
	"testing"

	"commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/pricing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderPricingSnapshotAuditIsReadOnlyAndReportsIssues(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&order.Order{}, &order.OrderItem{}))

	validOrderSnapshot, err := pricing.MarshalOrderPricingSnapshot(pricing.OrderPricingSnapshotInput{
		Currency: "USD", Subtotal: money.MustNew(1000, "USD"), Shipping: money.MustNew(100, "USD"),
		Tax: money.MustNew(100, "USD"), MemberDiscount: money.MustNew(50, "USD"), CouponDiscount: money.MustNew(40, "USD"),
		DiscountTotal: money.MustNew(90, "USD"), Total: money.MustNew(1110, "USD"),
	})
	require.NoError(t, err)
	lineSnapshot, err := pricing.NewSnapshot([]pricing.LineInput{{Key: "0:10:20", ProductID: 10, VariantID: 20, Quantity: 2, UnitPrice: money.MustNew(500, "USD")}})
	require.NoError(t, err)
	lineRaw, err := pricing.MarshalLineSnapshot(lineSnapshot.Lines()[0])
	require.NoError(t, err)
	variantID := uint(20)
	require.NoError(t, db.Create(&order.Order{
		ID: 1, OrderNumber: "ORD-VALID", Currency: "USD", SubtotalAmountMinor: 1000, ShippingFeeMinor: 100, TaxAmountMinor: 100,
		DiscountAmountMinor: 90, TotalAmountMinor: 1110, PricingSnapshotData: validOrderSnapshot,
		Items: []order.OrderItem{{ID: 11, ProductID: 10, VariantID: &variantID, Currency: "USD", Quantity: 2, PriceMinor: 500, SubtotalMinor: 1000, TotalMinor: 1000, PricingSnapshotData: lineRaw}},
	}).Error)
	require.NoError(t, db.Create(&order.Order{ID: 2, OrderNumber: "ORD-MISSING", Currency: "USD"}).Error)
	require.NoError(t, db.Create(&order.Order{ID: 3, OrderNumber: "ORD-INVALID", Currency: "USD", PricingSnapshotData: []byte(`{"schema_version":99}`)}).Error)

	report, err := NewOrderPricingSnapshotAuditService(db).Audit(context.Background(), OrderPricingSnapshotAuditOptions{BatchSize: 1, MaxIssues: 10})
	require.NoError(t, err)
	require.Equal(t, 3, report.OrdersScanned)
	require.Equal(t, 1, report.OrdersWithValidSnapshot)
	require.Equal(t, 1, report.OrdersMissingSnapshot)
	require.Equal(t, 1, report.OrdersInvalidSnapshot)
	require.Equal(t, 1, report.ItemsScanned)
	require.Equal(t, 1, report.ItemsWithValidSnapshot)
	require.Equal(t, 2, report.TotalIssues)

	var persisted order.Order
	require.NoError(t, db.Preload("Items").First(&persisted, 1).Error)
	require.Equal(t, string(validOrderSnapshot), string(persisted.PricingSnapshotData))
	require.Equal(t, string(lineRaw), string(persisted.Items[0].PricingSnapshotData))
}

func TestOrderPricingSnapshotAuditBoundsIssuesAndOrders(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&order.Order{}, &order.OrderItem{}))
	for id := uint(1); id <= 3; id++ {
		require.NoError(t, db.Create(&order.Order{ID: id, OrderNumber: "ORD-" + fmt.Sprint(id), Currency: "USD"}).Error)
	}
	report, err := NewOrderPricingSnapshotAuditService(db).Audit(context.Background(), OrderPricingSnapshotAuditOptions{BatchSize: 2, MaxOrders: 2, MaxIssues: 1})
	require.NoError(t, err)
	require.Equal(t, 2, report.OrdersScanned)
	require.True(t, report.OrdersTruncated)
	require.Equal(t, 2, report.TotalIssues)
	require.Len(t, report.Issues, 1)
	require.True(t, report.IssuesTruncated)
}

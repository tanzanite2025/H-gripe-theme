package repository

import (
	"encoding/json"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShipmentRecordRepositoryListsShippedOrdersWithoutCreatingAttachments(t *testing.T) {
	db := newShipmentRecordTestDB(t)
	shippedAt := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	orderRecord := createShippedOrder(t, db, shippedAt)

	repo := NewShipmentRecordRepository(db)
	records, total, err := repo.FindAll(1, 20, ShipmentRecordFilter{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, records, 1)
	require.Equal(t, orderRecord.ID, records[0].ID)
	require.Equal(t, orderRecord.ID, records[0].OrderID)
	require.False(t, records[0].RecordBound)
	require.Equal(t, "shipped", records[0].OrderStatus)
	require.Equal(t, "shipped", records[0].ShippingState)
	require.Contains(t, string(records[0].ItemsSnapshot), `"product_name":"Snapshot wheel"`)

	var attachments int64
	require.NoError(t, db.Model(&shippingdomain.ShipmentRecord{}).Count(&attachments).Error)
	require.Equal(t, int64(0), attachments)
}

func TestShipmentRecordRepositoryWritesOnlyWhenDetailsAreAttached(t *testing.T) {
	db := newShipmentRecordTestDB(t)
	shippedAt := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	orderRecord := createShippedOrder(t, db, shippedAt)
	repo := NewShipmentRecordRepository(db)

	start := shippedAt.Add(24 * time.Hour)
	created, err := repo.UpsertDetailsForOrder(
		orderRecord.ID,
		"拍摄了包装和轮组外观",
		[]string{"https://cdn.example/shipment-1.jpg"},
		[]string{"WHEEL-7-SERIAL"},
		24,
		start,
		start.AddDate(0, 24, 0),
	)
	require.NoError(t, err)
	require.Equal(t, orderRecord.ID, created.ID)
	require.Equal(t, orderRecord.ID, created.OrderID)
	require.True(t, created.RecordBound)
	require.Equal(t, "拍摄了包装和轮组外观", created.ShippingNote)
	require.JSONEq(t, `["https://cdn.example/shipment-1.jpg"]`, string(created.ShippingImages))
	require.JSONEq(t, `["WHEEL-7-SERIAL"]`, string(created.ProductCodes))
	require.Equal(t, 24, created.WarrantyMonths)
	require.True(t, created.WarrantyStartAt.Equal(start))
	require.True(t, created.WarrantyExpires.Equal(start.AddDate(0, 24, 0)))

	updated, err := repo.FindByOrderID(orderRecord.ID)
	require.NoError(t, err)
	require.True(t, updated.RecordBound)
	var codes []string
	require.NoError(t, json.Unmarshal(updated.ProductCodes, &codes))
	require.Equal(t, []string{"WHEEL-7-SERIAL"}, codes)
}

func newShipmentRecordTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&orderdomain.Order{}, &orderdomain.OrderItem{}, &shippingdomain.ShipmentRecord{}))
	return db
}

func createShippedOrder(t *testing.T, db *gorm.DB, shippedAt time.Time) *orderdomain.Order {
	t.Helper()
	variantID := uint(17)
	orderRecord := &orderdomain.Order{
		OrderNumber:    "TZ-2026-TEST-SHIPMENT",
		UserID:         42,
		Status:         "shipped",
		ShippingStatus: "shipped",
		Currency:       "USD",
		TotalAmount:    100,
		SubtotalAmount: 100,
		ShippingAddress: orderdomain.Address{
			FirstName: "Test",
			LastName:  "Buyer",
			Email:     "buyer@example.com",
		},
		ShippedAt: &shippedAt,
		Items: []orderdomain.OrderItem{{
			ProductID:   7,
			VariantID:   &variantID,
			ProductName: "Snapshot wheel",
			SKU:         "WHEEL-7",
			Quantity:    1,
		}},
	}
	require.NoError(t, db.Create(orderRecord).Error)
	return orderRecord
}

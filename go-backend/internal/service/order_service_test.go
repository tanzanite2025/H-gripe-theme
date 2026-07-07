package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/order"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderServiceCreateOrderPersistsPricingAndAdjustments(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	seedUserLoyalty(t, db, userID, 1000)
	seedCoupon(t, db, "SAVE10", "fixed", 10, 1)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 2}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"SAVE10",
		100,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	require.NotZero(t, createdOrder.ID)
	assert.InDelta(t, 100, createdOrder.SubtotalAmount, 0.001)
	assert.InDelta(t, 16, createdOrder.DiscountAmount, 0.001)
	assert.InDelta(t, 84, createdOrder.TotalAmount, 0.001)
	assert.Equal(t, 100, createdOrder.PointsUsed)
	assert.InDelta(t, 1, createdOrder.PointsValue, 0.001)
	assert.Equal(t, "SAVE10", createdOrder.CouponCode)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 1)
	require.NotNil(t, savedOrder.Items[0].VariantID)
	assert.Equal(t, productRecord.Name, savedOrder.Items[0].ProductName)
	assert.Equal(t, productRecord.SKU, savedOrder.Items[0].SKU)
	assert.InDelta(t, 100, savedOrder.Items[0].Subtotal, 0.001)
	assert.InDelta(t, 100, savedOrder.Items[0].Total, 0.001)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 3, savedProduct.Stock)

	var savedLoyalty loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", userID).First(&savedLoyalty).Error)
	assert.Equal(t, 900, savedLoyalty.AvailablePoints)
	assert.Equal(t, 100, savedLoyalty.UsedPoints)

	var pointTransaction loyalty.LoyaltyTransaction
	require.NoError(t, db.Where("user_id = ? AND source = ? AND source_id = ?", userID, "order", createdOrder.ID).First(&pointTransaction).Error)
	assert.Equal(t, -100, pointTransaction.Points)
	assert.Equal(t, 900, pointTransaction.Balance)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.Where("code = ?", "SAVE10").First(&savedCoupon).Error)
	assert.Equal(t, 1, savedCoupon.UsedCount)

	var usage coupon.CouponUsage
	require.NoError(t, db.Where("coupon_id = ? AND order_id = ?", savedCoupon.ID, createdOrder.ID).First(&usage).Error)
	assert.InDelta(t, 10, usage.Discount, 0.001)
}

func TestOrderServiceCreateOrderUsesVariantPricingAndStock(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProductShell(t, db, 999, 99)
	salePrice := 80.0
	variant := product.ProductVariant{
		ProductID:    productRecord.ID,
		SKU:          "SKU-TEST-BLK-24H",
		Title:        "Black / 24H",
		OptionValues: `{"color":"black","spoke_holes":"24"}`,
		Price:        90,
		SalePrice:    &salePrice,
		Stock:        3,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, VariantID: &variant.ID, Quantity: 2}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.InDelta(t, 160, createdOrder.SubtotalAmount, 0.001)
	assert.InDelta(t, 160, createdOrder.TotalAmount, 0.001)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 1)
	require.NotNil(t, savedOrder.Items[0].VariantID)
	assert.Equal(t, variant.ID, *savedOrder.Items[0].VariantID)
	assert.Equal(t, variant.SKU, savedOrder.Items[0].SKU)
	assert.Equal(t, variant.OptionValues, savedOrder.Items[0].Attributes)
	assert.InDelta(t, 80, savedOrder.Items[0].Price, 0.001)
	assert.InDelta(t, 160, savedOrder.Items[0].Subtotal, 0.001)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 1, savedVariant.Stock)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 1, savedProduct.Stock)
}

func TestOrderServiceCreateOrderRollsBackWhenStockIsInsufficient(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 1)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 2}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)

	require.Error(t, err)
	assert.Nil(t, createdOrder)
	assert.True(t, strings.Contains(strings.ToLower(err.Error()), "stock"))

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Equal(t, int64(0), orderCount)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 1, savedProduct.Stock)
}

func TestOrderServiceCreateOrderRejectsProductWithoutVariant(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProductShell(t, db, 50, 1)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)

	require.Error(t, err)
	assert.Nil(t, createdOrder)
	assert.True(t, strings.Contains(strings.ToLower(err.Error()), "not found"))

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Equal(t, int64(0), orderCount)
}

func TestOrderStatusTransitionUsesDomainRules(t *testing.T) {
	assert.False(t, (&order.Order{Status: "pending"}).CanTransitionTo("paid"))
	assert.True(t, (&order.Order{Status: "shipped"}).CanTransitionTo("completed"))
	assert.False(t, (&order.Order{Status: "shipped"}).CanTransitionTo("delivered"))
	assert.False(t, (&order.Order{Status: "paid"}).CanTransitionTo("refunded"))
	assert.False(t, (&order.Order{Status: "cancelled"}).CanTransitionTo("paid"))
}

func TestOrderServiceRejectsPaymentManagedStatusUpdates(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:   "ORD-SYSTEM-STATUS",
		UserID:        42,
		Status:        "pending",
		PaymentStatus: "unpaid",
		TotalAmount:   100,
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "paid"), ErrSystemManagedOrderStatus)
	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "refunded"), ErrSystemManagedOrderStatus)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "pending", savedOrder.Status)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)
}

func TestOrderServiceGenerateOrderNumberFormat(t *testing.T) {
	orderNumber := (&OrderService{}).generateOrderNumber()

	assert.True(t, strings.HasPrefix(orderNumber, "ORD"+time.Now().Format("20060102")))
	assert.Len(t, orderNumber, 19)
}

func newTestOrderService(t *testing.T) (*gorm.DB, *OrderService) {
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

	require.NoError(t, db.AutoMigrate(
		&product.ProductType{},
		&product.SpecDefinition{},
		&product.Product{},
		&product.ProductImage{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&order.Order{},
		&order.OrderItem{},
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&loyalty.UserLoyalty{},
		&loyalty.LoyaltyTransaction{},
		&paymentdomain.TaxRate{},
	))

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	checkoutService := NewCheckoutService(productRepo, couponRepo, paymentRepo, loyaltyRepo)
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo)

	return db, NewOrderService(txManager, orderRepo, checkoutService)
}

func seedProduct(t *testing.T, db *gorm.DB, price float64, stock int) product.Product {
	t.Helper()

	record := seedProductShell(t, db, price, stock)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:    record.ID,
		SKU:          record.SKU,
		Title:        "Default",
		OptionValues: "{}",
		Price:        price,
		Stock:        stock,
		IsDefault:    true,
		IsActive:     true,
	}).Error)
	return record
}

func seedProductShell(t *testing.T, db *gorm.DB, price float64, stock int) product.Product {
	t.Helper()

	record := product.Product{
		SKU:   "SKU-TEST",
		Name:  "Test Product",
		Slug:  "test-product",
		Price: price,
		Stock: stock,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedUserLoyalty(t *testing.T, db *gorm.DB, userID uint, points int) {
	t.Helper()

	require.NoError(t, db.Create(&loyalty.UserLoyalty{
		UserID:          userID,
		TotalPoints:     points,
		AvailablePoints: points,
	}).Error)
}

func seedCoupon(t *testing.T, db *gorm.DB, code, couponType string, value float64, usageLimit int) {
	t.Helper()

	now := time.Now()
	require.NoError(t, db.Create(&coupon.Coupon{
		Code:       code,
		Type:       couponType,
		Value:      value,
		UsageLimit: usageLimit,
		StartDate:  now.Add(-time.Hour),
		EndDate:    now.Add(time.Hour),
		Enabled:    true,
	}).Error)
}

func testAddress() order.Address {
	return order.Address{
		FirstName:  "Test",
		LastName:   "Buyer",
		Address1:   "123 Test Street",
		City:       "Test City",
		State:      "CA",
		PostalCode: "90001",
		Country:    "US",
		Phone:      "1234567890",
		Email:      "buyer@example.com",
	}
}

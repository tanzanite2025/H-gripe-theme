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
	"tanzanite/internal/domain/setting"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/ordernumber"
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

func TestOrderServiceCreateOrderUsesVersionedLoyaltyExchangeRate(t *testing.T) {
	db, orderService := newTestOrderService(t)
	programService := newTestLoyaltyProgramService(t, db)
	config, err := programService.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		ExchangeRatePoints:        80,
		MinRedeemPoints:           0,
		MaxValuePerDayCents:       50000,
		CardExpiryDays:            365,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
		RedeemValuesCents:         []int64{1000},
	})
	require.NoError(t, err)
	orderService.checkout.ConfigureLoyaltyProgram(programService)

	userID := uint(43)
	productRecord := seedProduct(t, db, 100, 5)
	seedUserLoyalty(t, db, userID, 1000)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		800,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, 800, createdOrder.PointsUsed)
	assert.InDelta(t, 10, createdOrder.PointsValue, 0.001)

	var pointTransaction loyalty.LoyaltyTransaction
	require.NoError(t, db.Where("user_id = ? AND source = ? AND source_id = ?", userID, "order", createdOrder.ID).First(&pointTransaction).Error)
	require.NotNil(t, pointTransaction.ProgramConfigID)
	assert.Equal(t, config.ID, *pointTransaction.ProgramConfigID)
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
		Weight:       11000,
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

func TestOrderServiceCreateOrderPersistsSelectedCarrierService(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)

	var template shippingdomain.ShippingTemplate
	require.NoError(t, db.Where("name = ?", "Test standard shipping").First(&template).Error)

	carrier := shippingdomain.Carrier{
		Name:    "DHL",
		Code:    "DHL",
		Enabled: true,
	}
	require.NoError(t, db.Create(&carrier).Error)

	carrierService := shippingdomain.CarrierService{
		CarrierID:   carrier.ID,
		TemplateID:  &template.ID,
		ServiceCode: "EXP-US",
		ServiceName: "Express",
		RouteName:   "US Express",
		Countries:   `["US"]`,
		Currency:    "USD",
		BillingMode: "actual_weight",
		Enabled:     true,
		Description: "Default checkout route",
	}
	require.NoError(t, db.Create(&carrierService).Error)

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

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	require.NotNil(t, createdOrder.CarrierID)
	require.NotNil(t, createdOrder.CarrierServiceID)
	assert.Equal(t, carrier.ID, *createdOrder.CarrierID)
	assert.Equal(t, carrierService.ID, *createdOrder.CarrierServiceID)
	assert.Equal(t, "DHL / US Express / Express (EXP-US)", createdOrder.ShippingMethod)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	require.NotNil(t, savedOrder.CarrierID)
	require.NotNil(t, savedOrder.CarrierServiceID)
	assert.Equal(t, carrier.ID, *savedOrder.CarrierID)
	assert.Equal(t, carrierService.ID, *savedOrder.CarrierServiceID)
	assert.Equal(t, "DHL / US Express / Express (EXP-US)", savedOrder.ShippingMethod)
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

func TestOrderServiceExpireStalePendingPaymentsReleasesReservations(t *testing.T) {
	db, orderService := newTestOrderService(t)
	now := time.Now().UTC()
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	seedUserLoyalty(t, db, userID, 1000)
	seedCoupon(t, db, "EXPIRE10", "fixed", 10, 1)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 2}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"EXPIRE10",
		100,
	)
	require.NoError(t, err)

	oldActivity := now.Add(-45 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": oldActivity,
		"updated_at": oldActivity,
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       createdOrder.ID,
		TransactionID: "pi_expire_old",
		PaymentMethod: "stripe",
		Amount:        createdOrder.TotalAmount,
		Currency:      "USD",
		Status:        "requires_action",
		CreatedAt:     oldActivity,
		UpdatedAt:     oldActivity,
	}).Error)

	result, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ScannedCandidates)
	assert.Equal(t, 1, result.ExpiredOrders)
	assert.Equal(t, int64(1), result.ExpiredOpenTransactions)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "payment_expired", savedOrder.Status)
	assert.Equal(t, "expired", savedOrder.PaymentStatus)
	assert.NotNil(t, savedOrder.CancelledAt)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "pi_expire_old").First(&savedTransaction).Error)
	assert.Equal(t, "expired", savedTransaction.Status)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 5, savedProduct.Stock)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 5, savedVariant.Stock)

	var savedLoyalty loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", userID).First(&savedLoyalty).Error)
	assert.Equal(t, 1000, savedLoyalty.AvailablePoints)
	assert.Equal(t, 0, savedLoyalty.UsedPoints)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.Where("code = ?", "EXPIRE10").First(&savedCoupon).Error)
	assert.Equal(t, 0, savedCoupon.UsedCount)

	var usageCount int64
	require.NoError(t, db.Model(&coupon.CouponUsage{}).Where("order_id = ?", createdOrder.ID).Count(&usageCount).Error)
	assert.Equal(t, int64(0), usageCount)
}

func TestOrderServiceExpireStalePendingPaymentsSkipsRecentPaymentActivity(t *testing.T) {
	db, orderService := newTestOrderService(t)
	now := time.Now().UTC()
	productRecord := seedProduct(t, db, 50, 5)
	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"stripe",
		"standard",
		"",
		0,
	)
	require.NoError(t, err)

	oldCreatedAt := now.Add(-2 * time.Hour)
	recentActivity := now.Add(-10 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": oldCreatedAt,
		"updated_at": oldCreatedAt,
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       createdOrder.ID,
		TransactionID: "pi_recent_activity",
		PaymentMethod: "stripe",
		Amount:        createdOrder.TotalAmount,
		Currency:      "USD",
		Status:        "processing",
		CreatedAt:     oldCreatedAt,
		UpdatedAt:     recentActivity,
	}).Error)

	result, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)

	require.NoError(t, err)
	assert.Equal(t, 0, result.ScannedCandidates)
	assert.Equal(t, 0, result.ExpiredOrders)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "pending", savedOrder.Status)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 4, savedProduct.Stock)
}

func TestOrderStatusTransitionUsesDomainRules(t *testing.T) {
	assert.False(t, (&order.Order{Status: "pending"}).CanTransitionTo("paid"))
	assert.True(t, (&order.Order{Status: "pending"}).CanTransitionTo("payment_expired"))
	assert.True(t, (&order.Order{Status: "shipped"}).CanTransitionTo("completed"))
	assert.False(t, (&order.Order{Status: "shipped"}).CanTransitionTo("delivered"))
	assert.False(t, (&order.Order{Status: "paid"}).CanTransitionTo("refunded"))
	assert.False(t, (&order.Order{Status: "cancelled"}).CanTransitionTo("paid"))
	assert.False(t, (&order.Order{Status: "payment_expired"}).CanTransitionTo("paid"))
}

func TestOrderServiceCompletingOrderAwardsPurchasePointsOnce(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	require.NoError(t, db.Create(&loyalty.MemberLevel{
		Name:         "Gold",
		MinPoints:    0,
		MaxPoints:    999999,
		DiscountRate: 0,
		Benefits:     "[]",
		SortOrder:    10,
	}).Error)

	orderRecord := order.Order{
		OrderNumber:    "ORD-COMPLETE-POINTS",
		UserID:         userID,
		Status:         "shipped",
		PaymentStatus:  "paid",
		SubtotalAmount: 120,
		DiscountAmount: 20,
		TotalAmount:    115,
		Currency:       "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.NoError(t, orderService.UpdateOrderStatus(orderRecord.ID, "completed"))

	var earnedTransactions []loyalty.LoyaltyTransaction
	require.NoError(t, db.Where(
		"user_id = ? AND type = ? AND source = ? AND source_id = ?",
		userID,
		"earn",
		"order",
		orderRecord.ID,
	).Find(&earnedTransactions).Error)
	require.Len(t, earnedTransactions, 1)
	assert.Equal(t, 100, earnedTransactions[0].Points)
	assert.Equal(t, 100, earnedTransactions[0].Balance)
	require.NotNil(t, earnedTransactions[0].ProgramConfigID)

	var userLoyalty loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", userID).First(&userLoyalty).Error)
	assert.Equal(t, 100, userLoyalty.AvailablePoints)
	assert.Equal(t, 100, userLoyalty.TotalPoints)

	require.Error(t, orderService.UpdateOrderStatus(orderRecord.ID, "completed"))
	var count int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ? AND source_id = ?", userID, "earn", "order", orderRecord.ID).
		Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestOrderServiceCompletionRejectsNonUSDPointsWithoutConversion(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(43)
	orderRecord := order.Order{
		OrderNumber:    "ORD-COMPLETE-EUR-POINTS",
		UserID:         userID,
		Status:         "shipped",
		PaymentStatus:  "paid",
		SubtotalAmount: 120,
		DiscountAmount: 20,
		TotalAmount:    115,
		Currency:       "EUR",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	err := orderService.UpdateOrderStatus(orderRecord.ID, "completed")

	require.ErrorIs(t, err, ErrInvalidLoyaltyProgramConfig)
	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "shipped", savedOrder.Status)

	var count int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ? AND source_id = ?", userID, "earn", "order", orderRecord.ID).
		Count(&count).Error)
	assert.EqualValues(t, 0, count)
}

func TestOrderServiceCreateOrderUsesDefaultOrderCurrency(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 50, 5)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, "USD", createdOrder.Currency)
}

func TestOrderServiceCreateOrderRejectsPaymentMethodUnsupportedDefaultCurrency(t *testing.T) {
	_, orderService := newTestOrderService(t)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: 1, Quantity: 1}},
		testAddress(),
		testAddress(),
		"wechat",
		"standard",
		"",
		0,
	)

	require.Error(t, err)
	assert.Nil(t, createdOrder)
	assert.Contains(t, err.Error(), "cannot collect order currency USD")
}

func TestOrderServiceRejectsPaymentManagedStatusUpdates(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:   "ORD-SYSTEM-STATUS",
		UserID:        42,
		Status:        "pending",
		PaymentStatus: "unpaid",
		TotalAmount:   100,
		Currency:      "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "paid"), ErrSystemManagedOrderStatus)
	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "refunded"), ErrSystemManagedOrderStatus)
	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "payment_expired"), ErrSystemManagedOrderStatus)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "pending", savedOrder.Status)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)
}

func TestOrderServiceUpdateTrackingInfoResolvesProviderCarrierCode(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, carrier, carrierService := seedTrackingProviderCarrierAndService(t, db)
	mapping := seedTrackingCarrierMapping(t, db, provider.ID, "carrier", &carrier.ID, nil, "DHL")

	orderRecord := order.Order{
		OrderNumber: "ORD-TRACKING",
		UserID:      42,
		Status:      "processing",
		TotalAmount: 100,
		Currency:    "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	err := orderService.UpdateTrackingInfo(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK123456",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})

	require.NoError(t, err)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "TRACK123456", savedOrder.TrackingNumber)
	assert.NotNil(t, savedOrder.TrackingProviderID)
	assert.Equal(t, provider.ID, *savedOrder.TrackingProviderID)
	assert.NotNil(t, savedOrder.CarrierID)
	assert.Equal(t, carrier.ID, *savedOrder.CarrierID)
	assert.NotNil(t, savedOrder.CarrierServiceID)
	assert.Equal(t, carrierService.ID, *savedOrder.CarrierServiceID)
	assert.NotNil(t, savedOrder.TrackingCarrierMappingID)
	assert.Equal(t, mapping.ID, *savedOrder.TrackingCarrierMappingID)
	assert.Equal(t, "DHL", savedOrder.ProviderCarrierCode)

	var shipment shippingdomain.TrackingShipment
	require.NoError(t, db.Where("order_id = ?", orderRecord.ID).First(&shipment).Error)
	assert.Equal(t, provider.ID, shipment.TrackingProviderID)
	assert.Equal(t, "TRACK123456", shipment.TrackingNumber)
	assert.Equal(t, "DHL", shipment.ProviderCarrierCode)
	assert.Equal(t, "pending", shipment.RegistrationStatus)
	assert.Equal(t, "pending", shipment.SyncStatus)
}

func TestOrderServiceUpdateTrackingInfoDefaultsToOrderCarrierService(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, carrier, carrierService := seedTrackingProviderCarrierAndService(t, db)
	mapping := seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-TRACKING-DEFAULT-SERVICE",
		UserID:           42,
		Status:           "processing",
		CarrierID:        &carrier.ID,
		CarrierServiceID: &carrierService.ID,
		TotalAmount:      100,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	err := orderService.UpdateTrackingInfo(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACKDEFAULT123",
		TrackingProviderID: provider.ID,
	})

	require.NoError(t, err)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "TRACKDEFAULT123", savedOrder.TrackingNumber)
	require.NotNil(t, savedOrder.CarrierID)
	require.NotNil(t, savedOrder.CarrierServiceID)
	require.NotNil(t, savedOrder.TrackingCarrierMappingID)
	assert.Equal(t, carrier.ID, *savedOrder.CarrierID)
	assert.Equal(t, carrierService.ID, *savedOrder.CarrierServiceID)
	assert.Equal(t, mapping.ID, *savedOrder.TrackingCarrierMappingID)
	assert.Equal(t, "DHL-EXP-US", savedOrder.ProviderCarrierCode)

	var shipment shippingdomain.TrackingShipment
	require.NoError(t, db.Where("order_id = ?", orderRecord.ID).First(&shipment).Error)
	assert.Equal(t, carrier.ID, *shipment.CarrierID)
	assert.Equal(t, carrierService.ID, *shipment.CarrierServiceID)
	assert.Equal(t, "DHL-EXP-US", shipment.ProviderCarrierCode)
}

func TestOrderServiceSyncOrderTrackingUsesStoredTrackingSource(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Provider",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)

	orderRecord := order.Order{
		OrderNumber:         "ORD-SYNC-TRACKING",
		UserID:              42,
		Status:              "shipped",
		TrackingNumber:      "MOCK123456",
		TrackingProviderID:  &provider.ID,
		ProviderCarrierCode: "DHL",
		TotalAmount:         100,
		Currency:            "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	result, err := orderService.SyncOrderTracking(context.Background(), orderRecord.ID)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Events, 2)
	require.NotNil(t, result.Shipment)
	assert.Equal(t, "synced", result.Shipment.SyncStatus)
	assert.Equal(t, 2, result.Shipment.EventCount)
	assert.NotNil(t, result.Shipment.LastSyncedAt)

	var events []shippingdomain.TrackingEvent
	require.NoError(t, db.Where("order_id = ?", orderRecord.ID).Order("event_time DESC").Find(&events).Error)
	require.Len(t, events, 2)
	assert.Equal(t, orderRecord.ID, events[0].OrderID)
	assert.Equal(t, "MOCK123456", events[0].TrackingNumber)
	assert.Equal(t, "DHL", events[0].ProviderCarrierCode)
}

func TestOrderServiceGenerateOrderNumberFormat(t *testing.T) {
	generator, err := ordernumber.NewGenerator("test-order-number-secret", 0)
	require.NoError(t, err)
	service := &OrderService{numberGenerator: generator}
	orderNumber, err := service.generateOrderNumber()

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(orderNumber, "TZ-"+time.Now().UTC().Format("2006")+"-"))
	assert.True(t, generator.Validate(orderNumber))
}

func TestOrderServiceRejectsSequentialPublicOrderNumber(t *testing.T) {
	generator, err := ordernumber.NewGenerator("test-order-number-secret", 0)
	require.NoError(t, err)
	service := &OrderService{numberGenerator: generator}

	assert.False(t, service.validatesKnownProtectedOrderNumber("1001"))
	assert.False(t, service.validatesKnownProtectedOrderNumber("#1001"))
	assert.True(t, service.validatesKnownProtectedOrderNumber("ORD-LEGACY-1001"))
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
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&order.Order{},
		&order.OrderItem{},
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&loyalty.UserLoyalty{},
		&loyalty.LoyaltyTransaction{},
		&loyalty.ProgramConfig{},
		&loyalty.ProgramRedeemOption{},
		&loyalty.MemberLevel{},
		&setting.Setting{},
		&paymentdomain.Transaction{},
		&paymentdomain.TaxRate{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.TrackingProviderConfig{},
		&shippingdomain.TrackingCarrierMapping{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.ShippingRule{},
		&shippingdomain.PackagingRule{},
		&shippingdomain.PackagingRuleApply{},
	))

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	shippingService := NewShippingService(shippingRepo, productRepo)
	seedDefaultShippingTemplate(t, db)
	checkoutService := NewCheckoutService(productRepo, couponRepo, paymentRepo, loyaltyRepo, shippingService)
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo, shippingRepo)
	currencyPolicyService := seedTestCurrencyPolicy(t, db)
	programRepo := repository.NewLoyaltyProgramRepository(db)
	txManager.ConfigureLoyaltyProgramRepository(programRepo)
	programService := NewLoyaltyProgramService(programRepo)
	programService.ConfigureCurrencyPolicy(currencyPolicyService)
	_, err = programService.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ExchangeRatePoints:        100,
		MinRedeemPoints:           1000,
		MaxValuePerDayCents:       50000,
		CardExpiryDays:            365,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
		RedeemValuesCents:         []int64{1000, 5000, 10000},
	})
	require.NoError(t, err)
	checkoutService.ConfigureLoyaltyProgram(programService)
	numberGenerator, err := ordernumber.NewGenerator("test-order-number-secret", 0)
	require.NoError(t, err)

	return db, NewOrderService(txManager, orderRepo, checkoutService, shippingService, currencyPolicyService, numberGenerator)
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
		Weight:       9000,
		IsDefault:    true,
		IsActive:     true,
	}).Error)
	return record
}

func seedProductShell(t *testing.T, db *gorm.DB, price float64, stock int) product.Product {
	t.Helper()

	shippingTemplateID := seedOrderTestShippingTemplateID(t, db)
	record := product.Product{
		ShippingTemplateID: shippingTemplateID,
		SKU:                "SKU-TEST",
		Name:               "Test Product",
		Slug:               "test-product",
		Price:              price,
		Stock:              stock,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedOrderTestShippingTemplateID(t *testing.T, db *gorm.DB) *uint {
	t.Helper()

	var template shippingdomain.ShippingTemplate
	require.NoError(t, db.First(&template).Error)
	return &template.ID
}

func seedDefaultShippingTemplate(t *testing.T, db *gorm.DB) {
	t.Helper()

	template := shippingdomain.ShippingTemplate{
		Name:          "Test standard shipping",
		Type:          "weight",
		FreeShipping:  true,
		FreeThreshold: 100,
		DefaultFee:    10,
		Enabled:       true,
		Rules: []shippingdomain.ShippingRule{
			{
				Region:   "US",
				MinValue: 0,
				MaxValue: 0,
				Fee:      10,
			},
		},
	}
	require.NoError(t, db.Create(&template).Error)
}

func seedUserLoyalty(t *testing.T, db *gorm.DB, userID uint, points int) {
	t.Helper()

	require.NoError(t, db.FirstOrCreate(&loyalty.MemberLevel{}, loyalty.MemberLevel{
		Name:         "Test Level",
		MinPoints:    0,
		MaxPoints:    999999,
		DiscountRate: 5,
	}).Error)

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

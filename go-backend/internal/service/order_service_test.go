package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	pricingdomain "commerce-platform/internal/domain/pricing"
	"commerce-platform/internal/domain/product"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	refundcancellationdomain "commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/domain/setting"
	shippingdomain "commerce-platform/internal/domain/shipping"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/ordernumber"
	"commerce-platform/internal/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderServiceCreateOrderRejectsChangedExpectedTotalMinorBeforeStockDeduction(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 100, 5)
	expectedTotalMinor := int64(10006)

	createdOrder, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{ExpectedTotalMinor: &expectedTotalMinor},
	)

	require.ErrorIs(t, err, ErrOrderTotalChanged)
	assert.Nil(t, createdOrder)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Zero(t, orderCount)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 5, savedVariant.Stock)
}

func TestOrderServiceCreateOrderRejectsExpectedTotalMinorDifference(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 100, 5)
	expectedTotalMinor := int64(10005)

	createdOrder, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{ExpectedTotalMinor: &expectedTotalMinor},
	)

	require.ErrorIs(t, err, ErrOrderTotalChanged)
	assert.Nil(t, createdOrder)
}

func TestOrderServiceCreateOrderUsesRequestedCarrierServiceForExpectedTotalMinor(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 100, 5)
	require.NotNil(t, productRecord.ShippingTemplateID)
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).
		Where("id = ?", *productRecord.ShippingTemplateID).
		Update("free_shipping", false).Error)

	carrier := shippingdomain.Carrier{Name: "DHL", Code: "DHL", Enabled: true}
	require.NoError(t, db.Create(&carrier).Error)
	standardService := shippingdomain.CarrierService{
		CarrierID: carrier.ID, TemplateID: productRecord.ShippingTemplateID,
		ServiceCode: "DHL-STD", ServiceName: "DHL Standard", Countries: `["US"]`,
		Currency: "USD", BillingMode: "actual_weight", Enabled: true,
	}
	expressService := shippingdomain.CarrierService{
		CarrierID: carrier.ID, TemplateID: productRecord.ShippingTemplateID,
		ServiceCode: "DHL-EXP", ServiceName: "DHL Express", Countries: `["US"]`,
		Currency: "USD", BillingMode: "actual_weight", RemoteSurchargeMinor: 10000, Enabled: true,
	}
	require.NoError(t, db.Create(&standardService).Error)
	require.NoError(t, db.Create(&expressService).Error)
	checkoutQuote, err := orderService.checkout.Quote(CheckoutQuoteInput{
		UserID:          42,
		Items:           []order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		ShippingAddress: testAddress(),
	})
	require.NoError(t, err)
	require.NotNil(t, checkoutQuote.ShippingQuote)
	expressPlanID := ""
	for _, plan := range checkoutQuote.ShippingQuote.Plans {
		if len(plan.Legs) == 1 && plan.Legs[0].CarrierServiceID == expressService.ID {
			expressPlanID = plan.ID
			break
		}
	}
	require.NotEmpty(t, expressPlanID)
	expectedTotalMinor := int64(21000)

	createdOrder, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{
			ExpectedTotalMinor:  &expectedTotalMinor,
			ShippingQuoteID:     checkoutQuote.ShippingQuote.ID,
			SelectedQuotePlanID: expressPlanID,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, int64(11000), createdOrder.ShippingFeeMinor)
	assert.Equal(t, expectedTotalMinor, createdOrder.TotalAmountMinor)
	assert.Contains(t, createdOrder.ShippingMethod, "DHL Express")
	assert.Equal(t, checkoutQuote.ShippingQuote.ID, createdOrder.ShippingQuoteID)
	assert.Equal(t, expressPlanID, createdOrder.ShippingQuotePlanID)
	assert.NotEmpty(t, createdOrder.ShippingPlanSnapshotData)
}

func TestOrderServiceRejectsUnavailableShippingRateBeforeStockDeduction(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 1200, 5)
	require.NotNil(t, productRecord.ShippingTemplateID)
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).
		Where("id = ?", *productRecord.ShippingTemplateID).
		Updates(map[string]interface{}{"free_shipping": false, "default_fee_minor": 0}).Error)
	shippingAddress := testAddress()
	shippingAddress.Country = "BR"

	createdOrder, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		shippingAddress,
		shippingAddress,
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{},
	)

	assert.Nil(t, createdOrder)
	require.ErrorIs(t, err, ErrCountryNotSupported)
	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Zero(t, orderCount)
	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 5, savedVariant.Stock)
}

func TestOrderServiceCreateOrderConsumesCheckoutCartAndStoresConsumedCartReference(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 100, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 100)

	createdOrder, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, VariantID: &variant.ID, Quantity: 99}},
		testAddress(),
		testAddress(),
		"paypal",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{CheckoutCartID: cart.ID},
	)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	require.NotNil(t, createdOrder.CheckoutCartID)
	assert.Equal(t, cart.ID, *createdOrder.CheckoutCartID)

	var remainingCartItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&remainingCartItems).Error)
	assert.Empty(t, remainingCartItems)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 4, savedVariant.Stock)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.NotNil(t, savedOrder.CheckoutCartID)
	assert.Equal(t, cart.ID, *savedOrder.CheckoutCartID)
	require.Len(t, savedOrder.Items, 1)
	assert.Equal(t, 1, savedOrder.Items[0].Quantity)
}

func TestOrderServiceCreateOrderRejectsSecondCheckoutAgainstConsumedCart(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 100, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 100)

	_, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		userID,
		nil,
		testAddress(),
		testAddress(),
		"paypal",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{CheckoutCartID: cart.ID},
	)
	require.NoError(t, err)

	_, err = orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		userID,
		nil,
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{CheckoutCartID: cart.ID},
	)
	require.ErrorIs(t, err, ErrCheckoutCartAlreadyConsumed)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Equal(t, int64(1), orderCount)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 4, savedVariant.Stock)
}

func TestOrderServiceConcurrentCheckoutAgainstSameCartCreatesOnlyOneOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 100, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 100)

	var waitGroup sync.WaitGroup
	errs := make(chan error, 2)
	waitGroup.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer waitGroup.Done()
			_, err := orderService.CreateOrderWithAttributionAndOptions(
				context.Background(),
				userID,
				nil,
				testAddress(),
				testAddress(),
				"paypal",
				"standard",
				"",
				0,
				attributionpkg.Context{},
				OrderCreationOptions{CheckoutCartID: cart.ID},
			)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)

	var successCount int
	var consumedCartErrorCount int
	for err := range errs {
		if err == nil {
			successCount++
			continue
		}
		if errors.Is(err, ErrCheckoutCartAlreadyConsumed) {
			consumedCartErrorCount++
			continue
		}
		t.Fatalf("unexpected concurrent checkout error: %v", err)
	}
	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, consumedCartErrorCount)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Equal(t, int64(1), orderCount)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 4, savedVariant.Stock)
}

func TestOrderServiceCheckoutTransactionFailureRestoresCartConsumption(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 100, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 100)
	require.NoError(t, db.Create(&setting.Setting{
		Key:      refundcancellationdomain.Key,
		Locale:   "en",
		Value:    "{invalid-json",
		Type:     "json",
		IsPublic: true,
	}).Error)
	orderService.ConfigureRefundCancellationPolicy(
		NewRefundCancellationPolicyService(repository.NewSettingRepository(db)),
	)

	_, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		userID,
		nil,
		testAddress(),
		testAddress(),
		"paypal",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{CheckoutCartID: cart.ID},
	)
	require.ErrorIs(t, err, ErrOrderPolicyDisclosureFailure)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Zero(t, orderCount)

	var cartItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&cartItems).Error)
	require.Len(t, cartItems, 1)
	assert.Equal(t, 1, cartItems[0].Quantity)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 5, savedVariant.Stock)
}

func TestOrderServiceCancelUnpaidCheckoutOrderRestoresCartExactlyOnce(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 100, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 100)

	createdOrder, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		userID,
		nil,
		testAddress(),
		testAddress(),
		"paypal",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		OrderCreationOptions{CheckoutCartID: cart.ID},
	)
	require.NoError(t, err)

	require.NoError(t, orderService.CancelOrder(createdOrder.ID, userID))

	var restoredCartItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&restoredCartItems).Error)
	require.Len(t, restoredCartItems, 1)
	assert.Equal(t, 1, restoredCartItems[0].Quantity)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 5, savedVariant.Stock)

	require.ErrorIs(t, orderService.CancelOrder(createdOrder.ID, userID), ErrOrderCancellationConflict)
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&restoredCartItems).Error)
	require.Len(t, restoredCartItems, 1)
	assert.Equal(t, 1, restoredCartItems[0].Quantity)
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 5, savedVariant.Stock)
}

func TestOrderServiceCreateOrderPersistsPricingAndAdjustments(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	productRecord.HSCode = "871499"
	productRecord.CNCode = "87149990"
	productRecord.CountryOfOrigin = "CN"
	productRecord.CustomsDescription = "Bicycle parts"
	require.NoError(t, db.Save(&productRecord).Error)
	seedUserLoyalty(t, db, userID, 1000)
	seedCoupon(t, db, "SAVE10", "fixed", 10, 1)

	incorrectDeclaredValueMinor := int64(99900)
	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{
			ProductID:              productRecord.ID,
			Quantity:               2,
			DeclaredValueMinor:     &incorrectDeclaredValueMinor,
			DeclaredValueConfirmed: true,
		}},
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
	assert.Equal(t, int64(10000), createdOrder.SubtotalAmountMinor)
	assert.Equal(t, int64(1600), createdOrder.DiscountAmountMinor)
	assert.Equal(t, int64(8400), createdOrder.TotalAmountMinor)
	assert.Equal(t, 100, createdOrder.PointsUsed)
	assert.Equal(t, int64(100), createdOrder.PointsValueMinor)
	assert.Equal(t, "SAVE10", createdOrder.CouponCode)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 1)
	require.NotNil(t, savedOrder.Items[0].VariantID)
	assert.Equal(t, productRecord.Name, savedOrder.Items[0].ProductName)
	assert.Equal(t, productRecord.SKU, savedOrder.Items[0].SKU)
	assert.Equal(t, int64(10000), savedOrder.Items[0].SubtotalMinor)
	assert.Equal(t, int64(8400), savedOrder.Items[0].TotalMinor)
	assert.Equal(t, "871499", savedOrder.Items[0].HSCode)
	assert.Equal(t, "87149990", savedOrder.Items[0].CNCode)
	assert.Equal(t, "CN", savedOrder.Items[0].CountryOfOrigin)
	assert.Equal(t, "Bicycle parts", savedOrder.Items[0].CustomsDescription)
	assert.False(t, savedOrder.Items[0].DeclaredValueConfirmed)

	require.NoError(t, db.Model(&product.Product{}).
		Where("id = ?", productRecord.ID).
		Updates(map[string]interface{}{
			"hs_code":             "999999",
			"customs_description": "Changed product data",
		}).Error)
	var unchangedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&unchangedOrder, createdOrder.ID).Error)
	assert.Equal(t, "871499", unchangedOrder.Items[0].HSCode)
	assert.Equal(t, "Bicycle parts", unchangedOrder.Items[0].CustomsDescription)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 3, savedVariant.Stock)

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
	assert.Equal(t, int64(1000), usage.DiscountMinor)
}

func TestOrderServiceCreateOrderSettlesZeroTotalDiscountOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 100, 5)
	seedCoupon(t, db, "FREE100", "fixed", 100, 1)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"FREE100",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	require.NotZero(t, createdOrder.ID)
	assert.Equal(t, int64(10000), createdOrder.SubtotalAmountMinor)
	assert.Equal(t, int64(10000), createdOrder.DiscountAmountMinor)
	assert.Equal(t, int64(0), createdOrder.TotalAmountMinor)
	assert.Equal(t, "paid", createdOrder.PaymentStatus)
	assert.Equal(t, "processing", createdOrder.Status)
	require.NotNil(t, createdOrder.PaidAt)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)
	assert.Equal(t, "processing", savedOrder.Status)
	require.NotNil(t, savedOrder.PaidAt)

	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("order_id = ? AND transaction_id = ?", createdOrder.ID, zeroTotalOrderTransactionID(createdOrder.ID)).First(&transaction).Error)
	assert.Equal(t, zeroTotalSettlementPaymentMethod, transaction.PaymentMethod)
	assert.Equal(t, "completed", transaction.Status)
	assert.Equal(t, int64(0), transaction.AmountMinor)
	require.NotNil(t, transaction.CompletedAt)

	var event outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", fmt.Sprintf("%s:%d:%s", outboxdomain.EventTypeOrderPaid, createdOrder.ID, transaction.TransactionID)).First(&event).Error)
	assert.Equal(t, outboxdomain.EventTypeOrderPaid, event.EventType)

	var conversionCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_type = ? AND aggregate_id = ?", outboxdomain.EventTypeVerifiedConversion, fmt.Sprint(createdOrder.ID)).Count(&conversionCount).Error)
	assert.Zero(t, conversionCount)
}

func TestOrderServiceCreateOrderSnapshotsMadeToOrderFulfillment(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 120, 5)
	productRecord.FulfillmentMode = product.FulfillmentModeMadeToOrder
	require.NoError(t, db.Save(&productRecord).Error)

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
	assert.Equal(t, order.FulfillmentModeMadeToOrder, createdOrder.FulfillmentMode)
	assert.Equal(t, order.ProductionStatusNotStarted, createdOrder.ProductionStatus)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 1)
	assert.Equal(t, order.FulfillmentModeMadeToOrder, savedOrder.Items[0].FulfillmentMode)
}

func TestOrderServiceCreateOrderRejectsCouponPerUserUsageLimit(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "WELCOME20",
		Type:              "fixed",
		ValueMinor:        2000,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID:      cp.ID,
		UserID:        userID,
		OrderID:       1001,
		DiscountMinor: 2000,
	}).Error)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"WELCOME20",
		0,
	)

	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)
	assert.Nil(t, createdOrder)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Equal(t, int64(0), orderCount)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.First(&savedCoupon, cp.ID).Error)
	assert.Equal(t, 0, savedCoupon.UsedCount)
}

func TestOrderServiceCreateOrderEnforcesGuestCouponUsageLimitByNormalizedEmail(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 50, 5)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "GUESTWELCOME",
		Type:              "fixed",
		ValueMinor:        2000,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)

	firstAddress := testAddress()
	firstAddress.Email = " Buyer@Example.com "
	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		0,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		firstAddress,
		firstAddress,
		"card",
		"standard",
		"GUESTWELCOME",
		0,
	)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	var usage coupon.CouponUsage
	require.NoError(t, db.Where("coupon_id = ? AND order_id = ?", cp.ID, createdOrder.ID).First(&usage).Error)
	assert.Equal(t, "buyer@example.com", usage.Email)

	secondAddress := testAddress()
	secondAddress.Email = "buyer@example.com"
	_, err = orderService.CreateOrder(
		context.Background(),
		0,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		secondAddress,
		secondAddress,
		"card",
		"standard",
		"GUESTWELCOME",
		0,
	)
	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)
}

func TestOrderServiceUpdateOrderItemCustoms(t *testing.T) {
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

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 1)
	orderItemID := savedOrder.Items[0].ID
	originalPricingSnapshot := append([]byte(nil), savedOrder.Items[0].PricingSnapshotData...)
	require.NotEqual(t, "{}", string(originalPricingSnapshot))
	require.NotEqual(t, "{}", string(savedOrder.PricingSnapshotData))
	pricingPayload, err := pricingdomain.ParseOrderPricingSnapshot(savedOrder.PricingSnapshotData)
	require.NoError(t, err)
	assert.Equal(t, int64(6000), pricingPayload.TotalMinor)

	declaredValue := int64(4275)
	require.NoError(t, orderService.UpdateOrderItemCustoms(createdOrder.ID, orderItemID, &declaredValue, true))

	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.NotNil(t, savedOrder.Items[0].DeclaredValueMinor)
	assert.Equal(t, declaredValue, *savedOrder.Items[0].DeclaredValueMinor)
	assert.True(t, savedOrder.Items[0].DeclaredValueConfirmed)
	assert.Equal(t, originalPricingSnapshot, []byte(savedOrder.Items[0].PricingSnapshotData))

	require.NoError(t, orderService.UpdateOrderItemCustoms(createdOrder.ID, orderItemID, nil, false))
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	assert.Nil(t, savedOrder.Items[0].DeclaredValueMinor)
	assert.False(t, savedOrder.Items[0].DeclaredValueConfirmed)

	require.ErrorIs(t, orderService.UpdateOrderItemCustoms(createdOrder.ID, orderItemID, nil, true), ErrDeclaredValueConfirmationRequired)
	negativeValue := int64(-1)
	require.ErrorIs(t, orderService.UpdateOrderItemCustoms(createdOrder.ID, orderItemID, &negativeValue, false), ErrDeclaredValueInvalid)
	require.ErrorIs(t, orderService.UpdateOrderItemCustoms(createdOrder.ID, orderItemID+1000, &declaredValue, true), ErrOrderItemNotFound)
}

func TestOrderServiceUpdateOrderItemCustomsRejectsPostShipmentChanges(t *testing.T) {
	testCases := []struct {
		name           string
		status         string
		shippingStatus string
		setShippedAt   bool
	}{
		{name: "shipped status", status: "shipped", shippingStatus: "pending"},
		{name: "completed status", status: "completed", shippingStatus: "shipped"},
		{name: "shipped shipping status", status: "processing", shippingStatus: "shipped"},
		{name: "shipped timestamp", status: "processing", shippingStatus: "pending", setShippedAt: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
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

			var savedOrder order.Order
			require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
			require.Len(t, savedOrder.Items, 1)
			orderItemID := savedOrder.Items[0].ID
			require.NoError(t, db.Model(&order.OrderItem{}).
				Where("id = ?", orderItemID).
				Updates(map[string]interface{}{
					"declared_value_minor":     1750,
					"declared_value_confirmed": true,
				}).Error)

			var shippedAt *time.Time
			if testCase.setShippedAt {
				value := time.Now().UTC()
				shippedAt = &value
			}
			require.NoError(t, db.Model(&order.Order{}).
				Where("id = ?", createdOrder.ID).
				Updates(map[string]interface{}{
					"status":          testCase.status,
					"shipping_status": testCase.shippingStatus,
					"shipped_at":      shippedAt,
				}).Error)

			updatedValue := int64(4275)
			require.ErrorIs(t, orderService.UpdateOrderItemCustoms(createdOrder.ID, orderItemID, &updatedValue, true), ErrOrderCustomsUpdateLocked)

			require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
			require.NotNil(t, savedOrder.Items[0].DeclaredValueMinor)
			assert.Equal(t, int64(1750), *savedOrder.Items[0].DeclaredValueMinor)
			assert.True(t, savedOrder.Items[0].DeclaredValueConfirmed)
		})
	}
}

func TestOrderServiceCreateOrderUsesVersionedLoyaltyExchangeRate(t *testing.T) {
	db, orderService := newTestOrderService(t)
	programService := newTestLoyaltyProgramService(t, db)
	config, err := programService.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		ExchangeRatePoints:        80,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
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
	assert.Equal(t, int64(1000), createdOrder.PointsValueMinor)

	var pointTransaction loyalty.LoyaltyTransaction
	require.NoError(t, db.Where("user_id = ? AND source = ? AND source_id = ?", userID, "order", createdOrder.ID).First(&pointTransaction).Error)
	require.NotNil(t, pointTransaction.ProgramConfigID)
	assert.Equal(t, config.ID, *pointTransaction.ProgramConfigID)
}

func TestOrderServiceCreateOrderUsesVariantPricingAndStock(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProductShell(t, db, 999, 99)
	cacheInvalidator := &recordingProductCacheInvalidator{}
	orderService.ConfigureProductCacheInvalidator(cacheInvalidator)
	variant := product.ProductVariant{
		ProductID:      productRecord.ID,
		SKU:            "SKU-TEST-BLK-24H",
		Title:          "Black / 24H",
		OptionValues:   `{"color":"black","spoke_holes":"24"}`,
		PriceMinor:     9000,
		SalePriceMinor: func() *int64 { v := int64(8000); return &v }(),
		Stock:          3,
		Weight:         11000,
		IsDefault:      true,
		IsActive:       true,
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
	assert.Equal(t, int64(16000), createdOrder.SubtotalAmountMinor)
	assert.Equal(t, int64(16000), createdOrder.TotalAmountMinor)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 1)
	require.NotNil(t, savedOrder.Items[0].VariantID)
	assert.Equal(t, variant.ID, *savedOrder.Items[0].VariantID)
	assert.Equal(t, variant.SKU, savedOrder.Items[0].SKU)
	assert.NotEmpty(t, savedOrder.Items[0].ConfigurationSnapshotData)
	assert.Equal(t, int64(8000), savedOrder.Items[0].PriceMinor)
	assert.Equal(t, int64(16000), savedOrder.Items[0].SubtotalMinor)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 1, savedVariant.Stock)

	assert.Equal(t, []uint{productRecord.ID}, cacheInvalidator.productIDs)

	var cacheEvent outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", outboxdomain.EventTypeProductCacheInvalidate).First(&cacheEvent).Error)
	assert.Equal(t, outboxdomain.EventStatusPending, cacheEvent.Status)
	assert.Contains(t, cacheEvent.EventKey, "order_stock_deducted")
}

func TestOrderServiceCreateOrderInvalidatesWarmedProductDetailCache(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	productService := newRedisBackedProductService(t, db)
	orderService.ConfigureProductCacheInvalidator(productService)

	warmed, err := productService.GetPublicByID(productRecord.ID)
	require.NoError(t, err)
	require.Equal(t, 5, warmed.TotalVariantStock())

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
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	reloaded, err := productService.GetPublicByID(productRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, reloaded.TotalVariantStock())
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
	assert.Equal(t, "DHL / US Express / Express (EXP-US)", createdOrder.ShippingMethod)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "DHL / US Express / Express (EXP-US)", savedOrder.ShippingMethod)
}

func TestOrderServiceCreateOrderRollsBackWhenStockIsInsufficient(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 1)
	cacheInvalidator := &recordingProductCacheInvalidator{}
	orderService.ConfigureProductCacheInvalidator(cacheInvalidator)

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

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 1, savedVariant.Stock)
	assert.Empty(t, cacheInvalidator.productIDs)
}

func TestOrderServiceCreateOrderPersistsImmutableEvidenceSnapshot(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 750, 5)
	rule := productrequirement.ProductQualityRequirementRule{
		ProductID:              productRecord.ID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: true,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "product-v1",
		Reason:                 "configured assembly product",
	}
	require.NoError(t, db.Create(&rule).Error)

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

	snapshot, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(createdOrder.ID)
	require.NoError(t, err)
	assert.True(t, snapshot.IsHighValue)
	assert.True(t, snapshot.HasSpokeTensionQC)
	assert.Equal(t, int64(75000), snapshot.OrderTotalUSDMinor)
	assert.Equal(t, orderevidence.OrderEvidenceSnapshotSchemaVersion, snapshot.SchemaVersion)
	require.NoError(t, snapshot.VerifyIntegrity())

	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg, err := evidenceRepo.FindLatestPackageByOrderID(createdOrder.ID)
	require.NoError(t, err)
	assert.Equal(t, snapshot.ID, pkg.SnapshotID)
	assert.True(t, pkg.IsHighValue)
	assert.True(t, pkg.HasSpokeTensionQC)
	assert.Equal(t, orderevidence.PackageStatusIncomplete, pkg.Status)

	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	require.Len(t, items, 5)
	var configurationCount, identityCount, tensionCount, outboundCount, podCount int
	for _, item := range items {
		switch item.ItemType {
		case orderevidence.EvidenceItemTypeConfigurationConfirmation:
			configurationCount++
			assert.Equal(t, orderevidence.EvidenceItemStatusComplete, item.Status)
			assert.Equal(t, snapshot.ID, *item.SnapshotID)
		case orderevidence.EvidenceItemTypeProductIdentity:
			identityCount++
		case orderevidence.EvidenceItemTypeSpokeQCTension:
			tensionCount++
		case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
			outboundCount++
		case orderevidence.EvidenceItemTypeSignedPOD:
			podCount++
		}
	}
	assert.Equal(t, 1, configurationCount)
	assert.Equal(t, 1, identityCount)
	assert.Equal(t, 1, tensionCount)
	assert.Equal(t, 1, outboundCount)
	assert.Equal(t, 1, podCount)

	require.NoError(t, db.Model(&rule).Updates(map[string]interface{}{
		"spoke_tension_qc_required": false,
		"status":                    productrequirement.RuleStatusInactive,
		"rule_version":              "product-v2",
	}).Error)
	foundAgain, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(createdOrder.ID)
	require.NoError(t, err)
	assert.True(t, foundAgain.HasSpokeTensionQC)
}

func TestOrderServiceCreateOrderUsesVariantRequirementOverrideAndFreezesSnapshot(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 100, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)

	productDefaultRule := productrequirement.ProductQualityRequirementRule{
		ProductID:              productRecord.ID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: false,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "product-v1",
		Reason:                 "product default does not require QC",
	}
	require.NoError(t, db.Create(&productDefaultRule).Error)

	variantRule := productrequirement.ProductQualityRequirementRule{
		ProductID:              productRecord.ID,
		VariantID:              &variant.ID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: true,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "variant-v1",
		Reason:                 "variant requires QC",
	}
	require.NoError(t, db.Create(&variantRule).Error)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, VariantID: &variant.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	snapshot, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(createdOrder.ID)
	require.NoError(t, err)
	require.True(t, snapshot.HasSpokeTensionQC)

	payload, err := orderevidence.ParseOrderEvidenceSnapshotPayload(snapshot)
	require.NoError(t, err)
	require.Len(t, payload.Items, 1)
	require.NotNil(t, payload.Items[0].ProductRequirementSnapshot.RuleID)
	assert.Equal(t, variantRule.ID, *payload.Items[0].ProductRequirementSnapshot.RuleID)
	assert.Equal(t, "variant-v1", payload.Items[0].ProductRequirementSnapshot.RuleVersion)
	assert.Equal(t, productrequirement.ResolutionSourceVariant, payload.Items[0].ProductRequirementSnapshot.Source)

	var storedRequirement productrequirement.ProductQualityRequirementRule
	require.NoError(t, db.First(&storedRequirement, variantRule.ID).Error)
	require.NoError(t, db.Model(&storedRequirement).Updates(map[string]interface{}{
		"spoke_tension_qc_required": false,
		"status":                    productrequirement.RuleStatusInactive,
		"rule_version":              "variant-v2",
	}).Error)

	foundAgain, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(createdOrder.ID)
	require.NoError(t, err)
	assert.True(t, foundAgain.HasSpokeTensionQC)

	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg, err := evidenceRepo.FindLatestPackageByOrderID(createdOrder.ID)
	require.NoError(t, err)
	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	var tensionItems int
	for _, item := range items {
		if item.ItemType == orderevidence.EvidenceItemTypeSpokeQCTension {
			tensionItems++
		}
	}
	assert.Equal(t, 1, tensionItems)
}

func TestOrderServiceEvidenceSeparatesHighValueFromSpokeTensionRequirement(t *testing.T) {
	db, orderService := newTestOrderService(t)
	ordinaryProduct := seedProductWithSKU(t, db, 750, 5, "SKU-EVIDENCE-ORDINARY")
	qcProduct := seedProductWithSKU(t, db, 100, 5, "SKU-EVIDENCE-QC")
	require.NoError(t, db.Create(&productrequirement.ProductQualityRequirementRule{
		ProductID:              qcProduct.ID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: true,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "qc-v1",
		Reason:                 "explicit configured requirement",
	}).Error)

	highValueOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: ordinaryProduct.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)
	require.NoError(t, err)
	highValueSnapshot, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(highValueOrder.ID)
	require.NoError(t, err)
	assert.True(t, highValueSnapshot.IsHighValue)
	assert.False(t, highValueSnapshot.HasSpokeTensionQC)
	highValuePackage, err := repository.NewOrderEvidenceRepository(db).FindLatestPackageByOrderID(highValueOrder.ID)
	require.NoError(t, err)
	highValueItems, err := repository.NewOrderEvidenceRepository(db).ListItemsByPackageID(highValuePackage.ID)
	require.NoError(t, err)
	for _, item := range highValueItems {
		assert.NotEqual(t, orderevidence.EvidenceItemTypeSpokeQCTension, item.ItemType)
	}

	lowValueOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: qcProduct.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)
	require.NoError(t, err)
	lowValueSnapshot, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(lowValueOrder.ID)
	require.NoError(t, err)
	assert.False(t, lowValueSnapshot.IsHighValue)
	assert.True(t, lowValueSnapshot.HasSpokeTensionQC)
	lowValuePackage, err := repository.NewOrderEvidenceRepository(db).FindLatestPackageByOrderID(lowValueOrder.ID)
	require.NoError(t, err)
	lowValueItems, err := repository.NewOrderEvidenceRepository(db).ListItemsByPackageID(lowValuePackage.ID)
	require.NoError(t, err)
	var lowValueTensionItems int
	for _, item := range lowValueItems {
		if item.ItemType == orderevidence.EvidenceItemTypeSpokeQCTension {
			lowValueTensionItems++
		}
	}
	assert.Equal(t, 1, lowValueTensionItems)
}

func TestOrderServiceCreateOrderRollsBackWhenEvidenceRuleIsAmbiguous(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 100, 1)
	for _, version := range []string{"product-v1", "product-v2"} {
		require.NoError(t, db.Create(&productrequirement.ProductQualityRequirementRule{
			ProductID:              productRecord.ID,
			RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
			SpokeTensionQCRequired: true,
			Status:                 productrequirement.RuleStatusActive,
			RuleVersion:            version,
		}).Error)
	}

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
	require.Error(t, err)
	assert.Nil(t, createdOrder)
	assert.Contains(t, err.Error(), "multiple active default spoke tension QC rules")

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Zero(t, orderCount)
	var snapshotCount int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceSnapshot{}).Count(&snapshotCount).Error)
	assert.Zero(t, snapshotCount)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 1, savedVariant.Stock)
}

func TestOrderServiceCreateOrderRequiresCompleteEvidenceConfiguration(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 100, 1)
	orderService.ConfigureOrderEvidence(nil)

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
	require.ErrorIs(t, err, ErrOrderEvidenceNotConfigured)
	assert.Nil(t, createdOrder)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.Zero(t, orderCount)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 1, savedVariant.Stock)
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
	cacheInvalidator := &recordingProductCacheInvalidator{}
	orderService.ConfigureProductCacheInvalidator(cacheInvalidator)
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
		AmountMinor:   createdOrder.TotalAmountMinor,
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

	var usage coupon.CouponUsage
	require.NoError(t, db.Where("order_id = ?", createdOrder.ID).First(&usage).Error)
	assert.Equal(t, coupon.CouponUsageStatusReversed, usage.Status)
	assert.Equal(t, "payment expired", usage.ReversalReason)
	assert.NotNil(t, usage.ReversedAt)
	assert.Equal(t, []uint{productRecord.ID, productRecord.ID}, cacheInvalidator.productIDs)
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
		AmountMinor:   createdOrder.TotalAmountMinor,
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

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 4, savedVariant.Stock)
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
		Name:                "Gold",
		MinPoints:           0,
		MaxPoints:           999999,
		DiscountRateDecimal: "0",
		Benefits:            "[]",
		SortOrder:           10,
	}).Error)

	orderRecord := order.Order{
		OrderNumber:         "ORD-COMPLETE-POINTS",
		UserID:              userID,
		Status:              "shipped",
		PaymentStatus:       "paid",
		SubtotalAmountMinor: 12000,
		DiscountAmountMinor: 2000,
		TotalAmountMinor:    11500,
		Currency:            "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.NoError(t, orderService.UpdateOrderStatus(orderRecord.ID, "completed"))
	var pendingRewardCount int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ? AND source_id = ?", userID, "earn", "order", orderRecord.ID).
		Count(&pendingRewardCount).Error)
	assert.Zero(t, pendingRewardCount)
	processOrderCompletedOutboxEvent(t, db, orderService, orderRecord.ID)
	processOrderCompletedOutboxEvent(t, db, orderService, orderRecord.ID)

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

func TestOrderServiceCompletingOrderDoesNotChangeShippingStatus(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:         "ORD-COMPLETE-SHIPPING-INDEPENDENT",
		UserID:              42,
		Status:              "shipped",
		ShippingStatus:      "shipped",
		PaymentStatus:       "paid",
		SubtotalAmountMinor: 10000,
		TotalAmountMinor:    10000,
		Currency:            "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.NoError(t, orderService.UpdateOrderStatus(orderRecord.ID, "completed"))
	processOrderCompletedOutboxEvent(t, db, orderService, orderRecord.ID)

	var saved order.Order
	require.NoError(t, db.First(&saved, orderRecord.ID).Error)
	assert.Equal(t, "completed", saved.Status)
	assert.Equal(t, "shipped", saved.ShippingStatus)
	assert.NotNil(t, saved.CompletedAt)
}

func TestOrderServiceCompletionConvertsNonUSDPointsUsingHistoricalFXSnapshot(t *testing.T) {
	db, orderService := newTestOrderService(t)
	testCases := []struct {
		currency       string
		rateDecimal    string
		subtotal       float64
		discount       float64
		expectedPoints int
	}{
		{currency: "EUR", rateDecimal: "0.9", subtotal: 120, discount: 20, expectedPoints: 111},
		{currency: "GBP", rateDecimal: "0.8", subtotal: 120, discount: 20, expectedPoints: 125},
		{currency: "JPY", rateDecimal: "150", subtotal: 15000, discount: 0, expectedPoints: 100},
		{currency: "CNY", rateDecimal: "7", subtotal: 120, discount: 20, expectedPoints: 14},
	}

	for index, testCase := range testCases {
		t.Run(testCase.currency, func(t *testing.T) {
			userID := uint(43 + index)
			multiplier := 100.0
			if testCase.currency == "JPY" {
				multiplier = 1
			}
			orderRecord := order.Order{
				OrderNumber:         "ORD-COMPLETE-" + testCase.currency + "-POINTS",
				UserID:              userID,
				Status:              "shipped",
				PaymentStatus:       "paid",
				SubtotalAmountMinor: int64(testCase.subtotal * multiplier),
				DiscountAmountMinor: int64(testCase.discount * multiplier),
				TotalAmountMinor:    int64((testCase.subtotal - testCase.discount) * multiplier),
				Currency:            testCase.currency,
				FXSnapshotData: currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
					Version:       currency.OrderFXSnapshotVersion,
					BaseCurrency:  "USD",
					OrderCurrency: testCase.currency,
					RateDecimal:   testCase.rateDecimal,
					Source:        "test-rate",
					CapturedAt:    time.Now().UTC(),
				}),
			}
			require.NoError(t, db.Create(&orderRecord).Error)

			require.NoError(t, orderService.UpdateOrderStatus(orderRecord.ID, "completed"))
			processOrderCompletedOutboxEvent(t, db, orderService, orderRecord.ID)

			var savedOrder order.Order
			require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
			assert.Equal(t, "completed", savedOrder.Status)

			var earnedTransactions []loyalty.LoyaltyTransaction
			require.NoError(t, db.Where(
				"user_id = ? AND type = ? AND source = ? AND source_id = ?",
				userID,
				"earn",
				"order",
				orderRecord.ID,
			).Find(&earnedTransactions).Error)
			require.Len(t, earnedTransactions, 1)
			assert.Equal(t, testCase.expectedPoints, earnedTransactions[0].Points)
		})
	}
}

func TestOrderServiceCompletionAllowsLegacyNonUSDOrderWithoutFXSnapshot(t *testing.T) {
	db, orderService := newTestOrderService(t)
	userID := uint(44)
	orderRecord := order.Order{
		OrderNumber:         "ORD-COMPLETE-LEGACY-EUR",
		UserID:              userID,
		Status:              "shipped",
		PaymentStatus:       "paid",
		SubtotalAmountMinor: 12000,
		DiscountAmountMinor: 2000,
		TotalAmountMinor:    11500,
		Currency:            "EUR",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.NoError(t, orderService.UpdateOrderStatus(orderRecord.ID, "completed"))
	processOrderCompletedOutboxEvent(t, db, orderService, orderRecord.ID)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "completed", savedOrder.Status)

	var count int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ? AND source_id = ?", userID, "earn", "order", orderRecord.ID).
		Count(&count).Error)
	assert.EqualValues(t, 0, count)
}

func TestOrderServiceCreateOrderUsesProductPriceCurrency(t *testing.T) {
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

func processOrderCompletedOutboxEvent(t *testing.T, db *gorm.DB, orderService *OrderService, orderID uint) {
	t.Helper()
	var event outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypeOrderCompleted,
		fmt.Sprint(orderID),
	).Order("id DESC").First(&event).Error)
	require.NoError(t, NewOrderCompletionOutboxHandler(orderService).Handle(context.Background(), event))
}

func TestOrderServiceCreateOrderPersistsDomesticProviderSettlementForCrossCurrencyOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 50, 5)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"wechat",
		"standard",
		"",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, "USD", createdOrder.Currency)
	assert.Equal(t, "CNY", createdOrder.PaymentCurrency)
	assert.Greater(t, createdOrder.PaymentAmountMinor, createdOrder.TotalAmountMinor)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&orderCount).Error)
	assert.EqualValues(t, 1, orderCount)

	var variant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	assert.Equal(t, 4, variant.Stock)
}

func TestOrderServiceCreateOrderAcceptsSupportedPaymentCurrency(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProductWithCurrency(t, db, 50, 5, "CNY")

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"wechat",
		"standard",
		"",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, "wechat", createdOrder.PaymentMethod)
	assert.Equal(t, "CNY", createdOrder.Currency)
}

func TestOrderServiceCreateOrderPersistsHistoricalFXSnapshot(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProductWithCurrency(t, db, 50, 5, "CNY")

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"wechat",
		"standard",
		"",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)

	snapshot, err := currency.ParseOrderFXSnapshot(savedOrder.FXSnapshotData)
	require.NoError(t, err)
	assert.Equal(t, "USD", snapshot.BaseCurrency)
	assert.Equal(t, "CNY", snapshot.OrderCurrency)
	assert.Equal(t, "7", snapshot.RateDecimal)
	assert.Equal(t, "test-rate", snapshot.Source)
	require.NotNil(t, snapshot.RateFetchedAt)
	assert.WithinDuration(t, time.Now().UTC().Add(-time.Minute), *snapshot.RateFetchedAt, 5*time.Second)
	assert.False(t, snapshot.CapturedAt.IsZero())
}

func TestOrderServiceRejectsPaymentManagedStatusUpdates(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:      "ORD-SYSTEM-STATUS",
		UserID:           42,
		Status:           "pending",
		PaymentStatus:    "unpaid",
		TotalAmountMinor: 10000,
		Currency:         "USD",
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
		OrderNumber:      "ORD-TRACKING",
		UserID:           42,
		Status:           "processing",
		TotalAmountMinor: 10000,
		Currency:         "USD",
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

	var shipment shippingdomain.TrackingShipment
	require.NoError(t, db.Where("order_id = ?", orderRecord.ID).First(&shipment).Error)
	assert.Equal(t, provider.ID, shipment.TrackingProviderID)
	assert.Equal(t, "TRACK123456", shipment.TrackingNumber)
	assert.Equal(t, "DHL", shipment.ProviderCarrierCode)
	assert.Equal(t, "pending", shipment.RegistrationStatus)
	assert.Equal(t, "pending", shipment.SyncStatus)
	require.NotNil(t, shipment.TrackingCarrierMappingID)
	assert.Equal(t, mapping.ID, *shipment.TrackingCarrierMappingID)
}

func TestOrderServiceUpdateTrackingInfoRequiresExplicitLocalTarget(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, _ := seedTrackingProviderCarrierAndService(t, db)

	orderRecord := order.Order{
		OrderNumber:      "ORD-TRACKING-LOCAL-TARGET-REQUIRED",
		UserID:           42,
		Status:           "processing",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	err := orderService.UpdateTrackingInfo(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACKDEFAULT123",
		TrackingProviderID: provider.ID,
	})

	require.ErrorIs(t, err, ErrTrackingLocalTargetRequired)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).Where("order_id = ?", orderRecord.ID).Count(&shipmentCount).Error)
	assert.Zero(t, shipmentCount)
}

func TestOrderServiceFulfillOrderMarksOrderShippedAndCreatesTrackingTask(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	result, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-FULFILL-123",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Order)
	require.Len(t, result.TrackingShipments, 1)
	assert.Empty(t, result.TrackingRegistrationError)
	assert.Equal(t, "shipped", result.Order.Status)
	assert.Equal(t, "shipped", result.Order.ShippingStatus)
	assert.NotNil(t, result.Order.ShippedAt)
	assert.Equal(t, orderRecord.ID, result.TrackingShipments[0].OrderID)
	assert.Equal(t, "TRACK-FULFILL-123", result.TrackingShipments[0].TrackingNumber)
	assert.Equal(t, "pending", result.TrackingShipments[0].SyncStatus)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "shipped", savedOrder.Status)
	assert.Equal(t, "shipped", savedOrder.ShippingStatus)
}

func TestOrderServiceFulfillOrderEnqueuesTrackingRegistration(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	provider.AutoRegister = true
	require.NoError(t, db.Save(&provider).Error)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-OUTBOX-REG",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-FULFILL-OUTBOX-123",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})
	require.NoError(t, err)

	var event outboxdomain.Event
	require.NoError(t, db.Where("event_type = ? AND aggregate_id = ?", outboxdomain.EventTypeTrackingShipmentRegistration, fmt.Sprint(orderRecord.ID)).First(&event).Error)
	assert.Equal(t, outboxdomain.EventStatusPending, event.Status)
	var payload outboxdomain.TrackingShipmentRegistrationPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, orderRecord.ID, payload.OrderID)
	assert.Equal(t, "TRACK-FULFILL-OUTBOX-123", payload.TrackingNumber)
}

func TestOrderServiceFulfillOrderRequiresPositiveConfirmedCustomsDeclaredValue(t *testing.T) {
	positiveDeclaredValue := 42.75
	zeroDeclaredValue := 0.0
	testCases := []struct {
		name           string
		declaredValue  *float64
		valueConfirmed bool
	}{
		{name: "missing value", declaredValue: nil, valueConfirmed: false},
		{name: "value not confirmed", declaredValue: &positiveDeclaredValue, valueConfirmed: false},
		{name: "zero value", declaredValue: &zeroDeclaredValue, valueConfirmed: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			db, orderService := newTestOrderService(t)
			orderRecord := order.Order{
				OrderNumber:      "ORD-FULFILL-CUSTOMS-" + strings.ToUpper(strings.ReplaceAll(testCase.name, " ", "-")),
				UserID:           42,
				Status:           "processing",
				PaymentStatus:    "paid",
				TotalAmountMinor: 10000,
				Currency:         "USD",
			}
			require.NoError(t, db.Create(&orderRecord).Error)
			variantID := uint(1)
			require.NoError(t, db.Create(&order.OrderItem{
				OrderID:       orderRecord.ID,
				ProductID:     1,
				VariantID:     &variantID,
				ProductName:   "Customs test product",
				SKU:           "CUSTOMS-TEST-SKU",
				Quantity:      1,
				PriceMinor:    10000,
				SubtotalMinor: 10000,
				TotalMinor:    10000,
				DeclaredValueMinor: func() *int64 {
					if testCase.declaredValue == nil {
						return nil
					}
					v := int64(*testCase.declaredValue * 100)
					return &v
				}(),
				DeclaredValueConfirmed: testCase.valueConfirmed,
			}).Error)

			_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{})

			require.ErrorIs(t, err, ErrOrderCustomsDeclarationIncomplete)
			assert.Contains(t, err.Error(), "CUSTOMS-TEST-SKU")

			var unchanged order.Order
			require.NoError(t, db.First(&unchanged, orderRecord.ID).Error)
			assert.Equal(t, "processing", unchanged.Status)
			assert.Equal(t, "pending", unchanged.ShippingStatus)
		})
	}
}

func TestOrderServiceFulfillOrderDoesNotMarkOrderShippedWhenCarrierMappingFails(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-MAPPING-FAIL",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	result, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-NO-MAPPING",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})

	require.ErrorIs(t, err, ErrTrackingCarrierMappingMissing)
	assert.Nil(t, result)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "processing", savedOrder.Status)
	assert.Equal(t, "pending", savedOrder.ShippingStatus)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).Where("order_id = ?", orderRecord.ID).Count(&shipmentCount).Error)
	assert.Zero(t, shipmentCount)
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
		OrderNumber:      "ORD-SYNC-TRACKING",
		UserID:           42,
		Status:           "shipped",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             orderRecord.ID,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "MOCK123456",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "pending",
		SyncStatus:          "pending",
		Enabled:             true,
	}).Error)

	result, err := orderService.SyncOrderTracking(context.Background(), orderRecord.ID, "MOCK123456")

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

func TestOrderServiceHideUnpaidCancelledOrPaymentExpiredOrderFromDefaultQueries(t *testing.T) {
	testCases := []struct {
		name          string
		status        string
		paymentStatus string
		canHide       bool
	}{
		{name: "paid order", status: "paid", paymentStatus: "paid", canHide: false},
		{name: "refunded order", status: "refunded", paymentStatus: "refunded", canHide: false},
		{name: "cancelled unpaid terminal order", status: "cancelled", paymentStatus: "unpaid", canHide: true},
		{name: "payment expired terminal order", status: "payment_expired", paymentStatus: "expired", canHide: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			db, orderService := newTestOrderService(t)
			orderRecord := order.Order{
				OrderNumber:      "ORD-HIDE-" + strings.ToUpper(strings.ReplaceAll(testCase.name, " ", "-")),
				UserID:           42,
				Status:           testCase.status,
				PaymentStatus:    testCase.paymentStatus,
				TotalAmountMinor: 10000,
				Currency:         "USD",
			}
			require.NoError(t, db.Create(&orderRecord).Error)

			err := orderService.HideUnpaidCancelledOrPaymentExpiredOrderFromDefaultQueries(orderRecord.ID)
			if testCase.canHide {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrOrderHideNotAllowed)
			}

			var retainedOrder order.Order
			require.NoError(t, db.Unscoped().First(&retainedOrder, orderRecord.ID).Error)
			assert.Equal(t, testCase.canHide, retainedOrder.DeletedAt.Valid)

			var defaultQueryOrder order.Order
			if testCase.canHide {
				require.ErrorIs(t, db.First(&defaultQueryOrder, orderRecord.ID).Error, gorm.ErrRecordNotFound)
			} else {
				require.NoError(t, db.First(&defaultQueryOrder, orderRecord.ID).Error)
			}
		})
	}
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
		&product.ProductSpecificationTemplate{},
		&product.SpecDefinition{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.Cart{},
		&product.CartItem{},
		&order.Order{},
		&order.OrderItem{},
		&order.OrderIdempotency{},
		&order.PolicyDisclosure{},
		&orderevidence.OrderEvidenceSnapshot{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceItem{},
		&orderevidence.OrderEvidenceAttachment{},
		&productrequirement.ProductQualityRequirementRule{},
		&outboxdomain.Event{},
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&currency.ExchangeRate{},
		&currency.ExchangeRateSyncLease{},
		&loyalty.UserLoyalty{},
		&loyalty.LoyaltyTransaction{},
		&loyalty.ProgramConfig{},
		&loyalty.MemberLevel{},
		&setting.Setting{},
		&paymentdomain.Transaction{},
		&paymentdomain.TaxRate{},
		&paymentdomain.StripeDispute{},
		&paymentdomain.PayPalDispute{},
		&paymentdomain.PaymentReview{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.QuoteSnapshot{},
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
	txManager.ConfigureCartRepository(repository.NewCartRepository(db))
	txManager.ConfigureOrderIdempotencyRepository(repository.NewOrderIdempotencyRepository(db))
	currencyPolicyService := seedTestCurrencyPolicy(t, db)
	exchangeRateRepo := repository.NewExchangeRateRepository(db)
	txManager.ConfigureSettingRepository(repository.NewSettingRepository(db))
	txManager.ConfigureOrderPolicyDisclosureRepository(repository.NewOrderPolicyDisclosureRepository(db))
	txManager.ConfigureExchangeRateRepository(exchangeRateRepo)
	txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	txManager.ConfigureProductQualityRequirementRepository(repository.NewProductQualityRequirementRepository(db))
	txManager.ConfigureOrderEvidenceSnapshotRepository(repository.NewOrderEvidenceSnapshotRepository(db))
	txManager.ConfigureOrderEvidenceRepository(repository.NewOrderEvidenceRepository(db))
	checkoutService.ConfigureCurrencyPolicy(currencyPolicyService)
	checkoutService.ConfigureExchangeRateRepository(exchangeRateRepo)
	eurRate := exchangeRateRecord("USD", "EUR", 0.9)
	cnyRate := exchangeRateRecord("USD", "CNY", 7.0)
	jpyRate := exchangeRateRecord("USD", "JPY", 150.0)
	require.NoError(t, db.Create(&eurRate).Error)
	require.NoError(t, db.Create(&cnyRate).Error)
	require.NoError(t, db.Create(&jpyRate).Error)
	programRepo := repository.NewLoyaltyProgramRepository(db)
	txManager.ConfigureLoyaltyProgramRepository(programRepo)
	programService := NewLoyaltyProgramService(programRepo)
	programService.ConfigureCurrencyPolicy(currencyPolicyService)
	_, err = programService.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ExchangeRatePoints:        100,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
	})
	require.NoError(t, err)
	checkoutService.ConfigureLoyaltyProgram(programService)
	numberGenerator, err := ordernumber.NewGenerator("test-order-number-secret", 0)
	require.NoError(t, err)

	orderService := NewOrderService(txManager, orderRepo, checkoutService, shippingService, numberGenerator)
	orderService.ConfigureOrderEvidenceSnapshot(NewOrderEvidenceSnapshotService())
	orderService.ConfigureOrderEvidence(NewOrderEvidenceService())
	return db, orderService
}

func newRedisBackedProductService(t *testing.T, db *gorm.DB) *ProductService {
	t.Helper()

	redisServer := miniredis.RunT(t)
	host, portText, err := net.SplitHostPort(redisServer.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)
	redisCache, err := cache.Init(config.RedisConfig{Host: host, Port: port})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = redisCache.Close()
	})
	return NewProductServiceWithCacheOptions(repository.NewProductRepository(db), redisCache, 60, 2)
}

func seedProduct(t *testing.T, db *gorm.DB, price float64, stock int) product.Product {
	t.Helper()

	return seedProductWithCurrency(t, db, price, stock, "USD")
}

func seedProductWithCurrency(t *testing.T, db *gorm.DB, price float64, stock int, currencyCode string) product.Product {
	t.Helper()

	record := seedProductShell(t, db, price, stock)
	record.Currency = currencyCode
	require.NoError(t, db.Save(&record).Error)
	if record.ShippingTemplateID != nil {
		require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).
			Where("id = ?", *record.ShippingTemplateID).
			Update("currency", currencyCode).Error)
		require.NoError(t, db.Model(&shippingdomain.ShippingRule{}).
			Where("template_id = ?", *record.ShippingTemplateID).
			Update("currency", currencyCode).Error)
	}
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:    record.ID,
		SKU:          record.SKU,
		Title:        "Default",
		OptionValues: "{}",
		Currency:     currencyCode,
		PriceMinor:   domainmoney.MustNew(int64(price*100), currencyCode).AmountMinor(),
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
		PriceMinor:         int64(price * 100),
		Stock:              stock,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedProductWithSKU(t *testing.T, db *gorm.DB, price float64, stock int, sku string) product.Product {
	t.Helper()

	shippingTemplateID := seedOrderTestShippingTemplateID(t, db)
	record := product.Product{
		ShippingTemplateID: shippingTemplateID,
		SKU:                sku,
		Name:               sku,
		Slug:               strings.ToLower(sku),
		Currency:           "USD",
		PriceMinor:         int64(price * 100),
		Stock:              stock,
		Status:             "active",
		Locale:             "en",
	}
	require.NoError(t, db.Create(&record).Error)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:    record.ID,
		SKU:          sku,
		Title:        "Default",
		OptionValues: "{}",
		Currency:     "USD",
		PriceMinor:   int64(price * 100),
		Stock:        stock,
		Weight:       9000,
		IsDefault:    true,
		IsActive:     true,
	}).Error)
	return record
}

func seedCheckoutCart(t *testing.T, db *gorm.DB, userID, productID, variantID uint, quantity int, price float64) product.Cart {
	t.Helper()

	cart := product.Cart{UserID: &userID}
	require.NoError(t, db.Create(&cart).Error)
	require.NoError(t, db.Create(&product.CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		VariantID: &variantID,
		Quantity:  quantity,
		PriceMinor: func() int64 {
			value, err := domainmoney.FromMajorFloat(price, "USD")
			if err != nil {
				t.Fatalf("invalid test price: %v", err)
			}
			return value.AmountMinor()
		}(),
		Currency: "USD",
	}).Error)
	return cart
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
		Name:               "Test standard shipping",
		Type:               "weight",
		FreeShipping:       true,
		FreeThresholdMinor: 10000,
		DefaultFeeMinor:    1000,
		Enabled:            true,
		Rules: []shippingdomain.ShippingRule{
			{
				Region:   "US",
				MinValue: 0,
				MaxValue: 0,
				FeeMinor: 1000,
			},
		},
	}
	require.NoError(t, db.Create(&template).Error)
}

func seedUserLoyalty(t *testing.T, db *gorm.DB, userID uint, points int) {
	t.Helper()

	require.NoError(t, db.FirstOrCreate(&loyalty.MemberLevel{}, loyalty.MemberLevel{
		Name:                "Test Level",
		MinPoints:           0,
		MaxPoints:           999999,
		DiscountRateDecimal: "5",
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
		Code: code,
		Type: couponType,
		ValueMinor: func() int64 {
			if couponType == "fixed" {
				return int64(value * 100)
			}
			return 0
		}(),
		ValueRateDecimal: func() string {
			if couponType == "percentage" {
				return strconv.FormatFloat(value, 'f', -1, 64)
			}
			return ""
		}(),
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

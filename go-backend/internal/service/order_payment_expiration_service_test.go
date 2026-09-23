package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/product"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderServiceExpiresStaleUnpaidOrderForNonCardPaymentMethod(t *testing.T) {
	db, orderService := newTestOrderService(t)
	now := time.Now().UTC()
	productRecord := seedProduct(t, db, 50, 5)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"paypal",
		"standard",
		"",
		0,
	)
	require.NoError(t, err)

	staleAt := now.Add(-45 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": staleAt,
		"updated_at": staleAt,
	}).Error)

	result, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)
	require.NoError(t, err)
	assert.Equal(t, 1, result.ScannedCandidates)
	assert.Equal(t, 1, result.ExpiredOrders)
	assert.Equal(t, int64(0), result.ExpiredOpenTransactions)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "payment_expired", savedOrder.Status)
	assert.Equal(t, "expired", savedOrder.PaymentStatus)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	assert.Equal(t, 5, savedVariant.Stock)
}

func TestOrderServiceDoesNotRollbackReservationsAgainWhenPaymentExpirationIsRepeated(t *testing.T) {
	db, orderService := newTestOrderService(t)
	now := time.Now().UTC()
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	cacheInvalidator := &recordingProductCacheInvalidator{}
	orderService.ConfigureProductCacheInvalidator(cacheInvalidator)
	seedUserLoyalty(t, db, userID, 1000)
	seedCoupon(t, db, "EXPIRE-ONCE", "fixed", 10, 1)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		userID,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 2}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"EXPIRE-ONCE",
		100,
	)
	require.NoError(t, err)

	oldActivity := now.Add(-45 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": oldActivity,
		"updated_at": oldActivity,
	}).Error)

	firstResult, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)
	require.NoError(t, err)
	require.Equal(t, 1, firstResult.ExpiredOrders)

	var savedVariant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	require.Equal(t, 5, savedVariant.Stock)

	var savedLoyalty loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", userID).First(&savedLoyalty).Error)
	require.Equal(t, 1000, savedLoyalty.AvailablePoints)
	require.Equal(t, 0, savedLoyalty.UsedPoints)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.Where("code = ?", "EXPIRE-ONCE").First(&savedCoupon).Error)
	require.Equal(t, 0, savedCoupon.UsedCount)

	secondExpiredTransactions, secondExpired, err := orderService.expirePaymentOrderIfStillEligible(
		createdOrder.ID,
		now.Add(-30*time.Minute),
		now.Add(time.Minute),
	)
	require.NoError(t, err)
	require.False(t, secondExpired)
	require.Zero(t, secondExpiredTransactions)
	require.Equal(t, []uint{productRecord.ID, productRecord.ID}, cacheInvalidator.productIDs)

	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&savedVariant).Error)
	require.Equal(t, 5, savedVariant.Stock)
	require.NoError(t, db.Where("user_id = ?", userID).First(&savedLoyalty).Error)
	require.Equal(t, 1000, savedLoyalty.AvailablePoints)
	require.Equal(t, 0, savedLoyalty.UsedPoints)
	require.NoError(t, db.Where("code = ?", "EXPIRE-ONCE").First(&savedCoupon).Error)
	require.Equal(t, 0, savedCoupon.UsedCount)
}

func TestOrderServicePaymentExpirationRestoresConsumedCheckoutCart(t *testing.T) {
	db, orderService := newTestOrderService(t)
	now := time.Now().UTC()
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 50)

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
	staleAt := now.Add(-45 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": staleAt,
		"updated_at": staleAt,
	}).Error)

	result, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)
	require.NoError(t, err)
	require.Equal(t, 1, result.ExpiredOrders)

	var restoredCartItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&restoredCartItems).Error)
	require.Len(t, restoredCartItems, 1)
	assert.Equal(t, 1, restoredCartItems[0].Quantity)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 5, savedVariant.Stock)
}

func TestOrderServicePaymentExpirationDoesNotRestoreCartForPaidOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	now := time.Now().UTC()
	userID := uint(42)
	productRecord := seedProduct(t, db, 50, 5)
	variant := product.ProductVariant{}
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	cart := seedCheckoutCart(t, db, userID, productRecord.ID, variant.ID, 1, 50)

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
	staleAt := now.Add(-45 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": staleAt,
		"updated_at": staleAt,
	}).Error)
	paymentService := NewPaymentService(orderService.txManager, repository.NewPaymentRepository(db))
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "paypal",
		OrderNumber:   createdOrder.OrderNumber,
		TransactionID: "capture_expiration_paid_first",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(createdOrder.PaymentAmountMinor, createdOrder.PaymentCurrency),
	}))
	// Model the expiration worker having selected this stale candidate before
	// payment won the order lock. The candidate transaction must then reload
	// the row and skip all rollback side effects.
	_, expired, err := orderService.expirePaymentOrderIfStillEligible(
		createdOrder.ID,
		now.Add(-30*time.Minute),
		now,
	)
	require.NoError(t, err)
	assert.False(t, expired)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "processing", savedOrder.Status)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)
	assert.False(t, savedOrder.FulfillmentHold)

	var savedVariant product.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 4, savedVariant.Stock)

	var remainingCartItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&remainingCartItems).Error)
	assert.Empty(t, remainingCartItems)
}

func TestOrderServicePaymentExpirationThenLateSuccessKeepsOrderExpiredAndRefunds(t *testing.T) {
	db, orderService := newTestOrderService(t)
	require.NoError(t, db.AutoMigrate(
		&paymentdomain.Refund{},
		&paymentdomain.RefundLineItem{},
		&paymentdomain.PaymentRefundExecution{},
		&paymentdomain.PaymentReview{},
	))
	orderService.txManager.ConfigurePaymentRefundExecutionRepository(repository.NewPaymentRefundExecutionRepository(db))
	now := time.Now().UTC()
	productRecord := seedProduct(t, db, 50, 5)
	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"paypal",
		"standard",
		"",
		0,
	)
	require.NoError(t, err)
	staleAt := now.Add(-45 * time.Minute)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"created_at": staleAt,
		"updated_at": staleAt,
	}).Error)

	result, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)
	require.NoError(t, err)
	require.Equal(t, 1, result.ExpiredOrders)
	var variant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	assert.Equal(t, 5, variant.Stock)

	paymentService := NewPaymentService(orderService.txManager, repository.NewPaymentRepository(db))
	input := VerifiedGatewayPaymentInput{
		Provider:      "paypal",
		OrderNumber:   createdOrder.OrderNumber,
		TransactionID: "capture_expiration_late_success",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(createdOrder.PaymentAmountMinor, createdOrder.PaymentCurrency),
	}
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(input))
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(input))

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "payment_expired", savedOrder.Status)
	assert.Equal(t, "expired", savedOrder.PaymentStatus)
	assert.True(t, savedOrder.FulfillmentHold)
	assert.Equal(t, 5, variantStock(t, db, variant.ID))

	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", input.TransactionID).First(&transaction).Error)
	assert.Equal(t, "completed", transaction.Status)
	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("payment_intent_id = ? AND reason = ?", input.TransactionID, "payment_succeeded_after_expiration").First(&review).Error)
	assert.Equal(t, "pending", review.Status)
	var refund paymentdomain.Refund
	require.NoError(t, db.Where("transaction_id = ? AND reason = ?", transaction.ID, "late_payment_refund:payment_succeeded_after_expiration").First(&refund).Error)
	assert.Equal(t, transaction.AmountMinor, refund.AmountMinor)
	assert.Equal(t, "pending", refund.Status)
	var execution paymentdomain.PaymentRefundExecution
	require.NoError(t, db.Where("refund_id = ?", refund.ID).First(&execution).Error)
	assert.Equal(t, paymentdomain.PaymentRefundExecutionStatusProcessing, execution.Status)
	var event outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypePaymentRefundExecutionRequested,
		fmt.Sprint(refund.ID),
	).First(&event).Error)
	assert.Equal(t, outboxdomain.EventStatusPending, event.Status)
}

func variantStock(t *testing.T, db *gorm.DB, variantID uint) int {
	t.Helper()
	var variant product.ProductVariant
	require.NoError(t, db.First(&variant, variantID).Error)
	return variant.Stock
}

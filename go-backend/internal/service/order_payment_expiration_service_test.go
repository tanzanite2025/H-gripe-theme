package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"
	attributionpkg "commerce-platform/internal/pkg/attribution"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 5, savedProduct.Stock)
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

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	require.Equal(t, 5, savedProduct.Stock)

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

	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	require.Equal(t, 5, savedProduct.Stock)
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
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"status":         "paid",
		"payment_status": "paid",
		"created_at":     now.Add(-45 * time.Minute),
		"updated_at":     now.Add(-45 * time.Minute),
	}).Error)

	result, err := orderService.ExpireStalePendingPayments(now, 30*time.Minute, 100)
	require.NoError(t, err)
	assert.Zero(t, result.ExpiredOrders)

	var remainingCartItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&remainingCartItems).Error)
	assert.Empty(t, remainingCartItems)
}

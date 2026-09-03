package service

import (
	"context"
	"sync"
	"testing"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelOrderRejectsPaidOrderWithoutChangingStateOrStock(t *testing.T) {
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
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"status":         "paid",
		"payment_status": "paid",
	}).Error)

	err = orderService.CancelOrder(createdOrder.ID, 42)

	require.ErrorIs(t, err, ErrPaidOrderCancellationNotAllowed)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "paid", savedOrder.Status)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 4, savedProduct.Stock)
}

func TestCancelOrderByNumberRejectsPaidPaymentStatusEvenWhenOrderIsPending(t *testing.T) {
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
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Update("payment_status", "paid").Error)

	err = orderService.CancelOrderByNumber(createdOrder.OrderNumber, 42)

	require.ErrorIs(t, err, ErrPaidOrderCancellationNotAllowed)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "pending", savedOrder.Status)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 4, savedProduct.Stock)
}

func TestCancelOrderRestoresStockForPendingUnpaidOrder(t *testing.T) {
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
	require.Equal(t, "pending", createdOrder.Status)
	require.Equal(t, "unpaid", createdOrder.PaymentStatus)

	require.NoError(t, orderService.CancelOrderByNumber(createdOrder.OrderNumber, 42))

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.Equal(t, "cancelled", savedOrder.Status)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 5, savedProduct.Stock)
}

func TestCancelOrderReversesCouponUsageWithoutDeletingAuditRecord(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 50, 5)
	seedCoupon(t, db, "CANCEL10", "fixed", 10, 1)
	require.NoError(t, db.Model(&coupon.Coupon{}).Where("code = ?", "CANCEL10").Update("usage_limit_per_user", 1).Error)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"CANCEL10",
		0,
	)
	require.NoError(t, err)

	var usage coupon.CouponUsage
	require.NoError(t, db.Where("order_id = ?", createdOrder.ID).First(&usage).Error)
	assert.Equal(t, coupon.CouponUsageStatusApplied, usage.Status)
	assert.Equal(t, float64(10), usage.Discount)

	require.NoError(t, orderService.CancelOrder(createdOrder.ID, 42))

	var savedCoupon coupon.Coupon
	require.NoError(t, db.Where("code = ?", "CANCEL10").First(&savedCoupon).Error)
	assert.Equal(t, 0, savedCoupon.UsedCount)

	var reversedUsage coupon.CouponUsage
	require.NoError(t, db.Where("order_id = ?", createdOrder.ID).First(&reversedUsage).Error)
	assert.Equal(t, usage.ID, reversedUsage.ID)
	assert.Equal(t, coupon.CouponUsageStatusReversed, reversedUsage.Status)
	assert.Equal(t, "cancelled", reversedUsage.ReversalReason)
	assert.NotNil(t, reversedUsage.ReversedAt)

	_, err = orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"CANCEL10",
		0,
	)
	require.NoError(t, err)
}

func TestCancelOrderConcurrentRequestsRollbackOnlyOnce(t *testing.T) {
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
	require.NoError(t, db.Create(&loyalty.UserLoyalty{
		UserID:          42,
		TotalPoints:     1000,
		AvailablePoints: 900,
		UsedPoints:      100,
	}).Error)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", createdOrder.ID).Updates(map[string]interface{}{
		"points_used":  100,
		"points_value": 1,
	}).Error)

	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer waitGroup.Done()
			results <- orderService.CancelOrderByNumber(createdOrder.OrderNumber, 42)
		}()
	}
	waitGroup.Wait()
	close(results)

	var succeeded int
	var cancellationErrors []error
	for result := range results {
		if result == nil {
			succeeded++
			continue
		}
		cancellationErrors = append(cancellationErrors, result)
	}
	require.Equal(t, 1, succeeded)
	require.Len(t, cancellationErrors, 1)
	require.ErrorIs(t, cancellationErrors[0], ErrOrderCancellationConflict)

	var savedProduct product.Product
	require.NoError(t, db.First(&savedProduct, productRecord.ID).Error)
	assert.Equal(t, 5, savedProduct.Stock)

	var savedLoyalty loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", 42).First(&savedLoyalty).Error)
	assert.Equal(t, 1000, savedLoyalty.AvailablePoints)
	assert.Equal(t, 0, savedLoyalty.UsedPoints)

	var refundTransactions []loyalty.LoyaltyTransaction
	require.NoError(t, db.Where("user_id = ? AND type = ? AND source = ? AND source_id = ?", 42, "refund", "order", createdOrder.ID).Find(&refundTransactions).Error)
	require.Len(t, refundTransactions, 1)
	assert.Equal(t, 100, refundTransactions[0].Points)
}

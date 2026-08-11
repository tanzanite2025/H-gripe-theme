package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"

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

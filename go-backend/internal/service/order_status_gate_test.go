package service

import (
	"testing"

	"commerce-platform/internal/domain/order"

	"github.com/stretchr/testify/require"
)

func TestOrderStatusUpdatesCannotSetShippedOutsideFulfillment(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:    "ORD-SHIPPED-STATUS-GATE",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		TotalAmount:    100,
		Currency:       "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "shipped"), ErrOrderFulfillmentStatusManaged)
	require.ErrorIs(t, orderService.UpdateShippingStatus(orderRecord.ID, "shipped"), ErrOrderFulfillmentStatusManaged)

	var unchanged order.Order
	require.NoError(t, db.First(&unchanged, orderRecord.ID).Error)
	require.Equal(t, "processing", unchanged.Status)
	require.Equal(t, "pending", unchanged.ShippingStatus)
}

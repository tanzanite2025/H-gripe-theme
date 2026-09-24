package service

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	shippingdomain "commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/require"
)

func TestOrderStatusUpdatesCannotSetShippedOutsideFulfillment(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:      "ORD-SHIPPED-STATUS-GATE",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.ErrorIs(t, orderService.UpdateOrderStatus(orderRecord.ID, "shipped"), ErrOrderFulfillmentStatusManaged)
	require.ErrorIs(t, orderService.UpdateShippingStatus(orderRecord.ID, "shipped"), ErrOrderFulfillmentStatusManaged)

	var unchanged order.Order
	require.NoError(t, db.First(&unchanged, orderRecord.ID).Error)
	require.Equal(t, "processing", unchanged.Status)
	require.Equal(t, "pending", unchanged.ShippingStatus)
}

func TestOrderShippingStatusCannotBeDeliveredBeforeAllPackagesArrive(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:      "ORD-SPLIT-DELIVERY-STATUS-GATE",
		UserID:           42,
		Status:           "shipped",
		PaymentStatus:    "paid",
		ShippingStatus:   "shipped",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "17TRACK",
		ProviderName: "17TRACK",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&[]shippingdomain.TrackingShipment{
		{OrderID: orderRecord.ID, TrackingProviderID: provider.ID, TrackingNumber: "SPLIT-WHEELSET-STATUS", ProviderCarrierCode: "DHL", Enabled: true},
		{OrderID: orderRecord.ID, TrackingProviderID: provider.ID, TrackingNumber: "SPLIT-FRAME-STATUS", ProviderCarrierCode: "DHL", Enabled: true},
	}).Error)

	require.ErrorIs(t, orderService.UpdateShippingStatus(orderRecord.ID, "delivered"), ErrOrderShippingDeliveryNotReady)

	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID: orderRecord.ID, TrackingNumber: "SPLIT-WHEELSET-STATUS", Status: "delivered", EventTime: time.Now().UTC(),
	}).Error)
	require.ErrorIs(t, orderService.UpdateShippingStatus(orderRecord.ID, "delivered"), ErrOrderShippingDeliveryNotReady)

	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID: orderRecord.ID, TrackingNumber: "SPLIT-FRAME-STATUS", Status: "delivered", EventTime: time.Now().UTC(),
	}).Error)
	require.NoError(t, orderService.UpdateShippingStatus(orderRecord.ID, "delivered"))

	var stored order.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	require.Equal(t, "delivered", stored.ShippingStatus)
	var deliveredEvent outbox.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outbox.EventTypeOrderDelivered,
		fmt.Sprint(orderRecord.ID),
	).First(&deliveredEvent).Error)
	var payload outbox.OrderDeliveredPayload
	require.NoError(t, json.Unmarshal(deliveredEvent.Payload, &payload))
	require.Equal(t, "admin_manual_update", payload.Source)
}

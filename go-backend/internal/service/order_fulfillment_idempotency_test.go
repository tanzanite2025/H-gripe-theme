package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderServiceFulfillOrderRetriesSameShippedTrackingWithoutChangingShipment(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-IDEMPOTENT",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	input := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-IDEMPOTENT-100",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	first, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, input)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NotNil(t, first.Order)
	require.Len(t, first.TrackingShipments, 1)

	second, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, input)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.NotNil(t, second.Order)
	require.Len(t, second.TrackingShipments, 1)
	assert.Equal(t, first.Order.ShippedAt, second.Order.ShippedAt)
	assert.Equal(t, first.TrackingShipments[0].ID, second.TrackingShipments[0].ID)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Count(&shipmentCount).Error)
	assert.Equal(t, int64(1), shipmentCount)
}

func TestOrderServiceFulfillOrderWithIdempotencyReplaysDurableRequest(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-DURABLE-IDEMPOTENCY",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	input := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-DURABLE-IDEMPOTENCY",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	first, err := orderService.FulfillOrderWithIdempotency(
		context.Background(), orderRecord.ID, input, 7, "fulfillment-request-1", "fulfillment-hash-1",
	)
	require.NoError(t, err)

	second, err := orderService.FulfillOrderWithIdempotency(
		context.Background(), orderRecord.ID, input, 7, "fulfillment-request-1", "fulfillment-hash-1",
	)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NotNil(t, second)
	assert.Equal(t, first.Order.ID, second.Order.ID)
	assert.Equal(t, first.TrackingShipments[0].ID, second.TrackingShipments[0].ID)

	var idempotencyCount int64
	require.NoError(t, db.Model(&order.OrderIdempotency{}).
		Where("user_id = ? AND scope = ? AND idempotency_key = ?", 7, "admin_order_fulfillment", "fulfillment-request-1").
		Count(&idempotencyCount).Error)
	assert.Equal(t, int64(1), idempotencyCount)

	_, err = orderService.FulfillOrderWithIdempotency(
		context.Background(), orderRecord.ID, input, 7, "fulfillment-request-1", "fulfillment-hash-2",
	)
	assert.ErrorIs(t, err, ErrOrderFulfillmentIdempotencyConflict)
}

func TestOrderServiceFulfillOrderQueuesShippingNotificationOnce(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, carrier, carrierService := seedTrackingProviderCarrierAndService(t, db)
	carrier.TrackingURL = "https://tracking.example.test/track/{tracking_number}"
	require.NoError(t, db.Save(&carrier).Error)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-EMAIL",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
		ShippingAddress: order.Address{
			FirstName: "Ada",
			LastName:  "Rider",
			Email:     "ada.rider@example.test",
		},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	input := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-NOTIFY-100",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, input)
	require.NoError(t, err)

	var event outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", outboxdomain.EventTypeOrderShipped).First(&event).Error)
	eventKey := event.EventKey
	assert.Equal(t, outboxdomain.EventTypeOrderShipped, event.EventType)
	assert.Equal(t, outboxdomain.EventStatusPending, event.Status)

	var payload outboxdomain.OrderShippedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, "ada.rider@example.test", payload.RecipientEmail)
	assert.Equal(t, "Ada Rider", payload.CustomerName)
	assert.Equal(t, orderRecord.OrderNumber, payload.OrderNumber)
	assert.Equal(t, "DHL", payload.CarrierName)
	assert.Equal(t, input.TrackingNumber, payload.TrackingNumber)
	assert.Equal(t, "https://tracking.example.test/track/TRACK-NOTIFY-100", payload.TrackingURL)
	assert.False(t, payload.ShippedAt.IsZero())

	_, err = orderService.FulfillOrder(context.Background(), orderRecord.ID, input)
	require.NoError(t, err)

	var eventCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_key = ?", eventKey).Count(&eventCount).Error)
	assert.Equal(t, int64(1), eventCount)
}

func TestOrderServiceFulfillOrderBlocksActivePaymentDispute(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-DISPUTE-HOLD",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)
	require.NoError(t, db.Create(&paymentdomain.StripeDispute{
		StripeDisputeID: "dp_fulfillment_hold",
		OrderID:         &orderRecord.ID,
		AmountMinor:     10000,
		Currency:        "USD",
		Status:          "under_review",
	}).Error)

	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-DISPUTE-HOLD",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})
	require.ErrorIs(t, err, ErrOrderFulfillmentOnHold)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "processing", savedOrder.Status)
	assert.Equal(t, "pending", savedOrder.ShippingStatus)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Count(&shipmentCount).Error)
	assert.Zero(t, shipmentCount)
}

func TestOrderServiceFulfillOrderBlocksPendingPaymentReview(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-REVIEW-HOLD",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)
	require.NoError(t, db.Create(&paymentdomain.PaymentReview{
		OrderID: &orderRecord.ID,
		Status:  "pending",
		Reason:  "manual_payment_review",
		Source:  "operator",
	}).Error)

	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-REVIEW-HOLD",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})
	require.ErrorIs(t, err, ErrOrderFulfillmentOnHold)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "processing", savedOrder.Status)
	assert.Equal(t, "pending", savedOrder.ShippingStatus)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Count(&shipmentCount).Error)
	assert.Zero(t, shipmentCount)
}

func TestOrderServiceFulfillOrderAllowsAdditionalTrackingAfterShipment(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-TRACKING-CONFLICT",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	firstInput := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-CONFLICT-ORIGINAL",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, firstInput)
	require.NoError(t, err)

	second, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-CONFLICT-REPLACEMENT",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})
	require.NoError(t, err)
	require.Len(t, second.TrackingShipments, 2)

	var stored order.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, "shipped", stored.Status)
	assert.Equal(t, "shipped", stored.ShippingStatus)
	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).Where("order_id = ?", orderRecord.ID).Count(&shipmentCount).Error)
	assert.Equal(t, int64(2), shipmentCount)
}

func TestOrderServiceFulfillOrderRejectsRetryWithoutLocalTrackingSource(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-TRACKING-STRICT-RETRY",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	firstInput := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-STRICT-RETRY",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, firstInput)
	require.NoError(t, err)

	_, err = orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-STRICT-RETRY",
		TrackingProviderID: provider.ID,
	})
	require.ErrorIs(t, err, ErrTrackingLocalTargetRequired)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Count(&shipmentCount).Error)
	assert.Equal(t, int64(1), shipmentCount)
}

func TestOrderServiceAddsPackageAfterShipmentWithoutReplacingExistingPackage(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:      "ORD-FULFILL-TRACKING-CORRECTION",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		ShippingStatus:   "pending",
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	originalInput := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-CORRECTION-ORIGINAL",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	first, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, originalInput)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NotNil(t, first.Order)
	require.Len(t, first.TrackingShipments, 1)
	require.NotNil(t, first.Order.ShippedAt)

	additionalInput := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-CORRECTION-CORRECTED",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	require.NoError(t, orderService.UpdateTrackingInfo(
		context.Background(),
		orderRecord.ID,
		additionalInput,
	))

	var stored order.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, "shipped", stored.Status)
	assert.Equal(t, "shipped", stored.ShippingStatus)
	require.NotNil(t, stored.ShippedAt)
	assert.WithinDuration(t, *first.Order.ShippedAt, *stored.ShippedAt, time.Second)

	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             orderRecord.ID,
		TrackingNumber:      additionalInput.TrackingNumber,
		ProviderCarrierCode: "DHL-EXP-US",
		Status:              "In Transit",
		Description:         "event attached to the corrected tracking number",
		EventTime:           time.Now().UTC(),
	}).Error)

	assembly, err := NewOrderEvidencePackageAssembler(
		repository.NewOrderRepository(db),
		repository.NewOrderEvidenceRepository(db),
		repository.NewShippingRepository(db),
	).Assemble(orderRecord.ID)
	require.NoError(t, err)
	require.NotNil(t, assembly)
	require.Len(t, assembly.TrackingEvents, 1)
	assert.Equal(t, additionalInput.TrackingNumber, assembly.TrackingEvents[0].TrackingNumber)

	retry, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, additionalInput)
	require.NoError(t, err)
	require.NotNil(t, retry)
	require.NotNil(t, retry.Order)
	require.Len(t, retry.TrackingShipments, 2)
	assert.Equal(t, first.Order.ShippedAt, retry.Order.ShippedAt)
	assert.Contains(t, []string{retry.TrackingShipments[0].TrackingNumber, retry.TrackingShipments[1].TrackingNumber}, additionalInput.TrackingNumber)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).Where("order_id = ?", orderRecord.ID).Count(&shipmentCount).Error)
	assert.Equal(t, int64(2), shipmentCount)

	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
}

package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
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
		OrderNumber:    "ORD-FULFILL-IDEMPOTENT",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		TotalAmount:    100,
		Currency:       "USD",
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
	require.NotNil(t, first.TrackingShipment)

	second, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, input)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.NotNil(t, second.Order)
	require.NotNil(t, second.TrackingShipment)
	assert.Equal(t, first.Order.TrackingNumber, second.Order.TrackingNumber)
	assert.Equal(t, first.Order.ShippedAt, second.Order.ShippedAt)
	assert.Equal(t, first.TrackingShipment.ID, second.TrackingShipment.ID)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Count(&shipmentCount).Error)
	assert.Equal(t, int64(1), shipmentCount)
}

func TestOrderServiceFulfillOrderRejectsDifferentTrackingAfterShipment(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:    "ORD-FULFILL-TRACKING-CONFLICT",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		TotalAmount:    100,
		Currency:       "USD",
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

	_, err = orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-CONFLICT-REPLACEMENT",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})
	require.ErrorIs(t, err, ErrOrderFulfillmentTrackingConflict)

	var stored order.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, "TRACK-CONFLICT-ORIGINAL", stored.TrackingNumber)
	assert.Equal(t, "shipped", stored.Status)
	assert.Equal(t, "shipped", stored.ShippingStatus)
}

func TestOrderServiceFulfillOrderRejectsRetryWithoutLocalTrackingSource(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:    "ORD-FULFILL-TRACKING-STRICT-RETRY",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		TotalAmount:    100,
		Currency:       "USD",
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

func TestOrderServiceCorrectsTrackingAfterShipmentAndKeepsFulfillmentIdempotency(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:    "ORD-FULFILL-TRACKING-CORRECTION",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		TotalAmount:    100,
		Currency:       "USD",
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
	require.NotNil(t, first.TrackingShipment)
	require.NotNil(t, first.Order.ShippedAt)

	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             orderRecord.ID,
		TrackingNumber:      originalInput.TrackingNumber,
		ProviderCarrierCode: "DHL-EXP-US",
		Status:              "Delivered",
		Description:         "event attached to the mistyped tracking number",
		EventTime:           time.Now().UTC(),
	}).Error)

	correctedInput := OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-CORRECTION-CORRECTED",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	}
	require.NoError(t, orderService.UpdateTrackingInfo(
		context.Background(),
		orderRecord.ID,
		correctedInput,
	))

	var stored order.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, correctedInput.TrackingNumber, stored.TrackingNumber)
	assert.Equal(t, "shipped", stored.Status)
	assert.Equal(t, "shipped", stored.ShippingStatus)
	require.NotNil(t, stored.ShippedAt)
	assert.WithinDuration(t, *first.Order.ShippedAt, *stored.ShippedAt, time.Second)

	correctedShipment, err := orderService.GetAdminOrderTrackingShipment(orderRecord.ID)
	require.NoError(t, err)
	require.NotNil(t, correctedShipment)
	assert.Equal(t, correctedInput.TrackingNumber, correctedShipment.TrackingNumber)

	var currentEvents []shippingdomain.TrackingEvent
	require.NoError(t, db.Where("order_id = ?", orderRecord.ID).
		Find(&currentEvents).Error)
	assert.Empty(t, currentEvents)

	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             orderRecord.ID,
		TrackingNumber:      correctedInput.TrackingNumber,
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
	assert.Equal(t, correctedInput.TrackingNumber, assembly.TrackingEvents[0].TrackingNumber)

	retry, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, correctedInput)
	require.NoError(t, err)
	require.NotNil(t, retry)
	require.NotNil(t, retry.Order)
	require.NotNil(t, retry.TrackingShipment)
	assert.Equal(t, correctedInput.TrackingNumber, retry.Order.TrackingNumber)
	assert.Equal(t, first.Order.ShippedAt, retry.Order.ShippedAt)
	assert.Equal(t, correctedShipment.ID, retry.TrackingShipment.ID)

	_, err = orderService.FulfillOrder(context.Background(), orderRecord.ID, originalInput)
	require.ErrorIs(t, err, ErrOrderFulfillmentTrackingConflict)

	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, correctedInput.TrackingNumber, stored.TrackingNumber)
}

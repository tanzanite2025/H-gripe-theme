package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/order"
	productdomain "commerce-platform/internal/domain/product"
	shippingdomain "commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderServiceCreateOrderSnapshotsHighValueSignatureRequirement(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, order.HighValueSignatureThresholdUSD, 5)

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
	assert.True(t, createdOrder.SignatureRequired)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, createdOrder.ID).Error)
	assert.True(t, savedOrder.SignatureRequired)
}

func TestOrderServiceFulfillHighValueOrderRequiresSignatureConfirmation(t *testing.T) {
	db, orderService := newTestOrderService(t)
	provider, _, carrierService := seedTrackingProviderCarrierAndService(t, db)
	mapping := seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	orderRecord := order.Order{
		OrderNumber:       "ORD-FULFILL-SIGNATURE",
		UserID:            42,
		Status:            "processing",
		PaymentStatus:     "paid",
		ShippingStatus:    "pending",
		SignatureRequired: true,
		TotalAmountMinor:  order.HighValueSignatureThresholdUSDMinor,
		Currency:          "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForTest(t, db, &orderRecord)

	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-FULFILL-SIGNATURE",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
	})
	require.ErrorIs(t, err, ErrOrderFulfillmentSignatureConfirmationRequired)

	var unchanged order.Order
	require.NoError(t, db.First(&unchanged, orderRecord.ID).Error)
	assert.Equal(t, "processing", unchanged.Status)
	assert.Equal(t, "pending", unchanged.ShippingStatus)

	var shipmentCount int64
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Count(&shipmentCount).Error)
	assert.Zero(t, shipmentCount)

	result, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{
		TrackingNumber:     "TRACK-FULFILL-SIGNATURE",
		TrackingProviderID: provider.ID,
		CarrierServiceID:   &carrierService.ID,
		SignatureConfirmed: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Order)
	assert.Equal(t, "shipped", result.Order.Status)
	require.Len(t, result.TrackingShipments, 1)
	require.NotNil(t, result.TrackingShipments[0].TrackingCarrierMappingID)
	assert.Equal(t, mapping.ID, *result.TrackingShipments[0].TrackingCarrierMappingID)
}

func TestOrderServiceCreateMadeToOrderAllowsZeroStockAndDoesNotDeductInventory(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 120, 0)
	productRecord.FulfillmentMode = productdomain.FulfillmentModeMadeToOrder
	require.NoError(t, db.Save(&productRecord).Error)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
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
	assert.Equal(t, order.FulfillmentModeMadeToOrder, createdOrder.FulfillmentMode)

	var variant productdomain.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	assert.Zero(t, variant.Stock)
}

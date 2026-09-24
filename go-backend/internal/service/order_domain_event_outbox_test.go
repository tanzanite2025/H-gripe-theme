package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/aftersales"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCanonicalDomainEventOutboxHandlerValidatesEnvelopeWithoutSending(t *testing.T) {
	handler := NewCanonicalDomainEventOutboxHandler()
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDelivered,
		Payload:   []byte(`{"schema_version":1,"idempotency_key":"delivery_fact:1:TRK-1","order_id":1,"order_number":"ORD-1"}`),
	}))
	assert.Error(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDelivered,
		Payload:   []byte(`{"schema_version":1}`),
	}))
}

func TestCanonicalDomainEventOutboxHandlerRejectsIncompleteFactPayload(t *testing.T) {
	handler := NewCanonicalDomainEventOutboxHandler()
	err := handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderShipped,
		Payload:   []byte(`{"schema_version":1,"idempotency_key":"shipment:1","order_id":1,"order_number":"ORD-1"}`),
	})
	assert.ErrorContains(t, err, "tracking_number")
}

func newCanonicalOutboxTestRepository(t *testing.T) (*gorm.DB, *repository.OutboxRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	return db, repository.NewOutboxRepository(db)
}

func TestCanonicalOrderPaymentSucceededEventIsVersionedAndIdempotent(t *testing.T) {
	db, repo := newCanonicalOutboxTestRepository(t)
	amount, err := domainmoney.New(12500, "USD")
	require.NoError(t, err)
	orderRecord := &order.Order{
		ID:            11,
		OrderNumber:   "ORD-EVENT-11",
		Status:        "pending",
		PaymentStatus: "unpaid",
		ShippingAddress: order.Address{
			FirstName: "Ada",
			LastName:  "Buyer",
			Email:     "ada@example.com",
		},
	}
	occurredAt := time.Date(2026, 9, 19, 8, 0, 0, 0, time.UTC)
	input := VerifiedGatewayPaymentInput{
		TransactionID: "txn-event-11",
		PaymentMethod: "stripe",
		Amount:        amount,
	}
	require.NoError(t, enqueueOrderPaymentSucceededDomainEvent(repo, orderRecord, input, "pending", "processing", occurredAt))
	require.NoError(t, enqueueOrderPaymentSucceededDomainEvent(repo, orderRecord, input, "pending", "processing", occurredAt))

	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeOrderPaymentSucceeded).First(&event).Error)
	assert.Equal(t, "order.payment_succeeded:11:txn-event-11", event.EventKey)
	assert.Equal(t, outbox.EventStatusPending, event.Status)
	var payload outbox.OrderPaymentSucceededPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, 1, payload.SchemaVersion)
	assert.Equal(t, "payment_transaction:txn-event-11", payload.IdempotencyKey)
	assert.Equal(t, "processing", payload.NewOrderStatus)
	assert.Equal(t, int64(12500), payload.AmountMinor)
	assert.Equal(t, "ada@example.com", payload.RecipientEmail)
	assert.Equal(t, "Ada Buyer", payload.CustomerName)
}

func TestCanonicalOrderShippedAndDeliveredEventsCarryStableFacts(t *testing.T) {
	db, repo := newCanonicalOutboxTestRepository(t)
	orderRecord := &order.Order{ID: 22, OrderNumber: "ORD-EVENT-22", Status: "processing", ShippingStatus: "pending"}
	tracking := &resolvedOrderTrackingUpdate{
		trackingShipment: TrackingShipmentInput{
			ID:             77,
			OrderID:        orderRecord.ID,
			TrackingNumber: "DHL-77",
		},
		carrierName:         "DHL",
		trackingURLTemplate: "https://tracking.example/{tracking_number}",
	}
	shippedAt := time.Date(2026, 9, 19, 9, 0, 0, 0, time.UTC)
	require.NoError(t, enqueueOrderShippedDomainEvent(repo, orderRecord, tracking, "processing", "pending", shippedAt))
	require.NoError(t, enqueueOrderDeliveredDomainEvent(repo, orderRecord, "DHL-77", "shipped", shippedAt.Add(24*time.Hour), "carrier_webhook"))

	var shipped outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeOrderShipped).First(&shipped).Error)
	var shippedPayload outbox.OrderShippedPayload
	require.NoError(t, json.Unmarshal(shipped.Payload, &shippedPayload))
	assert.Equal(t, uint(77), shippedPayload.ShipmentID)
	assert.Equal(t, "https://tracking.example/DHL-77", shippedPayload.TrackingURL)

	var delivered outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeOrderDelivered).First(&delivered).Error)
	var deliveredPayload outbox.OrderDeliveredPayload
	require.NoError(t, json.Unmarshal(delivered.Payload, &deliveredPayload))
	assert.Equal(t, "delivery_fact:22:DHL-77", deliveredPayload.IdempotencyKey)
	assert.Equal(t, "carrier_webhook", deliveredPayload.Source)
}

func TestCanonicalAfterSalesStatusEventUsesTransitionID(t *testing.T) {
	db, repo := newCanonicalOutboxTestRepository(t)
	createdAt := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	caseRecord := &aftersales.AfterSalesCase{
		ID:          33,
		OrderID:     44,
		OrderNumber: "ORD-EVENT-44",
		Type:        aftersales.TypeReturnRefund,
		Status:      aftersales.StatusApproved,
	}
	transition := &aftersales.AfterSalesCaseEvent{
		ID:         88,
		CaseID:     caseRecord.ID,
		FromStatus: aftersales.StatusReviewing,
		ToStatus:   aftersales.StatusApproved,
		UpdatedBy:  9,
		CreatedAt:  createdAt,
	}
	shipment := &aftersales.AfterSalesReturnShipment{ID: 99, Carrier: "DHL", TrackingNumber: "RET-99", LabelURL: "https://shop.example/labels/ret-99.pdf"}
	require.NoError(t, enqueueAfterSalesStatusChangedDomainEvent(repo, caseRecord, transition, shipment))

	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeAfterSalesStatusChanged).First(&event).Error)
	var payload outbox.AfterSalesStatusChangedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, "after_sales.status_changed:33:88", event.EventKey)
	assert.Equal(t, uint(88), payload.TransitionID)
	assert.Equal(t, "after_sales_case:33:transition:88", payload.IdempotencyKey)
	assert.Equal(t, "approved", payload.NewStatus)
	assert.Equal(t, uint(99), payload.ReturnShipmentID)
	assert.Equal(t, "https://shop.example/labels/ret-99.pdf", payload.LabelURL)
}

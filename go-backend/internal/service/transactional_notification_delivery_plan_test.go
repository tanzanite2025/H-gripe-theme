package service

import (
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTransactionalNotificationDeliveryPlanBuildsStableShippingPlan(t *testing.T) {
	shippedAt := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(outbox.OrderShippedPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{
			SchemaVersion:  1,
			OccurredAt:     shippedAt,
			IdempotencyKey: "shipment:77",
		},
		NotificationAudienceSnapshot: outbox.NotificationAudienceSnapshot{
			RecipientEmail: "buyer@example.com",
			Locale:         "de",
			CustomerName:   "Buyer",
		},
		OrderID:           42,
		OrderNumber:       "ORD-42",
		CarrierName:       "DHL",
		TrackingNumber:    "DHL-77",
		TrackingURL:       "https://tracking.example/DHL-77",
		ShippedAt:         shippedAt,
		NewOrderStatus:    "shipped",
		NewShippingStatus: "shipped",
	})
	require.NoError(t, err)

	plan, err := BuildTransactionalNotificationDeliveryPlan(
		outbox.Event{
			ID:        9,
			EventKey:  "order.shipped:shipment:77",
			EventType: outbox.EventTypeOrderShipped,
			Payload:   payload,
		},
		TransactionalNotificationDeliveryContext{
			Locale: "fr-CA",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, NotificationTemplateOrderShippingNotification, plan.TemplateCode)
	assert.Equal(t, "fr", plan.Locale)
	assert.Equal(t, "buyer@example.com", plan.RecipientEmail)
	assert.Equal(t, "order.shipped:shipment:77", plan.SourceEventKey)
	assert.Equal(t, "notification:shipment:77:order_shipping_notification", plan.IdempotencyKey)
	assert.Equal(t, "ORD-42", plan.Variables["order_number"])
	assert.Equal(t, "DHL-77", plan.Variables["tracking_number"])
	assert.Equal(t, shippedAt.Format(time.RFC3339), plan.Variables["shipped_at"])
	assert.Equal(t, "Buyer", plan.Variables["customer_name"])
}

func TestBuildTransactionalNotificationDeliveryPlanRejectsAudienceRecipientMismatch(t *testing.T) {
	payload, err := json.Marshal(outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{OccurredAt: time.Now().UTC()},
		NotificationAudienceSnapshot: outbox.NotificationAudienceSnapshot{RecipientEmail: "snapshot@example.com"},
		OrderID:                      42,
		OrderNumber:                  "ORD-42",
		DeliveredAt:                  time.Now().UTC(),
	})
	require.NoError(t, err)

	_, err = BuildTransactionalNotificationDeliveryPlan(
		outbox.Event{EventType: outbox.EventTypeOrderDelivered, Payload: payload},
		TransactionalNotificationDeliveryContext{RecipientEmail: "other@example.com"},
	)
	assert.ErrorIs(t, err, ErrNotificationDeliveryContextConflict)
}

func TestBuildTransactionalNotificationDeliveryPlanRejectsMissingRecipientAndUnknownVariables(t *testing.T) {
	payload, err := json.Marshal(outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{OccurredAt: time.Now().UTC()},
		OrderID:                      42,
		OrderNumber:                  "ORD-42",
		DeliveredAt:                  time.Now().UTC(),
	})
	require.NoError(t, err)
	event := outbox.Event{EventType: outbox.EventTypeOrderDelivered, Payload: payload}

	_, err = BuildTransactionalNotificationDeliveryPlan(event, TransactionalNotificationDeliveryContext{})
	assert.ErrorIs(t, err, ErrNotificationDeliveryRecipientRequired)

	_, err = BuildTransactionalNotificationDeliveryPlan(event, TransactionalNotificationDeliveryContext{
		RecipientEmail: "buyer@example.com",
		Variables:      map[string]string{"internal_secret": "must-not-render"},
	})
	assert.ErrorIs(t, err, ErrNotificationTemplateVariableUnknown)
}

func TestBuildTransactionalNotificationDeliveryPlanDoesNotAllowContextToRewriteEventFacts(t *testing.T) {
	payload, err := json.Marshal(outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{OccurredAt: time.Now().UTC()},
		OrderID:                      42,
		OrderNumber:                  "ORD-42",
		DeliveredAt:                  time.Now().UTC(),
	})
	require.NoError(t, err)

	_, err = BuildTransactionalNotificationDeliveryPlan(
		outbox.Event{EventType: outbox.EventTypeOrderDelivered, Payload: payload},
		TransactionalNotificationDeliveryContext{
			RecipientEmail: "buyer@example.com",
			Variables:      map[string]string{"order_number": "FORGED"},
		},
	)
	assert.ErrorIs(t, err, ErrNotificationDeliveryContextConflict)
}

func TestBuildTransactionalNotificationDeliveryPlanLeavesSilentAfterSalesFactsWithoutRecipient(t *testing.T) {
	payload, err := json.Marshal(outbox.AfterSalesStatusChangedPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{OccurredAt: time.Now().UTC()},
		CaseID:                       99,
		NewStatus:                    "inspecting",
	})
	require.NoError(t, err)

	plan, err := BuildTransactionalNotificationDeliveryPlan(
		outbox.Event{EventType: outbox.EventTypeAfterSalesStatusChanged, Payload: payload},
		TransactionalNotificationDeliveryContext{},
	)
	require.NoError(t, err)
	assert.Empty(t, plan.TemplateCode)
	assert.Empty(t, plan.RecipientEmail)
}

func TestBuildTransactionalNotificationDeliveryPlanUsesSnapshotOverridesForAfterSales(t *testing.T) {
	payload, err := json.Marshal(outbox.AfterSalesStatusChangedPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{OccurredAt: time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)},
		CaseID:                       99,
		OrderNumber:                  "ORD-99",
		CaseType:                     "refund_only",
		NewStatus:                    "completed",
	})
	require.NoError(t, err)

	plan, err := BuildTransactionalNotificationDeliveryPlan(
		outbox.Event{EventKey: "after_sales.status_changed:99:2", EventType: outbox.EventTypeAfterSalesStatusChanged, Payload: payload},
		TransactionalNotificationDeliveryContext{
			RecipientEmail: "buyer@example.com",
			Variables:      map[string]string{"refund_amount": "15.00"},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, NotificationTemplateAfterSalesCompleted, plan.TemplateCode)
	assert.Equal(t, "99", plan.Variables["after_sales_case_number"])
	assert.Equal(t, "15.00", plan.Variables["refund_amount"])
}

func TestBuildTransactionalNotificationDeliveryPlanCarriesReturnLabelSnapshot(t *testing.T) {
	payload, err := json.Marshal(outbox.AfterSalesStatusChangedPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{OccurredAt: time.Now().UTC()},
		CaseID:                       101,
		OrderNumber:                  "ORD-101",
		CaseType:                     "return_refund",
		NewStatus:                    "awaiting_return",
		WarehouseName:                "EU Returns Hub",
		WarehouseAddress:             "1 Returns Way",
		LabelURL:                     "https://shop.example/labels/101.pdf",
	})
	require.NoError(t, err)

	plan, err := BuildTransactionalNotificationDeliveryPlan(
		outbox.Event{EventType: outbox.EventTypeAfterSalesStatusChanged, Payload: payload},
		TransactionalNotificationDeliveryContext{RecipientEmail: "buyer@example.com"},
	)
	require.NoError(t, err)
	assert.Equal(t, NotificationTemplateAfterSalesAwaitingReturn, plan.TemplateCode)
	assert.Equal(t, "https://shop.example/labels/101.pdf", plan.Variables["return_label_url"])
}

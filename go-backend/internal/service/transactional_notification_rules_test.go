package service

import (
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/outbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTransactionalNotificationTemplateMapsOrderFacts(t *testing.T) {
	tests := []struct {
		eventType    string
		templateCode string
	}{
		{outbox.EventTypeOrderPaymentSucceeded, NotificationTemplateOrderConfirmation},
		{outbox.EventTypeOrderPaymentExpired, NotificationTemplateOrderPaymentExpired},
		{outbox.EventTypeOrderCancelled, NotificationTemplateOrderCancelled},
		{outbox.EventTypeOrderShipped, NotificationTemplateOrderShippingNotification},
		{outbox.EventTypeOrderDelivered, NotificationTemplateOrderDelivered},
		{outbox.EventTypeOrderCompleted, NotificationTemplateOrderCompleted},
		{outbox.EventTypeOrderRefunded, NotificationTemplateOrderRefunded},
	}

	for _, test := range tests {
		resolution, err := ResolveTransactionalNotificationTemplate(test.eventType, nil)
		require.NoError(t, err, test.eventType)
		assert.Equal(t, test.templateCode, resolution.TemplateCode, test.eventType)
		assert.True(t, resolution.CustomerVisible, test.eventType)
		assert.NotEmpty(t, resolution.RequiredVariables, test.eventType)
	}
}

func TestResolveTransactionalNotificationTemplateMapsAfterSalesStatuses(t *testing.T) {
	tests := []struct {
		status       string
		templateCode string
	}{
		{aftersales.StatusRequested, NotificationTemplateAfterSalesRequested},
		{aftersales.StatusApproved, NotificationTemplateAfterSalesApproved},
		{aftersales.StatusAwaitingReturn, NotificationTemplateAfterSalesAwaitingReturn},
		{aftersales.StatusReturnInTransit, NotificationTemplateAfterSalesReturnInTransit},
		{aftersales.StatusReceived, NotificationTemplateAfterSalesReceived},
		{aftersales.StatusResolving, NotificationTemplateAfterSalesResolving},
		{aftersales.StatusCompleted, NotificationTemplateAfterSalesCompleted},
		{aftersales.StatusRejected, NotificationTemplateAfterSalesRejected},
	}

	for _, test := range tests {
		payload, err := json.Marshal(outbox.AfterSalesStatusChangedPayload{
			CaseID:       42,
			NewStatus:    test.status,
			TransitionID: 9,
		})
		require.NoError(t, err)
		resolution, err := ResolveTransactionalNotificationTemplate(outbox.EventTypeAfterSalesStatusChanged, payload)
		require.NoError(t, err, test.status)
		assert.Equal(t, test.templateCode, resolution.TemplateCode, test.status)
		assert.True(t, resolution.CustomerVisible, test.status)
	}
}

func TestResolveTransactionalNotificationTemplateLeavesInternalAfterSalesFactsSilent(t *testing.T) {
	for _, status := range []string{aftersales.StatusInspecting, aftersales.StatusException, aftersales.StatusCancelled} {
		payload, err := json.Marshal(outbox.AfterSalesStatusChangedPayload{CaseID: 42, NewStatus: status})
		require.NoError(t, err)
		resolution, err := ResolveTransactionalNotificationTemplate(outbox.EventTypeAfterSalesStatusChanged, payload)
		require.NoError(t, err)
		assert.Empty(t, resolution.TemplateCode, status)
		assert.False(t, resolution.CustomerVisible, status)
	}
}

func TestResolveTransactionalNotificationTemplateRejectsUnknownEventAndMalformedAfterSalesPayload(t *testing.T) {
	_, err := ResolveTransactionalNotificationTemplate("order.unknown", nil)
	assert.ErrorIs(t, err, ErrNotificationTemplateRuleUnsupportedEvent)

	_, err = ResolveTransactionalNotificationTemplate(outbox.EventTypeAfterSalesStatusChanged, []byte("{}"))
	assert.Error(t, err)
}

package service

import (
	"testing"

	"commerce-platform/internal/domain/outbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionalNotificationTemplateDefinitionsCoverAllCustomerVisibleRules(t *testing.T) {
	for _, eventType := range []string{
		outbox.EventTypeOrderPaymentSucceeded,
		outbox.EventTypeOrderPaymentExpired,
		outbox.EventTypeOrderCancelled,
		outbox.EventTypeOrderShipped,
		outbox.EventTypeOrderDelivered,
		outbox.EventTypeOrderCompleted,
		outbox.EventTypeOrderRefunded,
	} {
		resolution, err := ResolveTransactionalNotificationTemplate(eventType, nil)
		require.NoError(t, err, eventType)
		definition, err := LookupTransactionalNotificationTemplate(resolution.TemplateCode)
		require.NoError(t, err, eventType)
		assert.Equal(t, resolution.RequiredVariables, definition.RequiredVariables, eventType)
		assert.NotEmpty(t, definition.Category, eventType)
	}
}

func TestResolveTransactionalNotificationTemplateDefinitionFallsBackToEnglish(t *testing.T) {
	definition, err := ResolveTransactionalNotificationTemplateDefinition(
		NotificationTemplateOrderConfirmation,
		"fr-CA",
		[]string{"en", "zh_cn"},
	)
	require.NoError(t, err)
	assert.Equal(t, "en", definition.Locale)

	definition, err = ResolveTransactionalNotificationTemplateDefinition(
		NotificationTemplateOrderConfirmation,
		"zh-CN",
		[]string{"en", "zh_cn"},
	)
	require.NoError(t, err)
	assert.Equal(t, "zh_cn", definition.Locale)
}

func TestValidateTransactionalNotificationVariablesRejectsMissingAndUnknownValues(t *testing.T) {
	err := ValidateTransactionalNotificationVariables(NotificationTemplateOrderConfirmation, map[string]string{
		"order_number": "ORD-1",
		"order_amount": "$10.00",
	})
	assert.ErrorIs(t, err, ErrNotificationTemplateVariableMissing)

	err = ValidateTransactionalNotificationVariables(NotificationTemplateOrderConfirmation, map[string]string{
		"order_number": "ORD-1",
		"order_amount": "$10.00",
		"paid_at":      "2026-09-20T00:00:00Z",
		"internal_id":  "secret",
	})
	assert.ErrorIs(t, err, ErrNotificationTemplateVariableUnknown)

	require.NoError(t, ValidateTransactionalNotificationVariables(NotificationTemplateOrderConfirmation, map[string]string{
		"order_number": "ORD-1",
		"order_amount": "$10.00",
		"paid_at":      "2026-09-20T00:00:00Z",
	}))
}

func TestResolveTransactionalNotificationTemplateDefinitionRequiresEnglishFallback(t *testing.T) {
	_, err := ResolveTransactionalNotificationTemplateDefinition(
		NotificationTemplateOrderConfirmation,
		"fr",
		[]string{"de"},
	)
	assert.ErrorIs(t, err, ErrNotificationTemplateLocaleUnavailable)
}

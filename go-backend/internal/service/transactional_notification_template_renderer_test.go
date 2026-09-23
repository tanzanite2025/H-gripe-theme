package service

import (
	"testing"

	"commerce-platform/internal/domain/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func renderableOrderConfirmationTemplate() *notification.EmailTemplate {
	return &notification.EmailTemplate{
		Code:              NotificationTemplateOrderConfirmation,
		Locale:            "en",
		Category:          notification.TemplateCategoryOrder,
		Name:              "Order confirmation",
		SubjectTemplate:   "Order {{order_number}}",
		BodyHTML:          `<p>Hello {{customer_name}}</p><p>Order {{order_number}}</p>`,
		BodyText:          "Hello {{customer_name}}\nOrder {{order_number}}",
		AllowedVariables:  datatypes.JSON(`["order_number","order_amount","paid_at","customer_name","currency","items"]`),
		RequiredVariables: datatypes.JSON(`["order_number","order_amount","paid_at"]`),
		IsEnabled:         true,
		Version:           1,
	}
}

func TestRenderTransactionalNotificationTemplateEscapesHTMLValues(t *testing.T) {
	rendered, err := RenderTransactionalNotificationTemplate(renderableOrderConfirmationTemplate(), map[string]string{
		"order_number":  "ORD-1",
		"order_amount":  "$10.00",
		"paid_at":       "2026-09-20",
		"customer_name": "<Admin>",
	})
	require.NoError(t, err)
	assert.Equal(t, "Order ORD-1", rendered.Subject)
	assert.Contains(t, rendered.HTML, "&lt;Admin&gt;")
	assert.Contains(t, rendered.Text, "<Admin>")
}

func TestRenderTransactionalNotificationTemplateRejectsUnknownPlaceholderAndSubjectInjection(t *testing.T) {
	templateRecord := renderableOrderConfirmationTemplate()
	templateRecord.BodyText = "{{unknown_field}}"
	_, err := RenderTransactionalNotificationTemplate(templateRecord, map[string]string{
		"order_number":  "ORD-1",
		"order_amount":  "$10.00",
		"paid_at":       "2026-09-20",
		"customer_name": "Customer",
	})
	assert.ErrorIs(t, err, ErrNotificationTemplateVariableMissing)

	templateRecord = renderableOrderConfirmationTemplate()
	templateRecord.SubjectTemplate = "Order\r\nBcc: attacker@example.com"
	_, err = RenderTransactionalNotificationTemplate(templateRecord, map[string]string{
		"order_number":  "ORD-1",
		"order_amount":  "$10.00",
		"paid_at":       "2026-09-20",
		"customer_name": "Customer",
	})
	assert.ErrorIs(t, err, ErrNotificationTemplateSubjectUnsafe)
}

func TestRenderTransactionalNotificationTemplateRejectsMalformedPlaceholder(t *testing.T) {
	templateRecord := renderableOrderConfirmationTemplate()
	templateRecord.BodyText = "Hello {{order_number"
	_, err := RenderTransactionalNotificationTemplate(templateRecord, map[string]string{
		"order_number":  "ORD-1",
		"order_amount":  "$10.00",
		"paid_at":       "2026-09-20",
		"customer_name": "Customer",
	})
	assert.ErrorIs(t, err, ErrNotificationTemplatePlaceholderInvalid)
}

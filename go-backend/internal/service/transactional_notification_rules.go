package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	afterSalesDomain "commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/outbox"
)

// Template codes are stable notification API identifiers. They are deliberately
// independent from database records, locale labels, and SMTP configuration.
const (
	NotificationTemplateOrderConfirmation         = "order_confirmation"
	NotificationTemplateOrderPaymentExpired       = "order_payment_expired"
	NotificationTemplateOrderCancelled            = "order_cancelled"
	NotificationTemplateOrderShippingNotification = "order_shipping_notification"
	NotificationTemplateOrderDelivered            = "order_delivered"
	NotificationTemplateOrderCompleted            = "order_completed"
	NotificationTemplateOrderRefunded             = "order_refunded"

	NotificationTemplateAfterSalesRequested       = "after_sales_requested"
	NotificationTemplateAfterSalesApproved        = "after_sales_approved"
	NotificationTemplateAfterSalesAwaitingReturn  = "after_sales_awaiting_return"
	NotificationTemplateAfterSalesReturnInTransit = "after_sales_return_in_transit"
	NotificationTemplateAfterSalesReceived        = "after_sales_received"
	NotificationTemplateAfterSalesResolving       = "after_sales_resolving"
	NotificationTemplateAfterSalesCompleted       = "after_sales_completed"
	NotificationTemplateAfterSalesRejected        = "after_sales_rejected"
)

var ErrNotificationTemplateRuleUnsupportedEvent = errors.New("notification template rule does not support this event")

const TransactionalNotificationSettingsGroup = "transactional_notifications"

// NotificationRuleToggleKey returns the admin setting key for a stable
// template code. Missing settings intentionally mean enabled for backwards
// compatibility; operators can disable a noisy status without changing the
// canonical event contract.
func NotificationRuleToggleKey(templateCode string) string {
	return "notification_" + strings.TrimSpace(templateCode) + "_enabled"
}

// NotificationTemplateResolution is the deterministic result of applying the
// event-to-template rule. An empty TemplateCode means that the fact is valid
// but intentionally has no automatic customer email (for example an
// after-sales exception or cancellation).
type NotificationTemplateResolution struct {
	EventType         string
	TemplateCode      string
	CustomerVisible   bool
	RequiredVariables []string
}

// ResolveTransactionalNotificationTemplate maps canonical domain facts to
// stable template codes. It performs no template lookup, rendering, SMTP I/O,
// or database access, so the rule can be tested and reused by a later worker.
func ResolveTransactionalNotificationTemplate(eventType string, payload []byte) (NotificationTemplateResolution, error) {
	eventType = strings.TrimSpace(eventType)
	resolution := NotificationTemplateResolution{EventType: eventType}
	switch eventType {
	case outbox.EventTypeOrderPaymentSucceeded:
		return withNotificationTemplate(resolution, NotificationTemplateOrderConfirmation, "order_number", "order_amount", "paid_at"), nil
	case outbox.EventTypeOrderPaymentExpired:
		return withNotificationTemplate(resolution, NotificationTemplateOrderPaymentExpired, "order_number"), nil
	case outbox.EventTypeOrderCancelled:
		return withNotificationTemplate(resolution, NotificationTemplateOrderCancelled, "order_number"), nil
	case outbox.EventTypeOrderShipped:
		return withNotificationTemplate(resolution, NotificationTemplateOrderShippingNotification, "order_number", "carrier_name", "tracking_number", "tracking_url"), nil
	case outbox.EventTypeOrderDelivered:
		return withNotificationTemplate(resolution, NotificationTemplateOrderDelivered, "order_number", "delivered_at"), nil
	case outbox.EventTypeOrderCompleted:
		return withNotificationTemplate(resolution, NotificationTemplateOrderCompleted, "order_number", "completed_at"), nil
	case outbox.EventTypeOrderRefunded:
		return withNotificationTemplate(resolution, NotificationTemplateOrderRefunded, "order_number", "refund_amount", "currency"), nil
	case outbox.EventTypeAfterSalesStatusChanged:
		return resolveAfterSalesNotificationTemplate(resolution, payload)
	default:
		return resolution, fmt.Errorf("%w: %s", ErrNotificationTemplateRuleUnsupportedEvent, eventType)
	}
}

func withNotificationTemplate(
	resolution NotificationTemplateResolution,
	templateCode string,
	requiredVariables ...string,
) NotificationTemplateResolution {
	resolution.TemplateCode = templateCode
	resolution.CustomerVisible = true
	resolution.RequiredVariables = append([]string(nil), requiredVariables...)
	return resolution
}

func resolveAfterSalesNotificationTemplate(
	resolution NotificationTemplateResolution,
	payload []byte,
) (NotificationTemplateResolution, error) {
	var eventPayload outbox.AfterSalesStatusChangedPayload
	if err := json.Unmarshal(payload, &eventPayload); err != nil {
		return resolution, fmt.Errorf("decode after-sales notification rule payload: %w", err)
	}
	if eventPayload.CaseID == 0 || strings.TrimSpace(eventPayload.NewStatus) == "" {
		return resolution, errors.New("after-sales notification rule payload is incomplete")
	}

	switch strings.TrimSpace(eventPayload.NewStatus) {
	case afterSalesDomain.StatusRequested:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesRequested, "after_sales_case_number", "order_number", "after_sales_type", "after_sales_status"), nil
	case afterSalesDomain.StatusApproved:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesApproved, "after_sales_case_number", "after_sales_type", "after_sales_status"), nil
	case afterSalesDomain.StatusAwaitingReturn:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesAwaitingReturn, "after_sales_case_number", "return_warehouse_name", "return_warehouse_address"), nil
	case afterSalesDomain.StatusReturnInTransit:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesReturnInTransit, "after_sales_case_number", "carrier_name", "tracking_number", "tracking_url"), nil
	case afterSalesDomain.StatusReceived:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesReceived, "after_sales_case_number", "received_at"), nil
	case afterSalesDomain.StatusResolving:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesResolving, "after_sales_case_number", "after_sales_status"), nil
	case afterSalesDomain.StatusCompleted:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesCompleted, "after_sales_case_number", "after_sales_status", "refund_amount"), nil
	case afterSalesDomain.StatusRejected:
		return withNotificationTemplate(resolution, NotificationTemplateAfterSalesRejected, "after_sales_case_number", "rejection_reason"), nil
	default:
		// Valid state facts such as inspecting, exception, and cancelled are
		// intentionally auditable without automatically emailing the customer.
		return resolution, nil
	}
}

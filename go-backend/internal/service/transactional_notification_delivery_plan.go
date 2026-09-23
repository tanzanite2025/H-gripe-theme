package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/locales"
)

var (
	ErrNotificationDeliveryRecipientRequired = errors.New("transactional notification recipient is required")
	ErrNotificationDeliveryEventIncomplete   = errors.New("transactional notification event is incomplete")
	ErrNotificationDeliveryContextConflict   = errors.New("transactional notification context conflicts with event facts")
)

// TransactionalNotificationDeliveryContext contains the customer-facing
// context that is intentionally not owned by the template registry. The
// recipient and locale should come from the immutable order/customer snapshot
// when a future delivery worker claims the event. Variables may carry
// snapshot-only values such as rendered line items or a refund amount.
//
// No SMTP/provider information belongs here. Keeping this boundary explicit
// prevents a provider configuration from becoming part of the domain event
// contract.
type TransactionalNotificationDeliveryContext struct {
	RecipientEmail string
	Locale         string
	CustomerName   string
	Variables      map[string]string
}

// TransactionalNotificationDeliveryPlan is the provider-neutral handoff
// between a canonical domain fact and a future email/notification worker.
// Building a plan performs no database lookup, template rendering, or network
// I/O. A silent fact returns a plan with an empty TemplateCode and should be
// acknowledged without creating a customer message.
type TransactionalNotificationDeliveryPlan struct {
	SourceEventID  uint
	SourceEventKey string
	EventType      string
	IdempotencyKey string
	TemplateCode   string
	RecipientEmail string
	Locale         string
	Variables      map[string]string
}

// BuildTransactionalNotificationDeliveryPlan translates one canonical event
// into a deterministic notification plan. The event remains the source of
// truth for status/timestamp/amount values; context only supplies recipient
// identity and values that are not part of the shared fact schema yet.
func BuildTransactionalNotificationDeliveryPlan(
	event outbox.Event,
	deliveryContext TransactionalNotificationDeliveryContext,
) (TransactionalNotificationDeliveryPlan, error) {
	event.EventType = strings.TrimSpace(event.EventType)
	resolution, err := ResolveTransactionalNotificationTemplate(event.EventType, event.Payload)
	if err != nil {
		return TransactionalNotificationDeliveryPlan{}, err
	}
	audience := notificationAudienceFromEvent(event)
	recipientEmail := strings.TrimSpace(deliveryContext.RecipientEmail)
	if recipientEmail == "" {
		recipientEmail = strings.TrimSpace(audience.RecipientEmail)
	} else if strings.TrimSpace(audience.RecipientEmail) != "" && recipientEmail != strings.TrimSpace(audience.RecipientEmail) {
		return TransactionalNotificationDeliveryPlan{}, fmt.Errorf("%w: recipient_email", ErrNotificationDeliveryContextConflict)
	}
	requestedLocale := strings.TrimSpace(deliveryContext.Locale)
	if requestedLocale == "" {
		requestedLocale = strings.TrimSpace(audience.Locale)
	}
	customerName := strings.TrimSpace(deliveryContext.CustomerName)
	if customerName == "" {
		customerName = strings.TrimSpace(audience.CustomerName)
	}

	plan := TransactionalNotificationDeliveryPlan{
		SourceEventID:  event.ID,
		SourceEventKey: strings.TrimSpace(event.EventKey),
		EventType:      strings.TrimSpace(event.EventType),
		TemplateCode:   strings.TrimSpace(resolution.TemplateCode),
		Locale:         normalizeNotificationLocale(requestedLocale),
		Variables:      map[string]string{},
	}
	if plan.TemplateCode == "" {
		return plan, nil
	}

	if recipientEmail == "" {
		return TransactionalNotificationDeliveryPlan{}, ErrNotificationDeliveryRecipientRequired
	}
	recipient, err := validateNotificationPlanRecipient(recipientEmail)
	if err != nil {
		return TransactionalNotificationDeliveryPlan{}, err
	}
	plan.RecipientEmail = recipient

	variables, err := notificationPlanVariables(event)
	if err != nil {
		return TransactionalNotificationDeliveryPlan{}, err
	}
	// The shared payload parser intentionally exposes every useful snapshot
	// field (for example a shipment carrier on an after-sales event). Only the
	// variables allowed by the selected template may cross the rendering
	// boundary; fields for a different status are discarded before caller
	// supplied overrides are validated.
	definition, err := LookupTransactionalNotificationTemplate(plan.TemplateCode)
	if err != nil {
		return TransactionalNotificationDeliveryPlan{}, err
	}
	allowed := make(map[string]struct{}, len(definition.AllowedVariables))
	for _, variable := range definition.AllowedVariables {
		allowed[variable] = struct{}{}
	}
	for key := range variables {
		if _, ok := allowed[key]; !ok {
			delete(variables, key)
		}
	}
	if customerName != "" {
		variables["customer_name"] = customerName
	}
	for key, value := range deliveryContext.Variables {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if existing, exists := variables[key]; exists && !isNotificationSnapshotOverrideAllowed(event, key, existing) {
			return TransactionalNotificationDeliveryPlan{}, fmt.Errorf("%w: %s", ErrNotificationDeliveryContextConflict, key)
		}
		variables[key] = strings.TrimSpace(value)
	}
	if err := ValidateTransactionalNotificationVariables(plan.TemplateCode, variables); err != nil {
		return TransactionalNotificationDeliveryPlan{}, err
	}

	plan.IdempotencyKey = canonicalNotificationIdempotencyKey(event, plan.TemplateCode)
	plan.Variables = variables
	return plan, nil
}

func notificationAudienceFromEvent(event outbox.Event) outbox.NotificationAudienceSnapshot {
	var audience outbox.NotificationAudienceSnapshot
	if len(event.Payload) == 0 {
		return audience
	}
	// All canonical customer-visible payloads embed the same JSON fields, so a
	// small structural decode keeps this helper independent from each event's
	// business fields and remains compatible with legacy payloads.
	_ = json.Unmarshal(event.Payload, &audience)
	return audience
}

func isNotificationSnapshotOverrideAllowed(event outbox.Event, key, existing string) bool {
	// The current after-sales event schema does not yet carry the monetary
	// resolution snapshot. Until it does, the explicit zero placeholder may be
	// replaced by the transaction's immutable refund amount. Every other value
	// derived from a domain fact remains authoritative and cannot be rewritten
	// by delivery context.
	return event.EventType == outbox.EventTypeAfterSalesStatusChanged &&
		key == "refund_amount" && existing == "0"
}

func validateNotificationPlanRecipient(value string) (string, error) {
	value = strings.TrimSpace(value)
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", fmt.Errorf("validate transactional notification recipient: %w", err)
	}
	return parsed.Address, nil
}

func normalizeNotificationLocale(value string) string {
	if resolved := locales.ResolveSupported(value); resolved != "" {
		return resolved
	}
	return "en"
}

func canonicalNotificationIdempotencyKey(event outbox.Event, templateCode string) string {
	var envelope struct {
		IdempotencyKey string `json:"idempotency_key"`
	}
	if len(event.Payload) > 0 && json.Unmarshal(event.Payload, &envelope) == nil && strings.TrimSpace(envelope.IdempotencyKey) != "" {
		return fmt.Sprintf("notification:%s:%s", strings.TrimSpace(envelope.IdempotencyKey), templateCode)
	}
	if strings.TrimSpace(event.EventKey) != "" {
		return fmt.Sprintf("notification:%s:%s", strings.TrimSpace(event.EventKey), templateCode)
	}
	return fmt.Sprintf("notification:event:%d:%s", event.ID, templateCode)
}

func notificationPlanVariables(event outbox.Event) (map[string]string, error) {
	variables := make(map[string]string)
	switch event.EventType {
	case outbox.EventTypeOrderPaymentSucceeded:
		var payload outbox.OrderPaymentSucceededPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["order_amount"] = notificationAmount(payload.AmountMinor, payload.Currency)
		variables["currency"] = strings.TrimSpace(payload.Currency)
		variables["paid_at"] = notificationTime(payload.OccurredAt)
	case outbox.EventTypeOrderPaymentExpired:
		var payload outbox.OrderPaymentExpiredPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["expired_at"] = notificationTime(payload.OccurredAt)
	case outbox.EventTypeOrderCancelled:
		var payload outbox.OrderCancelledPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["cancelled_at"] = notificationTime(payload.OccurredAt)
	case outbox.EventTypeOrderShipped:
		var payload outbox.OrderShippedPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" || strings.TrimSpace(payload.CarrierName) == "" || strings.TrimSpace(payload.TrackingNumber) == "" || strings.TrimSpace(payload.TrackingURL) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["carrier_name"] = strings.TrimSpace(payload.CarrierName)
		variables["tracking_number"] = strings.TrimSpace(payload.TrackingNumber)
		variables["tracking_url"] = strings.TrimSpace(payload.TrackingURL)
		variables["shipped_at"] = notificationTime(payload.ShippedAt)
	case outbox.EventTypeOrderDelivered:
		var payload outbox.OrderDeliveredPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["delivered_at"] = notificationTime(payload.DeliveredAt)
		variables["tracking_number"] = strings.TrimSpace(payload.TrackingNumber)
	case outbox.EventTypeOrderCompleted:
		var payload outbox.OrderCompletedPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["completed_at"] = notificationTime(payload.CompletedAt)
	case outbox.EventTypeOrderRefunded:
		var payload outbox.OrderRefundedPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" || payload.RefundID == 0 {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["refund_amount"] = notificationAmount(payload.AmountMinor, payload.Currency)
		variables["currency"] = strings.TrimSpace(payload.Currency)
		variables["refunded_at"] = notificationTime(payload.OccurredAt)
	case outbox.EventTypeAfterSalesStatusChanged:
		var payload outbox.AfterSalesStatusChangedPayload
		if err := decodeNotificationPlanPayload(event, &payload); err != nil {
			return nil, err
		}
		if payload.CaseID == 0 || strings.TrimSpace(payload.NewStatus) == "" {
			return nil, ErrNotificationDeliveryEventIncomplete
		}
		variables["after_sales_case_number"] = strconv.FormatUint(uint64(payload.CaseID), 10)
		variables["order_number"] = strings.TrimSpace(payload.OrderNumber)
		variables["after_sales_type"] = strings.TrimSpace(payload.CaseType)
		variables["after_sales_status"] = strings.TrimSpace(payload.NewStatus)
		variables["requested_at"] = notificationTime(payload.OccurredAt)
		variables["approved_at"] = notificationTime(payload.OccurredAt)
		variables["return_warehouse_name"] = strings.TrimSpace(payload.WarehouseName)
		variables["return_warehouse_address"] = strings.TrimSpace(payload.WarehouseAddress)
		variables["carrier_name"] = strings.TrimSpace(payload.Carrier)
		variables["tracking_number"] = strings.TrimSpace(payload.TrackingNumber)
		variables["tracking_url"] = strings.TrimSpace(payload.TrackingURL)
		variables["return_label_url"] = strings.TrimSpace(payload.LabelURL)
		variables["shipped_at"] = notificationTime(pointerTime(payload.ShippedAt))
		variables["received_at"] = notificationTime(pointerTime(payload.ReceivedAt))
		variables["resolution_note"] = strings.TrimSpace(payload.Resolution)
		variables["resolution"] = strings.TrimSpace(payload.Resolution)
		variables["rejection_reason"] = strings.TrimSpace(payload.Resolution)
		// A completed case may be an exchange/reshipment with no refund. The
		// amount is supplied by the future case snapshot when it exists; zero is
		// still an explicit, renderable value for non-refund resolutions.
		variables["refund_amount"] = "0"
	default:
		return nil, fmt.Errorf("%w: unsupported event %s", ErrNotificationDeliveryEventIncomplete, event.EventType)
	}
	return variables, nil
}

func decodeNotificationPlanPayload(event outbox.Event, target interface{}) error {
	if len(event.Payload) == 0 {
		return ErrNotificationDeliveryEventIncomplete
	}
	if err := json.Unmarshal(event.Payload, target); err != nil {
		return fmt.Errorf("decode transactional notification event payload: %w", err)
	}
	return nil
}

func notificationAmount(minor int64, currency string) string {
	currency = strings.TrimSpace(currency)
	if money, err := domainmoney.New(minor, currency); err == nil {
		if formatted, err := money.FormatMajor(); err == nil {
			return formatted
		}
	}
	return strconv.FormatInt(minor, 10)
}

func notificationTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func pointerTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

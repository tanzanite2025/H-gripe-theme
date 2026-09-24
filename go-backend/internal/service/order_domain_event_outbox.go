package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	afterSalesDomain "commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	paymentDomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

const canonicalDomainEventSchemaVersion = 1

// CanonicalDomainEventOutboxHandler validates persisted business facts and,
// when configured with a sender, executes the provider-neutral notification
// plan/render/send path. Malformed facts and delivery failures are returned to
// the Outbox worker so they remain observable and retryable.
type CanonicalDomainEventOutboxHandler struct {
	templateService *TransactionalNotificationTemplateService
	sender          TransactionalNotificationSender
	settings        *SettingService
}

func NewCanonicalDomainEventOutboxHandler(templateServices ...*TransactionalNotificationTemplateService) *CanonicalDomainEventOutboxHandler {
	var templateService *TransactionalNotificationTemplateService
	if len(templateServices) > 0 {
		templateService = templateServices[0]
	}
	return &CanonicalDomainEventOutboxHandler{templateService: templateService}
}

// NewCanonicalDomainEventOutboxHandlerWithSender enables the real delivery
// path while keeping the original constructor available to validation-only
// callers and contract tests.
func NewCanonicalDomainEventOutboxHandlerWithSender(
	templateService *TransactionalNotificationTemplateService,
	sender TransactionalNotificationSender,
) *CanonicalDomainEventOutboxHandler {
	return &CanonicalDomainEventOutboxHandler{
		templateService: templateService,
		sender:          sender,
	}
}

func (h *CanonicalDomainEventOutboxHandler) ConfigureNotificationSettings(settings *SettingService) {
	if h != nil {
		h.settings = settings
	}
}

func (h *CanonicalDomainEventOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	event.EventType = strings.TrimSpace(event.EventType)
	if !isCanonicalDomainEventType(event.EventType) {
		return fmt.Errorf("unsupported canonical domain event type %s", event.EventType)
	}
	var envelope struct {
		SchemaVersion  int    `json:"schema_version"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := json.Unmarshal(event.Payload, &envelope); err != nil {
		return fmt.Errorf("decode canonical domain event envelope: %w", err)
	}
	if envelope.SchemaVersion <= 0 || strings.TrimSpace(envelope.IdempotencyKey) == "" {
		return errors.New("canonical domain event envelope is incomplete")
	}
	if err := validateCanonicalDomainEventPayload(event.EventType, event.Payload); err != nil {
		return err
	}
	resolution, err := ResolveTransactionalNotificationTemplate(event.EventType, event.Payload)
	if err != nil {
		return err
	}
	if err := ValidateTransactionalNotificationTemplateResolution(resolution); err != nil {
		return err
	}
	if resolution.TemplateCode != "" && h.settings != nil {
		enabled, err := h.settings.IsTransactionalNotificationEnabled(resolution.TemplateCode)
		if err != nil {
			return fmt.Errorf("read transactional notification rule %q: %w", resolution.TemplateCode, err)
		}
		if !enabled {
			return nil
		}
	}
	if h.sender != nil {
		if err := deliverTransactionalNotification(ctx, event, h.templateService, h.sender); err != nil {
			return err
		}
		return nil
	}
	if h.templateService != nil && resolution.TemplateCode != "" {
		if _, err := h.templateService.Resolve(resolution.TemplateCode, "en"); err != nil {
			return fmt.Errorf("resolve transactional notification template %q: %w", resolution.TemplateCode, err)
		}
	}
	return nil
}

// validateCanonicalDomainEventPayload checks only the minimum identity facts
// required to safely acknowledge a canonical event. It deliberately does not
// require a recipient, locale, or a configured template because those belong
// to the later notification-delivery plan and may be absent for silent facts.
func validateCanonicalDomainEventPayload(eventType string, payload []byte) error {
	if len(payload) == 0 {
		return errors.New("canonical domain event payload is empty")
	}
	missing := func(fields ...string) error {
		return fmt.Errorf("canonical domain event %s payload is incomplete: missing %s", eventType, strings.Join(fields, ", "))
	}
	switch eventType {
	case outbox.EventTypeOrderPaymentSucceeded:
		var value outbox.OrderPaymentSucceededPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.OrderID == 0 || strings.TrimSpace(value.OrderNumber) == "" || strings.TrimSpace(value.PaymentTransactionID) == "" {
			return missing("order_id", "order_number", "payment_transaction_id")
		}
	case outbox.EventTypeOrderPaymentExpired:
		var value outbox.OrderPaymentExpiredPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.OrderID == 0 || strings.TrimSpace(value.OrderNumber) == "" {
			return missing("order_id", "order_number")
		}
	case outbox.EventTypeOrderCancelled:
		var value outbox.OrderCancelledPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.OrderID == 0 || strings.TrimSpace(value.OrderNumber) == "" {
			return missing("order_id", "order_number")
		}
	case outbox.EventTypeOrderShipped:
		var value outbox.OrderShippedPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.OrderID == 0 || strings.TrimSpace(value.OrderNumber) == "" || strings.TrimSpace(value.TrackingNumber) == "" {
			return missing("order_id", "order_number", "tracking_number")
		}
	case outbox.EventTypeOrderDelivered:
		var value outbox.OrderDeliveredPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.OrderID == 0 || strings.TrimSpace(value.OrderNumber) == "" {
			return missing("order_id", "order_number")
		}
	case outbox.EventTypeOrderRefunded:
		var value outbox.OrderRefundedPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.OrderID == 0 || value.RefundID == 0 {
			return missing("order_id", "refund_id")
		}
	case outbox.EventTypeAfterSalesStatusChanged:
		var value outbox.AfterSalesStatusChangedPayload
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("decode %s payload: %w", eventType, err)
		}
		if value.CaseID == 0 || value.TransitionID == 0 || strings.TrimSpace(value.NewStatus) == "" {
			return missing("case_id", "transition_id", "new_status")
		}
	default:
		return fmt.Errorf("unsupported canonical domain event type %s", eventType)
	}
	return nil
}

func isCanonicalDomainEventType(eventType string) bool {
	switch eventType {
	case outbox.EventTypeOrderPaymentSucceeded,
		outbox.EventTypeOrderPaymentExpired,
		outbox.EventTypeOrderCancelled,
		outbox.EventTypeOrderShipped,
		outbox.EventTypeOrderDelivered,
		outbox.EventTypeOrderRefunded,
		outbox.EventTypeAfterSalesStatusChanged:
		return true
	default:
		return false
	}
}

// createCanonicalDomainEvent persists a business fact in the same transaction
// as the aggregate mutation. The event key is the durable idempotency key;
// callers should derive it from a payment, shipment, refund, or transition ID
// rather than from a wall-clock timestamp.
func createCanonicalDomainEvent(
	repo *repository.OutboxRepository,
	eventType string,
	eventKey string,
	aggregateType string,
	aggregateID string,
	payload interface{},
	occurredAt time.Time,
) error {
	if repo == nil {
		return nil
	}
	if strings.TrimSpace(eventType) == "" || strings.TrimSpace(eventKey) == "" {
		return errors.New("canonical domain event type and key are required")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode canonical domain event %s: %w", eventType, err)
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	} else {
		occurredAt = occurredAt.UTC()
	}
	return repo.CreateEvent(&outbox.Event{
		EventKey:      strings.TrimSpace(eventKey),
		EventType:     strings.TrimSpace(eventType),
		AggregateType: strings.TrimSpace(aggregateType),
		AggregateID:   strings.TrimSpace(aggregateID),
		Payload:       datatypes.JSON(encoded),
		AvailableAt:   occurredAt,
	})
}

func canonicalDomainEventMetadata(idempotencyKey string, occurredAt time.Time) outbox.CanonicalDomainEventMetadata {
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	} else {
		occurredAt = occurredAt.UTC()
	}
	return outbox.CanonicalDomainEventMetadata{
		SchemaVersion:  canonicalDomainEventSchemaVersion,
		OccurredAt:     occurredAt,
		IdempotencyKey: strings.TrimSpace(idempotencyKey),
	}
}

func orderNotificationAudience(orderRecord *order.Order) outbox.NotificationAudienceSnapshot {
	if orderRecord == nil {
		return outbox.NotificationAudienceSnapshot{}
	}
	return outbox.NotificationAudienceSnapshot{
		RecipientEmail: strings.TrimSpace(orderRecord.ShippingAddress.Email),
		CustomerName:   orderCustomerName(orderRecord),
	}
}

func enqueueOrderPaymentSucceededDomainEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	input VerifiedGatewayPaymentInput,
	previousOrderStatus string,
	newOrderStatus string,
	occurredAt time.Time,
) error {
	if orderRecord == nil || orderRecord.ID == 0 {
		return errors.New("order payment succeeded event requires a persisted order")
	}
	transactionID := strings.TrimSpace(input.TransactionID)
	if transactionID == "" {
		return errors.New("order payment succeeded event requires a transaction ID")
	}
	idempotencyKey := fmt.Sprintf("payment_transaction:%s", transactionID)
	payload := outbox.OrderPaymentSucceededPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, occurredAt),
		NotificationAudienceSnapshot: orderNotificationAudience(orderRecord),
		OrderID:                      orderRecord.ID,
		OrderNumber:                  strings.TrimSpace(orderRecord.OrderNumber),
		PaymentTransactionID:         transactionID,
		PreviousOrderStatus:          strings.TrimSpace(previousOrderStatus),
		NewOrderStatus:               strings.TrimSpace(newOrderStatus),
		PreviousPaymentStatus:        "unpaid",
		NewPaymentStatus:             "paid",
		AmountMinor:                  input.Amount.AmountMinor(),
		Currency:                     strings.TrimSpace(input.Amount.Currency().String()),
		PaymentMethod:                strings.TrimSpace(input.PaymentMethod),
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeOrderPaymentSucceeded,
		fmt.Sprintf("%s:%d:%s", outbox.EventTypeOrderPaymentSucceeded, orderRecord.ID, transactionID),
		outbox.AggregateTypeOrder,
		strconv.FormatUint(uint64(orderRecord.ID), 10),
		payload,
		occurredAt,
	)
}

func enqueueOrderPaymentExpiredDomainEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	occurredAt time.Time,
) error {
	if orderRecord == nil || orderRecord.ID == 0 {
		return errors.New("order payment expired event requires a persisted order")
	}
	idempotencyKey := fmt.Sprintf("order_expiration:%d", orderRecord.ID)
	payload := outbox.OrderPaymentExpiredPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, occurredAt),
		NotificationAudienceSnapshot: orderNotificationAudience(orderRecord),
		OrderID:                      orderRecord.ID,
		OrderNumber:                  strings.TrimSpace(orderRecord.OrderNumber),
		PreviousOrderStatus:          strings.TrimSpace(orderRecord.Status),
		NewOrderStatus:               "payment_expired",
		PreviousPaymentStatus:        strings.TrimSpace(orderRecord.PaymentStatus),
		NewPaymentStatus:             "expired",
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeOrderPaymentExpired,
		fmt.Sprintf("%s:%d", outbox.EventTypeOrderPaymentExpired, orderRecord.ID),
		outbox.AggregateTypeOrder,
		strconv.FormatUint(uint64(orderRecord.ID), 10),
		payload,
		occurredAt,
	)
}

func enqueueOrderCancelledDomainEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	occurredAt time.Time,
) error {
	if orderRecord == nil || orderRecord.ID == 0 {
		return errors.New("order cancelled event requires a persisted order")
	}
	idempotencyKey := fmt.Sprintf("order_cancellation:%d", orderRecord.ID)
	payload := outbox.OrderCancelledPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, occurredAt),
		NotificationAudienceSnapshot: orderNotificationAudience(orderRecord),
		OrderID:                      orderRecord.ID,
		OrderNumber:                  strings.TrimSpace(orderRecord.OrderNumber),
		PreviousOrderStatus:          strings.TrimSpace(orderRecord.Status),
		NewOrderStatus:               "cancelled",
		PaymentStatus:                strings.TrimSpace(orderRecord.PaymentStatus),
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeOrderCancelled,
		fmt.Sprintf("%s:%d", outbox.EventTypeOrderCancelled, orderRecord.ID),
		outbox.AggregateTypeOrder,
		strconv.FormatUint(uint64(orderRecord.ID), 10),
		payload,
		occurredAt,
	)
}

func enqueueOrderShippedDomainEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	tracking *resolvedOrderTrackingUpdate,
	previousOrderStatus string,
	previousShippingStatus string,
	shippedAt time.Time,
) error {
	if orderRecord == nil || orderRecord.ID == 0 || tracking == nil {
		return errors.New("order shipped event requires an order and tracking update")
	}
	trackingNumber := strings.TrimSpace(tracking.trackingShipment.TrackingNumber)
	if trackingNumber == "" {
		return errors.New("order shipped event requires a tracking number")
	}
	shipmentID := tracking.trackingShipment.ID
	idempotencyKey := fmt.Sprintf("shipment:%d", shipmentID)
	if shipmentID == 0 {
		// This is only a defensive fallback for non-persisted test fixtures. A
		// committed fulfillment row always receives an ID before this helper runs.
		idempotencyKey = fmt.Sprintf("shipment:%d:%s", orderRecord.ID, trackingNumber)
	}
	payload := outbox.OrderShippedPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, shippedAt),
		NotificationAudienceSnapshot: orderNotificationAudience(orderRecord),
		OrderID:                      orderRecord.ID,
		OrderNumber:                  strings.TrimSpace(orderRecord.OrderNumber),
		ShipmentID:                   shipmentID,
		CarrierName:                  strings.TrimSpace(tracking.carrierName),
		TrackingNumber:               trackingNumber,
		TrackingURL:                  resolveOrderTrackingURL(tracking.trackingURLTemplate, trackingNumber),
		PreviousOrderStatus:          strings.TrimSpace(previousOrderStatus),
		NewOrderStatus:               "shipped",
		PreviousShippingStatus:       strings.TrimSpace(previousShippingStatus),
		NewShippingStatus:            "shipped",
		ShippedAt:                    shippedAt.UTC(),
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeOrderShipped,
		fmt.Sprintf("%s:%s", outbox.EventTypeOrderShipped, idempotencyKey),
		outbox.AggregateTypeOrder,
		strconv.FormatUint(uint64(orderRecord.ID), 10),
		payload,
		shippedAt,
	)
}

func enqueueOrderDeliveredDomainEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	trackingNumber string,
	previousShippingStatus string,
	deliveredAt time.Time,
	source string,
) error {
	if orderRecord == nil || orderRecord.ID == 0 {
		return errors.New("order delivered event requires a persisted order")
	}
	trackingNumber = strings.TrimSpace(trackingNumber)
	idempotencyKey := fmt.Sprintf("delivery_fact:%d:%s", orderRecord.ID, trackingNumber)
	payload := outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, deliveredAt),
		NotificationAudienceSnapshot: orderNotificationAudience(orderRecord),
		OrderID:                      orderRecord.ID,
		OrderNumber:                  strings.TrimSpace(orderRecord.OrderNumber),
		TrackingNumber:               trackingNumber,
		PreviousShippingStatus:       strings.TrimSpace(previousShippingStatus),
		NewShippingStatus:            "delivered",
		DeliveredAt:                  deliveredAt.UTC(),
		Source:                       strings.TrimSpace(source),
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeOrderDelivered,
		fmt.Sprintf("%s:%d:%s", outbox.EventTypeOrderDelivered, orderRecord.ID, trackingNumber),
		outbox.AggregateTypeOrder,
		strconv.FormatUint(uint64(orderRecord.ID), 10),
		payload,
		deliveredAt,
	)
}

func enqueueAfterSalesStatusChangedDomainEvent(
	repo *repository.OutboxRepository,
	caseRecord *afterSalesDomain.AfterSalesCase,
	transition *afterSalesDomain.AfterSalesCaseEvent,
	shipment *afterSalesDomain.AfterSalesReturnShipment,
	audienceSnapshots ...outbox.NotificationAudienceSnapshot,
) error {
	if caseRecord == nil || caseRecord.ID == 0 || transition == nil || transition.ID == 0 {
		return errors.New("after-sales status event requires a persisted case transition")
	}
	idempotencyKey := fmt.Sprintf("after_sales_case:%d:transition:%d", caseRecord.ID, transition.ID)
	payload := outbox.AfterSalesStatusChangedPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, transition.CreatedAt),
		NotificationAudienceSnapshot: firstNotificationAudienceSnapshot(audienceSnapshots),
		CaseID:                       caseRecord.ID,
		OrderID:                      caseRecord.OrderID,
		OrderNumber:                  strings.TrimSpace(caseRecord.OrderNumber),
		CaseType:                     strings.TrimSpace(caseRecord.Type),
		PreviousStatus:               strings.TrimSpace(transition.FromStatus),
		NewStatus:                    strings.TrimSpace(transition.ToStatus),
		TransitionID:                 transition.ID,
		Resolution:                   strings.TrimSpace(transition.Resolution),
		UpdatedBy:                    transition.UpdatedBy,
	}
	if shipment != nil {
		payload.ReturnShipmentID = shipment.ID
		payload.Carrier = strings.TrimSpace(shipment.Carrier)
		payload.TrackingNumber = strings.TrimSpace(shipment.TrackingNumber)
		payload.TrackingURL = strings.TrimSpace(shipment.TrackingURL)
		payload.LabelURL = strings.TrimSpace(shipment.LabelURL)
		payload.WarehouseName = strings.TrimSpace(shipment.WarehouseName)
		payload.WarehouseAddress = strings.TrimSpace(shipment.WarehouseAddress)
		payload.ShippedAt = shipment.ShippedAt
		payload.ReceivedAt = shipment.ReceivedAt
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeAfterSalesStatusChanged,
		fmt.Sprintf("%s:%d:%d", outbox.EventTypeAfterSalesStatusChanged, caseRecord.ID, transition.ID),
		outbox.AggregateTypeAfterSalesCase,
		strconv.FormatUint(uint64(caseRecord.ID), 10),
		payload,
		transition.CreatedAt,
	)
}

func firstNotificationAudienceSnapshot(values []outbox.NotificationAudienceSnapshot) outbox.NotificationAudienceSnapshot {
	if len(values) == 0 {
		return outbox.NotificationAudienceSnapshot{}
	}
	return values[0]
}

func enqueueOrderRefundedDomainEvent(
	repo *repository.OutboxRepository,
	refund *paymentDomain.Refund,
	provider string,
	providerRefundID string,
	occurredAt time.Time,
	orderNumbers ...string,
) error {
	if refund == nil || refund.ID == 0 {
		return errors.New("order refunded event requires a persisted refund")
	}
	amountMoney, err := refund.AmountMoney()
	if err != nil {
		return fmt.Errorf("refund amount: %w", err)
	}
	idempotencyKey := fmt.Sprintf("refund:%d", refund.ID)
	orderNumber := ""
	if len(orderNumbers) > 0 {
		orderNumber = strings.TrimSpace(orderNumbers[0])
	}
	payload := outbox.OrderRefundedPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(idempotencyKey, occurredAt),
		// Refund completion currently receives only a refund aggregate. The
		// notification worker may supply the immutable audience context from
		// the linked order until the refund payload is extended with it.
		OrderID:          refund.OrderID,
		OrderNumber:      orderNumber,
		RefundID:         refund.ID,
		TransactionID:    refund.TransactionID,
		Provider:         strings.TrimSpace(provider),
		ProviderRefundID: strings.TrimSpace(providerRefundID),
		AmountMinor:      amountMoney.AmountMinor(),
		Currency:         strings.TrimSpace(amountMoney.Currency().String()),
		RefundStatus:     strings.TrimSpace(refund.Status),
	}
	return createCanonicalDomainEvent(
		repo,
		outbox.EventTypeOrderRefunded,
		fmt.Sprintf("%s:%d", outbox.EventTypeOrderRefunded, refund.ID),
		outbox.AggregateTypeOrder,
		strconv.FormatUint(uint64(refund.OrderID), 10),
		payload,
		occurredAt,
	)
}

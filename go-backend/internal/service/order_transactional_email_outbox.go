package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

type TransactionalOrderEmailSender interface {
	SendOrderConfirmation(to string, orderData interface{}) error
	SendShippingNotification(to string, shippingData interface{}) error
}

type OrderTransactionalEmailOutboxHandler struct {
	sender TransactionalOrderEmailSender
}

func NewOrderTransactionalEmailOutboxHandler(sender TransactionalOrderEmailSender) *OrderTransactionalEmailOutboxHandler {
	return &OrderTransactionalEmailOutboxHandler{sender: sender}
}

func (h *OrderTransactionalEmailOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.sender == nil {
		return errors.New("transactional order email sender is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	switch event.EventType {
	case outbox.EventTypeOrderConfirmationEmail:
		var payload outbox.OrderConfirmationEmailPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode order confirmation email event: %w", err)
		}
		recipient, err := validateTransactionalOrderEmailRecipient(payload.RecipientEmail)
		if err != nil {
			return fmt.Errorf("validate order confirmation email event: %w", err)
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" || payload.Amount <= 0 || strings.TrimSpace(payload.Currency) == "" || payload.PaidAt.IsZero() {
			return errors.New("order confirmation email event is incomplete")
		}
		return h.sender.SendOrderConfirmation(recipient, payload)
	case outbox.EventTypeOrderShippingNotificationEmail:
		var payload outbox.OrderShippingNotificationEmailPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode order shipping notification email event: %w", err)
		}
		recipient, err := validateTransactionalOrderEmailRecipient(payload.RecipientEmail)
		if err != nil {
			return fmt.Errorf("validate order shipping notification email event: %w", err)
		}
		if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" || strings.TrimSpace(payload.TrackingNumber) == "" || payload.ShippedAt.IsZero() {
			return errors.New("order shipping notification email event is incomplete")
		}
		return h.sender.SendShippingNotification(recipient, payload)
	default:
		return fmt.Errorf("unsupported transactional order email outbox event type %s", event.EventType)
	}
}

func enqueueOrderConfirmationEmailOutboxEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	input VerifiedGatewayPaymentInput,
	paidAt time.Time,
) error {
	if repo == nil {
		return nil
	}
	if orderRecord == nil || orderRecord.ID == 0 {
		return errors.New("order confirmation email event requires a persisted order")
	}
	amount, err := input.Amount.MajorFloat()
	if err != nil {
		return fmt.Errorf("serialize order confirmation email amount: %w", err)
	}

	payload, err := json.Marshal(outbox.OrderConfirmationEmailPayload{
		RecipientEmail: strings.TrimSpace(orderRecord.ShippingAddress.Email),
		CustomerName:   orderCustomerName(orderRecord),
		OrderID:        orderRecord.ID,
		OrderNumber:    strings.TrimSpace(orderRecord.OrderNumber),
		Amount:         amount,
		Currency:       strings.TrimSpace(input.Amount.Currency().String()),
		PaidAt:         paidAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("encode order confirmation email event: %w", err)
	}

	return repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%s", outbox.EventTypeOrderConfirmationEmail, orderRecord.ID, input.TransactionID),
		EventType:     outbox.EventTypeOrderConfirmationEmail,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   strconv.FormatUint(uint64(orderRecord.ID), 10),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   paidAt.UTC(),
	})
}

func enqueueOrderShippingNotificationEmailOutboxEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	tracking *resolvedOrderTrackingUpdate,
	shippedAt time.Time,
) error {
	if repo == nil {
		return nil
	}
	if orderRecord == nil || orderRecord.ID == 0 || tracking == nil {
		return errors.New("order shipping notification email event requires a persisted order and tracking update")
	}

	trackingNumber := strings.TrimSpace(tracking.trackingInfo.TrackingNumber)
	if trackingNumber == "" {
		return errors.New("order shipping notification email event requires a tracking number")
	}
	payload, err := json.Marshal(outbox.OrderShippingNotificationEmailPayload{
		RecipientEmail: strings.TrimSpace(orderRecord.ShippingAddress.Email),
		CustomerName:   orderCustomerName(orderRecord),
		OrderID:        orderRecord.ID,
		OrderNumber:    strings.TrimSpace(orderRecord.OrderNumber),
		CarrierName:    strings.TrimSpace(tracking.carrierName),
		TrackingNumber: trackingNumber,
		TrackingURL:    resolveOrderTrackingURL(tracking.trackingURLTemplate, trackingNumber),
		ShippedAt:      shippedAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("encode order shipping notification email event: %w", err)
	}

	return repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%s", outbox.EventTypeOrderShippingNotificationEmail, orderRecord.ID, trackingNumber),
		EventType:     outbox.EventTypeOrderShippingNotificationEmail,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   strconv.FormatUint(uint64(orderRecord.ID), 10),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   shippedAt.UTC(),
	})
}

func orderCustomerName(orderRecord *order.Order) string {
	if orderRecord == nil {
		return ""
	}
	return strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(orderRecord.ShippingAddress.FirstName),
		strings.TrimSpace(orderRecord.ShippingAddress.LastName),
	}, " "))
}

func resolveOrderTrackingURL(template, trackingNumber string) string {
	template = strings.TrimSpace(template)
	trackingNumber = strings.TrimSpace(trackingNumber)
	if template == "" || trackingNumber == "" {
		return ""
	}

	resolved := strings.ReplaceAll(template, "{tracking_number}", url.PathEscape(trackingNumber))
	parsed, err := url.Parse(resolved)
	if err != nil || parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.String()
}

func validateTransactionalOrderEmailRecipient(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("recipient email is required")
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", errors.New("recipient email is invalid")
	}
	return parsed.Address, nil
}

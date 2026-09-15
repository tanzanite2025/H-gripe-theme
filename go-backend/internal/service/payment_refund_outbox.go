package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

func enqueuePaymentRefundPendingOutboxEvent(
	repo *repository.OutboxRepository,
	refund *payment.Refund,
	currencyCode string,
	provider string,
	executionStatus string,
	attempt int,
	occurredAt time.Time,
) error {
	if repo == nil {
		return nil
	}
	payload, err := paymentRefundOutboxPayload(refund, currencyCode, provider, executionStatus, "", "", occurredAt)
	if err != nil {
		return err
	}
	eventKey := fmt.Sprintf("%s:%d:%d", outbox.EventTypePaymentRefundPending, refund.ID, attempt)
	return createPaymentRefundOutboxEvent(repo, eventKey, outbox.EventTypePaymentRefundPending, refund.ID, payload, occurredAt)
}

func enqueuePaymentRefundCompletedOutboxEvent(
	repo *repository.OutboxRepository,
	refund *payment.Refund,
	currencyCode string,
	provider string,
	providerRefundID string,
	executionStatus string,
	occurredAt time.Time,
) error {
	if repo == nil {
		return nil
	}
	payload, err := paymentRefundOutboxPayload(refund, currencyCode, provider, executionStatus, providerRefundID, "", occurredAt)
	if err != nil {
		return err
	}
	eventKey := fmt.Sprintf("%s:%d", outbox.EventTypePaymentRefundCompleted, refund.ID)
	return createPaymentRefundOutboxEvent(repo, eventKey, outbox.EventTypePaymentRefundCompleted, refund.ID, payload, occurredAt)
}

func enqueuePaymentRefundFailedOutboxEvent(
	repo *repository.OutboxRepository,
	refund *payment.Refund,
	currencyCode string,
	provider string,
	providerRefundID string,
	executionStatus string,
	errorMessage string,
	attempt int,
	occurredAt time.Time,
) error {
	if repo == nil {
		return nil
	}
	payload, err := paymentRefundOutboxPayload(refund, currencyCode, provider, executionStatus, providerRefundID, errorMessage, occurredAt)
	if err != nil {
		return err
	}
	eventKey := fmt.Sprintf("%s:%d:%d:%s", outbox.EventTypePaymentRefundFailed, refund.ID, attempt, strings.TrimSpace(providerRefundID))
	return createPaymentRefundOutboxEvent(repo, eventKey, outbox.EventTypePaymentRefundFailed, refund.ID, payload, occurredAt)
}

func paymentRefundOutboxPayload(
	refund *payment.Refund,
	currencyCode string,
	provider string,
	executionStatus string,
	providerRefundID string,
	errorMessage string,
	occurredAt time.Time,
) (outbox.PaymentRefundPayload, error) {
	if refund == nil || refund.ID == 0 {
		return outbox.PaymentRefundPayload{}, errors.New("payment refund outbox event requires a persisted refund")
	}
	currencyCode = strings.TrimSpace(currencyCode)
	if currencyCode == "" {
		currencyCode = strings.TrimSpace(refund.Currency)
	}
	amountMoney, err := refund.AmountMoney()
	if err != nil {
		return outbox.PaymentRefundPayload{}, fmt.Errorf("refund amount: %w", err)
	}
	giftCardMoney, err := refund.GiftCardRefundMoney()
	if err != nil {
		return outbox.PaymentRefundPayload{}, fmt.Errorf("gift card refund amount: %w", err)
	}
	requestedMoney, err := refund.RequestedAmountMoney()
	if err != nil {
		return outbox.PaymentRefundPayload{}, fmt.Errorf("requested refund amount: %w", err)
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	} else {
		occurredAt = occurredAt.UTC()
	}
	return outbox.PaymentRefundPayload{
		RefundID:             refund.ID,
		OrderID:              refund.OrderID,
		TransactionID:        refund.TransactionID,
		Provider:             strings.TrimSpace(provider),
		ProviderRefundID:     strings.TrimSpace(providerRefundID),
		RefundStatus:         strings.TrimSpace(refund.Status),
		ExecutionStatus:      strings.TrimSpace(executionStatus),
		AmountMinor:          amountMoney.AmountMinor(),
		GiftCardAmountMinor:  giftCardMoney.AmountMinor(),
		RequestedAmountMinor: requestedMoney.AmountMinor(),
		Currency:             currencyCode,
		Reason:               strings.TrimSpace(refund.Reason),
		ErrorMessage:         strings.TrimSpace(errorMessage),
		OccurredAt:           occurredAt,
	}, nil
}

func createPaymentRefundOutboxEvent(
	repo *repository.OutboxRepository,
	eventKey string,
	eventType string,
	refundID uint,
	payload outbox.PaymentRefundPayload,
	availableAt time.Time,
) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode payment refund outbox event: %w", err)
	}
	if availableAt.IsZero() {
		availableAt = time.Now().UTC()
	} else {
		availableAt = availableAt.UTC()
	}
	return repo.CreateEvent(&outbox.Event{
		EventKey:      eventKey,
		EventType:     eventType,
		AggregateType: outbox.AggregateTypePayment,
		AggregateID:   strconv.FormatUint(uint64(refundID), 10),
		Payload:       datatypes.JSON(encoded),
		AvailableAt:   availableAt,
	})
}

func paymentAmountMinor(amount float64, currencyCode string) (int64, error) {
	money, err := parseRefundMoney(amount, currencyCode)
	if err != nil {
		return 0, err
	}
	return money.AmountMinor(), nil
}

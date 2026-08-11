package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/datatypes"
)

const (
	gatewayPaymentAmountTolerance = 0.01
)

type VerifiedGatewayPaymentInput struct {
	Provider        string
	OrderNumber     string
	TransactionID   string
	PaymentMethod   string
	Amount          float64
	Currency        string
	GatewayResponse string
}

func (s *PaymentService) GetTransaction(id uint) (*payment.Transaction, error) {
	return s.paymentRepo.FindTransactionByID(id)
}

func (s *PaymentService) GetOrderTransactions(orderID uint) ([]payment.Transaction, error) {
	return s.paymentRepo.FindTransactionByOrderID(orderID)
}

func (s *PaymentService) RecordVerifiedGatewayPayment(input VerifiedGatewayPaymentInput) error {
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.OrderNumber == "" {
		return errors.New("order_number is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = input.Provider
	}
	input.Currency = normalizePaymentCurrency(input.Currency)
	if input.Currency == "" {
		return errors.New("currency is required")
	}

	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var existingTransaction *payment.Transaction
		if transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID); err == nil {
			existingTransaction = transaction
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		if existingTransaction != nil && existingTransaction.Status == "completed" {
			if o.PaymentStatus == "paid" || o.Status == "payment_expired" {
				return nil
			}
		}
		if o.PaymentStatus == "paid" {
			return errors.New("order is already paid")
		}
		if o.Status == "cancelled" || o.Status == "refunded" {
			return fmt.Errorf("cannot mark %s order as paid", o.Status)
		}
		expectedCurrency, err := orderPaymentCurrency(o)
		if err != nil {
			return err
		}
		if input.Currency != expectedCurrency {
			return fmt.Errorf("transaction currency %s does not match order currency %s", input.Currency, expectedCurrency)
		}
		if math.Abs(o.TotalAmount-input.Amount) >= gatewayPaymentAmountTolerance {
			return fmt.Errorf("payment amount %.2f does not match order total %.2f", input.Amount, o.TotalAmount)
		}

		completedAt := time.Now().UTC()
		if err := saveCompletedGatewayTransaction(repos.Payment, existingTransaction, completedAt, payment.Transaction{
			OrderID:         o.ID,
			TransactionID:   input.TransactionID,
			PaymentMethod:   input.PaymentMethod,
			Amount:          input.Amount,
			Currency:        input.Currency,
			Status:          "completed",
			GatewayResponse: input.GatewayResponse,
			CompletedAt:     &completedAt,
		}); err != nil {
			return err
		}
		if o.Status == "payment_expired" {
			return createLatePaymentReview(repos.Payment, o.ID, input.TransactionID, o.OrderNumber)
		}
		if err := repos.Order.UpdatePaymentStatus(o.ID, "paid"); err != nil {
			return err
		}
		if o.Status == "pending" || o.Status == "paid" {
			if err := repos.Order.UpdateStatus(o.ID, "processing"); err != nil {
				return err
			}
		}
		if err := enqueueOrderPaidOutboxEvent(repos.Outbox, o, input, completedAt); err != nil {
			return err
		}
		return enqueueVerifiedConversionOutboxEvent(repos.Outbox, repos.OrderAttribution, o, input, completedAt)
	})
	if err == nil && s.risk != nil {
		s.risk.RecordProviderSuccess(input.Provider)
	}
	return err
}

func enqueueVerifiedConversionOutboxEvent(
	outboxRepo *repository.OutboxRepository,
	attributionRepo *repository.OrderAttributionRepository,
	o *order.Order,
	input VerifiedGatewayPaymentInput,
	verifiedAt time.Time,
) error {
	if outboxRepo == nil || o == nil {
		return nil
	}

	payload := outbox.VerifiedConversionPayload{
		OrderID:    o.ID,
		Amount:     input.Amount,
		Currency:   input.Currency,
		VerifiedAt: verifiedAt.UTC(),
	}
	if attributionRepo != nil {
		value, err := attributionRepo.FindByOrderID(o.ID)
		if err != nil && !repository.IsRecordNotFound(err) {
			return err
		}
		if value != nil {
			payload.Attribution = &outbox.VerifiedConversionAttribution{
				Source:      value.Source,
				Medium:      value.Medium,
				Campaign:    value.Campaign,
				Term:        value.Term,
				Content:     value.Content,
				ClickIDKind: value.ClickIDKind,
				ClickID:     value.ClickID,
			}
		}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return outboxRepo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%s", outbox.EventTypeVerifiedConversion, o.ID, input.TransactionID),
		EventType:     outbox.EventTypeVerifiedConversion,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprint(o.ID),
		Payload:       datatypes.JSON(encoded),
	})
}

func enqueueOrderPaidOutboxEvent(repo *repository.OutboxRepository, o *order.Order, input VerifiedGatewayPaymentInput, paidAt time.Time) error {
	if repo == nil || o == nil {
		return nil
	}
	customerName := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(o.ShippingAddress.FirstName),
		strings.TrimSpace(o.ShippingAddress.LastName),
	}, " "))
	payload, err := json.Marshal(outbox.OrderPaidPayload{
		OrderID:              o.ID,
		OrderNumber:          o.OrderNumber,
		UserID:               o.UserID,
		PaymentTransactionID: input.TransactionID,
		PaymentMethod:        input.PaymentMethod,
		Amount:               input.Amount,
		Currency:             input.Currency,
		PaidAt:               paidAt.UTC(),
		CustomerEmail:        strings.TrimSpace(o.ShippingAddress.Email),
		CustomerName:         customerName,
		ShippingCountry:      strings.TrimSpace(o.ShippingAddress.Country),
	})
	if err != nil {
		return err
	}
	return repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%s", outbox.EventTypeOrderPaid, o.ID, input.TransactionID),
		EventType:     outbox.EventTypeOrderPaid,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprint(o.ID),
		Payload:       datatypes.JSON(payload),
	})
}

func saveCompletedGatewayTransaction(repo *repository.PaymentRepository, existing *payment.Transaction, completedAt time.Time, next payment.Transaction) error {
	if existing == nil {
		return repo.CreateTransaction(&next)
	}
	existing.OrderID = next.OrderID
	existing.PaymentMethod = next.PaymentMethod
	existing.Amount = next.Amount
	existing.Currency = next.Currency
	existing.Status = "completed"
	existing.GatewayResponse = next.GatewayResponse
	existing.ErrorMessage = ""
	existing.CompletedAt = &completedAt
	return repo.UpdateTransaction(existing)
}

func createLatePaymentReview(repo *repository.PaymentRepository, orderID uint, paymentIntentID, orderNumber string) error {
	return repo.CreatePaymentReview(&payment.PaymentReview{
		OrderID:         &orderID,
		PaymentIntentID: paymentIntentID,
		Status:          "pending",
		Reason:          "payment_succeeded_after_expiration",
		Source:          "webhook",
		Notes:           fmt.Sprintf("Payment succeeded after order %s was marked payment_expired. Review inventory and refund/manual fulfillment.", orderNumber),
	})
}

func normalizePaymentCurrency(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func orderPaymentCurrency(o *order.Order) (string, error) {
	if o == nil {
		return "", errors.New("order is required")
	}
	value := currency.NormalizeCode(o.Currency)
	if !currency.IsValidCode(value) || !currency.IsCatalogCode(value) {
		return "", errors.New("order currency is not configured")
	}
	return value, nil
}

func normalizeGatewayAttemptStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "processing":
		return "processing"
	case "requires_action", "requires_confirmation", "requires_payment_method", "requires_capture":
		return "requires_action"
	case "failed", "payment_failed", "canceled", "cancelled":
		return "failed"
	default:
		return "pending"
	}
}

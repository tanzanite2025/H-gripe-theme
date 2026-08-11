package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
)

var (
	ErrStripeDisputeNotFound = errors.New("stripe dispute not found")
	ErrPayPalDisputeNotFound = errors.New("paypal dispute not found")
)

type StripeDisputeInput struct {
	StripeDisputeID string
	StripeChargeID  string
	PaymentIntentID string
	OrderID         *uint
	Amount          float64
	Currency        string
	Reason          string
	Status          string
	EvidenceDueAt   *time.Time
	RawPayload      string
}

type PayPalDisputeInput struct {
	PayPalDisputeID       string
	ProviderPaymentID     string
	OrderReference        string
	Amount                float64
	Currency              string
	Reason                string
	Status                string
	DisputeState          string
	DisputeLifeCycleStage string
	RawPayload            string
}

func (s *PaymentService) RecordStripeDispute(input StripeDisputeInput) (*paymentdomain.StripeDispute, error) {
	if strings.TrimSpace(input.StripeDisputeID) == "" {
		return nil, errors.New("stripe dispute id is required")
	}
	if input.Amount <= 0 {
		return nil, errors.New("dispute amount must be greater than zero")
	}
	input.Currency = currency.NormalizeCode(input.Currency)
	if input.Currency == "" {
		return nil, errors.New("dispute currency is required")
	}
	if !currency.IsValidCode(input.Currency) || !currency.IsCatalogCode(input.Currency) {
		return nil, errors.New("dispute currency must be a supported ISO 4217 code")
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = "needs_response"
	}

	var transactionID *uint
	orderID := input.OrderID
	if strings.TrimSpace(input.PaymentIntentID) != "" {
		if transaction, err := s.paymentRepo.FindTransactionByTransactionID(input.PaymentIntentID); err == nil {
			transactionID = &transaction.ID
			if orderID == nil {
				orderID = &transaction.OrderID
			}
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	record := &paymentdomain.StripeDispute{
		StripeDisputeID: input.StripeDisputeID,
		StripeChargeID:  input.StripeChargeID,
		PaymentIntentID: input.PaymentIntentID,
		OrderID:         orderID,
		TransactionID:   transactionID,
		Amount:          input.Amount,
		Currency:        input.Currency,
		Reason:          input.Reason,
		Status:          input.Status,
		EvidenceDueAt:   input.EvidenceDueAt,
		RawPayload:      input.RawPayload,
	}
	if err := s.paymentRepo.UpsertStripeDispute(record); err != nil {
		return nil, err
	}

	if orderID != nil && disputeNeedsResponse(record.Status) {
		if _, err := s.paymentRepo.FindPendingPaymentReviewByOrderID(*orderID); repository.IsRecordNotFound(err) {
			_ = s.paymentRepo.CreatePaymentReview(&paymentdomain.PaymentReview{
				OrderID:         orderID,
				TransactionID:   transactionID,
				DisputeID:       &record.ID,
				PaymentIntentID: input.PaymentIntentID,
				Status:          "pending",
				Reason:          "stripe_dispute",
				Source:          "dispute",
				Notes:           fmt.Sprintf("Stripe dispute %s requires review.", record.StripeDisputeID),
			})
		}
	}

	return record, nil
}

func (s *PaymentService) GetStripeDispute(id uint) (*paymentdomain.StripeDispute, error) {
	record, err := s.paymentRepo.FindStripeDisputeByID(id)
	if repository.IsRecordNotFound(err) {
		return nil, ErrStripeDisputeNotFound
	}
	return record, err
}

func (s *PaymentService) ListStripeDisputes(status string, page, pageSize int) ([]paymentdomain.StripeDispute, int64, error) {
	return s.paymentRepo.ListStripeDisputes(status, page, pageSize)
}

func (s *PaymentService) RecordPayPalDispute(input PayPalDisputeInput) (*paymentdomain.PayPalDispute, error) {
	if strings.TrimSpace(input.PayPalDisputeID) == "" {
		return nil, errors.New("paypal dispute id is required")
	}

	var transactionID *uint
	var orderID *uint
	if providerPaymentID := strings.TrimSpace(input.ProviderPaymentID); providerPaymentID != "" {
		if transaction, err := s.paymentRepo.FindTransactionByTransactionID(providerPaymentID); err == nil {
			transactionID = &transaction.ID
			orderID = &transaction.OrderID
			if input.Amount <= 0 {
				input.Amount = transaction.Amount
			}
			if strings.TrimSpace(input.Currency) == "" {
				input.Currency = transaction.Currency
			}
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	if orderID == nil && s.orderRepo != nil && strings.TrimSpace(input.OrderReference) != "" {
		if orderRecord, err := s.orderRepo.FindByOrderNumber(strings.TrimSpace(input.OrderReference)); err == nil {
			orderID = &orderRecord.ID
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	if input.Amount <= 0 {
		return nil, errors.New("paypal dispute amount must be greater than zero")
	}
	input.Currency = currency.NormalizeCode(input.Currency)
	if input.Currency == "" {
		return nil, errors.New("paypal dispute currency is required")
	}
	if !currency.IsValidCode(input.Currency) || !currency.IsCatalogCode(input.Currency) {
		return nil, errors.New("paypal dispute currency must be a supported ISO 4217 code")
	}

	record := &paymentdomain.PayPalDispute{
		PayPalDisputeID:       strings.TrimSpace(input.PayPalDisputeID),
		OrderID:               orderID,
		TransactionID:         transactionID,
		ProviderPaymentID:     strings.TrimSpace(input.ProviderPaymentID),
		Amount:                input.Amount,
		Currency:              input.Currency,
		Reason:                strings.TrimSpace(input.Reason),
		Status:                strings.TrimSpace(input.Status),
		DisputeState:          strings.TrimSpace(input.DisputeState),
		DisputeLifeCycleStage: strings.TrimSpace(input.DisputeLifeCycleStage),
		RawPayload:            input.RawPayload,
	}
	if record.Status == "" {
		record.Status = "WAITING_FOR_SELLER_RESPONSE"
	}
	if err := s.paymentRepo.UpsertPayPalDispute(record); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *PaymentService) GetPayPalDispute(id uint) (*paymentdomain.PayPalDispute, error) {
	record, err := s.paymentRepo.FindPayPalDisputeByID(id)
	if repository.IsRecordNotFound(err) {
		return nil, ErrPayPalDisputeNotFound
	}
	return record, err
}

func (s *PaymentService) ListPayPalDisputes(status string, page, pageSize int) ([]paymentdomain.PayPalDispute, int64, error) {
	return s.paymentRepo.ListPayPalDisputes(status, page, pageSize)
}

func disputeNeedsResponse(status string) bool {
	return status == "needs_response" || status == "warning_needs_response"
}

package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
)

var (
	ErrStripeDisputeNotFound             = errors.New("stripe dispute not found")
	ErrPayPalDisputeNotFound             = errors.New("paypal dispute not found")
	ErrPaymentDisputeTransactionRequired = errors.New("payment dispute transaction is not configured")
)

type StripeDisputeInput struct {
	StripeDisputeID string
	StripeChargeID  string
	PaymentIntentID string
	OrderID         *uint
	AmountMinor     int64
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
	AmountMinor           int64
	Currency              string
	Reason                string
	Status                string
	DisputeState          string
	DisputeLifeCycleStage string
	RawPayload            string
}

func (s *PaymentService) RecordStripeDispute(input StripeDisputeInput) (*paymentdomain.StripeDispute, error) {
	input.StripeDisputeID = strings.TrimSpace(input.StripeDisputeID)
	input.StripeChargeID = strings.TrimSpace(input.StripeChargeID)
	input.PaymentIntentID = strings.TrimSpace(input.PaymentIntentID)
	if input.StripeDisputeID == "" {
		return nil, errors.New("stripe dispute id is required")
	}
	if input.AmountMinor <= 0 {
		return nil, errors.New("dispute amount must be greater than zero")
	}
	input.Currency = currency.NormalizeCode(input.Currency)
	if input.Currency == "" {
		return nil, errors.New("dispute currency is required")
	}
	if !currency.IsValidCode(input.Currency) || !currency.IsCatalogCode(input.Currency) {
		return nil, errors.New("dispute currency must be a supported ISO 4217 code")
	}
	amountMoney, err := domainmoney.New(input.AmountMinor, input.Currency)
	if err != nil || amountMoney.AmountMinor() <= 0 {
		return nil, errors.New("dispute amount must be a valid positive monetary amount")
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = "needs_response"
	}

	if s == nil || s.txManager == nil {
		return nil, ErrPaymentDisputeTransactionRequired
	}

	var record *paymentdomain.StripeDispute
	err = s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var existing *paymentdomain.StripeDispute
		if found, err := repos.Payment.FindStripeDisputeByStripeID(input.StripeDisputeID); err == nil {
			existing = found
			if input.StripeChargeID == "" {
				input.StripeChargeID = found.StripeChargeID
			}
			if input.PaymentIntentID == "" {
				input.PaymentIntentID = found.PaymentIntentID
			}
			if input.OrderID == nil {
				input.OrderID = found.OrderID
			}
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		var transactionID *uint
		orderID := input.OrderID
		// Resolve provider identity without a lock so the order can be selected;
		// then acquire locks in the canonical order -> transaction sequence.
		if orderID == nil && existing != nil {
			orderID = existing.OrderID
		}
		if input.PaymentIntentID != "" {
			if transaction, err := repos.Payment.FindTransactionByTransactionID(input.PaymentIntentID); err == nil {
				if orderID == nil {
					orderID = &transaction.OrderID
				} else if *orderID != transaction.OrderID {
					return fmt.Errorf("stripe dispute order does not match payment transaction")
				}
			} else if !repository.IsRecordNotFound(err) {
				return err
			}
		}
		if orderID != nil {
			if _, err := repos.Order.FindByIDForUpdate(*orderID); err != nil {
				return normalizeOrderError(err)
			}
		}
		if input.PaymentIntentID != "" {
			if transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.PaymentIntentID); err == nil {
				transactionID = &transaction.ID
				if orderID == nil {
					orderID = &transaction.OrderID
				} else if *orderID != transaction.OrderID {
					return fmt.Errorf("stripe dispute order does not match payment transaction")
				}
			} else if !repository.IsRecordNotFound(err) {
				return err
			}
		}
		if transactionID == nil && existing != nil {
			transactionID = existing.TransactionID
		}
		record = &paymentdomain.StripeDispute{
			StripeDisputeID: input.StripeDisputeID,
			StripeChargeID:  input.StripeChargeID,
			PaymentIntentID: input.PaymentIntentID,
			OrderID:         orderID,
			TransactionID:   transactionID,
			AmountMinor:     amountMoney.AmountMinor(),
			Currency:        input.Currency,
			Reason:          input.Reason,
			Status:          input.Status,
			EvidenceDueAt:   input.EvidenceDueAt,
			RawPayload:      input.RawPayload,
		}
		if err := repos.Payment.UpsertStripeDispute(record); err != nil {
			return err
		}

		if record.OrderID != nil && disputeIsActive(record.Status) {
			if err := ensureDisputePaymentReview(
				repos.Payment,
				record.OrderID,
				record.TransactionID,
				&record.ID,
				record.PaymentIntentID,
				"stripe_dispute",
				fmt.Sprintf("Stripe dispute %s requires review.", record.StripeDisputeID),
			); err != nil {
				return err
			}
			if err := repos.Order.MarkDisputed(*record.OrderID); err != nil {
				return err
			}
			return enqueueReferralOrderInvalidatedOutboxEvent(
				repos.Outbox,
				*record.OrderID,
				time.Now().UTC(),
				"Stripe payment dispute opened",
				"stripe_dispute",
				record.StripeDisputeID,
			)
		}
		if record.OrderID != nil {
			if err := resolveDisputeOrderProjection(repos, *record.OrderID, &record.ID, record.Status, ""); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
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
	input.PayPalDisputeID = strings.TrimSpace(input.PayPalDisputeID)
	input.ProviderPaymentID = strings.TrimSpace(input.ProviderPaymentID)
	input.OrderReference = strings.TrimSpace(input.OrderReference)
	if input.PayPalDisputeID == "" {
		return nil, errors.New("paypal dispute id is required")
	}
	if s == nil || s.txManager == nil {
		return nil, ErrPaymentDisputeTransactionRequired
	}

	var record *paymentdomain.PayPalDispute
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var paypalAmountMinor int64
		var orderID *uint
		var existing *paymentdomain.PayPalDispute
		if found, err := repos.Payment.FindPayPalDisputeByPayPalID(input.PayPalDisputeID); err == nil {
			existing = found
			if input.ProviderPaymentID == "" {
				input.ProviderPaymentID = found.ProviderPaymentID
			}
			if orderID == nil {
				orderID = found.OrderID
			}
			if input.Status == "" {
				input.Status = found.Status
			}
			if input.DisputeState == "" {
				input.DisputeState = found.DisputeState
			}
			if input.DisputeLifeCycleStage == "" {
				input.DisputeLifeCycleStage = found.DisputeLifeCycleStage
			}
			if input.AmountMinor <= 0 {
				paypalAmountMinor = found.AmountMinor
			}
			if strings.TrimSpace(input.Currency) == "" {
				input.Currency = found.Currency
			}
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		var transactionID *uint
		// Resolve the transaction without a row lock first so its order can be
		// identified. The canonical lock order is order -> transaction; taking
		// the transaction lock here would deadlock with payment callbacks that
		// already hold the order projection lock.
		if input.ProviderPaymentID != "" {
			if transaction, err := repos.Payment.FindTransactionByTransactionID(input.ProviderPaymentID); err == nil {
				if orderID == nil {
					orderID = &transaction.OrderID
				} else if *orderID != transaction.OrderID {
					return fmt.Errorf("paypal dispute order does not match payment transaction")
				}
				if input.AmountMinor <= 0 {
					paypalAmountMinor = transaction.AmountMinor
				}
				if strings.TrimSpace(input.Currency) == "" {
					input.Currency = transaction.Currency
				}
			} else if !repository.IsRecordNotFound(err) {
				return err
			}
		}
		if orderID == nil && existing != nil {
			orderID = existing.OrderID
		}
		if orderID == nil && input.OrderReference != "" {
			orderRecord, err := repos.Order.FindByOrderNumberForVerificationForUpdate(input.OrderReference)
			if err == nil {
				orderID = &orderRecord.ID
			} else if !repository.IsRecordNotFound(err) {
				return err
			}
		}
		if orderID != nil {
			if _, err := repos.Order.FindByIDForUpdate(*orderID); err != nil {
				return normalizeOrderError(err)
			}
		}
		if input.ProviderPaymentID != "" {
			if transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.ProviderPaymentID); err == nil {
				transactionID = &transaction.ID
				if orderID == nil {
					orderID = &transaction.OrderID
				} else if *orderID != transaction.OrderID {
					return fmt.Errorf("paypal dispute order does not match payment transaction")
				}
				if input.AmountMinor <= 0 {
					paypalAmountMinor = transaction.AmountMinor
				}
				if strings.TrimSpace(input.Currency) == "" {
					input.Currency = transaction.Currency
				}
			} else if !repository.IsRecordNotFound(err) {
				return err
			}
		}
		if transactionID == nil && existing != nil {
			transactionID = existing.TransactionID
		}
		if input.AmountMinor <= 0 && paypalAmountMinor <= 0 {
			return errors.New("paypal dispute amount must be greater than zero")
		}
		input.Currency = currency.NormalizeCode(input.Currency)
		if input.Currency == "" {
			return errors.New("paypal dispute currency is required")
		}
		if !currency.IsValidCode(input.Currency) || !currency.IsCatalogCode(input.Currency) {
			return errors.New("paypal dispute currency must be a supported ISO 4217 code")
		}
		if input.Status == "" {
			input.Status = "WAITING_FOR_SELLER_RESPONSE"
		}

		if input.AmountMinor > 0 {
			paypalAmountMoney, amountErr := domainmoney.New(input.AmountMinor, input.Currency)
			if amountErr != nil || paypalAmountMoney.AmountMinor() <= 0 {
				return errors.New("paypal dispute amount must be a valid positive monetary amount")
			}
			paypalAmountMinor = paypalAmountMoney.AmountMinor()
		}
		record = &paymentdomain.PayPalDispute{
			PayPalDisputeID:       input.PayPalDisputeID,
			OrderID:               orderID,
			TransactionID:         transactionID,
			ProviderPaymentID:     input.ProviderPaymentID,
			AmountMinor:           paypalAmountMinor,
			Currency:              input.Currency,
			Reason:                strings.TrimSpace(input.Reason),
			Status:                strings.TrimSpace(input.Status),
			DisputeState:          strings.TrimSpace(input.DisputeState),
			DisputeLifeCycleStage: strings.TrimSpace(input.DisputeLifeCycleStage),
			RawPayload:            input.RawPayload,
		}
		if err := repos.Payment.UpsertPayPalDispute(record); err != nil {
			return err
		}

		if record.OrderID != nil && paypalDisputeIsActive(record.Status, record.DisputeState) {
			if err := ensureDisputePaymentReview(
				repos.Payment,
				record.OrderID,
				record.TransactionID,
				&record.ID,
				record.ProviderPaymentID,
				"paypal_dispute",
				fmt.Sprintf("PayPal dispute %s requires review.", record.PayPalDisputeID),
			); err != nil {
				return err
			}
			if err := repos.Order.MarkDisputed(*record.OrderID); err != nil {
				return err
			}
			return enqueueReferralOrderInvalidatedOutboxEvent(
				repos.Outbox,
				*record.OrderID,
				time.Now().UTC(),
				"PayPal payment dispute opened",
				"paypal_dispute",
				record.PayPalDisputeID,
			)
		}
		if record.OrderID != nil {
			if err := resolveDisputeOrderProjection(repos, *record.OrderID, &record.ID, record.Status, record.DisputeState); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
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
	status = strings.ToLower(strings.TrimSpace(status))
	return status == "needs_response" || status == "warning_needs_response"
}

func disputeIsActive(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "won", "lost", "closed", "resolved", "cancelled", "canceled", "denied", "rejected", "withdrawn", "refunded":
		return false
	default:
		return true
	}
}

func paypalDisputeIsActive(status, state string) bool {
	return disputeIsActive(status) && disputeIsActive(state)
}

// resolveDisputeOrderProjection closes only the dispute hold represented by
// this provider record. A winning/closed dispute may restore the order's
// pre-dispute status; a lost/refunded dispute remains held for reconciliation.
func resolveDisputeOrderProjection(
	repos repository.TxRepositories,
	orderID uint,
	disputeID *uint,
	status string,
	state string,
) error {
	if disputeIsActive(status) && (state == "" || disputeIsActive(state)) {
		return nil
	}
	reviewStatus := "cancelled"
	if disputeIsLoss(status) || disputeIsLoss(state) {
		reviewStatus = "rejected"
	}
	if disputeID != nil {
		if err := repos.Payment.ResolvePendingDisputeReview(
			*disputeID,
			reviewStatus,
			"Provider dispute is no longer active.",
		); err != nil {
			return err
		}
	}
	if disputeIsLoss(status) || disputeIsLoss(state) {
		return nil
	}
	activePaymentReview, err := repos.Payment.HasActivePaymentReviewByOrderID(orderID)
	if err != nil {
		return err
	}
	activeStripeDispute, err := repos.Payment.HasActiveStripeDisputeByOrderID(orderID)
	if err != nil {
		return err
	}
	activePayPalDispute, err := repos.Payment.HasActivePayPalDisputeByOrderID(orderID)
	if err != nil {
		return err
	}
	if activePaymentReview || activeStripeDispute || activePayPalDispute {
		return nil
	}
	return repos.Order.RestoreDisputeProjection(orderID)
}

func disputeIsLoss(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "lost", "refunded":
		return true
	default:
		return false
	}
}

func ensureDisputePaymentReview(
	repo *repository.PaymentRepository,
	orderID *uint,
	transactionID *uint,
	disputeID *uint,
	paymentIntentID string,
	reason string,
	notes string,
) error {
	if orderID == nil {
		return nil
	}
	if _, err := repo.FindPendingPaymentReviewByOrderID(*orderID); err == nil {
		return nil
	} else if !repository.IsRecordNotFound(err) {
		return err
	}

	return repo.CreatePaymentReview(&paymentdomain.PaymentReview{
		OrderID:         orderID,
		TransactionID:   transactionID,
		DisputeID:       disputeID,
		PaymentIntentID: strings.TrimSpace(paymentIntentID),
		Status:          "pending",
		Reason:          reason,
		Source:          "dispute",
		Notes:           notes,
	})
}

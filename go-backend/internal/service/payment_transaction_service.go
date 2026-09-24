package service

import (
	"commerce-platform/internal/domain/currency"
	currencydomain "commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
)

type VerifiedGatewayPaymentInput struct {
	Provider         string
	OrderNumber      string
	TransactionID    string
	PaymentMethod    string
	Amount           domainmoney.Money
	GatewayResponse  string
	LiabilityShifted *bool
}

type VerifiedGatewayPaymentResult struct {
	DuplicatePaid bool
	TransactionID string
	RefundID      uint
}

const (
	duplicatePaidRefundReasonCode  = "duplicate_paid_payment"
	duplicatePaidRefundReason      = "duplicate_paid_payment: Duplicate payment received after the order was already paid; refund the duplicate gateway transaction."
	highValueLiabilityReviewReason = "high_value_liability_shift_not_transferred"
)

func (s *PaymentService) GetTransaction(id uint) (*payment.Transaction, error) {
	return s.paymentRepo.FindTransactionByID(id)
}

func (s *PaymentService) GetOrderTransactions(orderID uint) ([]payment.Transaction, error) {
	return s.paymentRepo.FindTransactionByOrderID(orderID)
}

func (s *PaymentService) RecordVerifiedGatewayPayment(input VerifiedGatewayPaymentInput) error {
	_, err := s.RecordVerifiedGatewayPaymentResult(input)
	return err
}

func (s *PaymentService) RecordVerifiedGatewayPaymentResult(input VerifiedGatewayPaymentInput) (VerifiedGatewayPaymentResult, error) {
	if input.Provider == "" {
		return VerifiedGatewayPaymentResult{}, errors.New("provider is required")
	}
	if input.OrderNumber == "" {
		return VerifiedGatewayPaymentResult{}, errors.New("order_number is required")
	}
	if input.TransactionID == "" {
		return VerifiedGatewayPaymentResult{}, errors.New("transaction_id is required")
	}
	if input.Amount.Currency().String() == "" || input.Amount.AmountMinor() <= 0 {
		return VerifiedGatewayPaymentResult{}, errors.New("amount must be greater than zero")
	}
	if err := input.Amount.Validate(); err != nil {
		return VerifiedGatewayPaymentResult{}, fmt.Errorf("invalid amount: %w", err)
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = input.Provider
	}
	inputCurrency := normalizePaymentCurrency(input.Amount.Currency().String())

	var result VerifiedGatewayPaymentResult
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByOrderNumberForVerificationForUpdate(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}

		// The order is the serialization point for the payment projection. A
		// concurrent Capture and webhook must reload the transaction only after
		// waiting on this lock; otherwise both requests can observe a missing
		// transaction and the loser may manufacture a duplicate-paid refund for
		// the very charge that actually won.
		var existingTransaction *payment.Transaction
		if transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID); err == nil {
			existingTransaction = transaction
		} else if !repository.IsRecordNotFound(err) {
			return err
		}
		if existingTransaction != nil && existingTransaction.OrderID != o.ID {
			return errors.New("transaction does not belong to order")
		}
		if existingTransaction != nil && existingTransaction.Status == "completed" {
			if o.PaymentStatus == "paid" {
				return nil
			}
			if reason, ok := latePaymentReviewReason(o.Status); ok {
				return s.createLatePaymentReview(
					repos,
					o.ID,
					existingTransaction.ID,
					existingTransaction.TransactionID,
					o.OrderNumber,
					reason,
				)
			}
		}
		if o.PaymentStatus == "paid" && existingTransaction != nil &&
			existingTransaction.Status == payment.TransactionStatusDuplicatePaid {
			refund, err := createDuplicatePaidRefundInTx(repos, o, existingTransaction)
			if err != nil {
				return err
			}
			result = VerifiedGatewayPaymentResult{
				DuplicatePaid: true,
				TransactionID: existingTransaction.TransactionID,
				RefundID:      refund.ID,
			}
			return nil
		}

		expectedSettlement, err := orderPaymentSettlement(o)
		if err != nil {
			return err
		}
		expectedCurrency := expectedSettlement.Currency().String()
		if inputCurrency != expectedCurrency {
			return fmt.Errorf("transaction currency %s does not match order currency %s", inputCurrency, expectedCurrency)
		}
		if o.PaymentStatus == "paid" {
			completedAt := time.Now().UTC()
			duplicateTransaction, err := saveVerifiedGatewayTransaction(repos.Payment, existingTransaction, completedAt, payment.Transaction{
				OrderID:          o.ID,
				TransactionID:    input.TransactionID,
				PaymentMethod:    input.PaymentMethod,
				AmountMinor:      input.Amount.AmountMinor(),
				Currency:         inputCurrency,
				Status:           payment.TransactionStatusDuplicatePaid,
				GatewayResponse:  input.GatewayResponse,
				LiabilityShifted: input.LiabilityShifted,
				CompletedAt:      &completedAt,
			})
			if err != nil {
				return err
			}
			refund, err := createDuplicatePaidRefundInTx(repos, o, duplicateTransaction)
			if err != nil {
				return err
			}
			result = VerifiedGatewayPaymentResult{
				DuplicatePaid: true,
				TransactionID: duplicateTransaction.TransactionID,
				RefundID:      refund.ID,
			}
			return nil
		}

		if expectedSettlement.AmountMinor() != input.Amount.AmountMinor() {
			actualAmount, actualErr := input.Amount.FormatMajor()
			expectedAmount, expectedErr := expectedSettlement.FormatMajor()
			if actualErr != nil || expectedErr != nil {
				return fmt.Errorf("payment amount does not match payable amount")
			}
			return fmt.Errorf("payment amount %s does not match payable amount %s", actualAmount, expectedAmount)
		}

		completedAt := time.Now().UTC()
		completedTransaction, err := saveVerifiedGatewayTransaction(repos.Payment, existingTransaction, completedAt, payment.Transaction{
			OrderID:          o.ID,
			TransactionID:    input.TransactionID,
			PaymentMethod:    input.PaymentMethod,
			AmountMinor:      input.Amount.AmountMinor(),
			Currency:         inputCurrency,
			Status:           "completed",
			GatewayResponse:  input.GatewayResponse,
			LiabilityShifted: input.LiabilityShifted,
			CompletedAt:      &completedAt,
		})
		if err != nil {
			return err
		}
		if reason, ok := latePaymentReviewReason(o.Status); ok {
			return s.createLatePaymentReview(
				repos,
				o.ID,
				completedTransaction.ID,
				input.TransactionID,
				o.OrderNumber,
				reason,
			)
		}
		if err := repos.Order.UpdatePaymentStatus(o.ID, "paid"); err != nil {
			return err
		}
		requiresLiabilityReview, orderTotalUSD, err := highValueLiabilityReviewRequired(o, input.LiabilityShifted)
		if err != nil {
			return err
		}
		newOrderStatus := o.Status
		if requiresLiabilityReview {
			if err := repos.Order.MarkPaymentLiabilityReviewHold(o.ID); err != nil {
				return err
			}
			newOrderStatus = "needs_review"
			orderAmountMoney, amountErr := o.TotalMoney()
			if amountErr != nil {
				return amountErr
			}
			if err := createHighValueLiabilityReview(
				repos.Payment,
				o.ID,
				completedTransaction.ID,
				input.TransactionID,
				o.OrderNumber,
				orderAmountMoney,
				o.Currency,
				orderTotalUSD,
				input.LiabilityShifted,
			); err != nil {
				return err
			}
		} else if o.Status == "pending" || o.Status == "paid" {
			if err := repos.Order.UpdateStatus(o.ID, o.Status, "processing"); err != nil {
				return err
			}
			newOrderStatus = "processing"
		}
		if err := enqueueOrderPaidOutboxEvent(repos.Outbox, o, input, completedAt); err != nil {
			return err
		}
		if err := enqueueOrderPaymentSucceededDomainEvent(repos.Outbox, o, input, o.Status, newOrderStatus, completedAt); err != nil {
			return err
		}
		return enqueueVerifiedConversionOutboxEvent(repos.Outbox, repos.OrderAttribution, o, input, completedAt)
	})
	if err == nil && s.risk != nil {
		s.risk.RecordProviderSuccess(input.Provider)
	}
	return result, err
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
	inputCurrency := input.Amount.Currency().String()
	payload := outbox.VerifiedConversionPayload{
		OrderID:     o.ID,
		AmountMinor: input.Amount.AmountMinor(),
		Currency:    inputCurrency,
		VerifiedAt:  verifiedAt.UTC(),
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
		AmountMinor:          input.Amount.AmountMinor(),
		Currency:             input.Amount.Currency().String(),
		PaidAt:               paidAt.UTC(),
		CustomerEmail:        strings.TrimSpace(o.ShippingAddress.Email),
		CustomerName:         customerName,
		ShippingCountry:      strings.TrimSpace(o.ShippingAddress.Country),
	})
	if err != nil {
		return err
	}
	if err := repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%s", outbox.EventTypeOrderPaid, o.ID, input.TransactionID),
		EventType:     outbox.EventTypeOrderPaid,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprint(o.ID),
		Payload:       datatypes.JSON(payload),
	}); err != nil {
		return err
	}
	return enqueueReferralOrderPaidOutboxEvent(repo, o, paidAt, input.TransactionID)
}

func saveVerifiedGatewayTransaction(repo *repository.PaymentRepository, existing *payment.Transaction, completedAt time.Time, next payment.Transaction) (*payment.Transaction, error) {
	if existing == nil {
		if err := repo.CreateTransaction(&next); err != nil {
			return nil, err
		}
		return &next, nil
	}
	existing.OrderID = next.OrderID
	existing.PaymentMethod = next.PaymentMethod
	existing.AmountMinor = next.AmountMinor
	existing.Currency = next.Currency
	existing.Status = next.Status
	existing.GatewayResponse = next.GatewayResponse
	existing.LiabilityShifted = next.LiabilityShifted
	existing.ErrorMessage = ""
	existing.CompletedAt = &completedAt
	if err := repo.UpdateTransaction(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func createDuplicatePaidRefundInTx(
	repos repository.TxRepositories,
	orderRecord *order.Order,
	duplicateTransaction *payment.Transaction,
) (*payment.Refund, error) {
	if orderRecord == nil {
		return nil, errors.New("order is required for duplicate paid refund")
	}
	if duplicateTransaction == nil {
		return nil, errors.New("duplicate payment transaction is required for duplicate paid refund")
	}
	if duplicateTransaction.OrderID != orderRecord.ID {
		return nil, errors.New("duplicate payment transaction does not belong to order")
	}
	if duplicateTransaction.Status != payment.TransactionStatusDuplicatePaid {
		return nil, fmt.Errorf("transaction status %s cannot create a duplicate paid refund", duplicateTransaction.Status)
	}
	duplicateAmount, err := duplicateTransaction.AmountMoney()
	if err != nil {
		return nil, fmt.Errorf("invalid duplicate payment amount: %w", err)
	}
	if duplicateAmount.AmountMinor() <= 0 {
		return nil, errors.New("duplicate payment amount must be greater than zero")
	}

	existingRefund, err := repos.Payment.FindRefundByTransactionIDAndReasonForUpdate(
		duplicateTransaction.ID,
		duplicatePaidRefundReason,
	)
	if err == nil {
		return existingRefund, nil
	}
	if !repository.IsRecordNotFound(err) {
		return nil, err
	}

	refundDraft := &payment.Refund{}
	fxSnapshot, _, err := ensureRefundFXSnapshot(refundDraft, orderRecord, duplicateTransaction.Currency)
	if err != nil {
		return nil, err
	}
	refundAmount, err := duplicateAmount.FormatMajor()
	if err != nil {
		return nil, err
	}
	reservedAmountMinor, err := repos.Payment.SumRefundAmountMinorByTransactionID(
		duplicateTransaction.ID,
		"pending",
		"completed",
	)
	if err != nil {
		return nil, err
	}
	transactionMoney, moneyErr := duplicateTransaction.AmountMoney()
	if moneyErr != nil {
		return nil, moneyErr
	}
	reservedMoney, moneyErr := domainmoney.New(reservedAmountMinor, duplicateTransaction.Currency)
	if moneyErr != nil {
		return nil, moneyErr
	}
	remainingAmountMoney, remainingErr := subtractRefundAmounts(transactionMoney, reservedMoney)
	if remainingErr != nil {
		return nil, remainingErr
	}
	remainingAmount, moneyErr := remainingAmountMoney.FormatMajor()
	if moneyErr != nil {
		return nil, moneyErr
	}
	refundAmountMoney := duplicateAmount
	exceeds, compareErr := refundAmountExceedsInMinorUnits(refundAmountMoney, remainingAmountMoney)
	if compareErr != nil {
		return nil, compareErr
	}
	if exceeds {
		return nil, fmt.Errorf(
			"duplicate paid refund amount %s exceeds refundable amount %s",
			refundAmount,
			remainingAmount,
		)
	}
	if err := validateHistoricalRefundFXCap(
		fxSnapshot,
		duplicateTransaction,
		refundAmountMoney,
		reservedMoney,
	); err != nil {
		return nil, err
	}

	refund := &payment.Refund{
		OrderID:              orderRecord.ID,
		TransactionID:        duplicateTransaction.ID,
		Currency:             duplicateTransaction.Currency,
		AmountMinor:          refundAmountMoney.AmountMinor(),
		RequestedAmountMinor: refundAmountMoney.AmountMinor(),
		FXSnapshotData:       currencydomain.OrderFXSnapshotJSON(fxSnapshot),
		Reason:               duplicatePaidRefundReason,
		Status:               "pending",
		RefundedBy:           0,
		RefundID:             nil,
		GatewayResponse:      "",
		CompletedAt:          nil,
	}
	if err := repos.Payment.CreateRefund(refund); err != nil {
		return nil, err
	}
	if err := repos.Order.SetFulfillmentHold(orderRecord.ID, true); err != nil {
		return nil, err
	}
	return refund, nil
}

func (s *PaymentService) createLatePaymentReview(repos repository.TxRepositories, orderID, transactionID uint, paymentIntentID, orderNumber, reason string) error {
	review, err := repos.Payment.FindPendingPaymentReviewByPaymentIntentIDAndReason(paymentIntentID, reason)
	if err == nil {
		changed := false
		if review.OrderID == nil {
			review.OrderID = &orderID
			changed = true
		}
		if review.TransactionID == nil {
			review.TransactionID = &transactionID
			changed = true
		}
		if changed {
			if err := repos.Payment.UpdatePaymentReview(review); err != nil {
				return err
			}
		}
		return s.ensureLatePaymentRefundExecutionInTx(repos, orderID, transactionID, reason)
	}
	if !repository.IsRecordNotFound(err) {
		return err
	}

	if err := repos.Payment.CreatePaymentReview(&payment.PaymentReview{
		OrderID:         &orderID,
		TransactionID:   &transactionID,
		PaymentIntentID: paymentIntentID,
		Status:          "pending",
		Reason:          reason,
		Source:          "webhook",
		Notes:           fmt.Sprintf("Payment succeeded after order %s was marked %s. Review payment and refund before fulfillment.", orderNumber, latePaymentOrderStatus(reason)),
	}); err != nil {
		return err
	}
	return s.ensureLatePaymentRefundExecutionInTx(repos, orderID, transactionID, reason)
}

// ensureLatePaymentRefundExecutionInTx creates the full-amount refund intent
// immediately when a terminal order receives a verified late charge. The
// review remains pending for audit, while the provider request is dispatched
// asynchronously through the normal refund execution Outbox workflow.
func (s *PaymentService) ensureLatePaymentRefundExecutionInTx(
	repos repository.TxRepositories,
	orderID, transactionID uint,
	reason string,
) error {
	if s == nil || repos.Payment == nil || repos.Order == nil {
		return errors.New("late payment refund dependencies are not configured")
	}
	transaction, err := repos.Payment.FindTransactionByIDForUpdate(transactionID)
	if err != nil {
		return err
	}
	orderRecord, err := repos.Order.FindByIDForUpdate(orderID)
	if err != nil {
		return normalizeOrderError(err)
	}
	return s.createLatePaymentRefundInTx(repos, orderRecord, transaction, latePaymentRefundReason(reason))
}

// createLatePaymentRefundInTx turns an operator-approved late payment into a
// durable refund intent. Provider execution remains in the normal refund
// execution workflow and is therefore retryable and auditable.
func (s *PaymentService) createLatePaymentRefundInTx(
	repos repository.TxRepositories,
	o *order.Order,
	transaction *payment.Transaction,
	reason string,
) error {
	if o == nil || transaction == nil {
		return errors.New("late payment refund requires order and transaction")
	}
	if transaction.Status != "completed" {
		return fmt.Errorf("late payment transaction status %s is not refundable", transaction.Status)
	}
	transactionAmount, err := transaction.AmountMoney()
	if err != nil {
		return err
	}
	if transactionAmount.AmountMinor() <= 0 {
		return errors.New("late payment refund amount must be greater than zero")
	}
	refund, err := repos.Payment.FindRefundByTransactionIDAndReasonForUpdate(transaction.ID, reason)
	if err == nil {
		if refund.Status != "pending" {
			return nil
		}
		return s.enqueueLatePaymentRefundExecutionInTx(repos, refund, transaction)
	} else if !repository.IsRecordNotFound(err) {
		return err
	}
	draft := &payment.Refund{}
	fxSnapshot, _, err := ensureRefundFXSnapshot(draft, o, transaction.Currency)
	if err != nil {
		return err
	}
	refund = &payment.Refund{
		OrderID:              o.ID,
		TransactionID:        transaction.ID,
		Currency:             transaction.Currency,
		AmountMinor:          transactionAmount.AmountMinor(),
		RequestedAmountMinor: transactionAmount.AmountMinor(),
		Reason:               reason,
		Status:               "pending",
		FXSnapshotData:       currencydomain.OrderFXSnapshotJSON(fxSnapshot),
	}
	if err := repos.Payment.CreateRefund(refund); err != nil {
		return err
	}
	if err := repos.Order.SetFulfillmentHold(o.ID, true); err != nil {
		return err
	}
	return s.enqueueLatePaymentRefundExecutionInTx(repos, refund, transaction)
}

func (s *PaymentService) enqueueLatePaymentRefundExecutionInTx(
	repos repository.TxRepositories,
	refund *payment.Refund,
	transaction *payment.Transaction,
) error {
	// Lightweight test repositories may omit the execution projection. The
	// durable intent is still created; production wiring always includes both
	// repositories and therefore gets automatic provider execution.
	if repos.RefundExecution == nil || repos.Outbox == nil {
		return nil
	}
	existingExecution, executionErr := repos.RefundExecution.FindByRefundIDForUpdate(refund.ID)
	if executionErr == nil {
		if existingExecution.Status == payment.PaymentRefundExecutionStatusProcessing ||
			existingExecution.Status == payment.PaymentRefundExecutionStatusSucceeded {
			return nil
		}
	} else if !repository.IsRecordNotFound(executionErr) {
		return executionErr
	}
	plan, err := s.beginPendingRefundExecutionInTx(repos, RequestPendingRefundExecutionInput{
		RefundID: refund.ID,
		Provider: transaction.PaymentMethod,
	}, time.Now().UTC())
	if err != nil {
		return err
	}
	return enqueuePaymentRefundExecutionRequestedOutboxEvent(
		repos.Outbox,
		plan.Execution,
		transaction.PaymentMethod,
		time.Now().UTC(),
	)
}

func highValueLiabilityReviewRequired(o *order.Order, liabilityShifted *bool) (bool, domainmoney.Money, error) {
	if o == nil {
		return false, domainmoney.Money{}, errors.New("order is required for liability review")
	}
	if liabilityShifted != nil && *liabilityShifted {
		return false, domainmoney.Money{}, nil
	}

	snapshot, err := currencydomain.ParseOrderFXSnapshot(o.FXSnapshotData)
	if err == nil {
		totalMoney, totalErr := o.TotalMoney()
		if totalErr != nil {
			return false, domainmoney.Money{}, totalErr
		}
		evaluation, evaluationErr := order.EvaluateHighValueOrder(totalMoney, snapshot)
		if evaluationErr != nil {
			return false, domainmoney.Money{}, evaluationErr
		}
		return evaluation.IsHighValue, evaluation.OrderTotalUSD, nil
	}

	// Legacy USD orders may predate the persisted FX snapshot. USD is already
	// the policy base currency, so applying the threshold directly is safe.
	if currency.NormalizeCode(o.Currency) == currencydomain.DefaultPrimaryCurrency {
		totalMoney, totalErr := o.TotalMoney()
		if totalErr != nil {
			return false, domainmoney.Money{}, totalErr
		}
		return totalMoney.AmountMinor() >= order.HighValueSignatureThresholdUSDMinor, totalMoney, nil
	}

	// A non-USD order without its immutable snapshot cannot be proven below the
	// threshold. Hold it conservatively instead of allowing an unknown-risk
	// payment to reach fulfillment.
	return true, domainmoney.Money{}, nil
}

func createHighValueLiabilityReview(
	repo *repository.PaymentRepository,
	orderID, transactionID uint,
	transactionReference, orderNumber string,
	orderAmount domainmoney.Money,
	orderCurrency string,
	orderTotalUSD domainmoney.Money,
	liabilityShifted *bool,
) error {
	review, err := repo.FindPendingPaymentReviewByOrderIDAndReasonForUpdate(orderID, highValueLiabilityReviewReason)
	if err == nil {
		if review.TransactionID == nil {
			review.TransactionID = &transactionID
		}
		if review.Notes == "" {
			review.Notes = highValueLiabilityReviewNotes(
				orderNumber,
				transactionReference,
				orderAmount,
				orderCurrency,
				orderTotalUSD,
				liabilityShifted,
			)
		}
		return repo.UpdatePaymentReview(review)
	}
	if !repository.IsRecordNotFound(err) {
		return err
	}

	return repo.CreatePaymentReview(&payment.PaymentReview{
		OrderID:         &orderID,
		TransactionID:   &transactionID,
		PaymentIntentID: transactionReference,
		Status:          "pending",
		Reason:          highValueLiabilityReviewReason,
		Source:          "gateway_verification",
		Notes: highValueLiabilityReviewNotes(
			orderNumber,
			transactionReference,
			orderAmount,
			orderCurrency,
			orderTotalUSD,
			liabilityShifted,
		),
	})
}

func highValueLiabilityReviewNotes(
	orderNumber,
	transactionReference string,
	orderAmount domainmoney.Money,
	orderCurrency string,
	orderTotalUSD domainmoney.Money,
	liabilityShifted *bool,
) string {
	liabilityState := "unknown"
	if liabilityShifted != nil {
		liabilityState = fmt.Sprintf("%t", *liabilityShifted)
	}
	orderAmountText, _ := orderAmount.FormatMajor()
	orderTotalUSDText, orderTotalUSDErr := orderTotalUSD.FormatMajor()
	if orderTotalUSDErr == nil && orderTotalUSD.AmountMinor() > 0 {
		return fmt.Sprintf(
			"Order %s payment %s requires manual identity or wire-transfer verification before fulfillment: order_total=%s %s, order_total_usd=%s, liability_shifted=%s.",
			orderNumber,
			transactionReference,
			currency.NormalizeCode(orderCurrency),
			orderAmountText,
			orderTotalUSDText,
			liabilityState,
		)
	}
	return fmt.Sprintf(
		"Order %s payment %s requires manual identity or wire-transfer verification before fulfillment: order_total=%s %s, order_total_usd=unknown, liability_shifted=%s.",
		orderNumber,
		transactionReference,
		currency.NormalizeCode(orderCurrency),
		orderAmountText,
		liabilityState,
	)
}

func latePaymentReviewReason(orderStatus string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(orderStatus)) {
	case "payment_expired":
		return "payment_succeeded_after_expiration", true
	case "cancelled":
		return "payment_succeeded_after_cancellation", true
	case "refunded":
		return "payment_succeeded_after_refund", true
	default:
		return "", false
	}
}

func latePaymentOrderStatus(reason string) string {
	switch reason {
	case "payment_succeeded_after_expiration":
		return "payment_expired"
	case "payment_succeeded_after_cancellation":
		return "cancelled"
	case "payment_succeeded_after_refund":
		return "refunded"
	default:
		return "a terminal state"
	}
}

func normalizePaymentCurrency(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

// orderPaymentSettlement returns the immutable channel settlement amount
// captured during checkout. Payment arithmetic stays in Money; persistence
// and gateway adapters may serialize it at their boundaries.
func orderPaymentSettlement(o *order.Order) (domainmoney.Money, error) {
	if o == nil {
		return domainmoney.Money{}, errors.New("order is required")
	}
	value := currency.NormalizeCode(o.PaymentCurrency)
	if value == "" {
		return domainmoney.Money{}, errors.New("order payment currency snapshot is required")
	}
	if !currency.IsValidCode(value) || !currency.IsCatalogCode(value) {
		return domainmoney.Money{}, errors.New("order payment currency is not configured")
	}
	money, err := o.PaymentMoney()
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("invalid order payment amount: %w", err)
	}
	if money.AmountMinor() <= 0 {
		return domainmoney.Money{}, errors.New("order payment amount must be greater than zero")
	}
	return money, nil
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

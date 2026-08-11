package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
)

var ErrPaymentRefundExecutionInProgress = errors.New("payment refund execution is already in progress")

const paymentRefundExecutionStaleAfter = 15 * time.Minute

type ExecutePendingRefundInput struct {
	RefundID uint
	AdminID  uint
	Provider string
	Gateway  pgateway.PaymentGateway
}

type pendingRefundExecutionPlan struct {
	Refund      *paymentdomain.Refund
	Transaction *paymentdomain.Transaction
	Order       *orderdomain.Order
	Execution   *paymentdomain.PaymentRefundExecution
}

func (s *PaymentService) ExecutePendingRefund(
	ctx context.Context,
	input ExecutePendingRefundInput,
) (*paymentdomain.Refund, *paymentdomain.PaymentRefundExecution, error) {
	if input.RefundID == 0 {
		return nil, nil, errors.New("refund id is required")
	}
	if input.AdminID == 0 {
		return nil, nil, errors.New("admin user id is required")
	}
	if input.Gateway == nil {
		return nil, nil, errors.New("payment gateway is required")
	}

	plan, err := s.beginPendingRefundExecution(input)
	if err != nil {
		return nil, nil, err
	}

	response, err := input.Gateway.RefundPaymentWithOptions(ctx, plan.Execution.ProviderTransactionID, plan.Refund.Amount, pgateway.RefundOptions{
		IdempotencyKey:        plan.Execution.IdempotencyKey,
		Reason:                plan.Refund.Reason,
		Currency:              plan.Transaction.Currency,
		OriginalAmount:        plan.Transaction.Amount,
		MerchantOrderNumber:   plan.Execution.MerchantOrderNumber,
		ProviderTransactionID: plan.Execution.ProviderTransactionID,
	})
	if err != nil {
		execution, failErr := s.failPendingRefundExecution(plan.Execution.RefundID, err.Error())
		if failErr != nil {
			return plan.Refund, plan.Execution, fmt.Errorf("gateway refund failed: %v; record failure failed: %w", err, failErr)
		}
		return plan.Refund, execution, err
	}
	if response == nil || strings.TrimSpace(response.ID) == "" {
		err := errors.New("payment gateway refund response is missing refund id")
		execution, failErr := s.failPendingRefundExecution(plan.Execution.RefundID, err.Error())
		if failErr != nil {
			return plan.Refund, plan.Execution, fmt.Errorf("%v; record failure failed: %w", err, failErr)
		}
		return plan.Refund, execution, err
	}

	refund, execution, err := s.completePendingRefundExecution(plan.Execution.RefundID, response)
	if err != nil {
		return plan.Refund, plan.Execution, err
	}
	return refund, execution, nil
}

func (s *PaymentService) beginPendingRefundExecution(input ExecutePendingRefundInput) (*pendingRefundExecutionPlan, error) {
	var plan *pendingRefundExecutionPlan
	now := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.RefundExecution == nil {
			return errors.New("payment refund execution repository is not configured")
		}
		refund, err := repos.Payment.FindRefundByIDForUpdate(input.RefundID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("refund not found")
			}
			return err
		}
		if refund.Status != "pending" {
			return errors.New("refund is not pending")
		}
		if refund.RefundID != nil && strings.TrimSpace(*refund.RefundID) != "" {
			return errors.New("refund already has provider refund id")
		}
		if refund.Amount <= 0 {
			return errors.New("refund amount must be greater than zero")
		}

		transaction, err := repos.Payment.FindTransactionByIDForUpdate(refund.TransactionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("transaction not found")
			}
			return err
		}
		if transaction.OrderID != refund.OrderID {
			return errors.New("refund transaction does not belong to order")
		}
		if transaction.Status != "completed" {
			return errors.New("transaction is not refundable")
		}
		providerTransactionID := strings.TrimSpace(transaction.TransactionID)
		if providerTransactionID == "" {
			return errors.New("provider transaction id is required for refund execution")
		}
		orderRecord, err := repos.Order.FindByIDForUpdate(refund.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		merchantOrderNumber := strings.TrimSpace(orderRecord.OrderNumber)
		if merchantOrderNumber == "" {
			return errors.New("merchant order number is required for refund execution")
		}
		if orderRecord.ID != transaction.OrderID {
			return errors.New("refund order does not belong to transaction")
		}
		provider, err := normalizeRefundExecutionProvider(transaction.PaymentMethod)
		if err != nil {
			return err
		}
		if requestedProvider := strings.ToLower(strings.TrimSpace(input.Provider)); requestedProvider != "" && requestedProvider != provider {
			return fmt.Errorf("refund provider %s does not match transaction provider %s", requestedProvider, provider)
		}
		fxSnapshot, snapshotNeedsPersistence, err := ensureRefundFXSnapshot(refund, orderRecord, transaction.Currency)
		if err != nil {
			return err
		}
		reservedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "pending", "completed")
		if err != nil {
			return err
		}
		reservedBeforeCurrent := reservedAmount - refund.Amount
		if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, refund.Amount, reservedBeforeCurrent); err != nil {
			return err
		}
		if snapshotNeedsPersistence {
			refund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
			if err := repos.Payment.UpdateRefund(refund); err != nil {
				return err
			}
		}

		execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refund.ID)
		if err != nil && !repository.IsRecordNotFound(err) {
			return err
		}
		if execution != nil && err == nil {
			if execution.Status == paymentdomain.PaymentRefundExecutionStatusSucceeded {
				return errors.New("refund execution already succeeded")
			}
			if execution.Status == paymentdomain.PaymentRefundExecutionStatusProcessing && !refundExecutionIsStale(execution, now) {
				return ErrPaymentRefundExecutionInProgress
			}
			execution.Status = paymentdomain.PaymentRefundExecutionStatusProcessing
			execution.AttemptCount++
			execution.RequestedByID = input.AdminID
			execution.RequestedAt = now
			execution.ProviderPaymentID = providerTransactionID
			execution.ErrorMessage = ""
			execution.ProviderRefundID = ""
			execution.ProviderStatus = ""
			execution.MerchantOrderNumber = merchantOrderNumber
			execution.ProviderTransactionID = providerTransactionID
			execution.GatewayResponseJSON = ""
			execution.CompletedAt = nil
			if err := repos.RefundExecution.Update(execution); err != nil {
				return err
			}
			plan = &pendingRefundExecutionPlan{Refund: refund, Transaction: transaction, Order: orderRecord, Execution: execution}
			return nil
		}

		execution = &paymentdomain.PaymentRefundExecution{
			RefundID:              refund.ID,
			OrderID:               refund.OrderID,
			TransactionID:         transaction.ID,
			Provider:              provider,
			ProviderPaymentID:     providerTransactionID,
			MerchantOrderNumber:   merchantOrderNumber,
			ProviderTransactionID: providerTransactionID,
			Amount:                refund.Amount,
			Currency:              transaction.Currency,
			Status:                paymentdomain.PaymentRefundExecutionStatusProcessing,
			IdempotencyKey:        refundExecutionIdempotencyKey(refund.ID),
			AttemptCount:          1,
			RequestedByID:         input.AdminID,
			RequestedAt:           now,
		}
		if err := repos.RefundExecution.Create(execution); err != nil {
			return err
		}
		plan = &pendingRefundExecutionPlan{Refund: refund, Transaction: transaction, Order: orderRecord, Execution: execution}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PaymentService) completePendingRefundExecution(
	refundID uint,
	response *pgateway.RefundResponse,
) (*paymentdomain.Refund, *paymentdomain.PaymentRefundExecution, error) {
	var completedRefund *paymentdomain.Refund
	var completedExecution *paymentdomain.PaymentRefundExecution
	now := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.RefundExecution == nil {
			return errors.New("payment refund execution repository is not configured")
		}
		refund, err := repos.Payment.FindRefundByIDForUpdate(refundID)
		if err != nil {
			return err
		}
		if refund.Status != "pending" {
			return errors.New("refund is not pending")
		}
		transaction, err := repos.Payment.FindTransactionByIDForUpdate(refund.TransactionID)
		if err != nil {
			return err
		}
		orderRecord, err := repos.Order.FindByIDForUpdate(refund.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refundID)
		if err != nil {
			return err
		}
		if execution.Status != paymentdomain.PaymentRefundExecutionStatusProcessing {
			return errors.New("refund execution is not processing")
		}
		if response.Amount > 0 && absRefundMoney(response.Amount-refund.Amount) > 0.01 {
			return fmt.Errorf("gateway refund amount %.2f does not match local refund amount %.2f", response.Amount, refund.Amount)
		}

		payload, err := json.Marshal(response)
		if err != nil {
			return err
		}
		providerRefundID := strings.TrimSpace(response.ID)
		refund.Status = "completed"
		refund.RefundID = &providerRefundID
		refund.GatewayResponse = string(payload)
		refund.CompletedAt = &now
		if err := repos.Payment.UpdateRefund(refund); err != nil {
			return err
		}
		if err := restoreRefundLineItemStock(repos, refund.LineItems, now); err != nil {
			return err
		}

		execution.Status = paymentdomain.PaymentRefundExecutionStatusSucceeded
		execution.ProviderRefundID = providerRefundID
		execution.ProviderStatus = strings.TrimSpace(response.Status)
		execution.GatewayResponseJSON = string(payload)
		execution.ErrorMessage = ""
		execution.CompletedAt = &now
		if err := repos.RefundExecution.Update(execution); err != nil {
			return err
		}

		completedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "completed")
		if err != nil {
			return err
		}
		if completedAmount >= transaction.Amount-0.01 {
			transaction.Status = "refunded"
			if err := repos.Payment.UpdateTransaction(transaction); err != nil {
				return err
			}
		}
		orderRefundedAmount, err := repos.Payment.SumRefundAmountByOrderID(orderRecord.ID, "completed")
		if err != nil {
			return err
		}
		if orderRefundedAmount >= orderRecord.TotalAmount-0.01 {
			if err := repos.Order.UpdatePaymentStatus(orderRecord.ID, "refunded"); err != nil {
				return err
			}
			if err := repos.Order.UpdateStatus(orderRecord.ID, "refunded"); err != nil {
				return err
			}
		}

		completedRefund = refund
		completedExecution = execution
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return completedRefund, completedExecution, nil
}

func (s *PaymentService) failPendingRefundExecution(refundID uint, message string) (*paymentdomain.PaymentRefundExecution, error) {
	var failedExecution *paymentdomain.PaymentRefundExecution
	now := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.RefundExecution == nil {
			return errors.New("payment refund execution repository is not configured")
		}
		execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refundID)
		if err != nil {
			return err
		}
		execution.Status = paymentdomain.PaymentRefundExecutionStatusFailed
		execution.ErrorMessage = strings.TrimSpace(message)
		execution.CompletedAt = &now
		if err := repos.RefundExecution.Update(execution); err != nil {
			return err
		}
		failedExecution = execution
		return nil
	})
	if err != nil {
		return nil, err
	}
	return failedExecution, nil
}

func normalizeRefundExecutionProvider(value string) (string, error) {
	provider, err := pgateway.ParseGatewayType(value)
	if err != nil {
		return "", err
	}
	return string(provider), nil
}

func refundExecutionIdempotencyKey(refundID uint) string {
	return fmt.Sprintf("rf_%d_v1", refundID)
}

func refundExecutionIsStale(execution *paymentdomain.PaymentRefundExecution, now time.Time) bool {
	timestamp := execution.UpdatedAt
	if timestamp.IsZero() {
		timestamp = execution.RequestedAt
	}
	return now.Sub(timestamp) > paymentRefundExecutionStaleAfter
}

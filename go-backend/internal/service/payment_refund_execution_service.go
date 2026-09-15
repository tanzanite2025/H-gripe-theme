package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	aftersalesdomain "commerce-platform/internal/domain/aftersales"
	currencydomain "commerce-platform/internal/domain/currency"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
)

var ErrPaymentRefundExecutionInProgress = errors.New("payment refund execution is already in progress")

const paymentRefundExecutionStaleAfter = 15 * time.Minute
const localZeroRefundStatus = "local_zero_refund"

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
	plan, err := s.beginPendingRefundExecution(input)
	if err != nil {
		return nil, nil, err
	}

	var response *pgateway.RefundResponse
	refundMoney, err := plan.Refund.AmountMoney()
	if err != nil {
		return plan.Refund, plan.Execution, err
	}
	if refundMoney.AmountMinor() <= 0 {
		response = &pgateway.RefundResponse{
			ID:        localZeroRefundID(plan.Refund.ID),
			PaymentID: plan.Execution.ProviderTransactionID,
			Amount:    0,
			Status:    localZeroRefundStatus,
			CreatedAt: time.Now().UTC(),
		}
	} else {
		refundMajor, majorErr := refundMoney.MajorFloat()
		if majorErr != nil {
			return plan.Refund, plan.Execution, majorErr
		}
		originalMoney, originalErr := plan.Transaction.AmountMoney()
		if originalErr != nil {
			return plan.Refund, plan.Execution, originalErr
		}
		originalMajor, originalErr := originalMoney.MajorFloat()
		if originalErr != nil {
			return plan.Refund, plan.Execution, originalErr
		}
		response, err = input.Gateway.RefundPaymentWithOptions(ctx, plan.Execution.ProviderTransactionID, refundMajor, pgateway.RefundOptions{
			IdempotencyKey:        plan.Execution.IdempotencyKey,
			Reason:                plan.Refund.Reason,
			Currency:              plan.Transaction.Currency,
			OriginalAmount:        originalMajor,
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
		if refund.Status != "pending" && refund.Status != "failed" {
			return errors.New("refund is not pending")
		}
		// A provider failure is a retryable local outcome. A new execution
		// attempt reopens the intent while retaining the failed execution row
		// as the audit trail.
		if refund.Status == "failed" {
			refund.Status = "pending"
			// The failed webhook may have associated a provider refund id with
			// this intent for idempotency. It is not a successful money movement,
			// so clear it before issuing a new provider request. The previous id
			// remains on the execution row as failure audit data.
			refund.RefundID = nil
			refund.CompletedAt = nil
			if err := repos.Payment.UpdateRefund(refund); err != nil {
				return err
			}
		}
		if refund.RefundID != nil && strings.TrimSpace(*refund.RefundID) != "" {
			return errors.New("refund already has provider refund id")
		}
		refundMoney, err := refund.AmountMoney()
		if err != nil {
			return fmt.Errorf("refund amount: %w", err)
		}
		if refundMoney.AmountMinor() > 0 && input.Gateway == nil {
			return errors.New("payment gateway is required for a positive refund amount")
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
		if !isRefundableGatewayTransactionStatus(transaction.Status) {
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
		reservedMoney, err := parseRefundMoney(reservedAmount, transaction.Currency)
		if err != nil {
			return err
		}
		refundMoney, err = parseRefundMoney(refund.Amount, transaction.Currency)
		if err != nil {
			return err
		}
		reservedBeforeCurrentMoney, err := subtractRefundAmounts(reservedMoney, refundMoney)
		if err != nil {
			return err
		}
		if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, refundMoney, reservedBeforeCurrentMoney); err != nil {
			return err
		}
		if snapshotNeedsPersistence {
			refund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
			if err := repos.Payment.UpdateRefund(refund); err != nil {
				return err
			}
		}
		if err := prepareRefundLoyaltySettlementInTx(repos, orderRecord, refund); err != nil {
			return err
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
			AmountMinor:           refundMoney.AmountMinor(),
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
	var affectedProductIDs []uint
	var amountMismatchError error
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
		amountMatches := true
		if response.Amount > 0 {
			responseMoney, moneyErr := parseRefundMoney(response.Amount, transaction.Currency)
			if moneyErr != nil {
				return moneyErr
			}
			refundMoney, moneyErr := parseRefundMoney(refund.Amount, transaction.Currency)
			if moneyErr != nil {
				return moneyErr
			}
			amountMatches, err = refundAmountsEqualInMinorUnits(responseMoney, refundMoney)
			if err != nil {
				return err
			}
		}
		if !amountMatches {
			payload, marshalErr := json.Marshal(response)
			if marshalErr != nil {
				return marshalErr
			}
			execution.Status = paymentdomain.PaymentRefundExecutionStatusFailed
			execution.ProviderRefundID = strings.TrimSpace(response.ID)
			execution.ProviderStatus = strings.TrimSpace(response.Status)
			execution.GatewayResponseJSON = string(payload)
			execution.ErrorMessage = fmt.Sprintf(
				"provider refund amount %.2f does not match local net refund amount %.2f; manual reconciliation required",
				response.Amount,
				refund.Amount,
			)
			execution.CompletedAt = &now
			if err := repos.RefundExecution.Update(execution); err != nil {
				return err
			}
			if err := enqueuePaymentRefundFailedOutboxEvent(
				repos.Outbox,
				refund,
				transaction.Currency,
				execution.Provider,
				execution.ProviderRefundID,
				execution.Status,
				execution.ErrorMessage,
				execution.AttemptCount,
				now,
			); err != nil {
				return err
			}
			amountMismatchError = errors.New(execution.ErrorMessage)
			return nil
		}
		if err := finalizeRefundLoyaltySettlementInTx(repos, orderRecord, refund); err != nil {
			return err
		}
		if err := restoreGiftCardRefundInTx(
			repos.Coupon,
			refund.OrderID,
			refund.ID,
			refund.GiftCardRefundAmount,
			"gateway refund completed",
		); err != nil {
			return err
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
		productIDs, err := restoreRefundLineItemStock(repos, refund.LineItems, now)
		if err != nil {
			return err
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
		if err := s.enqueueProductCacheInvalidationInTx(repos, productIDs, "refund stock restored"); err != nil {
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
		if err := enqueuePaymentRefundCompletedOutboxEvent(
			repos.Outbox,
			refund,
			transaction.Currency,
			execution.Provider,
			providerRefundID,
			execution.Status,
			now,
		); err != nil {
			return err
		}
		if err := completeLinkedAfterSalesCaseInTx(repos, refund, execution.RequestedByID); err != nil {
			return err
		}

		completedAmount, err := repos.Payment.SumRefundTotalAmountByTransactionID(transaction.ID, transaction.Currency, "completed")
		if err != nil {
			return err
		}
		giftCardPaymentAmount, err := sumGiftCardUsageForOrderInTx(repos.Coupon, orderRecord.ID)
		if err != nil {
			return err
		}
		transactionMoney, err := transaction.AmountMoney()
		if err != nil {
			return err
		}
		giftCardPaymentMoney, err := parseRefundMoney(giftCardPaymentAmount, transaction.Currency)
		if err != nil {
			return err
		}
		transactionRefundTarget, err := addRefundAmounts(transactionMoney, giftCardPaymentMoney)
		if err != nil {
			return err
		}
		completedMoney, err := parseRefundMoney(completedAmount, transaction.Currency)
		if err != nil {
			return err
		}
		transactionFullyRefunded, compareErr := refundAmountAtLeastInMinorUnits(completedMoney, transactionRefundTarget)
		if compareErr != nil {
			return compareErr
		}
		if transactionFullyRefunded && transaction.Status != paymentdomain.TransactionStatusDuplicatePaid {
			transaction.Status = "refunded"
			if err := repos.Payment.UpdateTransaction(transaction); err != nil {
				return err
			}
		}
		if !isDuplicatePaidRefund(refund) {
			orderRefundedAmount, err := repos.Payment.SumRefundTotalAmountByOrderID(orderRecord.ID, orderRecord.Currency, "completed")
			if err != nil {
				return err
			}
			orderRefundedMoney, err := parseRefundMoney(orderRefundedAmount, orderRecord.Currency)
			if err != nil {
				return err
			}
			orderTotalMoney, err := orderRecord.TotalMoney()
			if err != nil {
				return err
			}
			orderFullyRefunded, compareErr := refundAmountAtLeastInMinorUnits(orderRefundedMoney, orderTotalMoney)
			if compareErr != nil {
				return compareErr
			}
			if orderFullyRefunded {
				if err := repos.Order.UpdatePaymentStatus(orderRecord.ID, "refunded"); err != nil {
					return err
				}
				if err := repos.Order.UpdateStatus(orderRecord.ID, orderRecord.Status, "refunded"); err != nil {
					return err
				}
				if err := enqueueReferralOrderInvalidatedOutboxEvent(
					repos.Outbox,
					orderRecord.ID,
					now,
					"order fully refunded",
					"refund_execution",
					fmt.Sprint(refund.ID),
				); err != nil {
					return err
				}
			}
		}

		completedRefund = refund
		completedExecution = execution
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if amountMismatchError != nil {
		return nil, nil, amountMismatchError
	}
	s.invalidateProductCacheAfterStockCommit(affectedProductIDs)
	return completedRefund, completedExecution, nil
}

func (s *PaymentService) failPendingRefundExecution(refundID uint, message string) (*paymentdomain.PaymentRefundExecution, error) {
	var failedExecution *paymentdomain.PaymentRefundExecution
	now := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.RefundExecution == nil {
			return errors.New("payment refund execution repository is not configured")
		}
		refund, err := repos.Payment.FindRefundByIDForUpdate(refundID)
		if err != nil {
			return err
		}
		orderRecord, err := repos.Order.FindByIDForUpdate(refund.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if err := releaseRefundLoyaltyReservationInTx(repos, orderRecord, refund); err != nil {
			return err
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
		if err := enqueuePaymentRefundFailedOutboxEvent(
			repos.Outbox,
			refund,
			execution.Currency,
			execution.Provider,
			execution.ProviderRefundID,
			execution.Status,
			execution.ErrorMessage,
			execution.AttemptCount,
			now,
		); err != nil {
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

func completeLinkedAfterSalesCaseInTx(
	repos repository.TxRepositories,
	refund *paymentdomain.Refund,
	updatedBy uint,
) error {
	if refund == nil || repos.AfterSalesRefund == nil || repos.AfterSalesCase == nil {
		return nil
	}

	review, err := repos.AfterSalesRefund.FindByLinkedRefundIDForUpdate(refund.ID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if review == nil || review.CaseID == 0 {
		return nil
	}

	caseRecord, err := repos.AfterSalesCase.FindByIDForUpdate(review.CaseID)
	if err != nil {
		return err
	}
	if caseRecord.OrderID != refund.OrderID {
		return fmt.Errorf("after-sales case %d does not belong to refund order", caseRecord.ID)
	}
	if caseRecord.Status != aftersalesdomain.StatusResolving {
		return nil
	}
	if updatedBy == 0 {
		updatedBy = refund.RefundedBy
	}
	_, err = repos.AfterSalesCase.UpdateStatusIfCurrentInTx(
		caseRecord.ID,
		aftersalesdomain.StatusResolving,
		aftersalesdomain.StatusCompleted,
		"退款执行完成",
		updatedBy,
	)
	return err
}

// markRefundExecutionSucceededInTx reconciles the local execution attempt
// from a verified provider webhook. The provider may have completed the
// refund even after the synchronous admin request timed out and marked the
// attempt failed, so a webhook must be allowed to move either processing or
// failed executions to succeeded. Older/external refunds may not have an
// execution row; those are still valid and are handled by the caller.
func markRefundExecutionSucceededInTx(
	repos repository.TxRepositories,
	refund *paymentdomain.Refund,
	providerRefundID string,
	providerStatus string,
	gatewayResponse string,
	completedAt time.Time,
) (*paymentdomain.PaymentRefundExecution, error) {
	if refund == nil || repos.RefundExecution == nil {
		return nil, nil
	}

	execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refund.ID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	providerRefundID = strings.TrimSpace(providerRefundID)
	if providerRefundID == "" {
		return nil, errors.New("provider refund id is required to reconcile refund execution")
	}
	providerStatus = strings.TrimSpace(providerStatus)
	if providerStatus == "" {
		providerStatus = "succeeded"
	}
	if strings.TrimSpace(gatewayResponse) == "" {
		gatewayResponse = execution.GatewayResponseJSON
	}
	if execution.Status == paymentdomain.PaymentRefundExecutionStatusSucceeded &&
		execution.ProviderRefundID == providerRefundID &&
		strings.TrimSpace(execution.ErrorMessage) == "" && execution.CompletedAt != nil {
		return execution, nil
	}

	execution.Status = paymentdomain.PaymentRefundExecutionStatusSucceeded
	execution.ProviderRefundID = providerRefundID
	execution.ProviderStatus = providerStatus
	execution.GatewayResponseJSON = gatewayResponse
	execution.ErrorMessage = ""
	execution.CompletedAt = &completedAt
	if err := repos.RefundExecution.Update(execution); err != nil {
		return nil, err
	}
	return execution, nil
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

func localZeroRefundID(refundID uint) string {
	return fmt.Sprintf("local_zero_refund_%d", refundID)
}

func refundExecutionIsStale(execution *paymentdomain.PaymentRefundExecution, now time.Time) bool {
	timestamp := execution.UpdatedAt
	if timestamp.IsZero() {
		timestamp = execution.RequestedAt
	}
	return now.Sub(timestamp) > paymentRefundExecutionStaleAfter
}

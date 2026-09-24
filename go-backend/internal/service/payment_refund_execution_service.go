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
	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/repository"
)

var ErrPaymentRefundExecutionInProgress = errors.New("payment refund execution is already in progress")

const paymentRefundExecutionStaleAfter = 15 * time.Minute
const localZeroRefundStatus = "local_zero_refund"

type pendingRefundExecutionPlan struct {
	Refund      *paymentdomain.Refund
	Transaction *paymentdomain.Transaction
	Order       *orderdomain.Order
	Execution   *paymentdomain.PaymentRefundExecution
}

func (s *PaymentService) executePendingRefundPlan(
	ctx context.Context,
	plan *pendingRefundExecutionPlan,
	gateway pgateway.PaymentGateway,
) (*paymentdomain.Refund, *paymentdomain.PaymentRefundExecution, error) {
	if plan == nil || plan.Refund == nil || plan.Execution == nil || plan.Transaction == nil {
		return nil, nil, errors.New("payment refund execution plan is required")
	}
	var response *pgateway.RefundResponse
	refundMoney, err := plan.Refund.AmountMoney()
	if err != nil {
		return plan.Refund, plan.Execution, err
	}
	if refundMoney.AmountMinor() <= 0 {
		response = &pgateway.RefundResponse{
			ID:          localZeroRefundID(plan.Refund.ID),
			PaymentID:   plan.Execution.ProviderTransactionID,
			AmountMinor: 0,
			Status:      localZeroRefundStatus,
			CreatedAt:   time.Now().UTC(),
		}
	} else {
		if gateway == nil {
			return plan.Refund, plan.Execution, errors.New("payment gateway is required for a positive refund amount")
		}
		originalMoney, originalErr := plan.Transaction.AmountMoney()
		if originalErr != nil {
			return plan.Refund, plan.Execution, originalErr
		}
		response, err = gateway.RefundPaymentWithOptions(ctx, plan.Execution.ProviderTransactionID, refundMoney.AmountMinor(), pgateway.RefundOptions{
			IdempotencyKey:        plan.Execution.IdempotencyKey,
			AmountMinor:           refundMoney.AmountMinor(),
			OriginalAmountMinor:   originalMoney.AmountMinor(),
			Reason:                plan.Refund.Reason,
			Currency:              plan.Transaction.Currency,
			MerchantOrderNumber:   plan.Execution.MerchantOrderNumber,
			ProviderTransactionID: plan.Execution.ProviderTransactionID,
		})
		if err != nil {
			if errors.Is(err, resilience.ErrExternalOutcomeUnknown) {
				return plan.Refund, plan.Execution, err
			}
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

type RequestPendingRefundExecutionInput struct {
	RefundID uint
	AdminID  uint
	Provider string
}

// RequestPendingRefundExecution persists the local execution intent and an
// outbox command atomically. Gateway credentials and network calls stay out of
// the request transaction.
func (s *PaymentService) RequestPendingRefundExecution(
	ctx context.Context,
	input RequestPendingRefundExecutionInput,
) (*paymentdomain.Refund, *paymentdomain.PaymentRefundExecution, error) {
	if input.RefundID == 0 {
		return nil, nil, errors.New("refund id is required")
	}
	if input.AdminID == 0 {
		return nil, nil, errors.New("admin user id is required")
	}
	var plan *pendingRefundExecutionPlan
	_, err := s.beginPendingRefundExecution(RequestPendingRefundExecutionInput{
		RefundID: input.RefundID,
		AdminID:  input.AdminID,
		Provider: input.Provider,
	}, func(repos repository.TxRepositories, pending *pendingRefundExecutionPlan) error {
		plan = pending
		return enqueuePaymentRefundExecutionRequestedOutboxEvent(
			repos.Outbox,
			pending.Execution,
			input.Provider,
			time.Now().UTC(),
		)
	})
	if err != nil {
		return nil, nil, err
	}
	return plan.Refund, plan.Execution, nil
}

func (s *PaymentService) beginPendingRefundExecution(
	input RequestPendingRefundExecutionInput,
	onCommitted func(repository.TxRepositories, *pendingRefundExecutionPlan) error,
) (*pendingRefundExecutionPlan, error) {
	var plan *pendingRefundExecutionPlan
	now := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var err error
		plan, err = s.beginPendingRefundExecutionInTx(repos, input, now)
		if err != nil {
			return err
		}
		if onCommitted != nil {
			return onCommitted(repos, plan)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// beginPendingRefundExecutionInTx prepares a provider refund attempt while the
// caller's business transaction is still open. This is used by both the admin
// retry endpoint and the automatic late-payment refund path so the refund
// intent, execution row, and Outbox command commit atomically.
func (s *PaymentService) beginPendingRefundExecutionInTx(
	repos repository.TxRepositories,
	input RequestPendingRefundExecutionInput,
	now time.Time,
) (*pendingRefundExecutionPlan, error) {
	if repos.RefundExecution == nil {
		return nil, errors.New("payment refund execution repository is not configured")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var plan *pendingRefundExecutionPlan
	refund, err := repos.Payment.FindRefundByIDForUpdate(input.RefundID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, errors.New("refund not found")
		}
		return nil, err
	}
	if refund.Status != "pending" && refund.Status != "failed" {
		return nil, errors.New("refund is not pending")
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
			return nil, err
		}
	}
	if refund.RefundID != nil && strings.TrimSpace(*refund.RefundID) != "" {
		return nil, errors.New("refund already has provider refund id")
	}
	refundMoney, err := refund.AmountMoney()
	if err != nil {
		return nil, fmt.Errorf("refund amount: %w", err)
	}
	transaction, err := repos.Payment.FindTransactionByIDForUpdate(refund.TransactionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	if transaction.OrderID != refund.OrderID {
		return nil, errors.New("refund transaction does not belong to order")
	}
	if !isRefundableGatewayTransactionStatus(transaction.Status) {
		return nil, errors.New("transaction is not refundable")
	}
	providerTransactionID := strings.TrimSpace(transaction.TransactionID)
	if providerTransactionID == "" {
		return nil, errors.New("provider transaction id is required for refund execution")
	}
	orderRecord, err := repos.Order.FindByIDForUpdate(refund.OrderID)
	if err != nil {
		return nil, normalizeOrderError(err)
	}
	merchantOrderNumber := strings.TrimSpace(orderRecord.OrderNumber)
	if merchantOrderNumber == "" {
		return nil, errors.New("merchant order number is required for refund execution")
	}
	if orderRecord.ID != transaction.OrderID {
		return nil, errors.New("refund order does not belong to transaction")
	}
	provider, err := normalizeRefundExecutionProvider(transaction.PaymentMethod)
	if err != nil {
		return nil, err
	}
	if requestedProvider := strings.ToLower(strings.TrimSpace(input.Provider)); requestedProvider != "" && requestedProvider != provider {
		return nil, fmt.Errorf("refund provider %s does not match transaction provider %s", requestedProvider, provider)
	}
	fxSnapshot, snapshotNeedsPersistence, err := ensureRefundFXSnapshot(refund, orderRecord, transaction.Currency)
	if err != nil {
		return nil, err
	}
	reservedAmountMinor, err := repos.Payment.SumRefundAmountMinorByTransactionID(transaction.ID, "pending", "completed")
	if err != nil {
		return nil, err
	}
	reservedMoney, err := domainmoney.New(reservedAmountMinor, transaction.Currency)
	if err != nil {
		return nil, err
	}
	refundMoney, err = refund.AmountMoney()
	if err != nil {
		return nil, err
	}
	reservedBeforeCurrentMoney, err := subtractRefundAmounts(reservedMoney, refundMoney)
	if err != nil {
		return nil, err
	}
	if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, refundMoney, reservedBeforeCurrentMoney); err != nil {
		return nil, err
	}
	if snapshotNeedsPersistence {
		refund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
		if err := repos.Payment.UpdateRefund(refund); err != nil {
			return nil, err
		}
	}
	if err := prepareRefundLoyaltySettlementInTx(repos, orderRecord, refund); err != nil {
		return nil, err
	}

	execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refund.ID)
	if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if execution != nil && err == nil {
		if execution.Status == paymentdomain.PaymentRefundExecutionStatusSucceeded {
			return nil, errors.New("refund execution already succeeded")
		}
		if execution.Status == paymentdomain.PaymentRefundExecutionStatusProcessing && !refundExecutionIsStale(execution, now) {
			return nil, ErrPaymentRefundExecutionInProgress
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
			return nil, err
		}
		plan = &pendingRefundExecutionPlan{Refund: refund, Transaction: transaction, Order: orderRecord, Execution: execution}
		return plan, nil
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
		Currency:              transaction.Currency,
		Status:                paymentdomain.PaymentRefundExecutionStatusProcessing,
		IdempotencyKey:        refundExecutionIdempotencyKey(refund.ID),
		AttemptCount:          1,
		RequestedByID:         input.AdminID,
		RequestedAt:           now,
	}
	if err := repos.RefundExecution.Create(execution); err != nil {
		return nil, err
	}
	plan = &pendingRefundExecutionPlan{Refund: refund, Transaction: transaction, Order: orderRecord, Execution: execution}
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
		refundMoney, moneyErr := refund.AmountMoney()
		if moneyErr != nil {
			return moneyErr
		}
		amountMatches := true
		// Some providers report the amount in the merchant settlement
		// currency when the customer paid in a non-USD currency. The customer
		// refund amount is already guarded against the historical FX cap before
		// the gateway call; a different settlement amount is an FX fact to
		// record, not a reason to retry an already successful refund.
		settlementCurrency := strings.ToUpper(strings.TrimSpace(response.SettlementCurrency))
		compareProviderAmount := response.AmountMinor > 0 &&
			(settlementCurrency == "" || strings.EqualFold(settlementCurrency, transaction.Currency))
		var responseMoney domainmoney.Money
		if compareProviderAmount {
			responseMoney, moneyErr = domainmoney.New(response.AmountMinor, transaction.Currency)
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
				"provider refund amount %s does not match local net refund amount %s; manual reconciliation required",
				formatRefundMoney(responseMoney),
				formatRefundMoney(refundMoney),
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

		payload, err := json.Marshal(response)
		if err != nil {
			return err
		}
		providerRefundID := strings.TrimSpace(response.ID)
		refund.Status = "completed"
		refund.RefundID = &providerRefundID
		refund.GatewayResponse = string(payload)
		refund.CompletedAt = &now
		if response.SettlementAmountMinor != 0 && strings.TrimSpace(response.SettlementCurrency) != "" {
			fxSnapshot, _, snapshotErr := ensureRefundFXSnapshot(refund, orderRecord, transaction.Currency)
			if snapshotErr != nil {
				return snapshotErr
			}
			refundMoney, moneyErr := refund.AmountMoney()
			if moneyErr != nil {
				return moneyErr
			}
			if err := applyRefundSettlementFacts(
				refund,
				fxSnapshot,
				refundMoney,
				response.SettlementAmountMinor,
				response.SettlementCurrency,
				response.SettlementBalanceTransactionID,
			); err != nil {
				return err
			}
		}
		if err := repos.Payment.UpdateRefund(refund); err != nil {
			return err
		}
		if refundCanRestockPhysicalItems(orderRecord) {
			productIDs, err := restoreRefundLineItemStock(repos, orderRecord, refund.LineItems, now)
			if err != nil {
				return err
			}
			affectedProductIDs = append(affectedProductIDs, productIDs...)
			if err := s.enqueueProductCacheInvalidationInTx(repos, productIDs, "refund stock restored"); err != nil {
				return err
			}
		}

		execution.Status = paymentdomain.PaymentRefundExecutionStatusSucceeded
		execution.ProviderRefundID = providerRefundID
		execution.ProviderStatus = strings.TrimSpace(response.Status)
		execution.SettlementAmountMinor = refund.SettlementAmountMinor
		execution.SettlementCurrency = refund.SettlementCurrency
		execution.SettlementBalanceTransactionID = refund.SettlementBalanceTransactionID
		execution.FXGainLossMinor = refund.FXGainLossMinor
		execution.FXGainLossCurrency = refund.FXGainLossCurrency
		execution.GatewayResponseJSON = string(payload)
		execution.ErrorMessage = ""
		execution.CompletedAt = &now
		if err := repos.RefundExecution.Update(execution); err != nil {
			return err
		}
		if err := releaseRefundPendingHoldIfClear(repos, refund.OrderID); err != nil {
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
			orderRecord.OrderNumber,
		); err != nil {
			return err
		}
		if err := completeLinkedAfterSalesCaseInTx(repos, refund, execution.RequestedByID); err != nil {
			return err
		}

		completedAmountMinor, err := repos.Payment.SumRefundTotalAmountMinorByTransactionID(transaction.ID, "completed")
		if err != nil {
			return err
		}
		transactionMoney, err := transaction.AmountMoney()
		if err != nil {
			return err
		}
		completedMoney, err := domainmoney.New(completedAmountMinor, transaction.Currency)
		if err != nil {
			return err
		}
		transactionFullyRefunded, compareErr := refundAmountAtLeastInMinorUnits(completedMoney, transactionMoney)
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
			orderRefundedAmountMinor, err := repos.Payment.SumRefundTotalAmountMinorByOrderID(orderRecord.ID, "completed")
			if err != nil {
				return err
			}
			orderRefundedMoney, err := domainmoney.New(orderRefundedAmountMinor, orderRecord.Currency)
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
				if refundCanRestockPhysicalItems(orderRecord) {
					if err := repos.Order.UpdateStatus(orderRecord.ID, orderRecord.Status, "refunded"); err != nil {
						return err
					}
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
		execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refundID)
		if err != nil {
			return err
		}
		// The provider may complete the refund asynchronously even though the
		// synchronous request returned an error. A verified webhook can therefore
		// win the race and commit the refund before this failure path acquires the
		// lock. Never overwrite that authoritative completed state with failed.
		if refund.Status == "completed" || execution.Status == paymentdomain.PaymentRefundExecutionStatusSucceeded {
			failedExecution = execution
			return nil
		}
		if err := releaseRefundLoyaltyReservationInTx(repos, orderRecord, refund); err != nil {
			return err
		}
		execution.Status = paymentdomain.PaymentRefundExecutionStatusFailed
		execution.ErrorMessage = strings.TrimSpace(message)
		execution.CompletedAt = &now
		if err := repos.RefundExecution.Update(execution); err != nil {
			return err
		}
		// The refund intent is the source of truth for the amount reserved by
		// this attempt. Mark it failed together with the execution so a failed
		// gateway request no longer counts as a pending refund when calculating
		// the transaction's remaining refundable amount.
		refund.Status = "failed"
		if err := repos.Payment.UpdateRefund(refund); err != nil {
			return err
		}
		if err := releaseRefundPendingHoldIfClear(repos, refund.OrderID); err != nil {
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
	transition, err := repos.AfterSalesCase.UpdateStatusIfCurrentInTxWithEvent(
		caseRecord.ID,
		aftersalesdomain.StatusResolving,
		aftersalesdomain.StatusCompleted,
		"退款执行完成",
		updatedBy,
	)
	if err != nil || transition == nil {
		return err
	}
	audience := outbox.NotificationAudienceSnapshot{}
	if repos.Order != nil {
		orderRecord, orderErr := repos.Order.FindByIDBasic(caseRecord.OrderID)
		if orderErr != nil {
			return orderErr
		}
		if caseRecord.OrderNumber == "" {
			caseRecord.OrderNumber = orderRecord.OrderNumber
		}
		audience = orderNotificationAudience(orderRecord)
	}
	return enqueueAfterSalesStatusChangedDomainEvent(repos.Outbox, caseRecord, transition, nil, audience)
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
		strings.TrimSpace(execution.ErrorMessage) == "" && execution.CompletedAt != nil &&
		execution.SettlementAmountMinor == refund.SettlementAmountMinor &&
		strings.EqualFold(execution.SettlementCurrency, refund.SettlementCurrency) &&
		execution.SettlementBalanceTransactionID == refund.SettlementBalanceTransactionID &&
		execution.FXGainLossMinor == refund.FXGainLossMinor &&
		strings.EqualFold(execution.FXGainLossCurrency, refund.FXGainLossCurrency) {
		return execution, nil
	}

	execution.Status = paymentdomain.PaymentRefundExecutionStatusSucceeded
	execution.ProviderRefundID = providerRefundID
	execution.ProviderStatus = providerStatus
	execution.GatewayResponseJSON = gatewayResponse
	execution.ErrorMessage = ""
	execution.SettlementAmountMinor = refund.SettlementAmountMinor
	execution.SettlementCurrency = refund.SettlementCurrency
	execution.SettlementBalanceTransactionID = refund.SettlementBalanceTransactionID
	execution.FXGainLossMinor = refund.FXGainLossMinor
	execution.FXGainLossCurrency = refund.FXGainLossCurrency
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

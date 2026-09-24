package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	aftersalesdomain "commerce-platform/internal/domain/aftersales"
	currencydomain "commerce-platform/internal/domain/currency"
	loyaltydomain "commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPaymentServiceRefundExecutionOutboxCompletesLocalRefund(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 12000)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 80)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:                             "re_gateway_1",
			PaymentID:                      transaction.TransactionID,
			Amount:                         "80.00",
			Status:                         "succeeded",
			SettlementAmountMinor:          8000,
			SettlementCurrency:             "USD",
			SettlementBalanceTransactionID: "txn_refund_1",
			CreatedAt:                      time.Now().UTC(),
		},
	}

	completedRefund, execution, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundCompleted, 1, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(8000), payload.AmountMinor)
		require.Equal(t, int64(8000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
	})

	require.Equal(t, "completed", completedRefund.Status)
	require.NotNil(t, completedRefund.RefundID)
	require.Equal(t, "re_gateway_1", *completedRefund.RefundID)
	require.NotNil(t, completedRefund.CompletedAt)
	require.Contains(t, completedRefund.GatewayResponse, "re_gateway_1")
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, "re_gateway_1", execution.ProviderRefundID)
	require.Equal(t, "succeeded", execution.ProviderStatus)
	require.Equal(t, int64(8000), execution.SettlementAmountMinor)
	require.Equal(t, "USD", execution.SettlementCurrency)
	require.Equal(t, "txn_refund_1", execution.SettlementBalanceTransactionID)
	require.Zero(t, execution.FXGainLossMinor)
	require.Equal(t, refundExecutionIdempotencyKey(refund.ID), gateway.options.IdempotencyKey)
	require.Equal(t, transaction.Currency, gateway.options.Currency)
	require.Equal(t, int64(12000), gateway.options.OriginalAmountMinor)
	require.Equal(t, orderRecord.OrderNumber, gateway.options.MerchantOrderNumber)
	require.Equal(t, transaction.TransactionID, gateway.options.ProviderTransactionID)
	require.Equal(t, transaction.TransactionID, gateway.paymentID)
	require.Equal(t, int64(8000), gateway.amountMinor)
	require.Equal(t, orderRecord.OrderNumber, execution.MerchantOrderNumber)
	require.Equal(t, transaction.TransactionID, execution.ProviderTransactionID)

	updatedTransaction, err := repository.NewPaymentRepository(db).FindTransactionByID(transaction.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", updatedTransaction.Status)
}

func TestPaymentServiceRefundExecutionOutboxCompletesZeroNetRefundLocally(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 10000)
	refund := paymentdomain.Refund{
		OrderID:                     orderRecord.ID,
		TransactionID:               transaction.ID,
		AmountMinor:                 0,
		RequestedAmountMinor:        majorTestMinor(t, 150, "USD"),
		DiscountClawbackAmountMinor: majorTestMinor(t, 150, "USD"),
		Currency:                    "USD",
		CalculationSnapshot:         `{"net_refund_amount_minor":0,"discount_clawback_amount_minor":15000}`,
		Status:                      "pending",
		Reason:                      "zero net promotional refund",
		RefundedBy:                  7,
	}
	require.NoError(t, db.Create(&refund).Error)
	gateway := &recordingRefundGateway{}

	completedRefund, execution, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  nil,
	})
	require.NoError(t, err)

	require.Equal(t, "completed", completedRefund.Status)
	require.NotNil(t, completedRefund.RefundID)
	require.Equal(t, localZeroRefundID(refund.ID), *completedRefund.RefundID)
	require.Zero(t, completedRefund.AmountMinor)
	require.Contains(t, completedRefund.GatewayResponse, localZeroRefundID(refund.ID))
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, localZeroRefundID(refund.ID), execution.ProviderRefundID)
	require.Equal(t, localZeroRefundStatus, execution.ProviderStatus)
	require.Equal(t, 0, gateway.refundCallCount)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.First(&savedTransaction, transaction.ID).Error)
	require.Equal(t, "completed", savedTransaction.Status)
}

func TestPaymentServiceExecuteZeroNetRefundCompletesLinkedAfterSalesCase(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&aftersalesdomain.AfterSalesCase{},
		&aftersalesdomain.AfterSalesCaseItem{},
		&aftersalesdomain.AfterSalesCaseEvent{},
		&aftersalesdomain.AfterSalesCaseAttachment{},
		&aftersalesdomain.AfterSalesRefundReview{},
	))
	service := newPaymentServiceWithRefundExecution(db)
	service.txManager.ConfigureAfterSalesCaseRepository(repository.NewAfterSalesCaseRepository(db))
	service.txManager.ConfigureAfterSalesRefundReviewRepository(repository.NewAfterSalesRefundReviewRepository(db))
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 10000)
	refund := paymentdomain.Refund{
		OrderID:                     orderRecord.ID,
		TransactionID:               transaction.ID,
		AmountMinor:                 0,
		RequestedAmountMinor:        majorTestMinor(t, 150, "USD"),
		DiscountClawbackAmountMinor: majorTestMinor(t, 150, "USD"),
		Currency:                    "USD",
		Status:                      "pending",
		Reason:                      "zero net after-sales refund",
		RefundedBy:                  7,
	}
	require.NoError(t, db.Create(&refund).Error)
	caseRecord := aftersalesdomain.AfterSalesCase{
		OrderID:   orderRecord.ID,
		Type:      aftersalesdomain.TypeRefundOnly,
		Status:    aftersalesdomain.StatusResolving,
		Reason:    "Returned accessory",
		CreatedBy: 7,
		UpdatedBy: 7,
	}
	require.NoError(t, db.Create(&caseRecord).Error)
	review := aftersalesdomain.AfterSalesRefundReview{
		CaseID:              caseRecord.ID,
		Status:              aftersalesdomain.RefundReviewStatusApproved,
		ProposedAmountMinor: majorTestMinor(t, 150, "USD"),
		Currency:            "USD",
		CreatedBy:           7,
		UpdatedBy:           7,
		LinkedRefundID:      &refund.ID,
	}
	require.NoError(t, db.Create(&review).Error)

	_, _, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
	})
	require.NoError(t, err)

	var savedCase aftersalesdomain.AfterSalesCase
	require.NoError(t, db.First(&savedCase, caseRecord.ID).Error)
	require.Equal(t, aftersalesdomain.StatusCompleted, savedCase.Status)
	require.NotNil(t, savedCase.ClosedAt)

	var event aftersalesdomain.AfterSalesCaseEvent
	require.NoError(t, db.Where("case_id = ?", caseRecord.ID).Order("id DESC").First(&event).Error)
	require.Equal(t, aftersalesdomain.StatusResolving, event.FromStatus)
	require.Equal(t, aftersalesdomain.StatusCompleted, event.ToStatus)
	require.Equal(t, uint(12), event.UpdatedBy)
}

func TestRecordVerifiedGatewayRefundReconcilesTimedOutExecutionAndCompletesLinkedCase(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&aftersalesdomain.AfterSalesCase{},
		&aftersalesdomain.AfterSalesCaseItem{},
		&aftersalesdomain.AfterSalesCaseEvent{},
		&aftersalesdomain.AfterSalesCaseAttachment{},
		&aftersalesdomain.AfterSalesRefundReview{},
	))
	service := newPaymentServiceWithRefundExecution(db)
	service.txManager.ConfigureAfterSalesCaseRepository(repository.NewAfterSalesCaseRepository(db))
	service.txManager.ConfigureAfterSalesRefundReviewRepository(repository.NewAfterSalesRefundReviewRepository(db))
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 10000)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)

	caseRecord := aftersalesdomain.AfterSalesCase{
		OrderID:   orderRecord.ID,
		Type:      aftersalesdomain.TypeRefundOnly,
		Status:    aftersalesdomain.StatusResolving,
		Reason:    "Refund gateway timed out",
		CreatedBy: 7,
		UpdatedBy: 7,
	}
	require.NoError(t, db.Create(&caseRecord).Error)
	review := aftersalesdomain.AfterSalesRefundReview{
		CaseID:              caseRecord.ID,
		Status:              aftersalesdomain.RefundReviewStatusApproved,
		ProposedAmountMinor: majorTestMinor(t, 100, "USD"),
		Currency:            "USD",
		CreatedBy:           7,
		UpdatedBy:           7,
		LinkedRefundID:      &refund.ID,
	}
	require.NoError(t, db.Create(&review).Error)
	execution := paymentdomain.PaymentRefundExecution{
		RefundID:              refund.ID,
		OrderID:               orderRecord.ID,
		TransactionID:         transaction.ID,
		Provider:              "stripe",
		ProviderPaymentID:     transaction.TransactionID,
		MerchantOrderNumber:   orderRecord.OrderNumber,
		ProviderTransactionID: transaction.TransactionID,
		AmountMinor:           majorTestMinor(t, 100, "USD"),
		Currency:              "USD",
		Status:                paymentdomain.PaymentRefundExecutionStatusFailed,
		IdempotencyKey:        refundExecutionIdempotencyKey(refund.ID),
		AttemptCount:          1,
		RequestedByID:         12,
		RequestedAt:           time.Now().UTC().Add(-time.Minute),
		ErrorMessage:          "gateway request timed out",
	}
	require.NoError(t, db.Create(&execution).Error)

	input := VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "re_timed_out_webhook",
		ProviderStatus:       "succeeded",
		ProviderRefundAmount: mustTestMoney(t, 100, "USD"),
		GatewayResponse:      `{"id":"re_timed_out_webhook","status":"succeeded"}`,
	}
	require.NoError(t, service.RecordVerifiedGatewayRefund(input))

	var savedExecution paymentdomain.PaymentRefundExecution
	require.NoError(t, db.First(&savedExecution, execution.ID).Error)
	assert.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, savedExecution.Status)
	assert.Equal(t, "re_timed_out_webhook", savedExecution.ProviderRefundID)
	assert.Equal(t, "succeeded", savedExecution.ProviderStatus)
	assert.Empty(t, savedExecution.ErrorMessage)
	assert.NotNil(t, savedExecution.CompletedAt)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	assert.Equal(t, "completed", savedRefund.Status)
	assert.Equal(t, "re_timed_out_webhook", *savedRefund.RefundID)

	var savedCase aftersalesdomain.AfterSalesCase
	require.NoError(t, db.First(&savedCase, caseRecord.ID).Error)
	assert.Equal(t, aftersalesdomain.StatusCompleted, savedCase.Status)
	assert.NotNil(t, savedCase.ClosedAt)

	var event aftersalesdomain.AfterSalesCaseEvent
	require.NoError(t, db.Where("case_id = ?", caseRecord.ID).Order("id DESC").First(&event).Error)
	assert.Equal(t, aftersalesdomain.StatusResolving, event.FromStatus)
	assert.Equal(t, aftersalesdomain.StatusCompleted, event.ToStatus)
	assert.Equal(t, uint(12), event.UpdatedBy)
	var domainEvent outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypeAfterSalesStatusChanged,
		fmt.Sprint(caseRecord.ID),
	).Order("id DESC").First(&domainEvent).Error)
	var domainPayload outboxdomain.AfterSalesStatusChangedPayload
	require.NoError(t, json.Unmarshal(domainEvent.Payload, &domainPayload))
	assert.Equal(t, aftersalesdomain.StatusCompleted, domainPayload.NewStatus)
	assert.Equal(t, orderRecord.OrderNumber, domainPayload.OrderNumber)
}

func TestRecordVerifiedGatewayRefundRepairsAlreadyCompletedRefundExecutionOnReplay(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&aftersalesdomain.AfterSalesCase{},
		&aftersalesdomain.AfterSalesCaseItem{},
		&aftersalesdomain.AfterSalesCaseEvent{},
		&aftersalesdomain.AfterSalesCaseAttachment{},
		&aftersalesdomain.AfterSalesRefundReview{},
	))
	service := newPaymentServiceWithRefundExecution(db)
	service.txManager.ConfigureAfterSalesCaseRepository(repository.NewAfterSalesCaseRepository(db))
	service.txManager.ConfigureAfterSalesRefundReviewRepository(repository.NewAfterSalesRefundReviewRepository(db))
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 10000)
	providerRefundID := "re_already_completed_replay"
	refund := paymentdomain.Refund{
		OrderID:              orderRecord.ID,
		TransactionID:        transaction.ID,
		RefundID:             &providerRefundID,
		AmountMinor:          majorTestMinor(t, 100, "USD"),
		RequestedAmountMinor: majorTestMinor(t, 100, "USD"),
		Currency:             "USD",
		Status:               "completed",
		CompletedAt:          func() *time.Time { value := time.Now().UTC().Add(-time.Minute); return &value }(),
	}
	require.NoError(t, db.Create(&refund).Error)

	caseRecord := aftersalesdomain.AfterSalesCase{
		OrderID:   orderRecord.ID,
		Type:      aftersalesdomain.TypeRefundOnly,
		Status:    aftersalesdomain.StatusResolving,
		Reason:    "Refund was completed by an earlier webhook",
		CreatedBy: 7,
		UpdatedBy: 7,
	}
	require.NoError(t, db.Create(&caseRecord).Error)
	review := aftersalesdomain.AfterSalesRefundReview{
		CaseID:              caseRecord.ID,
		Status:              aftersalesdomain.RefundReviewStatusApproved,
		ProposedAmountMinor: majorTestMinor(t, 100, "USD"),
		Currency:            "USD",
		CreatedBy:           7,
		UpdatedBy:           7,
		LinkedRefundID:      &refund.ID,
	}
	require.NoError(t, db.Create(&review).Error)
	execution := paymentdomain.PaymentRefundExecution{
		RefundID:              refund.ID,
		OrderID:               orderRecord.ID,
		TransactionID:         transaction.ID,
		Provider:              "stripe",
		ProviderPaymentID:     transaction.TransactionID,
		MerchantOrderNumber:   orderRecord.OrderNumber,
		ProviderTransactionID: transaction.TransactionID,
		AmountMinor:           majorTestMinor(t, 100, "USD"),
		Currency:              "USD",
		Status:                paymentdomain.PaymentRefundExecutionStatusFailed,
		IdempotencyKey:        refundExecutionIdempotencyKey(refund.ID),
		AttemptCount:          1,
		RequestedByID:         12,
		RequestedAt:           time.Now().UTC().Add(-time.Minute),
		ErrorMessage:          "gateway request timed out",
	}
	require.NoError(t, db.Create(&execution).Error)

	require.NoError(t, service.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		RefundID:             providerRefundID,
		ProviderStatus:       "succeeded",
		ProviderRefundAmount: mustTestMoney(t, 100, "USD"),
		GatewayResponse:      `{"id":"re_already_completed_replay","status":"succeeded"}`,
	}))

	var savedExecution paymentdomain.PaymentRefundExecution
	require.NoError(t, db.First(&savedExecution, execution.ID).Error)
	assert.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, savedExecution.Status)
	assert.Equal(t, providerRefundID, savedExecution.ProviderRefundID)
	assert.Empty(t, savedExecution.ErrorMessage)

	var savedCase aftersalesdomain.AfterSalesCase
	require.NoError(t, db.First(&savedCase, caseRecord.ID).Error)
	assert.Equal(t, aftersalesdomain.StatusCompleted, savedCase.Status)
}

func TestPaymentServiceRefundExecutionOutboxInvalidatesRestockedProductCache(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	cacheInvalidator := &recordingProductCacheInvalidator{}
	service.ConfigureProductCacheInvalidator(cacheInvalidator)
	variant := seedPaymentProductVariant(t, db, 4)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 12000)
	orderItem := seedPaymentOrderItemWithVariant(t, db, orderRecord.ID, variant.ProductID, variant.ID, 1, 120, 120, 0, 0, 120)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 120)
	require.NoError(t, db.Create(&paymentdomain.RefundLineItem{
		RefundID:       refund.ID,
		OrderID:        orderRecord.ID,
		OrderItemID:    orderItem.ID,
		ProductID:      variant.ProductID,
		VariantID:      &variant.ID,
		ProductName:    "Carbon component",
		SKU:            variant.SKU,
		Quantity:       1,
		Currency:       "USD",
		UnitPriceMinor: majorTestMinor(t, 120, "USD"),
		LineTotalMinor: majorTestMinor(t, 120, "USD"),
		Restock:        true,
	}).Error)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_gateway_restock",
			PaymentID: transaction.TransactionID,
			Amount:    "120.00",
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	_, _, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})

	require.NoError(t, err)
	assert.Equal(t, []uint{variant.ProductID}, cacheInvalidator.productIDs)
}

func TestPaymentServiceRefundExecutionOutboxRecordsGatewayFailure(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 9000)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 40)
	gateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}

	_, execution, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "gateway unavailable")
	require.NotNil(t, execution)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, execution.Status)
	require.Contains(t, execution.ErrorMessage, "gateway unavailable")
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundFailed, 1, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(4000), payload.AmountMinor)
		require.Equal(t, int64(4000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
		require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, payload.ExecutionStatus)
		require.Contains(t, payload.ErrorMessage, "gateway unavailable")
	})

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	require.Equal(t, "failed", storedRefund.Status)
	require.Nil(t, storedRefund.RefundID)
	require.Nil(t, storedRefund.CompletedAt)
	reservedAmount, err := repository.NewPaymentRepository(db).SumRefundAmountMinorByTransactionID(transaction.ID, "pending", "completed")
	require.NoError(t, err)
	require.Zero(t, reservedAmount)
}

func TestPaymentServiceGatewayRefundFailureWebhookIsIdempotentAndRetryable(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 9000)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 40)

	// A synchronous attempt can fail before the provider's asynchronous failure
	// notification arrives. Keep that execution row so the webhook can enrich
	// the same audit record.
	firstGateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}
	_, _, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  firstGateway,
	})
	require.Error(t, err)

	failure := VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "re_provider_failed",
		ProviderStatus:       "failed",
		ProviderRefundAmount: mustTestMoney(t, 40, "USD"),
		ErrorMessage:         "provider rejected refund",
		GatewayResponse:      `{"id":"re_provider_failed","status":"failed"}`,
	}
	require.NoError(t, service.RecordGatewayRefundFailure(failure))
	require.NoError(t, service.RecordGatewayRefundFailure(failure))
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundFailed, 2, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(4000), payload.AmountMinor)
		require.Equal(t, int64(4000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
		require.Equal(t, "re_provider_failed", payload.ProviderRefundID)
	})

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	require.Equal(t, "failed", savedRefund.Status)
	require.NotNil(t, savedRefund.RefundID)
	require.Equal(t, "re_provider_failed", *savedRefund.RefundID)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).
		Where("transaction_id = ?", transaction.ID).Count(&refundCount).Error)
	require.Equal(t, int64(1), refundCount)

	retryGateway := &recordingRefundGateway{response: &pgateway.RefundResponse{
		ID:        "re_provider_retry",
		PaymentID: transaction.TransactionID,
		Amount:    "40.00",
		Status:    "succeeded",
	}}
	completed, execution, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  retryGateway,
	})
	require.NoError(t, err)
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundCompleted, 1, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(4000), payload.AmountMinor)
		require.Equal(t, int64(4000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
		require.Equal(t, "re_provider_retry", payload.ProviderRefundID)
	})
	require.Equal(t, "completed", completed.Status)
	require.Equal(t, "re_provider_retry", *completed.RefundID)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, 2, execution.AttemptCount)
	require.Equal(t, "re_provider_retry", execution.ProviderRefundID)
}

func TestPaymentServiceExecuteDuplicatePaidCompensationRefundRetriesAndPreservesOrderPaymentState(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, _ := createRefundRecommendationPaidOrder(t, db, 10000)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 25, 0)

	result, err := service.RecordVerifiedGatewayPaymentResult(VerifiedGatewayPaymentInput{
		Provider:      "paypal",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "CAPTURE-DUPLICATE-1",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(10000, "USD"),
	})
	require.NoError(t, err)
	require.True(t, result.DuplicatePaid)

	gateway := &recordingRefundGateway{
		err:     errors.New("paypal temporarily unavailable"),
		errOnce: true,
		response: &pgateway.RefundResponse{
			ID:        "REFUND-DUPLICATE-1",
			PaymentID: "CAPTURE-DUPLICATE-1",
			Amount:    "100.00",
			Status:    "COMPLETED",
			CreatedAt: time.Now().UTC(),
		},
	}
	input := executePendingRefundTestInput{
		RefundID: result.RefundID,
		AdminID:  12,
		Provider: "paypal",
		Gateway:  gateway,
	}

	_, failedExecution, err := executeRefundViaOutbox(t, service, db, context.Background(), input)
	require.Error(t, err)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, failedExecution.Status)
	require.Equal(t, 1, failedExecution.AttemptCount)
	var failedRefund paymentdomain.Refund
	require.NoError(t, db.First(&failedRefund, result.RefundID).Error)
	require.Equal(t, "failed", failedRefund.Status)

	_, succeededExecution, err := executeRefundViaOutbox(t, service, db, context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, succeededExecution.Status)
	require.Equal(t, 2, succeededExecution.AttemptCount)
	require.Equal(t, "REFUND-DUPLICATE-1", succeededExecution.ProviderRefundID)
	require.Equal(t, refundExecutionIdempotencyKey(result.RefundID), gateway.options.IdempotencyKey)
	require.Equal(t, int64(10000), gateway.amountMinor)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, result.RefundID).Error)
	require.Equal(t, "completed", savedRefund.Status)
	require.Equal(t, duplicatePaidRefundReason, savedRefund.Reason)
	require.Equal(t, majorTestMinor(t, 100, "USD"), savedRefund.AmountMinor)
	require.False(t, savedRefund.LoyaltySettlementPrepared)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "CAPTURE-DUPLICATE-1").First(&savedTransaction).Error)
	require.Equal(t, paymentdomain.TransactionStatusDuplicatePaid, savedTransaction.Status)

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "paid", savedOrder.PaymentStatus)
	require.Equal(t, "paid", savedOrder.Status)

	var loyaltyTransactionCount int64
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source LIKE ?", "refund_loyalty_%").
		Count(&loyaltyTransactionCount).Error)
	require.Zero(t, loyaltyTransactionCount)
}

func TestPaymentServiceRefundExecutionOutboxPersistsHistoricalFXSnapshot(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	fetchedAt := time.Date(2026, time.August, 6, 9, 0, 0, 0, time.UTC)
	capturedAt := time.Date(2026, time.August, 6, 10, 0, 0, 0, time.UTC)
	fxSnapshot := currencydomain.OrderFXSnapshot{
		Version:       currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "EUR",
		RateDecimal:   "0.9",
		Source:        "historical_order_quote",
		CapturedAt:    capturedAt,
		RateFetchedAt: &fetchedAt,
	}

	orderRecord := orderdomain.Order{
		OrderNumber:        "ORD-RFX-1",
		UserID:             11,
		Status:             "paid",
		PaymentStatus:      "paid",
		TotalAmountMinor:   majorTestMinor(t, 90, "EUR"),
		Currency:           "EUR",
		PaymentAmountMinor: majorTestMinor(t, 90, "EUR"),
		PaymentCurrency:    "EUR",
		FXSnapshotData:     currencydomain.OrderFXSnapshotJSON(fxSnapshot),
		ShippingAddress:    orderdomain.Address{Country: "DE"},
		BillingAddress:     orderdomain.Address{Country: "DE"},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_1",
		PaymentMethod: "stripe",
		AmountMinor:   majorTestMinor(t, 90, "EUR"),
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 30)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_rfx_1",
			PaymentID: transaction.TransactionID,
			Amount:    "30.00",
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, execution, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, "completed", completedRefund.Status)

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	persistedSnapshot, err := currencydomain.ParseOrderFXSnapshot(storedRefund.FXSnapshotData)
	require.NoError(t, err)
	require.Equal(t, fxSnapshot.BaseCurrency, persistedSnapshot.BaseCurrency)
	require.Equal(t, fxSnapshot.OrderCurrency, persistedSnapshot.OrderCurrency)
	require.Equal(t, fxSnapshot.RateDecimal, persistedSnapshot.RateDecimal)
	require.Equal(t, fxSnapshot.Source, persistedSnapshot.Source)
	require.NotNil(t, persistedSnapshot.RateFetchedAt)
	require.True(t, fetchedAt.Equal(*persistedSnapshot.RateFetchedAt))
}

func TestPaymentServiceRefundExecutionOutboxCompletesNonUSDFXSettlementAboveHistoricalValue(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	fxSnapshot := currencydomain.OrderFXSnapshot{
		Version:       currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "EUR",
		RateDecimal:   "0.9",
		Source:        "historical_order_quote",
		CapturedAt:    time.Date(2026, time.August, 6, 10, 0, 0, 0, time.UTC),
	}
	orderRecord := orderdomain.Order{
		OrderNumber: "ORD-RFX-SETTLEMENT", UserID: 11, Status: "paid", PaymentStatus: "paid",
		TotalAmountMinor: majorTestMinor(t, 90, "EUR"), Currency: "EUR",
		PaymentAmountMinor: majorTestMinor(t, 90, "EUR"), PaymentCurrency: "EUR",
		FXSnapshotData: currencydomain.OrderFXSnapshotJSON(fxSnapshot),
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID: orderRecord.ID, TransactionID: "pi_rfx_settlement", PaymentMethod: "stripe",
		AmountMinor: majorTestMinor(t, 90, "EUR"), Currency: "EUR", Status: "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 90)
	gateway := &recordingRefundGateway{response: &pgateway.RefundResponse{
		ID: "re_rfx_settlement", PaymentID: transaction.TransactionID, AmountMinor: 11000,
		Status: "succeeded", SettlementAmountMinor: 11000, SettlementCurrency: "USD",
		SettlementBalanceTransactionID: "txn_rfx_settlement", CreatedAt: time.Now().UTC(),
	}}

	completedRefund, execution, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID, AdminID: 12, Provider: "stripe", Gateway: gateway,
	})
	require.NoError(t, err)
	require.Equal(t, "completed", completedRefund.Status)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, int64(11000), completedRefund.SettlementAmountMinor)
	require.Equal(t, "USD", completedRefund.SettlementCurrency)
	require.Equal(t, int64(1000), completedRefund.FXGainLossMinor)
	require.Equal(t, "USD", completedRefund.FXGainLossCurrency)
}

func TestPaymentServiceRefundExecutionOutboxRejectsMissingHistoricalFXSnapshot(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	orderRecord := orderdomain.Order{
		OrderNumber:        "ORD-RFX-MISSING",
		UserID:             11,
		Status:             "paid",
		PaymentStatus:      "paid",
		TotalAmountMinor:   majorTestMinor(t, 90, "EUR"),
		Currency:           "EUR",
		PaymentAmountMinor: majorTestMinor(t, 90, "EUR"),
		PaymentCurrency:    "EUR",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_missing",
		PaymentMethod: "stripe",
		AmountMinor:   majorTestMinor(t, 90, "EUR"),
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 30)
	gateway := &recordingRefundGateway{}

	_, _, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.ErrorIs(t, err, ErrHistoricalRefundFXSnapshotMissing)
	require.Empty(t, gateway.paymentID)

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", storedRefund.Status)
	require.Equal(t, "{}", string(storedRefund.FXSnapshotData))
}

func TestPaymentServiceRefundExecutionOutboxRejectsNonUSDRefundAboveHistoricalFXCap(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	fxSnapshot := currencydomain.OrderFXSnapshot{
		Version:       currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "EUR",
		RateDecimal:   "0.9",
		Source:        "historical_order_quote",
		CapturedAt:    time.Date(2026, time.August, 6, 10, 0, 0, 0, time.UTC),
	}
	orderRecord := orderdomain.Order{
		OrderNumber:        "ORD-RFX-CAP",
		UserID:             11,
		Status:             "paid",
		PaymentStatus:      "paid",
		TotalAmountMinor:   majorTestMinor(t, 90, "EUR"),
		Currency:           "EUR",
		PaymentAmountMinor: majorTestMinor(t, 90, "EUR"),
		PaymentCurrency:    "EUR",
		FXSnapshotData:     currencydomain.OrderFXSnapshotJSON(fxSnapshot),
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_cap",
		PaymentMethod: "stripe",
		AmountMinor:   majorTestMinor(t, 90, "EUR"),
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{}

	_, _, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "historical FX cap")
	require.Empty(t, gateway.paymentID)

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", storedRefund.Status)
}

func TestPaymentServiceRefundExecutionOutboxReleasesLoyaltyReservationWhenGatewayFails(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 10000)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 25, 0)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}

	_, _, err := executeRefundViaOutbox(t, service, db, context.Background(), executePendingRefundTestInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.Error(t, err)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	require.Equal(t, majorTestMinor(t, 100, "USD"), savedRefund.AmountMinor)
	require.False(t, savedRefund.LoyaltySettlementPrepared)
	require.Zero(t, savedRefund.LoyaltyPointsClawback)

	var userLoyalty loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&userLoyalty).Error)
	require.Equal(t, 25, userLoyalty.AvailablePoints)

	var netReservedPoints int
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source IN ?", []string{refundLoyaltyClawbackSource, refundLoyaltyClawbackReversalSource}).
		Select("COALESCE(SUM(points), 0)").
		Scan(&netReservedPoints).Error)
	require.Zero(t, netReservedPoints)
}

func seedRefundLoyaltyProgramConfig(t *testing.T, db *gorm.DB, exchangeRatePoints int) loyaltydomain.ProgramConfig {
	t.Helper()

	config := loyaltydomain.ProgramConfig{
		Version:                   1,
		Status:                    "active",
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
	}
	require.NoError(t, db.Create(&config).Error)
	return config
}

func seedRefundOrderLoyalty(
	t *testing.T,
	db *gorm.DB,
	orderRecord orderdomain.Order,
	config loyaltydomain.ProgramConfig,
	earnedPoints int,
	availablePoints int,
	usedPoints int,
) {
	t.Helper()

	require.NoError(t, db.Create(&loyaltydomain.UserLoyalty{
		UserID:          orderRecord.UserID,
		TotalPoints:     earnedPoints,
		AvailablePoints: availablePoints,
		UsedPoints:      usedPoints,
	}).Error)
	require.NoError(t, db.Create(&loyaltydomain.LoyaltyTransaction{
		UserID:          orderRecord.UserID,
		Type:            "earn",
		Points:          earnedPoints,
		Balance:         availablePoints,
		Source:          "order",
		SourceID:        orderRecord.ID,
		ProgramConfigID: &config.ID,
		Description:     "Order completion reward",
	}).Error)
}

func newPaymentServiceWithRefundExecution(db *gorm.DB) *PaymentService {
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(
		db,
		repository.NewOrderRepository(db),
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		paymentRepo,
	)
	txManager.ConfigureLoyaltyProgramRepository(repository.NewLoyaltyProgramRepository(db))
	txManager.ConfigurePaymentRefundExecutionRepository(repository.NewPaymentRefundExecutionRepository(db))
	txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	return NewPaymentService(txManager, paymentRepo)
}

func createPendingRefundRecord(t *testing.T, db *gorm.DB, orderID uint, transactionID uint, amount float64) paymentdomain.Refund {
	t.Helper()
	var orderRecord orderdomain.Order
	require.NoError(t, db.First(&orderRecord, orderID).Error)
	currency := orderRecord.Currency
	if currency == "" {
		currency = "USD"
	}

	refund := paymentdomain.Refund{
		OrderID:              orderID,
		TransactionID:        transactionID,
		AmountMinor:          majorTestMinor(t, amount, currency),
		RequestedAmountMinor: majorTestMinor(t, amount, currency),
		Currency:             currency,
		Status:               "pending",
		Reason:               "manual pending refund",
		RefundedBy:           7,
	}
	require.NoError(t, db.Create(&refund).Error)
	return refund
}

type executePendingRefundTestInput struct {
	RefundID uint
	AdminID  uint
	Provider string
	Gateway  pgateway.PaymentGateway
}

// executeRefundViaOutbox models the production boundary: the request records
// the local intent and command, then the worker consumes that command and
// performs the provider call. Tests inject a recording gateway at the worker
// boundary so no network credentials or calls are required.
func executeRefundViaOutbox(
	t *testing.T,
	service *PaymentService,
	db *gorm.DB,
	ctx context.Context,
	input executePendingRefundTestInput,
) (*paymentdomain.Refund, *paymentdomain.PaymentRefundExecution, error) {
	t.Helper()
	_, _, requestErr := service.RequestPendingRefundExecution(ctx, RequestPendingRefundExecutionInput{
		RefundID: input.RefundID,
		AdminID:  input.AdminID,
		Provider: input.Provider,
	})
	if requestErr != nil {
		return nil, nil, requestErr
	}

	var event outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypePaymentRefundExecutionRequested,
		fmt.Sprint(input.RefundID),
	).Order("id DESC").First(&event).Error)
	var payload outboxdomain.PaymentRefundExecutionRequestedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, input.RefundID, payload.RefundID)
	require.Equal(t, input.AdminID, payload.AdminID)
	require.Equal(t, input.Provider, payload.Provider)
	require.NotEmpty(t, payload.IdempotencyKey)
	worker := NewPaymentRefundExecutionOutboxHandler(
		service,
		nil,
		func(_ *pgateway.Config) (pgateway.PaymentGateway, error) {
			return input.Gateway, nil
		},
	)
	workerErr := worker.Handle(ctx, event)

	refund, refundErr := repository.NewPaymentRepository(db).FindRefundByID(input.RefundID)
	if refundErr != nil {
		return nil, nil, refundErr
	}
	execution, executionErr := repository.NewPaymentRefundExecutionRepository(db).FindByRefundIDForUpdate(input.RefundID)
	if executionErr != nil {
		return refund, nil, executionErr
	}
	return refund, execution, workerErr
}

func assertPaymentRefundOutboxEvent(
	t *testing.T,
	db *gorm.DB,
	eventType string,
	expectedCount int64,
	assertPayload func(outboxdomain.PaymentRefundPayload),
) {
	t.Helper()
	var events []outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", eventType).Order("id ASC").Find(&events).Error)
	require.Len(t, events, int(expectedCount))
	if len(events) == 0 || assertPayload == nil {
		return
	}
	var payload outboxdomain.PaymentRefundPayload
	require.NoError(t, json.Unmarshal(events[len(events)-1].Payload, &payload))
	assertPayload(payload)
}

type recordingRefundGateway struct {
	paymentID       string
	amountMinor     int64
	options         pgateway.RefundOptions
	response        *pgateway.RefundResponse
	err             error
	errOnce         bool
	refundCallCount int
}

func (g *recordingRefundGateway) CreatePayment(ctx context.Context, req *pgateway.PaymentRequest) (*pgateway.PaymentResponse, error) {
	return nil, errors.New("not implemented")
}

func (g *recordingRefundGateway) CapturePayment(ctx context.Context, paymentID string) (*pgateway.PaymentResponse, error) {
	return nil, errors.New("not implemented")
}

func (g *recordingRefundGateway) RefundPayment(ctx context.Context, paymentID string, amountMinor int64) (*pgateway.RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amountMinor, pgateway.RefundOptions{})
}

func (g *recordingRefundGateway) RefundPaymentWithOptions(ctx context.Context, paymentID string, amountMinor int64, options pgateway.RefundOptions) (*pgateway.RefundResponse, error) {
	g.refundCallCount++
	g.paymentID = paymentID
	g.amountMinor = amountMinor
	g.options = options
	if g.err != nil {
		err := g.err
		if g.errOnce {
			g.err = nil
		}
		return nil, err
	}
	return g.response, nil
}

func (g *recordingRefundGateway) GetPayment(ctx context.Context, paymentID string) (*pgateway.PaymentResponse, error) {
	return nil, errors.New("not implemented")
}

func (g *recordingRefundGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	return false, errors.New("not implemented")
}

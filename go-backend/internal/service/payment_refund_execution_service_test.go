package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	aftersalesdomain "commerce-platform/internal/domain/aftersales"
	coupondomain "commerce-platform/internal/domain/coupon"
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

func seedRefundGiftCardPayment(
	t *testing.T,
	db *gorm.DB,
	orderID uint,
	code string,
	usedCents int64,
) coupondomain.GiftCard {
	t.Helper()

	card := coupondomain.GiftCard{
		Code:         code,
		InitialCents: usedCents,
		BalanceCents: 0,
		Currency:     "USD",
		Status:       "used",
	}
	require.NoError(t, db.Create(&card).Error)
	require.NoError(t, db.Create(&coupondomain.GiftCardTransaction{
		GiftCardID:   card.ID,
		OrderID:      orderID,
		Currency:     "USD",
		Type:         "use",
		AmountCents:  -usedCents,
		BalanceCents: 0,
		Note:         "Gift card payment",
	}).Error)
	return card
}

func TestPaymentServiceAdminRefundSplitsMixedGiftCardAndGatewayPayment(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 1000)
	transaction.Amount = 700
	require.NoError(t, db.Save(&transaction).Error)
	card := seedRefundGiftCardPayment(t, db, orderRecord.ID, "GC-PAY-03-MIXED", 30000)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        1000,
		Reason:        "full mixed-payment refund",
	}
	require.NoError(t, service.CreateAdminRefund(&refund, 12))
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundPending, 1, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(70000), payload.AmountMinor)
		require.Equal(t, int64(30000), payload.GiftCardAmountMinor)
		require.Equal(t, int64(100000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
	})
	require.InDelta(t, 700, refund.Amount, 0.001)
	require.InDelta(t, 300, refund.GiftCardRefundAmount, 0.001)
	require.InDelta(t, 1000, refund.RequestedAmount, 0.001)

	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_pay_03_mixed",
			PaymentID: transaction.TransactionID,
			Amount:    700,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}
	completedRefund, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundCompleted, 1, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(70000), payload.AmountMinor)
		require.Equal(t, int64(30000), payload.GiftCardAmountMinor)
		require.Equal(t, int64(100000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
	})
	require.InDelta(t, 700, gateway.amount, 0.001)
	require.InDelta(t, 300, completedRefund.GiftCardRefundAmount, 0.001)

	var savedCard coupondomain.GiftCard
	require.NoError(t, db.First(&savedCard, card.ID).Error)
	require.Equal(t, int64(30000), savedCard.BalanceCents)
	require.Equal(t, "active", savedCard.Status)

	var restorationCount int64
	require.NoError(t, db.Model(&coupondomain.GiftCardTransaction{}).
		Where("refund_id = ? AND type = ?", refund.ID, "refund").
		Count(&restorationCount).Error)
	require.Equal(t, int64(1), restorationCount)

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "refunded", savedOrder.Status)
	require.Equal(t, "refunded", savedOrder.PaymentStatus)
}

func TestPaymentServiceExecutePendingRefundRestoresGiftCardWithoutGateway(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 300)
	transaction.Amount = 0
	require.NoError(t, db.Save(&transaction).Error)
	card := seedRefundGiftCardPayment(t, db, orderRecord.ID, "GC-PAY-03-FULL", 30000)
	refund := paymentdomain.Refund{
		OrderID:              orderRecord.ID,
		TransactionID:        transaction.ID,
		Amount:               0,
		GiftCardRefundAmount: 300,
		RequestedAmount:      300,
		Status:               "pending",
		Reason:               "full gift-card refund",
		RefundedBy:           12,
	}
	require.NoError(t, db.Create(&refund).Error)

	gateway := &recordingRefundGateway{}
	completedRefund, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  nil,
	})
	require.NoError(t, err)
	require.Equal(t, 0, gateway.refundCallCount)
	require.InDelta(t, 0, completedRefund.Amount, 0.001)

	var savedCard coupondomain.GiftCard
	require.NoError(t, db.First(&savedCard, card.ID).Error)
	require.Equal(t, int64(30000), savedCard.BalanceCents)

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "refunded", savedOrder.Status)
	require.Equal(t, "refunded", savedOrder.PaymentStatus)
}

func TestPaymentServicePartialGiftCardRefundDoesNotMarkTransactionFullyRefunded(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 300)
	transaction.Amount = 0
	require.NoError(t, db.Save(&transaction).Error)
	card := seedRefundGiftCardPayment(t, db, orderRecord.ID, "GC-PAY-03-PARTIAL", 30000)
	refund := paymentdomain.Refund{
		OrderID:              orderRecord.ID,
		TransactionID:        transaction.ID,
		Amount:               0,
		GiftCardRefundAmount: 100,
		RequestedAmount:      100,
		Status:               "pending",
		Reason:               "partial gift-card refund",
		RefundedBy:           12,
	}
	require.NoError(t, db.Create(&refund).Error)

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
	})
	require.NoError(t, err)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.First(&savedTransaction, transaction.ID).Error)
	require.Equal(t, "completed", savedTransaction.Status)

	var savedCard coupondomain.GiftCard
	require.NoError(t, db.First(&savedCard, card.ID).Error)
	require.Equal(t, int64(10000), savedCard.BalanceCents)
}

func TestPaymentServiceExecutePendingRefundDoesNotRestoreGiftCardWhenGatewayFails(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 1000)
	transaction.Amount = 700
	require.NoError(t, db.Save(&transaction).Error)
	card := seedRefundGiftCardPayment(t, db, orderRecord.ID, "GC-PAY-03-FAIL", 30000)
	refund := paymentdomain.Refund{
		OrderID:              orderRecord.ID,
		TransactionID:        transaction.ID,
		Amount:               700,
		GiftCardRefundAmount: 300,
		RequestedAmount:      1000,
		Status:               "pending",
		Reason:               "failed mixed-payment refund",
		RefundedBy:           12,
	}
	require.NoError(t, db.Create(&refund).Error)

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  &recordingRefundGateway{err: errors.New("gateway unavailable")},
	})
	require.Error(t, err)

	var savedCard coupondomain.GiftCard
	require.NoError(t, db.First(&savedCard, card.ID).Error)
	require.Equal(t, int64(0), savedCard.BalanceCents)

	var restorationCount int64
	require.NoError(t, db.Model(&coupondomain.GiftCardTransaction{}).
		Where("refund_id = ? AND type = ?", refund.ID, "refund").
		Count(&restorationCount).Error)
	require.Zero(t, restorationCount)
}

func TestRecordVerifiedGatewayRefundRestoresMixedGiftCardPaymentIdempotently(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 1000)
	transaction.Amount = 700
	require.NoError(t, db.Save(&transaction).Error)
	card := seedRefundGiftCardPayment(t, db, orderRecord.ID, "GC-PAY-03-WEBHOOK", 30000)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 700)
	refund.GiftCardRefundAmount = 300
	refund.RequestedAmount = 1000
	require.NoError(t, db.Save(&refund).Error)

	input := VerifiedGatewayRefundInput{
		Provider:              "stripe",
		OrderNumber:           orderRecord.OrderNumber,
		TransactionID:         transaction.TransactionID,
		RefundID:              "re_pay_03_webhook",
		ProviderRefundAmount:  mustTestMoney(t, 700, "USD"),
		RequestedRefundAmount: mustTestMoney(t, 1000, "USD"),
	}
	require.NoError(t, service.RecordVerifiedGatewayRefund(input))
	require.NoError(t, service.RecordVerifiedGatewayRefund(input))

	var savedCard coupondomain.GiftCard
	require.NoError(t, db.First(&savedCard, card.ID).Error)
	require.Equal(t, int64(30000), savedCard.BalanceCents)

	var restorationCount int64
	require.NoError(t, db.Model(&coupondomain.GiftCardTransaction{}).
		Where("refund_id = ? AND type = ?", refund.ID, "refund").
		Count(&restorationCount).Error)
	require.Equal(t, int64(1), restorationCount)

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "refunded", savedOrder.Status)
	require.Equal(t, "refunded", savedOrder.PaymentStatus)
}

func TestPaymentServiceExecutePendingRefundCompletesLocalRefund(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 120)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 80)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_gateway_1",
			PaymentID: transaction.TransactionID,
			Amount:    80,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	assertPaymentRefundOutboxEvent(t, db, outboxdomain.EventTypePaymentRefundCompleted, 1, func(payload outboxdomain.PaymentRefundPayload) {
		require.Equal(t, int64(8000), payload.AmountMinor)
		require.Equal(t, int64(0), payload.GiftCardAmountMinor)
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
	require.Equal(t, refundExecutionIdempotencyKey(refund.ID), gateway.options.IdempotencyKey)
	require.Equal(t, transaction.Currency, gateway.options.Currency)
	require.Equal(t, transaction.Amount, gateway.options.OriginalAmount)
	require.Equal(t, orderRecord.OrderNumber, gateway.options.MerchantOrderNumber)
	require.Equal(t, transaction.TransactionID, gateway.options.ProviderTransactionID)
	require.Equal(t, transaction.TransactionID, gateway.paymentID)
	require.Equal(t, 80.0, gateway.amount)
	require.Equal(t, orderRecord.OrderNumber, execution.MerchantOrderNumber)
	require.Equal(t, transaction.TransactionID, execution.ProviderTransactionID)

	updatedTransaction, err := repository.NewPaymentRepository(db).FindTransactionByID(transaction.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", updatedTransaction.Status)
}

func TestPaymentServiceExecutePendingRefundCompletesZeroNetRefundLocally(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	refund := paymentdomain.Refund{
		OrderID:                orderRecord.ID,
		TransactionID:          transaction.ID,
		Amount:                 0,
		RequestedAmount:        150,
		DiscountClawbackAmount: 150,
		CalculationSnapshot:    `{"net_refund_amount":0,"discount_clawback_amount":150}`,
		Status:                 "pending",
		Reason:                 "zero net promotional refund",
		RefundedBy:             7,
	}
	require.NoError(t, db.Create(&refund).Error)
	gateway := &recordingRefundGateway{}

	completedRefund, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  nil,
	})
	require.NoError(t, err)

	require.Equal(t, "completed", completedRefund.Status)
	require.NotNil(t, completedRefund.RefundID)
	require.Equal(t, localZeroRefundID(refund.ID), *completedRefund.RefundID)
	require.InDelta(t, 0, completedRefund.Amount, 0.001)
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
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	refund := paymentdomain.Refund{
		OrderID:                orderRecord.ID,
		TransactionID:          transaction.ID,
		Amount:                 0,
		RequestedAmount:        150,
		DiscountClawbackAmount: 150,
		Status:                 "pending",
		Reason:                 "zero net after-sales refund",
		RefundedBy:             7,
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
		CaseID:         caseRecord.ID,
		Status:         aftersalesdomain.RefundReviewStatusApproved,
		ProposedAmount: 150,
		Currency:       "USD",
		CreatedBy:      7,
		UpdatedBy:      7,
		LinkedRefundID: &refund.ID,
	}
	require.NoError(t, db.Create(&review).Error)

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
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
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
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
		CaseID:         caseRecord.ID,
		Status:         aftersalesdomain.RefundReviewStatusApproved,
		ProposedAmount: 100,
		Currency:       "USD",
		CreatedBy:      7,
		UpdatedBy:      7,
		LinkedRefundID: &refund.ID,
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
		Amount:                100,
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
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	providerRefundID := "re_already_completed_replay"
	refund := paymentdomain.Refund{
		OrderID:         orderRecord.ID,
		TransactionID:   transaction.ID,
		RefundID:        &providerRefundID,
		Amount:          100,
		RequestedAmount: 100,
		Status:          "completed",
		CompletedAt:     func() *time.Time { value := time.Now().UTC().Add(-time.Minute); return &value }(),
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
		CaseID:         caseRecord.ID,
		Status:         aftersalesdomain.RefundReviewStatusApproved,
		ProposedAmount: 100,
		Currency:       "USD",
		CreatedBy:      7,
		UpdatedBy:      7,
		LinkedRefundID: &refund.ID,
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
		Amount:                100,
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

func TestPaymentServiceExecutePendingRefundInvalidatesRestockedProductCache(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	cacheInvalidator := &recordingProductCacheInvalidator{}
	service.ConfigureProductCacheInvalidator(cacheInvalidator)
	variant := seedPaymentProductVariant(t, db, 4)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 120)
	orderItem := seedPaymentOrderItemWithVariant(t, db, orderRecord.ID, variant.ProductID, variant.ID, 1, 120, 120, 0, 0, 120)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 120)
	require.NoError(t, db.Create(&paymentdomain.RefundLineItem{
		RefundID:        refund.ID,
		OrderID:         orderRecord.ID,
		OrderItemID:     orderItem.ID,
		ProductID:       variant.ProductID,
		VariantID:       &variant.ID,
		ProductName:     "Carbon component",
		SKU:             variant.SKU,
		Quantity:        1,
		UnitPrice:       120,
		LineTotalAmount: 120,
		Restock:         true,
	}).Error)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_gateway_restock",
			PaymentID: transaction.TransactionID,
			Amount:    120,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})

	require.NoError(t, err)
	assert.Equal(t, []uint{variant.ProductID}, cacheInvalidator.productIDs)
}

func TestPaymentServiceExecutePendingRefundRecordsGatewayFailure(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 90)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 40)
	gateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}

	_, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
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
		require.Equal(t, int64(0), payload.GiftCardAmountMinor)
		require.Equal(t, int64(4000), payload.RequestedAmountMinor)
		require.Equal(t, "USD", payload.Currency)
		require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, payload.ExecutionStatus)
		require.Contains(t, payload.ErrorMessage, "gateway unavailable")
	})

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", storedRefund.Status)
	require.Nil(t, storedRefund.RefundID)
	require.Nil(t, storedRefund.CompletedAt)
}

func TestPaymentServiceGatewayRefundFailureWebhookIsIdempotentAndRetryable(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 90)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 40)

	// A synchronous attempt can fail before the provider's asynchronous failure
	// notification arrives. Keep that execution row so the webhook can enrich
	// the same audit record.
	firstGateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}
	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
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
		Amount:    40,
		Status:    "succeeded",
	}}
	completed, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
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
	orderRecord, _ := createRefundRecommendationPaidOrder(t, db, 100)
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
			Amount:    100,
			Status:    "COMPLETED",
			CreatedAt: time.Now().UTC(),
		},
	}
	input := ExecutePendingRefundInput{
		RefundID: result.RefundID,
		AdminID:  12,
		Provider: "paypal",
		Gateway:  gateway,
	}

	_, failedExecution, err := service.ExecutePendingRefund(context.Background(), input)
	require.Error(t, err)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, failedExecution.Status)
	require.Equal(t, 1, failedExecution.AttemptCount)

	_, succeededExecution, err := service.ExecutePendingRefund(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, succeededExecution.Status)
	require.Equal(t, 2, succeededExecution.AttemptCount)
	require.Equal(t, "REFUND-DUPLICATE-1", succeededExecution.ProviderRefundID)
	require.Equal(t, refundExecutionIdempotencyKey(result.RefundID), gateway.options.IdempotencyKey)
	require.Equal(t, 100.0, gateway.amount)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, result.RefundID).Error)
	require.Equal(t, "completed", savedRefund.Status)
	require.Equal(t, duplicatePaidRefundReason, savedRefund.Reason)
	require.InDelta(t, 100, savedRefund.Amount, 0.001)
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

func TestPaymentServiceExecutePendingRefundKeepsLoyaltyReservationForMismatchedProviderAmount(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 25, 0)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_amount_mismatch",
			PaymentID: transaction.TransactionID,
			Amount:    80,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "manual reconciliation required")

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	require.Equal(t, "pending", savedRefund.Status)
	require.True(t, savedRefund.LoyaltySettlementPrepared)
	require.Equal(t, 25, savedRefund.LoyaltyPointsClawback)
	require.InDelta(t, 92.5, savedRefund.Amount, 0.001)
	require.InDelta(t, 7.5, savedRefund.LoyaltyCashDeductionAmount, 0.001)

	var savedExecution paymentdomain.PaymentRefundExecution
	require.NoError(t, db.Where("refund_id = ?", refund.ID).First(&savedExecution).Error)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, savedExecution.Status)
	require.Equal(t, "re_amount_mismatch", savedExecution.ProviderRefundID)
	require.Contains(t, savedExecution.ErrorMessage, "manual reconciliation required")
	require.NotNil(t, savedExecution.CompletedAt)

	var userLoyalty loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&userLoyalty).Error)
	require.Equal(t, 0, userLoyalty.AvailablePoints)
}

func TestPaymentServiceExecutePendingRefundPersistsHistoricalFXSnapshot(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	fetchedAt := time.Date(2026, time.August, 6, 9, 0, 0, 0, time.UTC)
	capturedAt := time.Date(2026, time.August, 6, 10, 0, 0, 0, time.UTC)
	fxSnapshot := currencydomain.OrderFXSnapshot{
		Version:         currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:    "USD",
		OrderCurrency:   "EUR",
		BaseToOrderRate: 0.9,
		Source:          "historical_order_quote",
		CapturedAt:      capturedAt,
		RateFetchedAt:   &fetchedAt,
	}

	orderRecord := orderdomain.Order{
		OrderNumber:     "ORD-RFX-1",
		UserID:          11,
		Status:          "paid",
		PaymentStatus:   "paid",
		TotalAmount:     90,
		Currency:        "EUR",
		PaymentAmount:   90,
		PaymentCurrency: "EUR",
		FXSnapshotData:  currencydomain.OrderFXSnapshotJSON(fxSnapshot),
		ShippingAddress: orderdomain.Address{Country: "DE"},
		BillingAddress:  orderdomain.Address{Country: "DE"},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_1",
		PaymentMethod: "stripe",
		Amount:        90,
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 30)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_rfx_1",
			PaymentID: transaction.TransactionID,
			Amount:    30,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
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
	require.InDelta(t, fxSnapshot.BaseToOrderRate, persistedSnapshot.BaseToOrderRate, 0.0001)
	require.Equal(t, fxSnapshot.Source, persistedSnapshot.Source)
	require.NotNil(t, persistedSnapshot.RateFetchedAt)
	require.True(t, fetchedAt.Equal(*persistedSnapshot.RateFetchedAt))
}

func TestPaymentServiceExecutePendingRefundRejectsMissingHistoricalFXSnapshot(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	orderRecord := orderdomain.Order{
		OrderNumber:     "ORD-RFX-MISSING",
		UserID:          11,
		Status:          "paid",
		PaymentStatus:   "paid",
		TotalAmount:     90,
		Currency:        "EUR",
		PaymentAmount:   90,
		PaymentCurrency: "EUR",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_missing",
		PaymentMethod: "stripe",
		Amount:        90,
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 30)
	gateway := &recordingRefundGateway{}

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
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

func TestPaymentServiceExecutePendingRefundSettlesFullLoyaltyEffectsIdempotently(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 100)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	orderRecord.PointsUsed = 20
	require.NoError(t, db.Save(&orderRecord).Error)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 100, 20)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_loyalty_full",
			PaymentID: transaction.TransactionID,
			Amount:    100,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	require.Equal(t, 100, completedRefund.LoyaltyPointsClawback)
	require.Equal(t, 20, completedRefund.LoyaltyPointsReturned)
	require.Zero(t, completedRefund.LoyaltyCashDeductionAmount)

	var userLoyalty loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&userLoyalty).Error)
	require.Equal(t, 20, userLoyalty.AvailablePoints)
	require.Equal(t, 100, userLoyalty.UsedPoints)
	require.Equal(t, 100, userLoyalty.TotalPoints)

	var clawbackCount int64
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ? AND source_id = ?", refundLoyaltyClawbackSource, refund.ID).
		Count(&clawbackCount).Error)
	require.Equal(t, int64(1), clawbackCount)

	var returnedCount int64
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ? AND source_id = ?", refundLoyaltyPointsReturnSource, refund.ID).
		Count(&returnedCount).Error)
	require.Equal(t, int64(1), returnedCount)

	require.NoError(t, service.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "re_loyalty_full",
		ProviderRefundAmount: mustTestMoney(t, 100, "USD"),
	}))

	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ? AND source_id = ?", refundLoyaltyClawbackSource, refund.ID).
		Count(&clawbackCount).Error)
	require.Equal(t, int64(1), clawbackCount)
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ? AND source_id = ?", refundLoyaltyPointsReturnSource, refund.ID).
		Count(&returnedCount).Error)
	require.Equal(t, int64(1), returnedCount)
}

func TestPaymentServiceRefundKeepsReferralPointsSeparateFromCashAndReturnsOrderPoints(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 150)
	orderRecord.PointsUsed = 50
	orderRecord.PaymentAmount = 100
	orderRecord.PaymentAmountMinor = 10000
	require.NoError(t, db.Save(&orderRecord).Error)

	// The referee received 75 referral points, then spent 50 of them on this
	// order and paid the remaining 100 USD through the gateway. The referral
	// credit is deliberately not an order-earned reward.
	require.NoError(t, db.Create(&loyaltydomain.UserLoyalty{
		UserID:          orderRecord.UserID,
		TotalPoints:     75,
		AvailablePoints: 25,
		UsedPoints:      50,
	}).Error)
	require.NoError(t, db.Create(&loyaltydomain.LoyaltyTransaction{
		UserID:          orderRecord.UserID,
		Type:            "earn",
		Points:          75,
		Balance:         75,
		Source:          "referral_referee",
		SourceID:        9001,
		ProgramConfigID: &config.ID,
		Description:     "Referral welcome benefit",
	}).Error)
	require.NoError(t, db.Create(&loyaltydomain.LoyaltyTransaction{
		UserID:      orderRecord.UserID,
		Type:        "spend",
		Points:      -50,
		Balance:     25,
		Source:      "order",
		SourceID:    orderRecord.ID,
		Description: "Spent referral points on order",
	}).Error)

	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_referral_points_cash_separate",
			PaymentID: transaction.TransactionID,
			Amount:    100,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	require.InDelta(t, 100, gateway.amount, 0.001)
	require.InDelta(t, 100, completedRefund.Amount, 0.001)
	require.Equal(t, 0, completedRefund.LoyaltyPointsClawback)
	require.Equal(t, 50, completedRefund.LoyaltyPointsReturned)
	require.Zero(t, completedRefund.LoyaltyCashDeductionAmount)

	var balance loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&balance).Error)
	require.Equal(t, 75, balance.AvailablePoints)
	require.Equal(t, 0, balance.UsedPoints)

	var returnedPoints int
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ? AND source_id = ?", refundLoyaltyPointsReturnSource, refund.ID).
		Select("COALESCE(SUM(points), 0)").
		Scan(&returnedPoints).Error)
	require.Equal(t, 50, returnedPoints)

	var referralReversalPoints int
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ?", referralRefereeReversalSource).
		Select("COALESCE(SUM(points), 0)").
		Scan(&referralReversalPoints).Error)
	// The referral reward is reversed by the referral invalidation outbox
	// handler, independently of this payment refund transaction.
	require.Zero(t, referralReversalPoints)
}

func TestPaymentServiceExecutePendingRefundAllocatesPartialLoyaltyEffectsAcrossRefunds(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 100)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	orderRecord.PointsUsed = 20
	require.NoError(t, db.Save(&orderRecord).Error)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 100, 20)

	firstRefund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 40)
	firstGateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_loyalty_partial_1",
			PaymentID: transaction.TransactionID,
			Amount:    40,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}
	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: firstRefund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  firstGateway,
	})
	require.NoError(t, err)

	secondRefund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 60)
	secondGateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_loyalty_partial_2",
			PaymentID: transaction.TransactionID,
			Amount:    60,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}
	_, _, err = service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: secondRefund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  secondGateway,
	})
	require.NoError(t, err)

	var savedFirst, savedSecond paymentdomain.Refund
	require.NoError(t, db.First(&savedFirst, firstRefund.ID).Error)
	require.NoError(t, db.First(&savedSecond, secondRefund.ID).Error)
	require.Equal(t, 40, savedFirst.LoyaltyPointsClawback)
	require.Equal(t, 8, savedFirst.LoyaltyPointsReturned)
	require.Equal(t, 60, savedSecond.LoyaltyPointsClawback)
	require.Equal(t, 12, savedSecond.LoyaltyPointsReturned)

	var userLoyalty loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&userLoyalty).Error)
	require.Equal(t, 20, userLoyalty.AvailablePoints)
	require.Equal(t, 100, userLoyalty.UsedPoints)

	var clawbackPoints, returnedPoints int
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ?", refundLoyaltyClawbackSource).
		Select("COALESCE(SUM(points), 0)").
		Scan(&clawbackPoints).Error)
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source = ?", refundLoyaltyPointsReturnSource).
		Select("COALESCE(SUM(points), 0)").
		Scan(&returnedPoints).Error)
	require.Equal(t, -100, clawbackPoints)
	require.Equal(t, 20, returnedPoints)
}

func TestPaymentServiceExecutePendingRefundDeductsCashWhenEarnedPointsWereSpent(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 25, 0)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_loyalty_cash_recovery",
			PaymentID: transaction.TransactionID,
			Amount:    92.5,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	require.InDelta(t, 92.5, gateway.amount, 0.001)
	require.InDelta(t, 92.5, completedRefund.Amount, 0.001)
	require.Equal(t, 25, completedRefund.LoyaltyPointsClawback)
	require.Zero(t, completedRefund.LoyaltyPointsReturned)
	require.InDelta(t, 7.5, completedRefund.LoyaltyCashDeductionAmount, 0.001)

	var userLoyalty loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&userLoyalty).Error)
	require.Equal(t, 0, userLoyalty.AvailablePoints)
	require.Equal(t, 25, userLoyalty.UsedPoints)
}

func TestPaymentServiceExecutePendingRefundReleasesLoyaltyReservationWhenGatewayFails(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 25, 0)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)
	gateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.Error(t, err)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	require.InDelta(t, 100, savedRefund.Amount, 0.001)
	require.False(t, savedRefund.LoyaltySettlementPrepared)
	require.Zero(t, savedRefund.LoyaltyPointsClawback)
	require.Zero(t, savedRefund.LoyaltyCashDeductionAmount)

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

func TestRecordVerifiedGatewayRefundSettlesLoyaltyEffectsOnlyOnce(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	orderRecord.PointsUsed = 10
	require.NoError(t, db.Save(&orderRecord).Error)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 0, 0)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 100)

	input := VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "re_webhook_loyalty",
		ProviderRefundAmount: mustTestMoney(t, 90, "USD"),
	}
	require.NoError(t, service.RecordVerifiedGatewayRefund(input))
	require.NoError(t, service.RecordVerifiedGatewayRefund(input))

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	require.InDelta(t, 90, savedRefund.Amount, 0.001)
	require.Equal(t, 0, savedRefund.LoyaltyPointsClawback)
	require.Equal(t, 10, savedRefund.LoyaltyPointsReturned)
	require.InDelta(t, 10, savedRefund.LoyaltyCashDeductionAmount, 0.001)

	var userLoyalty loyaltydomain.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", orderRecord.UserID).First(&userLoyalty).Error)
	require.Equal(t, 10, userLoyalty.AvailablePoints)

	var refundTransactionCount int64
	require.NoError(t, db.Model(&loyaltydomain.LoyaltyTransaction{}).
		Where("source IN ?", []string{refundLoyaltyClawbackSource, refundLoyaltyPointsReturnSource}).
		Count(&refundTransactionCount).Error)
	require.Equal(t, int64(1), refundTransactionCount)
}

func TestRecordVerifiedGatewayRefundWithoutPendingRefundUsesExplicitRequestedAmount(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	config := seedRefundLoyaltyProgramConfig(t, db, 10)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 100)
	seedRefundOrderLoyalty(t, db, orderRecord, config, 100, 25, 0)

	err := service.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:              "stripe",
		OrderNumber:           orderRecord.OrderNumber,
		TransactionID:         transaction.TransactionID,
		RefundID:              "re_external_refund",
		ProviderRefundAmount:  mustTestMoney(t, 92.5, "USD"),
		RequestedRefundAmount: mustTestMoney(t, 100, "USD"),
	})
	require.NoError(t, err)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.Where("refund_id = ?", "re_external_refund").First(&savedRefund).Error)
	require.Equal(t, "completed", savedRefund.Status)
	require.InDelta(t, 100, savedRefund.RequestedAmount, 0.001)
	require.InDelta(t, 92.5, savedRefund.Amount, 0.001)
	require.InDelta(t, 7.5, savedRefund.LoyaltyCashDeductionAmount, 0.001)
	require.True(t, savedRefund.LoyaltySettlementPrepared)
}

func seedRefundLoyaltyProgramConfig(t *testing.T, db *gorm.DB, exchangeRatePoints int) loyaltydomain.ProgramConfig {
	t.Helper()

	config := loyaltydomain.ProgramConfig{
		Version:                   1,
		Status:                    "active",
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ExchangeRatePoints:        exchangeRatePoints,
		MinRedeemPoints:           1,
		MaxValuePerDayCents:       50000,
		CardExpiryDays:            365,
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

	refund := paymentdomain.Refund{
		OrderID:         orderID,
		TransactionID:   transactionID,
		Amount:          amount,
		RequestedAmount: amount,
		Status:          "pending",
		Reason:          "manual pending refund",
		RefundedBy:      7,
	}
	require.NoError(t, db.Create(&refund).Error)
	return refund
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
	amount          float64
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

func (g *recordingRefundGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*pgateway.RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amount, pgateway.RefundOptions{})
}

func (g *recordingRefundGateway) RefundPaymentWithOptions(ctx context.Context, paymentID string, amount float64, options pgateway.RefundOptions) (*pgateway.RefundResponse, error) {
	g.refundCallCount++
	g.paymentID = paymentID
	g.amount = amount
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

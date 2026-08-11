package service

import (
	"testing"
	"time"

	coupondomain "commerce-platform/internal/domain/coupon"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaymentRefundRecommendationServiceEnqueuesIdempotently(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	repo := repository.NewPaymentRefundRecommendationRepository(db)
	service := NewPaymentRefundRecommendationService(repo)
	now := time.Date(2026, time.August, 2, 8, 0, 0, 0, time.UTC)

	transaction := paymentdomain.Transaction{
		OrderID:       42,
		TransactionID: "pi_123",
		PaymentMethod: "stripe",
		Amount:        120,
		Currency:      "USD",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	riskEvent := paymentdomain.PaymentRiskEvent{
		Provider:          "stripe",
		Kind:              paymentdomain.PaymentRiskEventEarlyFraudWarning,
		ExternalReference: "efw_123",
		WebhookEventID:    "evt_123",
		ProviderPaymentID: "pi_123",
		Amount:            120,
		Currency:          "USD",
		OccurredAt:        now,
		MetadataJSON:      "{}",
	}
	require.NoError(t, db.Create(&riskEvent).Error)

	input := PaymentRiskEventInput{
		Provider:          "stripe",
		Kind:              paymentdomain.PaymentRiskEventEarlyFraudWarning,
		ExternalReference: "efw_123",
		WebhookEventID:    "evt_123",
		ProviderPaymentID: "pi_123",
		Amount:            120,
		Currency:          "usd",
		OccurredAt:        now,
		Metadata:          map[string]string{"fraud_type": "card"},
	}
	first, err := service.EnqueueFromRiskEvent(input)
	require.NoError(t, err)
	second, err := service.EnqueueFromRiskEvent(input)
	require.NoError(t, err)

	require.Equal(t, first.ID, second.ID)
	require.Equal(t, paymentdomain.PaymentRefundRecommendationStatusPending, first.Status)
	require.Equal(t, paymentdomain.PaymentRefundRecommendationActionReviewRefundBeforeDispute, first.RecommendedAction)
	require.Equal(t, paymentdomain.PaymentRefundRecommendationPriorityHigh, first.Priority)
	require.Equal(t, "USD", first.Currency)
	require.NotNil(t, first.OrderID)
	require.Equal(t, uint(42), *first.OrderID)
	require.NotNil(t, first.TransactionID)
	require.Equal(t, transaction.ID, *first.TransactionID)
	require.NotNil(t, first.RiskEventID)
	require.Equal(t, riskEvent.ID, *first.RiskEventID)
	require.NotNil(t, first.ReviewBy)
	require.Equal(t, now.Add(24*time.Hour), *first.ReviewBy)

	var count int64
	require.NoError(t, db.Model(&paymentdomain.PaymentRefundRecommendation{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestPaymentRefundRecommendationServiceCancelsNoLongerActionableDispute(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := NewPaymentRefundRecommendationService(repository.NewPaymentRefundRecommendationRepository(db))

	input := PaymentRiskEventInput{
		Provider:          "paypal",
		Kind:              paymentdomain.PaymentRiskEventDispute,
		ExternalReference: "PP-D-1",
		WebhookEventID:    "WH-1",
		ProviderPaymentID: "PAY-1",
		Amount:            50,
		Currency:          "USD",
		OccurredAt:        time.Date(2026, time.August, 2, 9, 0, 0, 0, time.UTC),
		Metadata:          map[string]string{"status": "WAITING_FOR_SELLER_RESPONSE", "reason": "unauthorized"},
	}
	record, err := service.EnqueueFromRiskEvent(input)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundRecommendationStatusPending, record.Status)

	input.Metadata["status"] = "closed"
	record, err = service.EnqueueFromRiskEvent(input)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundRecommendationStatusCancelled, record.Status)
	require.Equal(t, "Provider risk event is no longer actionable.", record.DecisionNotes)
}

func TestPaymentRefundRecommendationDecisionDoesNotCreateRefund(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := NewPaymentRefundRecommendationService(repository.NewPaymentRefundRecommendationRepository(db))
	record, err := service.EnqueueFromRiskEvent(PaymentRiskEventInput{
		Provider:          "stripe",
		Kind:              paymentdomain.PaymentRiskEventEarlyFraudWarning,
		ExternalReference: "efw_decision",
		WebhookEventID:    "evt_decision",
		ProviderPaymentID: "pi_decision",
		Amount:            88,
		Currency:          "USD",
		OccurredAt:        time.Now().UTC(),
		Metadata:          map[string]string{"fraud_type": "card"},
	})
	require.NoError(t, err)

	record, err = service.UpdateRecommendationDecision(record.ID, paymentdomain.PaymentRefundRecommendationStatusAccepted, "Manual refund will be reviewed from the order page.", 7)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundRecommendationStatusAccepted, record.Status)
	require.NotNil(t, record.ReviewedByID)
	require.Equal(t, uint(7), *record.ReviewedByID)
	require.NotNil(t, record.ReviewedAt)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Count(&refundCount).Error)
	require.Equal(t, int64(0), refundCount)
}

func TestPaymentRefundRecommendationCreatesPendingRefundDraft(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentRefundRecommendationWorkflowService(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 120)
	recommendation := createRefundRecommendationRecord(t, db, orderRecord.ID, transaction.ID, 100)

	updated, refund, err := service.CreatePendingRefundFromRecommendation(CreatePendingRefundFromRecommendationInput{
		RecommendationID: recommendation.ID,
		Amount:           80,
		Reason:           "manual risk review refund",
		DecisionNotes:    "Customer contacted before dispute escalation.",
		AdminID:          7,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.NotNil(t, refund)

	require.Equal(t, paymentdomain.PaymentRefundRecommendationStatusAccepted, updated.Status)
	require.NotNil(t, updated.LinkedRefundID)
	require.Equal(t, refund.ID, *updated.LinkedRefundID)
	require.NotNil(t, updated.RefundCreatedByID)
	require.Equal(t, uint(7), *updated.RefundCreatedByID)
	require.NotNil(t, updated.RefundCreatedAt)
	require.Contains(t, updated.DecisionNotes, "Created local pending refund")

	require.Equal(t, orderRecord.ID, refund.OrderID)
	require.Equal(t, transaction.ID, refund.TransactionID)
	require.Equal(t, "pending", refund.Status)
	require.Equal(t, 80.0, refund.Amount)
	require.Equal(t, 80.0, refund.RequestedAmount)
	require.Equal(t, "manual risk review refund", refund.Reason)
	require.Nil(t, refund.RefundID)
	require.Empty(t, refund.GatewayResponse)
	require.Nil(t, refund.CompletedAt)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Count(&refundCount).Error)
	require.Equal(t, int64(1), refundCount)
}

func TestPaymentRefundRecommendationPendingRefundDraftIsIdempotent(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentRefundRecommendationWorkflowService(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 75)
	recommendation := createRefundRecommendationRecord(t, db, orderRecord.ID, transaction.ID, 75)

	_, firstRefund, err := service.CreatePendingRefundFromRecommendation(CreatePendingRefundFromRecommendationInput{
		RecommendationID: recommendation.ID,
		AdminID:          9,
	})
	require.NoError(t, err)
	updated, secondRefund, err := service.CreatePendingRefundFromRecommendation(CreatePendingRefundFromRecommendationInput{
		RecommendationID: recommendation.ID,
		AdminID:          9,
	})
	require.NoError(t, err)

	require.Equal(t, firstRefund.ID, secondRefund.ID)
	require.NotNil(t, updated.LinkedRefundID)
	require.Equal(t, firstRefund.ID, *updated.LinkedRefundID)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Count(&refundCount).Error)
	require.Equal(t, int64(1), refundCount)
}

func newPaymentRefundRecommendationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&coupondomain.Coupon{},
		&coupondomain.CouponUsage{},
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
		&paymentdomain.RefundLineItem{},
		&paymentdomain.PaymentRefundExecution{},
		&paymentdomain.PaymentRiskEvent{},
		&paymentdomain.PaymentRefundRecommendation{},
	))
	return db
}

func newPaymentRefundRecommendationWorkflowService(db *gorm.DB) *PaymentRefundRecommendationService {
	refundRecommendationRepo := repository.NewPaymentRefundRecommendationRepository(db)
	txManager := repository.NewTxManager(
		db,
		repository.NewOrderRepository(db),
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		repository.NewPaymentRepository(db),
	)
	txManager.ConfigurePaymentRefundRecommendationRepository(refundRecommendationRepo)
	txManager.ConfigurePaymentRefundExecutionRepository(repository.NewPaymentRefundExecutionRepository(db))
	return NewPaymentRefundRecommendationService(refundRecommendationRepo, txManager)
}

func createRefundRecommendationPaidOrder(t *testing.T, db *gorm.DB, amount float64) (orderdomain.Order, paymentdomain.Transaction) {
	t.Helper()

	orderRecord := orderdomain.Order{
		OrderNumber:    "ORD-RISK-1",
		UserID:         11,
		Status:         "paid",
		PaymentStatus:  "paid",
		SubtotalAmount: amount,
		TotalAmount:    amount,
		Currency:       "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_refund_draft",
		PaymentMethod: "stripe",
		Amount:        amount,
		Currency:      "USD",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	return orderRecord, transaction
}

func createRefundRecommendationRecord(t *testing.T, db *gorm.DB, orderID uint, transactionID uint, amount float64) paymentdomain.PaymentRefundRecommendation {
	t.Helper()

	recommendation := paymentdomain.PaymentRefundRecommendation{
		Provider:           "stripe",
		SourceKind:         paymentdomain.PaymentRiskEventEarlyFraudWarning,
		ExternalReference:  "efw_refund_draft",
		WebhookEventID:     "evt_refund_draft",
		OrderID:            &orderID,
		TransactionID:      &transactionID,
		ProviderPaymentID:  "pi_refund_draft",
		RecommendedAction:  paymentdomain.PaymentRefundRecommendationActionReviewRefundBeforeDispute,
		RecommendedAmount:  amount,
		Currency:           "USD",
		Priority:           paymentdomain.PaymentRefundRecommendationPriorityHigh,
		Status:             paymentdomain.PaymentRefundRecommendationStatusPending,
		Reason:             "Early fraud warning received; review a local refund draft.",
		SourceMetadataJSON: "{}",
	}
	require.NoError(t, db.Create(&recommendation).Error)
	return recommendation
}

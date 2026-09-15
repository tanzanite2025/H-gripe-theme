package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	attributiondomain "commerce-platform/internal/domain/attribution"
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	productdomain "commerce-platform/internal/domain/product"
	shippingdomain "commerce-platform/internal/domain/shipping"
	ticketdomain "commerce-platform/internal/domain/ticket"
	userdomain "commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/invoice"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	paypalapi "github.com/plutov/paypal/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func mustTestMoney(t *testing.T, amount float64, code string) domainmoney.Money {
	t.Helper()
	money, err := domainmoney.FromMajorFloat(amount, code)
	require.NoError(t, err)
	return money
}

func TestRecordVerifiedGatewayPaymentCreatesLedgerAndMarksOrderPaid(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-1", 84, "pending", "unpaid")
	orderRecord.ShippingAddress = order.Address{
		FirstName: "Ada",
		LastName:  "Rider",
		Email:     "ada.rider@example.test",
	}
	require.NoError(t, db.Save(&orderRecord).Error)

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_123",
		Amount:        domainmoney.MustNew(8400, "USD"),
	})

	require.NoError(t, err)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "txn_123").First(&savedTransaction).Error)
	assert.Equal(t, orderRecord.ID, savedTransaction.OrderID)
	assert.Equal(t, "stripe", savedTransaction.PaymentMethod)
	assert.Equal(t, "completed", savedTransaction.Status)
	assert.NotNil(t, savedTransaction.CompletedAt)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)
	assert.Equal(t, "processing", savedOrder.Status)
	assert.NotNil(t, savedOrder.PaidAt)

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_123",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}))

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("transaction_id = ?", "txn_123").Count(&transactionCount).Error)
	assert.Equal(t, int64(1), transactionCount)

	var event outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", fmt.Sprintf("%s:%d:%s", outboxdomain.EventTypeOrderPaid, orderRecord.ID, "txn_123")).First(&event).Error)
	assert.Equal(t, outboxdomain.EventTypeOrderPaid, event.EventType)
	assert.Equal(t, outboxdomain.AggregateTypeOrder, event.AggregateType)
	assert.Equal(t, "pending", event.Status)

	var payload outboxdomain.OrderPaidPayload
	require.NoError(t, json.Unmarshal([]byte(event.Payload), &payload))
	assert.Equal(t, orderRecord.ID, payload.OrderID)
	assert.Equal(t, orderRecord.OrderNumber, payload.OrderNumber)
	assert.Equal(t, "txn_123", payload.PaymentTransactionID)
	assert.InDelta(t, 84, payload.Amount, 0.001)

	var referralEvent outboxdomain.Event
	require.NoError(t, db.Where(
		"event_key = ?",
		fmt.Sprintf("%s:%d:%s", outboxdomain.EventTypeReferralOrderPaid, orderRecord.ID, "txn_123"),
	).First(&referralEvent).Error)
	assert.Equal(t, outboxdomain.EventTypeReferralOrderPaid, referralEvent.EventType)
	var referralPayload outboxdomain.ReferralOrderPaidPayload
	require.NoError(t, json.Unmarshal([]byte(referralEvent.Payload), &referralPayload))
	assert.Equal(t, orderRecord.ID, referralPayload.OrderID)
	assert.Equal(t, int64(8400), referralPayload.AmountMinor)
	assert.Equal(t, "USD", referralPayload.Currency)

	var confirmation outboxdomain.Event
	require.NoError(t, db.Where(
		"event_key = ?",
		fmt.Sprintf("%s:%d:%s", outboxdomain.EventTypeOrderConfirmationEmail, orderRecord.ID, "txn_123"),
	).First(&confirmation).Error)
	assert.Equal(t, outboxdomain.EventTypeOrderConfirmationEmail, confirmation.EventType)
	assert.Equal(t, outboxdomain.AggregateTypeOrder, confirmation.AggregateType)
	assert.Equal(t, outboxdomain.EventStatusPending, confirmation.Status)

	var confirmationPayload outboxdomain.OrderConfirmationEmailPayload
	require.NoError(t, json.Unmarshal([]byte(confirmation.Payload), &confirmationPayload))
	assert.Equal(t, "ada.rider@example.test", confirmationPayload.RecipientEmail)
	assert.Equal(t, "Ada Rider", confirmationPayload.CustomerName)
	assert.Equal(t, orderRecord.ID, confirmationPayload.OrderID)
	assert.Equal(t, orderRecord.OrderNumber, confirmationPayload.OrderNumber)
	assert.InDelta(t, 84, confirmationPayload.Amount, 0.001)
	assert.Equal(t, "USD", confirmationPayload.Currency)
	assert.False(t, confirmationPayload.PaidAt.IsZero())

	var outboxCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_key = ?", event.EventKey).Count(&outboxCount).Error)
	assert.Equal(t, int64(1), outboxCount)
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_key = ?", confirmation.EventKey).Count(&outboxCount).Error)
	assert.Equal(t, int64(1), outboxCount)

	var conversion outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", fmt.Sprintf("%s:%d:%s", outboxdomain.EventTypeVerifiedConversion, orderRecord.ID, "txn_123")).First(&conversion).Error)
	assert.Equal(t, outboxdomain.EventTypeVerifiedConversion, conversion.EventType)
}

func TestRecordVerifiedGatewayPaymentRecordsDuplicatePaidTransactionAndCreatesFullRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-ALREADY-PAID", 84, "processing", "paid")
	seedCompletedTransaction(t, db, orderRecord.ID, "txn_original_paid", 84, "USD")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "paypal",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_provider_later_event",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(150000, "USD"),
	})

	require.NoError(t, err)

	var duplicateTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "txn_provider_later_event").First(&duplicateTransaction).Error)
	assert.Equal(t, paymentdomain.TransactionStatusDuplicatePaid, duplicateTransaction.Status)
	assert.InDelta(t, 1500, duplicateTransaction.Amount, 0.001)
	assert.Equal(t, "USD", duplicateTransaction.Currency)

	var refunds []paymentdomain.Refund
	require.NoError(t, db.Where("transaction_id = ?", duplicateTransaction.ID).Find(&refunds).Error)
	require.Len(t, refunds, 1)
	assert.Equal(t, "pending", refunds[0].Status)
	assert.InDelta(t, 1500, refunds[0].Amount, 0.001)
	assert.InDelta(t, 1500, refunds[0].RequestedAmount, 0.001)
	assert.Equal(t, uint(0), refunds[0].RefundedBy)
	assert.Contains(t, refunds[0].Reason, duplicatePaidRefundReasonCode)

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "paypal",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_provider_later_event",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(150000, "USD"),
	}))

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).
		Where("transaction_id = ?", "txn_provider_later_event").
		Count(&transactionCount).Error)
	assert.Equal(t, int64(1), transactionCount)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).
		Where("transaction_id = ?", duplicateTransaction.ID).
		Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)

	var outboxCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Count(&outboxCount).Error)
	assert.Equal(t, int64(0), outboxCount)
}

func TestRecordVerifiedGatewayPaymentPersistsLiabilityShifted(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	shifted := true
	notShifted := false

	shiftedOrder := seedPaymentOrder(t, db, "ORD-PAY-LIABILITY-SHIFTED", 84, "pending", "unpaid")
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:         "stripe",
		OrderNumber:      shiftedOrder.OrderNumber,
		TransactionID:    "txn_liability_shifted",
		Amount:           domainmoney.MustNew(8400, "USD"),
		LiabilityShifted: &shifted,
	}))

	var shiftedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "txn_liability_shifted").First(&shiftedTransaction).Error)
	require.NotNil(t, shiftedTransaction.LiabilityShifted)
	assert.True(t, *shiftedTransaction.LiabilityShifted)

	notShiftedOrder := seedPaymentOrder(t, db, "ORD-PAY-LIABILITY-NOT-SHIFTED", 92, "pending", "unpaid")
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:         "stripe",
		OrderNumber:      notShiftedOrder.OrderNumber,
		TransactionID:    "txn_liability_not_shifted",
		Amount:           domainmoney.MustNew(9200, "USD"),
		LiabilityShifted: &notShifted,
	}))

	var notShiftedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "txn_liability_not_shifted").First(&notShiftedTransaction).Error)
	require.NotNil(t, notShiftedTransaction.LiabilityShifted)
	assert.False(t, *notShiftedTransaction.LiabilityShifted)
}

func TestRecordVerifiedGatewayPaymentHoldsHighValueLiabilityNotShiftedOrders(t *testing.T) {
	tests := []struct {
		name             string
		total            float64
		liabilityShifted *bool
		wantStatus       string
		wantHold         bool
		wantReview       bool
	}{
		{
			name:             "below threshold",
			total:            749.99,
			liabilityShifted: boolPtr(false),
			wantStatus:       "processing",
		},
		{
			name:             "at threshold without liability shift",
			total:            750,
			liabilityShifted: boolPtr(false),
			wantStatus:       "needs_review",
			wantHold:         true,
			wantReview:       true,
		},
		{
			name:       "at threshold with unknown liability shift",
			total:      750,
			wantStatus: "needs_review",
			wantHold:   true,
			wantReview: true,
		},
		{
			name:             "above threshold with liability shifted",
			total:            2500,
			liabilityShifted: boolPtr(true),
			wantStatus:       "processing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, paymentService := newTestPaymentService(t)
			orderRecord := seedPaymentOrder(t, db, "ORD-PAY-HIGH-VALUE-"+strings.ReplaceAll(test.name, " ", "-"), test.total, "pending", "unpaid")

			require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
				Provider:         "stripe",
				OrderNumber:      orderRecord.OrderNumber,
				TransactionID:    "pi-" + strings.ReplaceAll(test.name, " ", "-"),
				Amount:           mustTestMoney(t, test.total, "USD"),
				LiabilityShifted: test.liabilityShifted,
			}))

			var savedOrder order.Order
			require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
			assert.Equal(t, "paid", savedOrder.PaymentStatus)
			assert.Equal(t, test.wantStatus, savedOrder.Status)
			assert.Equal(t, test.wantHold, savedOrder.FulfillmentHold)

			var transaction paymentdomain.Transaction
			require.NoError(t, db.Where("order_id = ?", orderRecord.ID).First(&transaction).Error)
			assert.Equal(t, test.liabilityShifted, transaction.LiabilityShifted)

			var reviews []paymentdomain.PaymentReview
			require.NoError(t, db.Where("order_id = ? AND reason = ?", orderRecord.ID, highValueLiabilityReviewReason).Find(&reviews).Error)
			if !test.wantReview {
				assert.Empty(t, reviews)
				return
			}
			require.Len(t, reviews, 1)
			assert.Equal(t, "pending", reviews[0].Status)
			assert.Equal(t, "gateway_verification", reviews[0].Source)
			assert.Equal(t, transaction.ID, *reviews[0].TransactionID)
			assert.Equal(t, transaction.TransactionID, reviews[0].PaymentIntentID)
			assert.Contains(t, reviews[0].Notes, "manual identity or wire-transfer verification")
		})
	}
}

func TestStripeAutomaticReviewReleasesHighValueOrderOnlyAfterApprovedLiabilityReview(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-REVIEW-ORDERING", 2500, "needs_review", "paid")
	orderRecord.FulfillmentHold = true
	require.NoError(t, db.Save(&orderRecord).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "pi_review_ordering", 2500, "USD")

	require.NoError(t, db.Create(&paymentdomain.PaymentReview{
		OrderID:         &orderRecord.ID,
		TransactionID:   &transaction.ID,
		PaymentIntentID: transaction.TransactionID,
		Status:          "pending",
		Reason:          highValueLiabilityReviewReason,
		Source:          "gateway_verification",
	}).Error)
	automaticReview, err := paymentService.CreatePaymentReview(CreatePaymentReviewInput{
		PaymentIntentID: transaction.TransactionID,
		StripeReviewID:  "prv_review_ordering",
		Status:          "pending",
		Reason:          stripeRadarReviewReason,
		Source:          "radar",
	})
	require.NoError(t, err)
	require.NotNil(t, automaticReview.OrderID)
	assert.Equal(t, orderRecord.ID, *automaticReview.OrderID)
	require.NotNil(t, automaticReview.TransactionID)
	assert.Equal(t, transaction.ID, *automaticReview.TransactionID)

	require.NoError(t, paymentService.ResolveStripeReview("prv_review_ordering", transaction.TransactionID, "approved"))

	var afterAutomaticClose order.Order
	require.NoError(t, db.First(&afterAutomaticClose, orderRecord.ID).Error)
	assert.Equal(t, "needs_review", afterAutomaticClose.Status)
	assert.True(t, afterAutomaticClose.FulfillmentHold)

	var liabilityReview paymentdomain.PaymentReview
	require.NoError(t, db.Where("order_id = ? AND reason = ?", orderRecord.ID, highValueLiabilityReviewReason).First(&liabilityReview).Error)
	_, err = paymentService.UpdatePaymentReview(liabilityReview.ID, "approved", "identity verified", 7)
	require.NoError(t, err)

	var released order.Order
	require.NoError(t, db.First(&released, orderRecord.ID).Error)
	assert.Equal(t, "processing", released.Status)
	assert.False(t, released.FulfillmentHold)
}

func TestStripeAutomaticReviewClosesBeforeReleasingApprovedHighValueOrder(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-REVIEW-RELEASE", 2500, "needs_review", "paid")
	orderRecord.FulfillmentHold = true
	require.NoError(t, db.Save(&orderRecord).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "pi_review_release", 2500, "USD")

	liabilityReview := paymentdomain.PaymentReview{
		OrderID:         &orderRecord.ID,
		TransactionID:   &transaction.ID,
		PaymentIntentID: transaction.TransactionID,
		Status:          "pending",
		Reason:          highValueLiabilityReviewReason,
		Source:          "gateway_verification",
	}
	require.NoError(t, db.Create(&liabilityReview).Error)
	automaticReview := paymentdomain.PaymentReview{
		OrderID:         &orderRecord.ID,
		TransactionID:   &transaction.ID,
		PaymentIntentID: transaction.TransactionID,
		StripeReviewID:  "prv_review_release",
		Status:          "pending",
		Reason:          stripeRadarReviewReason,
		Source:          "radar",
	}
	require.NoError(t, db.Create(&automaticReview).Error)

	_, err := paymentService.UpdatePaymentReview(liabilityReview.ID, "approved", "identity verified", 7)
	require.NoError(t, err)
	var stillHeld order.Order
	require.NoError(t, db.First(&stillHeld, orderRecord.ID).Error)
	assert.Equal(t, "needs_review", stillHeld.Status)
	assert.True(t, stillHeld.FulfillmentHold)

	require.NoError(t, paymentService.ResolveStripeReview("prv_review_release", transaction.TransactionID, "approved"))
	var released order.Order
	require.NoError(t, db.First(&released, orderRecord.ID).Error)
	assert.Equal(t, "processing", released.Status)
	assert.False(t, released.FulfillmentHold)
}

func TestStripeAutomaticReviewDoesNotReleaseHighValueOrderWithManualReview(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-REVIEW-MANUAL", 2500, "needs_review", "paid")
	orderRecord.FulfillmentHold = true
	require.NoError(t, db.Save(&orderRecord).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "pi_review_manual", 2500, "USD")

	require.NoError(t, db.Create(&paymentdomain.PaymentReview{
		OrderID:         &orderRecord.ID,
		TransactionID:   &transaction.ID,
		PaymentIntentID: transaction.TransactionID,
		Status:          "approved",
		Reason:          highValueLiabilityReviewReason,
		Source:          "gateway_verification",
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.PaymentReview{
		OrderID:         &orderRecord.ID,
		PaymentIntentID: transaction.TransactionID,
		StripeReviewID:  "prv_review_manual",
		Status:          "pending",
		Reason:          stripeRadarReviewReason,
		Source:          "radar",
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.PaymentReview{
		OrderID: &orderRecord.ID,
		Status:  "pending",
		Reason:  "manual_payment_review",
		Source:  "operator",
	}).Error)

	require.NoError(t, paymentService.ResolveStripeReview("prv_review_manual", transaction.TransactionID, "approved"))

	var stillHeld order.Order
	require.NoError(t, db.First(&stillHeld, orderRecord.ID).Error)
	assert.Equal(t, "needs_review", stillHeld.Status)
	assert.True(t, stillHeld.FulfillmentHold)
}

func TestRecordVerifiedGatewayPaymentEmitsOneVerifiedConversionWithOrderAttribution(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-ATTRIBUTION", 84, "pending", "unpaid")
	require.NoError(t, db.Create(&attributiondomain.OrderAttribution{
		OrderID:     orderRecord.ID,
		Source:      "newsletter",
		Campaign:    "summer",
		ClickIDKind: "gclid",
		ClickID:     "click_123",
		CapturedAt:  time.Now().UTC(),
	}).Error)

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_attr_bad_amount",
		Amount:        domainmoney.MustNew(8399, "USD"),
	})
	require.Error(t, err)
	var beforeCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_type = ?", outboxdomain.EventTypeVerifiedConversion).Count(&beforeCount).Error)
	require.Zero(t, beforeCount)

	input := VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_attr_success",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(input))
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(input))

	var events []outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", outboxdomain.EventTypeVerifiedConversion).Find(&events).Error)
	require.Len(t, events, 1)
	var payload outboxdomain.VerifiedConversionPayload
	require.NoError(t, json.Unmarshal([]byte(events[0].Payload), &payload))
	require.Equal(t, orderRecord.ID, payload.OrderID)
	require.NotNil(t, payload.Attribution)
	require.Equal(t, "newsletter", payload.Attribution.Source)
	require.Equal(t, "summer", payload.Attribution.Campaign)
	require.Equal(t, "click_123", payload.Attribution.ClickID)
}

func TestEnsureGatewayPaymentAttemptReusesProviderRequestKeyForSameAttempt(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-ATTEMPT-KEY", 84, "pending", "unpaid")

	first, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:           "stripe",
		OrderNumber:        orderRecord.OrderNumber,
		AttemptKey:         "checkout-attempt-1",
		ProviderRequestKey: PaymentProviderRequestKey("stripe", orderRecord.ID, "checkout-attempt-1"),
		PaymentMethod:      "stripe",
		Amount:             domainmoney.MustNew(8400, "USD"),
	})
	require.NoError(t, err)

	second, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		AttemptKey:    "checkout-attempt-1",
		PaymentMethod: "stripe",
		Amount:        domainmoney.MustNew(8400, "USD"),
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.ProviderRequestKey, second.ProviderRequestKey)
	assert.Equal(t, first.TransactionID, second.TransactionID)

	var count int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).
		Where("order_id = ? AND payment_method = ? AND attempt_key = ?", orderRecord.ID, "stripe", "checkout-attempt-1").
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestEnsureGatewayPaymentAttemptRejectsMissingAttemptKey(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-ATTEMPT-KEY-MISSING", 84, "pending", "unpaid")

	_, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		PaymentMethod: "stripe",
		Amount:        domainmoney.MustNew(8400, "USD"),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "payment attempt key is required")
}

func TestEnsureGatewayPaymentAttemptCreatesDifferentProviderRequestKeyForNewAttempt(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-ATTEMPT-KEY-2", 84, "pending", "unpaid")

	first, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:      "paypal",
		OrderNumber:   orderRecord.OrderNumber,
		AttemptKey:    "checkout-attempt-1",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(8400, "USD"),
	})
	require.NoError(t, err)

	second, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:      "paypal",
		OrderNumber:   orderRecord.OrderNumber,
		AttemptKey:    "checkout-attempt-2",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(8400, "USD"),
	})
	require.NoError(t, err)

	assert.NotEqual(t, first.ID, second.ID)
	assert.NotEqual(t, first.ProviderRequestKey, second.ProviderRequestKey)
	assert.NotEqual(t, first.TransactionID, second.TransactionID)
}

func TestEnsureGatewayPaymentAttemptKeepsExistingDurableKeyWhenCallerSuppliesAnotherKey(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-ATTEMPT-KEY-3", 84, "pending", "unpaid")

	first, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:           "stripe",
		OrderNumber:        orderRecord.OrderNumber,
		AttemptKey:         "checkout-attempt-1",
		ProviderRequestKey: "provider-key-original",
		PaymentMethod:      "stripe",
		Amount:             domainmoney.MustNew(8400, "USD"),
	})
	require.NoError(t, err)

	second, err := paymentService.EnsureGatewayPaymentAttempt(EnsureGatewayPaymentAttemptInput{
		Provider:           "stripe",
		OrderNumber:        orderRecord.OrderNumber,
		AttemptKey:         "checkout-attempt-1",
		ProviderRequestKey: "provider-key-forged-or-raced",
		PaymentMethod:      "stripe",
		Amount:             domainmoney.MustNew(8400, "USD"),
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	assert.Equal(t, "provider-key-original", second.ProviderRequestKey)
}

func TestRecordVerifiedGatewayPaymentCompletesExistingPendingAttempt(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-PENDING-ATTEMPT", 84, "pending", "unpaid")

	require.NoError(t, paymentService.RecordGatewayPaymentAttempt(GatewayPaymentAttemptInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "pi_pending_attempt",
		Status:        "requires_action",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}))

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "pi_pending_attempt",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}))

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "pi_pending_attempt").First(&savedTransaction).Error)
	assert.Equal(t, "completed", savedTransaction.Status)
	assert.NotNil(t, savedTransaction.CompletedAt)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)
	assert.Equal(t, "processing", savedOrder.Status)
}

func TestRecordVerifiedGatewayPaymentCreatesReviewForExpiredOrderLateSuccess(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-LATE-SUCCESS", 84, "payment_expired", "expired")
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_late_success",
		PaymentMethod: "stripe",
		Amount:        84,
		Currency:      "USD",
		Status:        "expired",
	}).Error)

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "pi_late_success",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}))

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "payment_expired", savedOrder.Status)
	assert.Equal(t, "expired", savedOrder.PaymentStatus)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "pi_late_success").First(&savedTransaction).Error)
	assert.Equal(t, "completed", savedTransaction.Status)

	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("payment_intent_id = ? AND reason = ?", "pi_late_success", "payment_succeeded_after_expiration").First(&review).Error)
	assert.Equal(t, "pending", review.Status)
}

func TestRecordVerifiedGatewayPaymentCreatesReviewForCancelledOrderLateSuccess(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-CANCELLED-LATE-SUCCESS", 84, "cancelled", "unpaid")

	input := VerifiedGatewayPaymentInput{
		Provider:      "paypal",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "capture_cancelled_late",
		PaymentMethod: "paypal",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(input))
	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(input))

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "cancelled", savedOrder.Status)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", input.TransactionID).First(&savedTransaction).Error)
	assert.Equal(t, "completed", savedTransaction.Status)
	assert.Equal(t, "paypal", savedTransaction.PaymentMethod)

	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("payment_intent_id = ? AND reason = ?", input.TransactionID, "payment_succeeded_after_cancellation").First(&review).Error)
	assert.Equal(t, "pending", review.Status)
	require.NotNil(t, review.TransactionID)
	assert.Equal(t, savedTransaction.ID, *review.TransactionID)

	var reviewCount int64
	require.NoError(t, db.Model(&paymentdomain.PaymentReview{}).
		Where("payment_intent_id = ? AND reason = ?", input.TransactionID, "payment_succeeded_after_cancellation").
		Count(&reviewCount).Error)
	assert.Equal(t, int64(1), reviewCount)
}

func TestApproveLatePaymentReviewCreatesPendingRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-CANCELLED-REVIEW-REFUND", 84, "cancelled", "unpaid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "capture_cancelled_review_refund", 84, "USD")
	review := paymentdomain.PaymentReview{
		OrderID:         &orderRecord.ID,
		TransactionID:   &transaction.ID,
		PaymentIntentID: transaction.TransactionID,
		Status:          "pending",
		Reason:          "payment_succeeded_after_cancellation",
		Source:          "webhook",
	}
	require.NoError(t, db.Create(&review).Error)

	_, err := paymentService.UpdatePaymentReview(review.ID, "approved", "Refund approved", 7)
	require.NoError(t, err)
	var refunds []paymentdomain.Refund
	require.NoError(t, db.Where("transaction_id = ?", transaction.ID).Find(&refunds).Error)
	require.Len(t, refunds, 1)
	assert.Equal(t, "pending", refunds[0].Status)
	assert.Equal(t, "late_payment_refund:payment_succeeded_after_cancellation", refunds[0].Reason)
}

func TestRecordVerifiedGatewayPaymentCreatesReviewForRefundedOrderLateSuccess(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-REFUNDED-LATE-SUCCESS", 84, "refunded", "refunded")

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "pi_refunded_late",
		Amount:        domainmoney.MustNew(8400, "USD"),
	}))

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "pi_refunded_late").First(&savedTransaction).Error)
	assert.Equal(t, "completed", savedTransaction.Status)

	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("payment_intent_id = ? AND reason = ?", "pi_refunded_late", "payment_succeeded_after_refund").First(&review).Error)
	assert.Equal(t, "pending", review.Status)
}

func TestRecordVerifiedGatewayPaymentNormalizesCurrency(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-LOWER-CURRENCY", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_lower_currency",
		Amount:        mustTestMoney(t, 84, " usd "),
	})

	require.NoError(t, err)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "txn_lower_currency").First(&savedTransaction).Error)
	assert.Equal(t, "USD", savedTransaction.Currency)
}

func TestRecordVerifiedGatewayPaymentRejectsAmountMismatch(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-2", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_bad_amount",
		Amount:        domainmoney.MustNew(8399, "USD"),
	})

	require.Error(t, err)

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("transaction_id = ?", "txn_bad_amount").Count(&transactionCount).Error)
	assert.Equal(t, int64(0), transactionCount)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)
	assert.Equal(t, "pending", savedOrder.Status)
}

func TestRecordVerifiedGatewayPaymentRejectsCurrencyMismatch(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-BAD-CURRENCY", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_bad_currency",
		Amount:        domainmoney.MustNew(8400, "EUR"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "transaction currency EUR does not match order currency USD")

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("transaction_id = ?", "txn_bad_currency").Count(&transactionCount).Error)
	assert.Equal(t, int64(0), transactionCount)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)
	assert.Equal(t, "pending", savedOrder.Status)
}

func TestRecordVerifiedGatewayPaymentUsesProviderSettlementSnapshot(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-CNY-SNAPSHOT", 1500, "pending", "unpaid")
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", orderRecord.ID).Updates(map[string]interface{}{
		"payment_currency":     "CNY",
		"payment_amount_minor": int64(1080000),
	}).Error)

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "wechat",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "wx_snapshot_1",
		Amount:        domainmoney.MustNew(1080000, "CNY"),
	}))

	var saved order.Order
	require.NoError(t, db.First(&saved, orderRecord.ID).Error)
	assert.Equal(t, "USD", saved.Currency)
	assert.Equal(t, "CNY", saved.PaymentCurrency)
	assert.Equal(t, "paid", saved.PaymentStatus)
}

func TestRecordVerifiedGatewayPaymentRejectsMissingCrossCurrencySnapshot(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-CNY-NO-SNAPSHOT", 1500, "pending", "unpaid")
	orderRecord.PaymentAmount = 0
	orderRecord.PaymentCurrency = ""
	require.NoError(t, db.Save(&orderRecord).Error)

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "wechat",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "wx_snapshot_missing",
		Amount:        domainmoney.MustNew(1080000, "CNY"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "order payment currency snapshot is required")
}

func TestCreateAdminRefundReservesPendingAmount(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-1", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_1", 100, "USD")

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        60,
		Reason:        "customer request",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	assert.Equal(t, "pending", refund.Status)
	assert.Equal(t, uint(7), refund.RefundedBy)
	assert.Nil(t, refund.RefundID)

	excessRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        50,
	}
	require.Error(t, paymentService.CreateAdminRefund(&excessRefund, 7))

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("transaction_id = ?", transaction.ID).Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestCreateAdminRefundWithIdempotencyReplaysDurableRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-IDEMPOTENT", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_idempotent", 100, "USD")

	first, err := paymentService.CreateAdminRefundWithIdempotency(
		&paymentdomain.Refund{
			OrderID:       orderRecord.ID,
			TransactionID: transaction.ID,
			Amount:        60,
			Reason:        "customer request",
		},
		7,
		"refund-request-1",
		"request-hash-1",
	)
	require.NoError(t, err)

	second, err := paymentService.CreateAdminRefundWithIdempotency(
		&paymentdomain.Refund{
			OrderID:       orderRecord.ID,
			TransactionID: transaction.ID,
			Amount:        60,
			Reason:        "customer request",
		},
		7,
		"refund-request-1",
		"request-hash-1",
	)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NotNil(t, second)
	assert.Equal(t, first.ID, second.ID)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("transaction_id = ?", transaction.ID).Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)

	_, err = paymentService.CreateAdminRefundWithIdempotency(
		&paymentdomain.Refund{
			OrderID:       orderRecord.ID,
			TransactionID: transaction.ID,
			Amount:        40,
			Reason:        "different request",
		},
		7,
		"refund-request-1",
		"request-hash-2",
	)
	assert.ErrorIs(t, err, ErrPaymentRefundIdempotencyConflict)
}

func TestCreateAdminRefundCreatesItemLevelRefundAndPreventsOverRefundQuantity(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-LINE-1", 300, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_line_1", 300, "USD")
	orderItem := seedPaymentOrderItem(t, db, orderRecord.ID, 2, 150, 300, 0, 0, 300)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Reason:        "return one rim",
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1, Restock: true},
		},
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	assert.InDelta(t, 150, refund.RequestedAmount, 0.001)
	assert.InDelta(t, 150, refund.Amount, 0.001)
	require.Len(t, refund.LineItems, 1)
	assert.Equal(t, orderItem.ID, refund.LineItems[0].OrderItemID)
	assert.Equal(t, 1, refund.LineItems[0].Quantity)
	assert.True(t, refund.LineItems[0].Restock)
	assert.InDelta(t, 150, refund.LineItems[0].LineSubtotalAmount, 0.001)
	assert.InDelta(t, 150, refund.LineItems[0].LineTotalAmount, 0.001)

	secondRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 2},
		},
	}
	require.Error(t, paymentService.CreateAdminRefund(&secondRefund, 7))

	var savedLineItems []paymentdomain.RefundLineItem
	require.NoError(t, db.Where("refund_id = ?", refund.ID).Find(&savedLineItems).Error)
	require.Len(t, savedLineItems, 1)
	assert.Equal(t, refund.ID, savedLineItems[0].RefundID)
}

func TestCreateAdminRefundAllowsItemLevelRefundAfterAmountOnlyRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-MIXED-AMOUNT-FIRST", 300, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_mixed_amount_first", 300, "USD")
	orderItem := seedPaymentOrderItem(t, db, orderRecord.ID, 2, 150, 300, 0, 0, 300)

	amountRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        30,
		Reason:        "repair compensation",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&amountRefund, 7))

	itemRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1},
		},
		Reason: "return one rim",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&itemRefund, 7))

	assert.InDelta(t, 30, amountRefund.RequestedAmount, 0.001)
	assert.InDelta(t, 150, itemRefund.RequestedAmount, 0.001)
}

func TestCreateAdminRefundAllowsAmountOnlyRefundAfterItemLevelRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-MIXED-ITEM-FIRST", 300, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_mixed_item_first", 300, "USD")
	orderItem := seedPaymentOrderItem(t, db, orderRecord.ID, 2, 150, 300, 0, 0, 300)

	itemRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1},
		},
		Reason: "return one rim",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&itemRefund, 7))

	amountRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        30,
		Reason:        "repair compensation",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&amountRefund, 7))

	assert.InDelta(t, 150, itemRefund.RequestedAmount, 0.001)
	assert.InDelta(t, 30, amountRefund.RequestedAmount, 0.001)
}

func TestRecordVerifiedGatewayRefundMixedTypesRestocksOnlyItemLevelRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	variant := seedPaymentProductVariant(t, db, 5)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-MIXED-COMPLETE", 300, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_mixed_complete", 300, "USD")
	orderItem := seedPaymentOrderItemWithVariant(t, db, orderRecord.ID, variant.ProductID, variant.ID, 2, 150, 300, 0, 0, 300)

	amountRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        30,
		Reason:        "repair compensation",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&amountRefund, 7))

	itemRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1, Restock: true},
		},
		Reason: "return one rim",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&itemRefund, 7))

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "rf_mixed_amount",
		ProviderRefundAmount: mustTestMoney(t, 30, "USD"),
	}))

	var savedVariant productdomain.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 5, savedVariant.Stock)

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "rf_mixed_item",
		ProviderRefundAmount: mustTestMoney(t, 150, "USD"),
	}))

	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 6, savedVariant.Stock)

	var savedAmountRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedAmountRefund, amountRefund.ID).Error)
	assert.Empty(t, savedAmountRefund.LineItems)

	var savedItemRefund paymentdomain.Refund
	require.NoError(t, db.Preload("LineItems").First(&savedItemRefund, itemRefund.ID).Error)
	require.Len(t, savedItemRefund.LineItems, 1)
	assert.NotNil(t, savedItemRefund.LineItems[0].RestockedAt)
}

func TestCreateAdminRefundClawsBackCouponDiscountUsingItemSubtotal(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-LINE-COUPON", 1000, "processing", "paid")
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", orderRecord.ID).Updates(map[string]interface{}{
		"subtotal_amount": 1100.0,
		"discount_amount": 100.0,
		"coupon_code":     "SAVE100",
	}).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_line_coupon", 1000, "USD")
	mainItem := seedPaymentOrderItem(t, db, orderRecord.ID, 1, 950, 950, 0, 0, 950)
	_ = mainItem
	accessory := seedPaymentOrderItem(t, db, orderRecord.ID, 1, 150, 150, 0, 0, 150)

	promo := seedPaymentCoupon(t, db, "SAVE100", "fixed", 100, 1000, 0)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID: promo.ID,
		UserID:   orderRecord.UserID,
		OrderID:  orderRecord.ID,
		Discount: 100,
	}).Error)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: accessory.ID, Quantity: 1},
		},
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	assert.InDelta(t, 150, refund.RequestedAmount, 0.001)
	assert.InDelta(t, 100, refund.DiscountClawbackAmount, 0.001)
	assert.InDelta(t, 50, refund.Amount, 0.001)
	assert.Contains(t, refund.CalculationSnapshot, `"requested_subtotal_amount":150`)
}

func TestCreateAdminRefundClawsBackCouponDiscountWhenPartialRefundBreaksThreshold(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-COUPON-1", 1000, "processing", "paid")
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", orderRecord.ID).Updates(map[string]interface{}{
		"subtotal_amount": 1100.0,
		"discount_amount": 100.0,
		"coupon_code":     "SAVE100",
	}).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_coupon_1", 1000, "USD")

	promo := seedPaymentCoupon(t, db, "SAVE100", "fixed", 100, 1000, 0)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID: promo.ID,
		UserID:   orderRecord.UserID,
		OrderID:  orderRecord.ID,
		Discount: 100,
	}).Error)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        150,
		Reason:        "return accessories",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	assert.Equal(t, "pending", refund.Status)
	assert.InDelta(t, 150, refund.RequestedAmount, 0.001)
	assert.InDelta(t, 100, refund.DiscountClawbackAmount, 0.001)
	assert.InDelta(t, 50, refund.Amount, 0.001)
	assert.Contains(t, refund.CalculationSnapshot, "coupon_recalculation")

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	assert.InDelta(t, 150, savedRefund.RequestedAmount, 0.001)
	assert.InDelta(t, 100, savedRefund.DiscountClawbackAmount, 0.001)
	assert.InDelta(t, 50, savedRefund.Amount, 0.001)
}

func TestCreateAdminRefundAllowsZeroNetRefundAfterCouponClawback(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-COUPON-ZERO", 1450, "processing", "paid")
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", orderRecord.ID).Updates(map[string]interface{}{
		"subtotal_amount": 1650.0,
		"discount_amount": 200.0,
		"coupon_code":     "SAVE200",
	}).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_coupon_zero", 1450, "USD")
	seedPaymentOrderItem(t, db, orderRecord.ID, 1, 1500, 1500, 0, 0, 1500)
	accessory := seedPaymentOrderItem(t, db, orderRecord.ID, 1, 150, 150, 0, 0, 150)

	promo := seedPaymentCoupon(t, db, "SAVE200", "fixed", 200, 1600, 0)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID: promo.ID,
		UserID:   orderRecord.UserID,
		OrderID:  orderRecord.ID,
		Discount: 200,
	}).Error)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: accessory.ID, Quantity: 1},
		},
		Reason: "return accessories",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	assert.Equal(t, "pending", refund.Status)
	assert.InDelta(t, 150, refund.RequestedAmount, 0.001)
	assert.InDelta(t, 150, refund.DiscountClawbackAmount, 0.001)
	assert.InDelta(t, 0, refund.Amount, 0.001)
	assert.Contains(t, refund.CalculationSnapshot, `"net_refund_amount":0`)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	assert.Equal(t, "pending", savedRefund.Status)
	assert.InDelta(t, 150, savedRefund.RequestedAmount, 0.001)
	assert.InDelta(t, 150, savedRefund.DiscountClawbackAmount, 0.001)
	assert.InDelta(t, 0, savedRefund.Amount, 0.001)
}

func TestCreateAdminRefundDoesNotDoubleClawBackCouponDiscount(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-COUPON-2", 1000, "processing", "paid")
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", orderRecord.ID).Updates(map[string]interface{}{
		"subtotal_amount": 1100.0,
		"discount_amount": 100.0,
		"coupon_code":     "SAVE100",
	}).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_coupon_2", 1000, "USD")

	promo := seedPaymentCoupon(t, db, "SAVE100", "fixed", 100, 1000, 0)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID: promo.ID,
		UserID:   orderRecord.UserID,
		OrderID:  orderRecord.ID,
		Discount: 100,
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.Refund{
		OrderID:                orderRecord.ID,
		TransactionID:          transaction.ID,
		Amount:                 50,
		RequestedAmount:        150,
		DiscountClawbackAmount: 100,
		Status:                 "completed",
	}).Error)

	secondRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        50,
		Reason:        "second partial return",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&secondRefund, 7))

	assert.InDelta(t, 50, secondRefund.RequestedAmount, 0.001)
	assert.InDelta(t, 0, secondRefund.DiscountClawbackAmount, 0.001)
	assert.InDelta(t, 50, secondRefund.Amount, 0.001)
}

func TestRecordVerifiedGatewayRefundCompletesPendingRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-2", 84, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_2", 84, "USD")
	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        84,
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	err := paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "rf_123",
		ProviderRefundAmount: mustTestMoney(t, 84, "USD"),
	})
	require.NoError(t, err)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	assert.Equal(t, "completed", savedRefund.Status)
	require.NotNil(t, savedRefund.RefundID)
	assert.Equal(t, "rf_123", *savedRefund.RefundID)
	assert.NotNil(t, savedRefund.CompletedAt)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.First(&savedTransaction, transaction.ID).Error)
	assert.Equal(t, "refunded", savedTransaction.Status)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "refunded", savedOrder.PaymentStatus)
	assert.Equal(t, "refunded", savedOrder.Status)

	var invalidatedEvent outboxdomain.Event
	require.NoError(t, db.Where(
		"event_key = ?",
		fmt.Sprintf("%s:%d:%s:%s", outboxdomain.EventTypeReferralOrderInvalidated, orderRecord.ID, "refund_webhook", "rf_123"),
	).First(&invalidatedEvent).Error)
	assert.Equal(t, outboxdomain.EventTypeReferralOrderInvalidated, invalidatedEvent.EventType)

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "rf_123",
		ProviderRefundAmount: mustTestMoney(t, 84, "USD"),
	}))

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("refund_id = ?", "rf_123").Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestRecordVerifiedGatewayRefundReplayByRefundIDDoesNotCreateDuplicatePartialRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-IDEMPOTENT-PARTIAL", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_idempotent_partial", 100, "USD")
	refundID := "rf_idempotent_partial"
	require.NoError(t, db.Create(&paymentdomain.Refund{
		OrderID:         orderRecord.ID,
		TransactionID:   transaction.ID,
		RefundID:        &refundID,
		Amount:          40,
		RequestedAmount: 40,
		Status:          "completed",
	}).Error)

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:              "stripe",
		OrderNumber:           orderRecord.OrderNumber,
		TransactionID:         transaction.TransactionID,
		RefundID:              refundID,
		ProviderRefundAmount:  mustTestMoney(t, 40, "USD"),
		RequestedRefundAmount: mustTestMoney(t, 40, "USD"),
	}))

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).
		Where("refund_id = ?", refundID).
		Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestRecordVerifiedGatewayRefundRestocksItemLevelRefundOnce(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	cacheInvalidator := &recordingProductCacheInvalidator{}
	paymentService.ConfigureProductCacheInvalidator(cacheInvalidator)
	variant := seedPaymentProductVariant(t, db, 5)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-RESTOCK", 300, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_restock", 300, "USD")
	orderItem := seedPaymentOrderItemWithVariant(t, db, orderRecord.ID, variant.ProductID, variant.ID, 2, 150, 300, 0, 0, 300)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1, Restock: true},
		},
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	input := VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "rf_restock_once",
		ProviderRefundAmount: mustTestMoney(t, 150, "USD"),
	}
	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(input))

	var savedVariant productdomain.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 6, savedVariant.Stock)

	var savedProduct productdomain.Product
	require.NoError(t, db.First(&savedProduct, variant.ProductID).Error)
	assert.Equal(t, 5, savedProduct.Stock)

	var savedLineItem paymentdomain.RefundLineItem
	require.NoError(t, db.Where("refund_id = ?", refund.ID).First(&savedLineItem).Error)
	assert.NotNil(t, savedLineItem.RestockedAt)

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(input))
	assert.Equal(t, []uint{variant.ProductID}, cacheInvalidator.productIDs)
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 6, savedVariant.Stock)
	assert.Equal(t, []uint{variant.ProductID}, cacheInvalidator.productIDs)

	var cacheEvent outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", outboxdomain.EventTypeProductCacheInvalidate).First(&cacheEvent).Error)
	assert.Contains(t, cacheEvent.EventKey, "refund_stock_restored")
}

func TestCreateAdminRefundIgnoresSoftDeletedLineItemRefundQuantity(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-SOFT-DELETED", 300, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_soft_deleted", 300, "USD")
	orderItem := seedPaymentOrderItem(t, db, orderRecord.ID, 1, 300, 300, 0, 0, 300)

	deletedRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        300,
		Status:        "completed",
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1, ProductID: orderItem.ProductID, VariantID: orderItem.VariantID, ProductName: orderItem.ProductName, SKU: orderItem.SKU, UnitPrice: 300, LineSubtotalAmount: 300, LineTotalAmount: 300},
		},
	}
	require.NoError(t, repository.NewPaymentRepository(db).CreateRefund(&deletedRefund))
	require.NoError(t, db.Delete(&paymentdomain.Refund{}, deletedRefund.ID).Error)

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		LineItems: []paymentdomain.RefundLineItem{
			{OrderItemID: orderItem.ID, Quantity: 1},
		},
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))
	assert.InDelta(t, 300, refund.Amount, 0.001)
}

func TestRecordVerifiedGatewayRefundRejectsOverRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-3", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_3", 100, "USD")

	err := paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:             "stripe",
		OrderNumber:          orderRecord.OrderNumber,
		TransactionID:        transaction.TransactionID,
		RefundID:             "rf_too_much",
		ProviderRefundAmount: mustTestMoney(t, 101, "USD"),
	})

	require.Error(t, err)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("refund_id = ?", "rf_too_much").Count(&refundCount).Error)
	assert.Equal(t, int64(0), refundCount)
}

func TestPaymentServicePublicTaxRatesOnlyReturnEnabledRates(t *testing.T) {
	db, paymentService := newTestPaymentService(t)

	enabledRate := paymentdomain.TaxRate{
		Name:    "Enabled",
		Country: "US",
		State:   "CA",
		Rate:    7.5,
		Enabled: true,
	}
	disabledRate := paymentdomain.TaxRate{
		Name:    "Disabled",
		Country: "US",
		State:   "NY",
		Rate:    8.5,
		Enabled: false,
	}
	require.NoError(t, db.Create(&enabledRate).Error)
	require.NoError(t, db.Create(&disabledRate).Error)
	require.NoError(t, db.Model(&paymentdomain.TaxRate{}).Where("id = ?", disabledRate.ID).Update("enabled", false).Error)

	rates, err := paymentService.ListPublicTaxRates()
	require.NoError(t, err)
	require.Len(t, rates, 1)
	assert.Equal(t, enabledRate.ID, rates[0].ID)

	_, err = paymentService.GetPublicTaxRate(disabledRate.ID)
	require.ErrorIs(t, err, ErrPaymentNotFound)
}

func TestPaymentServiceCalculateTaxPrefersPostalCodeAndFallsBackToDefault(t *testing.T) {
	db, paymentService := newTestPaymentService(t)

	defaultRate := paymentdomain.TaxRate{
		Name:    "California default",
		Country: "US",
		State:   "CA",
		Rate:    7.25,
		Enabled: true,
	}
	postalRate := paymentdomain.TaxRate{
		Name:       "Beverly Hills",
		Country:    "US",
		State:      "CA",
		PostalCode: "90210",
		Rate:       9.5,
		Enabled:    true,
	}
	require.NoError(t, db.Create(&defaultRate).Error)
	require.NoError(t, db.Create(&postalRate).Error)

	amount, err := domainmoney.New(10000, "USD")
	require.NoError(t, err)
	rate, taxMoney, err := paymentService.CalculateTaxMoney(amount, "us", "ca", "90210")
	require.NoError(t, err)
	tax, err := taxMoney.MajorFloat()
	require.NoError(t, err)
	assert.InDelta(t, 9.5, rate, 0.001)
	assert.InDelta(t, 9.5, tax, 0.001)

	rate, taxMoney, err = paymentService.CalculateTaxMoney(amount, "US", "CA", "10001")
	require.NoError(t, err)
	tax, err = taxMoney.MajorFloat()
	require.NoError(t, err)
	assert.InDelta(t, 7.25, rate, 0.001)
	assert.InDelta(t, 7.25, tax, 0.001)
}

func TestStripeWebhookEventClaimIsIdempotentAndFailedEventsRetry(t *testing.T) {
	db, paymentService := newTestPaymentService(t)

	claimed, err := paymentService.ClaimStripeWebhookEvent("evt_1", "payment_intent.succeeded", "{}")
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = paymentService.ClaimStripeWebhookEvent("evt_1", "payment_intent.succeeded", "{}")
	require.NoError(t, err)
	assert.False(t, claimed)

	require.NoError(t, paymentService.MarkStripeWebhookEventFailed("evt_1", assert.AnError))

	claimed, err = paymentService.ClaimStripeWebhookEvent("evt_1", "payment_intent.succeeded", `{"retry":true}`)
	require.NoError(t, err)
	assert.True(t, claimed)

	require.NoError(t, paymentService.MarkStripeWebhookEventProcessed("evt_1"))

	var saved paymentdomain.StripeWebhookEvent
	require.NoError(t, db.Where("event_id = ?", "evt_1").First(&saved).Error)
	assert.Equal(t, "processed", saved.Status)
	assert.NotNil(t, saved.ProcessedAt)
}

func TestPaymentReviewDeduplicatesPaymentIntentAndResolvesStripeReview(t *testing.T) {
	db, paymentService := newTestPaymentService(t)

	first, err := paymentService.CreatePaymentReview(CreatePaymentReviewInput{
		PaymentIntentID: "pi_review_1",
		Status:          "pending",
		Reason:          "stripe_requires_action",
		Source:          "radar",
		Notes:           "requires action",
	})
	require.NoError(t, err)

	second, err := paymentService.CreatePaymentReview(CreatePaymentReviewInput{
		PaymentIntentID: "pi_review_1",
		StripeReviewID:  "prv_1",
		Status:          "pending",
		Reason:          "stripe_review_opened",
		Source:          "radar",
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "prv_1", second.StripeReviewID)

	var reviewCount int64
	require.NoError(t, db.Model(&paymentdomain.PaymentReview{}).Where("payment_intent_id = ?", "pi_review_1").Count(&reviewCount).Error)
	assert.Equal(t, int64(1), reviewCount)

	require.NoError(t, paymentService.ResolveStripeReview("prv_1", "pi_review_1", "approved"))

	var saved paymentdomain.PaymentReview
	require.NoError(t, db.First(&saved, first.ID).Error)
	assert.Equal(t, "approved", saved.Status)
	assert.NotNil(t, saved.ReviewedAt)
}

func TestResolveStripeRequiresActionReviewApprovesOnlyMatchingPendingReview(t *testing.T) {
	db, paymentService := newTestPaymentService(t)

	review, err := paymentService.CreatePaymentReview(CreatePaymentReviewInput{
		PaymentIntentID: "pi_requires_action_success",
		Status:          "pending",
		Reason:          "stripe_requires_action",
		Source:          "radar",
		Notes:           "requires action",
	})
	require.NoError(t, err)
	require.NoError(t, db.Create(&paymentdomain.PaymentReview{
		PaymentIntentID: "pi_requires_action_success",
		Status:          "pending",
		Reason:          "stripe_review_opened",
		Source:          "radar",
	}).Error)

	require.NoError(t, paymentService.ResolveStripeRequiresActionReview(" pi_requires_action_success "))

	var saved paymentdomain.PaymentReview
	require.NoError(t, db.First(&saved, review.ID).Error)
	assert.Equal(t, "approved", saved.Status)
	assert.NotNil(t, saved.ReviewedAt)
	assert.Contains(t, saved.Notes, "Stripe 3DS authentication completed successfully.")

	var other paymentdomain.PaymentReview
	require.NoError(t, db.Where(
		"payment_intent_id = ? AND reason = ?",
		"pi_requires_action_success",
		"stripe_review_opened",
	).First(&other).Error)
	assert.Equal(t, "pending", other.Status)

	require.NoError(t, paymentService.ResolveStripeRequiresActionReview("pi_requires_action_success"))
}

func TestResolveStripeRequiresActionReviewIsIdempotentWhenMissing(t *testing.T) {
	_, paymentService := newTestPaymentService(t)

	require.NoError(t, paymentService.ResolveStripeRequiresActionReview(""))
	require.NoError(t, paymentService.ResolveStripeRequiresActionReview("pi_without_review"))
}

func TestRecordStripeDisputeCreatesReviewWhenResponseNeeded(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-DISPUTE-1", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "pi_dispute_1", 100, "USD")

	dispute, err := paymentService.RecordStripeDispute(StripeDisputeInput{
		StripeDisputeID: "dp_1",
		PaymentIntentID: transaction.TransactionID,
		Amount:          100,
		Currency:        "USD",
		Reason:          "fraudulent",
		Status:          "needs_response",
		RawPayload:      "{}",
	})
	require.NoError(t, err)
	assert.Equal(t, orderRecord.ID, *dispute.OrderID)
	assert.Equal(t, transaction.ID, *dispute.TransactionID)

	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("dispute_id = ? AND status = ?", dispute.ID, "pending").First(&review).Error)
	assert.Equal(t, "stripe_dispute", review.Reason)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "disputed", savedOrder.Status)
	assert.True(t, savedOrder.FulfillmentHold)

	var invalidatedEvent outboxdomain.Event
	require.NoError(t, db.Where(
		"event_key = ?",
		fmt.Sprintf("%s:%d:%s:%s", outboxdomain.EventTypeReferralOrderInvalidated, orderRecord.ID, "stripe_dispute", "dp_1"),
	).First(&invalidatedEvent).Error)
	assert.Equal(t, outboxdomain.EventTypeReferralOrderInvalidated, invalidatedEvent.EventType)
}

func TestRecordStripeDisputePreservesEvidenceSubmissionAudit(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-DISPUTE-2", 120, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "pi_dispute_2", 120, "USD")

	dispute, err := paymentService.RecordStripeDispute(StripeDisputeInput{
		StripeDisputeID: "dp_preserve_evidence",
		PaymentIntentID: transaction.TransactionID,
		Amount:          120,
		Currency:        "USD",
		Reason:          "product_not_received",
		Status:          "needs_response",
		RawPayload:      `{"version":1}`,
	})
	require.NoError(t, err)

	submittedAt := time.Now().UTC()
	require.NoError(t, paymentService.paymentRepo.UpdateStripeDisputeEvidenceSubmission(
		dispute.ID,
		&submittedAt,
		`{"shipping_documentation_file_id":"file_pod"}`,
		"",
		"under_review",
	))

	updated, err := paymentService.RecordStripeDispute(StripeDisputeInput{
		StripeDisputeID: "dp_preserve_evidence",
		PaymentIntentID: transaction.TransactionID,
		Amount:          120,
		Currency:        "USD",
		Reason:          "product_not_received",
		Status:          "lost",
		RawPayload:      `{"version":2}`,
	})
	require.NoError(t, err)
	assert.Equal(t, dispute.ID, updated.ID)

	var saved paymentdomain.StripeDispute
	require.NoError(t, db.First(&saved, dispute.ID).Error)
	require.NotNil(t, saved.EvidenceSubmittedAt)
	assert.WithinDuration(t, submittedAt, *saved.EvidenceSubmittedAt, time.Second)
	assert.Equal(t, `{"shipping_documentation_file_id":"file_pod"}`, saved.EvidenceSubmissionPayload)
	assert.Empty(t, saved.EvidenceSubmissionError)
	assert.Equal(t, "lost", saved.Status)
	assert.Equal(t, `{"version":2}`, saved.RawPayload)
}

func TestRecordStripeDisputeRestoresPreDisputeShippingStateWhenWon(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-DISPUTE-RESTORE", 100, "shipped", "paid")
	orderRecord.FulfillmentHold = false
	require.NoError(t, db.Save(&orderRecord).Error)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "pi_dispute_restore", 100, "USD")

	dispute, err := paymentService.RecordStripeDispute(StripeDisputeInput{
		StripeDisputeID: "dp_restore",
		PaymentIntentID: transaction.TransactionID,
		Amount:          100,
		Currency:        "USD",
		Status:          "needs_response",
	})
	require.NoError(t, err)
	var held order.Order
	require.NoError(t, db.First(&held, orderRecord.ID).Error)
	assert.Equal(t, "disputed", held.Status)
	assert.Equal(t, "shipped", held.DisputePreviousStatus)
	assert.True(t, held.FulfillmentHold)

	_, err = paymentService.RecordStripeDispute(StripeDisputeInput{
		StripeDisputeID: "dp_restore",
		PaymentIntentID: transaction.TransactionID,
		Amount:          100,
		Currency:        "USD",
		Status:          "won",
	})
	require.NoError(t, err)
	var restored order.Order
	require.NoError(t, db.First(&restored, orderRecord.ID).Error)
	assert.Equal(t, "shipped", restored.Status)
	assert.False(t, restored.FulfillmentHold)
	assert.Empty(t, restored.DisputePreviousStatus)
	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("dispute_id = ?", dispute.ID).First(&review).Error)
	assert.Equal(t, "cancelled", review.Status)
}

func TestBuildStripeDisputeEvidencePackageCollectsOrderShippingAndCommunication(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 42, "rider@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-EVIDENCE-1", customer.ID)
	disputeRecord := seedStripeDispute(t, db, "dp_evidence_1", orderRecord.ID, "needs_response")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL123")
	seedCustomerCommunication(t, db, orderRecord.ID, customer.ID, orderRecord.OrderNumber)

	pkg, err := paymentService.BuildStripeDisputeEvidencePackage(disputeRecord.ID)

	require.NoError(t, err)
	require.NotNil(t, pkg.Order)
	assert.Equal(t, orderRecord.OrderNumber, pkg.Order.OrderNumber)
	assert.True(t, pkg.CanSubmit)
	assert.Equal(t, "rider@example.test", pkg.Evidence.CustomerEmailAddress)
	assert.Contains(t, pkg.Evidence.ProductDescription, "Carbon wheelset")
	assert.Equal(t, "DHL123", pkg.Evidence.ShippingTrackingNumber)
	assert.Contains(t, pkg.Evidence.UncategorizedText, "Delivered tracking event")
	require.NotEmpty(t, pkg.Communications)
	assert.Contains(t, pkg.Evidence.CommunicationSummary, "Please confirm delivery")
	require.Len(t, pkg.TrackingEventEvidence, 1)
	assert.Equal(t, "DHL123", pkg.TrackingEventEvidence[0].TrackingNumber)
	assert.NotNil(t, pkg.TrackingContext)
	assert.NotNil(t, pkg.TrackingContext.LatestDeliveryEvent)

	payload, err := json.Marshal(pkg)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"tracking_events"`)
	assert.Contains(t, string(payload), "Delivered tracking event")
	assert.NotContains(t, string(payload), "api_key")
}

func TestSubmitStripeDisputeEvidenceCallsStripeAndRecordsSubmission(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 43, "submit@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-EVIDENCE-2", customer.ID)
	disputeRecord := seedStripeDispute(t, db, "dp_submit_1", orderRecord.ID, "needs_response")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL999")
	fakeSubmitter := &fakeStripeDisputeEvidenceSubmitter{status: stripe.DisputeStatusUnderReview}
	paymentService.stripeDisputeEvidenceSubmitter = fakeSubmitter

	result, err := paymentService.SubmitStripeDisputeEvidence(nil, SubmitStripeDisputeEvidenceInput{
		DisputeID:                    disputeRecord.ID,
		APIKey:                       "sk_test_fake",
		Confirm:                      true,
		Submit:                       true,
		IncludeCustomerCommunication: true,
		AdditionalStatement:          "Manual operator note.",
		ShippingDocumentationFileID:  "file_shipping",
		CustomerCommunicationFileID:  "file_chat",
	})

	require.NoError(t, err)
	require.NotNil(t, result.SubmittedAt)
	assert.Equal(t, "under_review", result.StripeStatus)
	assert.Equal(t, "dp_submit_1", fakeSubmitter.disputeID)
	require.NotNil(t, fakeSubmitter.params)
	require.NotNil(t, fakeSubmitter.params.Submit)
	assert.True(t, *fakeSubmitter.params.Submit)
	require.NotNil(t, fakeSubmitter.params.Evidence)
	assert.Equal(t, "file_shipping", stripe.StringValue(fakeSubmitter.params.Evidence.ShippingDocumentation))
	assert.Equal(t, "file_chat", stripe.StringValue(fakeSubmitter.params.Evidence.CustomerCommunication))
	assert.Contains(t, stripe.StringValue(fakeSubmitter.params.Evidence.UncategorizedText), "Manual operator note.")

	var saved paymentdomain.StripeDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.NotNil(t, saved.EvidenceSubmittedAt)
	assert.Equal(t, "under_review", saved.Status)
	assert.Empty(t, saved.EvidenceSubmissionError)
	assert.Contains(t, saved.EvidenceSubmissionPayload, "file_shipping")
}

func TestSubmitStripeDisputeEvidenceRequiresConfirmation(t *testing.T) {
	_, paymentService := newTestPaymentService(t)

	_, err := paymentService.SubmitStripeDisputeEvidence(nil, SubmitStripeDisputeEvidenceInput{
		DisputeID: 1,
		APIKey:    "sk_test_fake",
		Confirm:   false,
		Submit:    true,
	})

	require.ErrorIs(t, err, ErrStripeDisputeEvidenceConfirmRequired)
}

func TestTruncateEvidenceTextPreservesUTF8(t *testing.T) {
	value := truncateEvidenceText("客户已确认签收凭证和售后沟通记录，客服已经发送订单号和物流轨迹截图。", 24)

	assert.True(t, utf8.ValidString(value))
	assert.Contains(t, value, "[truncated]")
}

func newTestPaymentService(t *testing.T) (*gorm.DB, *PaymentService) {
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
		&userdomain.User{},
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&coupon.GiftCard{},
		&coupon.GiftCardTransaction{},
		&loyalty.ProgramConfig{},
		&loyalty.ProgramRedeemOption{},
		&loyalty.UserLoyalty{},
		&loyalty.LoyaltyTransaction{},
		&productdomain.Product{},
		&productdomain.ProductVariant{},
		&order.Order{},
		&order.OrderItem{},
		&order.PolicyDisclosure{},
		&orderevidence.OrderEvidenceSnapshot{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceItem{},
		&orderevidence.OrderEvidenceAttachment{},
		&orderevidence.OrderEvidenceSubmissionSnapshot{},
		&attributiondomain.OrderAttribution{},
		&outboxdomain.Event{},
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
		&paymentdomain.RefundIdempotency{},
		&paymentdomain.RefundLineItem{},
		&paymentdomain.TaxRate{},
		&paymentdomain.StripeWebhookEvent{},
		&paymentdomain.StripeDispute{},
		&paymentdomain.PayPalDispute{},
		&paymentdomain.PaymentReview{},
		&shippingdomain.TrackingProviderConfig{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
		&ticketdomain.Ticket{},
		&ticketdomain.TicketMessage{},
	))

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	orderEvidenceSubmissionRepo := repository.NewOrderEvidenceSubmissionSnapshotRepository(db)
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo, shippingRepo)
	txManager.ConfigureLoyaltyProgramRepository(repository.NewLoyaltyProgramRepository(db))
	txManager.ConfigureOutboxRepository(outboxRepo)
	txManager.ConfigureOrderAttributionRepository(repository.NewOrderAttributionRepository(db))
	txManager.ConfigurePaymentRefundIdempotencyRepository(repository.NewPaymentRefundIdempotencyRepository(db))
	txManager.ConfigureOrderEvidenceSubmissionSnapshotRepository(orderEvidenceSubmissionRepo)
	policyDisclosureRepo := repository.NewOrderPolicyDisclosureRepository(db)

	paymentService := NewPaymentService(txManager, paymentRepo)
	paymentService.ConfigureEvidenceSources(orderRepo, ticketRepo)
	paymentService.ConfigureOrderEvidenceSubmissionSnapshotRepository(orderEvidenceSubmissionRepo)
	paymentService.ConfigureOrderEvidenceAssembler(
		NewOrderEvidencePackageAssembler(
			orderRepo,
			repository.NewOrderEvidenceRepository(db),
			shippingRepo,
		),
	)
	paymentService.ConfigurePolicyDisclosureRepository(policyDisclosureRepo)
	return db, paymentService
}

func seedPaymentOrder(t *testing.T, db *gorm.DB, orderNumber string, total float64, status, paymentStatus string) order.Order {
	t.Helper()

	record := order.Order{
		OrderNumber:     orderNumber,
		UserID:          42,
		Status:          status,
		PaymentStatus:   paymentStatus,
		TotalAmount:     total,
		Currency:        "USD",
		PaymentAmount:   total,
		PaymentCurrency: "USD",
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedCompletedTransaction(t *testing.T, db *gorm.DB, orderID uint, transactionID string, amount float64, currency string) paymentdomain.Transaction {
	t.Helper()

	completedAt := time.Now()
	record := paymentdomain.Transaction{
		OrderID:       orderID,
		TransactionID: transactionID,
		PaymentMethod: "stripe",
		Amount:        amount,
		Currency:      currency,
		Status:        "completed",
		CompletedAt:   &completedAt,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedPaymentOrderItem(t *testing.T, db *gorm.DB, orderID uint, quantity int, price float64, subtotal float64, taxAmount float64, discount float64, total float64) order.OrderItem {
	t.Helper()

	variantID := uint(1)
	return seedPaymentOrderItemWithVariant(t, db, orderID, 1, variantID, quantity, price, subtotal, taxAmount, discount, total)
}

func seedPaymentOrderItemWithVariant(t *testing.T, db *gorm.DB, orderID uint, productID uint, variantID uint, quantity int, price float64, subtotal float64, taxAmount float64, discount float64, total float64) order.OrderItem {
	t.Helper()

	record := order.OrderItem{
		OrderID:     orderID,
		ProductID:   productID,
		VariantID:   &variantID,
		ProductName: "Carbon component",
		SKU:         "TEST-SKU",
		Quantity:    quantity,
		Price:       price,
		Subtotal:    subtotal,
		TaxAmount:   taxAmount,
		Discount:    discount,
		Total:       total,
	}
	record.PricingSnapshotData = refundPricingLineForTest(t, productID, &variantID, quantity, price, discount, "USD")
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedPaymentProductVariant(t *testing.T, db *gorm.DB, stock int) productdomain.ProductVariant {
	t.Helper()

	productRecord := productdomain.Product{
		SKU:    "TEST-PRODUCT",
		Name:   "Test Product",
		Slug:   "test-product",
		Price:  150,
		Stock:  stock,
		Status: "active",
		Locale: "en",
	}
	require.NoError(t, db.Create(&productRecord).Error)

	variant := productdomain.ProductVariant{
		ProductID:    productRecord.ID,
		SKU:          "TEST-VARIANT",
		Title:        "Test Variant",
		OptionValues: `{"size":"test"}`,
		Price:        150,
		Stock:        stock,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)
	return variant
}

func seedPaymentCoupon(t *testing.T, db *gorm.DB, code string, couponType string, value float64, minAmount float64, maxDiscount float64) coupon.Coupon {
	t.Helper()

	record := coupon.Coupon{
		Code:        code,
		Type:        couponType,
		Value:       value,
		MinAmount:   minAmount,
		MaxDiscount: maxDiscount,
		Enabled:     true,
		StartDate:   time.Now().Add(-24 * time.Hour),
		EndDate:     time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

type fakeStripeDisputeEvidenceSubmitter struct {
	disputeID     string
	params        *stripe.DisputeParams
	paramsHistory []*stripe.DisputeParams
	status        stripe.DisputeStatus
	err           error
}

func (f *fakeStripeDisputeEvidenceSubmitter) Update(id string, params *stripe.DisputeParams) (*stripe.Dispute, error) {
	f.disputeID = id
	f.params = params
	f.paramsHistory = append(f.paramsHistory, params)
	if f.err != nil {
		return nil, f.err
	}
	return &stripe.Dispute{Status: f.status}, nil
}

func TestRecordPayPalDisputeLinksTransactionAndOrder(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-DP-1", 42)
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "PAYPAL-CAPTURE-1", orderRecord.TotalAmount, "USD")
	transaction.PaymentMethod = "paypal"
	require.NoError(t, db.Save(&transaction).Error)

	dispute, err := paymentService.RecordPayPalDispute(PayPalDisputeInput{
		PayPalDisputeID:   "PP-D-1",
		ProviderPaymentID: "PAYPAL-CAPTURE-1",
		Reason:            "MERCHANDISE_OR_SERVICE_NOT_RECEIVED",
		Status:            "WAITING_FOR_SELLER_RESPONSE",
		DisputeState:      "REQUIRED_ACTION",
	})

	require.NoError(t, err)
	require.NotNil(t, dispute.OrderID)
	require.NotNil(t, dispute.TransactionID)
	assert.Equal(t, orderRecord.ID, *dispute.OrderID)
	assert.Equal(t, transaction.ID, *dispute.TransactionID)
	assert.Equal(t, orderRecord.TotalAmount, dispute.Amount)
	assert.Equal(t, "USD", dispute.Currency)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "disputed", savedOrder.Status)
	assert.True(t, savedOrder.FulfillmentHold)

	var invalidatedEvent outboxdomain.Event
	require.NoError(t, db.Where(
		"event_key = ?",
		fmt.Sprintf("%s:%d:%s:%s", outboxdomain.EventTypeReferralOrderInvalidated, orderRecord.ID, "paypal_dispute", "PP-D-1"),
	).First(&invalidatedEvent).Error)
	assert.Equal(t, outboxdomain.EventTypeReferralOrderInvalidated, invalidatedEvent.EventType)

	var review paymentdomain.PaymentReview
	require.NoError(t, db.Where("order_id = ? AND status = ?", orderRecord.ID, "pending").First(&review).Error)
	assert.Equal(t, "paypal_dispute", review.Reason)
}

func TestBuildPayPalDisputeEvidencePackageCollectsTrackingInvoiceAndCommunication(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 44, "paypal-evidence@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-EVIDENCE-1", customer.ID)
	orderRecord.SignatureRequired = true
	require.NoError(t, db.Save(&orderRecord).Error)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-EVIDENCE-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL777")
	seedCustomerCommunication(t, db, orderRecord.ID, customer.ID, orderRecord.OrderNumber)

	pkg, err := paymentService.BuildPayPalDisputeEvidencePackage(disputeRecord.ID)

	require.NoError(t, err)
	require.NotNil(t, pkg.Order)
	require.NotNil(t, pkg.FulfillmentEvidence)
	require.NotNil(t, pkg.FulfillmentEvidence.Shipment)
	assert.True(t, pkg.CanSubmit)
	assert.Equal(t, "DHL777", pkg.Evidence.ShippingTrackingNumber)
	assert.Equal(t, "DHL", pkg.Evidence.ShippingCarrier)
	assert.Contains(t, pkg.Evidence.InvoiceSummary, orderRecord.OrderNumber)
	assert.Contains(t, pkg.Evidence.ProofOfDeliverySummary, "Delivered and signed by recipient")
	assert.Contains(t, pkg.Evidence.ProofOfDeliverySummary, "recipient_signature_name=Test Rider")
	assert.Contains(t, pkg.Evidence.ProofOfDeliverySummary, "proof_of_delivery_url=https://carrier.example.test/pod/DHL777.pdf")
	assert.Contains(t, pkg.Evidence.Notes, "Invoice summary:")
	assert.Contains(t, pkg.Evidence.Notes, "Proof of delivery summary")
	assert.Contains(t, pkg.Evidence.Notes, "PayPal order-policy Signature POD")
	assert.Contains(t, pkg.Evidence.Notes, "Customer communication summary")
}

func TestBuildPayPalDisputeEvidencePackageWarnsOrderPolicySignatureWithoutSignaturePOD(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 53, "paypal-signature-pod@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-SIGNATURE-POD-1", customer.ID)
	orderRecord.SignatureRequired = true
	require.NoError(t, db.Save(&orderRecord).Error)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-SIGNATURE-POD-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidenceWithPOD(t, db, orderRecord.ID, "DHL778", "", "")

	pkg, err := paymentService.BuildPayPalDisputeEvidencePackage(disputeRecord.ID)

	require.NoError(t, err)
	assert.Equal(t, float64(1235), orderRecord.TotalAmount)
	assert.True(t, orderRecord.SignatureRequired)
	require.True(t, pkg.CanSubmit)
	require.NotNil(t, pkg.FulfillmentEvidence)
	require.NotNil(t, pkg.FulfillmentEvidence.Shipment)
	require.True(t, pkg.SubmissionCheck.Ready)
	require.False(t, pkg.SubmissionCheck.OverrideRequired)
	assert.Empty(t, pkg.SubmissionCheck.Blockers)
	assert.Contains(t, strings.Join(pkg.SubmissionCheck.Warnings, "\n"), "收件人签名或官方 POD")
	assert.Contains(t, strings.Join(pkg.Warnings, "\n"), "order shipping policy")
	assert.Contains(t, pkg.Evidence.ProofOfDeliverySummary, "Delivery scan")
	assert.NotContains(t, pkg.Evidence.ProofOfDeliverySummary, "Signature POD")
	assert.NotContains(t, pkg.Evidence.ProofOfDeliverySummary, "recipient_signature_name=")
	assert.Contains(t, pkg.Evidence.Notes, "Signature POD: missing")
}

func TestBuildPayPalDisputeCommercialInvoicePDFReturnsPDF(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 48, "paypal-invoice-preview@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-INVOICE-PREVIEW-1", customer.ID)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-INVOICE-PREVIEW-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	paymentService.ConfigurePayPalDisputeInvoiceOptions(PayPalDisputeInvoiceOptions{
		Seller: invoice.SellerProfile{
			Name:    "Commerce Platform Factory",
			Address: "100 Factory Road\nAustin, TX 78701\nUS",
			Email:   "support@example.test",
		},
	})

	pdf, err := paymentService.BuildPayPalDisputeCommercialInvoicePDF(disputeRecord.ID)

	require.NoError(t, err)
	require.NotNil(t, pdf)
	assert.Equal(t, disputeRecord.ID, pdf.DisputeID)
	assert.Equal(t, "PP-D-INVOICE-PREVIEW-1", pdf.PayPalDisputeID)
	assert.Contains(t, pdf.Filename, "CI-ORD-PAYPAL-INVOICE-PREVIEW-1")
	require.True(t, bytes.HasPrefix(pdf.Bytes, []byte("%PDF-")))
}

func TestSubmitPayPalDisputeEvidenceCallsPayPalAndRecordsSubmission(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 45, "paypal-submit@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-EVIDENCE-2", customer.ID)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-SUBMIT-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL888")
	fakeSubmitter := &fakePayPalDisputeEvidenceSubmitter{}
	paymentService.paypalDisputeEvidenceSubmitter = fakeSubmitter

	result, err := paymentService.SubmitPayPalDisputeEvidence(nil, SubmitPayPalDisputeEvidenceInput{
		DisputeID:           disputeRecord.ID,
		ClientID:            "paypal-client",
		SecretKey:           "paypal-secret",
		Environment:         "sandbox",
		AdditionalStatement: "Order shipped and delivered to the confirmed shipping address.",
	})

	require.NoError(t, err)
	require.NotNil(t, result.SubmittedAt)
	assert.Equal(t, "PP-D-SUBMIT-1", fakeSubmitter.disputeID)
	require.NotNil(t, fakeSubmitter.params)
	require.NotNil(t, fakeSubmitter.params.Evidences)
	assert.Equal(t, paypalapi.EvidenceTypeProofOfFulfillment, fakeSubmitter.params.Evidences.EvidenceType)
	require.NotNil(t, fakeSubmitter.params.Evidences.EvidenceInfo)
	require.Len(t, fakeSubmitter.params.Evidences.EvidenceInfo.TrackingInfo, 1)
	assert.Equal(t, "DHL888", fakeSubmitter.params.Evidences.EvidenceInfo.TrackingInfo[0].TrackingNumber)
	assert.Equal(t, "DHL", fakeSubmitter.params.Evidences.EvidenceInfo.TrackingInfo[0].CarrierName)
	assert.Contains(t, fakeSubmitter.params.Evidences.Notes, "confirmed shipping address")
	assert.Contains(t, fakeSubmitter.params.Evidences.Notes, "Invoice summary")

	var saved paymentdomain.PayPalDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.NotNil(t, saved.EvidenceSubmittedAt)
	assert.Empty(t, saved.EvidenceSubmissionError)
	assert.Contains(t, saved.EvidenceSubmissionPayload, "DHL888")
}

func TestSubmitPayPalDisputeEvidenceDoesNotAttachCommercialInvoicePDFWhenDisabled(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 49, "paypal-invoice-disabled@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-INVOICE-DISABLED-1", customer.ID)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-INVOICE-DISABLED-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL890")
	fakeSubmitter := &fakePayPalDisputeEvidenceSubmitter{}
	fakeStorage := &fakePayPalDisputeDocumentStorage{url: "https://cdn.example.test/evidence/commercial-invoice.pdf"}
	paymentService.ConfigurePayPalDisputeEvidenceSubmitter(fakeSubmitter)
	paymentService.ConfigurePayPalDisputeEvidenceDocumentStorage(fakeStorage)
	paymentService.ConfigurePayPalDisputeInvoiceOptions(PayPalDisputeInvoiceOptions{
		Seller: invoice.SellerProfile{
			Name:    "Commerce Platform Factory",
			Address: "100 Factory Road\nAustin, TX 78701\nUS",
		},
		AutoAttachPDF: false,
	})

	result, err := paymentService.SubmitPayPalDisputeEvidence(context.Background(), SubmitPayPalDisputeEvidenceInput{
		DisputeID:   disputeRecord.ID,
		ClientID:    "paypal-client",
		SecretKey:   "paypal-secret",
		Environment: "sandbox",
	})

	require.NoError(t, err)
	require.Empty(t, result.Documents)
	require.NotNil(t, fakeSubmitter.params)
	require.NotNil(t, fakeSubmitter.params.Evidences)
	require.Empty(t, fakeSubmitter.params.Evidences.Documents)
	require.Empty(t, fakeStorage.data)

	var saved paymentdomain.PayPalDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.Contains(t, saved.EvidenceSubmissionPayload, "auto-attachment is not enabled")
}

func TestSubmitPayPalDisputeEvidenceAttachesCommercialInvoicePDF(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 47, "paypal-invoice@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-INVOICE-1", customer.ID)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-INVOICE-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL889")
	fakeSubmitter := &fakePayPalDisputeEvidenceSubmitter{}
	fakeStorage := &fakePayPalDisputeDocumentStorage{url: "https://cdn.example.test/evidence/commercial-invoice.pdf"}
	paymentService.ConfigurePayPalDisputeEvidenceSubmitter(fakeSubmitter)
	paymentService.ConfigurePayPalDisputeEvidenceDocumentStorage(fakeStorage)
	paymentService.ConfigurePayPalDisputeInvoiceOptions(PayPalDisputeInvoiceOptions{
		Seller: invoice.SellerProfile{
			Name:    "Commerce Platform Factory",
			Address: "100 Factory Road\nAustin, TX 78701\nUS",
			Email:   "support@example.test",
		},
		AutoAttachPDF: true,
	})

	result, err := paymentService.SubmitPayPalDisputeEvidence(context.Background(), SubmitPayPalDisputeEvidenceInput{
		DisputeID:   disputeRecord.ID,
		ClientID:    "paypal-client",
		SecretKey:   "paypal-secret",
		Environment: "sandbox",
	})

	require.NoError(t, err)
	require.Len(t, result.Documents, 1)
	assert.Equal(t, "commercial_invoice", result.Documents[0].Type)
	require.NotNil(t, fakeSubmitter.params)
	require.NotNil(t, fakeSubmitter.params.Evidences)
	require.Len(t, fakeSubmitter.params.Evidences.Documents, 1)
	assert.Equal(t, "https://cdn.example.test/evidence/commercial-invoice.pdf", fakeSubmitter.params.Evidences.Documents[0].URL)
	assert.Contains(t, fakeSubmitter.params.Evidences.Documents[0].Name, "CI-ORD-PAYPAL-INVOICE-1")
	require.True(t, bytes.HasPrefix(fakeStorage.data, []byte("%PDF-")))

	var saved paymentdomain.PayPalDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.Contains(t, saved.EvidenceSubmissionPayload, "commercial_invoice")
	assert.Contains(t, saved.EvidenceSubmissionPayload, "https://cdn.example.test/evidence/commercial-invoice.pdf")
}

func TestSubmitPayPalDisputeEvidenceUsesCompactInvoiceWhenCommercialInvoiceExceedsBudget(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 50, "paypal-invoice-budget@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-INVOICE-BUDGET-1", customer.ID)
	seedPayPalDisputeEvidenceOrderItems(t, db, orderRecord.ID, 4)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-INVOICE-BUDGET-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL991")
	fakeSubmitter := &fakePayPalDisputeEvidenceSubmitter{}
	fakeStorage := &fakePayPalDisputeDocumentStorage{url: "https://cdn.example.test/evidence/commercial-invoice-budget.pdf"}
	paymentService.ConfigurePayPalDisputeEvidenceSubmitter(fakeSubmitter)
	paymentService.ConfigurePayPalDisputeEvidenceDocumentStorage(fakeStorage)
	paymentService.ConfigurePayPalDisputeInvoiceOptions(PayPalDisputeInvoiceOptions{
		Seller: invoice.SellerProfile{
			Name:    "Commerce Platform Factory",
			Address: "100 Factory Road\nAustin, TX 78701\nUS",
			Email:   "support@example.test",
		},
		AutoAttachPDF: true,
	})
	paymentService.paypalDisputeCommercialInvoiceRenderer = func(document invoice.CommercialInvoice, _ string) ([]byte, error) {
		if strings.TrimSpace(document.PaymentReference) == "" {
			return []byte("%PDF-compact"), nil
		}
		return bytes.Repeat([]byte("x"), paypalDisputeCommercialInvoiceMaxBytes+1), nil
	}

	result, err := paymentService.SubmitPayPalDisputeEvidence(context.Background(), SubmitPayPalDisputeEvidenceInput{
		DisputeID:   disputeRecord.ID,
		ClientID:    "paypal-client",
		SecretKey:   "paypal-secret",
		Environment: "sandbox",
	})

	require.NoError(t, err)
	require.Len(t, result.Documents, 1)
	assert.Equal(t, "commercial_invoice", result.Documents[0].Type)
	require.NotNil(t, fakeSubmitter.params)
	require.NotNil(t, fakeSubmitter.params.Evidences)
	require.Len(t, fakeSubmitter.params.Evidences.Documents, 1)
	assert.Equal(t, "https://cdn.example.test/evidence/commercial-invoice-budget.pdf", fakeSubmitter.params.Evidences.Documents[0].URL)
	require.Equal(t, []byte("%PDF-compact"), fakeStorage.data)

	var saved paymentdomain.PayPalDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.Contains(t, saved.EvidenceSubmissionPayload, "compact fallback was used before upload")
}

func TestSubmitPayPalDisputeEvidenceAllowsMissingTrackingWithoutWarningOverride(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 46, "paypal-no-track@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-EVIDENCE-3", customer.ID)
	orderRecord.TrackingNumber = ""
	orderRecord.ProviderCarrierCode = ""
	orderRecord.ProviderCarrierName = ""
	require.NoError(t, db.Save(&orderRecord).Error)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-NO-TRACK-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	fakeSubmitter := &fakePayPalDisputeEvidenceSubmitter{}
	paymentService.ConfigurePayPalDisputeEvidenceSubmitter(fakeSubmitter)

	result, err := paymentService.SubmitPayPalDisputeEvidence(nil, SubmitPayPalDisputeEvidenceInput{
		DisputeID: disputeRecord.ID,
		ClientID:  "paypal-client",
		SecretKey: "paypal-secret",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, fakeSubmitter.params)
	require.NotNil(t, fakeSubmitter.params.Evidences)
	require.Empty(t, fakeSubmitter.params.Evidences.EvidenceInfo.TrackingInfo)
	var saved paymentdomain.PayPalDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.Empty(t, saved.EvidenceSubmissionError)
	assert.Contains(t, saved.EvidenceSubmissionPayload, "submission_warnings")
	assert.Contains(t, saved.EvidenceSubmissionPayload, `"override_warnings":false`)
}

type fakePayPalDisputeEvidenceSubmitter struct {
	disputeID     string
	params        *paypalapi.DisputeProvideEvidenceParams
	paramsHistory []*paypalapi.DisputeProvideEvidenceParams
	err           error
}

func (f *fakePayPalDisputeEvidenceSubmitter) ProvideEvidence(_ context.Context, disputeID string, params *paypalapi.DisputeProvideEvidenceParams) error {
	f.disputeID = disputeID
	f.params = params
	f.paramsHistory = append(f.paramsHistory, params)
	return f.err
}

type fakePayPalDisputeDocumentStorage struct {
	url      string
	filename string
	data     []byte
	uploads  int
}

func (f *fakePayPalDisputeDocumentStorage) UploadFromReader(_ context.Context, reader io.Reader, filename string) (string, error) {
	f.filename = filename
	f.uploads++
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	f.data = data
	return f.url, nil
}

func seedPaymentUser(t *testing.T, db *gorm.DB, id uint, email string) userdomain.User {
	t.Helper()

	record := userdomain.User{
		ID:        id,
		Email:     email,
		Username:  email,
		Password:  "hashed",
		FirstName: "Test",
		LastName:  "Rider",
		Role:      "user",
		Status:    "active",
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedDisputeEvidenceOrder(t *testing.T, db *gorm.DB, orderNumber string, userID uint) order.Order {
	t.Helper()

	shippedAt := time.Now().Add(-72 * time.Hour)
	paidAt := time.Now().Add(-96 * time.Hour)
	variantID := uint(1)
	record := order.Order{
		OrderNumber:         orderNumber,
		UserID:              userID,
		Status:              "shipped",
		PaymentStatus:       "paid",
		ShippingStatus:      "delivered",
		TrackingNumber:      "DHL123",
		ProviderCarrierCode: "DHL",
		ProviderCarrierName: "DHL Express",
		SubtotalAmount:      1200,
		ShippingFee:         35,
		TotalAmount:         1235,
		Currency:            "USD",
		PaidAt:              &paidAt,
		ShippedAt:           &shippedAt,
		ShippingAddress: order.Address{
			FirstName:  "Test",
			LastName:   "Rider",
			Address1:   "1 Carbon Road",
			City:       "Los Angeles",
			State:      "CA",
			PostalCode: "90001",
			Country:    "US",
			Email:      "rider@example.test",
		},
		BillingAddress: order.Address{
			FirstName:  "Test",
			LastName:   "Rider",
			Address1:   "1 Carbon Road",
			City:       "Los Angeles",
			State:      "CA",
			PostalCode: "90001",
			Country:    "US",
			Email:      "rider@example.test",
		},
		Items: []order.OrderItem{
			{
				ProductID:   1,
				VariantID:   &variantID,
				ProductName: "Carbon wheelset",
				SKU:         "C50-DT240",
				Quantity:    1,
				Price:       1200,
				Subtotal:    1200,
				Total:       1200,
			},
		},
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedStripeDispute(t *testing.T, db *gorm.DB, stripeID string, orderID uint, status string) paymentdomain.StripeDispute {
	t.Helper()

	dueAt := time.Now().Add(5 * 24 * time.Hour)
	record := paymentdomain.StripeDispute{
		StripeDisputeID: stripeID,
		OrderID:         &orderID,
		PaymentIntentID: "pi_" + stripeID,
		Amount:          1235,
		Currency:        "USD",
		Reason:          "fraudulent",
		Status:          status,
		EvidenceDueAt:   &dueAt,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedPayPalDispute(t *testing.T, db *gorm.DB, paypalID string, orderID uint, status string, disputeState string) paymentdomain.PayPalDispute {
	t.Helper()

	record := paymentdomain.PayPalDispute{
		PayPalDisputeID:       paypalID,
		OrderID:               &orderID,
		ProviderPaymentID:     "capture_" + paypalID,
		Amount:                1235,
		Currency:              "USD",
		Reason:                "MERCHANDISE_OR_SERVICE_NOT_RECEIVED",
		Status:                status,
		DisputeState:          disputeState,
		DisputeLifeCycleStage: "CHARGEBACK",
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedTrackingEvidence(t *testing.T, db *gorm.DB, orderID uint, trackingNumber string) {
	t.Helper()

	seedTrackingEvidenceWithPOD(
		t,
		db,
		orderID,
		trackingNumber,
		"Test Rider",
		"https://carrier.example.test/pod/"+trackingNumber+".pdf",
	)
}

func seedTrackingEvidenceWithPOD(t *testing.T, db *gorm.DB, orderID uint, trackingNumber string, recipientSignatureName string, proofOfDeliveryURL string) {
	t.Helper()

	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Tracking",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             orderID,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		EventCount:          2,
		Enabled:             true,
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:                orderID,
		TrackingNumber:         trackingNumber,
		ProviderCarrierCode:    "DHL",
		Status:                 "Delivered",
		Location:               "Los Angeles, US",
		Description:            "Delivered and signed by recipient",
		RecipientSignatureName: recipientSignatureName,
		ProofOfDeliveryURL:     proofOfDeliveryURL,
		EventTime:              time.Now().Add(-24 * time.Hour),
	}).Error)
}

func seedCustomerCommunication(t *testing.T, db *gorm.DB, orderID, userID uint, orderNumber string) {
	t.Helper()

	customerID := userID
	conversation := ticketdomain.Ticket{
		UserID:         userID,
		CustomerUserID: &customerID,
		Subject:        "Question about " + orderNumber,
		Category:       "customer_service",
		Status:         "resolved",
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticketdomain.TicketMessage{
		TicketID:    conversation.ID,
		UserID:      &userID,
		IsStaff:     false,
		Content:     fmt.Sprintf("Please confirm delivery for order %s.", orderNumber),
		MessageType: "text",
		CreatedAt:   time.Now().Add(-48 * time.Hour),
	}).Error)
	_ = orderID
}

func seedPayPalDisputeEvidenceOrderItems(t *testing.T, db *gorm.DB, orderID uint, count int) {
	t.Helper()

	variantID := uint(1)
	for i := 0; i < count; i++ {
		require.NoError(t, db.Create(&order.OrderItem{
			OrderID:     orderID,
			ProductID:   uint(100 + i),
			VariantID:   &variantID,
			ProductName: fmt.Sprintf("Accessory %d", i+1),
			SKU:         fmt.Sprintf("ACC-%d", i+1),
			Quantity:    1,
			Price:       1,
			Subtotal:    1,
			Total:       1,
		}).Error)
	}
}

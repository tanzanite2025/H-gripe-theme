package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"
	"unicode/utf8"

	attributiondomain "commerce-platform/internal/domain/attribution"
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/order"
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

func TestRecordVerifiedGatewayPaymentCreatesLedgerAndMarksOrderPaid(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-1", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_123",
		Amount:        84,
		Currency:      "USD",
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
		Amount:        84,
		Currency:      "USD",
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

	var outboxCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_key = ?", event.EventKey).Count(&outboxCount).Error)
	assert.Equal(t, int64(1), outboxCount)

	var conversion outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", fmt.Sprintf("%s:%d:%s", outboxdomain.EventTypeVerifiedConversion, orderRecord.ID, "txn_123")).First(&conversion).Error)
	assert.Equal(t, outboxdomain.EventTypeVerifiedConversion, conversion.EventType)
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
		Amount:        83.99,
		Currency:      "USD",
	})
	require.Error(t, err)
	var beforeCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Where("event_type = ?", outboxdomain.EventTypeVerifiedConversion).Count(&beforeCount).Error)
	require.Zero(t, beforeCount)

	input := VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_attr_success",
		Amount:        84,
		Currency:      "USD",
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

func TestRecordVerifiedGatewayPaymentCompletesExistingPendingAttempt(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-PENDING-ATTEMPT", 84, "pending", "unpaid")

	require.NoError(t, paymentService.RecordGatewayPaymentAttempt(GatewayPaymentAttemptInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "pi_pending_attempt",
		Status:        "requires_action",
		Amount:        84,
		Currency:      "USD",
	}))

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "pi_pending_attempt",
		Amount:        84,
		Currency:      "USD",
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
		Amount:        84,
		Currency:      "USD",
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

func TestRecordVerifiedGatewayPaymentNormalizesCurrency(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-LOWER-CURRENCY", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_lower_currency",
		Amount:        84,
		Currency:      " usd ",
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
		Amount:        83.99,
		Currency:      "USD",
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
		Amount:        84,
		Currency:      "EUR",
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
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_123",
		Amount:        84,
		Currency:      "USD",
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

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_123",
		Amount:        84,
		Currency:      "USD",
	}))

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("refund_id = ?", "rf_123").Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestRecordVerifiedGatewayRefundRestocksItemLevelRefundOnce(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
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
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_restock_once",
		Amount:        150,
		Currency:      "USD",
	}
	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(input))

	var savedVariant productdomain.ProductVariant
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 6, savedVariant.Stock)

	var savedProduct productdomain.Product
	require.NoError(t, db.First(&savedProduct, variant.ProductID).Error)
	assert.Equal(t, 6, savedProduct.Stock)

	var savedLineItem paymentdomain.RefundLineItem
	require.NoError(t, db.Where("refund_id = ?", refund.ID).First(&savedLineItem).Error)
	assert.NotNil(t, savedLineItem.RestockedAt)

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(input))
	require.NoError(t, db.First(&savedVariant, variant.ID).Error)
	assert.Equal(t, 6, savedVariant.Stock)
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
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_too_much",
		Amount:        101,
		Currency:      "USD",
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
		&productdomain.Product{},
		&productdomain.ProductVariant{},
		&order.Order{},
		&order.OrderItem{},
		&attributiondomain.OrderAttribution{},
		&outboxdomain.Event{},
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
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
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo, shippingRepo)
	txManager.ConfigureOutboxRepository(outboxRepo)
	txManager.ConfigureOrderAttributionRepository(repository.NewOrderAttributionRepository(db))

	paymentService := NewPaymentService(txManager, paymentRepo)
	paymentService.ConfigureEvidenceSources(orderRepo, shippingRepo, ticketRepo)
	return db, paymentService
}

func seedPaymentOrder(t *testing.T, db *gorm.DB, orderNumber string, total float64, status, paymentStatus string) order.Order {
	t.Helper()

	record := order.Order{
		OrderNumber:   orderNumber,
		UserID:        42,
		Status:        status,
		PaymentStatus: paymentStatus,
		TotalAmount:   total,
		Currency:      "USD",
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
	disputeID string
	params    *stripe.DisputeParams
	status    stripe.DisputeStatus
}

func (f *fakeStripeDisputeEvidenceSubmitter) Update(id string, params *stripe.DisputeParams) (*stripe.Dispute, error) {
	f.disputeID = id
	f.params = params
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
}

func TestBuildPayPalDisputeEvidencePackageCollectsTrackingInvoiceAndCommunication(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 44, "paypal-evidence@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-EVIDENCE-1", customer.ID)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-EVIDENCE-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL777")
	seedCustomerCommunication(t, db, orderRecord.ID, customer.ID, orderRecord.OrderNumber)

	pkg, err := paymentService.BuildPayPalDisputeEvidencePackage(disputeRecord.ID)

	require.NoError(t, err)
	require.NotNil(t, pkg.Order)
	assert.True(t, pkg.CanSubmit)
	assert.Equal(t, "DHL777", pkg.Evidence.ShippingTrackingNumber)
	assert.Equal(t, "DHL", pkg.Evidence.ShippingCarrier)
	assert.Contains(t, pkg.Evidence.InvoiceSummary, orderRecord.OrderNumber)
	assert.Contains(t, pkg.Evidence.ProofOfDeliverySummary, "Delivered and signed by recipient")
	assert.Contains(t, pkg.Evidence.Notes, "Invoice summary:")
	assert.Contains(t, pkg.Evidence.Notes, "Proof of delivery / signature event")
	assert.Contains(t, pkg.Evidence.Notes, "Customer communication summary")
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

func TestSubmitPayPalDisputeEvidenceRequiresTracking(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 46, "paypal-no-track@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-EVIDENCE-3", customer.ID)
	orderRecord.TrackingNumber = ""
	orderRecord.ProviderCarrierCode = ""
	orderRecord.ProviderCarrierName = ""
	require.NoError(t, db.Save(&orderRecord).Error)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-NO-TRACK-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")

	_, err := paymentService.SubmitPayPalDisputeEvidence(nil, SubmitPayPalDisputeEvidenceInput{
		DisputeID: disputeRecord.ID,
		ClientID:  "paypal-client",
		SecretKey: "paypal-secret",
	})

	require.ErrorIs(t, err, ErrPayPalDisputeEvidenceTrackingNeeded)
	var saved paymentdomain.PayPalDispute
	require.NoError(t, db.First(&saved, disputeRecord.ID).Error)
	assert.Contains(t, saved.EvidenceSubmissionError, ErrPayPalDisputeEvidenceTrackingNeeded.Error())
}

type fakePayPalDisputeEvidenceSubmitter struct {
	disputeID string
	params    *paypalapi.DisputeProvideEvidenceParams
}

func (f *fakePayPalDisputeEvidenceSubmitter) ProvideEvidence(_ context.Context, disputeID string, params *paypalapi.DisputeProvideEvidenceParams) error {
	f.disputeID = disputeID
	f.params = params
	return nil
}

type fakePayPalDisputeDocumentStorage struct {
	url      string
	filename string
	data     []byte
}

func (f *fakePayPalDisputeDocumentStorage) UploadFromReader(_ context.Context, reader io.Reader, filename string) (string, error) {
	f.filename = filename
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
		OrderID:             orderID,
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: "DHL",
		Status:              "Delivered",
		Location:            "Los Angeles, US",
		Description:         "Delivered and signed by recipient",
		EventTime:           time.Now().Add(-24 * time.Hour),
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
		UserID:      userID,
		IsStaff:     false,
		Content:     fmt.Sprintf("Please confirm delivery for order %s.", orderNumber),
		MessageType: "text",
		CreatedAt:   time.Now().Add(-48 * time.Hour),
	}).Error)
	_ = orderID
}

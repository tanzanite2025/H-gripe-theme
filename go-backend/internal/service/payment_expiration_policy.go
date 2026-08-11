package service

import (
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"
)

func paymentExpirationStillEligible(orderRecord *order.Order, transactions []payment.Transaction, cutoff time.Time) bool {
	if orderRecord == nil || orderRecord.Status != "pending" || orderRecord.PaymentStatus != "unpaid" {
		return false
	}

	latestActivity := orderRecord.CreatedAt
	for _, transaction := range transactions {
		if transaction.Status == "completed" {
			return false
		}
		if transaction.UpdatedAt.After(latestActivity) {
			latestActivity = transaction.UpdatedAt
		}
	}
	return !latestActivity.After(cutoff)
}

package service

import (
	"time"
)

const defaultShippingQuoteCleanupBatchLimit = 500

// CleanupExpiredShippingQuoteSnapshots physically removes short-lived quote
// snapshots after their expiry. Checkout still treats an expired row as
// invalid when it races with this sweep, so cleanup is safe to retry.
func (s *ShippingService) CleanupExpiredShippingQuoteSnapshots(now time.Time, limit int) (int64, error) {
	if s == nil || s.shippingRepo == nil {
		return 0, errorsWithCause(ErrShippingRateUnavailable, "shipping quote repository is not configured")
	}
	if limit <= 0 {
		limit = defaultShippingQuoteCleanupBatchLimit
	}
	return s.shippingRepo.DeleteExpiredQuoteSnapshots(now.UTC(), limit)
}

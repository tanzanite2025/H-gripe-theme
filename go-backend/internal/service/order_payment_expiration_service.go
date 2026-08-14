package service

import (
	"time"

	"commerce-platform/internal/repository"
)

const (
	defaultPaymentExpirationTTL        = 30 * time.Minute
	defaultPaymentExpirationBatchLimit = 100
)

type PaymentExpirationCleanupResult struct {
	ScannedCandidates         int       `json:"scanned_candidates"`
	ExpiredOrders             int       `json:"expired_orders"`
	SkippedOrders             int       `json:"skipped_orders"`
	ExpiredOpenTransactions   int64     `json:"expired_open_transactions"`
	CleanupReferenceTimestamp time.Time `json:"cleanup_reference_timestamp"`
	Cutoff                    time.Time `json:"cutoff"`
}

func (s *OrderService) ExpireStalePendingPayments(now time.Time, ttl time.Duration, limit int) (PaymentExpirationCleanupResult, error) {
	result := PaymentExpirationCleanupResult{}
	if s == nil || s.orderRepo == nil || s.txManager == nil {
		return result, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if ttl <= 0 {
		ttl = defaultPaymentExpirationTTL
	}
	if limit <= 0 {
		limit = defaultPaymentExpirationBatchLimit
	}

	cutoff := now.Add(-ttl)
	result.CleanupReferenceTimestamp = now
	result.Cutoff = cutoff

	candidates, err := s.orderRepo.FindPaymentExpirationCandidates(cutoff, limit)
	if err != nil {
		return result, err
	}
	result.ScannedCandidates = len(candidates)

	for _, candidate := range candidates {
		expiredTransactions, expired, err := s.expirePaymentOrderIfStillEligible(candidate.ID, cutoff, now)
		if err != nil {
			return result, err
		}
		if expired {
			result.ExpiredOrders++
			result.ExpiredOpenTransactions += expiredTransactions
		} else {
			result.SkippedOrders++
		}
	}

	return result, nil
}

func (s *OrderService) expirePaymentOrderIfStillEligible(orderID uint, cutoff, now time.Time) (int64, bool, error) {
	var expiredTransactions int64
	expired := false
	var affectedProductIDs []uint

	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		orderRecord, err := repos.Order.FindByIDForUpdateWithItems(orderID)
		if err != nil {
			return err
		}
		transactions, err := repos.Payment.FindTransactionByOrderID(orderRecord.ID)
		if err != nil {
			return err
		}
		if !paymentExpirationStillEligible(orderRecord, transactions, cutoff) {
			return nil
		}

		if err := repos.Order.MarkPaymentExpired(orderRecord.ID, now); err != nil {
			return err
		}
		expiredTransactions, err = repos.Payment.ExpireOpenTransactionsByOrderID(orderRecord.ID, now)
		if err != nil {
			return err
		}
		productIDs, err := rollbackOrderReservationsInTx(repos, orderRecord, "payment expired")
		if err != nil {
			return err
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
		if err := s.enqueueProductCacheInvalidationInTx(repos, productIDs, "order stock restored payment expired"); err != nil {
			return err
		}
		expired = true
		return nil
	})
	if err == nil && expired {
		s.invalidateProductCacheAfterStockCommit(affectedProductIDs)
	}

	return expiredTransactions, expired, err
}

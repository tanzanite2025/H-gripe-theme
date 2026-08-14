package service

import "commerce-platform/internal/repository"

func (s *OrderService) invalidateProductCacheAfterStockCommit(productIDs []uint) {
	if s == nil || s.productCache == nil {
		return
	}
	s.productCache.InvalidateProductCacheByIDs(uniqueUintIDs(productIDs))
}

func (s *OrderService) enqueueProductCacheInvalidationInTx(repos repository.TxRepositories, productIDs []uint, reason string) error {
	ids := uniqueUintIDs(productIDs)
	if len(ids) == 0 {
		return nil
	}
	if repos.Outbox != nil {
		return NewProductCacheOutboxPublisher(repos.Outbox).EnqueueProductCacheInvalidateByIDs(ids, reason)
	}
	if s != nil && s.productCacheEvents != nil {
		return s.productCacheEvents.EnqueueProductCacheInvalidateByIDs(ids, reason)
	}
	return nil
}

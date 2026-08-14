package service

import (
	"errors"

	"commerce-platform/internal/repository"
)

var ErrTransactionalOutboxNotConfigured = errors.New("transactional outbox is not configured")

func newTransactionalProductPublishers(
	outboxRepo *repository.OutboxRepository,
) (ProductCacheEventPublisher, MerchantProductEventPublisher, error) {
	if outboxRepo == nil {
		return nil, nil, ErrTransactionalOutboxNotConfigured
	}
	return NewProductCacheOutboxPublisher(outboxRepo), NewMerchantOutboxPublisher(outboxRepo), nil
}

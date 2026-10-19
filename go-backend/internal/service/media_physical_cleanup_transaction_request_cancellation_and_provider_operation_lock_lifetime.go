package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"commerce-platform/internal/repository"
)

func (s *MediaService) withinMediaPhysicalCleanupTransactionHoldingLocksUntilProviderReturns(ctx context.Context, operation func(*repository.MediaRepository, *gorm.DB) error) error {
	if err := s.repo.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	// database/sql must not release row/object locks on request cancellation
	// while a provider operation can still delete bytes. Every individual SQL
	// statement and provider call keeps the request deadline. The transaction
	// lifetime alone waits until the synchronous operation has returned.
	return s.repo.WithinMediaAssetTransactionWithRequestContext(context.WithoutCancel(ctx), func(repo *repository.MediaRepository, tx *gorm.DB) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		requestTx := tx.WithContext(ctx)
		operationError := operation(repo.WithTx(requestTx), requestTx)
		return errors.Join(operationError, ctx.Err())
	})
}

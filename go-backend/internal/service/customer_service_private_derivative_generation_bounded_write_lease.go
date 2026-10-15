package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

var ErrCustomerServiceAttachmentWritesInProgress = errors.New("customer-service attachment writes are still in progress")

// A short SQL transaction claims the write lease; object IO occurs after commit.
// Every generation path shares this boundary, including background backfill.
func (s *MediaService) acquireCustomerServicePrivateDerivativeWriteLease(ctx context.Context, asset *media.MediaAsset) (context.Context, func(), error) {
	if asset.Visibility != "private" || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(assetStorageKey(asset)) {
		return ctx, func() {}, nil
	}
	token := uuid.NewString()
	now := time.Now().UTC()
	deadline := now.Add(media.CustomerServiceAttachmentUploadMaximumWriteDuration)
	if requestDeadline, ok := ctx.Deadline(); ok && requestDeadline.Before(deadline) {
		deadline = requestDeadline
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	err := s.repo.WithinMediaAssetTransactionWithRequestContext(ctx, func(repo *repository.MediaRepository, tx *gorm.DB) error {
		current, err := repo.FindAssetByIDForUpdate(asset.ID)
		if err != nil {
			return err
		}
		if current.Status != "active" || current.Visibility != "private" || assetStorageKey(current) != assetStorageKey(asset) {
			return ErrMediaAssetForbidden
		}
		if current.CustomerServiceDerivativeWriteExpiresAt != nil && current.CustomerServiceDerivativeWriteExpiresAt.After(now) {
			return ErrCustomerServiceAttachmentWritesInProgress
		}
		return tx.Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Updates(map[string]interface{}{
			"customer_service_derivative_write_token": token, "customer_service_derivative_write_expires_at": deadline,
		}).Error
	})
	if err != nil {
		return nil, nil, err
	}
	writeContext, cancel := context.WithDeadline(ctx, deadline)
	return writeContext, func() {
		cancel()
		releaseContext, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer releaseCancel()
		_ = s.repo.ReleaseCustomerServiceDerivativeWriteLease(releaseContext, asset.ID, token)
	}, nil
}

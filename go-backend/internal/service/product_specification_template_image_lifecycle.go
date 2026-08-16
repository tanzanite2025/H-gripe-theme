package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/pkg/logger"

	"go.uber.org/zap"
)

type mediaAssetDeleter interface {
	DeleteAsset(ctx context.Context, id uint, confirmation string) error
}

// CleanupDetachedProductSpecificationTemplateImageAsset lets an HTTP boundary hand back an
// uploaded asset that never became part of a product specification template relation.
func (s *ProductService) CleanupDetachedProductSpecificationTemplateImageAsset(assetID uint, reason string) {
	s.cleanupProductSpecificationTemplateImageAsset(assetID, reason)
}

// cleanupProductSpecificationTemplateImageAsset is intentionally best-effort. The product specification template
// relation is already updated or deleted when this runs, while MediaService
// rechecks all references before removing the physical object and asset row.
func (s *ProductService) cleanupProductSpecificationTemplateImageAsset(assetID uint, reason string) {
	if s == nil || s.mediaService == nil || assetID == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.mediaService.DeleteAsset(ctx, assetID, MediaAssetDeleteConfirmation(assetID))
	if err == nil {
		return
	}
	if errors.Is(err, ErrMediaAssetInUse) {
		logger.Info(
			"retaining product specification template image asset because it is still referenced",
			zap.Uint("asset_id", assetID),
			zap.String("reason", reason),
		)
		return
	}
	if errors.Is(err, ErrMediaAssetNotFound) {
		logger.Info(
			"product specification template image asset was already removed",
			zap.Uint("asset_id", assetID),
			zap.String("reason", reason),
		)
		return
	}

	logger.Warn(
		"failed to clean up product specification template image asset",
		zap.Uint("asset_id", assetID),
		zap.String("reason", reason),
		zap.Error(err),
	)
}

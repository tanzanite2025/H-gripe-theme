package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"gorm.io/gorm"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

const mediaDerivativeBackfillBatchSize = 25

// MediaDerivativeBackfillResult describes one bounded, one-shot repair pass.
// It is deliberately not a recurring scheduler: once assets have their
// persistent derivatives, subsequent runs find nothing to do.
type MediaDerivativeBackfillResult struct {
	ScannedAssets           int
	GeneratedAssets         int
	GeneratedDerivatives    int
	UpdatedProductMediaRows int64
}

// BackfillMissingImageDerivatives upgrades legacy image assets that predate
// persistent derivatives. It is idempotent and processes one source image at
// a time, keeping memory bounded independently of the storefront process.
func (s *MediaService) BackfillMissingImageDerivatives(ctx context.Context) (MediaDerivativeBackfillResult, error) {
	var result MediaDerivativeBackfillResult
	if s == nil || s.repo == nil || s.storage == nil {
		return result, ErrMediaStorageUnavailable
	}

	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return result, err
	}
	requiredDerivatives := mediaDerivativePresetRequirements(presets)
	var afterID uint
	for {
		assets, err := s.repo.ListImageAssetsMissingDerivatives(afterID, mediaDerivativeBackfillBatchSize, requiredDerivatives)
		if err != nil {
			return result, fmt.Errorf("list image assets missing derivatives: %w", err)
		}
		if len(assets) == 0 {
			return result, nil
		}

		for index := range assets {
			asset := &assets[index]
			afterID = asset.ID
			result.ScannedAssets++

			assetResult, err := s.ensureAssetDerivatives(ctx, asset)
			if err != nil {
				return result, fmt.Errorf("backfill media asset %d: %w", asset.ID, err)
			}
			if assetResult.GeneratedDerivatives > 0 {
				result.GeneratedAssets++
				result.GeneratedDerivatives += assetResult.GeneratedDerivatives
			}
			result.UpdatedProductMediaRows += assetResult.UpdatedProductMediaRows
		}
	}
}

type mediaDerivativeEnsureResult struct {
	GeneratedDerivatives    int
	UpdatedProductMediaRows int64
}

func (s *MediaService) ensureAssetDerivatives(
	ctx context.Context,
	asset *mediadomain.MediaAsset,
) (mediaDerivativeEnsureResult, error) {
	var result mediaDerivativeEnsureResult
	if s == nil || s.repo == nil || s.storage == nil || asset == nil {
		return result, ErrMediaStorageUnavailable
	}
	if asset.MediaType != "image" {
		return result, nil
	}

	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return result, err
	}
	missingPresets := missingMediaDerivativePresets(asset.Derivatives, presets)
	var retiredDerivatives []mediadomain.MediaAssetDerivative
	if len(missingPresets) > 0 {
		if s.objectCleanupOutbox == nil {
			return result, ErrObjectStorageCleanupUnavailable
		}
		openSource, err := s.openMediaDerivativeSource(ctx, asset)
		if err != nil {
			return result, err
		}

		persistGenerated := func(repo *repository.MediaRepository, _ *gorm.DB, generated []mediadomain.MediaAssetDerivative) error {
			return repo.ReplaceAssetDerivativesAndRecordRetiredObjectsWithinTransaction(generated, func(tx *gorm.DB, retired []mediadomain.MediaAssetDerivative) error {
				if err := s.recordRetiredMediaDerivativeCleanupEventsWithinReplacementTransaction(tx, retired); err != nil {
					return err
				}
				// Snapshot replacement shares the metadata/event commit. Otherwise a
				// failed snapshot update makes the next missing-generation backfill
				// skip this asset and leaves its consumers permanently on old URLs.
				transactionService := *s
				transactionService.repo = s.repo.WithTx(tx)
				variants, thumbnail, err := transactionService.productMediaImageVariantsUsingPresetDefinitions(asset.ID, presets)
				if err != nil {
					return err
				}
				rows, err := transactionService.repo.UpdateProductMediaImageVariantsForAsset(asset.ID, variants, thumbnail)
				if err != nil {
					return err
				}
				result.UpdatedProductMediaRows = rows
				retiredDerivatives = retired
				return nil
			})
		}
		var generated []mediadomain.MediaAssetDerivative
		if asset.Visibility != "private" {
			generated, err = s.generateAndPersistPublicMediaDerivativesWithPrecommittedExactKeyReceipts(ctx, asset, openSource, missingPresets, persistGenerated)
			if err != nil {
				return result, err
			}
		} else {
			generated, err = s.generateAssetDerivativesFromSource(ctx, asset, openSource, missingPresets)
			if err != nil {
				return result, err
			}
			err = persistGenerated(s.repo, nil, generated)
			if err != nil {
				uploadedURLs := make([]string, 0, len(generated))
				for _, derivative := range generated {
					uploadedURLs = append(uploadedURLs, derivative.URL)
				}
				return result, errors.Join(fmt.Errorf("persist generated derivatives: %w", err), s.persistFailedMediaUploadCompensationBeforeAttemptingExactObjectCleanup(ctx, asset, uploadedURLs))
			}
		}
		asset.Derivatives = replaceMediaAssetDerivatives(asset.Derivatives, generated)
		result.GeneratedDerivatives = len(generated)
	}

	if len(missingPresets) == 0 {
		variants, thumbnail, err := s.ProductMediaImageVariants(asset.ID)
		if err != nil {
			return result, fmt.Errorf("resolve generated image variants: %w", err)
		}
		rows, err := s.repo.UpdateProductMediaImageVariantsForAsset(asset.ID, variants, thumbnail)
		if err != nil {
			return result, fmt.Errorf("update product media image variants: %w", err)
		}
		result.UpdatedProductMediaRows = rows
	}
	for _, retired := range retiredDerivatives {
		key, err := s.exactRegisteredDerivativeStorageKeyForCleanup(retired)
		if err == nil {
			// Synchronize storefront snapshots first. The committed event remains
			// the retry path if any historical snapshot or other consumer remains.
			_ = s.cleanupUnregisteredOrRetiredMediaObjectFromDurableEvent(ctx, objectCleanupResourceRetiredMediaDerivative,
				strconv.FormatUint(uint64(retired.ID), 10), []string{key})
		}
	}
	return result, nil
}

func (s *MediaService) openMediaDerivativeSource(
	ctx context.Context,
	asset *mediadomain.MediaAsset,
) (mediaDerivativeSourceOpener, error) {
	if s == nil || s.storage == nil || asset == nil {
		return nil, ErrMediaStorageUnavailable
	}

	key := storageObjectKey(s.storage, asset.URL)
	if key == "" {
		key = assetStorageKey(asset)
	}
	if key == "" || IsMediaDerivativeStorageKey(key) {
		return nil, ErrMediaAssetURLUnavailable
	}

	opener, ok := s.storage.(storage.ObjectOpener)
	if !ok {
		return nil, ErrMediaStorageUnavailable
	}
	return func() (io.ReadCloser, error) {
		object, err := opener.Open(ctx, key)
		if err != nil {
			return nil, err
		}
		if object == nil || object.ReadCloser == nil {
			return nil, ErrMediaStorageUnavailable
		}
		return object.ReadCloser, nil
	}, nil
}

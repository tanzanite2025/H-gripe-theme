package service

import (
	"context"
	"fmt"
	"io"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
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
	if len(missingPresets) > 0 {
		openSource, err := s.openMediaDerivativeSource(ctx, asset)
		if err != nil {
			return result, err
		}

		generated, err := s.generateAssetDerivativesFromSource(ctx, asset, openSource, missingPresets)
		if err != nil {
			return result, err
		}
		replacedURLs, err := s.repo.ReplaceAssetDerivatives(generated)
		if err != nil {
			uploadedURLs := make([]string, 0, len(generated))
			for _, derivative := range generated {
				uploadedURLs = append(uploadedURLs, derivative.URL)
			}
			deleteUploadedMediaObjectsBestEffort(ctx, s.storage, uploadedURLs)
			return result, fmt.Errorf("persist generated derivatives: %w", err)
		}
		deleteUploadedMediaObjectsBestEffort(ctx, s.storage, replacedURLs)
		asset.Derivatives = replaceMediaAssetDerivatives(asset.Derivatives, generated)
		result.GeneratedDerivatives = len(generated)
	}

	variants, thumbnail, err := s.ProductMediaImageVariants(asset.ID)
	if err != nil {
		return result, fmt.Errorf("resolve generated image variants: %w", err)
	}
	rows, err := s.repo.UpdateProductMediaImageVariantsForAsset(asset.ID, variants, thumbnail)
	if err != nil {
		return result, fmt.Errorf("update product media image variants: %w", err)
	}
	result.UpdatedProductMediaRows = rows
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

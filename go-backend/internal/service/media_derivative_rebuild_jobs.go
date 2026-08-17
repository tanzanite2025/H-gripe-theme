package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
)

const (
	mediaDerivativeRebuildBatchSize = 10
	mediaDerivativeRebuildLeaseTTL  = 15 * time.Minute
)

type MediaDerivativeRebuildProcessResult struct {
	Claimed                 bool
	Completed               bool
	JobID                   uint
	ScannedAssets           int
	GeneratedAssets         int
	GeneratedDerivatives    int
	FailedAssets            int
	UpdatedProductMediaRows int64
}

func (s *MediaService) ListMediaDerivativeRebuildJobs() ([]mediadomain.MediaDerivativeRebuildJob, error) {
	if s == nil || s.derivativeRebuildJobs == nil {
		return nil, ErrMediaDerivativePresetUnavailable
	}
	return s.derivativeRebuildJobs.ListRecent(10)
}

func (s *MediaService) RequestMediaDerivativeRebuild(reason string) (*mediadomain.MediaDerivativeRebuildJob, error) {
	if s == nil || s.derivativeRebuildJobs == nil {
		return nil, ErrMediaDerivativePresetUnavailable
	}
	job, _, err := s.derivativeRebuildJobs.Enqueue(reason, time.Now().UTC())
	return job, err
}

// ProcessNextMediaDerivativeRebuild claims one persistent rebuild and advances
// at most one bounded batch. Configuration changes made during a running pass
// create a later pending pass, which starts from asset zero with the newest
// active registry.
func (s *MediaService) ProcessNextMediaDerivativeRebuild(
	ctx context.Context,
	workerID string,
	batchSize int,
) (MediaDerivativeRebuildProcessResult, error) {
	var output MediaDerivativeRebuildProcessResult
	if s == nil || s.repo == nil || s.storage == nil || s.derivativeRebuildJobs == nil {
		return output, ErrMediaStorageUnavailable
	}
	if batchSize <= 0 || batchSize > 100 {
		batchSize = mediaDerivativeRebuildBatchSize
	}

	now := time.Now().UTC()
	job, err := s.derivativeRebuildJobs.ClaimNext(now, workerID, mediaDerivativeRebuildLeaseTTL)
	if err != nil {
		return output, err
	}
	if job == nil {
		return output, nil
	}
	output.Claimed = true
	output.JobID = job.ID

	batch, completed, processErr := s.processMediaDerivativeRebuildBatch(ctx, job.CursorAssetID, batchSize)
	output.Completed = completed
	output.ScannedAssets = batch.ScannedAssets
	output.GeneratedAssets = batch.GeneratedAssets
	output.GeneratedDerivatives = batch.GeneratedDerivatives
	output.FailedAssets = batch.FailedAssets
	output.UpdatedProductMediaRows = batch.UpdatedProductMediaRows

	if processErr != nil {
		batch.LastError = truncateMediaDerivativeRebuildError(processErr.Error())
		completed = false
	}
	if err := s.derivativeRebuildJobs.CompleteBatch(
		job,
		workerID,
		batch,
		completed,
		time.Now().UTC(),
	); err != nil {
		return output, err
	}
	if processErr != nil {
		return output, processErr
	}
	return output, nil
}

func (s *MediaService) processMediaDerivativeRebuildBatch(
	ctx context.Context,
	afterID uint,
	batchSize int,
) (repository.MediaDerivativeRebuildBatchUpdate, bool, error) {
	var result repository.MediaDerivativeRebuildBatchUpdate
	assets, err := s.repo.ListImageAssetsForDerivativeRebuild(afterID, batchSize)
	if err != nil {
		return result, false, fmt.Errorf("list image assets for derivative rebuild: %w", err)
	}
	if len(assets) == 0 {
		return result, true, nil
	}

	errorsSeen := make([]string, 0, 3)
	for index := range assets {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return result, false, err
			}
		}
		asset := &assets[index]
		result.CursorAssetID = asset.ID
		result.ScannedAssets++

		assetResult, err := s.rebuildImageAssetDerivatives(ctx, asset)
		if err != nil {
			result.FailedAssets++
			if len(errorsSeen) < cap(errorsSeen) {
				errorsSeen = append(errorsSeen, fmt.Sprintf("asset %d: %v", asset.ID, err))
			}
			continue
		}
		if assetResult.GeneratedDerivatives > 0 {
			result.GeneratedAssets++
			result.GeneratedDerivatives += assetResult.GeneratedDerivatives
		}
		result.UpdatedProductMediaRows += assetResult.UpdatedProductMediaRows
	}
	if len(errorsSeen) > 0 {
		result.LastError = truncateMediaDerivativeRebuildError(strings.Join(errorsSeen, "; "))
	}
	return result, len(assets) < batchSize, nil
}

func (s *MediaService) rebuildImageAssetDerivatives(
	ctx context.Context,
	asset *mediadomain.MediaAsset,
) (mediaDerivativeEnsureResult, error) {
	if asset == nil {
		return mediaDerivativeEnsureResult{}, errors.New("media asset is required")
	}
	if asset.Width <= 0 || asset.Height <= 0 {
		if err := s.syncImageAssetDimensions(ctx, asset); err != nil {
			return mediaDerivativeEnsureResult{}, err
		}
	}
	return s.ensureAssetDerivatives(ctx, asset)
}

func truncateMediaDerivativeRebuildError(value string) string {
	value = strings.TrimSpace(value)
	const limit = 2000
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	showcasedomain "commerce-platform/internal/domain/showcase"

	"gorm.io/datatypes"
)

const showcaseExpiredPendingReason = "pending submission expired"

type ShowcaseImageCleanupResult struct {
	Cutoff                 time.Time `json:"cutoff"`
	ScannedCandidates      int       `json:"scanned_candidates"`
	ExpiredPendingRecords  int       `json:"expired_pending_records"`
	DeletedPendingImages   int       `json:"deleted_pending_images"`
	RetainedFailedImages   int       `json:"retained_failed_images"`
	UpdatedImageReferences int       `json:"updated_image_references"`
}

func (s *ShowcaseService) CleanupExpiredPendingImages(
	ctx context.Context,
	now time.Time,
	retention time.Duration,
	limit int,
) (ShowcaseImageCleanupResult, error) {
	result := ShowcaseImageCleanupResult{}
	if s == nil || s.repo == nil || s.storage == nil {
		return result, ErrShowcaseStorageUnavailable
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if retention <= 0 {
		return result, fmt.Errorf("showcase pending image retention must be positive")
	}
	if limit <= 0 {
		return result, fmt.Errorf("showcase image cleanup batch limit must be positive")
	}

	result.Cutoff = now.Add(-retention)
	candidates, err := s.repo.ListImageCleanupCandidates(result.Cutoff, limit)
	if err != nil {
		return result, err
	}
	result.ScannedCandidates = len(candidates)

	for _, item := range candidates {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		cleanupStatus := item.Status
		if item.Status == showcasedomain.StatusPending {
			updated, err := s.repo.UpdatePendingStatus(item.ID, showcasedomain.StatusRejected, showcaseExpiredPendingReason)
			if err != nil {
				return result, err
			}
			if !updated {
				continue
			}
			cleanupStatus = showcasedomain.StatusRejected
			result.ExpiredPendingRecords++
		}

		remainingReferences, deleted, retained, err := s.deletePendingImageReferences(ctx, item.Images)
		if err != nil {
			return result, err
		}
		result.DeletedPendingImages += deleted
		result.RetainedFailedImages += retained
		if deleted == 0 {
			continue
		}

		imagesJSON, err := json.Marshal(remainingReferences)
		if err != nil {
			return result, fmt.Errorf("encode showcase cleanup images: %w", err)
		}
		updated, err := s.repo.UpdateImagesByStatus(item.ID, cleanupStatus, datatypes.JSON(imagesJSON))
		if err != nil {
			return result, err
		}
		if updated {
			result.UpdatedImageReferences++
		}
	}

	return result, nil
}

func (s *ShowcaseService) deletePendingImageReferences(
	ctx context.Context,
	rawImages datatypes.JSON,
) ([]string, int, int, error) {
	imageReferences, err := decodeShowcaseImageURLs(rawImages)
	if err != nil {
		return nil, 0, 0, err
	}

	remainingReferences := make([]string, 0, len(imageReferences))
	deleted := 0
	retained := 0
	for _, imageReference := range imageReferences {
		key, err := s.storage.ObjectKey(imageReference)
		if err != nil || !showcaseStorageKeyIsPending(key) {
			remainingReferences = append(remainingReferences, imageReference)
			continue
		}
		if err := s.storage.Delete(ctx, imageReference); err != nil {
			remainingReferences = append(remainingReferences, imageReference)
			retained++
			continue
		}
		deleted++
	}
	return remainingReferences, deleted, retained, nil
}

package service

import (
	"context"
	"fmt"
	"time"

	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const showcaseExpiredPendingReason = "pending submission expired"

type UGCShowcaseImageCleanupResult struct {
	Cutoff                 time.Time `json:"cutoff"`
	ScannedCandidates      int       `json:"scanned_candidates"`
	ExpiredPendingRecords  int       `json:"expired_pending_records"`
	DeletedPendingImages   int       `json:"deleted_pending_images"`
	RetainedFailedImages   int       `json:"retained_failed_images"`
	UpdatedImageReferences int       `json:"updated_image_references"`
}

func (s *UGCShowcaseService) CleanupExpiredPendingImages(ctx context.Context, now time.Time, retention time.Duration, limit int) (UGCShowcaseImageCleanupResult, error) {
	result := UGCShowcaseImageCleanupResult{}
	if s == nil || s.repo == nil || s.storage == nil {
		return result, ErrShowcaseStorageUnavailable
	}
	if retention <= 0 || limit <= 0 {
		return result, fmt.Errorf("showcase cleanup retention and batch limit must be positive")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now()
	}
	result.Cutoff = now.UTC().Add(-retention)
	candidates, err := s.repo.WithContext(ctx).ListImageCleanupCandidates(result.Cutoff, limit)
	if err != nil {
		return result, err
	}
	result.ScannedCandidates = len(candidates)
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		var keys []string
		changed := false
		err := s.repo.WithContext(ctx).WithTransaction(func(repo *repository.UGCShowcaseRepository, tx *gorm.DB) error {
			current, err := repo.FindUGCShowcaseByIDForUpdateWithinLifecycleTransaction(candidate.ID)
			if repository.IsRecordNotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if current.Status == ugcshowcasedomain.StatusPending {
				if !current.CreatedAt.Before(result.Cutoff) {
					return nil
				}
				changed, err = repo.UpdatePendingStatus(current.ID, ugcshowcasedomain.StatusRejected, showcaseExpiredPendingReason)
				if err != nil || !changed {
					return err
				}
			} else if current.Status != ugcshowcasedomain.StatusRejected {
				return nil
			}
			references, err := decodeShowcaseImageURLs(current.Images)
			if err != nil {
				return err
			}
			keys = s.showcasePendingCleanupKeys(references)
			return s.enqueueShowcaseCleanup(tx, current.ID, keys)
		})
		if err != nil {
			return result, err
		}
		if changed {
			result.ExpiredPendingRecords++
		}
		pruned := false
		for _, key := range keys {
			updated, err := s.cleanupUGCShowcaseObjectsWithRetainedReferencesAndLockedRejectedOwner(ctx, candidate.ID, []string{key}, false)
			if err != nil {
				result.RetainedFailedImages++
				continue
			}
			result.DeletedPendingImages++
			pruned = pruned || updated
		}
		if pruned {
			result.UpdatedImageReferences++
		}
	}
	return result, ctx.Err()
}

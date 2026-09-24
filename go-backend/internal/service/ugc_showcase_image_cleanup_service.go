package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
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

func (s *UGCShowcaseService) CleanupExpiredPendingImages(
	ctx context.Context,
	now time.Time,
	retention time.Duration,
	limit int,
) (UGCShowcaseImageCleanupResult, error) {
	result := UGCShowcaseImageCleanupResult{}
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

	if ctx == nil {
		ctx = context.Background()
	}

	for _, item := range candidates {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		// The database state transition and cleanup event must commit together.
		// The previous implementation deleted objects first and then updated the
		// record, which could permanently lose an object when the second write
		// failed. We now enqueue durable cleanup before any best-effort immediate
		// deletion; the event remains responsible for retries and stale references.
		var (
			pendingReferences   []string
			remainingReferences []string
			cleanupKeys         []string
			cleanupStatus       string
			statusChanged       bool
		)
		err := s.repo.WithTransaction(func(repo *repository.UGCShowcaseRepository, tx *gorm.DB) error {
			current, err := repo.GetByID(item.ID)
			if err != nil {
				if repository.IsRecordNotFound(err) {
					return nil
				}
				return err
			}

			cleanupStatus = current.Status
			if cleanupStatus == ugcshowcasedomain.StatusPending {
				updated, updateErr := repo.UpdatePendingStatus(current.ID, ugcshowcasedomain.StatusRejected, showcaseExpiredPendingReason)
				if updateErr != nil {
					return updateErr
				}
				if !updated {
					return nil
				}
				cleanupStatus = ugcshowcasedomain.StatusRejected
				statusChanged = true
			} else if cleanupStatus != ugcshowcasedomain.StatusRejected {
				return nil
			}

			imageReferences, decodeErr := decodeShowcaseImageURLs(current.Images)
			if decodeErr != nil {
				return decodeErr
			}
			pendingReferences, remainingReferences = splitPendingShowcaseImageReferences(s, imageReferences)
			cleanupKeys = s.showcasePendingCleanupKeys(pendingReferences)
			if len(cleanupKeys) == 0 {
				return nil
			}

			return s.enqueueShowcaseCleanup(tx, current.ID, cleanupKeys)
		})
		if err != nil {
			return result, err
		}

		if statusChanged {
			result.ExpiredPendingRecords++
		}
		if len(cleanupKeys) == 0 {
			continue
		}

		deletedReferences, retainedReferences := s.deletePendingImageReferences(ctx, pendingReferences)
		result.DeletedPendingImages += len(deletedReferences)
		result.RetainedFailedImages += len(retainedReferences)
		if len(deletedReferences) == 0 {
			continue
		}

		// Prune references only after a successful best-effort delete. If the
		// immediate delete or this non-critical update fails, the durable event
		// remains available and the handler will retry both operations.
		remaining := append([]string(nil), remainingReferences...)
		remaining = append(remaining, retainedReferences...)
		imagesJSON, marshalErr := json.Marshal(remaining)
		if marshalErr != nil {
			return result, fmt.Errorf("encode showcase cleanup images: %w", marshalErr)
		}
		updated, updateErr := s.repo.UpdateImagesByStatus(item.ID, cleanupStatus, datatypes.JSON(imagesJSON))
		if updateErr != nil {
			return result, updateErr
		}
		if updated {
			result.UpdatedImageReferences++
		}
	}

	return result, nil
}

func splitPendingShowcaseImageReferences(
	s *UGCShowcaseService,
	imageReferences []string,
) (pendingReferences, remainingReferences []string) {
	remainingReferences = make([]string, 0, len(imageReferences))
	for _, imageReference := range imageReferences {
		key, err := s.storage.ObjectKey(imageReference)
		if err != nil || !showcaseStorageKeyIsPending(key) {
			remainingReferences = append(remainingReferences, imageReference)
			continue
		}
		pendingReferences = append(pendingReferences, imageReference)
	}
	return pendingReferences, remainingReferences
}

func (s *UGCShowcaseService) deletePendingImageReferences(
	ctx context.Context,
	pendingReferences []string,
) (deletedReferences, retainedReferences []string) {
	for _, imageReference := range pendingReferences {
		if err := s.storage.Delete(ctx, imageReference); err != nil {
			retainedReferences = append(retainedReferences, imageReference)
			continue
		}
		deletedReferences = append(deletedReferences, imageReference)
	}
	return deletedReferences, retainedReferences
}

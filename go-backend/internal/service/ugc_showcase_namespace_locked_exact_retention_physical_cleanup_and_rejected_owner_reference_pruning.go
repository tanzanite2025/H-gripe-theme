package service

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

// A rejected owner's pending reference is the only excluded claim. Every other
// UGC/business/registration claim retains the original task for a later retry.
func (s *UGCShowcaseService) cleanupUGCShowcaseObjectsWithRetainedReferencesAndLockedRejectedOwner(ctx context.Context, ownerID uint, keys []string, failedUpload bool) (bool, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return false, ErrObjectStorageCleanupUnavailable
	}
	pruned := false
	err := s.repo.WithLockedUGCShowcaseLifecycleUntilProviderOperationReturns(ctx, func(repo *repository.UGCShowcaseRepository, tx *gorm.DB) error {
		var owner *ugcshowcase.UGCShowcase
		if ownerID != 0 {
			var err error
			owner, err = repo.FindUGCShowcaseByIDForUpdateWithinLifecycleTransaction(ownerID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		referencesRepo := repository.NewMediaRepository(tx)
		referencesRepo.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: s.storage, siteURL: s.siteURL})
		lockedService := *s
		lockedService.repo = repo
		for _, key := range keys {
			if !showcaseStorageKeyIsPending(key) && !(failedUpload && showcaseStorageKeyIsApproved(key)) {
				continue
			}
			if key == showcasePendingUploadPrefix || key == showcaseApprovedUploadPrefix {
				continue
			}
			references, err := referencesRepo.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
			if err != nil {
				return fmt.Errorf("check retained UGC image references: %w", err)
			}
			retained := make([]media.AssetReference, 0, len(references))
			for _, reference := range references {
				if owner != nil && owner.Status == ugcshowcase.StatusRejected && showcaseStorageKeyIsPending(key) && reference.ResourceType == "showcase" && reference.ResourceID == ownerID && reference.Field == "images" {
					continue
				}
				retained = append(retained, reference)
			}
			if len(retained) > 0 {
				return &MediaAssetInUseError{References: retained}
			}
			if err := s.storage.Delete(ctx, key); err != nil {
				return fmt.Errorf("delete unreferenced UGC image: %w", err)
			}
			if owner != nil && owner.Status == ugcshowcase.StatusRejected {
				updated, err := lockedService.pruneDeletedPendingUGCShowcaseImageFromLockedRejectedOwner(owner, key)
				if err != nil {
					return err
				}
				pruned = pruned || updated
			}
		}
		return nil
	})
	return pruned && err == nil, err
}

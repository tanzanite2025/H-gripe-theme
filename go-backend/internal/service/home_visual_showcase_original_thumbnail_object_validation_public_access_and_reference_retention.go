package service

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

// Called by the repository under the same namespace lock as physical cleanup.
// Existing placements may retain historical URLs; new placements need an exact
// provider-owned key and a readable object, including independent thumbnails.
func (s *HomeVisualTileService) validateVisualShowcaseImageBindingsWithObjectExistenceUnderTransactionLock(tx *gorm.DB, items, previous []homevisualtile.Tile) error {
	previousBySlot := make(map[int]homevisualtile.Tile, len(previous))
	for _, item := range previous {
		previousBySlot[item.DesktopOrder] = item
	}
	checked := make(map[string]struct{})
	for _, item := range items {
		old, exists := previousBySlot[item.DesktopOrder]
		retainsSlot := exists && old.TileSetKey == item.TileSetKey && old.Locale == item.Locale
		for index, reference := range []string{item.ImageURL, item.ThumbnailURL} {
			oldReference := old.ImageURL
			if index == 1 {
				oldReference = old.ThumbnailURL
				if oldReference == "" {
					oldReference = old.ImageURL
				}
			}
			if retainsSlot && reference == oldReference && (index == 1 || item.StorageKey == old.StorageKey) {
				continue
			}
			key := s.visualShowcaseStorageKeyFromReference(reference)
			if key == "" || (index == 0 && key != item.StorageKey) {
				return fmt.Errorf("%w: original and thumbnail must identify dedicated showcase uploads", ErrHomeVisualTileImageInvalid)
			}
			if _, exists := checked[key]; exists {
				continue
			}
			opener, ok := s.storage.(storage.ObjectOpener)
			if !ok {
				return ErrHomeVisualTileStorageUnavailable
			}
			object, err := opener.Open(tx.Statement.Context, key)
			if err != nil {
				return fmt.Errorf("%w: selected file is missing or unreadable: %v", ErrHomeVisualTileImageInvalid, err)
			}
			if object == nil || object.ReadCloser == nil {
				return fmt.Errorf("%w: selected file is unreadable", ErrHomeVisualTileImageInvalid)
			}
			_ = object.ReadCloser.Close()
			checked[key] = struct{}{}
		}
	}
	keys := make([]string, 0, len(items)*2)
	for _, item := range items {
		keys = append(keys, s.dedicatedVisualShowcaseObjectKeysReferencedByItem(item)...)
	}
	err := s.repo.WithTx(tx).BindVisualShowcaseDraftUploadReceiptsWithinOwnerTransaction(keys, s.visualShowcaseStorageKeyFromReference)
	if errors.Is(err, repository.ErrVisualShowcaseDraftUploadReceiptUnavailable) {
		return fmt.Errorf("%w: upload expired or is not ready; upload the image again: %w", ErrHomeVisualTileImageInvalid, err)
	}
	return err
}

func (s *HomeVisualTileService) dedicatedVisualShowcaseObjectKeysReferencedByItem(item homevisualtile.Tile) []string {
	keys := []string{s.visualShowcaseStorageKeyFromInput(item.StorageKey, item.ImageURL), s.visualShowcaseStorageKeyFromReference(item.ThumbnailURL)}
	// An old stored key is retention evidence even when its URL alias no longer
	// belongs to the current provider. It does not authorize a new binding.
	if key, ok := storage.NormalizeObjectKey(item.StorageKey); ok && IsHomeVisualTileStorageKey(key) {
		keys = append(keys, key)
	}
	return normalizeObjectCleanupKeys(keys)
}

func (s *HomeVisualTileService) visualShowcaseObjectHasActiveReference(repo *repository.HomeVisualTileRepository, key string, publicOnly bool) (bool, error) {
	found := false
	err := repo.VisitActiveVisualShowcaseImageReferences(publicOnly, func(item homevisualtile.Tile) bool {
		for _, candidate := range s.dedicatedVisualShowcaseObjectKeysReferencedByItem(item) {
			if candidate == key {
				found = true
				return false
			}
		}
		return true
	})
	return found, err
}

func (s *HomeVisualTileService) CanServePublicVisualShowcaseImage(ctx context.Context, key string) (bool, error) {
	normalized, ok := storage.NormalizeObjectKey(key)
	if !ok || !IsHomeVisualTileStorageKey(normalized) || s == nil || s.repo == nil {
		return false, nil
	}
	allowed, err := s.visualShowcaseObjectHasActiveReference(s.repo.WithContext(ctx), normalized, true)
	if err != nil || allowed {
		return allowed, err
	}
	return s.repo.HasPublishedBusinessConsumerForDedicatedPublicObject(ctx, normalized, &MediaService{storage: s.storage, siteURL: s.siteURL})
}

func (s *HomeVisualTileService) cleanupUnreferencedDedicatedVisualShowcaseObjectsUnderTransactionLock(ctx context.Context, keys []string) error {
	return s.repo.WithLockedVisualShowcaseReferencesForObjectCleanup(ctx, func(repo *repository.HomeVisualTileRepository, references *repository.MediaRepository) error {
		retained := make(map[string]struct{})
		if err := repo.VisitActiveVisualShowcaseImageReferences(false, func(item homevisualtile.Tile) bool {
			for _, key := range s.dedicatedVisualShowcaseObjectKeysReferencedByItem(item) {
				retained[key] = struct{}{}
			}
			return true
		}); err != nil {
			return fmt.Errorf("check dedicated showcase references: %w", err)
		}
		for _, key := range keys {
			if !IsHomeVisualTileStorageKey(key) {
				continue
			}
			if _, referenced := retained[key]; referenced {
				continue
			}
			if err := checkDedicatedPublicObjectBusinessReferencesBeforePhysicalCleanup(references, s.storage, s.siteURL, key); err != nil {
				return err
			}
			if err := s.storage.Delete(ctx, key); err != nil {
				return fmt.Errorf("delete retired home visual tile %s: %w", key, err)
			}
		}
		return nil
	})
}

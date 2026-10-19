package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrMediaAssetInUse                 = errors.New("media asset is currently in use")
	ErrMediaDeleteConfirmationRequired = errors.New("media delete confirmation is invalid")
)

type MediaAssetInUseError struct {
	References []media.AssetReference
}

func (e *MediaAssetInUseError) Error() string {
	return ErrMediaAssetInUse.Error()
}

func (e *MediaAssetInUseError) Unwrap() error {
	return ErrMediaAssetInUse
}

func MediaAssetDeleteConfirmation(id uint) string {
	return fmt.Sprintf("DELETE %d", id)
}

// DeleteAsset atomically hides an unreferenced asset and records durable object
// cleanup. Physical deletion happens after commit and is retried by Outbox.
func (s *MediaService) DeleteAsset(ctx context.Context, id uint, confirmation string) error {
	if !isMediaDeleteConfirmationValid(id, confirmation) {
		return ErrMediaDeleteConfirmationRequired
	}
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrMediaStorageUnavailable
	}
	if s.objectCleanupOutbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}

	err := s.repo.WithTransaction(func(repo *repository.MediaRepository, tx *gorm.DB) error {
		asset, err := repo.FindAssetByIDForUpdate(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMediaAssetNotFound
			}
			return err
		}
		return s.retireUnreferencedMediaAssetWithinTransaction(repo, tx, asset)
	})
	if err != nil {
		return err
	}
	// Durable cleanup is the correctness path; immediate cleanup only reduces
	// object-store retention latency and is safe to repeat by the worker.
	_ = s.cleanupDeletedAsset(ctx, id)

	return nil
}

func (s *MediaService) cleanupDeletedAsset(ctx context.Context, id uint, originalReceiptKeys ...[]string) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var cleanedAsset *media.MediaAsset
	err := s.withinMediaPhysicalCleanupTransactionHoldingLocksUntilProviderReturns(ctx, func(repo *repository.MediaRepository, _ *gorm.DB) error {
		asset, err := repo.FindAssetByIDUnscopedForUpdate(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if len(originalReceiptKeys) > 0 {
				return s.cleanupLateWrittenMediaObjectsWithinTransactionAfterRegistrationRemovalFromOriginalReceipt(ctx, repo, id, originalReceiptKeys[0])
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("load deleted media asset: %w", err)
		}
		if !asset.DeletedAt.Valid {
			return nil
		}
		if err := repo.LockRegisteredMediaAssetObjectsBeforePhysicalCleanupWithinTransaction(asset); err != nil {
			return err
		}
		if err := s.cleanupDeletedMediaAssetObjectsAndRegistrationWithinLockedTransaction(ctx, repo, asset); err != nil {
			return err
		}
		cleanedAsset = asset
		return nil
	})
	if err != nil {
		return err
	}
	if cleanedAsset != nil && s.cdnPurger != nil {
		s.cdnPurger.PurgeAsync(s.CanonicalPublicMediaURL(cleanedAsset.URL))
		for _, derivative := range cleanedAsset.Derivatives {
			s.cdnPurger.PurgeAsync(s.CanonicalPublicMediaURL(derivative.URL))
		}
	}
	return nil
}

// Keep the row and object locks through reference revalidation, provider IO and
// hard deletion. Concurrent workers and registrations use this same boundary.
func (s *MediaService) cleanupDeletedMediaAssetObjectsAndRegistrationWithinLockedTransaction(ctx context.Context, repo *repository.MediaRepository, asset *media.MediaAsset) error {
	if err := s.validateSharedMediaCleanupDoesNotOwnDedicatedPublicObjects(asset); err != nil {
		return err
	}
	if asset.CustomerServiceDerivativeWriteExpiresAt != nil && asset.CustomerServiceDerivativeWriteExpiresAt.After(time.Now().UTC()) {
		return ErrCustomerServiceAttachmentWritesInProgress
	}
	references, err := repo.FindAssetReferences(asset)
	if err != nil {
		return err
	}
	if len(references) > 0 {
		return &MediaAssetInUseError{References: references}
	}
	if err := s.cleanupRetiredCustomerServiceAttachmentDerivativeNamespace(ctx, asset, repo); err != nil {
		return err
	}

	for _, derivative := range asset.Derivatives {
		key := derivative.StorageKey
		if key == "" {
			key = storageObjectKey(s.storage, derivative.URL)
		}
		if key == "" {
			continue
		}
		if err := s.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete media derivative object: %w", err)
		}
	}
	key, keyErr := s.mediaAssetOriginalStorageKeyForDurableCleanup(asset)
	if keyErr != nil {
		return keyErr
	}
	if key != "" {
		if err := s.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete media object: %w", err)
		}
	}
	if err := repo.HardDeleteAsset(asset.ID); err != nil {
		return fmt.Errorf("delete media asset record: %w", err)
	}
	return nil
}

type mediaCDNPurger struct {
	url    string
	token  string
	client *http.Client
}

func newMediaCDNPurgerFromEnv() *mediaCDNPurger {
	url := strings.TrimSpace(os.Getenv("MEDIA_CDN_PURGE_URL"))
	token := strings.TrimSpace(os.Getenv("MEDIA_CDN_PURGE_TOKEN"))
	if url == "" || token == "" {
		return nil
	}
	return &mediaCDNPurger{url: url, token: token, client: &http.Client{Timeout: 3 * time.Second}}
}

func (p *mediaCDNPurger) PurgeAsync(reference string) {
	if p == nil || strings.TrimSpace(reference) == "" {
		return
	}
	client := p.client
	if client == nil {
		client = http.DefaultClient
	}
	go func() {
		payload, err := json.Marshal(map[string]interface{}{"files": []string{reference}})
		if err != nil {
			return
		}
		req, err := http.NewRequest(http.MethodPost, p.url, bytes.NewReader(payload))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.token)
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}()
}

func isMediaDeleteConfirmationValid(id uint, confirmation string) bool {
	return strings.TrimSpace(confirmation) == MediaAssetDeleteConfirmation(id)
}

// customerServiceAttachmentReferencedElsewhere protects shared media-library
// objects during conversation retention. A media asset may be attached to
// several messages or to an unrelated domain; only an object with no
// references outside the purged ticket can be physically removed by the
// retention service.
func (s *MediaService) customerServiceAttachmentReferencedElsewhere(reference string, ticketID uint) (bool, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return false, nil
	}
	key, err := s.storage.ObjectKey(reference)
	if err != nil || strings.TrimSpace(key) == "" {
		return false, nil
	}
	asset, err := s.repo.FindAssetByStorageKey(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	references, err := s.repo.FindAssetReferences(asset)
	if err != nil {
		return false, err
	}
	for _, item := range references {
		if item.ResourceType == "ticket_message" && item.ParentResourceID == ticketID {
			continue
		}
		return true, nil
	}
	return false, nil
}

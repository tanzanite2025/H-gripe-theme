package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type CustomerServicePendingAttachmentUploadExpiryBatchResult struct {
	Scanned int
	Retired int
	Blocked int
	Errors  int
}

func (s *CustomerServiceRetentionService) managedCustomerServiceAttachmentRetirementAvailable() bool {
	return s != nil && s.cleanupOutbox != nil && s.mediaService != nil && s.mediaService.repo != nil &&
		s.mediaService.storage != nil && s.mediaService.objectCleanupOutbox != nil
}

// Uses the existing scheduler even when conversation retention is disabled.
// Only expired, unbound receipts are candidates; bound media follows retention.
func (s *CustomerServiceRetentionService) ExpireUnboundCustomerServiceAttachmentUploads(ctx context.Context, limit int) (CustomerServicePendingAttachmentUploadExpiryBatchResult, error) {
	var result CustomerServicePendingAttachmentUploadExpiryBatchResult
	if !s.managedCustomerServiceAttachmentRetirementAvailable() {
		return result, nil
	}
	now := s.currentTime()
	sessions, err := s.mediaService.repo.FindExpiredUnboundCustomerServiceUploadSessions(ctx, now, limit)
	if err != nil {
		return result, err
	}
	result.Scanned = len(sessions)
	for _, candidate := range sessions {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		retired, err := s.retireExpiredUnboundCustomerServiceUploadSession(ctx, candidate, now)
		if err != nil {
			result.Errors++
		} else if retired {
			result.Retired++
		} else {
			result.Blocked++
		}
	}
	return result, nil
}

func (s *CustomerServiceRetentionService) retireExpiredUnboundCustomerServiceUploadSession(ctx context.Context, candidate media.MediaUploadSession, now time.Time) (bool, error) {
	retired := false
	err := s.mediaService.repo.WithinMediaAssetTransactionWithRequestContext(ctx, func(repo *repository.MediaRepository, tx *gorm.DB) error {
		var asset *media.MediaAsset
		var err error
		// Match message binding order: asset lock first, receipt lock second.
		if candidate.MediaAssetID != nil {
			asset, err = repo.FindCustomerServiceAttachmentAssetUnscopedForUpdate(*candidate.MediaAssetID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					_, lookupErr := repo.FindCustomerServiceMediaUploadSession(ctx, candidate.ID)
					if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
						return nil
					}
				}
				return err
			}
		}
		session, err := repo.FindCustomerServiceUploadSessionForUpdate(ctx, candidate.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if session.BusinessContextType != media.MediaUploadSessionCustomerServiceContext || session.ExpiresAt.After(now) ||
			session.CreatedAt.After(now.Add(-media.CustomerServiceAttachmentUploadMaximumWriteDuration)) ||
			(session.Status != "ready" && session.Status != "failed" && session.Status != "uploading") {
			return nil
		}
		if (session.MediaAssetID == nil) != (candidate.MediaAssetID == nil) ||
			(session.MediaAssetID != nil && *session.MediaAssetID != *candidate.MediaAssetID) {
			return nil
		}
		bindingCount, err := repo.CountCustomerServiceMessageBindingsForUploadSession(session.ID)
		if err != nil {
			return err
		}
		if bindingCount > 0 {
			return nil
		}
		if asset != nil {
			if err := validateCustomerServiceRetirementUploadSessionAsset(session, asset, now); err != nil {
				if errors.Is(err, ErrCustomerServiceAttachmentWritesInProgress) {
					return nil
				}
				return err
			}
		} else if !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(session.ObjectKey) {
			return ErrMediaAssetForbidden
		}
		if err := repo.DeleteRetiredCustomerServiceUploadSession(session.ID); err != nil {
			return err
		}
		if asset != nil {
			if err := s.retireCustomerServiceOwnedPrivateAssetIfUnreferenced(repo, tx, asset); err != nil {
				return err
			}
		} else if err := s.queueRetiredUnregisteredCustomerServiceUploadOriginal(tx, *session); err != nil {
			return err
		}
		retired = true
		return nil
	})
	return retired, err
}

func validateCustomerServiceRetirementUploadSessionAsset(session *media.MediaUploadSession, asset *media.MediaAsset, now time.Time) error {
	if asset.CustomerServiceDerivativeWriteExpiresAt != nil && asset.CustomerServiceDerivativeWriteExpiresAt.After(now) {
		return ErrCustomerServiceAttachmentWritesInProgress
	}
	key := assetStorageKey(asset)
	if asset.Visibility != "private" || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) ||
		!customerServiceUploadSessionMatchesRegisteredAsset(session, asset, key) {
		return ErrMediaAssetForbidden
	}
	return nil
}

func (s *CustomerServiceRetentionService) queueRetiredUnregisteredCustomerServiceUploadOriginal(tx *gorm.DB, session media.MediaUploadSession) error {
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceCustomerServiceUnregisteredUpload, session.ID,
		outbox.AggregateTypeCustomerServiceConversation, strconv.FormatUint(uint64(session.BusinessContextID), 10), []string{session.ObjectKey})
	if err != nil {
		return err
	}
	return s.cleanupOutbox.WithTx(tx).CreateEvent(event)
}

func (s *CustomerServiceRetentionService) retireCustomerServiceOwnedPrivateAssetIfUnreferenced(repo *repository.MediaRepository, tx *gorm.DB, asset *media.MediaAsset) error {
	// Public assets remain media-library-owned, even after their last chat use.
	if asset.Visibility != "private" || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(assetStorageKey(asset)) {
		return nil
	}
	err := s.mediaService.retireUnreferencedMediaAssetWithinTransaction(repo, tx, asset)
	if errors.Is(err, ErrMediaAssetInUse) {
		return nil
	}
	return err
}

func (s *CustomerServiceRetentionService) retireManagedCustomerServiceConversationAttachmentsWithinTransaction(repo *repository.TicketRepository, tx *gorm.DB, conversation ticket.Ticket) error {
	now := s.currentTime()
	if !conversation.DeletedAt.Valid || conversation.DeletedAt.Time.Add(s.recoveryWindow).After(now) {
		return ErrCustomerServiceRetentionWindow
	}
	sessions, err := repo.FindCustomerServiceConversationManagedUploadSessions(conversation.ID)
	if err != nil {
		return err
	}
	assetIDs, err := repo.FindCustomerServiceConversationBoundMediaAssetIDs(conversation.ID)
	if err != nil {
		return err
	}
	uniqueAssetIDs := make(map[uint]bool)
	for _, id := range assetIDs {
		uniqueAssetIDs[id] = true
	}
	for _, session := range sessions {
		if session.CreatedAt.After(now.Add(-media.CustomerServiceAttachmentUploadMaximumWriteDuration)) {
			return ErrCustomerServiceAttachmentWritesInProgress
		}
		if session.MediaAssetID != nil {
			uniqueAssetIDs[*session.MediaAssetID] = true
		}
	}
	mediaRepo := s.mediaService.repo.WithTx(tx)
	legacy, err := repo.FindCustomerServiceConversationAttachmentReferenceScan(conversation.ID)
	if err != nil {
		return err
	}
	for _, reference := range legacy.References {
		key := storageObjectKey(s.mediaService.storage, reference)
		if key == "" {
			continue
		}
		asset, err := mediaRepo.FindAssetByStorageKey(key)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		uniqueAssetIDs[asset.ID] = true
	}
	assetIDs = assetIDs[:0]
	for id := range uniqueAssetIDs {
		assetIDs = append(assetIDs, id)
	}
	sort.Slice(assetIDs, func(i, j int) bool { return assetIDs[i] < assetIDs[j] })
	assets := make(map[uint]*media.MediaAsset, len(assetIDs))
	for _, id := range assetIDs {
		asset, err := mediaRepo.FindCustomerServiceAttachmentAssetUnscopedForUpdate(id)
		if err != nil {
			return err
		}
		if asset.CustomerServiceDerivativeWriteExpiresAt != nil && asset.CustomerServiceDerivativeWriteExpiresAt.After(now) {
			return ErrCustomerServiceAttachmentWritesInProgress
		}
		assets[id] = asset
	}
	for _, snapshot := range sessions {
		session, err := mediaRepo.FindCustomerServiceUploadSessionForUpdate(context.Background(), snapshot.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if session.BusinessContextID != conversation.ID {
			return fmt.Errorf("customer-service upload changed parent while retiring")
		}
		if session.MediaAssetID != nil {
			asset := assets[*session.MediaAssetID]
			if asset == nil {
				return ErrCustomerServiceAttachmentWritesInProgress
			}
			if err := validateCustomerServiceRetirementUploadSessionAsset(session, asset, now); err != nil {
				return err
			}
		} else {
			if !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(session.ObjectKey) {
				return ErrMediaAssetForbidden
			}
			if err := s.queueRetiredUnregisteredCustomerServiceUploadOriginal(tx, *session); err != nil {
				return err
			}
		}
	}
	if err := repo.DeleteCustomerServiceConversationManagedAttachmentsAndMessages(conversation.ID); err != nil {
		return err
	}
	for _, id := range assetIDs {
		if err := s.retireCustomerServiceOwnedPrivateAssetIfUnreferenced(mediaRepo, tx, assets[id]); err != nil {
			return err
		}
	}
	return nil
}

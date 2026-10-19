package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"path"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *MediaService) uploadPermanentMediaOriginalWithPrecommittedReceiptAndAtomicAssetRegistration(ctx context.Context, input MediaUploadInput, classification trustedMediaAssetUploadClassification, mimeType, contentHash string) (*media.MediaAsset, error) {
	writer, available := s.storage.(storage.PreallocatedPermanentMediaOriginalObjectWriter)
	if !available {
		return nil, ErrMediaStorageUnavailable
	}
	if _, available := s.storage.(storage.ObjectOpener); !available {
		return nil, ErrMediaStorageUnavailable
	}
	if s.objectCleanupOutbox == nil {
		return nil, ErrObjectStorageCleanupUnavailable
	}
	if err := s.repo.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return nil, err
	}
	copyrightClaim, err := s.buildCopyrightClaim(input, contentHash)
	if err != nil {
		return nil, err
	}
	extension := strings.ToLower(path.Ext(input.File.Filename))
	switch mimeType {
	case "image/png":
		extension = ".png"
	case "image/jpeg":
		extension = ".jpg"
	case "image/webp":
		extension = ".webp"
	case "image/gif":
		extension = ".gif"
	}
	key, err := storage.AllocatePermanentMediaOriginalObjectKey(strings.TrimSpace(input.StoragePrefix), extension)
	if err != nil {
		return nil, err
	}
	if err := validateFailedSharedMediaOriginalCleanupOwnership(key); err != nil {
		return nil, err
	}
	expectedURL := s.storage.GetURL(key)
	if err := s.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(expectedURL, key); err != nil {
		return nil, err
	}
	identity := input.UploadActorIdentity
	asset := &media.MediaAsset{
		Filename: path.Base(key), OriginalFilename: input.File.Filename, URL: expectedURL, StorageKey: key,
		MimeType: mimeType, MediaType: input.MediaType, Size: input.File.Size, Width: input.Width, Height: input.Height,
		ContentSHA256: contentHash, CopyrightClaimJSON: copyrightClaim, Alt: strings.TrimSpace(input.Alt), Caption: strings.TrimSpace(input.Caption),
		UploaderID: identity.AccountID, UploaderActorType: identity.ActorType, UploaderVisitorSessionHash: identity.VisitorSessionHash, Status: "active", Visibility: "public",
		LibraryPartition: classification.partition, UploadOrigin: classification.origin, ClassificationMethod: classification.method, CustomerServiceMediaPurpose: classification.customerServicePurpose, UploadPurpose: input.UploadPurpose,
	}
	resourceID := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	receipt, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaUpload, resourceID, "media_asset", resourceID, []string{key}, map[string]int64{key: input.File.Size})
	if err != nil {
		return nil, err
	}
	receipt.EventKey = repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix + uuid.NewString()
	err = s.repo.RunPermanentMediaOriginalUploadWithPrecommittedCleanupReceiptAndAtomicAssetExchange(ctx, identity, receipt, func(repo *repository.MediaRepository, tx *gorm.DB) error {
		used, err := repo.AssetStorageUsageByUploadActor(identity)
		if err != nil {
			return err
		}
		quota := s.storageQuotaBytesForMediaUploadActor(identity)
		if quota <= 0 || used >= quota || asset.Size > quota-used {
			return repository.ErrMediaUploadActorStorageQuotaExceeded
		}
		reader, err := input.File.Open()
		if err != nil {
			return err
		}
		reference, writeError := writer.WritePermanentMediaOriginalObjectAtPreallocatedKey(tx.Statement.Context, reader, key)
		if err := errors.Join(writeError, reader.Close()); err != nil {
			return err
		}
		if err := s.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(reference, key); err != nil {
			return err
		}
		if err := s.verifyStoredMediaObjectSourceSizeDigestAndMimeType(tx.Statement.Context, asset); err != nil {
			return err
		}
		return repo.CreateAssetWithinUploadActorStorageQuota(tx.Statement.Context, asset, quota)
	})
	if errors.Is(err, repository.ErrMediaUploadActorStorageQuotaExceeded) {
		err = mediaStorageQuotaExceededErrorForUploadActor(identity)
	}
	if err != nil {
		return nil, err
	}
	return s.finishRegisteredMediaAssetUploadWithBestEffortDerivatives(ctx, input, asset)
}

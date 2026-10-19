package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

const businessDraftMediaUploadLifetime = 24 * time.Hour
const businessDraftMediaUploadMaximumWriteDuration = 5 * time.Minute
const objectCleanupResourceExpiredBusinessDraftMediaUpload = "expired_business_draft_media_upload"

type preparedBusinessDraftMediaDerivative struct {
	payload  []byte
	metadata media.MediaAssetDerivative
}

// Only authorized business editors call this method. Media-library uploads keep
// UploadAsset's permanent semantics, even when their classification is catalog.
func (s *MediaService) UploadBusinessDraftMediaAsset(ctx context.Context, input MediaUploadInput) (*media.MediaAsset, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if s.objectCleanupOutbox == nil {
		return nil, ErrObjectStorageCleanupUnavailable
	}
	if err := s.repo.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return nil, err
	}
	writer, available := s.storage.(storage.PreallocatedBusinessDraftMediaObjectWriter)
	if !available {
		return nil, ErrMediaStorageUnavailable
	}
	if input.File == nil {
		return nil, ErrMediaUploadFileRequired
	}
	identity, err := resolveMediaUploadActorIdentity(input)
	if err != nil {
		return nil, err
	}
	if identity.ActorType != media.MediaUploadActorAdministrator || input.PreparedCustomerServiceUploadSessionID != "" || input.StoragePrefix != "" {
		return nil, ErrMediaAssetForbidden
	}
	switch input.UploadOrigin {
	case media.MediaUploadOriginAdministratorCatalog, media.MediaUploadOriginAdministratorContent, media.MediaUploadOriginAdministratorWebsiteProfile, media.MediaUploadOriginAdministratorWorkbenchFeed:
	default:
		return nil, ErrMediaAssetForbidden
	}
	input.MediaType, err = normalizeMediaType(input.MediaType)
	if err != nil {
		return nil, err
	}
	input.UploadPurpose = strings.TrimSpace(input.UploadPurpose)
	if input.MediaType == "video" {
		input.UploadPurpose = ""
	}
	classification, err := resolveMediaAssetClassificationFromTrustedUploadOrigin(input, identity)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMediaUploadActorStorageQuota(identity, input.File.Size); err != nil {
		return nil, err
	}
	mimeType, err := trustedMediaUploadMimeType(input.File, input.MediaType)
	if err != nil {
		return nil, err
	}
	contentHash, err := contentSHA256FromMultipartFile(input.File)
	if err != nil {
		return nil, err
	}
	input.UploaderID, input.UploadActorIdentity = identity.AccountID, identity
	copyrightClaim, err := s.buildCopyrightClaim(input, contentHash)
	if err != nil {
		return nil, err
	}
	receiptID := uuid.NewString()
	extension := path.Ext(input.File.Filename)
	// Use sniffed content for image extensions, including disguised JPEG names.
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
	key, err := storage.AllocateBusinessDraftMediaObjectKey(receiptID, extension)
	if err != nil {
		return nil, err
	}
	asset := &media.MediaAsset{
		Filename: path.Base(key), OriginalFilename: input.File.Filename, URL: s.storage.GetURL(key), StorageKey: key,
		MimeType: mimeType, MediaType: input.MediaType, Size: input.File.Size, Width: input.Width, Height: input.Height,
		ContentSHA256: contentHash, CopyrightClaimJSON: copyrightClaim, Alt: strings.TrimSpace(input.Alt), Caption: strings.TrimSpace(input.Caption),
		UploaderID: identity.AccountID, UploaderActorType: identity.ActorType, Status: "active", Visibility: "public",
		LibraryPartition: classification.partition, UploadOrigin: classification.origin, ClassificationMethod: classification.method, UploadPurpose: input.UploadPurpose,
	}
	prepared, err := s.prepareBusinessDraftMediaDerivativesBeforeAnyObjectWrite(ctx, input, asset, receiptID)
	if err != nil {
		// Conversion can reject an otherwise valid animated image. No objects have
		// been written yet, so keep the validated original with a one-key receipt.
		if errors.Is(err, ErrMediaDerivativeGenerationFailed) && ctx.Err() == nil {
			prepared = nil
		} else {
			return nil, err
		}
	}
	keys := []string{key}
	for _, derivative := range prepared {
		keys = append(keys, derivative.metadata.StorageKey)
		asset.Derivatives = append(asset.Derivatives, derivative.metadata)
	}
	encodedKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	receipt := &media.BusinessDraftMediaUploadReceipt{ID: receiptID, UploaderID: identity.AccountID, UploadOrigin: classification.origin,
		OriginalObjectKey: key, ObjectKeysJSON: encodedKeys, ContentSHA256: contentHash, Size: input.File.Size, MimeType: mimeType,
		Status: "uploading", CreatedAt: now, ExpiresAt: now.Add(businessDraftMediaUploadLifetime), NextExpiryAttemptAt: now}
	if err := s.repo.CreateBusinessDraftMediaUploadReceipt(ctx, receipt); err != nil {
		return nil, err
	}
	writeContext, cancel := context.WithTimeout(ctx, businessDraftMediaUploadMaximumWriteDuration)
	defer cancel()
	// No quota lock is acquired while these receipt/object locks are held.
	// The transaction lifetime survives cancellation until synchronous IO returns.
	err = s.withinMediaPhysicalCleanupTransactionHoldingLocksUntilProviderReturns(writeContext, func(repo *repository.MediaRepository, _ *gorm.DB) error {
		current, err := repo.FindBusinessDraftMediaUploadReceiptForUpdate(receiptID)
		if err != nil {
			return err
		}
		if current.Status != "uploading" || current.AssetID != nil || !current.ExpiresAt.After(time.Now().UTC()) {
			return repository.ErrBusinessDraftMediaUploadReceiptUnavailable
		}
		if err := repo.LockBusinessDraftMediaUploadObjectsWithinTransaction(current); err != nil {
			return err
		}
		source, err := input.File.Open()
		if err != nil {
			return err
		}
		returnedURL, writeErr := writer.WriteBusinessDraftMediaObjectAtPreallocatedKey(writeContext, source, key)
		closeErr := source.Close()
		if writeErr != nil || closeErr != nil {
			return errors.Join(writeErr, closeErr)
		}
		if err := s.validateBusinessDraftMediaProviderReturnedExactObjectKey(returnedURL, key); err != nil {
			return err
		}
		if err := s.verifyWrittenBusinessDraftMediaObjectSizeAndDigest(writeContext, key, asset.Size, asset.ContentSHA256); err != nil {
			return err
		}
		for _, derivative := range prepared {
			returnedURL, err := writer.WriteBusinessDraftMediaObjectAtPreallocatedKey(writeContext, bytes.NewReader(derivative.payload), derivative.metadata.StorageKey)
			if err != nil {
				return err
			}
			if err := s.validateBusinessDraftMediaProviderReturnedExactObjectKey(returnedURL, derivative.metadata.StorageKey); err != nil {
				return err
			}
			digest := sha256.Sum256(derivative.payload)
			if err := s.verifyWrittenBusinessDraftMediaObjectSizeAndDigest(writeContext, derivative.metadata.StorageKey, int64(len(derivative.payload)), fmt.Sprintf("%x", digest)); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		err = s.repo.CreateBusinessDraftMediaAssetWithinUploadActorStorageQuota(writeContext, asset, s.storageQuotaBytesForMediaUploadActor(identity), receiptID)
	}
	if err != nil {
		cleanupContext, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cleanupCancel()
		scheduleErr := s.repo.ScheduleFailedBusinessDraftMediaUploadReceiptForExpiry(cleanupContext, receiptID)
		if errors.Is(err, repository.ErrMediaUploadActorStorageQuotaExceeded) {
			err = mediaStorageQuotaExceededErrorForUploadActor(identity)
		}
		return nil, errors.Join(err, scheduleErr)
	}
	s.hydrateAssetAccessURL(asset)
	return asset, nil
}

func (s *MediaService) verifyWrittenBusinessDraftMediaObjectSizeAndDigest(ctx context.Context, key string, size int64, expectedDigest string) error {
	opener, available := s.storage.(storage.ObjectOpener)
	if !available {
		return ErrMediaStorageUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		return fmt.Errorf("%w: draft object cannot be verified", ErrMediaAssetURLUnavailable)
	}
	if object == nil || object.ReadCloser == nil {
		return ErrMediaAssetURLUnavailable
	}
	digest := sha256.New()
	copied, readErr := io.Copy(digest, io.LimitReader(object.ReadCloser, size+1))
	closeErr := object.ReadCloser.Close()
	if readErr != nil || closeErr != nil || copied != size || fmt.Sprintf("%x", digest.Sum(nil)) != expectedDigest {
		return fmt.Errorf("%w: draft object integrity mismatch", ErrMediaAssetURLUnavailable)
	}
	return ctx.Err()
}

func (s *MediaService) validateBusinessDraftMediaProviderReturnedExactObjectKey(reference, expectedKey string) error {
	key, err := s.storage.ObjectKey(reference)
	if err != nil || key != expectedKey || strings.TrimSpace(reference) == "" {
		return ErrMediaAssetURLUnavailable
	}
	return nil
}

func (s *MediaService) prepareBusinessDraftMediaDerivativesBeforeAnyObjectWrite(ctx context.Context, input MediaUploadInput, asset *media.MediaAsset, receiptID string) ([]preparedBusinessDraftMediaDerivative, error) {
	if asset.MediaType != "image" {
		return nil, nil
	}
	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return nil, err
	}
	release, err := acquireMediaDerivativeGenerationSlot(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	decoded, err := decodeDerivativeSource(func() (io.ReadCloser, error) { return input.File.Open() })
	if err != nil {
		return nil, err
	}
	asset.Width, asset.Height = decoded.Bounds().Dx(), decoded.Bounds().Dy()
	var prepared []preparedBusinessDraftMediaDerivative
	for _, preset := range presets {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		resized := resizeImageToMaxWidth(decoded, preset.MaxWidth)
		payload, filename, mimeType, width, height, err := encodeDerivativeImage(asset, resized, preset, imageHasAlpha(decoded))
		if err != nil {
			return nil, err
		}
		key, err := storage.AllocateBusinessDraftMediaObjectKey(receiptID, path.Ext(filename))
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, preparedBusinessDraftMediaDerivative{payload: payload, metadata: media.MediaAssetDerivative{
			Preset: preset.Name, PresetVersion: mediaDerivativePresetVersion(preset), StorageKey: key, URL: s.storage.GetURL(key), MimeType: mimeType, Width: width, Height: height, Size: int64(len(payload)),
		}})
	}
	return prepared, nil
}

type BusinessDraftMediaUploadExpiryBatchResult struct{ Scanned, Retired, Blocked, Errors int }

func (s *MediaService) ExpireUnboundBusinessDraftMediaUploads(ctx context.Context, limit int) (BusinessDraftMediaUploadExpiryBatchResult, error) {
	result := BusinessDraftMediaUploadExpiryBatchResult{}
	if s == nil || s.repo == nil || s.objectCleanupOutbox == nil {
		return result, ErrObjectStorageCleanupUnavailable
	}
	if err := s.repo.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return result, err
	}
	candidates, err := s.repo.FindExpiredBusinessDraftMediaUploadReceipts(ctx, time.Now().UTC(), limit)
	if err != nil {
		return result, err
	}
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		result.Scanned++
		retired := false
		err := s.repo.WithinMediaAssetTransactionWithRequestContext(ctx, func(repo *repository.MediaRepository, tx *gorm.DB) error {
			// Asset -> receipt -> object, matching business binding and retirement.
			var asset *media.MediaAsset
			if candidate.AssetID != nil {
				var err error
				asset, err = repo.FindAssetByIDUnscopedForUpdate(*candidate.AssetID)
				if err != nil {
					return err
				}
			}
			receipt, err := repo.FindBusinessDraftMediaUploadReceiptForUpdate(candidate.ID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			if receipt.Status == "bound" || receipt.ExpiresAt.After(time.Now().UTC()) {
				return nil
			}
			// Registration may have committed since candidate enumeration. Retry with
			// its asset lock on the next pass; never acquire an asset after its receipt.
			if (receipt.AssetID == nil) != (candidate.AssetID == nil) {
				return nil
			}
			if receipt.AssetID != nil && *receipt.AssetID != *candidate.AssetID {
				return repository.ErrBusinessDraftMediaUploadReceiptUnavailable
			}
			if receipt.Status != "uploading" && receipt.Status != "ready" {
				return repository.ErrBusinessDraftMediaUploadReceiptUnavailable
			}
			if asset != nil {
				if err := s.retireUnreferencedMediaAssetWithinTransaction(repo, tx, asset); err != nil {
					return err
				}
			} else {
				keys, err := repository.BusinessDraftMediaUploadReceiptObjectKeys(receipt)
				if err != nil {
					return err
				}
				event, err := newObjectStorageCleanupEvent(objectCleanupResourceExpiredBusinessDraftMediaUpload, receipt.ID, "media_asset", receipt.ID, keys)
				if err != nil {
					return err
				}
				if err := s.objectCleanupOutbox.WithTx(tx).CreateEvent(event); err != nil {
					return err
				}
				if err := repo.DeleteBusinessDraftMediaUploadReceiptWithinTransaction(receipt.ID); err != nil {
					return err
				}
			}
			retired = true
			return nil
		})
		if err != nil {
			if errors.Is(err, ErrMediaAssetInUse) {
				result.Blocked++
			} else {
				result.Errors++
			}
			// Keep the original receipt/deadline, but let the next bounded batch move
			// beyond blocked history. A successful binding cannot be downgraded here.
			if retryErr := s.repo.DelayRetryOfBlockedBusinessDraftMediaUploadExpiry(ctx, candidate.ID); retryErr != nil {
				return result, retryErr
			}
		} else if retired {
			result.Retired++
		}
	}
	return result, nil
}

func (s *MediaService) cleanupExpiredBusinessDraftMediaObjectsFromDurableReceipt(ctx context.Context, receiptID string, keys []string) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if err := s.repo.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return err
	}
	if len(keys) == 0 || len(keys) > 13 {
		return repository.ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	for _, key := range keys {
		id, valid := storage.BusinessDraftMediaReceiptIDFromObjectKey(key)
		if !valid || id != receiptID {
			return fmt.Errorf("expired draft cleanup object does not belong to its receipt")
		}
	}
	return s.withinMediaPhysicalCleanupTransactionHoldingLocksUntilProviderReturns(ctx, func(repo *repository.MediaRepository, _ *gorm.DB) error {
		encodedKeys, _ := json.Marshal(keys)
		receipt := &media.BusinessDraftMediaUploadReceipt{ID: receiptID, OriginalObjectKey: keys[0], ObjectKeysJSON: encodedKeys}
		if err := repo.LockBusinessDraftMediaUploadObjectsWithinTransaction(receipt); err != nil {
			return err
		}
		if err := repo.EnsureBusinessDraftMediaReceiptIsAbsentBeforeExpiredObjectCleanup(receiptID); err != nil {
			return err
		}
		for _, key := range keys {
			references, err := repo.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
			if err != nil {
				return err
			}
			if len(references) > 0 {
				return &MediaAssetInUseError{References: references}
			}
		}
		for _, key := range keys {
			if err := s.storage.Delete(ctx, key); err != nil {
				return err
			}
		}
		return nil
	})
}

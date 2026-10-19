package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

var ErrPublicMediaDerivativeRegistrationExchangeFailed = errors.New("public media derivative registration exchange failed")

type plannedPublicMediaDerivativeExactObjectWrite struct {
	derivative media.MediaAssetDerivative
	preset     MediaDerivativePresetDefinition
}

func (s *MediaService) generateAndPersistPublicMediaDerivativesWithPrecommittedExactKeyReceipts(ctx context.Context, asset *media.MediaAsset, openSource mediaDerivativeSourceOpener, presets []MediaDerivativePresetDefinition, persist func(*repository.MediaRepository, *gorm.DB, []media.MediaAssetDerivative) error) ([]media.MediaAssetDerivative, error) {
	if s == nil || s.repo == nil || s.storage == nil || asset == nil || openSource == nil || persist == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if len(presets) > mediaDerivativeActivePresetLimit {
		return nil, ErrInvalidMediaDerivativePreset
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if asset.Visibility != "public" || storage.IsPrivateObjectKey(assetStorageKey(asset)) {
		return nil, ErrMediaAssetForbidden
	}
	writer, available := s.storage.(storage.PreallocatedPublicMediaDerivativeObjectWriter)
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
	if err := s.repo.ValidateBusinessDraftMediaDerivativeGenerationIsPermanentlyBound(asset); err != nil {
		return nil, err
	}
	release, err := acquireMediaDerivativeGenerationSlot(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	decoded, err := decodeDerivativeSource(openSource)
	if err != nil {
		return nil, err
	}
	preserveAlpha := imageHasAlpha(decoded)
	extension, mimeType := ".jpg", "image/jpeg"
	if preserveAlpha {
		extension, mimeType = ".png", "image/png"
	}
	plans := make([]plannedPublicMediaDerivativeExactObjectWrite, 0, len(presets))
	receipts := make([]*outbox.Event, 0, len(presets))
	seenPresets := make(map[string]bool)
	for _, preset := range presets {
		if !isValidMediaDerivativePresetCode(preset.Name) || preset.MaxWidth <= 0 || preset.MaxWidth > mediaDerivativeMaxDimension || seenPresets[preset.Name] {
			return nil, ErrInvalidMediaDerivativePreset
		}
		seenPresets[preset.Name] = true
		key, err := storage.AllocatePublicMediaDerivativeObjectKey(asset.ID, preset.Name, mediaDerivativePresetVersion(preset), extension)
		if err != nil {
			return nil, err
		}
		if err := validateMediaDerivativeObjectCleanupOwnership(asset, key); err != nil {
			return nil, err
		}
		reference := s.storage.GetURL(key)
		if err := s.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(reference, key); err != nil {
			return nil, err
		}
		derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: preset.Name, PresetVersion: mediaDerivativePresetVersion(preset), URL: reference, StorageKey: key, MimeType: mimeType}
		plans = append(plans, plannedPublicMediaDerivativeExactObjectWrite{derivative: derivative, preset: preset})
		parentID := strconv.FormatUint(uint64(asset.ID), 10)
		receipt, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaDerivative, parentID, "media_asset", parentID, []string{key})
		if err != nil {
			return nil, err
		}
		receipt.EventKey = repository.PublicMediaDerivativeUploadReceiptEventKeyPrefix + uuid.NewString()
		receipts = append(receipts, receipt)
	}
	if len(plans) == 0 {
		return nil, nil
	}
	var derivatives []media.MediaAssetDerivative
	err = s.repo.RunPublicMediaDerivativeBatchWithPrecommittedReceiptsAndAtomicRegistration(ctx, asset, receipts, func(repo *repository.MediaRepository, tx *gorm.DB) error {
		for _, plan := range plans {
			if err := tx.Statement.Context.Err(); err != nil {
				return errors.Join(ErrMediaDerivativeGenerationFailed, err)
			}
			// Keep only one encoded payload alive while all planned keys remain protected.
			resized := resizeImageToMaxWidth(decoded, plan.preset.MaxWidth)
			payload, _, actualMime, width, height, err := encodeDerivativeImage(asset, resized, plan.preset, preserveAlpha)
			if err != nil {
				return err
			}
			plan.derivative.MimeType, plan.derivative.Width, plan.derivative.Height, plan.derivative.Size = actualMime, width, height, int64(len(payload))
			reference, err := writer.WritePublicMediaDerivativeObjectAtPreallocatedKey(tx.Statement.Context, bytes.NewReader(payload), plan.derivative.StorageKey)
			if err != nil {
				return fmt.Errorf("%w: public derivative exact-key write failed", ErrMediaDerivativeGenerationFailed)
			}
			if err := s.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(reference, plan.derivative.StorageKey); err != nil {
				return errors.Join(ErrMediaDerivativeGenerationFailed, err)
			}
			expected := &media.MediaAsset{StorageKey: plan.derivative.StorageKey, Size: plan.derivative.Size, MimeType: plan.derivative.MimeType, ContentSHA256: fmt.Sprintf("%x", sha256.Sum256(payload))}
			if err := s.verifyStoredMediaObjectSourceSizeDigestAndMimeType(tx.Statement.Context, expected); err != nil {
				return errors.Join(ErrMediaDerivativeGenerationFailed, err)
			}
			derivatives = append(derivatives, plan.derivative)
		}
		return persist(repo, tx, derivatives)
	})
	if errors.Is(err, repository.ErrPublicMediaDerivativeUploadReceiptPersistenceUnavailable) {
		return nil, errors.Join(ErrObjectStorageCleanupUnavailable, err)
	}
	if err != nil {
		if errors.Is(err, ErrMediaDerivativeGenerationFailed) {
			return nil, err
		}
		return nil, errors.Join(ErrPublicMediaDerivativeRegistrationExchangeFailed, err)
	}
	return derivatives, nil
}

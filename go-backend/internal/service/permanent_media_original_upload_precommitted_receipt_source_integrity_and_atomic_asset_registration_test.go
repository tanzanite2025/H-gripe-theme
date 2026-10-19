package service

import (
	"encoding/json"
	"errors"
	"path"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPermanentMediaOriginalUploadCommitsExactReceiptBeforeWritingAndAtomicallyRegistersAsset(t *testing.T) {
	service, provider := newMediaAssetUploadTestService(t)
	assets := service.repo
	provider.onPermanentWrite = func(key string) {
		require.NoError(t, assets.WithTransaction(func(_ *repository.MediaRepository, tx *gorm.DB) error {
			var receipt outbox.Event
			if err := tx.Where("event_key LIKE ?", repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix+"%").First(&receipt).Error; err != nil {
				return err
			}
			require.Equal(t, outbox.EventStatusPending, receipt.Status)
			var payload outbox.ObjectStorageCleanupPayload
			if err := json.Unmarshal(receipt.Payload, &payload); err != nil {
				return err
			}
			require.Equal(t, []string{key}, payload.ObjectKeys)
			var assetCount int64
			if err := tx.Model(&media.MediaAsset{}).Count(&assetCount).Error; err != nil {
				return err
			}
			require.Zero(t, assetCount)
			return nil
		}))
	}
	input := MediaUploadInput{
		File: multipartFileHeader(t, "permanent-source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image",
		UploadOrigin: media.MediaUploadOriginAdministratorCatalog, UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42},
	}

	asset, err := service.UploadAsset(t.Context(), input)
	require.NoError(t, err)
	require.NotNil(t, asset)
	require.True(t, storage.IsPreallocatedPermanentMediaOriginalObjectKey(asset.StorageKey))
	require.Equal(t, path.Base(asset.StorageKey), asset.Filename)
	require.Equal(t, "catalog/product-images/preallocated-permanent-originals/", path.Dir(asset.StorageKey)+"/")
	require.Equal(t, int64(len(mediaUploadTestPNG(t))), asset.Size)
	require.NotEmpty(t, asset.ContentSHA256)
	require.Nil(t, asset.BusinessDraftUploadExpiresAt)
	require.Empty(t, asset.PendingCustomerServiceUploadSessionID)
	assetCount, receiptCount := permanentMediaOriginalUploadAssetAndReceiptCounts(t, assets)
	require.EqualValues(t, 1, assetCount)
	require.Zero(t, receiptCount)
	require.Equal(t, asset.StorageKey, onlyRecordedMediaUploadObjectKey(t, provider))
}

func TestPermanentMediaOriginalUploadKeepsOriginalReceiptWhenProviderWritesButLosesResponse(t *testing.T) {
	service, provider := newMediaAssetUploadTestService(t)
	provider.permanentWriteResponseError = errors.New("simulated provider lost response after accepting bytes")
	asset, err := service.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "lost-response.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42,
	})
	require.Error(t, err)
	require.Nil(t, asset)
	key := onlyRecordedMediaUploadObjectKey(t, provider)
	require.True(t, storage.IsPreallocatedPermanentMediaOriginalObjectKey(key))
	assetCount, receiptCount := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
	require.Zero(t, assetCount)
	require.EqualValues(t, 1, receiptCount)
	receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
	require.Equal(t, outbox.EventStatusPending, receipt.Status)
	var payload outbox.ObjectStorageCleanupPayload
	require.NoError(t, json.Unmarshal(receipt.Payload, &payload))
	require.Equal(t, []string{key}, payload.ObjectKeys)
}

func TestPermanentMediaOriginalUploadRejectsWrongProviderIdentityAndRetainsExactReceipt(t *testing.T) {
	service, provider := newMediaAssetUploadTestService(t)
	provider.permanentWriteReturnWrongKey = true
	asset, err := service.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "wrong-identity.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42,
	})
	require.Error(t, err)
	require.Nil(t, asset)
	key := onlyRecordedMediaUploadObjectKey(t, provider)
	_, receiptCount := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
	require.EqualValues(t, 1, receiptCount)
	receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
	var payload outbox.ObjectStorageCleanupPayload
	require.NoError(t, json.Unmarshal(receipt.Payload, &payload))
	require.Equal(t, []string{key}, payload.ObjectKeys)
	require.Equal(t, outbox.EventStatusPending, receipt.Status)
}

func TestPermanentMediaOriginalUploadFailsClosedWhenProviderCannotPreallocateOrReadSource(t *testing.T) {
	for _, missing := range []string{"preallocated writer", "source reader"} {
		t.Run(missing, func(t *testing.T) {
			service, provider := newMediaAssetUploadTestService(t)
			service.storage = permanentMediaOriginalUploadLegacyProviderWithoutPreallocatedWriter{StorageService: provider}
			if missing == "source reader" {
				service.storage = permanentMediaOriginalUploadProviderWithoutSourceReader{StorageService: provider, PreallocatedPermanentMediaOriginalObjectWriter: provider}
			}
			asset, err := service.UploadAsset(t.Context(), MediaUploadInput{
				File: multipartFileHeader(t, "legacy-provider.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42,
			})
			require.ErrorIs(t, err, ErrMediaStorageUnavailable)
			require.Nil(t, asset)
			require.Empty(t, provider.objects)
			assets, receipts := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
			require.Zero(t, assets)
			require.Zero(t, receipts)
		})
	}
}

type permanentMediaOriginalUploadProviderWithoutSourceReader struct {
	storage.StorageService
	storage.PreallocatedPermanentMediaOriginalObjectWriter
}

type permanentMediaOriginalUploadLegacyProviderWithoutPreallocatedWriter struct {
	storage.StorageService
}

func onlyRecordedMediaUploadObjectKey(t *testing.T, provider *recordingMediaUploadStorage) string {
	t.Helper()
	require.Len(t, provider.objects, 1)
	for key := range provider.objects {
		return key
	}
	t.Fatal("object key missing")
	return ""
}

func permanentMediaOriginalUploadAssetAndReceiptCounts(t *testing.T, repo *repository.MediaRepository) (int64, int64) {
	t.Helper()
	var assets, receipts int64
	require.NoError(t, repo.WithTransaction(func(_ *repository.MediaRepository, tx *gorm.DB) error {
		if err := tx.Model(&media.MediaAsset{}).Count(&assets).Error; err != nil {
			return err
		}
		return tx.Model(&outbox.Event{}).Where("event_key LIKE ?", repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix+"%").Count(&receipts).Error
	}))
	return assets, receipts
}
func permanentMediaOriginalUploadOnlyReceipt(t *testing.T, repo *repository.MediaRepository) outbox.Event {
	t.Helper()
	var receipt outbox.Event
	require.NoError(t, repo.WithTransaction(func(_ *repository.MediaRepository, tx *gorm.DB) error {
		return tx.Where("event_key LIKE ?", repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix+"%").First(&receipt).Error
	}))
	return receipt
}

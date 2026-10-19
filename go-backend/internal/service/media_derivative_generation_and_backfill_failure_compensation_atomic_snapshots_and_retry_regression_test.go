package service

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type derivativeUploadStorageWithPartialGenerationFailure struct {
	storage.StorageService
	failSecondDerivative   bool
	derivativeAttempts     int
	uploadedDerivativeURLs []string
}

func (s *derivativeUploadStorageWithPartialGenerationFailure) UploadFromReaderWithPrefix(ctx context.Context, reader io.Reader, filename, prefix string) (string, error) {
	s.derivativeAttempts++
	if s.failSecondDerivative && s.derivativeAttempts == 2 {
		return "", errors.New("injected second derivative upload failure")
	}
	address, err := s.StorageService.UploadFromReaderWithPrefix(ctx, reader, filename, prefix)
	if err == nil {
		s.uploadedDerivativeURLs = append(s.uploadedDerivativeURLs, address)
	}
	return address, err
}

func (s *derivativeUploadStorageWithPartialGenerationFailure) WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.derivativeAttempts++
	if s.failSecondDerivative && s.derivativeAttempts == 2 {
		return "", errors.New("injected second derivative write failure")
	}
	reference, err := s.StorageService.(storage.PreallocatedPublicMediaDerivativeObjectWriter).WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx, reader, key)
	if err == nil {
		s.uploadedDerivativeURLs = append(s.uploadedDerivativeURLs, reference)
	}
	return reference, err
}

func (s *derivativeUploadStorageWithPartialGenerationFailure) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	return s.StorageService.(storage.PreallocatedPermanentMediaOriginalObjectWriter).WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx, reader, key)
}

func (s *derivativeUploadStorageWithPartialGenerationFailure) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func TestMediaDerivativeGenerationAndRegistrationFailureRetainPrecommittedReceiptsWithoutLosingOriginal(t *testing.T) {
	for _, failure := range []string{"second-upload", "derivative-registration", "compensation-outbox"} {
		t.Run(failure, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			provider := &derivativeUploadStorageWithPartialGenerationFailure{StorageService: assets.storage, failSecondDerivative: failure != "derivative-registration"}
			assets.storage = provider
			if failure != "second-upload" {
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_derivative_failure_test_write", func(tx *gorm.DB) {
					if (failure == "derivative-registration" && tx.Statement.Table == "media_asset_derivatives") ||
						(failure == "compensation-outbox" && tx.Statement.Table == "outbox_events" && !strings.HasPrefix(tx.Statement.Dest.(*outbox.Event).EventKey, repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix)) {
						tx.AddError(errors.New("injected persistence failure"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Create().Remove("reject_derivative_failure_test_write") })
			}
			asset, err := assets.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "partial.png", "image/png", testOpaquePNG(t, 80, 40)), MediaType: "image", UploaderID: 7, Width: 80, Height: 40})
			if failure == "second-upload" {
				require.NoError(t, err, "ordinary conversion failure preserves the original upload contract")
				require.NotNil(t, asset)
			} else {
				require.Error(t, err)
			}
			if failure == "compensation-outbox" {
				require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
			}
			var originals []media.MediaAsset
			require.NoError(t, db.Find(&originals).Error)
			require.Len(t, originals, 1)
			require.FileExists(t, filepath.Join(root, filepath.FromSlash(originals[0].StorageKey)))
			var derivativeCount int64
			require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Count(&derivativeCount).Error)
			require.Zero(t, derivativeCount)
			if failure == "compensation-outbox" {
				require.Empty(t, provider.uploadedDerivativeURLs)
			} else {
				require.NotEmpty(t, provider.uploadedDerivativeURLs)
			}
			for _, address := range provider.uploadedDerivativeURLs {
				key, err := provider.ObjectKey(address)
				require.NoError(t, err)
				require.FileExists(t, filepath.Join(root, filepath.FromSlash(key)), "failed batch remains covered by its original precommitted receipts")
			}
			var events []outbox.Event
			require.NoError(t, db.Find(&events).Error)
			if failure == "compensation-outbox" {
				require.Empty(t, events)
			} else {
				require.Len(t, events, 3, "all planned exact keys have receipts, including never-written objects")
				for _, event := range events {
					require.NoError(t, NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil).Handle(t.Context(), event))
				}
			}
		})
	}
}

func TestMediaDerivativeBackfillCommitsSnapshotsAndCleanupEventsAtomicallyAndRetriesFailedSnapshotUpdate(t *testing.T) {
	for _, failure := range []string{"none", "snapshot-update"} {
		t.Run(failure, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.MediaDerivativePreset{}, &product.ProductMedia{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			presets := repository.NewMediaDerivativePresetRepository(db)
			require.NoError(t, SeedDefaultMediaDerivativePresets(presets))
			assets.ConfigureDerivativePresetRepository(presets)
			asset, err := assets.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "atomic-backfill.png", "image/png", testOpaquePNG(t, 80, 40)), MediaType: "image", UploaderID: 7, Width: 80, Height: 40})
			require.NoError(t, err)
			var oldCard media.MediaAssetDerivative
			for _, derivative := range asset.Derivatives {
				if derivative.Preset == "card" {
					oldCard = derivative
				}
			}
			require.NotZero(t, oldCard.ID)
			variants, thumbnail, err := assets.ProductMediaImageVariants(asset.ID)
			require.NoError(t, err)
			placement := product.ProductMedia{ProductID: 81, MediaAssetID: &asset.ID, MediaType: "image", URL: asset.URL, ThumbnailURL: thumbnail, ImageVariantData: product.ProductMediaImageVariantsJSON(variants), IsVisible: true}
			require.NoError(t, db.Create(&placement).Error)
			require.NoError(t, db.Model(&media.MediaDerivativePreset{}).Where("code = ?", "card").Update("generation_version", 2).Error)
			if failure == "snapshot-update" {
				require.NoError(t, db.Callback().Update().Before("gorm:update").Register("reject_derivative_snapshot_update", func(tx *gorm.DB) {
					if tx.Statement.Table == "product_media" {
						tx.AddError(errors.New("injected snapshot update failure"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Update().Remove("reject_derivative_snapshot_update") })
				_, err := assets.BackfillMissingImageDerivatives(t.Context())
				require.Error(t, err)
				var stillActive media.MediaAssetDerivative
				require.NoError(t, db.First(&stillActive, oldCard.ID).Error, "failed snapshot write must preserve previous active generation")
				require.FileExists(t, filepath.Join(root, filepath.FromSlash(oldCard.StorageKey)))
				var events []outbox.Event
				require.NoError(t, db.Find(&events).Error)
				for _, event := range events {
					require.NotContains(t, string(event.Payload), objectCleanupResourceRetiredMediaDerivative, "rolled-back replacement cannot leave a retired cleanup event")
				}
				require.NoError(t, db.Callback().Update().Remove("reject_derivative_snapshot_update"))
			}
			result, err := assets.BackfillMissingImageDerivatives(t.Context())
			require.NoError(t, err)
			require.Equal(t, 1, result.GeneratedDerivatives)
			require.EqualValues(t, 1, result.UpdatedProductMediaRows)
			require.NoError(t, db.First(&placement, placement.ID).Error)
			nextVariants := product.ParseProductMediaImageVariants(placement.ImageVariantData)
			require.NotEqual(t, oldCard.URL, nextVariants["card"].URL)
			require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(oldCard.StorageKey)), "the old generation can be deleted after the snapshot commit")
			var tombstone media.MediaAssetDerivative
			require.NoError(t, db.Unscoped().First(&tombstone, oldCard.ID).Error)
			require.True(t, tombstone.DeletedAt.Valid)
		})
	}
}

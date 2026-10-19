package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPostgreSQLFailedUploadAndRetiredDerivativeCleanupSerializeAgainstNewOwnershipInBothOrders(t *testing.T) {
	for _, resource := range []string{"failed-original", "failed-public-derivative", "failed-private-derivative", "retired-derivative"} {
		for _, registration := range []string{"original", "derivative"} {
			for _, first := range []string{"registration", "cleanup"} {
				t.Run(resource+"/"+registration+"/"+first+"-first", func(t *testing.T) {
					env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
					ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
					defer cancel()
					db, provider := env.db.WithContext(ctx), env.storage.StorageService
					assets := NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
					assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
					key, root := "catalog/failed-exact-object.png", env.publicRoot
					resourceType, resourceID := objectCleanupResourceUnregisteredMediaUpload, "synthetic-failed-upload"
					var event *outbox.Event
					if resource == "retired-derivative" {
						_, previous, _, stored := createRetiredMediaDerivativeWithDurableCleanupEventForTest(t, db, assets, root)
						key, event = previous.StorageKey, &stored
					} else if resource != "failed-original" {
						parentKey := "catalog/failed-derivative-owner.png"
						if resource == "failed-private-derivative" {
							parentKey = storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/failed-derivative-owner.png"
						}
						parent := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, parentKey)
						prefix := mediaDerivativePrefix
						if resource == "failed-private-derivative" {
							require.NoError(t, db.Model(parent).Update("visibility", "private").Error)
							prefix, root = storage.CustomerServiceConversationAttachmentDerivativesPrefix, env.privateRoot
						}
						key = fmt.Sprintf("%s/%d/card/v1/failed.png", prefix, parent.ID)
						resourceType, resourceID = objectCleanupResourceUnregisteredMediaDerivative, fmt.Sprint(parent.ID)
					}
					if event == nil {
						var err error
						event, err = newObjectStorageCleanupEvent(resourceType, resourceID, "media_asset", resourceID, []string{key})
						require.NoError(t, err)
						require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
					}
					writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
					candidate := &media.MediaAsset{Filename: "new-claim.png", URL: provider.GetURL(key) + "?new-owner=1", StorageKey: key, Status: "active", Visibility: "public"}
					if resource == "failed-private-derivative" {
						candidate.Visibility = "private"
					}
					var derivative media.MediaAssetDerivative
					if registration == "derivative" {
						owner := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, "catalog/other-generation-owner.png")
						derivative = media.MediaAssetDerivative{MediaAssetID: owner.ID, Preset: "card", PresetVersion: 1, StorageKey: key, URL: candidate.URL}
					}
					register := func() error {
						if registration == "derivative" {
							return assets.repo.CreateAssetDerivatives([]media.MediaAssetDerivative{derivative})
						}
						return assets.repo.CreateAsset(candidate)
					}
					acquired, releaseFirst := make(chan int, 1), make(chan struct{})
					var releaseOnce, pauseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
					t.Cleanup(release)
					callbackName := "pause_exact_media_generation_cleanup_" + uuid.NewString()
					if first == "registration" {
						table := "media_assets"
						if registration == "derivative" {
							table = "media_asset_derivatives"
						}
						require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == table {
								pauseOnce.Do(func() {
									pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, releaseFirst)
								})
							}
						}))
						t.Cleanup(func() { _ = db.Callback().Create().Remove(callbackName) })
					} else {
						require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "media_assets" && strings.Contains(strings.Join(tx.Statement.Selects, " "), "AS retained_deleted") {
								pauseOnce.Do(func() {
									pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, releaseFirst)
								})
							}
						}))
						t.Cleanup(func() { _ = db.Callback().Query().Remove(callbackName) })
					}
					handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
					registrationResult, cleanupResult := make(chan error, 1), make(chan error, 1)
					if first == "registration" {
						go func() { registrationResult <- register() }()
					} else {
						go func() { cleanupResult <- handler.Handle(ctx, *event) }()
					}
					pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
					if first == "registration" {
						go func() { cleanupResult <- handler.Handle(ctx, *event) }()
					} else {
						go func() { registrationResult <- register() }()
					}
					waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
					release()
					registrationError := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, registrationResult)
					cleanupError := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult)
					if first == "registration" {
						require.NoError(t, registrationError)
						require.ErrorIs(t, cleanupError, ErrMediaAssetInUse)
						require.FileExists(t, filepath.Join(root, filepath.FromSlash(key)))
						if registration == "derivative" {
							require.NoError(t, db.Unscoped().Where("url = ?", derivative.URL).Delete(&media.MediaAssetDerivative{}).Error)
						} else {
							require.NoError(t, repository.NewMediaRepository(db).HardDeleteAsset(candidate.ID))
						}
						require.NoError(t, handler.Handle(ctx, *event), "retry the original event after detaching the last new owner")
					} else {
						require.NoError(t, cleanupError)
						require.ErrorIs(t, registrationError, ErrMediaAssetURLUnavailable)
					}
					require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(key)))
					require.NoError(t, handler.Handle(ctx, *event))
				})
			}
		}
	}
}

func TestPostgreSQLMediaDerivativeBackfillSnapshotFailureRollsBackGenerationAndCleanupEventThenRetries(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	db := env.db.WithContext(ctx)
	require.NoError(t, db.AutoMigrate(&media.MediaDerivativePreset{}, &product.ProductMedia{}))
	assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	presets := repository.NewMediaDerivativePresetRepository(db)
	require.NoError(t, SeedDefaultMediaDerivativePresets(presets))
	assets.ConfigureDerivativePresetRepository(presets)
	asset, err := assets.UploadAsset(ctx, MediaUploadInput{File: multipartFileHeader(t, "postgres-atomic-snapshots.png", "image/png", testOpaquePNG(t, 80, 40)),
		MediaType: "image", UploaderID: 701, Width: 80, Height: 40})
	require.NoError(t, err)
	var previous media.MediaAssetDerivative
	for _, derivative := range asset.Derivatives {
		if derivative.Preset == "card" {
			previous = derivative
		}
	}
	require.NotZero(t, previous.ID)
	variants, thumbnail, err := assets.ProductMediaImageVariants(asset.ID)
	require.NoError(t, err)
	placement := product.ProductMedia{ProductID: 771, MediaAssetID: &asset.ID, MediaType: "image", URL: asset.URL, ThumbnailURL: thumbnail,
		ImageVariantData: product.ProductMediaImageVariantsJSON(variants), IsVisible: true}
	require.NoError(t, db.Create(&placement).Error)
	require.NoError(t, db.Model(&media.MediaDerivativePreset{}).Where("code = ?", "card").Update("generation_version", 2).Error)
	callbackName := "reject_postgresql_derivative_snapshot_update_" + uuid.NewString()
	require.NoError(t, db.Callback().Update().Before("gorm:update").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "product_media" {
			tx.AddError(errors.New("injected PostgreSQL snapshot update failure"))
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Update().Remove(callbackName) })
	_, err = assets.ensureAssetDerivatives(ctx, asset)
	require.Error(t, err)
	var retained media.MediaAssetDerivative
	require.NoError(t, db.First(&retained, previous.ID).Error)
	require.FileExists(t, filepath.Join(env.publicRoot, filepath.FromSlash(previous.StorageKey)))
	var events []outbox.Event
	require.NoError(t, db.Find(&events).Error)
	for _, event := range events {
		require.NotContains(t, string(event.Payload), objectCleanupResourceRetiredMediaDerivative)
	}
	require.NoError(t, db.Callback().Update().Remove(callbackName))
	result, err := assets.ensureAssetDerivatives(ctx, asset)
	require.NoError(t, err)
	require.Equal(t, 1, result.GeneratedDerivatives)
	require.EqualValues(t, 1, result.UpdatedProductMediaRows)
	require.NoError(t, db.First(&placement, placement.ID).Error)
	require.NotEqual(t, previous.URL, product.ParseProductMediaImageVariants(placement.ImageVariantData)["card"].URL)
	require.NoFileExists(t, filepath.Join(env.publicRoot, filepath.FromSlash(previous.StorageKey)))
	var tombstone media.MediaAssetDerivative
	require.NoError(t, db.Unscoped().First(&tombstone, previous.ID).Error)
	require.True(t, tombstone.DeletedAt.Valid)
}

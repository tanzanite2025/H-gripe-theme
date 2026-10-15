package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLProductMediaBindingAndRetirementSerializePrimaryThumbnailAndPosterReferences(t *testing.T) {
	for _, field := range []string{"primary", "thumbnail", "poster"} {
		for _, firstOperation := range []string{"bind", "retire"} {
			t.Run(field+"/"+firstOperation+" first", func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				require.NoError(t, env.db.AutoMigrate(&product.Product{}, &product.ProductMedia{}))
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				db := env.db.WithContext(ctx)
				assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
				assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				products := repository.NewProductRepository(db)
				products.ConfigureProductMediaTransactionalAssetBindingResolver(assets)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/racing-photo.png")
				placement := product.ProductMedia{URL: "https://external.example.test/unmanaged.png", MediaType: "image"}
				switch field {
				case "primary":
					placement.URL = "/uploads/" + asset.StorageKey
				case "thumbnail":
					placement.ThumbnailURL = "/uploads/" + asset.StorageKey
				case "poster":
					placement.MediaType, placement.URL = "video", "https://external.example.test/unmanaged.webm"
					placement.PosterURL = "/uploads/" + asset.StorageKey
				}
				firstLocked := make(chan int, 1)
				releaseFirst := make(chan struct{})
				var releaseOnce, pauseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
				defer release()
				pauseAfterFirstOperationAcquiresAssetLock := func(tx *gorm.DB) {
					pauseOnce.Do(func() {
						var pid int
						if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
							tx.AddError(err)
							return
						}
						firstLocked <- pid
						select {
						case <-releaseFirst:
						case <-ctx.Done():
							tx.AddError(ctx.Err())
						}
					})
				}
				callbackName := "pause_product_media_binding_retirement_race_" + uuid.NewString()
				if firstOperation == "bind" {
					require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == "product_media" {
							pauseAfterFirstOperationAcquiresAssetLock(tx)
						}
					}))
					defer db.Callback().Create().Remove(callbackName)
				} else {
					require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == "media_assets" {
							pauseAfterFirstOperationAcquiresAssetLock(tx)
						}
					}))
					defer db.Callback().Delete().Remove(callbackName)
				}
				bindResult, retirementResult := make(chan error, 1), make(chan error, 1)
				bind := func() {
					item := &product.Product{Name: "Concurrent product", Slug: "concurrent-product", Locale: "en", Status: "draft"}
					bindResult <- products.CreateWithSpecValuesVariantsOptionValuesAndMedia(item, nil, nil, nil, []product.ProductMedia{placement})
				}
				retire := func() { retirementResult <- assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
				if firstOperation == "bind" {
					go bind()
				} else {
					go retire()
				}
				var blockerPID int
				select {
				case blockerPID = <-firstLocked:
				case <-ctx.Done():
					t.Fatal("first operation failed to reach its asset lock")
				}
				if firstOperation == "bind" {
					go retire()
				} else {
					go bind()
				}
				require.Eventually(t, func() bool {
					var waiting bool
					err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%media_assets%')", blockerPID).Scan(&waiting).Error
					return err == nil && waiting
				}, 4*time.Second, 20*time.Millisecond, "the second operation must wait on the shared asset row")
				release()
				var bindError, retirementError error
				select {
				case bindError = <-bindResult:
				case <-ctx.Done():
					t.Fatal("binding did not finish")
				}
				select {
				case retirementError = <-retirementResult:
				case <-ctx.Done():
					t.Fatal("retirement did not finish")
				}
				var placements, cleanupEvents int64
				require.NoError(t, db.Model(&product.ProductMedia{}).Count(&placements).Error)
				require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&cleanupEvents).Error)
				if firstOperation == "bind" {
					require.NoError(t, bindError)
					require.ErrorIs(t, retirementError, ErrMediaAssetInUse)
					require.EqualValues(t, 1, placements)
					require.Zero(t, cleanupEvents)
					stored, err := assets.repo.FindAssetByID(asset.ID)
					require.NoError(t, err)
					require.False(t, stored.DeletedAt.Valid)
				} else {
					require.NoError(t, retirementError)
					require.ErrorIs(t, bindError, repository.ErrProductMediaReferenceInvalid)
					require.Zero(t, placements)
					require.EqualValues(t, 1, cleanupEvents)
					var registrations int64
					require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Count(&registrations).Error)
					require.Zero(t, registrations, "the next assertion must exercise cleanup after the registration is physically removed")
					// After physical cleanup removes the registration, a URL-only
					// request must not resurrect the retired managed storage path.
					item := &product.Product{Name: "After cleanup", Slug: "after-cleanup", Locale: "en"}
					err := products.CreateWithSpecValuesVariantsOptionValuesAndMedia(item, nil, nil, nil, []product.ProductMedia{placement})
					require.ErrorIs(t, err, repository.ErrProductMediaReferenceInvalid)
				}
			})
		}
	}
}

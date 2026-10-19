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

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLProductBrandLogoCreateAndUpdateSerializeBindingAndMediaRetirement(t *testing.T) {
	for _, mutation := range []string{"create", "update"} {
		for _, firstOperation := range []string{"bind", "retire"} {
			t.Run(mutation+"/"+firstOperation+" first", func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				require.NoError(t, env.db.AutoMigrate(&product.ProductBrand{}))
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				db := env.db.WithContext(ctx)
				assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
				assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				brands := repository.NewProductBrandRepository(db)
				brands.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/racing-brand.png")
				var brandID uint
				if mutation == "update" {
					brand := product.ProductBrand{Name: "Existing", Slug: "existing", LogoURL: "https://external.example.test/previous.png"}
					require.NoError(t, brands.Create(&brand))
					brandID = brand.ID
				}
				firstLocked := make(chan int, 1)
				releaseChannel := make(chan struct{})
				var releaseOnce, pauseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
				defer release()
				pauseAfterLock := func(tx *gorm.DB) {
					pauseOnce.Do(func() {
						pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, firstLocked, releaseChannel)
					})
				}
				callbackName := "pause_brand_logo_binding_retirement_" + uuid.NewString()
				if firstOperation == "bind" {
					callback := func(tx *gorm.DB) {
						if tx.Statement.Table == "product_brands" {
							pauseAfterLock(tx)
						}
					}
					if mutation == "create" {
						require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, callback))
						defer db.Callback().Create().Remove(callbackName)
					} else {
						require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, callback))
						defer db.Callback().Update().Remove(callbackName)
					}
				} else {
					require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == "media_assets" {
							pauseAfterLock(tx)
						}
					}))
					defer db.Callback().Delete().Remove(callbackName)
				}
				bindLogo := func() error {
					brand := product.ProductBrand{ID: brandID, Name: "Binding", Slug: "binding", LogoURL: "/uploads/" + asset.StorageKey}
					if mutation == "create" {
						return brands.Create(&brand)
					}
					return brands.Update(&brand)
				}
				bindResult, retireResult := make(chan error, 1), make(chan error, 1)
				bind := func() { bindResult <- bindLogo() }
				retire := func() { retireResult <- assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
				if firstOperation == "bind" {
					go bind()
				} else {
					go retire()
				}
				var blockerPID int
				select {
				case blockerPID = <-firstLocked:
				case <-ctx.Done():
					t.Fatal("first brand operation did not acquire its media lock")
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
				}, 4*time.Second, 20*time.Millisecond)
				release()
				var bindError, retireError error
				select {
				case bindError = <-bindResult:
				case <-ctx.Done():
					t.Fatal("brand logo binding timed out")
				}
				select {
				case retireError = <-retireResult:
				case <-ctx.Done():
					t.Fatal("brand logo media retirement timed out")
				}
				var cleanupEvents int64
				require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&cleanupEvents).Error)
				if firstOperation == "bind" {
					require.NoError(t, bindError)
					require.ErrorIs(t, retireError, ErrMediaAssetInUse)
					require.Zero(t, cleanupEvents)
				} else {
					require.NoError(t, retireError)
					require.ErrorIs(t, bindError, repository.ErrProductBrandLogoReferenceInvalid)
					require.EqualValues(t, 1, cleanupEvents)
					var event outbox.Event
					require.NoError(t, db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).First(&event).Error)
					require.NoError(t, NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil).Handle(ctx, event))
					require.ErrorIs(t, bindLogo(), repository.ErrProductBrandLogoReferenceInvalid)
					if brandID != 0 {
						stored, err := brands.FindByID(brandID)
						require.NoError(t, err)
						require.Equal(t, "Existing", stored.Name)
						require.Equal(t, "https://external.example.test/previous.png", stored.LogoURL)
					}
				}
			})
		}
	}
}

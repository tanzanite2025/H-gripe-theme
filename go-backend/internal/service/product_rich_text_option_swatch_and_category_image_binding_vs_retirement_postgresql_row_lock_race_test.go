package service

import (
	"context"
	"fmt"
	"strings"
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

func TestPostgreSQLProductRichTextSwatchesAndCategoryImagesSerializeBindingAndRetirement(t *testing.T) {
	for _, field := range []string{"description", "short_description", "video poster", "file link", "product swatch", "template swatch", "category image"} {
		for _, firstOperation := range []string{"bind", "retire"} {
			t.Run(field+"/"+firstOperation+" first", func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				require.NoError(t, env.db.AutoMigrate(&product.Product{}, &product.ProductMedia{}, &product.ProductVariantOptionValue{}, &product.ProductSpecificationTemplate{}, &product.SpecDefinition{}, &product.ProductSpecOptionItem{}, &product.ProductCategory{}))
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				db := env.db.WithContext(ctx)
				assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
				assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				products := repository.NewProductRepository(db)
				products.ConfigureProductMediaTransactionalAssetBindingResolver(assets)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/racing-photo.png")
				bindingTable := "products"
				bindingError := repository.ErrProductMediaReferenceInvalid
				switch field {
				case "product swatch":
					bindingTable = "product_variant_option_values"
				case "template swatch":
					bindingTable = "product_spec_option_items"
				case "category image":
					bindingTable, bindingError = "product_categories", repository.ErrProductCategoryImageReferenceInvalid
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
						if tx.Statement.Table == bindingTable {
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
				bindMediaReference := func() error {
					item := &product.Product{Name: "Concurrent content", Slug: "concurrent-content", Locale: "en", Status: "draft"}
					var options []product.ProductVariantOptionValue
					switch field {
					case "description":
						item.Description = "<img src=\"/uploads/" + asset.StorageKey + "\">"
					case "short_description":
						item.ShortDesc = "<img src=\"/uploads/" + asset.StorageKey + "\">"
					case "video poster":
						item.Description = "<video src=\"https://external.example.test/video.webm\" poster=\"/uploads/" + asset.StorageKey + "\"></video>"
					case "file link":
						item.Description = "<a href=\"/uploads/" + asset.StorageKey + "\">Download</a>"
					case "product swatch":
						options = []product.ProductVariantOptionValue{{SpecDefinitionID: 1, ValueKey: "ruby", Label: "Ruby", SwatchURL: "/uploads/" + asset.StorageKey}}
					case "template swatch":
						return products.CreateProductSpecificationTemplate(&product.ProductSpecificationTemplate{Name: "Concurrent template", Slug: "concurrent-template", SpecDefinitions: []product.SpecDefinition{{Name: "Finish", Slug: "finish", FieldType: "select", OptionItems: []product.ProductSpecOptionItem{{ValueKey: "ruby", DefaultLabel: "Ruby", SwatchURL: "/uploads/" + asset.StorageKey}}}}})
					case "category image":
						return repository.NewProductCategoryRepository(db).Create(&product.ProductCategory{Name: "Concurrent category", Slug: "concurrent-category", ImageMediaAssetID: &asset.ID})
					}
					return products.CreateWithSpecValuesVariantsOptionValuesAndMedia(item, nil, nil, options, nil)
				}
				bind := func() { bindResult <- bindMediaReference() }
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
				require.NoError(t, db.Table(bindingTable).Count(&placements).Error)
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
					require.ErrorIs(t, bindError, bindingError)
					require.Zero(t, placements)
					require.EqualValues(t, 1, cleanupEvents)
					var registrations int64
					require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Count(&registrations).Error)
					require.Zero(t, registrations, "the next assertion must exercise cleanup after the registration is physically removed")
					// After physical cleanup removes the registration, a URL-only
					// request must not resurrect the retired managed storage path.
					err := bindMediaReference()
					require.ErrorIs(t, err, bindingError)
				}
			})
		}
	}
}

func TestPostgreSQLProductRichTextAndGallerySwappedAssetsAcquireOneConsistentLockOrder(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&product.Product{}, &product.ProductMedia{}))
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	db := env.db.WithContext(ctx)
	assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	products := repository.NewProductRepository(db)
	products.ConfigureProductMediaTransactionalAssetBindingResolver(assets)
	low := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/low.png")
	high := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/high.png")
	first := &product.Product{Name: "First", Slug: "first", Locale: "en"}
	second := &product.Product{Name: "Second", Slug: "second", Locale: "en"}
	require.NoError(t, products.Create(first))
	require.NoError(t, products.Create(second))
	first.Description = `<img src="` + low.URL + `">`
	second.Description = `<img src="` + high.URL + `">`
	firstLocked := make(chan int, 1)
	releaseFirst := make(chan struct{})
	var releaseOnce, pauseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	defer release()
	callbackName := "pause_combined_product_media_asset_order_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" || !strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
			return
		}
		pauseOnce.Do(func() {
			asset, ok := tx.Statement.Dest.(*media.MediaAsset)
			if !ok || asset.ID != low.ID {
				tx.AddError(fmt.Errorf("first lock must be the lowest asset ID"))
				return
			}
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
	}))
	defer db.Callback().Query().Remove(callbackName)
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() {
		firstResult <- products.UpdateWithSpecValuesVariantsOptionValuesAndMedia(first, nil, false, nil, false, nil, false, []product.ProductMedia{{URL: high.URL, MediaType: "image"}}, true)
	}()
	var blockerPID int
	select {
	case blockerPID = <-firstLocked:
	case <-ctx.Done():
		t.Fatal("first combined mutation did not acquire its first asset lock")
	}
	go func() {
		secondResult <- products.UpdateWithSpecValuesVariantsOptionValuesAndMedia(second, nil, false, nil, false, nil, false, []product.ProductMedia{{URL: low.URL, MediaType: "image"}}, true)
	}()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%media_assets%')", blockerPID).Scan(&waiting).Error
		return err == nil && waiting
	}, 4*time.Second, 20*time.Millisecond)
	release()
	for _, result := range []chan error{firstResult, secondResult} {
		select {
		case err := <-result:
			require.NoError(t, err)
		case <-ctx.Done():
			t.Fatal("swapped media references caused a lock order deadlock")
		}
	}
}

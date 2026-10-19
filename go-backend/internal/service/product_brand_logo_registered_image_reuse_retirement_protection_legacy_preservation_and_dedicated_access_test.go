package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
)

func createProductBrandLogoRegisteredMediaTestServices(t *testing.T, useTransactionManager bool) (*gorm.DB, *ProductBrandService, *MediaService) {
	t.Helper()
	db := newProductBrandServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}))
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.siteURL = "https://shop.example.test"
	brands := NewProductBrandService(repository.NewProductBrandRepository(db))
	brands.ConfigureMediaService(assets)
	if useTransactionManager {
		manager := repository.NewTxManager(db, nil, repository.NewProductRepository(db), nil, nil, nil)
		manager.ConfigureProductBrandRepository(brands.repo)
		manager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
		brands.ConfigureTxManager(manager)
	}
	return db, brands, assets
}

func TestProductBrandLogoCreateAndUpdateRejectIneligibleMediaWithoutPartialWrites(t *testing.T) {
	for _, useManager := range []bool{false, true} {
		for _, scenario := range []string{"unregistered", "retired", "inactive", "private", "customer", "visitor", "customer submission", "conversation", "video", "ambiguous", "dedicated original", "dedicated derivative", "private path", "unsafe scheme"} {
			t.Run(fmt.Sprintf("manager=%t/%s", useManager, scenario), func(t *testing.T) {
				db, brands, _ := createProductBrandLogoRegisteredMediaTestServices(t, useManager)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/brand-logo.png")
				logoURL := "/uploads/" + asset.StorageKey
				switch scenario {
				case "unregistered":
					logoURL = "/uploads/missing-brand.png"
				case "retired":
					require.NoError(t, db.Delete(asset).Error)
				case "inactive":
					require.NoError(t, db.Model(asset).Update("status", "inactive").Error)
				case "private":
					require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
				case "customer", "visitor":
					require.NoError(t, db.Model(asset).Update("uploader_actor_type", scenario).Error)
				case "customer submission":
					require.NoError(t, db.Model(asset).Update("library_partition", media.MediaLibraryPartitionCustomerSubmissions).Error)
				case "conversation":
					require.NoError(t, db.Model(asset).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
				case "video":
					require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
				case "ambiguous":
					duplicate := *asset
					duplicate.ID = 0
					duplicate.URL = "https://alternate.example.test/uploads/" + asset.StorageKey
					require.NoError(t, db.Create(&duplicate).Error)
				case "dedicated original":
					logoURL = "/uploads/site-logo/historical.png"
					require.NoError(t, db.Model(asset).Updates(map[string]any{"storage_key": "site-logo/historical.png", "url": "https://media.example.test" + logoURL}).Error)
				case "dedicated derivative":
					logoURL = "/uploads/site-favicon/derivative.png"
					require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: asset.ID, StorageKey: "site-favicon/derivative.png", URL: "https://media.example.test" + logoURL}).Error)
				case "private path":
					logoURL = "/uploads/customer-service/attachments/originals/private.png"
				case "unsafe scheme":
					logoURL = "ftp://example.test/brand.png"
				}
				input := ProductBrandInput{Name: "Invalid", Slug: "invalid", LogoURL: logoURL, IsEnabled: true}
				_, err := brands.Create(input)
				require.ErrorIs(t, err, ErrProductBrandLogoReferenceInvalid)
				var count int64
				require.NoError(t, db.Model(&product.ProductBrand{}).Count(&count).Error)
				require.Zero(t, count)
				original := product.ProductBrand{Name: "Original", Slug: "original", LogoURL: "https://external.example.test/old.png", IsEnabled: true}
				require.NoError(t, db.Create(&original).Error)
				_, err = brands.Update(original.ID, input)
				require.ErrorIs(t, err, ErrProductBrandLogoReferenceInvalid)
				stored, err := brands.Get(original.ID)
				require.NoError(t, err)
				require.Equal(t, original.Name, stored.Name)
				require.Equal(t, original.LogoURL, stored.LogoURL)
				require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
				require.Zero(t, count)
			})
		}
	}
}

func TestProductBrandLogoCanonicalReuseRetainsSourceAndProtectsFilesUntilDetachment(t *testing.T) {
	for _, useManager := range []bool{false, true} {
		t.Run(fmt.Sprintf("manager=%t", useManager), func(t *testing.T) {
			db, brands, assets := createProductBrandLogoRegisteredMediaTestServices(t, useManager)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/page-media/original-brand.png")
			require.NoError(t, db.Model(asset).Updates(map[string]any{"upload_origin": media.MediaUploadOriginAdministratorContent, "library_partition": media.MediaLibraryPartitionContent}).Error)
			derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, URL: "https://media.example.test/uploads/derivatives/brand-small.webp", StorageKey: "derivatives/brand-small.webp"}
			require.NoError(t, db.Create(&derivative).Error)
			brand, err := brands.Create(ProductBrandInput{Name: "Brand", Slug: "brand", LogoURL: "/uploads/derivatives/brand-small.webp#logo", IsEnabled: false})
			require.NoError(t, err)
			require.Equal(t, asset.URL+"#logo", brand.LogoURL)
			report, err := assets.GetAssetReferences(asset.ID)
			require.NoError(t, err)
			require.Equal(t, 1, report.Total, "disabled brands retain logo bytes")
			require.Equal(t, "product_brand_logo", report.References[0].ResourceType)
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			var storedAsset media.MediaAsset
			require.NoError(t, db.First(&storedAsset, asset.ID).Error)
			require.Equal(t, media.MediaUploadOriginAdministratorContent, storedAsset.UploadOrigin)
			require.Equal(t, media.MediaLibraryPartitionContent, storedAsset.LibraryPartition)
			_, err = brands.Update(brand.ID, ProductBrandInput{Name: brand.Name, Slug: brand.Slug, LogoURL: ""})
			require.NoError(t, err)
			require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
		})
	}
}

func TestProductBrandLogoReferenceScanningPagesExactAliasesAndHistoricalDerivatives(t *testing.T) {
	db, _, assets := createProductBrandLogoRegisteredMediaTestServices(t, false)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/exact-brand.png")
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, URL: "https://media.example.test/uploads/derivatives/brand.webp", StorageKey: "derivatives/brand.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	var brands []product.ProductBrand
	for index := 0; index < 260; index++ {
		brands = append(brands, product.ProductBrand{Name: "Unrelated", Slug: fmt.Sprintf("unrelated-%d", index), LogoURL: "https://foreign.example.test/uploads/" + asset.StorageKey})
	}
	require.NoError(t, db.Create(&brands).Error)
	for index, reference := range []string{"https://shop.example.test/uploads/" + asset.StorageKey + "?version=old", "/uploads/derivatives/brand.webp#small", "/uploads/" + asset.StorageKey + ".backup"} {
		require.NoError(t, db.Create(&product.ProductBrand{Name: "Historical", Slug: fmt.Sprintf("historical-%d", index), LogoURL: reference}).Error)
	}
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	for _, reference := range report.References {
		require.Equal(t, "product_brand_logo", reference.ResourceType)
	}
}

func TestProductBrandHistoricalDedicatedLogoRetentionAndPublicAccessFollowEnabledOwner(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		t.Run(namespace, func(t *testing.T) {
			db, provider, cleanup, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
			require.NoError(t, db.AutoMigrate(&product.ProductBrand{}))
			brand := product.ProductBrand{Name: "Historical brand", Slug: "historical-brand", LogoURL: "/uploads/" + strings.ReplaceAll(key, " ", "%20"), IsEnabled: true}
			require.NoError(t, db.Create(&brand).Error)
			access := NewPublicUploadAccessService(nil, nil)
			access.ConfigureSiteLogoService(cleanup.siteLogo)
			access.ConfigureSiteFaviconService(cleanup.siteFavicon)
			access.ConfigureHomeVisualTileService(cleanup.homeVisual)
			for _, enabled := range []bool{true, false, true} {
				require.NoError(t, db.Model(&brand).Update("is_enabled", enabled).Error)
				allowed, err := access.CanServePublicUpload(t.Context(), key)
				require.NoError(t, err)
				require.Equal(t, enabled, allowed)
				require.ErrorIs(t, cleanup.Handle(t.Context(), event), ErrMediaAssetInUse)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
			}
			brands := repository.NewProductBrandRepository(db)
			stale := brand
			brand.LogoURL = ""
			require.NoError(t, brands.Update(&brand))
			require.ErrorIs(t, brands.Update(&stale), repository.ErrProductBrandLogoReferenceInvalid)
			allowed, err := access.CanServePublicUpload(t.Context(), key)
			require.NoError(t, err)
			require.False(t, allowed)
			require.NoError(t, cleanup.Handle(t.Context(), event))
		})
	}
}

func TestProductBrandLogoOmittedRepositoryUpdateReloadsCurrentRowAndCannotRecreateDeletedBrand(t *testing.T) {
	db, brands, _ := createProductBrandLogoRegisteredMediaTestServices(t, false)
	brand := product.ProductBrand{Name: "Old", Slug: "old", LogoURL: "/uploads/legacy-logo.png"}
	require.NoError(t, db.Create(&brand).Error)
	stale := brand
	require.NoError(t, db.Model(&brand).Update("logo_url", "/uploads/current-legacy-logo.png").Error)
	stale.Name = "Metadata edit"
	require.NoError(t, brands.repo.Update(&stale, repository.ProductBrandLogoFieldMutation{UpdateLogo: false}))
	require.Equal(t, "/uploads/current-legacy-logo.png", stale.LogoURL)
	require.NoError(t, db.Delete(&brand).Error)
	require.ErrorIs(t, brands.repo.Update(&stale), gorm.ErrRecordNotFound)
}

func TestProductBrandLogoUpdateRollsBackWithDependentOutboxFailure(t *testing.T) {
	db, brands, _ := createProductBrandLogoRegisteredMediaTestServices(t, true)
	brand, err := brands.Create(ProductBrandInput{Name: "Original", Slug: "original", LogoURL: "https://external.example.test/previous.png", IsEnabled: true})
	require.NoError(t, err)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/replacement-brand.png")
	forcedOutboxError := errors.New("forced product brand dependent event failure")
	callbackName := "reject_product_brand_dependent_outbox_write"
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			tx.AddError(forcedOutboxError)
		}
	}))
	defer db.Callback().Create().Remove(callbackName)
	_, err = brands.Update(brand.ID, ProductBrandInput{Name: "Replacement", Slug: brand.Slug, LogoURL: "/uploads/" + asset.StorageKey, IsEnabled: true})
	require.ErrorIs(t, err, forcedOutboxError)
	stored, err := brands.Get(brand.ID)
	require.NoError(t, err)
	require.Equal(t, brand.Name, stored.Name)
	require.Equal(t, brand.LogoURL, stored.LogoURL)
	var events int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
	require.Zero(t, events)
}

package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
)

func createProductMediaTransactionalBindingTestServices(t *testing.T, useApplicationTransactionManager bool) (*gorm.DB, *ProductService, *MediaService) {
	t.Helper()
	db, products := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.MediaDerivativePreset{}))
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.siteURL = "https://shop.example.test"
	assets.ConfigureDerivativePresetRepository(repository.NewMediaDerivativePresetRepository(db))
	require.NoError(t, db.Create(&media.MediaDerivativePreset{Code: "thumbnail", Label: "Thumbnail", MaxWidth: 160, Enabled: true, GenerationVersion: 2}).Error)
	products.ConfigureMediaService(assets)
	if useApplicationTransactionManager {
		manager := repository.NewTxManager(db, repository.NewOrderRepository(db), products.productRepo, repository.NewCouponRepository(db), repository.NewLoyaltyRepository(db), repository.NewPaymentRepository(db))
		manager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
		products.ConfigureTxManager(manager)
	}
	return db, products, assets
}

func createRegisteredPublicImageForProductMediaBindingTest(t *testing.T, db *gorm.DB, key string) *media.MediaAsset {
	t.Helper()
	asset := &media.MediaAsset{Filename: "photo.png", StorageKey: key, URL: "https://media.example.test/uploads/" + key, MediaType: "image", Visibility: "public", Status: "active"}
	require.NoError(t, db.Create(asset).Error)
	return asset
}

func createProductInputForTransactionalMediaBindingTest(items ...ProductMediaInput) ProductCreateInput {
	return ProductCreateInput{MerchantProductIdentifier: "TEST-MERCHANT-PRODUCT", ManufacturerProductIdentifier: "TEST-MANUFACTURER-PRODUCT", Name: "Media binding test", Slug: "media-binding-test", Locale: "en", Status: "draft", Media: items,
		Variants: []ProductVariantInput{{SKU: "MEDIA-BINDING-TEST", PriceMinor: 399, Stock: 5, IsDefault: true, IsActive: boolPtr(true)}},
	}
}

func TestProductMediaTransactionalBindingRejectsIneligibleAssetsAndMismatchedIdentityAtomically(t *testing.T) {
	for _, useManager := range []bool{false, true} {
		for _, scenario := range []string{"different ID and URL", "foreign host same upload path", "missing asset", "retired asset", "inactive asset", "private asset", "conversation origin", "legacy conversation Windows key", "legacy conversation cloud URL", "customer submission", "customer actor", "visitor actor", "type mismatch", "private thumbnail", "private poster", "video poster", "ambiguous storage key", "retired URL only", "unregistered managed URL", "unregistered private namespace"} {
			t.Run(fmt.Sprintf("application manager=%t/%s", useManager, scenario), func(t *testing.T) {
				db, products, _ := createProductMediaTransactionalBindingTestServices(t, useManager)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/product-images/photo.png")
				item := ProductMediaInput{MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}
				switch scenario {
				case "different ID and URL":
					item.URL = createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/page-media/photo.png").URL
				case "foreign host same upload path":
					item.URL = "https://foreign.example.test/uploads/" + asset.StorageKey
				case "missing asset":
					missingID := uint(99999)
					item.MediaAssetID = &missingID
				case "retired asset", "retired URL only":
					require.NoError(t, db.Delete(asset).Error)
					if scenario == "retired URL only" {
						item.MediaAssetID = nil
					}
				case "inactive asset":
					require.NoError(t, db.Model(asset).Update("status", "inactive").Error)
				case "private asset":
					require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
				case "conversation origin":
					require.NoError(t, db.Model(asset).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
				case "legacy conversation Windows key":
					require.NoError(t, db.Model(asset).Update("storage_key", `customer-service\attachments\originals\photo.png`).Error)
				case "legacy conversation cloud URL":
					item.URL = "https://legacy-cloud.example.test/customer-service/public-conversation-attachments/photo.png"
					require.NoError(t, db.Model(asset).Updates(map[string]any{"storage_key": "", "url": item.URL}).Error)
				case "customer submission":
					require.NoError(t, db.Model(asset).Update("library_partition", media.MediaLibraryPartitionCustomerSubmissions).Error)
				case "customer actor":
					require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
				case "visitor actor":
					require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorVisitor).Error)
				case "type mismatch":
					item.MediaType = "video"
				case "private thumbnail", "private poster":
					private := createRegisteredPublicImageForProductMediaBindingTest(t, db, "another/private.png")
					require.NoError(t, db.Model(private).Update("visibility", "private").Error)
					if scenario == "private thumbnail" {
						item.ThumbnailURL = "/uploads/" + private.StorageKey
					} else {
						item.PosterURL = private.URL
					}
				case "video poster":
					video := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/video.webm")
					require.NoError(t, db.Model(video).Update("media_type", "video").Error)
					item.PosterURL = video.URL
				case "ambiguous storage key":
					duplicate := media.MediaAsset{Filename: "other.png", StorageKey: asset.StorageKey, URL: "https://legacy.example.test/same-key.png"}
					require.NoError(t, db.Create(&duplicate).Error)
					item.URL = "/uploads/" + asset.StorageKey
				case "unregistered managed URL":
					item.MediaAssetID, item.URL = nil, "/uploads/deleted-or-unregistered.png"
				case "unregistered private namespace":
					item.MediaAssetID, item.URL = nil, "/uploads/customer-service/attachments/originals/not-registered.png"
				}
				_, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(item))
				require.ErrorIs(t, err, ErrProductMediaInvalid)
				for _, model := range []any{&product.Product{}, &product.ProductMedia{}, &outbox.Event{}} {
					var count int64
					require.NoError(t, db.Model(model).Count(&count).Error)
					require.Zero(t, count, "a rejected binding must roll back all product-owned writes and events")
				}
			})
		}
	}
}

func TestProductMediaTransactionalBindingResolvesURLOnlyAndCurrentDerivativesWithoutChangingProvenance(t *testing.T) {
	for _, useManager := range []bool{false, true} {
		t.Run(fmt.Sprintf("application manager=%t", useManager), func(t *testing.T) {
			db, products, assets := createProductMediaTransactionalBindingTestServices(t, useManager)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/product-images/photo.png")
			other := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/page-media/photo.png")
			derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", PresetVersion: 2, URL: "https://media.example.test/uploads/derivatives/thumbnail.webp", StorageKey: "derivatives/thumbnail.webp", MimeType: "image/webp", Width: 160, Height: 80}
			require.NoError(t, db.Create(&derivative).Error)
			created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(
				ProductMediaInput{URL: "/uploads/" + asset.StorageKey + "?v=3", MediaType: "image"},
				ProductMediaInput{URL: "https://shop.example.test/uploads/" + other.StorageKey, MediaType: "image"},
				ProductMediaInput{URL: "https://foreign.example.test/uploads/" + asset.StorageKey, MediaType: "image"},
				ProductMediaInput{URL: "/images/legacy-static.png", MediaType: "image"},
			))
			require.NoError(t, err)
			require.Len(t, created.Media, 4)
			require.Equal(t, asset.ID, *created.Media[0].MediaAssetID)
			require.Equal(t, asset.URL, created.Media[0].URL)
			require.Equal(t, derivative.URL, created.Media[0].ThumbnailURL)
			require.Equal(t, 160, product.ParseProductMediaImageVariants(created.Media[0].ImageVariantData)["thumbnail"].Width)
			require.Equal(t, other.ID, *created.Media[1].MediaAssetID)
			for _, item := range created.Media[2:] {
				require.Nil(t, item.MediaAssetID)
			}
			stored, err := assets.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			require.Equal(t, media.MediaUploadOriginLegacyUnknown, stored.UploadOrigin)
			require.Equal(t, media.MediaLibraryPartitionUnclassified, stored.LibraryPartition)
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
		})
	}
}

func TestProductMediaAuxiliaryImagesUseStableOriginalIdentityAndUnbindingKeepsFiles(t *testing.T) {
	db, products, assets := createProductMediaTransactionalBindingTestServices(t, false)
	video := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/product-videos/video.webm")
	require.NoError(t, db.Model(video).Update("media_type", "video").Error)
	poster := createRegisteredPublicImageForProductMediaBindingTest(t, db, "shared/poster.png")
	derivative := media.MediaAssetDerivative{MediaAssetID: poster.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/poster.webp", StorageKey: "derivatives/poster.webp", MimeType: "image/webp"}
	require.NoError(t, db.Create(&derivative).Error)
	created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(ProductMediaInput{URL: video.URL, MediaType: "video", PosterURL: "/uploads/derivatives/poster.webp", ThumbnailURL: "/uploads/" + poster.StorageKey}))
	require.NoError(t, err)
	require.Equal(t, poster.URL, created.Media[0].PosterURL)
	require.Equal(t, poster.URL, created.Media[0].ThumbnailURL)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), poster.ID, MediaAssetDeleteConfirmation(poster.ID)), ErrMediaAssetInUse)
	updated, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true, Media: []ProductMediaInput{}})
	require.NoError(t, err)
	require.Empty(t, updated.Media)
	var activeAssets, cleanupEvents int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&activeAssets).Error)
	require.EqualValues(t, 2, activeAssets)
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Count(&cleanupEvents).Error)
	require.Zero(t, cleanupEvents)
}

func TestProductMediaOrdinaryEditsPreserveExistingBindingsWhileReplacementAndTranslationRevalidate(t *testing.T) {
	for _, useManager := range []bool{false, true} {
		t.Run(fmt.Sprintf("application manager=%t", useManager), func(t *testing.T) {
			db, products, _ := createProductMediaTransactionalBindingTestServices(t, useManager)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/product-images/photo.png")
			created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(ProductMediaInput{URL: asset.URL, MediaType: "image"}))
			require.NoError(t, err)
			translated, _, err := products.CopyAdminProductTranslation(created.ID, "fr")
			require.NoError(t, err)
			require.Equal(t, asset.ID, *translated.Media[0].MediaAssetID)
			require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
			name := "Ordinary edit remains available"
			updated, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &name})
			require.NoError(t, err)
			require.Equal(t, asset.URL, updated.Media[0].URL)
			rejectedName := "This must roll back"
			_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &rejectedName, UpdateMedia: true, Media: []ProductMediaInput{{MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}}})
			require.ErrorIs(t, err, ErrProductMediaInvalid)
			_, _, err = products.CopyAdminProductTranslation(created.ID, "de")
			require.ErrorIs(t, err, ErrProductMediaInvalid)
			stored, err := products.GetAdminProduct(created.ID)
			require.NoError(t, err)
			require.Equal(t, name, stored.Name)
			var count int64
			require.NoError(t, db.Model(&product.Product{}).Count(&count).Error)
			require.EqualValues(t, 2, count)
			require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
			require.EqualValues(t, 1, count)
		})
	}
}

func TestProductMediaExplicitAssetIDCannotBypassValidationWithoutMediaServiceOrMediaTable(t *testing.T) {
	_, products := newTestProductService(t)
	assetID := uint(77)
	_, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(ProductMediaInput{MediaAssetID: &assetID, URL: "/uploads/legacy.png", MediaType: "image"}))
	require.ErrorIs(t, err, ErrProductMediaInvalid)
}

func TestProductMediaUnregisteredHistoricalUploadURLCanOnlyRemainOnItsOriginalPlacement(t *testing.T) {
	db, products, _ := createProductMediaTransactionalBindingTestServices(t, false)
	created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest())
	require.NoError(t, err)
	legacy := product.ProductMedia{ProductID: created.ID, URL: "/uploads/legacy-unregistered.png", MediaType: "image"}
	require.NoError(t, db.Create(&legacy).Error)
	updated, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true, Media: []ProductMediaInput{{ID: &legacy.ID, URL: legacy.URL, MediaType: "image", Alt: "Preserved legacy placement"}}})
	require.NoError(t, err)
	require.Equal(t, legacy.URL, updated.Media[0].URL)
	require.Nil(t, updated.Media[0].MediaAssetID)
	for _, replacement := range []ProductMediaInput{
		{URL: legacy.URL, MediaType: "image"},
		{ID: &legacy.ID, URL: "/uploads/different-unregistered.png", MediaType: "image"},
		{ID: &legacy.ID, URL: legacy.URL, ThumbnailURL: legacy.URL, MediaType: "image"},
	} {
		_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true, Media: []ProductMediaInput{replacement}})
		require.ErrorIs(t, err, ErrProductMediaInvalid)
	}
	_, _, err = products.CopyAdminProductTranslation(created.ID, "fr")
	require.ErrorIs(t, err, ErrProductMediaInvalid)
	_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true})
	require.NoError(t, err)
}

func TestProductMediaHistoricalUnconfiguredPortAliasCanOnlyRemainWithItsOriginalPlacementAndAssetID(t *testing.T) {
	db, products, _ := createProductMediaTransactionalBindingTestServices(t, false)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/photo.png")
	created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(ProductMediaInput{MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}))
	require.NoError(t, err)
	legacyURL := "https://media.example.test:9443/uploads/" + asset.StorageKey
	placementID := created.Media[0].ID
	require.NoError(t, db.Model(&product.ProductMedia{}).Where("id = ?", placementID).Update("url", legacyURL).Error)
	updated, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true, Media: []ProductMediaInput{{ID: &placementID, MediaAssetID: &asset.ID, URL: legacyURL, MediaType: "image", Alt: "Keep the historical address"}}})
	require.NoError(t, err)
	require.Equal(t, legacyURL, updated.Media[0].URL)
	require.Equal(t, asset.ID, *updated.Media[0].MediaAssetID)
	for _, replacement := range []ProductMediaInput{
		{MediaAssetID: &asset.ID, URL: legacyURL, MediaType: "image"},
		{ID: &placementID, MediaAssetID: &asset.ID, URL: "https://different.example.test/uploads/" + asset.StorageKey, MediaType: "image"},
	} {
		_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true, Media: []ProductMediaInput{replacement}})
		require.ErrorIs(t, err, ErrProductMediaInvalid)
	}
	_, _, err = products.CopyAdminProductTranslation(created.ID, "fr")
	require.ErrorIs(t, err, ErrProductMediaInvalid)
}

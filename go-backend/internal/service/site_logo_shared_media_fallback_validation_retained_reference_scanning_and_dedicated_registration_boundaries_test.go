package service

import (
	"fmt"
	"testing"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSiteLogoSharedFallbackAllSettingWriteContractsValidateCanonicalizeAndProtectWithoutCopyingAssets(t *testing.T) {
	for _, entry := range []string{"single", "batch", "transaction", "import"} {
		t.Run(entry, func(t *testing.T) {
			db, settings, _, assets := createWebsiteProfileRegisteredMediaBindingTestServices(t)
			image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/shared-logo.png")
			derivative := media.MediaAssetDerivative{MediaAssetID: image.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/logo.webp", StorageKey: "derivatives/logo.webp"}
			require.NoError(t, db.Create(&derivative).Error)
			write := func(reference string) error {
				row := setting.Setting{Key: setting.SiteLogoSettingKey, Value: reference, Group: "site", Locale: "en", IsPublic: true}
				switch entry {
				case "single":
					return settings.Set(row.Key, row.Value, "string", row.Group, row.Locale)
				case "transaction":
					return db.Transaction(func(tx *gorm.DB) error { return settings.BatchSetTx(tx, []setting.Setting{row}) })
				case "import":
					return settings.settingRepo.SetMultipleValuesPreservingExistingVisibilityAndDescription([]setting.Setting{row})
				default:
					return settings.BatchSet([]setting.Setting{row})
				}
			}
			for _, reference := range []string{"/uploads/" + derivative.StorageKey, "https://shop.example.test/uploads/" + image.StorageKey} {
				require.NoError(t, write(reference))
				stored, err := settings.Get(setting.SiteLogoSettingKey, "en")
				require.NoError(t, err)
				require.Equal(t, image.URL, stored.Value)
			}
			report, err := assets.GetAssetReferences(image.ID)
			require.NoError(t, err)
			require.Equal(t, 1, report.Total)
			require.Equal(t, "site_identity_setting", report.References[0].ResourceType)
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
			var count int64
			require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
			require.EqualValues(t, 1, count)
			require.NoError(t, write(""))
			require.NoError(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)))
		})
	}
}

func TestSiteLogoSharedFallbackRejectsInvalidManagedAssetsAndRollsBackOtherSettings(t *testing.T) {
	for _, scenario := range []string{"missing", "private", "inactive", "customer", "conversation", "retired", "wrong type", "dedicated"} {
		t.Run(scenario, func(t *testing.T) {
			db, settings, _, _ := createWebsiteProfileRegisteredMediaBindingTestServices(t)
			image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/rejected-logo.png")
			switch scenario {
			case "missing":
				require.NoError(t, db.Unscoped().Delete(image).Error)
			case "private":
				require.NoError(t, db.Model(image).Update("visibility", "private").Error)
			case "inactive":
				require.NoError(t, db.Model(image).Update("status", "inactive").Error)
			case "customer":
				require.NoError(t, db.Model(image).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
			case "conversation":
				require.NoError(t, db.Model(image).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
			case "retired":
				require.NoError(t, db.Delete(image).Error)
			case "wrong type":
				require.NoError(t, db.Model(image).Update("media_type", "video").Error)
			case "dedicated":
				image.StorageKey = "site-logo/duplicate.png"
				require.NoError(t, db.Model(image).Update("storage_key", image.StorageKey).Error)
			}
			err := settings.BatchSet([]setting.Setting{{Key: "site_title", Value: "Must roll back", Locale: "en", Group: "site"}, {Key: setting.SiteLogoSettingKey, Value: "/uploads/" + image.StorageKey, Locale: "en", Group: "site"}})
			require.ErrorIs(t, err, ErrSettingInvalid)
			var count int64
			require.NoError(t, db.Unscoped().Model(&setting.Setting{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestSiteIdentitySettingReferenceScanningIncludesPaginationHiddenDeletedDerivativesAndOnlyTrustedExactAliases(t *testing.T) {
	db, _, _, assets := createWebsiteProfileRegisteredMediaBindingTestServices(t)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/retained-logo.png")
	derivative := media.MediaAssetDerivative{MediaAssetID: image.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/retained-logo.webp", StorageKey: "derivatives/retained-logo.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	for index := 0; index < 257; index++ {
		require.NoError(t, db.Create(&setting.Setting{Key: setting.SiteLogoSettingKey, Locale: fmt.Sprintf("unused-%d", index), Value: "/static/unrelated.png"}).Error)
	}
	for index, value := range []string{image.URL, "/uploads/" + derivative.StorageKey, "https://shop.example.test/uploads/" + image.StorageKey + "?cache=1", "https://foreign.example.test/uploads/" + image.StorageKey, "https://shop.example.test:8443/uploads/" + image.StorageKey, "/uploads/another/retained-logo.png"} {
		row := setting.Setting{Key: setting.SiteLogoSettingKey, Locale: fmt.Sprintf("target-%d", index), Value: value}
		if index == 1 {
			row.Key = setting.SiteFaviconSettingKey
		}
		require.NoError(t, db.Create(&row).Error)
		if index == 0 {
			require.NoError(t, db.Model(&row).Update("is_public", false).Error)
		}
		if index == 1 {
			require.NoError(t, db.Delete(&row).Error)
		}
	}
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Equal(t, 3, report.Total)
	retained := 0
	for _, reference := range report.References {
		if reference.UsageState == "retained_deleted_parent" {
			retained++
		}
	}
	require.Equal(t, 1, retained)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
}

func TestCurrentDedicatedOwnersAndDraftShowcaseProtectConflictingSharedRegistrations(t *testing.T) {
	db, _, _, assets := createWebsiteProfileRegisteredMediaBindingTestServices(t)
	require.NoError(t, db.AutoMigrate(&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &homevisualtile.Tile{}))
	logo := createRegisteredPublicImageForProductMediaBindingTest(t, db, "site-logo/collision.webp")
	favicon := createRegisteredPublicImageForProductMediaBindingTest(t, db, "site-favicon/collision.webp")
	thumbnail := createRegisteredPublicImageForProductMediaBindingTest(t, db, "visual-showcase/thumbnail.webp")
	require.NoError(t, db.Create(&sitelogodomain.Asset{ID: 1, StorageKey: logo.StorageKey, URL: "https://other-store.example.test/logo.webp"}).Error)
	require.NoError(t, db.Create(&sitefavicondomain.Asset{ID: 1, StorageKey: favicon.StorageKey}).Error)
	tile := homevisualtile.Tile{TileSetKey: "home-hero", Locale: "en", ImageURL: "/uploads/visual-showcase/original.webp", ThumbnailURL: "https://shop.example.test/uploads/" + thumbnail.StorageKey}
	require.NoError(t, db.Create(&tile).Error)
	require.NoError(t, db.Model(&tile).Update("is_published", false).Error)
	for _, image := range []*media.MediaAsset{logo, favicon, thumbnail} {
		report, err := assets.GetAssetReferences(image.ID)
		require.NoError(t, err)
		require.Equal(t, 1, report.Total)
		require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
	}
	require.NoError(t, db.Delete(&tile).Error)
	report, err := assets.GetAssetReferences(thumbnail.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total, "deleted historical tile versions do not become permanent references")
}

func TestDedicatedDuplicateRegistrationsCannotCreateProductCategoryOrProductMediaPlacements(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		t.Run(namespace, func(t *testing.T) {
			db, products, assets := createProductMediaTransactionalBindingTestServices(t, false)
			require.NoError(t, db.AutoMigrate(&product.ProductCategory{}))
			image := createRegisteredPublicImageForProductMediaBindingTest(t, db, namespace+"/duplicate.webp")
			_, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(ProductMediaInput{MediaAssetID: &image.ID, URL: image.URL, MediaType: "image"}))
			require.Error(t, err)
			categories := repository.NewProductCategoryRepository(db)
			categories.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
			category := product.ProductCategory{Name: "New", Slug: "new", ImageMediaAssetID: &image.ID}
			require.ErrorIs(t, categories.Create(&category), repository.ErrProductCategoryImageReferenceInvalid)
			urlOnly := createRegisteredPublicImageForProductMediaBindingTest(t, db, namespace+"/url-only.webp")
			require.NoError(t, db.Model(urlOnly).Update("storage_key", "").Error)
			require.ErrorIs(t, categories.Create(&product.ProductCategory{Name: "URL only", Slug: "url-only", ImageMediaAssetID: &urlOnly.ID}), repository.ErrProductCategoryImageReferenceInvalid)
			foreignURLOnly := createRegisteredPublicImageForProductMediaBindingTest(t, db, namespace+"/foreign.webp")
			require.NoError(t, db.Model(foreignURLOnly).Updates(map[string]interface{}{"storage_key": "", "url": "https://foreign.example.test/uploads/" + foreignURLOnly.StorageKey}).Error)
			require.NoError(t, categories.Create(&product.ProductCategory{Name: "Foreign", Slug: "foreign", ImageMediaAssetID: &foreignURLOnly.ID}), "a foreign origin is not identified by its path suffix")
			// Preserve existing placements exactly when unrelated metadata changes.
			category.ID, category.ImageURL = 0, "/uploads/"+image.StorageKey
			require.NoError(t, db.Create(&category).Error)
			category.Name = "Edited"
			require.NoError(t, categories.Update(&category, nil))
			require.Equal(t, "/uploads/"+image.StorageKey, category.ImageURL)
			ordinary := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/original.webp")
			created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest(ProductMediaInput{URL: ordinary.URL, MediaType: "image"}))
			require.NoError(t, err)
			placementID := created.Media[0].ID
			legacyURL := "/uploads/" + image.StorageKey
			require.NoError(t, db.Model(&product.ProductMedia{}).Where("id = ?", placementID).Updates(map[string]interface{}{"url": legacyURL, "media_asset_id": image.ID, "thumbnail_url": legacyURL}).Error)
			updated, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateMedia: true, Media: []ProductMediaInput{{ID: &placementID, MediaAssetID: &image.ID, URL: legacyURL, ThumbnailURL: legacyURL, MediaType: "image", Alt: "Updated metadata"}}})
			require.NoError(t, err)
			require.Equal(t, legacyURL, updated.Media[0].URL)
			require.Equal(t, legacyURL, updated.Media[0].ThumbnailURL)
			_, _, err = products.CopyAdminProductTranslation(created.ID, "fr")
			require.ErrorIs(t, err, ErrProductMediaInvalid)
		})
	}
}

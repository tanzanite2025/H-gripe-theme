package service

import (
	"errors"
	"fmt"
	"testing"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createGalleryTransactionalMediaBindingTestServices(t *testing.T) (*gorm.DB, *GalleryService, *MediaService) {
	t.Helper()
	db, galleries := newTestGalleryService(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAssetDerivative{}))
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.siteURL = "https://shop.example.test"
	galleries.ConfigureMediaService(assets)
	return db, galleries, assets
}

func TestGalleryCanonicalCoverOriginalAndIndependentThumbnailReuseWithoutCopiesOrStaleAssociationWrites(t *testing.T) {
	db, galleries, assets := createGalleryTransactionalMediaBindingTestServices(t)
	cover := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/gallery-cover.png")
	original := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/gallery-original.png")
	preview := createRegisteredPublicImageForProductMediaBindingTest(t, db, "shared/gallery-preview.png")
	require.NoError(t, db.Model(original).Update("library_partition", media.MediaLibraryPartitionCatalog).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: original.ID, Preset: "thumbnail", StorageKey: "derivatives/gallery.webp", URL: "https://media.example.test/uploads/derivatives/gallery.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	item := &gallery.Gallery{Name: "Reused", Slug: "reused", CoverImage: "/uploads/" + cover.StorageKey, Images: []gallery.GalleryImage{{MediaAssetID: &original.ID, URL: "https://shop.example.test/uploads/derivatives/gallery.webp?v=1", Thumbnail: "/uploads/" + preview.StorageKey}}}
	require.NoError(t, galleries.CreateGallery(item))
	require.Equal(t, cover.URL, item.CoverImage)
	require.Len(t, item.Images, 1)
	require.Equal(t, original.URL, item.Images[0].URL)
	require.Equal(t, preview.URL, item.Images[0].Thumbnail)
	for _, asset := range []*media.MediaAsset{cover, original, preview} {
		require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	}
	stale, err := galleries.GetGalleryByID(item.ID)
	require.NoError(t, err)
	current := *stale
	current.CoverImage = ""
	require.NoError(t, galleries.UpdateGallery(&current))
	image := stale.Images[0]
	image.MediaAssetID = nil
	image.URL = "https://external.example.test/replacement.png"
	image.Thumbnail = ""
	require.NoError(t, galleries.UpdateGalleryImage(&image))
	for _, asset := range []*media.MediaAsset{cover, original, preview} {
		require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	}
	stale.Name = "Ordinary edit after unbinding"
	require.NoError(t, galleries.repo.UpdateGallery(stale, repository.GalleryMediaFieldMutation{}))
	stored, err := galleries.GetGalleryByID(item.ID)
	require.NoError(t, err)
	require.Empty(t, stored.CoverImage)
	require.Equal(t, image.URL, stored.Images[0].URL)
	require.Nil(t, stored.Images[0].MediaAssetID)
	require.NoError(t, galleries.DeleteGalleryImage(image.ID))
	require.ErrorIs(t, galleries.UpdateGalleryImage(&image), gorm.ErrRecordNotFound)
	require.NoError(t, galleries.DeleteGallery(item.ID))
	require.ErrorIs(t, galleries.UpdateGallery(stale), gorm.ErrRecordNotFound)
	var count int64
	require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestGalleryNewCoverOriginalAndThumbnailRejectIneligibleOrUnregisteredMediaAtomically(t *testing.T) {
	for _, field := range []string{"cover", "original", "thumbnail"} {
		for _, scenario := range []string{"missing", "private", "customer", "conversation", "retired", "inactive", "video", "ambiguous", "mismatched ID"} {
			if scenario == "mismatched ID" && field != "original" {
				continue
			}
			t.Run(field+"/"+scenario, func(t *testing.T) {
				db, galleries, _ := createGalleryTransactionalMediaBindingTestServices(t)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/rejected-gallery.png")
				reference := "/uploads/" + asset.StorageKey
				switch scenario {
				case "missing":
					reference = "/uploads/content/not-registered.png"
				case "private":
					require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
				case "customer":
					require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
				case "conversation":
					require.NoError(t, db.Model(asset).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
				case "retired":
					require.NoError(t, db.Delete(asset).Error)
				case "inactive":
					require.NoError(t, db.Model(asset).Update("status", "inactive").Error)
				case "video":
					require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
				case "ambiguous":
					duplicate := *asset
					duplicate.ID = 0
					duplicate.URL = "https://old.example.test/same-gallery-file.png"
					require.NoError(t, db.Create(&duplicate).Error)
				case "mismatched ID":
					other := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/other-gallery.png")
					reference = other.URL
				}
				item := &gallery.Gallery{Name: "Rejected", Slug: "rejected"}
				switch field {
				case "cover":
					item.CoverImage = reference
				case "original":
					item.Images = []gallery.GalleryImage{{URL: reference}}
					if scenario == "mismatched ID" {
						item.Images[0].MediaAssetID = &asset.ID
					}
				case "thumbnail":
					item.Images = []gallery.GalleryImage{{URL: "https://external.example.test/image.png", Thumbnail: reference}}
				}
				require.Error(t, galleries.CreateGallery(item))
				var count int64
				require.NoError(t, db.Unscoped().Model(&gallery.Gallery{}).Count(&count).Error)
				require.Zero(t, count)
				require.NoError(t, db.Unscoped().Model(&gallery.GalleryImage{}).Count(&count).Error)
				require.Zero(t, count)
			})
		}
	}
}

func TestGalleryLegacyFieldsRemainLocalAndOmittedImageFieldsReloadInsideTransaction(t *testing.T) {
	db, galleries, assets := createGalleryTransactionalMediaBindingTestServices(t)
	item := &gallery.Gallery{Name: "Legacy", Slug: "legacy", CoverImage: "/uploads/old-cover.png"}
	require.NoError(t, db.Create(item).Error)
	legacy := &gallery.GalleryImage{GalleryID: item.ID, URL: "/uploads/old-original.png", Thumbnail: "/uploads/old-thumbnail.png"}
	require.NoError(t, db.Create(legacy).Error)
	title := "Ordinary edit"
	_, err := galleries.UpdateAdminGallery(item.ID, GalleryAdminUpdateInput{Title: &title})
	require.NoError(t, err)
	_, err = galleries.UpdateAdminGalleryImage(legacy.ID, GalleryImageAdminUpdateInput{Title: &title})
	require.NoError(t, err)
	copy := *legacy
	copy.ID = 0
	require.ErrorIs(t, galleries.repo.CreateGalleryImage(&copy), ErrGalleryMediaAssetInvalid)
	changed := *legacy
	changed.Thumbnail = legacy.URL
	require.ErrorIs(t, galleries.UpdateGalleryImage(&changed), ErrGalleryMediaAssetInvalid)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/stale-gallery.png")
	image := &gallery.GalleryImage{GalleryID: item.ID, MediaAssetID: &asset.ID}
	require.NoError(t, galleries.CreateGalleryImage(image))
	stale := *image
	image.MediaAssetID = nil
	image.URL = "https://external.example.test/detached.png"
	image.Thumbnail = ""
	require.NoError(t, galleries.UpdateGalleryImage(image))
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	stale.Title = "Metadata edit"
	require.NoError(t, galleries.repo.UpdateGalleryImage(&stale, repository.GalleryImageMediaFieldMutation{}))
	stored, err := galleries.GetGalleryImageByID(image.ID)
	require.NoError(t, err)
	require.Nil(t, stored.MediaAssetID)
	require.Equal(t, image.URL, stored.URL)
	require.Equal(t, image.Thumbnail, stored.Thumbnail)
}

func TestGalleryImageInsertAndProductLinkFailuresRollBackWholeGalleryAndMediaChanges(t *testing.T) {
	db, galleries, assets := createGalleryTransactionalMediaBindingTestServices(t)
	enableGalleryProductLinkTestTables(t, db)
	require.NoError(t, db.Exec("INSERT INTO products (id,name,slug,locale,status) VALUES (1,'Product','product','en','active')").Error)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/rollback-gallery.png")
	failure := errors.New("gallery image insert failed")
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("fail_gallery_image_insert", func(tx *gorm.DB) {
		if tx.Statement.Table == "gallery_images" {
			tx.AddError(failure)
		}
	}))
	_, err := galleries.CreateAdminGallery(GalleryAdminCreateInput{Title: "Rollback", Slug: "rollback", ProductIDs: []uint{1}, Images: []GalleryImageAdminCreateInput{{MediaAssetID: asset.ID, Title: "Image"}}})
	require.ErrorIs(t, err, failure)
	require.NoError(t, db.Callback().Create().Remove("fail_gallery_image_insert"))
	for _, table := range []string{"galleries", "gallery_images", "gallery_product_links"} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		require.Zero(t, count)
	}
	item := &gallery.Gallery{Name: "Existing", Slug: "existing"}
	require.NoError(t, galleries.CreateGallery(item))
	item.CoverImage = asset.URL
	require.ErrorIs(t, galleries.repo.UpdateGalleryWithProductLinks(item, []uint{999}), gorm.ErrRecordNotFound)
	stored, err := galleries.GetGalleryByID(item.ID)
	require.NoError(t, err)
	require.Empty(t, stored.CoverImage)
	images := []gallery.GalleryImage{{GalleryID: item.ID, MediaAssetID: &asset.ID}, {GalleryID: item.ID, URL: "/uploads/missing-batch.png"}}
	require.Error(t, galleries.BatchCreateImages(images))
	var count int64
	require.NoError(t, db.Model(&gallery.GalleryImage{}).Count(&count).Error)
	require.Zero(t, count)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
}

func TestGalleryMetadataUpdateCannotRewriteStalePreloadedProductLinksOrProducts(t *testing.T) {
	db, galleries, _ := createGalleryTransactionalMediaBindingTestServices(t)
	enableGalleryProductLinkTestTables(t, db)
	require.NoError(t, db.Exec("INSERT INTO products (id,name,slug,locale,status) VALUES (1,'First product','first','en','active'),(2,'Second product','second','en','active')").Error)
	item, err := galleries.CreateAdminGallery(GalleryAdminCreateInput{Title: "Links", Slug: "links", ProductIDs: []uint{1}})
	require.NoError(t, err)
	stale, err := galleries.GetGalleryByID(item.ID)
	require.NoError(t, err)
	require.NoError(t, galleries.ReplaceGalleryProductLinks(item.ID, []uint{2}))
	stale.Name = "Metadata edit"
	stale.ProductLinks[0].Product.Name = "Must not be written"
	require.NoError(t, galleries.UpdateGallery(stale))
	stored, err := galleries.GetGalleryByID(item.ID)
	require.NoError(t, err)
	require.Len(t, stored.ProductLinks, 1)
	require.EqualValues(t, 2, stored.ProductLinks[0].ProductID)
	var name string
	require.NoError(t, db.Table("products").Where("id = ?", 1).Pluck("name", &name).Error)
	require.Equal(t, "First product", name)
}

func TestGalleryPagedExactReferencesProtectSoftDeletedImagesAndDeletedOrMissingParents(t *testing.T) {
	db, galleries, assets := createGalleryTransactionalMediaBindingTestServices(t)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/scanned-gallery.png")
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", StorageKey: "derivatives/gallery-scanned.webp", URL: "https://media.example.test/uploads/derivatives/gallery-scanned.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	item := &gallery.Gallery{Name: "Scanned", Slug: "scanned", CoverImage: "https://shop.example.test/uploads/" + asset.StorageKey + "?v=1"}
	require.NoError(t, db.Create(item).Error)
	rows := make([]gallery.GalleryImage, 257)
	for index := range rows {
		rows[index] = gallery.GalleryImage{GalleryID: item.ID, URL: fmt.Sprintf("https://foreign.example.test/uploads/%s?row=%d", asset.StorageKey, index), Thumbnail: asset.URL + ".backup"}
	}
	rows[256].Thumbnail = derivative.URL + "?v=1"
	require.NoError(t, db.Create(&rows).Error)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	require.NoError(t, galleries.DeleteGalleryImage(rows[256].ID))
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	retained := false
	for _, reference := range report.References {
		retained = retained || reference.UsageState == "retained_deleted_placement"
	}
	require.True(t, retained)
	require.NoError(t, galleries.DeleteGallery(item.ID))
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	for _, reference := range report.References {
		require.Equal(t, "retained_deleted_parent", reference.UsageState)
	}
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	require.NoError(t, db.Unscoped().Where("id = ?", item.ID).Delete(&gallery.Gallery{}).Error)
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "missing_parent", report.References[0].UsageState)
}

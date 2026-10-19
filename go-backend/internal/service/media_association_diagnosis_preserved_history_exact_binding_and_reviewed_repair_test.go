package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
)

func createMediaAssociationRepairReviewFixture(t *testing.T, db *gorm.DB, assets *MediaService) (*media.MediaAsset, *gallery.GalleryImage, *gallery.Gallery) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&gallery.Gallery{}, &gallery.GalleryImage{}))
	if !db.Migrator().HasTable(&audit.AuditLog{}) {
		require.NoError(t, db.AutoMigrate(&audit.AuditLog{}))
	}
	if db.Name() == "postgres" {
		// Reproduce legacy corruption only inside this test's isolated schema.
		require.NoError(t, db.Exec("ALTER TABLE gallery_images DROP CONSTRAINT IF EXISTS fk_galleries_images").Error)
	}
	asset := &media.MediaAsset{Filename: "repair.png", StorageKey: "catalog/repair.png", URL: assets.storage.GetURL("catalog/repair.png"), MediaType: "image", Visibility: "public", Status: "active", LibraryPartition: media.MediaLibraryPartitionShared, UploaderActorType: media.MediaUploadActorAdministrator}
	require.NoError(t, db.Create(asset).Error)
	destination := &gallery.Gallery{Name: "Reviewed draft", Slug: "reviewed-draft", Status: "draft"}
	require.NoError(t, db.Create(destination).Error)
	image := &gallery.GalleryImage{GalleryID: 987654, MediaAssetID: &asset.ID, URL: asset.URL, Thumbnail: asset.URL, Title: "Original title", Description: "Preserve text", Alt: "Preserve alt", Tags: "one,two", Order: 17, Size: 42}
	require.NoError(t, db.Create(image).Error)
	require.NoError(t, db.First(asset, asset.ID).Error)
	require.NoError(t, db.First(image, image.ID).Error)
	require.NoError(t, db.First(destination, destination.ID).Error)
	return asset, image, destination
}

func mediaAssociationGalleryOnlyReviewScope() MediaUsageSummaryAccessScope {
	return MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"gallery_image": true, "gallery_cover": true, "media_asset_registration": true, "media_derivative_registration": true}}
}

func TestMediaAssociationDiagnosisDistinguishesMissingParentDeletedImageAndRestrictedReferences(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	asset, image, _ := createMediaAssociationRepairReviewFixture(t, db, assets)
	retained := *image
	retained.ID = 0
	require.NoError(t, db.Create(&retained).Error)
	require.NoError(t, db.Delete(&retained).Error)
	require.NoError(t, db.Exec("CREATE TABLE ticket_messages(id INTEGER PRIMARY KEY,ticket_id INTEGER,attachments TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO ticket_messages VALUES(876543,765432,?)", `["`+asset.URL+`"]`).Error)
	report, err := assets.DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(t.Context(), asset.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	require.Len(t, report.Entries, 2)
	require.Equal(t, "missing_parent_requires_business_review", report.Entries[0].Diagnosis)
	require.Equal(t, "deleted_history_must_remain_retained", report.Entries[1].Diagnosis)
	require.True(t, report.HasRestrictedReferences)
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	for _, secret := range []string{"876543", "765432", "ticket_message", asset.URL, "storage_key"} {
		require.NotContains(t, string(encoded), secret)
	}
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	var after gallery.GalleryImage
	require.NoError(t, db.First(&after, image.ID).Error)
	require.Equal(t, image.GalleryID, after.GalleryID)
}

func TestMissingGalleryParentPreviewRejectsRetainedPublishedAmbiguousAndUnboundRecords(t *testing.T) {
	for _, scenario := range []string{"valid", "deleted image", "deleted source gallery", "existing source gallery", "published destination", "deleted destination", "different thumbnail", "URL alias", "private asset", "customer asset", "missing selection", "registration conflict", "pending draft", "restricted reference", "historical conversation"} {
		t.Run(scenario, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			assets := newMediaDeleteTestService(t, db, t.TempDir())
			asset, image, destination := createMediaAssociationRepairReviewFixture(t, db, assets)
			scope := mediaAssociationGalleryOnlyReviewScope()
			switch scenario {
			case "deleted image":
				require.NoError(t, db.Delete(image).Error)
			case "existing source gallery", "deleted source gallery":
				parent := gallery.Gallery{ID: image.GalleryID, Name: "Historical parent", Slug: "historical-parent"}
				require.NoError(t, db.Create(&parent).Error)
				if scenario == "deleted source gallery" {
					require.NoError(t, db.Delete(&parent).Error)
				}
			case "published destination":
				require.NoError(t, db.Model(destination).Update("status", "published").Error)
			case "deleted destination":
				require.NoError(t, db.Delete(destination).Error)
			case "different thumbnail":
				require.NoError(t, db.Model(image).Update("thumbnail", "https://other.example.test/photo.png").Error)
			case "URL alias":
				require.NoError(t, db.Model(image).Update("url", "/uploads/"+asset.StorageKey).Error)
			case "private asset":
				require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
			case "customer asset":
				require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
			case "missing selection":
				require.NoError(t, db.Model(image).Update("media_asset_id", nil).Error)
			case "registration conflict":
				other := *asset
				other.ID = 0
				other.URL += "?duplicate=1"
				require.NoError(t, db.Create(&other).Error)
			case "pending draft":
				require.NoError(t, db.AutoMigrate(&media.BusinessDraftMediaUploadReceipt{}))
				require.NoError(t, db.Create(&media.BusinessDraftMediaUploadReceipt{ID: "00000000-0000-0000-0000-000000000001", AssetID: &asset.ID, Status: "uploaded", ObjectKeysJSON: datatypes.JSON(`[]`)}).Error)
			case "restricted reference", "historical conversation":
				require.NoError(t, db.Exec("CREATE TABLE ticket_messages(id INTEGER PRIMARY KEY,ticket_id INTEGER,attachments TEXT)").Error)
				require.NoError(t, db.Exec("INSERT INTO ticket_messages VALUES(33,44,?)", `["`+asset.URL+`"]`).Error)
				if scenario == "historical conversation" {
					require.NoError(t, db.Model(asset).Updates(map[string]interface{}{"library_partition": "unclassified", "uploader_actor_type": "legacy_account"}).Error)
					scope.AllowedReferenceTypes["ticket_message"] = true
				}
			}
			preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, scope)
			if scenario == "valid" {
				require.NoError(t, err)
				require.Len(t, preview.Fingerprint, 64)
			} else {
				require.Error(t, err)
				require.Empty(t, preview.Fingerprint)
			}
			var after gallery.GalleryImage
			require.NoError(t, db.Unscoped().First(&after, image.ID).Error)
			require.Equal(t, image.GalleryID, after.GalleryID)
			var count int64
			require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestMissingGalleryParentRepairRejectsInvalidInputBeforeAnyDatabaseOrStorageAccess(t *testing.T) {
	assets := &MediaService{}
	for _, scenario := range []string{"empty fingerprint", "invalid fingerprint", "missing reason", "long reason", "missing identity", "missing scope", "missing audit"} {
		t.Run(scenario, func(t *testing.T) {
			fingerprint, reason, id, scope, factory := strings.Repeat("a", 64), "审核通过", uint(1), mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} }
			switch scenario {
			case "empty fingerprint":
				fingerprint = ""
			case "invalid fingerprint":
				fingerprint = strings.Repeat("g", 64)
			case "missing reason":
				reason = " "
			case "long reason":
				reason = strings.Repeat("审", 1001)
			case "missing identity":
				id = 0
			case "missing scope":
				scope = MediaUsageSummaryAccessScope{}
			case "missing audit":
				factory = nil
			}
			_, err := assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(context.Background(), id, 1, 2, fingerprint, reason, scope, factory)
			require.ErrorIs(t, err, ErrMediaAssociationRepairInvalidInput)
		})
	}
}

func TestMediaAssociationDiagnosisRegistrationConflictNeverDisclosesOtherRegistrationIdentity(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	asset, _, _ := createMediaAssociationRepairReviewFixture(t, db, assets)
	other := *asset
	other.ID = 456789
	other.URL += "?private-identity"
	other.Visibility = "private"
	require.NoError(t, db.Create(&other).Error)
	report, err := assets.DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(t.Context(), asset.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), fmt.Sprint(other.ID))
	require.NotContains(t, string(encoded), "private-identity")
}

func TestMediaAssociationDraftGalleryCoverAndImageReferencesRemainRetainedUntilPublication(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	asset, image, destination := createMediaAssociationRepairReviewFixture(t, db, assets)
	require.NoError(t, db.Model(image).Update("gallery_id", destination.ID).Error)
	require.NoError(t, db.Model(destination).Update("cover_image", asset.URL).Error)
	references, err := assets.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, references, 2)
	for _, reference := range references {
		require.Equal(t, "retained_unpublished_parent", reference.UsageState)
	}
	report, err := assets.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	require.True(t, report.Rows[0].Retained)
	require.False(t, report.Rows[0].NoRecognizedReferences)
	require.False(t, report.Rows[0].Anomalous)
	require.NoError(t, db.Model(destination).Update("status", "published").Error)
	references, err = assets.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	for _, reference := range references {
		require.Equal(t, "active", reference.UsageState)
	}
}

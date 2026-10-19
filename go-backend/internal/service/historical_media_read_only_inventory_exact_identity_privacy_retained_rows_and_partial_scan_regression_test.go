package service

import (
	"context"
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t *testing.T) (*gorm.DB, *HistoricalMediaReadOnlyInventoryService) {
	t.Helper()
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.ProductBrand{}, &post.Post{}, &gallery.Gallery{}, &gallery.GalleryImage{}, &setting.Setting{})
	config := &storage.Config{Type: storage.StorageTypeLocal, LocalPath: "unused-audit-path", BaseURL: "https://media.example.test"}
	logoConfig := &storage.Config{Type: storage.StorageTypeLocal, LocalPath: "unused-separate-logo-path", BaseURL: "https://logos.example.test"}
	return db, NewHistoricalMediaReadOnlyInventoryService(db, config, logoConfig, "https://shop.example.test")
}

func createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t *testing.T, db *gorm.DB, key string) media.MediaAsset {
	t.Helper()
	asset := media.MediaAsset{URL: "https://media.example.test/uploads/" + key, StorageKey: key, MediaType: "image", Status: "active", Visibility: "public", UploadOrigin: media.MediaUploadOriginAdministratorShared, LibraryPartition: media.MediaLibraryPartitionShared, ClassificationMethod: media.MediaClassificationRecordedUploadContext}
	require.NoError(t, db.Create(&asset).Error)
	return asset
}

func defaultHistoricalMediaReadOnlyInventoryRegressionOptions() HistoricalMediaReadOnlyInventoryOptions {
	return HistoricalMediaReadOnlyInventoryOptions{BatchSize: 2, MaxRecordsPerSource: 100, MaxIssues: 100}
}

func TestHistoricalMediaReadOnlyInventoryFindsExactIdentityIssuesWithoutWritingOrExposingContent(t *testing.T) {
	db, audit := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	good := createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, db, "catalog/good.png")
	other := createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, db, "catalog/other.png")
	duplicate := createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, db, "catalog/duplicate.png")
	require.NoError(t, db.Model(&duplicate).Update("storage_key", `catalog\good.png`).Error)
	retired := createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, db, "catalog/retired.png")
	require.NoError(t, db.Delete(&retired).Error)
	require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: other.ID, URL: "https://media.example.test/uploads/derivatives/other.webp", StorageKey: "derivatives/other.webp"}).Error)
	require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: 999, URL: "https://media.example.test/uploads/derivatives/orphan.webp", StorageKey: "derivatives/orphan.webp"}).Error)
	article := post.Post{Title: "Sensitive title must not be copied", Slug: "audit", Status: "draft", Content: `<p>Private customer text and /uploads/not-an-attribute.png</p><script><img src="/uploads/script-decoy.png"></script><img src="https://shop.example.test/uploads/missing.png?secret=private-token"><img src="/uploads/derivatives/other.webp"><img src="/uploads/catalog/good.png"><a href="/uploads/catalog/retired.png">File</a><img src="https://foreign.example.test/uploads/catalog/other.png">`}
	require.NoError(t, db.Create(&article).Error)
	require.NoError(t, db.Delete(&article).Error)
	require.NoError(t, db.Create(&product.ProductBrand{Name: "Dedicated history", Slug: "dedicated-history", LogoURL: "https://logos.example.test/uploads/site-logo/old.png"}).Error)
	missingID := uint(999)
	require.NoError(t, db.Create(&gallery.GalleryImage{GalleryID: 444, MediaAssetID: &missingID}).Error)
	require.NoError(t, db.Create(&gallery.GalleryImage{GalleryID: 444, MediaAssetID: &good.ID, URL: "/uploads/catalog/other.png"}).Error)
	// SQLite itself rejects any attempted write during the audit.
	require.NoError(t, db.Exec("PRAGMA query_only = ON").Error)
	defer db.Exec("PRAGMA query_only = OFF")
	report, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	require.True(t, report.RegistrationIndexComplete)
	require.False(t, report.RequestedSourcesComplete, "unavailable business tables must remain visible in overall coverage")
	for _, issue := range report.Issues {
		if issue.Table == "posts" {
			require.True(t, issue.RetainedDeletedRecord)
		}
	}
	require.Equal(t, 6, report.RegistrationsScanned)
	require.Equal(t, 1, report.IssueCounts["duplicate_registered_object_identity"])
	require.Equal(t, 1, report.IssueCounts["registration_url_and_storage_key_disagree"])
	require.Equal(t, 1, report.IssueCounts["derivative_parent_registration_missing"])
	require.Equal(t, 1, report.IssueCounts["business_managed_url_not_registered"])
	require.Equal(t, 1, report.IssueCounts["untrusted_upload_url_origin_requires_review"])
	require.Equal(t, 1, report.IssueCounts["business_reference_registration_identity_ambiguous"])
	require.Equal(t, 1, report.IssueCounts["business_asset_id_not_registered"])
	require.Equal(t, 1, report.IssueCounts["business_asset_id_and_url_identity_disagree"])
	require.Equal(t, 1, report.IssueCounts["business_reference_uses_retired_asset"])
	require.Equal(t, 1, report.IssueCounts["historical_dedicated_object_business_reference"])
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	for _, secret := range []string{"private-token", "Sensitive title", "Private customer text", "not-an-attribute.png", "missing.png", "script-decoy.png", "old.png"} {
		require.NotContains(t, string(encoded), secret)
	}
	var stored post.Post
	require.NoError(t, db.Unscoped().First(&stored, article.ID).Error)
	require.Equal(t, article.Content, stored.Content)
	require.True(t, stored.DeletedAt.Valid)
	require.False(t, report.PhysicalObjectsChecked)
	require.False(t, report.UnreferencedObjectsEvaluated)
}

func TestHistoricalMediaReadOnlyInventoryTruncationPreservesTotalsAndDoesNotInferMissingRegistrations(t *testing.T) {
	db, audit := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, db, "catalog/first.png")
	second := createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, db, "catalog/second.png")
	require.NoError(t, db.Create(&product.ProductBrand{Name: "Brand", Slug: "brand", LogoURL: second.URL}).Error)
	options := defaultHistoricalMediaReadOnlyInventoryRegressionOptions()
	options.MaxRecordsPerSource = 1
	report, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), options)
	require.NoError(t, err)
	require.False(t, report.RegistrationIndexComplete)
	require.False(t, report.RequestedSourcesComplete)
	require.Zero(t, report.IssueCounts["business_managed_url_not_registered"])
	require.Zero(t, report.IssueCounts["untrusted_upload_url_origin_requires_review"])
	for _, source := range report.Sources {
		if source.Table == "media_assets" {
			require.EqualValues(t, 1, source.RemainingRows)
		}
	}
	for index, path := range []string{"/uploads/missing-one.png", "/uploads/missing-two.png"} {
		require.NoError(t, db.Create(&product.ProductBrand{Name: "Missing", Slug: string(rune('a' + index)), LogoURL: path}).Error)
	}
	options.MaxRecordsPerSource, options.MaxIssues = 100, 1
	report, err = audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), options)
	require.NoError(t, err)
	require.True(t, report.RegistrationIndexComplete)
	require.Equal(t, 2, report.IssueCounts["business_managed_url_not_registered"])
	require.True(t, report.IssuesTruncated)
	require.Len(t, report.Issues, 1)
	require.Greater(t, report.TotalIssues, len(report.Issues))
}

func TestHistoricalMediaReadOnlyInventoryReportsUnavailableSchemaAndHonorsCancellation(t *testing.T) {
	db, audit := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	require.NoError(t, db.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY)").Error)
	report, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	require.Greater(t, report.IssueCounts["business_source_column_unavailable"], 0)
	require.False(t, report.RequestedSourcesComplete)
	require.NoError(t, db.Migrator().DropTable(&media.MediaAssetDerivative{}))
	report, err = audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	require.False(t, report.RegistrationIndexComplete)
	require.Equal(t, 1, report.IssueCounts["registration_table_unavailable"])
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = audit.AuditHistoricalRegistrationsAndBusinessReferences(canceled, defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.ErrorIs(t, err, context.Canceled)
}

func TestHistoricalMediaReadOnlyInventoryConfigurationProvenanceOmitsCredentialsAndTokens(t *testing.T) {
	_, audit := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	audit.sharedStorageConfig.BaseURL = "https://sensitive-user:sensitive-password@media.example.test/public?signature=sensitive-signature#sensitive-fragment"
	audit.sharedStorageConfig.AccessKeyID = "sensitive-access-key"
	audit.sharedStorageConfig.SecretAccessKey = "sensitive-secret-key"
	audit.sharedStorageConfig.PrivateBaseURL = "https://sensitive-private.example.test"
	audit.sharedStorageConfig.LocalPath = "sensitive-local-directory"
	audit.storefrontURL = "https://sensitive-store-user:sensitive-store-password@shop.example.test/?token=sensitive-store-token"
	report, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	require.Equal(t, "https://media.example.test/public", report.IdentityConfiguration["shared_public_base_url"])
	require.Equal(t, "https://shop.example.test/", report.IdentityConfiguration["storefront_base_url"])
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "sensitive-")
}

func TestHistoricalMediaReadOnlyInventoryDistinguishesUnverifiedDedicatedOriginsFromProvenObjectKeyConflicts(t *testing.T) {
	db, audit := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	require.NoError(t, db.Exec("CREATE TABLE site_logo_assets (id INTEGER PRIMARY KEY, url TEXT, storage_key TEXT)").Error)
	require.NoError(t, db.Table("site_logo_assets").Create([]map[string]any{
		{"id": 1, "url": "https://retired-unknown.example.test/uploads/site-logo/original.png", "storage_key": "site-logo/original.png"},
		{"id": 2, "url": "https://logos.example.test/uploads/site-logo/original.png", "storage_key": "site-logo/different.png"},
		{"id": 3, "url": "https://logos.example.test/uploads/site-logo/original.png", "storage_key": "site-logo/original.png"},
	}).Error)
	report, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	require.Equal(t, 1, report.IssueCounts["dedicated_owner_url_identity_requires_review"])
	require.Equal(t, 1, report.IssueCounts["dedicated_owner_url_and_storage_key_disagree"])
	for _, issue := range report.Issues {
		if issue.Code == "dedicated_owner_url_and_storage_key_disagree" {
			require.EqualValues(t, 2, issue.RecordID, "unverified origins alone do not prove a storage key mismatch")
		}
	}
}

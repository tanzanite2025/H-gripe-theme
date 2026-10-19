package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type legacyJSONMediaReferenceRetentionTestSource struct {
	Table, ResourceType, JSONColumn, ParentColumn, Schema string
	HasSoftDeletion                                       bool
}

func legacyJSONMediaReferenceSourcesForRetentionTest() []legacyJSONMediaReferenceRetentionTestSource {
	return []legacyJSONMediaReferenceRetentionTestSource{
		{Table: "showcases", ResourceType: "showcase", JSONColumn: "images", Schema: "CREATE TABLE showcases (id INTEGER PRIMARY KEY, title TEXT, images TEXT)"},
		{Table: "reviews", ResourceType: "review", JSONColumn: "images", ParentColumn: "product_id", HasSoftDeletion: true, Schema: "CREATE TABLE reviews (id INTEGER PRIMARY KEY, product_id INTEGER, images TEXT, deleted_at TIMESTAMP)"},
		{Table: "warranty_claims", ResourceType: "warranty_claim", JSONColumn: "images", HasSoftDeletion: true, Schema: "CREATE TABLE warranty_claims (id INTEGER PRIMARY KEY, images TEXT, video_url TEXT, deleted_at TIMESTAMP)"},
		{Table: "suggestion_feedback", ResourceType: "suggestion_feedback", JSONColumn: "attachments", HasSoftDeletion: true, Schema: "CREATE TABLE suggestion_feedback (id INTEGER PRIMARY KEY, attachments TEXT, deleted_at TIMESTAMP)"},
		{Table: "ticket_auto_replies", ResourceType: "ticket_auto_reply", JSONColumn: "attachments", Schema: "CREATE TABLE ticket_auto_replies (id INTEGER PRIMARY KEY, attachments TEXT)"},
		{Table: "ticket_messages", ResourceType: "ticket_message", JSONColumn: "attachments", ParentColumn: "ticket_id", Schema: "CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, ticket_id INTEGER, attachments TEXT)"},
	}
}

func encodeLegacyMediaReferenceRetentionTestJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}

func insertLegacyMediaReferenceRetentionTestRow(t *testing.T, db *gorm.DB, source legacyJSONMediaReferenceRetentionTestSource, id uint, rawJSON string, deleted bool, videoURL string) {
	t.Helper()
	values := map[string]any{"id": id, source.JSONColumn: rawJSON}
	if source.ParentColumn != "" {
		values[source.ParentColumn] = 71
	}
	if source.Table == "showcases" {
		values["title"] = "Synthetic retained media placement"
	}
	if source.Table == "warranty_claims" {
		values["video_url"] = videoURL
	}
	if source.HasSoftDeletion && deleted {
		values["deleted_at"] = time.Now().UTC()
	}
	require.NoError(t, db.Table(source.Table).Create(values).Error)
}

func TestLegacyJSONMediaReferencesDecodeEscapesAndMatchExactOriginalDerivativeAndTrustedAliasIdentities(t *testing.T) {
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		for _, scenario := range []string{"escaped-slashes", "unicode-slashes", "nested-object", "scalar-url", "query-fragment", "storefront-alias", "original-url-only-alias", "retired-derivative-url-only-alias", "foreign-host", "longer-filename", "embedded-prose", "non-address-values"} {
			t.Run(source.Table+"/"+scenario, func(t *testing.T) {
				db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
				require.NoError(t, db.Exec(source.Schema).Error)
				root := t.TempDir()
				service := newMediaDeleteTestService(t, db, root)
				service.siteURL = "https://shop.example.test"
				originalKey, derivativeKey := "catalog/legacy-original.png", "media-derivatives/legacy-card.png"
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, originalKey)
				derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "card", PresetVersion: 1, URL: service.storage.GetURL(derivativeKey), StorageKey: derivativeKey}
				require.NoError(t, db.Create(&derivative).Error)
				writeMediaCleanupOwnershipConflictTestObjects(t, root, originalKey, derivativeKey)
				payload := encodeLegacyMediaReferenceRetentionTestJSON(t, []string{asset.URL})
				shouldRetain := true
				switch scenario {
				case "escaped-slashes":
					payload = strings.ReplaceAll(payload, "/", `\/`)
				case "unicode-slashes":
					payload = strings.ReplaceAll(payload, "/", `\u002f`)
				case "nested-object":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []any{map[string]any{"file": map[string]any{"url": derivative.URL}}})
				case "scalar-url":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, asset.URL)
				case "query-fragment":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []string{asset.URL + "?historical=1#preview"})
				case "storefront-alias", "original-url-only-alias":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []string{"https://shop.example.test/uploads/" + originalKey})
					if scenario == "original-url-only-alias" {
						require.NoError(t, db.Model(asset).Update("storage_key", "").Error)
					}
				case "retired-derivative-url-only-alias":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []string{"https://shop.example.test/uploads/" + derivativeKey})
					require.NoError(t, db.Model(&derivative).Update("storage_key", "").Error)
					require.NoError(t, db.Delete(&derivative).Error)
				case "foreign-host":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []string{"https://foreign.example.test/uploads/" + originalKey})
					shouldRetain = false
				case "longer-filename":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []string{asset.URL + ".another.png"})
					shouldRetain = false
				case "embedded-prose":
					payload = encodeLegacyMediaReferenceRetentionTestJSON(t, []string{"see " + asset.URL + " for details"})
					shouldRetain = false
				case "non-address-values":
					payload = `[null,42,true,{},[]]`
					shouldRetain = false
				}
				insertLegacyMediaReferenceRetentionTestRow(t, db, source, 1, payload, false, "")
				report, err := service.GetAssetReferences(asset.ID)
				require.NoError(t, err)
				if shouldRetain {
					require.Equal(t, 1, report.Total)
					require.Equal(t, source.ResourceType, report.References[0].ResourceType)
					require.Equal(t, source.JSONColumn, report.References[0].Field)
					require.Equal(t, "active", report.References[0].UsageState)
					if source.ParentColumn != "" {
						require.EqualValues(t, 71, report.References[0].ParentResourceID)
					}
					require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				} else {
					require.Zero(t, report.Total)
					require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
				}
				for _, key := range []string{originalKey, derivativeKey} {
					_, err := os.Stat(filepath.Join(root, filepath.FromSlash(key)))
					if shouldRetain {
						require.NoError(t, err)
					} else {
						require.ErrorIs(t, err, os.ErrNotExist)
					}
				}
			})
		}
	}
}

func TestRetainedDeletedLegacyMediaReferencesSurviveRestorationAndDelayOriginalCleanupEventUntilPermanentDetachment(t *testing.T) {
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		if !source.HasSoftDeletion {
			continue
		}
		fields := []string{source.JSONColumn}
		if source.Table == "warranty_claims" {
			fields = append(fields, "video_url")
		}
		for _, field := range fields {
			t.Run(source.Table+"/"+field, func(t *testing.T) {
				db := newMediaDeleteTestDB(t, &media.MediaAsset{})
				require.NoError(t, db.Exec(source.Schema).Error)
				root := t.TempDir()
				service := newMediaDeleteTestService(t, db, root)
				key := "catalog/retained-business-media.png"
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, key)
				writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
				payload, video := encodeLegacyMediaReferenceRetentionTestJSON(t, []map[string]string{{"url": asset.URL}}), ""
				if field == "video_url" {
					payload, video = "[]", asset.URL+"?video=1"
				}
				insertLegacyMediaReferenceRetentionTestRow(t, db, source, 1, payload, true, video)
				report, err := service.GetAssetReferences(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 1, report.Total)
				require.Equal(t, field, report.References[0].Field)
				require.Equal(t, "retained_deleted_parent", report.References[0].UsageState)
				review, err := service.ReviewMediaAssetLibraryOwnershipForManualClassification(t.Context(), asset.ID)
				require.NoError(t, err)
				require.Equal(t, 1, review.RetainedDeletedParentReferences)
				require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				require.NoError(t, db.Table(source.Table).Where("id = ?", 1).Update("deleted_at", nil).Error)
				report, err = service.GetAssetReferences(asset.ID)
				require.NoError(t, err)
				require.Equal(t, "active", report.References[0].UsageState)
				require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				// Only remove synthetic placements in this test, never implicitly
				// turn business soft deletion into permanent attachment detachment.
				require.NoError(t, db.Table(source.Table).Where("id = ?", 1).Delete(map[string]any{}).Error)
				provider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: service.storage, deleteError: errors.New("synthetic delayed legacy cleanup")}
				service.storage = provider
				require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
				var event outbox.Event
				require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
				insertLegacyMediaReferenceRetentionTestRow(t, db, source, 1, payload, true, video)
				provider.deleteError = nil
				handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
				require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
				_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
				require.NoError(t, err)
				var tombstone media.MediaAsset
				require.NoError(t, db.Unscoped().First(&tombstone, asset.ID).Error)
				require.True(t, tombstone.DeletedAt.Valid)
				detach := map[string]any{field: "[]"}
				if field == "video_url" {
					detach[field] = ""
				}
				require.NoError(t, db.Table(source.Table).Where("id = ?", 1).Updates(detach).Error)
				require.NoError(t, handler.Handle(t.Context(), event))
				require.NoError(t, handler.Handle(t.Context(), event))
				_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
				require.ErrorIs(t, err, os.ErrNotExist)
				var eventCount int64
				require.NoError(t, db.Model(&outbox.Event{}).Where("event_key = ?", event.EventKey).Count(&eventCount).Error)
				require.EqualValues(t, 1, eventCount)
			})
		}
	}
}

func TestMalformedLegacyMediaJSONBlocksRetirementAndReplayWithoutExposingAttachmentValues(t *testing.T) {
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		t.Run(source.Table, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{})
			require.NoError(t, db.Exec(source.Schema).Error)
			root := t.TempDir()
			service := newMediaDeleteTestService(t, db, root)
			key := "catalog/malformed-retained-json.png"
			asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, key)
			writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
			privateValue := "https://private.example.test/never-log-this-attachment.png?synthetic-secret=private"
			malformed := `["` + privateValue + `"`
			insertLegacyMediaReferenceRetentionTestRow(t, db, source, 1, malformed, source.HasSoftDeletion, "")
			_, err := service.GetAssetReferences(asset.ID)
			require.ErrorContains(t, err, source.Table+"."+source.JSONColumn)
			require.NotContains(t, err.Error(), privateValue)
			require.Error(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
			_, err = service.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			var count int64
			require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
			require.Zero(t, count)
			require.NoError(t, db.Table(source.Table).Where("id = ?", 1).Update(source.JSONColumn, "[]").Error)
			provider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: service.storage, deleteError: errors.New("synthetic storage outage")}
			service.storage = provider
			require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
			var event outbox.Event
			require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
			require.NoError(t, db.Table(source.Table).Where("id = ?", 1).Update(source.JSONColumn, malformed).Error)
			provider.deleteError = nil
			handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
			err = handler.Handle(t.Context(), event)
			require.ErrorContains(t, err, "invalid JSON")
			require.NotContains(t, err.Error(), privateValue)
			_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
			require.NoError(t, err)
			require.NoError(t, db.Table(source.Table).Where("id = ?", 1).Update(source.JSONColumn, "[]").Error)
			require.NoError(t, handler.Handle(t.Context(), event))
			_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestLegacyMediaReferencePaginationIncludesRetainedRowsAfterFirstPage(t *testing.T) {
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		t.Run(source.Table, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{})
			require.NoError(t, db.Exec(source.Schema).Error)
			root := t.TempDir()
			service := newMediaDeleteTestService(t, db, root)
			key := "catalog/reference-after-first-page.png"
			asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, key)
			writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
			for id := uint(1); id <= 256; id++ {
				insertLegacyMediaReferenceRetentionTestRow(t, db, source, id, "[]", false, "")
			}
			payload := strings.ReplaceAll(encodeLegacyMediaReferenceRetentionTestJSON(t, []string{asset.URL}), "/", `\/`)
			insertLegacyMediaReferenceRetentionTestRow(t, db, source, 257, payload, source.HasSoftDeletion, "")
			report, err := service.GetAssetReferences(asset.ID)
			require.NoError(t, err)
			require.Equal(t, 1, report.Total)
			require.EqualValues(t, 257, report.References[0].ResourceID)
			require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
		})
	}
}

func TestWarrantyMediaReferenceReportsEveryMatchingFieldOnce(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	root := t.TempDir()
	service := newMediaDeleteTestService(t, db, root)
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/shared-image-and-video-reference.png")
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		if source.Table == "warranty_claims" {
			require.NoError(t, db.Exec(source.Schema).Error)
			insertLegacyMediaReferenceRetentionTestRow(t, db, source, 1, encodeLegacyMediaReferenceRetentionTestJSON(t, []string{asset.URL, asset.URL}), true, asset.URL)
		}
	}
	report, err := service.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "video_url, images", report.References[0].Field)
	require.Equal(t, fmt.Sprintf("保修申请 #%d", 1), report.References[0].Label)
}

func TestLegacyPrivateAttachmentRetentionRecognizesConfiguredAliasesAndPreservesPrivateVisibility(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &outbox.Event{})
	require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, ticket_id INTEGER, attachments TEXT)").Error)
	root := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root,
		PrivateLocalPath: root + "-private", BaseURL: "https://media.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	key := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/private-legacy-retention.png"
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, key)
	require.NoError(t, db.Model(asset).Updates(map[string]any{"storage_key": "", "url": "https://private.example.test/uploads/" + key, "visibility": "private"}).Error)
	writeMediaCleanupOwnershipConflictTestObjects(t, root+"-private", key)
	payload := encodeLegacyMediaReferenceRetentionTestJSON(t, []string{"https://media.example.test/uploads/" + key + "?old-client=1"})
	require.NoError(t, db.Table("ticket_messages").Create(map[string]any{"id": 1, "ticket_id": 71, "attachments": payload}).Error)
	report, err := service.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	encodedReport := encodeLegacyMediaReferenceRetentionTestJSON(t, report)
	require.NotContains(t, encodedReport, key)
	require.NotContains(t, encodedReport, "private.example.test")
	require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	current, err := service.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, "private", current.Visibility)
	_, err = os.Stat(filepath.Join(root+"-private", filepath.FromSlash(key)))
	require.NoError(t, err)
}

func TestLegacyJSONMediaRetentionResolvesURLOnlyOriginalAndDerivativeIdentitiesAcrossConfiguredProviders(t *testing.T) {
	for _, providerType := range []storage.StorageType{storage.StorageTypeLocal, storage.StorageTypeS3, storage.StorageTypeOSS} {
		for _, referenceType := range []string{"original", "retained-derivative"} {
			t.Run(string(providerType)+"/"+referenceType, func(t *testing.T) {
				db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{})
				require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, ticket_id INTEGER, attachments TEXT)").Error)
				config := &storage.Config{Type: providerType, LocalPath: t.TempDir(), BaseURL: "https://media.example.test/assets",
					Bucket: "media-public", Region: "us-east-1", AccessKeyID: "synthetic-test-key", SecretAccessKey: "synthetic-test-secret"}
				if providerType == storage.StorageTypeOSS {
					config.Endpoint = "https://oss-cn-hangzhou.aliyuncs.com"
				}
				provider, err := storage.NewStorageService(config)
				require.NoError(t, err)
				recorder := &mediaCleanupStorageDeletionAttemptRecorder{StorageService: provider}
				service := NewMediaService(repository.NewMediaRepository(db), recorder, nil, "https://shop.example.test", 20<<30)
				service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				key := "catalog/url-only-legacy-original.png"
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, key)
				require.NoError(t, db.Model(asset).Update("storage_key", "").Error)
				if referenceType == "retained-derivative" {
					key = "media-derivatives/url-only-retained.png"
					derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "card", PresetVersion: 1, URL: provider.GetURL(key)}
					require.NoError(t, db.Create(&derivative).Error)
					require.NoError(t, db.Delete(&derivative).Error)
				}
				reference := "https://shop.example.test/uploads/" + key
				switch providerType {
				case storage.StorageTypeS3:
					reference = "https://media-public.s3.us-east-1.amazonaws.com/" + key
				case storage.StorageTypeOSS:
					reference = "https://media-public.oss-cn-hangzhou.aliyuncs.com/" + key
				}
				require.NoError(t, db.Table("ticket_messages").Create(map[string]any{"id": 1, "ticket_id": 71,
					"attachments": encodeLegacyMediaReferenceRetentionTestJSON(t, []string{reference})}).Error)
				report, err := service.GetAssetReferences(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 1, report.Total)
				require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				require.Zero(t, recorder.deleteCalls)
			})
		}
	}
}

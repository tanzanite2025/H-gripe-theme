package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createCustomerServiceDanglingAttachmentReferenceFixture(t *testing.T) (*gorm.DB, *MediaService, *media.MediaAsset) {
	t.Helper()
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.MediaUploadSession{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	for _, statement := range []string{
		"CREATE TABLE tickets(id INTEGER PRIMARY KEY,deleted_at DATETIME)",
		"CREATE TABLE ticket_messages(id INTEGER PRIMARY KEY,ticket_id INTEGER,attachments TEXT,metadata TEXT,message_type TEXT)",
		"CREATE TABLE ticket_message_media_attachments(id INTEGER PRIMARY KEY,ticket_message_id INTEGER,media_asset_id INTEGER,upload_session_id TEXT)",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	asset := &media.MediaAsset{Filename: "protected.png", URL: "https://media.example.test/uploads/protected.png", StorageKey: "protected.png", Status: "active", MediaType: "image", Visibility: "public", LibraryPartition: media.MediaLibraryPartitionShared, UploaderActorType: media.MediaUploadActorAdministrator}
	require.NoError(t, db.Create(asset).Error)
	return db, assets, asset
}

func customerServiceAttachmentReferenceDiagnosisScope() MediaUsageSummaryAccessScope {
	return MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"ticket_message": true, "ticket_message_attachment_binding": true, "media_upload_session": true}}
}

func TestCustomerServiceDanglingAuthoritativeAttachmentBindingRetainsAssetWithoutInventingConversation(t *testing.T) {
	for _, scenario := range []string{"missing message", "missing messages table", "missing conversation", "deleted conversation", "active conversation"} {
		t.Run(scenario, func(t *testing.T) {
			db, assets, asset := createCustomerServiceDanglingAttachmentReferenceFixture(t)
			require.NoError(t, db.Exec("INSERT INTO ticket_message_media_attachments VALUES(876543,765432,?,NULL)", asset.ID).Error)
			if scenario == "missing messages table" {
				require.NoError(t, db.Exec("DROP TABLE ticket_messages").Error)
			} else if scenario != "missing message" {
				require.NoError(t, db.Exec("INSERT INTO ticket_messages VALUES(765432,654321,'[]','','text')").Error)
			}
			if scenario == "active conversation" || scenario == "deleted conversation" {
				require.NoError(t, db.Exec("INSERT INTO tickets VALUES(654321,NULL)").Error)
				if scenario == "deleted conversation" {
					require.NoError(t, db.Exec("UPDATE tickets SET deleted_at = ? WHERE id=654321", time.Now()).Error)
				}
			}
			refs, err := assets.repo.FindAssetReferences(asset)
			require.NoError(t, err)
			require.Len(t, refs, 1)
			expectedState := "missing_parent"
			if scenario == "active conversation" {
				expectedState = "active"
			}
			if scenario == "deleted conversation" {
				expectedState = "retained_deleted_parent"
			}
			require.Equal(t, expectedState, refs[0].UsageState)
			if scenario == "missing message" || scenario == "missing messages table" {
				require.Equal(t, "ticket_message_attachment_binding", refs[0].ResourceType)
				require.EqualValues(t, 876543, refs[0].ResourceID)
				require.Zero(t, refs[0].ParentResourceID)
			}
			report, err := assets.DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(t.Context(), asset.ID, customerServiceAttachmentReferenceDiagnosisScope())
			require.NoError(t, err)
			if scenario == "active conversation" {
				require.Empty(t, report.Entries)
			} else {
				require.Len(t, report.Entries, 1)
			}
			restricted, err := assets.DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(t.Context(), asset.ID, MediaUsageSummaryAccessScope{})
			require.NoError(t, err)
			require.True(t, restricted.HasRestrictedReferences)
			require.Empty(t, restricted.Entries)
			encoded, err := json.Marshal(restricted)
			require.NoError(t, err)
			for _, secret := range []string{"876543", "765432", "654321", "ticket_message", asset.URL} {
				require.NotContains(t, string(encoded), secret)
			}
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			var count int64
			require.NoError(t, db.Table("ticket_message_media_attachments").Count(&count).Error)
			require.EqualValues(t, 1, count)
		})
	}
}

func TestCustomerServiceLegacyAttachmentAndStructuredMetadataShareConversationRetentionDiagnosis(t *testing.T) {
	for _, source := range []string{"attachments", "metadata", "both"} {
		t.Run(source, func(t *testing.T) {
			db, assets, asset := createCustomerServiceDanglingAttachmentReferenceFixture(t)
			attachments, metadata := "[]", ""
			if source != "metadata" {
				attachments = `["` + asset.URL + `"]`
			}
			if source != "attachments" {
				metadata = `{"image_url":"` + asset.URL + `"}`
			}
			require.NoError(t, db.Exec("INSERT INTO ticket_messages VALUES(17,23,?,?,'image')", attachments, metadata).Error)
			for _, state := range []string{"missing_parent", "retained_deleted_parent", "active"} {
				if state == "retained_deleted_parent" {
					require.NoError(t, db.Exec("INSERT INTO tickets VALUES(23,?)", time.Now()).Error)
				}
				if state == "active" {
					require.NoError(t, db.Exec("UPDATE tickets SET deleted_at=NULL WHERE id=23").Error)
				}
				refs, err := assets.repo.FindAssetReferences(asset)
				require.NoError(t, err)
				require.Len(t, refs, 1)
				require.Equal(t, state, refs[0].UsageState)
				require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			}
		})
	}
}

func TestCustomerServiceUploadReceiptRetainsItsOwnAssetDespiteMissingOrMismatchedBindings(t *testing.T) {
	for _, scenario := range []string{"ready", "bound without binding", "bound missing message", "bound other asset", "bound missing conversation", "bound deleted conversation"} {
		t.Run(scenario, func(t *testing.T) {
			db, assets, asset := createCustomerServiceDanglingAttachmentReferenceFixture(t)
			status := "bound"
			if scenario == "ready" {
				status = "ready"
			}
			session := media.MediaUploadSession{ID: "11111111-1111-1111-1111-111111111111", BusinessContextType: media.MediaUploadSessionCustomerServiceContext, BusinessContextID: 23, ObjectKey: asset.StorageKey, MediaAssetID: &asset.ID, Status: status, ExpiresAt: time.Now().Add(-48 * time.Hour)}
			require.NoError(t, db.Create(&session).Error)
			if scenario != "bound missing conversation" {
				require.NoError(t, db.Exec("INSERT INTO tickets VALUES(23,NULL)").Error)
			}
			if scenario == "bound deleted conversation" {
				require.NoError(t, db.Exec("UPDATE tickets SET deleted_at=? WHERE id=23", time.Now()).Error)
			}
			if scenario == "bound missing message" {
				require.NoError(t, db.Exec("INSERT INTO ticket_message_media_attachments VALUES(17,19,?,?)", asset.ID, session.ID).Error)
			}
			if scenario == "bound other asset" {
				require.NoError(t, db.Exec("INSERT INTO ticket_message_media_attachments VALUES(17,19,999999,?)", session.ID).Error)
			}
			refs, err := assets.repo.FindAssetReferences(asset)
			require.NoError(t, err)
			expectedState := "retained_upload_receipt"
			if scenario == "ready" {
				expectedState = "pending_upload"
			}
			if scenario == "bound missing conversation" {
				expectedState = "missing_parent"
			}
			if scenario == "bound deleted conversation" {
				expectedState = "retained_deleted_parent"
			}
			found := false
			for _, reference := range refs {
				if reference.ResourceType == "media_upload_session" {
					found = true
					require.Equal(t, expectedState, reference.UsageState)
				}
			}
			require.True(t, found)
			row := summarizeMediaUsageWithPermissionScopedReferenceStates(*asset, refs, customerServiceAttachmentReferenceDiagnosisScope())
			require.True(t, row.Retained)
			require.False(t, row.NoRecognizedReferences)
			for _, reference := range row.References {
				if reference.ResourceType == "media_upload_session" {
					require.Equal(t, expectedState, reference.UsageState)
				}
			}
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
		})
	}
}

func TestCustomerServiceManagedOwnersKeepUnclassifiedHistoricalAssetsConversationScopedWithoutMessageProjections(t *testing.T) {
	for _, source := range []string{"orphan binding", "bound receipt", "exact object receipt"} {
		t.Run(source, func(t *testing.T) {
			db, assets, asset := createCustomerServiceDanglingAttachmentReferenceFixture(t)
			require.NoError(t, db.Model(asset).Updates(map[string]any{"library_partition": "unclassified", "uploader_actor_type": "legacy_account"}).Error)
			require.NoError(t, db.First(asset, asset.ID).Error)
			require.NoError(t, db.Exec("DROP TABLE ticket_messages").Error)
			if source == "orphan binding" {
				require.NoError(t, db.Exec("INSERT INTO ticket_message_media_attachments VALUES(17,19,?,NULL)", asset.ID).Error)
			} else {
				var assetID *uint
				if source == "bound receipt" {
					assetID = &asset.ID
				}
				require.NoError(t, db.Create(&media.MediaUploadSession{ID: "11111111-1111-1111-1111-111111111111", BusinessContextType: media.MediaUploadSessionCustomerServiceContext, BusinessContextID: 23, ObjectKey: asset.StorageKey, MediaAssetID: assetID, Status: "bound", ExpiresAt: time.Now()}).Error)
			}
			required, err := assets.repo.HistoricalUnclassifiedMediaRequiresConversationAccess(asset)
			require.NoError(t, err)
			require.True(t, required)
		})
	}
}

func TestCustomerServiceDanglingBindingPreservesOriginalCleanupEventAndRetriesOnlyAfterLastUnlink(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	root := t.TempDir()
	assets := newMediaDeleteTestService(t, db, root)
	asset := media.MediaAsset{Filename: "cleanup-protected.png", URL: assets.storage.GetURL("cleanup-protected.png"), StorageKey: "cleanup-protected.png", Status: "active", Visibility: "public", LibraryPartition: media.MediaLibraryPartitionShared}
	require.NoError(t, db.Create(&asset).Error)
	storedPath := filepath.Join(root, asset.StorageKey)
	require.NoError(t, os.WriteFile(storedPath, []byte("synthetic protected bytes"), 0600))
	// Commit retirement and its receipt without running DeleteAsset's immediate
	// best-effort cleanup, so the original worker event is genuinely delayed.
	require.NoError(t, assets.repo.WithTransaction(func(repo *repository.MediaRepository, tx *gorm.DB) error {
		locked, err := repo.FindAssetByIDForUpdate(asset.ID)
		if err != nil {
			return err
		}
		return assets.retireUnreferencedMediaAssetWithinTransaction(repo, tx, locked)
	}))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
	require.NoError(t, db.Exec("CREATE TABLE ticket_message_media_attachments(id INTEGER PRIMARY KEY,ticket_message_id INTEGER,media_asset_id INTEGER)").Error)
	require.NoError(t, db.Exec("INSERT INTO ticket_message_media_attachments VALUES(17,19,?)", asset.ID).Error)
	handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	bytes, err := os.ReadFile(storedPath)
	require.NoError(t, err)
	require.Equal(t, "synthetic protected bytes", string(bytes))
	var originalEvent outbox.Event
	require.NoError(t, db.First(&originalEvent, event.ID).Error)
	require.Equal(t, event.Payload, originalEvent.Payload)
	require.NoError(t, db.Exec("DELETE FROM ticket_message_media_attachments WHERE id=17").Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(storedPath)
	require.True(t, os.IsNotExist(err))
	require.NoError(t, db.First(&originalEvent, event.ID).Error)
	require.Equal(t, event.Payload, originalEvent.Payload)
}

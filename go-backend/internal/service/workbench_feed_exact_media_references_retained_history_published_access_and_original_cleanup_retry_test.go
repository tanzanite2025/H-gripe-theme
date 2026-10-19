package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func createWorkbenchFeedReferenceRetentionTestTables(t *testing.T, db *gorm.DB, withParents bool) {
	t.Helper()
	require.NoError(t, db.Exec("CREATE TABLE workbench_feed_media (id INTEGER PRIMARY KEY, entry_id INTEGER, file_path TEXT, deleted_at TIMESTAMP)").Error)
	if withParents {
		require.NoError(t, db.Exec("CREATE TABLE workbench_feed_entries (id INTEGER PRIMARY KEY, status TEXT, deleted_at TIMESTAMP)").Error)
	}
}

func TestWorkbenchFeedReferenceScannerMatchesExactOriginalDerivativeAndTrustedAliasesInRetainedStates(t *testing.T) {
	for _, scenario := range []string{"original", "derivative", "trusted-alias", "url-only-registration", "private-alias", "soft-deleted-parent", "soft-deleted-placement", "draft", "archived", "missing-parent", "missing-parent-table", "foreign-host", "longer-path", "embedded-prose", "pagination"} {
		t.Run(scenario, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			assets.siteURL = "https://shop.example.test"
			createWorkbenchFeedReferenceRetentionTestTables(t, db, scenario != "missing-parent-table")
			key := "workbench-feed/original.png"
			if scenario == "private-alias" {
				key = "private/customer-service/conversations/test/original.png"
			}
			asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, key)
			derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, URL: assets.storage.GetURL("derivatives/feed.webp"), StorageKey: "derivatives/feed.webp", Preset: "card", PresetVersion: 1}
			require.NoError(t, db.Create(&derivative).Error)
			path, expectedState := asset.URL, "active"
			switch scenario {
			case "derivative":
				path = derivative.URL + "?historical=1#preview"
			case "trusted-alias", "private-alias", "url-only-registration":
				path = "https://shop.example.test/uploads/" + key
				if scenario == "url-only-registration" {
					require.NoError(t, db.Model(asset).Update("storage_key", "").Error)
				}
			case "soft-deleted-parent":
				expectedState = "retained_deleted_parent"
			case "soft-deleted-placement":
				expectedState = "retained_deleted_placement"
			case "draft", "archived":
				expectedState = scenario
			case "missing-parent", "missing-parent-table":
				expectedState = "missing_parent"
			case "foreign-host":
				path = "https://foreign.example.test/uploads/" + key
			case "longer-path":
				path += ".other.png"
			case "embedded-prose":
				path = "read " + path + " please"
			}
			if scenario != "missing-parent-table" && scenario != "missing-parent" {
				parent := map[string]any{"id": 71, "status": "published"}
				if scenario == "draft" || scenario == "archived" {
					parent["status"] = scenario
				}
				if scenario == "soft-deleted-parent" {
					parent["deleted_at"] = time.Now().UTC()
				}
				require.NoError(t, db.Table("workbench_feed_entries").Create(parent).Error)
			}
			id := 1
			if scenario == "pagination" {
				for index := 1; index <= 256; index++ {
					require.NoError(t, db.Table("workbench_feed_media").Create(map[string]any{"id": index, "entry_id": 71, "file_path": "https://external.example.test/unrelated.png"}).Error)
				}
				id = 257
			}
			row := map[string]any{"id": id, "entry_id": 71, "file_path": path}
			if scenario == "soft-deleted-placement" {
				row["deleted_at"] = time.Now().UTC()
			}
			require.NoError(t, db.Table("workbench_feed_media").Create(row).Error)
			report, err := assets.GetAssetReferences(asset.ID)
			require.NoError(t, err)
			if scenario == "foreign-host" || scenario == "longer-path" || scenario == "embedded-prose" {
				require.Zero(t, report.Total)
				return
			}
			require.Equal(t, 1, report.Total)
			require.Equal(t, expectedState, report.References[0].UsageState)
			require.Equal(t, "workbench_feed_media", report.References[0].ResourceType)
			require.EqualValues(t, 71, report.References[0].ParentResourceID)
			require.Equal(t, "file_path", report.References[0].Field)
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
		})
	}
}

func TestWorkbenchFeedRetainedHistoryBlocksOriginalSharedMediaCleanupEventUntilPermanentDetachment(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	root := t.TempDir()
	assets := newMediaDeleteTestService(t, db, root)
	createWorkbenchFeedReferenceRetentionTestTables(t, db, true)
	key := "workbench-feed/delayed-cleanup.png"
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, key)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
	provider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: assets.storage, deleteError: errors.New("synthetic temporary storage outage")}
	assets.storage = provider
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
	require.NoError(t, db.Table("workbench_feed_entries").Create(map[string]any{"id": 71, "status": "published", "deleted_at": time.Now().UTC()}).Error)
	require.NoError(t, db.Table("workbench_feed_media").Create(map[string]any{"id": 1, "entry_id": 71, "file_path": asset.URL, "deleted_at": time.Now().UTC()}).Error)
	provider.deleteError = nil
	handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.NoError(t, err)
	require.NoError(t, db.Table("workbench_feed_media").Where("id = ?", 1).Delete(nil).Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.ErrorIs(t, err, os.ErrNotExist)
	var events int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_key = ?", event.EventKey).Count(&events).Error)
	require.EqualValues(t, 1, events)
}

func TestWorkbenchFeedDedicatedHistoryRetentionDoesNotGrantAccessWithoutPublishedActivePlacement(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		t.Run(namespace, func(t *testing.T) {
			db, _, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
			access := NewPublicUploadAccessService(nil, nil)
			access.ConfigureSiteLogoService(handler.siteLogo)
			access.ConfigureSiteFaviconService(handler.siteFavicon)
			access.ConfigureHomeVisualTileService(handler.homeVisual)
			createWorkbenchFeedReferenceRetentionTestTables(t, db, true)
			require.NoError(t, db.Table("workbench_feed_entries").Create(map[string]any{"id": 71, "status": "draft"}).Error)
			require.NoError(t, db.Table("workbench_feed_media").Create(map[string]any{"id": 1, "entry_id": 71, "file_path": "/uploads/" + key}).Error)
			repo := repository.NewMediaRepository(db)
			state, err := repo.FindDedicatedPublicObjectBusinessReferenceState(key)
			require.NoError(t, err)
			require.Len(t, state.RetainedReferences, 1)
			require.False(t, state.HasPublishedConsumer)
			allowed, err := access.CanServePublicUpload(t.Context(), key)
			require.NoError(t, err)
			require.False(t, allowed)
			require.Error(t, handler.Handle(t.Context(), event))
			require.NoError(t, db.Table("workbench_feed_entries").Where("id = ?", 71).Update("status", "published").Error)
			state, err = repo.FindDedicatedPublicObjectBusinessReferenceState(key)
			require.NoError(t, err)
			require.True(t, state.HasPublishedConsumer)
			allowed, err = access.CanServePublicUpload(t.Context(), key)
			require.NoError(t, err)
			require.True(t, allowed)
			for _, mutation := range []struct {
				table, column string
				value         any
			}{
				{"workbench_feed_entries", "status", "archived"},
				{"workbench_feed_entries", "status", "published"},
				{"workbench_feed_media", "deleted_at", time.Now().UTC()},
				{"workbench_feed_media", "deleted_at", nil},
				{"workbench_feed_entries", "deleted_at", time.Now().UTC()},
			} {
				require.NoError(t, db.Table(mutation.table).Where("id > ?", 0).Update(mutation.column, mutation.value).Error)
				state, err = repo.FindDedicatedPublicObjectBusinessReferenceState(key)
				require.NoError(t, err)
				require.Len(t, state.RetainedReferences, 1)
				require.Equal(t, mutation.value == "published" || (mutation.column == "deleted_at" && mutation.value == nil), state.HasPublishedConsumer)
				allowed, err = access.CanServePublicUpload(t.Context(), key)
				require.NoError(t, err)
				require.Equal(t, state.HasPublishedConsumer, allowed)
				require.Error(t, handler.Handle(t.Context(), event))
			}
			require.NoError(t, db.Table("workbench_feed_media").Where("id = ?", 1).Delete(nil).Error)
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event))
			allowed, err = access.CanServePublicUpload(t.Context(), key)
			require.NoError(t, err)
			require.False(t, allowed)
		})
	}
}

func TestWorkbenchFeedReadOnlyInventoryIncludesRetainedImagePathsWithoutExposingAttachmentValues(t *testing.T) {
	db, inventory := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	createWorkbenchFeedReferenceRetentionTestTables(t, db, false)
	require.NoError(t, db.Table("workbench_feed_media").Create(map[string]any{"id": 1, "entry_id": 71, "file_path": "/uploads/workbench-feed/unregistered-private-label.png?secret=synthetic", "deleted_at": time.Now().UTC()}).Error)
	require.NoError(t, db.Exec("PRAGMA query_only = ON").Error)
	defer db.Exec("PRAGMA query_only = OFF")
	report, err := inventory.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	var foundSource, foundIssue bool
	for _, source := range report.Sources {
		if source.Table == "workbench_feed_media" && source.Field == "file_path" {
			foundSource = true
			require.True(t, source.Complete)
			require.Equal(t, 1, source.RowsScanned)
		}
	}
	for _, issue := range report.Issues {
		if issue.Table == "workbench_feed_media" {
			foundIssue = true
			require.Equal(t, "business_managed_url_not_registered", issue.Code)
			require.True(t, issue.RetainedDeletedRecord)
		}
	}
	require.True(t, foundSource)
	require.True(t, foundIssue)
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "unregistered-private-label")
	require.NotContains(t, string(encoded), "secret=synthetic")
}

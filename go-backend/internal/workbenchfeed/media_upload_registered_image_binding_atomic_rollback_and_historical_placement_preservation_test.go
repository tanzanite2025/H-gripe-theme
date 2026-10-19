package workbenchfeed

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	mediaservice "commerce-platform/internal/service"
)

func createWorkbenchFeedMediaBindingTestServices(t *testing.T, db *gorm.DB) (*gorm.DB, *Service, *mediaservice.MediaService, string) {
	t.Helper()
	if db == nil {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		connection, err := db.DB()
		require.NoError(t, err)
		connection.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = connection.Close() })
	}
	require.NoError(t, db.AutoMigrate(&FeedEntry{}, &FeedMedia{}, &TaggedProduct{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{}, &outbox.Event{}))
	root := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	assets := mediaservice.NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return db, NewService(db, NewRepository(db), NewCatalog(db), assets), assets, root
}

func createWorkbenchFeedRegisteredImageAndPhysicalObject(t *testing.T, db *gorm.DB, root, key string) *media.MediaAsset {
	t.Helper()
	asset := &media.MediaAsset{URL: "https://media.example.test/uploads/" + key, StorageKey: key, Filename: filepath.Base(key), MediaType: "image", Status: "active", Visibility: "public", UploadOrigin: media.MediaUploadOriginAdministratorShared, LibraryPartition: media.MediaLibraryPartitionShared, ClassificationMethod: media.MediaClassificationRecordedUploadContext}
	require.NoError(t, db.Create(asset).Error)
	path := filepath.Join(root, filepath.FromSlash(key))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("synthetic image object"), 0o644))
	return asset
}

func workbenchFeedEntryInputWithImagePaths(paths ...string) EntryInput {
	input := EntryInput{Content: "Synthetic feed entry", Status: StatusDraft, ActorID: 42}
	for index, path := range paths {
		input.Media = append(input.Media, MediaInput{FilePath: path, Width: 64, Height: 32, FileSizeKB: 1, SortOrder: index})
	}
	return input
}

func TestWorkbenchFeedImageUploadRegistrationBindingAndRetirementProtection(t *testing.T) {
	db, feed, assets, root := createWorkbenchFeedMediaBindingTestServices(t, nil)
	var content bytes.Buffer
	require.NoError(t, png.Encode(&content, image.NewRGBA(image.Rect(0, 0, 64, 32))))
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "feed.png")
	require.NoError(t, err)
	_, err = part.Write(content.Bytes())
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequestWithContext(t.Context(), "POST", "/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(1<<20))
	t.Cleanup(func() { _ = request.MultipartForm.RemoveAll() })
	upload, err := feed.UploadMedia(t.Context(), 42, request.MultipartForm.File["file"][0], "Uploaded")
	require.NoError(t, err)
	var asset media.MediaAsset
	require.NoError(t, db.Where("url = ?", upload.FilePath).First(&asset).Error)
	require.Equal(t, media.MediaUploadOriginAdministratorWorkbenchFeed, asset.UploadOrigin)
	require.Equal(t, media.MediaLibraryPartitionContent, asset.LibraryPartition)
	require.Equal(t, media.MediaUploadActorAdministrator, asset.UploaderActorType)
	require.EqualValues(t, 42, asset.UploaderID)
	entry, err := feed.Create(workbenchFeedEntryInputWithImagePaths(upload.FilePath))
	require.NoError(t, err)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "workbench_feed_media", report.References[0].ResourceType)
	require.Equal(t, entry.ID, report.References[0].ParentResourceID)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, mediaservice.MediaAssetDeleteConfirmation(asset.ID)), mediaservice.ErrMediaAssetInUse)
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(asset.StorageKey)))
	require.NoError(t, err)
}

func TestWorkbenchFeedNewInvalidImageBindingsRollBackCreateAndEntireUpdate(t *testing.T) {
	for _, mutation := range []string{"create", "update"} {
		for _, scenario := range []string{"unregistered", "private", "conversation", "retired", "inactive", "video", "duplicate-registration", "dedicated-logo", "dedicated-favicon", "dedicated-showcase", "dedicated-registered"} {
			t.Run(mutation+"/"+scenario, func(t *testing.T) {
				db, feed, _, root := createWorkbenchFeedMediaBindingTestServices(t, nil)
				asset := createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "content/selected.png")
				path := asset.URL
				switch scenario {
				case "unregistered":
					path = "/uploads/content/missing.png"
				case "private":
					require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
				case "conversation":
					path = "/uploads/customer-service/conversations/private.png"
				case "retired":
					require.NoError(t, db.Delete(asset).Error)
				case "inactive":
					require.NoError(t, db.Model(asset).Update("status", "disabled").Error)
				case "video":
					require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
				case "duplicate-registration":
					duplicate := createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "content/duplicate.png")
					require.NoError(t, db.Model(duplicate).Update("storage_key", asset.StorageKey).Error)
				case "dedicated-logo", "dedicated-registered":
					path = "/uploads/site-logo/historical.png"
					if scenario == "dedicated-registered" {
						createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "site-logo/historical.png")
					}
				case "dedicated-favicon":
					path = "/uploads/site-favicon/historical.png"
				case "dedicated-showcase":
					path = "/uploads/visual-showcase/historical.png"
				}
				var parent *FeedEntry
				if mutation == "update" {
					var err error
					parent, err = feed.Create(workbenchFeedEntryInputWithImagePaths("https://external.example.test/before.png"))
					require.NoError(t, err)
				}
				input := workbenchFeedEntryInputWithImagePaths("https://external.example.test/valid.png", path)
				input.Content, input.Status = "Must roll back", StatusPublished
				var err error
				if mutation == "create" {
					_, err = feed.Create(input)
				} else {
					_, err = feed.Update(parent.ID, input)
				}
				require.ErrorIs(t, err, ErrMediaReferenceInvalid)
				var entries, placements int64
				require.NoError(t, db.Unscoped().Model(&FeedEntry{}).Count(&entries).Error)
				require.NoError(t, db.Unscoped().Model(&FeedMedia{}).Count(&placements).Error)
				if mutation == "create" {
					require.Zero(t, entries)
					require.Zero(t, placements)
				} else {
					require.EqualValues(t, 1, entries)
					require.EqualValues(t, 1, placements)
					stored, err := feed.Get(parent.ID)
					require.NoError(t, err)
					require.Equal(t, parent.Content, stored.Content)
					require.Equal(t, parent.Status, stored.Status)
					require.Equal(t, parent.Media[0].ID, stored.Media[0].ID)
				}
			})
		}
	}
}

func TestWorkbenchFeedDerivativeAndTrustedAliasBindingsCanonicalizeAndRetainReplacedHistory(t *testing.T) {
	db, feed, assets, root := createWorkbenchFeedMediaBindingTestServices(t, nil)
	asset := createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "content/canonical.png")
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, URL: "https://media.example.test/uploads/derivatives/card.webp", StorageKey: "derivatives/card.webp", Preset: "card", PresetVersion: 1}
	require.NoError(t, db.Create(&derivative).Error)
	entry, err := feed.Create(workbenchFeedEntryInputWithImagePaths("https://shop.example.test/uploads/derivatives/card.webp?revision=1#preview"))
	require.NoError(t, err)
	require.Equal(t, asset.URL+"#preview", entry.Media[0].FilePath)
	entry, err = feed.Update(entry.ID, workbenchFeedEntryInputWithImagePaths("https://external.example.test/next.png"))
	require.NoError(t, err)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "retained_deleted_placement", report.References[0].UsageState)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, mediaservice.MediaAssetDeleteConfirmation(asset.ID)), mediaservice.ErrMediaAssetInUse)
	for _, status := range []string{StatusPublished, StatusArchived, StatusDraft} {
		entry, err = feed.UpdateStatus(entry.ID, status, 81)
		require.NoError(t, err)
		require.Equal(t, status, entry.Status)
		require.EqualValues(t, 81, entry.UpdatedBy)
		require.Len(t, entry.Media, 1)
		require.Equal(t, "https://external.example.test/next.png", entry.Media[0].FilePath)
	}
	var allPlacements int64
	require.NoError(t, db.Unscoped().Model(&FeedMedia{}).Count(&allPlacements).Error)
	require.EqualValues(t, 2, allPlacements, "status changes cannot rewrite associations")
	require.NoError(t, feed.Delete(entry.ID))
	_, err = feed.UpdateStatus(entry.ID, StatusPublished, 81)
	require.ErrorIs(t, err, ErrEntryNotFound)
	_, err = feed.Update(entry.ID, workbenchFeedEntryInputWithImagePaths(asset.URL))
	require.ErrorIs(t, err, ErrEntryNotFound)
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, "retained_deleted_parent", report.References[0].UsageState)
}

func TestWorkbenchFeedExactHistoricalPathsCanBePreservedButCannotExpandOrReturnAfterDetachment(t *testing.T) {
	for _, path := range []string{"/uploads/site-logo/historical.png", "/uploads/content/unregistered.png"} {
		t.Run(fmt.Sprint(path), func(t *testing.T) {
			db, feed, _, _ := createWorkbenchFeedMediaBindingTestServices(t, nil)
			entry, err := feed.Create(workbenchFeedEntryInputWithImagePaths())
			require.NoError(t, err)
			require.NoError(t, db.Create(&FeedMedia{EntryID: entry.ID, FilePath: path, Width: 64, Height: 32}).Error)
			input := workbenchFeedEntryInputWithImagePaths(path)
			input.Content = "Edit exact historical placement"
			_, err = feed.Update(entry.ID, input)
			require.NoError(t, err)
			_, err = feed.Update(entry.ID, workbenchFeedEntryInputWithImagePaths(path, path))
			require.ErrorIs(t, err, ErrMediaReferenceInvalid)
			_, err = feed.Update(entry.ID, workbenchFeedEntryInputWithImagePaths())
			require.NoError(t, err)
			_, err = feed.Update(entry.ID, input)
			require.ErrorIs(t, err, ErrMediaReferenceInvalid, "retained deleted rows do not authorize a new binding")
		})
	}
}

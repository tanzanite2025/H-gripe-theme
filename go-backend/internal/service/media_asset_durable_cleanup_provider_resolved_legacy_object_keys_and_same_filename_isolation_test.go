package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func TestMediaDurableCleanupResolvesLegacyOriginalFullProviderKeyAndPreservesOtherObjectsWithSameFilename(t *testing.T) {
	for _, provider := range []storage.StorageType{storage.StorageTypeLocal, storage.StorageTypeS3, storage.StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			originalKey := "catalog/2024/old/shared.png"
			neighborKeys := []string{"shared.png", "catalog/another/shared.png"}
			publicRoot := t.TempDir()
			var remoteObjectsMutex sync.Mutex
			remoteObjects := map[string]bool{originalKey: true, neighborKeys[0]: true, neighborKeys[1]: true}
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete || !strings.HasPrefix(r.URL.Path, "/media-public/") {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				remoteObjectsMutex.Lock()
				delete(remoteObjects, strings.TrimPrefix(r.URL.Path, "/media-public/"))
				remoteObjectsMutex.Unlock()
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(endpoint.Close)
			store, err := storage.NewStorageService(&storage.Config{Type: provider, LocalPath: publicRoot,
				BaseURL: "https://media.example.test/assets", Bucket: "media-public", Region: "us-east-1", Endpoint: endpoint.URL,
				AccessKeyID: "legacy-cleanup-test-key", SecretAccessKey: "legacy-cleanup-test-secret"})
			require.NoError(t, err)
			if provider == storage.StorageTypeLocal {
				for _, key := range append([]string{originalKey}, neighborKeys...) {
					objectPath := filepath.Join(publicRoot, filepath.FromSlash(key))
					require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
					require.NoError(t, os.WriteFile(objectPath, []byte(key), 0o644))
				}
			}
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &outbox.Event{})
			asset := media.MediaAsset{Filename: "shared.png", OriginalFilename: "shared.png", URL: store.GetURL(originalKey),
				MediaType: "image", MimeType: "image/png", Size: 10, Status: "active", Visibility: "public"}
			require.NoError(t, db.Create(&asset).Error)
			mediaService := NewMediaService(repository.NewMediaRepository(db), store, nil, "", 20<<30)
			mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			require.NoError(t, mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
			var event outbox.Event
			require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
			var payload outbox.ObjectStorageCleanupPayload
			require.NoError(t, json.Unmarshal(event.Payload, &payload))
			require.Equal(t, []string{originalKey}, payload.ObjectKeys)
			handler := NewObjectStorageCleanupOutboxHandler(mediaService, nil, nil, nil)
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event))
			if provider == storage.StorageTypeLocal {
				_, err = os.Stat(filepath.Join(publicRoot, filepath.FromSlash(originalKey)))
				require.True(t, os.IsNotExist(err))
				for _, key := range neighborKeys {
					_, err = os.Stat(filepath.Join(publicRoot, filepath.FromSlash(key)))
					require.NoError(t, err)
				}
			} else {
				remoteObjectsMutex.Lock()
				remainingObjects := make(map[string]bool, len(remoteObjects))
				for key, exists := range remoteObjects {
					remainingObjects[key] = exists
				}
				remoteObjectsMutex.Unlock()
				require.Equal(t, map[string]bool{neighborKeys[0]: true, neighborKeys[1]: true}, remainingObjects)
			}
		})
	}
}

func TestMediaDurableCleanupRejectsUnresolvableLegacyURLBeforeRetiringAssetOrDeletingSameFilename(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	publicRoot := t.TempDir()
	mediaService := newMediaDeleteTestService(t, db, publicRoot)
	asset := media.MediaAsset{Filename: "shared.png", OriginalFilename: "shared.png", URL: "https://untrusted.example.test/shared.png",
		MediaType: "image", MimeType: "image/png", Size: 10, Status: "active", Visibility: "public"}
	require.NoError(t, db.Create(&asset).Error)
	neighborPath := filepath.Join(publicRoot, "shared.png")
	require.NoError(t, os.WriteFile(neighborPath, []byte("unrelated file"), 0o644))
	require.Error(t, mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	_, err := mediaService.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	_, err = os.Stat(neighborPath)
	require.NoError(t, err)
	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&eventCount).Error)
	require.Zero(t, eventCount)
}

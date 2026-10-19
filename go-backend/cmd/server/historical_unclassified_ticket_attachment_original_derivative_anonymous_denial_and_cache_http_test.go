package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestHistoricalUnclassifiedTicketAttachmentOriginalAndDerivativeAnonymousHTTPReadsAreNotCacheable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("STORAGE_TYPE", "local")
	t.Setenv("STORAGE_LOCAL_PATH", root)
	t.Setenv("STORAGE_BASE_URL", "https://shop.example.test")
	t.Setenv("SITE_LOGO_STORAGE_TYPE", "local")
	t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", t.TempDir())
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &ticket.TicketMessage{}))
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	asset := media.MediaAsset{Filename: "historical.png", URL: "http://localhost:9200/uploads/2026/08/20/historical.png", StorageKey: "2026/08/20/historical.png", MimeType: "image/png"}
	require.NoError(t, db.Create(&asset).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", PresetVersion: 1, StorageKey: "media-derivatives/historical-thumbnail.png", URL: "http://localhost:9200/uploads/media-derivatives/historical-thumbnail.png", MimeType: "image/png"}
	require.NoError(t, db.Create(&derivative).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 999, Attachments: `["http://localhost:9199/uploads/` + asset.StorageKey + `"]`}).Error)
	other := media.MediaAsset{Filename: "shared.png", URL: "https://shop.example.test/uploads/shared/shared.png", StorageKey: "shared/shared.png", UploadOrigin: media.MediaUploadOriginAdministratorShared}
	require.NoError(t, db.Create(&other).Error)
	payload := []byte("same physical bytes in separate registered keys")
	for _, key := range []string{asset.StorageKey, derivative.StorageKey, other.StorageKey} {
		filename := filepath.Join(root, filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, payload, 0600))
	}
	assets := service.NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
	deps := &app.Dependencies{Services: app.Services{PublicUploadAccess: service.NewPublicUploadAccessService(assets, nil)}}
	router := gin.New()
	registerLocalUploadsRoute(router, deps)
	for _, key := range []string{asset.StorageKey, derivative.StorageKey, other.StorageKey} {
		response := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/uploads/"+key, nil)
		request.Header.Set("Range", "bytes=0-3")
		router.ServeHTTP(response, request)
		if key == other.StorageKey {
			require.Equal(t, http.StatusPartialContent, response.Code, response.Body.String())
			require.Equal(t, payload[:4], response.Body.Bytes())
		} else {
			require.Equal(t, http.StatusNotFound, response.Code, response.Body.String())
			require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
			require.NotContains(t, response.Body.String(), string(payload))
		}
		stored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(key)))
		require.NoError(t, err)
		require.Equal(t, payload, stored)
	}
}

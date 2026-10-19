package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/api/admin"
	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLocalUploadRouteServesConfiguredVisualShowcaseOriginalAndThumbnailAndRejectsInvalidNewBindings(t *testing.T) {
	root := t.TempDir()
	t.Setenv("STORAGE_TYPE", "local")
	t.Setenv("STORAGE_LOCAL_PATH", root)
	t.Setenv("STORAGE_BASE_URL", "https://media.example.test")
	t.Setenv("SITE_LOGO_STORAGE_TYPE", "local")
	t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", t.TempDir())
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&homevisualtile.Tile{}, &outbox.Event{}))
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 600, 600))))
	imageURL, err := provider.UploadFromReaderWithPrefix(t.Context(), bytes.NewReader(encoded.Bytes()), "original.png", "visual-showcase/home-hero/en")
	require.NoError(t, err)
	thumbnailURL, err := provider.UploadFromReaderWithPrefix(t.Context(), bytes.NewReader(encoded.Bytes()), "thumbnail.png", "visual-showcase/home-hero/en")
	require.NoError(t, err)
	imageKey, err := provider.ObjectKey(imageURL)
	require.NoError(t, err)
	thumbnailKey, err := provider.ObjectKey(thumbnailURL)
	require.NoError(t, err)
	visuals := service.NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), provider)
	visuals.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	access := service.NewPublicUploadAccessService(nil, nil)
	access.ConfigureHomeVisualTileService(visuals)
	deps := &app.Dependencies{Services: app.Services{PublicUploadAccess: access}}
	router := gin.New()
	registerLocalUploadsRoute(router, deps)
	handler := admin.NewHomeVisualTileHandler(visuals)
	router.PUT("/test/showcases/:showcase_key/slots/:desktop_order", handler.SaveSingleVisualShowcaseItem)
	save := func(thumbnail string) *httptest.ResponseRecorder {
		body, err := json.Marshal(map[string]any{"locale": "en", "item": map[string]any{
			"image_url": imageURL, "thumbnail_url": thumbnail, "storage_key": imageKey,
			"title": "Hero", "alt_text": "Hero image", "width": 600, "height": 600,
		}})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/test/showcases/home-hero/slots/1", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	get := func(key string, status int) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/uploads/"+key, nil))
		require.Equal(t, status, response.Code, response.Body.String())
		if status == http.StatusOK {
			require.Equal(t, encoded.Bytes(), response.Body.Bytes())
		}
	}
	get(imageKey, http.StatusNotFound)
	get(thumbnailKey, http.StatusNotFound)
	response := save(thumbnailURL)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	get(imageKey, http.StatusOK)
	get(thumbnailKey, http.StatusOK)
	for _, invalidThumbnail := range []string{"https://foreign.example.test/uploads/" + thumbnailKey, "/uploads/visual-showcase/home-hero/en/missing.png"} {
		response := save(invalidThumbnail)
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		get(imageKey, http.StatusOK)
		get(thumbnailKey, http.StatusOK)
	}
	_, err = visuals.ReplaceAdminItems(t.Context(), "home-hero", "en", nil)
	require.NoError(t, err)
	get(imageKey, http.StatusNotFound)
	get(thumbnailKey, http.StatusNotFound)
}

package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"commerce-platform/internal/testfixtures"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createHistoricalDedicatedObjectReplacementMultipartWebPHeader(t *testing.T) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "replacement.webp")
	require.NoError(t, err)
	_, err = part.Write(testfixtures.SiteLogoAndFaviconUploadLifecycle512By512LosslessWebPFixtureBytes())
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test-upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(2<<20))
	t.Cleanup(func() { _ = request.MultipartForm.RemoveAll() })
	return request.MultipartForm.File["image"][0]
}

func TestHistoricalDedicatedUploadsRemainReadableAfterReplacementOnlyWhilePublishedConsumersExistAndRetryOriginalCleanup(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		t.Run(namespace, func(t *testing.T) {
			sharedRoot, logoRoot := t.TempDir(), t.TempDir()
			t.Setenv("STORAGE_TYPE", "local")
			t.Setenv("STORAGE_LOCAL_PATH", sharedRoot)
			t.Setenv("STORAGE_BASE_URL", "https://shared.example.test")
			t.Setenv("SITE_LOGO_STORAGE_TYPE", "local")
			t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", logoRoot)
			t.Setenv("SITE_LOGO_STORAGE_BASE_URL", "https://logos.example.test")
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}, &setting.Setting{},
				&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &homevisualtile.Tile{}, &post.Post{}, &product.Product{}, &product.ProductMedia{}))
			root, baseURL := sharedRoot, "https://shared.example.test"
			if namespace == "site-logo" {
				root, baseURL = logoRoot, "https://logos.example.test"
			}
			provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: baseURL})
			require.NoError(t, err)
			key := namespace + "/historical/old.png"
			objectPath := filepath.Join(root, filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
			originalBytes := []byte("historical object from " + namespace + " owner storage")
			require.NoError(t, os.WriteFile(objectPath, originalBytes, 0o644))
			if namespace == "site-logo" {
				decoyPath := filepath.Join(sharedRoot, filepath.FromSlash(key))
				require.NoError(t, os.MkdirAll(filepath.Dir(decoyPath), 0o755))
				require.NoError(t, os.WriteFile(decoyPath, []byte("same key in a different physical store"), 0o644))
			}
			settings := service.NewSettingService(repository.NewSettingRepository(db), nil, 0)
			logos := service.NewSiteLogoService(repository.NewSiteLogoRepository(db), provider, "https://shop.example.test")
			favicons := service.NewSiteFaviconService(repository.NewSiteFaviconRepository(db), provider, "https://shop.example.test")
			visuals := service.NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), provider, "https://shop.example.test")
			logos.ConfigureSettingService(settings)
			favicons.ConfigureSettingService(settings)
			logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			favicons.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			visuals.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			switch namespace {
			case "site-logo":
				err = db.Create(&sitelogodomain.Asset{ID: sitelogodomain.CurrentAssetID, StorageKey: key, URL: provider.GetURL(key)}).Error
			case "site-favicon":
				err = db.Create(&sitefavicondomain.Asset{ID: sitefavicondomain.CurrentAssetID, StorageKey: key, URL: provider.GetURL(key)}).Error
			case "visual-showcase":
				err = db.Create(&homevisualtile.Tile{TileSetKey: "home-hero", Locale: "en", ImageURL: "/uploads/" + key, ThumbnailURL: "/uploads/" + key, StorageKey: key, Title: "Old hero", AltText: "Old hero", DesktopOrder: 1}).Error
			}
			require.NoError(t, err)
			access := service.NewPublicUploadAccessService(nil, nil)
			access.ConfigureSiteLogoService(logos)
			access.ConfigureSiteFaviconService(favicons)
			access.ConfigureHomeVisualTileService(visuals)
			router := gin.New()
			registerLocalUploadsRoute(router, &app.Dependencies{Services: app.Services{PublicUploadAccess: access}})
			get := func(status int) {
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/uploads/"+key, nil))
				require.Equal(t, status, response.Code, response.Body.String())
				if status == http.StatusOK {
					require.Equal(t, originalBytes, response.Body.Bytes(), "serve bytes from the owning store")
				}
			}
			get(http.StatusOK)
			article := post.Post{Title: "Published historical image", Slug: "historical-image", Status: "published", Content: `<img src="https://shop.example.test/uploads/` + key + `">`}
			require.NoError(t, db.Create(&article).Error)
			item := product.Product{Name: "Published product", Slug: "historical-product", Status: "active"}
			require.NoError(t, db.Create(&item).Error)
			placement := product.ProductMedia{ProductID: item.ID, URL: "/static/unrelated.png", ThumbnailURL: "/uploads/" + key, IsVisible: true}
			require.NoError(t, db.Create(&placement).Error)
			switch namespace {
			case "site-logo":
				_, err = logos.UploadCurrent(t.Context(), createHistoricalDedicatedObjectReplacementMultipartWebPHeader(t), 11)
			case "site-favicon":
				_, err = favicons.UploadCurrent(t.Context(), createHistoricalDedicatedObjectReplacementMultipartWebPHeader(t), 11)
			case "visual-showcase":
				_, err = visuals.ReplaceAdminItems(t.Context(), "home-hero", "en", nil)
			}
			require.NoError(t, err)
			get(http.StatusOK)
			cleanup := service.NewObjectStorageCleanupOutboxHandler(nil, logos, visuals, nil)
			cleanup.ConfigureSiteFaviconService(favicons)
			worker := service.NewOutboxService(repository.NewOutboxRepository(db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, cleanup.Handle)
			result, err := worker.ProcessPending(t.Context(), time.Now().Add(time.Second), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Failed)
			var event outbox.Event
			require.NoError(t, db.First(&event).Error)
			require.Equal(t, outbox.EventStatusFailed, event.Status)
			require.NoError(t, db.Model(&article).Update("status", "draft").Error)
			get(http.StatusOK) // The visible product still authorizes its thumbnail.
			require.NoError(t, db.Model(&placement).Update("is_visible", false).Error)
			get(http.StatusNotFound)
			require.ErrorIs(t, cleanup.Handle(t.Context(), event), service.ErrMediaAssetInUse)
			require.NoError(t, db.Model(&article).Update("status", "published").Error)
			get(http.StatusOK)
			require.NoError(t, db.Model(&article).Update("content", "<p>Detached</p>").Error)
			get(http.StatusNotFound)
			require.ErrorIs(t, cleanup.Handle(t.Context(), event), service.ErrMediaAssetInUse, "hidden product still retains the file")
			require.NoError(t, db.Model(&placement).Update("thumbnail_url", "").Error)
			get(http.StatusNotFound)
			result, err = worker.ProcessPending(t.Context(), event.AvailableAt.Add(time.Second), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Processed)
			_, err = os.Stat(objectPath)
			require.True(t, os.IsNotExist(err))
			var events []outbox.Event
			require.NoError(t, db.Find(&events).Error)
			require.Len(t, events, 1)
			require.Equal(t, event.ID, events[0].ID)
			require.Equal(t, outbox.EventStatusProcessed, events[0].Status)
			get(http.StatusNotFound)
		})
	}
}

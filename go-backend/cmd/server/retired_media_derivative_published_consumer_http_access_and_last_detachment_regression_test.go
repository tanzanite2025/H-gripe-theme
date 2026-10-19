package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRetiredMediaDerivativeHTTPReadRetainsPublishedConsumersAndRevokesAfterLastDetachment(t *testing.T) {
	for _, source := range []string{"product-html", "variant-snapshot", "draft-only", "historical-ticket-only", "ambiguous-owner"} {
		t.Run(source, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("STORAGE_TYPE", "local")
			t.Setenv("STORAGE_LOCAL_PATH", root)
			t.Setenv("STORAGE_BASE_URL", "https://shop.example.test")
			t.Setenv("SITE_LOGO_STORAGE_TYPE", "local")
			t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", t.TempDir())
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{}, &product.ProductMedia{}, &ticket.TicketMessage{}))
			provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://shop.example.test"})
			require.NoError(t, err)
			parent := media.MediaAsset{Filename: "original.png", StorageKey: "catalog/original.png", URL: provider.GetURL("catalog/original.png"), MimeType: "image/png", UploadOrigin: media.MediaUploadOriginAdministratorShared}
			require.NoError(t, db.Create(&parent).Error)
			key := "media-derivatives/1/card/v1/history.png"
			previous := media.MediaAssetDerivative{MediaAssetID: parent.ID, Preset: "card", PresetVersion: 1, StorageKey: key, URL: provider.GetURL(key)}
			require.NoError(t, db.Create(&previous).Error)
			require.NoError(t, db.Delete(&previous).Error)
			path := filepath.Join(root, filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			payload := []byte("retained historical derivative bytes")
			require.NoError(t, os.WriteFile(path, payload, 0o644))
			owner := product.Product{SKU: "published", Slug: "published", Name: "Published", Status: "active", Description: `<img src="` + previous.URL + `">`}
			if source == "draft-only" {
				owner.Status = "draft"
			}
			if source != "historical-ticket-only" {
				require.NoError(t, db.Create(&owner).Error)
			}
			if source == "variant-snapshot" {
				require.NoError(t, db.Model(&owner).Update("description", "").Error)
				require.NoError(t, db.Create(&product.ProductMedia{ProductID: owner.ID, MediaType: "image", URL: parent.URL, IsVisible: true,
					ImageVariantData: product.ProductMediaImageVariantsJSON(map[string]product.ProductMediaImageVariant{"card": {URL: previous.URL}})}).Error)
			}
			if source == "historical-ticket-only" {
				require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 99, Attachments: `["` + previous.URL + `"]`}).Error)
			}
			if source == "ambiguous-owner" {
				require.NoError(t, db.Create(&media.MediaAsset{Filename: "duplicate", StorageKey: key, URL: previous.URL + "?duplicate=1"}).Error)
			}
			assets := service.NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
			router := gin.New()
			registerLocalUploadsRoute(router, &app.Dependencies{Services: app.Services{PublicUploadAccess: service.NewPublicUploadAccessService(assets, nil)}})
			get := func(expected int) {
				response := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, "/uploads/"+key, nil)
				request.Header.Set("Range", "bytes=0-7")
				router.ServeHTTP(response, request)
				require.Equal(t, expected, response.Code, response.Body.String())
				if expected == http.StatusPartialContent {
					require.Equal(t, payload[:8], response.Body.Bytes())
				} else {
					require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
				}
			}
			if source == "product-html" || source == "variant-snapshot" {
				get(http.StatusPartialContent)
			} else {
				get(http.StatusNotFound)
			}
			require.NoError(t, db.Model(&product.Product{}).Where("id = ?", owner.ID).Update("description", "Detached").Error)
			require.NoError(t, db.Model(&product.ProductMedia{}).Where("product_id = ?", owner.ID).Update("image_variants", "{}").Error)
			get(http.StatusNotFound)
			require.FileExists(t, path, "access revocation does not itself delete the object")
		})
	}
}

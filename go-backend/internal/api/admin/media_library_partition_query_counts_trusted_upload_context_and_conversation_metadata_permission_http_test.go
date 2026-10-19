package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func createMediaLibraryPartitionHTTPTestEnvironment(t *testing.T) (*gorm.DB, *MediaHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}))
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://shop.example.test", 20<<30)
	mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return db, NewMediaHandler(mediaService)
}

func createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler *MediaHandler, role string) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Set("user_role", role)
		c.Set("role", role)
		c.Next()
	})
	imageDimensionsHandler := NewMediaImageDimensionsHandler(service.NewMediaImageDimensionEngine(handler.mediaService))
	registerMediaAndPreflightRoutes(router.Group("/api/admin"), handler, imageDimensionsHandler, nil, nil, nil)
	return router
}

func TestMediaLibraryConversationMetadataCannotLeakThroughImagePreflightListCountsOrReconcileHTTP(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	assets := []media.MediaAsset{
		{Filename: "catalog.png", URL: "https://shop.example.test/uploads/catalog/product-images/catalog.png", LibraryPartition: media.MediaLibraryPartitionCatalog},
		{Filename: "customer.png", URL: "https://shop.example.test/uploads/customer-service/public-conversation-attachments/customer.png", LibraryPartition: media.MediaLibraryPartitionCustomerService, UploadOrigin: media.MediaUploadOriginCustomerServiceConversation, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment},
		{Filename: "old-private.png", URL: "https://private.example.test/customer-service/attachments/originals/old-private.png", Visibility: "private"},
	}
	require.NoError(t, db.Create(&assets).Error)
	for _, role := range []string{"editor", "support", "manager"} {
		router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/preflight/image-dimensions?state=all", nil))
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		require.NotContains(t, recorder.Body.String(), "customer-service/")
		require.NotContains(t, recorder.Body.String(), "customer.png")
		var response struct {
			Data struct {
				Items   []service.MediaImageDimensionFinding
				Summary service.MediaImageDimensionSummary
			}
			Pagination struct{ Total int64 }
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		require.Len(t, response.Data.Items, 1)
		require.EqualValues(t, 1, response.Data.Summary.Total)
		require.EqualValues(t, 1, response.Pagination.Total)
		for _, asset := range assets[1:] {
			recorder = httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("/api/admin/preflight/image-dimensions/%d/reconcile", asset.ID), nil))
			require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
			require.NotContains(t, recorder.Body.String(), "customer-service/")
		}
	}
}

func createMediaLibraryClassificationForgeryMultipartRequest(t *testing.T, path string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for field, value := range map[string]string{"media_type": "image", "image_purpose": "product_image", "library_partition": "customer_service",
		"upload_origin": "customer_service_conversation_upload", "classification_method": "historical_unknown", "customer_service_media_purpose": "conversation_attachment"} {
		require.NoError(t, writer.WriteField(field, value))
	}
	file, err := writer.CreateFormFile("file", "new-product.png")
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	require.NoError(t, writer.Close())
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestMediaLibraryAuthorizedUploadRouteDeterminesOriginAndOwnershipDespiteClientClassificationForgery(t *testing.T) {
	for _, route := range []struct {
		path, role string
		status     int
		partition  media.MediaAssetLibraryPartition
		origin     media.MediaAssetUploadOrigin
	}{
		{"/api/admin/media/uploads/catalog", "editor", http.StatusCreated, media.MediaLibraryPartitionCatalog, media.MediaUploadOriginAdministratorCatalog},
		{"/api/admin/media/uploads/catalog", "support", http.StatusForbidden, "", ""},
		{"/api/admin/media/uploads/website-profile", "editor", http.StatusForbidden, "", ""},
		{"/api/admin/media/assets", "editor", http.StatusCreated, media.MediaLibraryPartitionUnclassified, media.MediaUploadOriginAdministratorGeneral},
	} {
		t.Run(route.path+route.role, func(t *testing.T) {
			db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			recorder := httptest.NewRecorder()
			createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, route.role).ServeHTTP(recorder, createMediaLibraryClassificationForgeryMultipartRequest(t, route.path))
			require.Equal(t, route.status, recorder.Code, recorder.Body.String())
			var assets []media.MediaAsset
			require.NoError(t, db.Find(&assets).Error)
			if route.status == http.StatusCreated {
				require.Len(t, assets, 1)
				require.Equal(t, route.partition, assets[0].LibraryPartition)
				require.Equal(t, route.origin, assets[0].UploadOrigin)
				require.Empty(t, assets[0].CustomerServiceMediaPurpose)
			} else {
				require.Empty(t, assets)
			}
		})
	}
}

func TestMediaLibraryCountsPaginationAndDirectMetadataRespectConversationScopeAndNeverExposeDeliveryKeys(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	assets := []media.MediaAsset{
		{Filename: "product.png", URL: "https://shop.example.test/uploads/product.png", LibraryPartition: media.MediaLibraryPartitionCatalog, Visibility: "public", Status: "active"},
		{Filename: "legacy.png", URL: "https://shop.example.test/uploads/legacy.png", Visibility: "public", Status: "active"},
		{Filename: "customer.png", URL: "https://shop.example.test/uploads/customer-service/public-conversation-attachments/customer.png", LibraryPartition: media.MediaLibraryPartitionCustomerService, UploadOrigin: media.MediaUploadOriginCustomerServiceConversation, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment, Visibility: "public", Status: "active"},
		{Filename: "old-private.png", URL: "https://private.example.test/customer-service/attachments/originals/old-private.png", Visibility: "private", Status: "active"},
	}
	require.NoError(t, db.Create(&assets).Error)
	for _, role := range []string{"editor", "manager"} {
		router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/media/assets?library_partition=all&page_size=1", nil))
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		var response struct {
			Data       struct{ Assets []media.MediaAsset }
			Pagination struct{ Total int64 }
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		require.Len(t, response.Data.Assets, 1)
		if role == "editor" {
			require.EqualValues(t, 2, response.Pagination.Total)
		} else {
			require.EqualValues(t, 4, response.Pagination.Total)
		}
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/media/assets/library-partition-counts?page_size=1&library_partition=catalog", nil))
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		var counts struct {
			Data struct {
				Counts map[string]int64 `json:"partition_counts"`
			}
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &counts))
		require.Equal(t, response.Pagination.Total, counts.Data.Counts["all"])
		for _, asset := range assets[2:] {
			recorder = httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("/api/admin/media/assets/%d", asset.ID), nil))
			if role == "editor" {
				require.Equal(t, http.StatusForbidden, recorder.Code)
				continue
			}
			require.Equal(t, http.StatusOK, recorder.Code)
			require.NotContains(t, recorder.Body.String(), asset.URL)
			require.Contains(t, recorder.Body.String(), "\"requires_conversation_access\":true")
			for _, method := range []string{http.MethodPatch, http.MethodDelete} {
				recorder = httptest.NewRecorder()
				request := httptest.NewRequestWithContext(context.Background(), method, fmt.Sprintf("/api/admin/media/assets/%d", asset.ID), strings.NewReader(`{"visibility":"public","confirm":"DELETE 1"}`))
				request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(recorder, request)
				require.Equal(t, http.StatusForbidden, recorder.Code)
			}
		}
	}
}

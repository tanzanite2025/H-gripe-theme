package admin

import (
	"context"
	"fmt"
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
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestAdminProductCreateReturnsBadRequestForTransactionalMediaAssetIdentityMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&product.Product{}, &product.ProductVariant{}, &media.MediaAsset{}))
	asset := media.MediaAsset{Filename: "photo.png", URL: "https://media.example.test/uploads/catalog/photo.png", StorageKey: "catalog/photo.png"}
	require.NoError(t, db.Create(&asset).Error)
	products := service.NewProductService(repository.NewProductRepository(db), nil, 0)
	handler := NewProductHandler(products)
	router := gin.New()
	router.POST("/api/admin/products", handler.CreateProduct)
	body := fmt.Sprintf(`{"name":"Media identity","slug":"media-identity","status":"active","locale":"en","variants":[{"sku":"MEDIA-IDENTITY","price_minor":399,"is_active":true,"is_default":true}],"media":[{"media_type":"image","media_asset_id":%d,"url":"https://foreign.example.test/uploads/catalog/photo.png"}]}`, asset.ID)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/admin/products", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "media asset ID and URL identify different files")
	var count int64
	require.NoError(t, db.Model(&product.Product{}).Count(&count).Error)
	require.Zero(t, count)
}

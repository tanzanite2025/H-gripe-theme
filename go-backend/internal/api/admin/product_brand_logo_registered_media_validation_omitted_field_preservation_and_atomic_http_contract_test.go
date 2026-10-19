package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func performProductBrandLogoJSONContractRequest(t *testing.T, handler *ProductBrandHandler, id uint, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/brands", bytes.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	if id == 0 {
		handler.Create(context)
	} else {
		context.Request.Method = http.MethodPut
		context.Params = gin.Params{{Key: "id", Value: fmt.Sprint(id)}}
		handler.Update(context)
	}
	return response
}

func TestProductBrandLogoHTTPRejectsInvalidBindingsAndPreservesOmittedHistoricalLogo(t *testing.T) {
	db, _, assets := createPostManagedMediaHTTPContractTestHandler(t)
	require.NoError(t, db.AutoMigrate(&product.ProductBrand{}))
	brands := service.NewProductBrandService(repository.NewProductBrandRepository(db))
	brands.ConfigureMediaService(assets)
	handler := NewProductBrandHandler(brands)
	legacy := product.ProductBrand{Name: "Historical", Slug: "historical", LogoURL: "/uploads/site-logo/historical.png", IsEnabled: true}
	require.NoError(t, db.Create(&legacy).Error)
	for _, id := range []uint{0, legacy.ID} {
		response := performProductBrandLogoJSONContractRequest(t, handler, id, map[string]any{"name": "Invalid", "slug": "invalid", "logo_url": "/uploads/unregistered.png"})
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		var count int64
		require.NoError(t, db.Model(&product.ProductBrand{}).Count(&count).Error)
		require.EqualValues(t, 1, count)
		stored, err := brands.Get(legacy.ID)
		require.NoError(t, err)
		require.Equal(t, legacy.Name, stored.Name)
		require.Equal(t, legacy.LogoURL, stored.LogoURL)
	}
	response := performProductBrandLogoJSONContractRequest(t, handler, legacy.ID, map[string]any{"name": "Metadata edit", "slug": legacy.Slug})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	stored, err := brands.Get(legacy.ID)
	require.NoError(t, err)
	require.Equal(t, legacy.LogoURL, stored.LogoURL)
	response = performProductBrandLogoJSONContractRequest(t, handler, legacy.ID, map[string]any{"name": "Detached", "slug": legacy.Slug, "logo_url": ""})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	stored, err = brands.Get(legacy.ID)
	require.NoError(t, err)
	require.Empty(t, stored.LogoURL)
	response = performProductBrandLogoJSONContractRequest(t, handler, legacy.ID, map[string]any{"name": "Stale restore", "slug": legacy.Slug, "logo_url": legacy.LogoURL})
	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
}

func TestProductBrandLogoHTTPCanonicalizesRegisteredImagesAndKeepsForeignAndStaticURLsUnmanaged(t *testing.T) {
	db, _, assets := createPostManagedMediaHTTPContractTestHandler(t)
	require.NoError(t, db.AutoMigrate(&product.ProductBrand{}))
	brands := service.NewProductBrandService(repository.NewProductBrandRepository(db))
	brands.ConfigureMediaService(assets)
	handler := NewProductBrandHandler(brands)
	asset := media.MediaAsset{Filename: "brand.png", StorageKey: "catalog/brand.png", URL: "https://shop.example.test/uploads/catalog/brand.png", Status: "active", Visibility: "public", MediaType: "image"}
	require.NoError(t, db.Create(&asset).Error)
	for index, reference := range []string{"/uploads/catalog/brand.png", "https://foreign.example.test/uploads/catalog/brand.png", "/static/brand.svg"} {
		slug := fmt.Sprintf("brand-%d", index)
		response := performProductBrandLogoJSONContractRequest(t, handler, 0, map[string]any{"name": "Brand", "slug": slug, "logo_url": reference})
		require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
		var stored product.ProductBrand
		require.NoError(t, db.Where("slug = ?", slug).First(&stored).Error)
		if index == 0 {
			require.Equal(t, asset.URL, stored.LogoURL)
		} else {
			require.Equal(t, reference, stored.LogoURL)
		}
	}
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
}

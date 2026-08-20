package recommendation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetRecommendationsReturnsOnlyAvailableProducts(t *testing.T) {
	router, db := newRecommendationTestRouter(t)
	available := seedRecommendationHandlerProduct(t, db, "available", "Available Wheel", 5, true)
	seedRecommendationHandlerProduct(t, db, "empty", "Empty Wheel", 0, true)

	payload := map[string]any{
		"surface": "shop_search_drawer",
		"locale":  "en-US",
		"limit":   6,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/recommendations", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Data service.RecommendationResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data.Items, 1)
	require.Equal(t, available.ID, response.Data.Items[0].ProductID)
}

func TestGetRecommendationsRejectsInvalidSurface(t *testing.T) {
	router, _ := newRecommendationTestRouter(t)
	body := bytes.NewBufferString(`{"surface":"","limit":1}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/recommendations", body)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"validation_error"`)
}

func newRecommendationTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&productdomain.ProductSpecificationTemplate{},
		&productdomain.SpecDefinition{},
		&productdomain.ProductInformationTemplate{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{},
		&productdomain.ProductVariant{},
		&productdomain.ProductVariantOptionValue{},
	))

	productService := service.NewProductService(repository.NewProductRepository(db), nil, 0)
	handler := NewHandler(service.NewRecommendationService(productService))
	router := gin.New()
	router.POST("/recommendations", handler.GetRecommendations)
	return router, db
}

func seedRecommendationHandlerProduct(
	t *testing.T,
	db *gorm.DB,
	slug string,
	name string,
	stock int,
	active bool,
) productdomain.Product {
	t.Helper()

	item := productdomain.Product{
		SKU:    slug + "-sku",
		Name:   name,
		Slug:   slug,
		Status: "active",
		Locale: "en",
		Price:  399,
	}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&productdomain.ProductVariant{
		ProductID: item.ID,
		SKU:       slug + "-variant",
		Title:     name,
		Price:     399,
		Stock:     stock,
		IsActive:  active,
		IsDefault: true,
	}).Error)
	return item
}

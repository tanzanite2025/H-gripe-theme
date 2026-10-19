package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductDeclarationIdentifierMissingOrInvalidFieldsReturnActionableBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&product.Product{}, &product.ProductVariant{}))
	handler := NewProductHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router := gin.New()
	router.POST("/products", handler.CreateProduct)
	for _, scenario := range []struct {
		field string
		value any
		label string
	}{
		{"merchant_product_identifier", nil, "MPID"},
		{"manufacturer_product_identifier", " \t ", "NSPID"},
		{"standard_product_identifier", strings.Repeat("0", 71), "SPID"},
	} {
		t.Run(scenario.field, func(t *testing.T) {
			body := map[string]any{
				"name": "Product identifiers", "slug": "product-identifiers", "locale": "en", "status": "active",
				"merchant_product_identifier": "MERCHANT-001", "manufacturer_product_identifier": "FACTORY-002",
				"variants": []map[string]any{{"sku": "SEPARATE-SKU", "price_minor": 1000, "is_default": true}},
			}
			body[scenario.field] = scenario.value
			encoded, err := json.Marshal(body)
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/products", strings.NewReader(string(encoded)))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), scenario.label)
			var count int64
			require.NoError(t, db.Model(&product.Product{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

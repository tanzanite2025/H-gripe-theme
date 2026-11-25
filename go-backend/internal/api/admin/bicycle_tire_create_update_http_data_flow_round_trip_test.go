package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBicycleTireCreateGetAndUpdateHTTPRequestsPreserveProductLevelSpecification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db, tireCategory := createBicycleTireHTTPDataFlowTestRouter(t)

	createBody, err := json.Marshal(map[string]any{
		"merchant_product_identifier":     "TIRE-MERCHANT-HTTP-001",
		"manufacturer_product_identifier": "TIRE-MANUFACTURER-HTTP-001",
		"product_type":                    product.ProductTypeBicycleTire,
		"bicycle_tire_specification":      map[string]any{"schema_version": 1},
		"product_category_id":             tireCategory.ID,
		"name":                            "HTTP Road Tire",
		"slug":                            "http-road-tire",
		"currency":                        "USD",
		"locale":                          "en",
		"status":                          "active",
		"variants": []map[string]any{{
			"sku":         "HTTP-ROAD-TIRE-001",
			"price_minor": 4500,
			"stock":       3,
			"is_default":  true,
			"is_active":   true,
		}},
	})
	require.NoError(t, err)
	createResponse := executeBicycleTireHTTPDataFlowRequest(t, router, http.MethodPost, "/products", createBody)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())
	createdProduct := decodeBicycleTireHTTPDataFlowProduct(t, createResponse)
	require.Equal(t, product.ProductTypeBicycleTire, createdProduct.ProductType)
	require.Equal(t, &product.BicycleTireSpecification{SchemaVersion: 1}, createdProduct.BicycleTireSpecification)
	require.Nil(t, createdProduct.ProductSpecificationTemplateID)
	require.Len(t, createdProduct.Variants, 1)

	productPath := "/products/" + formatBicycleTireHTTPDataFlowProductID(createdProduct.ID)
	getResponse := executeBicycleTireHTTPDataFlowRequest(t, router, http.MethodGet, productPath, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	loadedProduct := decodeBicycleTireHTTPDataFlowProduct(t, getResponse)
	require.Equal(t, product.ProductTypeBicycleTire, loadedProduct.ProductType)
	require.Equal(t, &product.BicycleTireSpecification{SchemaVersion: 1}, loadedProduct.BicycleTireSpecification)
	require.Equal(t, tireCategory.ID, *loadedProduct.ProductCategoryID)

	updateBody, err := json.Marshal(map[string]any{
		"product_type": product.ProductTypeBicycleTire,
		"bicycle_tire_specification": map[string]any{
			"schema_version":   1,
			"etrto":            "40-622",
			"max_pressure_bar": 5,
			"max_pressure_psi": 72.5,
		},
	})
	require.NoError(t, err)
	updateResponse := executeBicycleTireHTTPDataFlowRequest(t, router, http.MethodPut, productPath, updateBody)
	require.Equal(t, http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	updatedProduct := decodeBicycleTireHTTPDataFlowProduct(t, updateResponse)
	require.Equal(t, "40-622", updatedProduct.BicycleTireSpecification.ETRTO)
	require.Equal(t, 5.0, *updatedProduct.BicycleTireSpecification.MaxPressureBar)
	require.Equal(t, 72.5, *updatedProduct.BicycleTireSpecification.MaxPressurePSI)

	getResponse = executeBicycleTireHTTPDataFlowRequest(t, router, http.MethodGet, productPath, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	loadedProduct = decodeBicycleTireHTTPDataFlowProduct(t, getResponse)
	require.Equal(t, "40-622", loadedProduct.BicycleTireSpecification.ETRTO)
	require.Equal(t, 5.0, *loadedProduct.BicycleTireSpecification.MaxPressureBar)
	require.Equal(t, 72.5, *loadedProduct.BicycleTireSpecification.MaxPressurePSI)

	var persistedProduct product.Product
	require.NoError(t, db.First(&persistedProduct, createdProduct.ID).Error)
	require.Equal(t, "40-622", persistedProduct.BicycleTireSpecification.ETRTO)
	require.Equal(t, 5.0, *persistedProduct.BicycleTireSpecification.MaxPressureBar)
}

func createBicycleTireHTTPDataFlowTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, product.ProductCategory) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
		&product.ProductBrand{},
		&product.ProductSpecificationTemplate{},
		&product.SpecDefinition{},
		&product.ProductSpecOptionItem{},
		&product.ProductInformationTemplate{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.ProductShippingTemplateBinding{},
		&product.ProductVariantOptionValue{},
		&product.ProductCustomOptionPolicy{},
		&product.ProductOptionGroupVariantRule{},
		&product.ProductOptionValueVariantRule{},
		&setting.Setting{},
	))
	root := product.ProductCategory{Name: "Wheel Components", Slug: "wheel-components", Depth: 1, IsEnabled: true}
	require.NoError(t, db.Create(&root).Error)
	tireCategory := product.ProductCategory{Name: "Tires", Slug: product.BicycleTireProductCategorySlug, ParentID: &root.ID, Depth: 2, IsEnabled: true}
	require.NoError(t, db.Create(&tireCategory).Error)
	productService := service.NewProductService(repository.NewProductRepository(db), nil, 0)
	productService.ConfigureProductCategoryRepository(repository.NewProductCategoryRepository(db))
	router := gin.New()
	handler := NewProductHandler(productService)
	router.POST("/products", handler.CreateProduct)
	router.GET("/products/:id", handler.GetProduct)
	router.PUT("/products/:id", handler.UpdateProduct)
	return router, db, tireCategory
}

func executeBicycleTireHTTPDataFlowRequest(t *testing.T, router http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func decodeBicycleTireHTTPDataFlowProduct(t *testing.T, response *httptest.ResponseRecorder) *product.Product {
	t.Helper()
	var payload struct {
		Product product.Product `json:"product"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	return &payload.Product
}

func formatBicycleTireHTTPDataFlowProductID(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

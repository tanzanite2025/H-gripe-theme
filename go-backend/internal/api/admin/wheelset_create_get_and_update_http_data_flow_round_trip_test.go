package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestWheelsetCreateGetAndUpdateHTTPRequestsPreserveProductLevelSpecification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db, wheelsetCategory := createWheelsetHTTPDataFlowTestRouter(t)
	createBody, err := json.Marshal(map[string]any{
		"merchant_product_identifier":     "WHEELSET-MERCHANT-HTTP-001",
		"manufacturer_product_identifier": " WHEELSET-MANUFACTURER-HTTP-001 ",
		"standard_product_identifier":     " 00012345678905 ",
		"product_type":                    product.ProductTypeWheelset,
		"product_category_id":             wheelsetCategory.ID,
		"wheelset_specification": map[string]any{
			"schema_version": 1,
			"front_wheel": map[string]any{
				"wheel_size": "700C", "spoke_hole_counts": []int{24, 28}, "lacing_patterns": []string{"2x", "3x"},
			},
			"rear_wheel": map[string]any{
				"wheel_size": "650B", "spoke_hole_counts": []int{28, 32}, "lacing_patterns": []string{"3x"},
			},
		},
		"name":     "HTTP Mixed Diameter Wheelset",
		"slug":     "http-mixed-diameter-wheelset",
		"currency": "USD",
		"locale":   "en",
		"status":   "active",
		"variants": []map[string]any{{
			"sku": "HTTP-MIXED-DIAMETER-WHEELSET", "price_minor": 25000, "stock": 2, "is_default": true, "is_active": true,
		}},
	})
	require.NoError(t, err)
	createResponse := executeWheelsetHTTPDataFlowRequest(t, router, http.MethodPost, "/products", createBody)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())
	createdProduct := decodeWheelsetHTTPDataFlowProduct(t, createResponse)
	require.Equal(t, product.ProductTypeWheelset, createdProduct.ProductType)
	require.Equal(t, "WHEELSET-MANUFACTURER-HTTP-001", createdProduct.ManufacturerProductIdentifier)
	require.Equal(t, "00012345678905", createdProduct.StandardProductIdentifier)
	require.Equal(t, "700C", createdProduct.WheelsetSpecification.FrontWheel.WheelSize)
	require.Equal(t, "650B", createdProduct.WheelsetSpecification.RearWheel.WheelSize)
	require.Nil(t, createdProduct.ProductSpecificationTemplateID)

	productPath := "/products/" + strconv.FormatUint(uint64(createdProduct.ID), 10)
	getResponse := executeWheelsetHTTPDataFlowRequest(t, router, http.MethodGet, productPath, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	loadedProduct := decodeWheelsetHTTPDataFlowProduct(t, getResponse)
	require.Equal(t, createdProduct.WheelsetSpecification, loadedProduct.WheelsetSpecification)
	require.Equal(t, createdProduct.ManufacturerProductIdentifier, loadedProduct.ManufacturerProductIdentifier)
	require.Equal(t, "00012345678905", loadedProduct.StandardProductIdentifier)

	updateBody, err := json.Marshal(map[string]any{
		"product_type":                    product.ProductTypeWheelset,
		"manufacturer_product_identifier": "WHEELSET-MANUFACTURER-HTTP-002",
		"standard_product_identifier":     " 036000291452 ",
		"wheelset_specification": map[string]any{
			"schema_version": 1,
			"front_wheel": map[string]any{
				"wheel_size": "700C", "spoke_hole_counts": []int{24, 28}, "lacing_patterns": []string{"2x", "3x"},
			},
			"rear_wheel": map[string]any{
				"wheel_size": "650B", "spoke_hole_counts": []int{28, 32}, "lacing_patterns": []string{"3x"}, "rim_depth_mm": 45,
			},
			"max_system_weight_kg": 120,
		},
	})
	require.NoError(t, err)
	updateResponse := executeWheelsetHTTPDataFlowRequest(t, router, http.MethodPut, productPath, updateBody)
	require.Equal(t, http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	updatedProduct := decodeWheelsetHTTPDataFlowProduct(t, updateResponse)
	require.Equal(t, "WHEELSET-MANUFACTURER-HTTP-002", updatedProduct.ManufacturerProductIdentifier)
	require.Equal(t, "036000291452", updatedProduct.StandardProductIdentifier)
	require.Equal(t, 45.0, *updatedProduct.WheelsetSpecification.RearWheel.RimDepthMM)
	require.Equal(t, 120.0, *updatedProduct.WheelsetSpecification.MaxSystemWeightKG)
	require.Equal(t, "700C", updatedProduct.WheelsetSpecification.FrontWheel.WheelSize)
	require.Equal(t, "650B", updatedProduct.WheelsetSpecification.RearWheel.WheelSize)

	getResponse = executeWheelsetHTTPDataFlowRequest(t, router, http.MethodGet, productPath, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	loadedProduct = decodeWheelsetHTTPDataFlowProduct(t, getResponse)
	require.Equal(t, 45.0, *loadedProduct.WheelsetSpecification.RearWheel.RimDepthMM)
	require.Equal(t, 120.0, *loadedProduct.WheelsetSpecification.MaxSystemWeightKG)
	var persistedProduct product.Product
	require.NoError(t, db.First(&persistedProduct, createdProduct.ID).Error)
	require.Equal(t, "650B", persistedProduct.WheelsetSpecification.RearWheel.WheelSize)
	require.Equal(t, "WHEELSET-MANUFACTURER-HTTP-002", persistedProduct.ManufacturerProductIdentifier)
	require.Equal(t, "036000291452", persistedProduct.StandardProductIdentifier)

	clearGTINResponse := executeWheelsetHTTPDataFlowRequest(t, router, http.MethodPut, productPath, []byte(`{"standard_product_identifier":""}`))
	require.Equal(t, http.StatusOK, clearGTINResponse.Code, clearGTINResponse.Body.String())
	require.Empty(t, decodeWheelsetHTTPDataFlowProduct(t, clearGTINResponse).StandardProductIdentifier)
	getResponse = executeWheelsetHTTPDataFlowRequest(t, router, http.MethodGet, productPath, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	loadedProduct = decodeWheelsetHTTPDataFlowProduct(t, getResponse)
	require.Empty(t, loadedProduct.StandardProductIdentifier)
	require.Equal(t, "WHEELSET-MANUFACTURER-HTTP-002", loadedProduct.ManufacturerProductIdentifier)
}

func TestWheelsetCreateHTTPRequestRejectsUnknownSpecificationFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, _, wheelsetCategory := createWheelsetHTTPDataFlowTestRouter(t)
	body := []byte(`{"merchant_product_identifier":"WHEELSET-MERCHANT-UNKNOWN-001","manufacturer_product_identifier":"WHEELSET-MANUFACTURER-UNKNOWN-001","product_type":"wheelset","product_category_id":` + strconv.FormatUint(uint64(wheelsetCategory.ID), 10) + `,"wheelset_specification":{"schema_version":1,"front_wheel":{"wheel_size":"700C","spoke_hole_counts":[24],"lacing_patterns":["2x"],"unsupported_engineering_field":"value"},"rear_wheel":{"wheel_size":"700C","spoke_hole_counts":[28],"lacing_patterns":["3x"]}},"name":"Unknown field wheelset","slug":"unknown-field-wheelset","currency":"USD","locale":"en","status":"active"}`)
	response := executeWheelsetHTTPDataFlowRequest(t, router, http.MethodPost, "/products", body)
	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
}

func createWheelsetHTTPDataFlowTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, product.ProductCategory) {
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
	))
	wheelsetCategory := product.ProductCategory{Name: "Wheelsets", Slug: "wheelset", Depth: 1, IsEnabled: true}
	require.NoError(t, db.Create(&wheelsetCategory).Error)
	productService := service.NewProductService(repository.NewProductRepository(db), nil, 0)
	productService.ConfigureProductCategoryRepository(repository.NewProductCategoryRepository(db))
	router := gin.New()
	handler := NewProductHandler(productService)
	router.POST("/products", handler.CreateProduct)
	router.GET("/products/:id", handler.GetProduct)
	router.PUT("/products/:id", handler.UpdateProduct)
	return router, db, wheelsetCategory
}

func executeWheelsetHTTPDataFlowRequest(t *testing.T, router http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func decodeWheelsetHTTPDataFlowProduct(t *testing.T, response *httptest.ResponseRecorder) *product.Product {
	t.Helper()
	var payload struct {
		Product product.Product `json:"product"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	return &payload.Product
}

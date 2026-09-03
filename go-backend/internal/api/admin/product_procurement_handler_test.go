package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	procurementdomain "commerce-platform/internal/domain/procurement"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductProcurementHandlerProductOptionsReturnsMinimalFields(t *testing.T) {
	db := newProductProcurementOptionsHandlerTestDB(t)
	procurementService := service.NewProductProcurementServiceWithProfitability(
		repository.NewProductProcurementRepository(db),
		repository.NewProductProfitCalculationRepository(db),
	)
	procurementService.ConfigureCatalogRepository(repository.NewProductProcurementCatalogRepository(db))
	handler := NewProductProcurementHandler(procurementService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/product-options", handler.ProductOptions)

	request := httptest.NewRequest(http.MethodGet, "/product-options?search=handler", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	var payload struct {
		Options []map[string]interface{} `json:"options"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Len(t, payload.Options, 1)
	require.Equal(t, "SKU-HANDLER-OPTION", payload.Options[0]["sku"])
	require.NotContains(t, response.Body.String(), "stock")
	require.NotContains(t, response.Body.String(), "description")
	require.NotContains(t, response.Body.String(), "product_id")
	require.NotContains(t, response.Body.String(), "variant_id")
	require.NotContains(t, response.Body.String(), "currency")
	require.NotContains(t, response.Body.String(), "price")
}

func TestProductProcurementHandlerCreateUsesCatalogSnapshotAndRejectsManualSKUContract(t *testing.T) {
	db := newProductProcurementOptionsHandlerTestDB(t)
	procurementService := service.NewProductProcurementServiceWithProfitability(
		repository.NewProductProcurementRepository(db),
		repository.NewProductProfitCalculationRepository(db),
	)
	procurementService.ConfigureCatalogRepository(repository.NewProductProcurementCatalogRepository(db))
	handler := NewProductProcurementHandler(procurementService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/records", handler.Create)

	legacyResponse := performProductProcurementJSONRequest(t, router, http.MethodPost, "/records", `{
		"product_code": "SKU-HANDLER-OPTION",
		"product_name": "Manually supplied name",
		"purchase_price": 12
	}`)
	require.Equal(t, http.StatusBadRequest, legacyResponse.Code)

	response := performProductProcurementJSONRequest(t, router, http.MethodPost, "/records", `{
		"sku": "SKU-HANDLER-OPTION",
		"product_name": "Manually supplied name",
		"purchase_price": 12,
		"currency": "USD"
	}`)
	require.Equal(t, http.StatusBadRequest, response.Code)

	response = performProductProcurementJSONRequest(t, router, http.MethodPost, "/records", `{
		"sku": "SKU-HANDLER-OPTION",
		"purchase_price": 12,
		"currency": "USD"
	}`)
	require.Equal(t, http.StatusCreated, response.Code)
	require.Contains(t, response.Body.String(), `"product_name":"Handler option product"`)
	require.NotContains(t, response.Body.String(), "Manually supplied name")
}

func TestProductProcurementHandlerUpdateKeepsStoredProductSnapshot(t *testing.T) {
	db := newProductProcurementOptionsHandlerTestDB(t)
	procurementService := service.NewProductProcurementServiceWithProfitability(
		repository.NewProductProcurementRepository(db),
		repository.NewProductProfitCalculationRepository(db),
	)
	procurementService.ConfigureCatalogRepository(repository.NewProductProcurementCatalogRepository(db))
	handler := NewProductProcurementHandler(procurementService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/records", handler.Create)
	router.PUT("/records/:id", handler.Update)

	createResponse := performProductProcurementJSONRequest(t, router, http.MethodPost, "/records", `{
		"sku": "SKU-HANDLER-OPTION",
		"purchase_price": 12,
		"currency": "USD"
	}`)
	require.Equal(t, http.StatusCreated, createResponse.Code)

	var created struct {
		Record struct {
			ID uint `json:"id"`
		} `json:"record"`
	}
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &created))

	legacyUpdateResponse := performProductProcurementJSONRequest(t, router, http.MethodPut, "/records/1", `{
		"product_code": "SKU-MANUAL-REWRITE",
		"product_name": "Manually supplied name",
		"purchase_price": 18,
		"currency": "USD"
	}`)
	require.Equal(t, http.StatusBadRequest, legacyUpdateResponse.Code)

	updateResponse := performProductProcurementJSONRequest(
		t,
		router,
		http.MethodPut,
		"/records/"+strconv.FormatUint(uint64(created.Record.ID), 10),
		`{"purchase_price":18,"currency":"USD"}`,
	)
	require.Equal(t, http.StatusOK, updateResponse.Code)
	require.Contains(t, updateResponse.Body.String(), `"product_code":"SKU-HANDLER-OPTION"`)
	require.Contains(t, updateResponse.Body.String(), `"product_name":"Handler option product"`)
	require.NotContains(t, updateResponse.Body.String(), "Manually supplied name")
}

func newProductProcurementOptionsHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProcurement{}))
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProfitCalculation{}))
	require.NoError(t, db.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			deleted_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE product_variants (
			id INTEGER PRIMARY KEY,
			product_id INTEGER NOT NULL,
			sku TEXT NOT NULL,
			title TEXT,
			currency TEXT NOT NULL,
			price REAL NOT NULL,
			sale_price REAL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			deleted_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO products (id, name, status, deleted_at) VALUES
			(21, 'Handler option product', 'active', NULL)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO product_variants (id, product_id, sku, title, currency, price, sale_price, is_active, deleted_at) VALUES
			(201, 21, 'SKU-HANDLER-OPTION', 'Default', 'USD', 100, 90, TRUE, NULL)
	`).Error)
	return db
}

func performProductProcurementJSONRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

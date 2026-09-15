package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductProfitabilityHandlerPreviewDoesNotWriteDatabase(t *testing.T) {
	db := newProductProfitabilityHandlerTestDB(t)
	handler := NewProductProfitabilityHandler(
		service.NewProductProfitabilityService(repository.NewProductProfitCalculationRepository(db)),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/preview", handler.Preview)

	response := performProductProfitabilityJSONRequest(t, router, http.MethodPost, "/preview", `{
		"items": [{
			"product_code": "SKU-HANDLER-PREVIEW",
			"product_name": "Handler preview",
			"currency": "USD",
			"list_price": 100,
			"sale_price": 90,
			"unit_cost": 50,
			"unit_cost_known": true
		}]
	}`)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		Items []struct {
			GrossProfit *float64 `json:"gross_profit"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Len(t, payload.Items, 1)
	require.NotNil(t, payload.Items[0].GrossProfit)
	require.Equal(t, 40.0, *payload.Items[0].GrossProfit)

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityHandlerBulkUpsertRejectsInvalidBatchWithoutPartialWrite(t *testing.T) {
	db := newProductProfitabilityHandlerTestDB(t)
	handler := NewProductProfitabilityHandler(
		service.NewProductProfitabilityService(repository.NewProductProfitCalculationRepository(db)),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/bulk-upsert", handler.BulkUpsert)

	response := performProductProfitabilityJSONRequest(t, router, http.MethodPost, "/bulk-upsert", `{
		"items": [
			{
				"product_code": "SKU-HANDLER-VALID",
				"product_name": "Valid item",
				"currency": "USD",
				"list_price": 100,
				"unit_cost": 40,
				"unit_cost_known": true
			},
			{
				"product_code": "SKU-HANDLER-INVALID",
				"product_name": "Invalid item",
				"currency": "USD",
				"cost_currency": "CNY",
				"list_price": 100,
				"unit_cost": 40,
				"unit_cost_known": true
			}
		]
	}`)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "product profitability batch is invalid")
	require.Contains(t, response.Body.String(), "SKU-HANDLER-INVALID")

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityHandlerBulkUpsertAcceptsNestedSupplierCostDetails(t *testing.T) {
	db := newProductProfitabilityHandlerTestDB(t)
	handler := NewProductProfitabilityHandler(
		service.NewProductProfitabilityServiceWithSupplierCostRecords(
			repository.NewProductProfitCalculationRepository(db),
			repository.NewProductSupplierCostRecordRepository(db),
		),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/bulk-upsert", handler.BulkUpsert)

	response := performProductProfitabilityJSONRequest(t, router, http.MethodPost, "/bulk-upsert", `{
		"items": [{
			"product_code": "SKU-HANDLER-SUPPLIER-COST",
			"product_name": "Handler supplier cost item",
			"currency": "USD",
			"list_price": 100,
			"unit_cost": 40,
			"unit_cost_known": true,
			"supplier_cost_details": {
				"supplier_name": "Nested Supplier",
				"supplier_contact_name": "Ming",
				"supplier_phone": "+86-456",
				"lead_time_days": 12,
				"minimum_order_quantity": 5
			}
		}]
	}`)
	require.Equal(t, http.StatusOK, response.Code)

	supplierCostRecord, err := repository.NewProductSupplierCostRecordRepository(db).
		FindByProductCode("SKU-HANDLER-SUPPLIER-COST")
	require.NoError(t, err)
	require.Equal(t, "Nested Supplier", supplierCostRecord.SupplierName)
	require.Equal(t, "Ming", supplierCostRecord.SupplierContactName)
	require.Equal(t, 12, supplierCostRecord.LeadTimeDays)
	require.Equal(t, 5, supplierCostRecord.MinimumOrderQuantity)
}

func TestProductProfitabilityHandlerAcceptsHistoricalLegacyCostAndSupplierDetailsEnvelopeAsSupplierCostCompatibility(t *testing.T) {
	db := newProductProfitabilityHandlerTestDB(t)
	handler := NewProductProfitabilityHandler(
		service.NewProductProfitabilityServiceWithSupplierCostRecords(
			repository.NewProductProfitCalculationRepository(db),
			repository.NewProductSupplierCostRecordRepository(db),
		),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/bulk-upsert", handler.BulkUpsert)

	response := performProductProfitabilityJSONRequest(t, router, http.MethodPost, "/bulk-upsert", `{
		"items": [{
			"product_code": "SKU-HANDLER-LEGACY-SUPPLIER-COST",
			"product_name": "Handler legacy supplier cost item",
			"currency": "USD",
			"list_price": 100,
			"purchase_price": 40,
			"purchase_price_known": true,
			"procurement": {
				"supplier_name": "Legacy Supplier",
				"supplier_contact_name": "Ming",
				"lead_time_days": 12,
				"minimum_order_quantity": 5
			}
		}]
	}`)
	require.Equal(t, http.StatusOK, response.Code)

	supplierCostRecord, err := repository.NewProductSupplierCostRecordRepository(db).
		FindByProductCode("SKU-HANDLER-LEGACY-SUPPLIER-COST")
	require.NoError(t, err)
	require.Equal(t, "Legacy Supplier", supplierCostRecord.SupplierName)
}

func newProductProfitabilityHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "product-profitability-handler.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductProfitCalculation{}))
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductSupplierCostRecord{}))
	return db
}

func performProductProfitabilityJSONRequest(
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

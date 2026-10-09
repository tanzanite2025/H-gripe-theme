package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	taxratedomain "commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaxRateRuleHandlerCreatesUpdatesListsAndDeletesCheckoutRules(t *testing.T) {
	router, database := newTaxRateRuleHandlerTestRouter(t)

	createResponse := performTaxRateRuleHandlerRequest(t, router, http.MethodPost, "/rules", `{
		"name":"Germany explicit zero rate",
		"country":"de",
		"state":"",
		"postal_code":"",
		"rate_decimal":"0",
		"priority":0,
		"enabled":true
	}`)
	require.Equal(t, http.StatusCreated, createResponse.Code)
	var createdEnvelope struct {
		Data taxratedomain.TaxRate `json:"data"`
	}
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &createdEnvelope))
	require.NotZero(t, createdEnvelope.Data.ID)
	assert.Equal(t, "DE", createdEnvelope.Data.Country)
	assert.Equal(t, "0", createdEnvelope.Data.RateDecimal)

	listResponse := performTaxRateRuleHandlerRequest(t, router, http.MethodGet, "/rules", "")
	require.Equal(t, http.StatusOK, listResponse.Code)
	var listEnvelope struct {
		Data struct {
			Rules []taxratedomain.TaxRate `json:"rules"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(listResponse.Body.Bytes(), &listEnvelope))
	require.Len(t, listEnvelope.Data.Rules, 1)

	updateResponse := performTaxRateRuleHandlerRequest(t, router, http.MethodPut, "/rules/"+taxRateRuleIDString(createdEnvelope.Data.ID), `{
		"name":"Germany VAT",
		"country":"DE",
		"state":"",
		"postal_code":"",
		"rate_decimal":"19",
		"priority":2,
		"enabled":false
	}`)
	require.Equal(t, http.StatusOK, updateResponse.Code)
	var updatedEnvelope struct {
		Data taxratedomain.TaxRate `json:"data"`
	}
	require.NoError(t, json.Unmarshal(updateResponse.Body.Bytes(), &updatedEnvelope))
	assert.Equal(t, "19", updatedEnvelope.Data.RateDecimal)
	assert.False(t, updatedEnvelope.Data.Enabled)

	var persistedRule taxratedomain.TaxRate
	require.NoError(t, database.First(&persistedRule, createdEnvelope.Data.ID).Error)
	assert.False(t, persistedRule.Enabled)

	deleteResponse := performTaxRateRuleHandlerRequest(t, router, http.MethodDelete, "/rules/"+taxRateRuleIDString(createdEnvelope.Data.ID), "")
	require.Equal(t, http.StatusOK, deleteResponse.Code)
	listResponse = performTaxRateRuleHandlerRequest(t, router, http.MethodGet, "/rules", "")
	require.Equal(t, http.StatusOK, listResponse.Code)
	require.NoError(t, json.Unmarshal(listResponse.Body.Bytes(), &listEnvelope))
	assert.Empty(t, listEnvelope.Data.Rules)
}

func TestTaxRateRuleHandlerRejectsInvalidRateRule(t *testing.T) {
	router, _ := newTaxRateRuleHandlerTestRouter(t)

	response := performTaxRateRuleHandlerRequest(t, router, http.MethodPost, "/rules", `{
		"name":"Invalid rule",
		"country":"US",
		"rate_decimal":"7.1234567890123456",
		"enabled":true
	}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
}

func newTaxRateRuleHandlerTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDatabase, err := database.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	require.NoError(t, database.AutoMigrate(&taxratedomain.TaxRate{}))

	handler := NewTaxRateRuleHandler(service.NewTaxRateService(repository.NewTaxRateRepository(database)))
	router := gin.New()
	router.GET("/rules", handler.ListRules)
	router.POST("/rules", handler.CreateRule)
	router.PUT("/rules/:id", handler.UpdateRule)
	router.DELETE("/rules/:id", handler.DeleteRule)
	return router, database
}

func performTaxRateRuleHandlerRequest(
	t *testing.T,
	router http.Handler,
	method, path, body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func taxRateRuleIDString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

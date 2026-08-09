package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExchangeRateHandlerConvertDisplayPricesUsesCachedRates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))

	repo := repository.NewExchangeRateRepository(db)
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{
		{
			BaseCurrency:  "CNY",
			QuoteCurrency: "USD",
			Rate:          0.14,
			Source:        "test_cache",
			FetchedAt:     time.Now().UTC(),
		},
	}))

	handler := NewExchangeRateHandler(service.NewExchangeRateService(repo, nil))
	router := gin.New()
	router.POST("/convert", handler.ConvertDisplayPrices)

	body := bytes.NewBufferString(`{"amount":699,"base_currency":"CNY","quote_currencies":["CNY","usd","USD"]}`)
	request := httptest.NewRequest(http.MethodPost, "/convert", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Amount          float64  `json:"amount"`
			BaseCurrency    string   `json:"base_currency"`
			QuoteCurrencies []string `json:"quote_currencies"`
			Prices          []struct {
				Amount    float64 `json:"amount"`
				Currency  string  `json:"currency"`
				Rate      float64 `json:"rate"`
				Source    string  `json:"source"`
				Converted bool    `json:"converted"`
			} `json:"prices"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Equal(t, 699.0, payload.Data.Amount)
	require.Equal(t, "CNY", payload.Data.BaseCurrency)
	require.Equal(t, []string{"USD"}, payload.Data.QuoteCurrencies)
	require.Len(t, payload.Data.Prices, 1)
	require.Equal(t, 97.86, payload.Data.Prices[0].Amount)
	require.Equal(t, "USD", payload.Data.Prices[0].Currency)
	require.Equal(t, 0.14, payload.Data.Prices[0].Rate)
	require.Equal(t, "direct_rate", payload.Data.Prices[0].Source)
	require.True(t, payload.Data.Prices[0].Converted)
}

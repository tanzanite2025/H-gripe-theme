package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRefreshDisplayPriceSnapshotsPreservesMismatchedSourceAmounts(t *testing.T) {
	db := newDisplayPriceRefreshTestDB(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, nil, 0)

	oldProductSnapshot := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{Amount: 100, Currency: "USD", QuoteCurrency: "USD", Rate: 0.1, Source: "old_rate", Converted: true},
	}, "CNY")
	oldVariantSnapshot := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{Amount: 101, Currency: "USD", QuoteCurrency: "USD", Rate: 0.1, Source: "old_rate", Converted: true},
	}, "CNY")

	productOne := product.Product{
		SKU:              "CNY-001",
		Name:             "CNY product",
		Slug:             "cny-product",
		Locale:           "en",
		Status:           "active",
		Currency:         "CNY",
		Price:            699,
		DisplayPriceData: oldProductSnapshot,
	}
	productTwo := product.Product{
		SKU:              "USD-001",
		Name:             "Mismatched product",
		Slug:             "mismatched-product",
		Locale:           "en",
		Status:           "active",
		Currency:         "USD",
		Price:            10,
		DisplayPriceData: oldProductSnapshot,
	}
	productThree := product.Product{
		SKU:              "CNY-002",
		Name:             "Mismatched variant product",
		Slug:             "mismatched-variant-product",
		Locale:           "en",
		Status:           "active",
		Currency:         "CNY",
		Price:            200,
		DisplayPriceData: oldProductSnapshot,
	}
	require.NoError(t, db.Create(&productOne).Error)
	require.NoError(t, db.Create(&productTwo).Error)
	require.NoError(t, db.Create(&productThree).Error)

	variantOne := product.ProductVariant{
		ProductID:        productOne.ID,
		SKU:              "CNY-001-V",
		OptionValues:     "{}",
		Currency:         "CNY",
		Price:            699,
		DisplayPriceData: oldVariantSnapshot,
		Stock:            1,
		IsDefault:        true,
		IsActive:         true,
	}
	variantTwo := product.ProductVariant{
		ProductID:        productTwo.ID,
		SKU:              "USD-001-V",
		OptionValues:     "{}",
		Currency:         "CNY",
		Price:            100,
		DisplayPriceData: oldVariantSnapshot,
		Stock:            1,
		IsDefault:        true,
		IsActive:         true,
	}
	variantThree := product.ProductVariant{
		ProductID:        productThree.ID,
		SKU:              "CNY-002-V",
		OptionValues:     "{}",
		Currency:         "USD",
		Price:            200,
		DisplayPriceData: oldVariantSnapshot,
		Stock:            1,
		IsDefault:        true,
		IsActive:         true,
	}
	require.NoError(t, db.Create(&variantOne).Error)
	require.NoError(t, db.Create(&variantTwo).Error)
	require.NoError(t, db.Create(&variantThree).Error)

	result, err := productService.RefreshDisplayPriceSnapshots(
		"CNY",
		[]string{"USD"},
		[]currency.ExchangeRate{{BaseCurrency: "CNY", QuoteCurrency: "USD", Rate: 0.14}},
	)

	require.NoError(t, err)
	require.Equal(t, 3, result.ProductsScanned)
	require.Equal(t, 2, result.ProductsUpdated)
	require.Equal(t, 3, result.VariantsScanned)
	require.Equal(t, 2, result.VariantsUpdated)
	require.Equal(t, 2, result.CurrencyMismatchCount)

	var storedProductOne, storedProductTwo, storedProductThree product.Product
	require.NoError(t, db.First(&storedProductOne, productOne.ID).Error)
	require.NoError(t, db.First(&storedProductTwo, productTwo.ID).Error)
	require.NoError(t, db.First(&storedProductThree, productThree.ID).Error)
	require.Equal(t, 699.0, storedProductOne.Price)
	require.Equal(t, "CNY", storedProductOne.Currency)
	require.Equal(t, 10.0, storedProductTwo.Price)
	require.Equal(t, "USD", storedProductTwo.Currency)
	require.Equal(t, 200.0, storedProductThree.Price)
	require.Equal(t, "CNY", storedProductThree.Currency)
	require.Equal(t, 97.86, displaySnapshotAmount(storedProductOne.DisplayPriceData, "USD"))
	require.Equal(t, 100.0, displaySnapshotAmount(storedProductTwo.DisplayPriceData, "USD"))
	require.Equal(t, 28.0, displaySnapshotAmount(storedProductThree.DisplayPriceData, "USD"))

	var storedVariantOne, storedVariantTwo, storedVariantThree product.ProductVariant
	require.NoError(t, db.First(&storedVariantOne, variantOne.ID).Error)
	require.NoError(t, db.First(&storedVariantTwo, variantTwo.ID).Error)
	require.NoError(t, db.First(&storedVariantThree, variantThree.ID).Error)
	require.Equal(t, 699.0, storedVariantOne.Price)
	require.Equal(t, "CNY", storedVariantOne.Currency)
	require.Equal(t, 97.86, displaySnapshotAmount(storedVariantOne.DisplayPriceData, "USD"))
	require.Equal(t, 100.0, storedVariantTwo.Price)
	require.Equal(t, "CNY", storedVariantTwo.Currency)
	require.Equal(t, 14.0, displaySnapshotAmount(storedVariantTwo.DisplayPriceData, "USD"))
	require.Equal(t, 200.0, storedVariantThree.Price)
	require.Equal(t, "USD", storedVariantThree.Currency)
	require.Equal(t, 101.0, displaySnapshotAmount(storedVariantThree.DisplayPriceData, "USD"))
}

func TestExchangeRateSyncRefreshesProductDisplayPriceSnapshots(t *testing.T) {
	db := newDisplayPriceRefreshTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&setting.Setting{},
		&currency.ExchangeRate{},
		&currency.ExchangeRateSyncLease{},
	))

	settingRepo := repository.NewSettingRepository(db)
	exchangeRateRepo := repository.NewExchangeRateRepository(db)
	policyService := NewCurrencyPolicyService(settingRepo)
	_, err := policyService.UpdatePolicy(currency.Policy{PrimaryCurrency: "CNY"})
	require.NoError(t, err)

	oldSnapshot := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{Amount: 100, Currency: "USD", QuoteCurrency: "USD", Rate: 0.1, Source: "old_rate", Converted: true},
	}, "CNY")
	storedProduct := product.Product{
		SKU:              "SYNC-001",
		Name:             "Sync product",
		Slug:             "sync-product",
		Locale:           "en",
		Status:           "active",
		Currency:         "CNY",
		Price:            100,
		DisplayPriceData: oldSnapshot,
	}
	require.NoError(t, db.Create(&storedProduct).Error)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v6/test-key/latest/CNY", r.URL.Path)
		_, _ = w.Write([]byte(`{
			"result": "success",
			"base_code": "CNY",
			"conversion_rates": {"USD": 0.14}
		}`))
	}))
	t.Cleanup(apiServer.Close)

	require.NoError(t, settingRepo.BatchSet([]setting.Setting{
		{Key: "exchange_rate_enabled", Value: "true", Type: "boolean", Locale: "en", Group: "api"},
		{Key: "exchange_rate_endpoint", Value: apiServer.URL + "/v6/{apiKey}/latest/{base}", Type: "string", Locale: "en", Group: "api"},
		{Key: "exchange_rate_api_key", Value: "test-key", Type: "string", Locale: "en", Group: "api"},
	}))

	productService := NewProductService(repository.NewProductRepository(db), nil, 0)
	exchangeRateService := NewExchangeRateService(exchangeRateRepo, settingRepo)
	exchangeRateService.ConfigureCurrencyPolicy(policyService)
	exchangeRateService.ConfigureProductService(productService)

	result, err := exchangeRateService.Sync()

	require.NoError(t, err)
	require.NotNil(t, result.DisplayPriceRefresh)
	require.Equal(t, 1, result.DisplayPriceRefresh.ProductsUpdated)
	require.Equal(t, 100.0, storedProduct.Price)

	var refreshed product.Product
	require.NoError(t, db.First(&refreshed, storedProduct.ID).Error)
	require.Equal(t, 100.0, refreshed.Price)
	require.Equal(t, "CNY", refreshed.Currency)
	require.Equal(t, 14.0, displaySnapshotAmount(refreshed.DisplayPriceData, "USD"))

	rate, err := exchangeRateRepo.Find("CNY", "USD")
	require.NoError(t, err)
	require.Equal(t, 0.14, rate.Rate)
}

func newDisplayPriceRefreshTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&product.Product{}, &product.ProductVariant{}))
	return db
}

func displaySnapshotAmount(raw []byte, quoteCurrency string) float64 {
	var snapshots []currency.DisplayPriceSnapshot
	if err := json.Unmarshal(raw, &snapshots); err != nil {
		return 0
	}
	for _, snapshot := range snapshots {
		if currency.NormalizeCode(snapshot.Currency) == currency.NormalizeCode(quoteCurrency) {
			return snapshot.Amount
		}
	}
	return 0
}

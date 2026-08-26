package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCurrencyPolicyPersistsBackendEntryCurrencyOnly(t *testing.T) {
	service := newTestCurrencyPolicyService(t)

	policy, err := service.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "cny",
		DisplayCurrencies: []string{"CNY", "usd", "eur", "USD"},
	})
	require.NoError(t, err)
	require.Equal(t, "CNY", policy.PrimaryCurrency)
	require.Empty(t, policy.DisplayCurrencies)

	loaded, err := service.GetPolicy()
	require.NoError(t, err)
	require.Equal(t, "CNY", loaded.PrimaryCurrency)
	require.Empty(t, loaded.DisplayCurrencies)
}

func TestCurrencyPolicyDefaultsDisplayCurrenciesToEmptyWhenUnset(t *testing.T) {
	service := newTestCurrencyPolicyService(t)

	policy, err := service.GetPolicy()

	require.NoError(t, err)
	require.Equal(t, currency.DefaultPrimaryCurrency, policy.PrimaryCurrency)
	require.Empty(t, policy.DisplayCurrencies)
}

func TestCurrencyPolicyDeletesLegacyGlobalDisplayCurrenciesWhenSaved(t *testing.T) {
	service := newTestCurrencyPolicyService(t)
	require.NoError(t, service.settings.BatchSet([]setting.Setting{
		{
			Key:      "currency_display_currencies",
			Value:    "USD,EUR",
			Type:     "string",
			Locale:   "en",
			Group:    "currency",
			IsPublic: true,
		},
	}))

	_, err := service.UpdatePolicy(currency.Policy{PrimaryCurrency: "CNY"})

	require.NoError(t, err)
	_, err = service.settings.Get("currency_display_currencies", "en")
	require.True(t, repository.IsRecordNotFound(err))
}

func TestCurrencyPolicyRejectsUnsupportedPrimaryCurrency(t *testing.T) {
	service := newTestCurrencyPolicyService(t)

	_, err := service.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "BTC",
		DisplayCurrencies: []string{"USD"},
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidCurrencyPolicy))
	require.Contains(t, err.Error(), "unsupported primary currency")
}

func newTestCurrencyPolicyService(t *testing.T) *CurrencyPolicyService {
	t.Helper()
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
	require.NoError(t, db.AutoMigrate(&setting.Setting{}))
	return NewCurrencyPolicyService(repository.NewSettingRepository(db))
}

package service

import (
	"errors"
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
	taxratedomain "commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaxRateServiceListsEnabledRatesAndHidesDisabledRates(t *testing.T) {
	database := newTaxRateServiceTestDatabase(t)
	service := NewTaxRateService(repository.NewTaxRateRepository(database))

	enabledRate := taxratedomain.TaxRate{
		Name:        "Enabled",
		Country:     "US",
		State:       "CA",
		RateDecimal: "7.5",
		Enabled:     true,
	}
	disabledRate := taxratedomain.TaxRate{
		Name:        "Disabled",
		Country:     "US",
		State:       "NY",
		RateDecimal: "8.5",
	}
	require.NoError(t, database.Create(&enabledRate).Error)
	require.NoError(t, database.Create(&disabledRate).Error)
	require.NoError(t, database.Model(&taxratedomain.TaxRate{}).
		Where("id = ?", disabledRate.ID).
		Update("enabled", false).Error)

	rates, err := service.ListPublicTaxRates()
	require.NoError(t, err)
	require.Len(t, rates, 1)
	assert.Equal(t, enabledRate.ID, rates[0].ID)

	_, err = service.GetPublicTaxRate(disabledRate.ID)
	assert.ErrorIs(t, err, taxratedomain.ErrTaxRateNotFound)
}

func TestTaxRateServicePrefersPostalRateAndFallsBackToStateRate(t *testing.T) {
	database := newTaxRateServiceTestDatabase(t)
	service := NewTaxRateService(repository.NewTaxRateRepository(database))

	defaultRate := taxratedomain.TaxRate{
		Name:        "California default",
		Country:     "US",
		State:       "CA",
		RateDecimal: "7.25",
		Enabled:     true,
	}
	postalRate := taxratedomain.TaxRate{
		Name:        "Beverly Hills",
		Country:     "US",
		State:       "CA",
		PostalCode:  "90210",
		RateDecimal: "9.5",
		Enabled:     true,
	}
	require.NoError(t, database.Create(&defaultRate).Error)
	require.NoError(t, database.Create(&postalRate).Error)

	amount := domainmoney.MustNew(10000, "USD")
	rate, taxMoney, err := service.CalculateTaxMoney(amount, "us", "ca", "90210")
	require.NoError(t, err)
	assert.Equal(t, "9.5", rate)
	assert.Equal(t, int64(950), taxMoney.AmountMinor())

	rate, taxMoney, err = service.CalculateTaxMoney(amount, "US", "CA", "10001")
	require.NoError(t, err)
	assert.Equal(t, "7.25", rate)
	assert.Equal(t, int64(725), taxMoney.AmountMinor())
}

func TestTaxRateServiceRejectsMissingRateInsteadOfReturningZeroTax(t *testing.T) {
	database := newTaxRateServiceTestDatabase(t)
	service := NewTaxRateService(repository.NewTaxRateRepository(database))

	rate, taxMoney, err := service.CalculateTaxMoney(domainmoney.MustNew(10000, "USD"), "US", "CA", "90210")

	assert.ErrorIs(t, err, ErrTaxRateUnavailable)
	assert.Empty(t, rate)
	assert.Zero(t, taxMoney.AmountMinor())
}

func TestTaxRateServiceAllowsExplicitZeroRate(t *testing.T) {
	database := newTaxRateServiceTestDatabase(t)
	service := NewTaxRateService(repository.NewTaxRateRepository(database))
	require.NoError(t, database.Create(&taxratedomain.TaxRate{
		Name:        "Explicit zero tax jurisdiction",
		Country:     "US",
		State:       "CA",
		RateDecimal: "0",
		Enabled:     true,
	}).Error)

	rate, taxMoney, err := service.CalculateTaxMoney(domainmoney.MustNew(10000, "USD"), "US", "CA")

	require.NoError(t, err)
	assert.Equal(t, "0", rate)
	assert.Zero(t, taxMoney.AmountMinor())
}

func TestTaxRateServiceCreatesUpdatesAndSoftDeletesCheckoutRules(t *testing.T) {
	database := newTaxRateServiceTestDatabase(t)
	service := NewTaxRateService(repository.NewTaxRateRepository(database))

	zeroRate, err := service.CreateTaxRateRule(TaxRateRuleInput{
		Name:        "Explicit zero rate",
		Country:     "us",
		RateDecimal: " 0 ",
		Enabled:     true,
	})
	require.NoError(t, err)
	assert.Equal(t, "US", zeroRate.Country)
	assert.Equal(t, "0", zeroRate.RateDecimal)
	assert.True(t, zeroRate.Enabled)

	disabledRate, err := service.CreateTaxRateRule(TaxRateRuleInput{
		Name:        "Disabled rate",
		Country:     "CA",
		RateDecimal: "7.25",
		Enabled:     false,
	})
	require.NoError(t, err)
	assert.False(t, disabledRate.Enabled)

	updatedRate, err := service.UpdateTaxRateRule(zeroRate.ID, TaxRateRuleInput{
		Name:        "Updated zero rate",
		Country:     "US",
		State:       "ca",
		RateDecimal: "5.5",
		Priority:    10,
		Enabled:     true,
	})
	require.NoError(t, err)
	assert.Equal(t, "CA", updatedRate.State)
	assert.Equal(t, "5.5", updatedRate.RateDecimal)
	assert.Equal(t, 10, updatedRate.Priority)

	taxRate, taxMoney, err := service.CalculateTaxMoney(domainmoney.MustNew(10000, "USD"), "US", "CA")
	require.NoError(t, err)
	assert.Equal(t, "5.5", taxRate)
	assert.Equal(t, int64(550), taxMoney.AmountMinor())

	require.NoError(t, service.DeleteTaxRateRule(zeroRate.ID))
	_, err = service.GetTaxRate(zeroRate.ID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	rules, err := service.ListTaxRates()
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, disabledRate.ID, rules[0].ID)
}

func TestTaxRateServiceRejectsInvalidRuleCountryAndRate(t *testing.T) {
	service := NewTaxRateService(repository.NewTaxRateRepository(newTaxRateServiceTestDatabase(t)))

	_, err := service.CreateTaxRateRule(TaxRateRuleInput{
		Name:        "Invalid country",
		Country:     "USA",
		RateDecimal: "5",
		Enabled:     true,
	})
	assert.ErrorIs(t, err, ErrTaxRateRuleInvalid)

	_, err = service.CreateTaxRateRule(TaxRateRuleInput{
		Name:        "Invalid rate",
		Country:     "US",
		RateDecimal: "7.1234567890123456",
		Enabled:     true,
	})
	assert.ErrorIs(t, err, ErrTaxRateRuleInvalid)
}

func newTaxRateServiceTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDatabase, err := database.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	require.NoError(t, database.AutoMigrate(&taxratedomain.TaxRate{}))
	return database
}

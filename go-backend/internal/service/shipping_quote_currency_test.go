package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestQuoteResolvedItemsConvertsTemplateCurrencyToQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))

	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Rate:          0.9,
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	exchangeRates := serviceExchangeRateForTest(db)
	shippingService.ConfigureExchangeRateService(exchangeRates)

	template := shippingdomain.ShippingTemplate{
		Name:       "USD source template",
		Type:       "weight",
		Currency:   "USD",
		DefaultFee: 10,
		Enabled:    true,
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Amount:   50,
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:          1,
			ShippingTemplateID: &template.ID,
			Quantity:           1,
			UnitPrice:          50,
			WeightGrams:        1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Equal(t, "EUR", quote.Currency)
	assert.InDelta(t, 9.0, quote.ShippingFee, 0.001)
	require.Len(t, quote.Items, 1)
	assert.InDelta(t, 9.0, quote.Items[0].ShippingFee, 0.001)
}

func TestQuoteResolvedItemsConvertsPriceRuleThresholdFromQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Rate:          0.9,
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	shippingService.ConfigureExchangeRateService(serviceExchangeRateForTest(db))

	template := shippingdomain.ShippingTemplate{
		Name:     "USD price template",
		Type:     "price",
		Currency: "USD",
		Enabled:  true,
		Rules: []shippingdomain.ShippingRule{{
			Region: "DE", Currency: "USD", MinValue: 100, MaxValue: 200, Fee: 20,
		}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Amount:   90,
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:          1,
			ShippingTemplateID: &template.ID,
			Quantity:           1,
			UnitPrice:          90,
			WeightGrams:        1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.InDelta(t, 18.0, quote.ShippingFee, 0.001)
}

func TestQuoteResolvedItemsConvertsFreeShippingThresholdFromQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Rate:          0.9,
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	shippingService.ConfigureExchangeRateService(serviceExchangeRateForTest(db))

	template := shippingdomain.ShippingTemplate{
		Name:          "USD free-shipping threshold",
		Type:          "weight",
		Currency:      "USD",
		DefaultFee:    10,
		FreeShipping:  true,
		FreeThreshold: 100,
		Enabled:       true,
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Amount:   90,
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:          1,
			ShippingTemplateID: &template.ID,
			Quantity:           1,
			UnitPrice:          90,
			WeightGrams:        1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Zero(t, quote.ShippingFee)
	assert.True(t, quote.FreeShipping)
}

func TestQuoteResolvedItemsConvertsCarrierSurchargeToQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Rate:          0.9,
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	shippingService.ConfigureExchangeRateService(serviceExchangeRateForTest(db))

	template := shippingdomain.ShippingTemplate{
		Name:       "USD carrier template",
		Type:       "weight",
		Currency:   "USD",
		DefaultFee: 10,
		Enabled:    true,
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	carrier := seedQuoteCarrier(t, db, "Test Carrier", "test-carrier")
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:     "test-service",
		ServiceName:     "Test Service",
		Countries:       `["DE"]`,
		Currency:        "USD",
		RemoteSurcharge: 2,
	})

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Amount:   50,
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:          1,
			ShippingTemplateID: &template.ID,
			Quantity:           1,
			UnitPrice:          50,
			WeightGrams:        1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, "EUR", leg.Currency)
	assert.InDelta(t, 10.8, leg.ShippingFee, 0.001)
}

func serviceExchangeRateForTest(db *gorm.DB) *ExchangeRateService {
	return NewExchangeRateService(repository.NewExchangeRateRepository(db), nil)
}

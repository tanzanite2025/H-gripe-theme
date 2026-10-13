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

func TestQuoteCartUsesConvertedCatalogPriceForFreeShippingThreshold(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{BaseCurrency: "USD", QuoteCurrency: "EUR", RateDecimal: "0.9", Source: "test", FetchedAt: time.Now().UTC()}).Error)
	shippingService.ConfigureExchangeRateService(serviceExchangeRateForTest(db))
	template := shippingdomain.ShippingTemplate{Name: "EUR threshold", Type: "quantity", Currency: "EUR", Enabled: true, FreeShipping: true, FreeThresholdMinor: 4700, DefaultFeeMinor: 1000}
	require.NoError(t, shippingService.CreateTemplate(&template))
	item, variant := seedQuoteProduct(t, db, 50, 800, template.ID)
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "EUR", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
	require.NoError(t, err)
	assert.Equal(t, int64(4500), quote.Items[0].AmountMinor)
	assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
	assert.False(t, quote.FreeShipping)
}

func TestQuoteCartRejectsMissingExchangeRateForCatalogPrice(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{Name: "EUR billing", Type: "quantity", Currency: "EUR", Enabled: true, DefaultFeeMinor: 1000}
	require.NoError(t, shippingService.CreateTemplate(&template))
	item, variant := seedQuoteProduct(t, db, 50, 800, template.ID)
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "EUR", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
	require.ErrorIs(t, err, ErrShippingRateUnavailable)
	assert.Nil(t, quote)
}

func TestQuoteResolvedItemsConvertsTemplateCurrencyToQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))

	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		RateDecimal:   "0.9",
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	exchangeRates := serviceExchangeRateForTest(db)
	shippingService.ConfigureExchangeRateService(exchangeRates)

	template := shippingdomain.ShippingTemplate{
		Name:            "USD source template",
		Type:            "weight",
		Currency:        "USD",
		DefaultFeeMinor: 1000,
		Enabled:         true,
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:           1,
			ShippingTemplateIDs: []uint{template.ID},
			Quantity:            1,
			UnitPriceMinor:      5000,
			WeightGrams:         1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Equal(t, "EUR", quote.Currency)
	assert.Equal(t, "9.00", quote.ShippingFeeDecimal)
	require.Len(t, quote.Items, 1)
	assert.Equal(t, "9.00", quote.Items[0].ShippingFeeDecimal)
}

func TestQuoteResolvedItemsConvertsPriceRuleThresholdFromQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		RateDecimal:   "0.9",
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
			Region: "DE", Currency: "USD", MinValueMinor: 10000, MaxValueMinor: 20000, FeeMinor: 2000,
		}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:           1,
			ShippingTemplateIDs: []uint{template.ID},
			Quantity:            1,
			UnitPriceMinor:      9000,
			WeightGrams:         1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Equal(t, "18.00", quote.ShippingFeeDecimal)
}

func TestQuoteResolvedItemsConvertsFreeShippingThresholdFromQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		RateDecimal:   "0.9",
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	shippingService.ConfigureExchangeRateService(serviceExchangeRateForTest(db))

	template := shippingdomain.ShippingTemplate{
		Name:               "USD free-shipping threshold",
		Type:               "weight",
		Currency:           "USD",
		DefaultFeeMinor:    1000,
		FreeShipping:       true,
		FreeThresholdMinor: 10000,
		Enabled:            true,
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:           1,
			ShippingTemplateIDs: []uint{template.ID},
			Quantity:            1,
			UnitPriceMinor:      9000,
			WeightGrams:         1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Equal(t, "0.00", quote.ShippingFeeDecimal)
	assert.True(t, quote.FreeShipping)
}

func TestQuoteResolvedItemsConvertsCarrierSurchargeToQuoteCurrency(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRate{}))
	require.NoError(t, db.Create(&currency.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		RateDecimal:   "0.9",
		Source:        "test",
		FetchedAt:     time.Now().UTC(),
	}).Error)
	shippingService.ConfigureExchangeRateService(serviceExchangeRateForTest(db))

	template := shippingdomain.ShippingTemplate{
		Name:            "USD carrier template",
		Type:            "weight",
		Currency:        "USD",
		DefaultFeeMinor: 1000,
		Enabled:         true,
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	carrier := seedQuoteCarrier(t, db, "Test Carrier", "test-carrier")
	service := seedQuoteCustomerBillingTemplateAndCarrierRoute(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:          "test-service",
		ServiceName:          "Test Service",
		Countries:            `["DE"]`,
		Currency:             "USD",
		RemoteSurchargeMinor: 200,
	})

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:  "DE",
		Currency: "EUR",
		Items: []ShippingQuoteItemInput{{
			ProductID:           1,
			ShippingTemplateIDs: []uint{template.ID},
			Quantity:            1,
			UnitPriceMinor:      5000,
			WeightGrams:         1000,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, "EUR", leg.Currency)
	assert.Equal(t, "10.80", leg.ShippingFeeDecimal)
}

func serviceExchangeRateForTest(db *gorm.DB) *ExchangeRateService {
	return NewExchangeRateService(repository.NewExchangeRateRepository(db), nil)
}

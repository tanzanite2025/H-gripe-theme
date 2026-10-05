package service

import (
	"testing"

	shippingdomain "commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestQuoteCartUsesCurrentFpxCollectionCountryScope(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createGlobalWeightQuoteTemplateForPublishedCollectionTest(t, db)
	product, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "4PX", "4PX")
	channel := shippingdomain.FpxChannel{
		ServiceCode: "EU-EXPRESS", DisplayName: "4PX Europe Express", Countries: `["US"]`, Enabled: true,
	}
	require.NoError(t, db.Create(&channel).Error)
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		FpxChannelID: &channel.ID,
		ServiceCode:  "EU-EXPRESS", ServiceName: "Old service name", Countries: `["CA"]`,
		BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})
	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: product.ID, VariantID: &variant.ID, Quantity: 1}},
	}

	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, "4PX Europe Express", leg.ServiceName)

	publicServices, err := shippingService.ListPublicCarrierServices()
	require.NoError(t, err)
	require.Len(t, publicServices, 1)
	assert.Equal(t, `["US"]`, publicServices[0].Countries)

	selectedQuoteInput := input
	selectedQuoteInput.ShippingQuoteID = quote.ID
	selectedQuoteInput.SelectedQuotePlanID = quote.SelectedPlan.ID

	channel.Countries = `["CA"]`
	require.NoError(t, db.Save(&channel).Error)
	lockedQuote, err := shippingService.QuoteCart(selectedQuoteInput)
	require.NoError(t, err)
	assert.Equal(t, service.ID, requireSelectedQuoteLeg(t, lockedQuote).CarrierServiceID)

	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)

	input.Country = "CA"
	currentQuote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, service.ID, requireSelectedQuoteLeg(t, currentQuote).CarrierServiceID)

	publicServices, err = shippingService.ListPublicCarrierServices()
	require.NoError(t, err)
	require.Len(t, publicServices, 1)
	assert.Equal(t, `["CA"]`, publicServices[0].Countries)

	publicService, err := shippingService.GetPublicCarrierService(service.ID)
	require.NoError(t, err)
	assert.Equal(t, "4PX Europe Express", publicService.ServiceName)
	assert.Equal(t, `["CA"]`, publicService.Countries)
}

func TestQuoteCartIgnoresTestEnvironmentFpxChannels(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createGlobalWeightQuoteTemplateForPublishedCollectionTest(t, db)
	product, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "4PX", "4PX")
	productionChannel := shippingdomain.FpxChannel{
		Environment: "production", ServiceCode: "SHARED-ROUTE", DisplayName: "Production route", Countries: `["US"]`, Enabled: true,
	}
	require.NoError(t, db.Create(&productionChannel).Error)
	require.NoError(t, db.Create(&shippingdomain.FpxChannel{
		Environment: "test", ServiceCode: "SHARED-ROUTE", DisplayName: "Test route", Countries: `["CA"]`, Enabled: true,
	}).Error)
	seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		FpxChannelID: &productionChannel.ID,
		ServiceCode:  "SHARED-ROUTE", ServiceName: "Stored route", Countries: `["US","CA"]`,
		BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})

	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: product.ID, VariantID: &variant.ID, Quantity: 1}},
	}
	productionQuote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, "Production route", requireSelectedQuoteLeg(t, productionQuote).ServiceName)

	input.Country = "CA"
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
}

func TestQuoteCartDoesNotFallBackWhenYanwenCollectionRouteIsDisabled(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createGlobalWeightQuoteTemplateForPublishedCollectionTest(t, db)
	product, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "Yanwen", "YANWEN")
	channel := shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "481", DisplayName: "Yanwen production standard", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&channel).Error)
	require.NoError(t, db.Create(&shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentFAT,
		ProductCode: "481", DisplayName: "Yanwen FAT standard", Countries: `["CA"]`, Enabled: true, VolumetricDivisor: 8000,
	}).Error)
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		YanwenPublishedChannelID: &channel.ID,
		ServiceCode:              "YANWEN:481", ServiceName: "Old Yanwen name", Countries: `["US"]`,
		BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})
	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: product.ID, VariantID: &variant.ID, Quantity: 1}},
	}

	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, "Yanwen production standard", requireSelectedQuoteLeg(t, quote).ServiceName)

	channel.Enabled = false
	require.NoError(t, db.Save(&channel).Error)
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)

	publicServices, err := shippingService.ListPublicCarrierServices()
	require.NoError(t, err)
	assert.Empty(t, publicServices)

	_, err = shippingService.GetPublicCarrierService(service.ID)
	require.ErrorIs(t, err, ErrShippingNotFound)
}

func TestQuoteCartDoesNotRebindStableFpxCollectionIDAfterCodeReuse(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createGlobalWeightQuoteTemplateForPublishedCollectionTest(t, db)
	product, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "4PX", "4PX")
	oldChannel := shippingdomain.FpxChannel{
		Environment: "production", ServiceCode: "REUSED-CODE", DisplayName: "Old collection", Countries: `["US"]`, Enabled: true,
	}
	require.NoError(t, db.Create(&oldChannel).Error)
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		FpxChannelID: &oldChannel.ID,
		ServiceCode:  "REUSED-CODE", ServiceName: "Stored route", Countries: `["US"]`,
		BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})

	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: product.ID, VariantID: &variant.ID, Quantity: 1}},
	}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, "Old collection", requireSelectedQuoteLeg(t, quote).ServiceName)

	require.NoError(t, db.Delete(&oldChannel).Error)
	newChannel := shippingdomain.FpxChannel{
		Environment: "production", ServiceCode: "REUSED-CODE", DisplayName: "New collection", Countries: `["US"]`, Enabled: true,
	}
	require.NoError(t, db.Create(&newChannel).Error)

	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
	publicService, publicErr := shippingService.GetPublicCarrierService(service.ID)
	require.ErrorIs(t, publicErr, ErrShippingNotFound)
	assert.Nil(t, publicService)
}

func TestQuoteCartDoesNotRebindStableYanwenCollectionIDAfterCodeReuse(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createGlobalWeightQuoteTemplateForPublishedCollectionTest(t, db)
	product, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "Yanwen", "YANWEN")
	oldChannel := shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "481", DisplayName: "Old Yanwen collection", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&oldChannel).Error)
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		YanwenPublishedChannelID: &oldChannel.ID,
		ServiceCode:              "YANWEN:481", ServiceName: "Stored Yanwen route", Countries: `["US"]`,
		BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})

	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: product.ID, VariantID: &variant.ID, Quantity: 1}},
	}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, "Old Yanwen collection", requireSelectedQuoteLeg(t, quote).ServiceName)

	require.NoError(t, db.Delete(&oldChannel).Error)
	newChannel := shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "481", DisplayName: "New Yanwen collection", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&newChannel).Error)

	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
	publicService, publicErr := shippingService.GetPublicCarrierService(service.ID)
	require.ErrorIs(t, publicErr, ErrShippingNotFound)
	assert.Nil(t, publicService)
}

func createGlobalWeightQuoteTemplateForPublishedCollectionTest(t *testing.T, db *gorm.DB) shippingdomain.ShippingTemplate {
	t.Helper()
	template := shippingdomain.ShippingTemplate{
		Name: "Published collection quote template", Type: "weight", Currency: "USD", DefaultFeeMinor: 9900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "*", MinValue: 0, MaxValue: 1, FeeMinor: 500},
			{Region: "*", MinValue: 1, MaxValue: 2, FeeMinor: 900},
		},
	}
	require.NoError(t, db.Create(&template).Error)
	return template
}

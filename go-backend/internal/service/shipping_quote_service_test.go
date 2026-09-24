package service

import (
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	productdomain "commerce-platform/internal/domain/product"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShippingServicePublicCatalogFiltersDisabledResources(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)

	enabledTemplate := shippingdomain.ShippingTemplate{Name: "Enabled", Type: "weight", DefaultFeeMinor: 1000, Enabled: true}
	disabledTemplate := shippingdomain.ShippingTemplate{Name: "Disabled", Type: "weight", DefaultFeeMinor: 2000, Enabled: false}
	require.NoError(t, shippingService.CreateTemplate(&enabledTemplate))
	require.NoError(t, shippingService.CreateTemplate(&disabledTemplate))

	templates, err := shippingService.ListPublicTemplates()
	require.NoError(t, err)
	require.Len(t, templates, 1)
	assert.Equal(t, enabledTemplate.ID, templates[0].ID)

	_, err = shippingService.GetPublicTemplate(disabledTemplate.ID)
	require.ErrorIs(t, err, ErrShippingNotFound)

	_, err = shippingService.CalculateShipping(ShippingCalculationInput{
		TemplateID: disabledTemplate.ID,
		Weight:     1,
		Country:    "US",
	})
	require.ErrorIs(t, err, ErrShippingNotFound)
}

func TestCalculateShippingRequiresCarrierWeightStepForAdditionalFee(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)

	template := shippingdomain.ShippingTemplate{
		Name:            "First unit shipping",
		Type:            "weight",
		DefaultFeeMinor: 9900,
		Enabled:         true,
		Rules: []shippingdomain.ShippingRule{
			{
				Region:          "US",
				MinValue:        0,
				MaxValue:        10,
				FeeMinor:        1500,
				AdditionalMinor: 500,
			},
		},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	_, err := shippingService.CalculateShipping(ShippingCalculationInput{
		TemplateID: template.ID,
		Weight:     0.8,
		Country:    "US",
	})
	require.ErrorIs(t, err, ErrShippingRateConfigurationInvalid)
}

func TestQuoteCartChargesCarrierSpecificAdditionalWeightStep(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name:            "500g carrier shipping",
		Type:            "weight",
		Currency:        "USD",
		DefaultFeeMinor: 9900,
		Enabled:         true,
		Rules: []shippingdomain.ShippingRule{{
			Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 1500, AdditionalMinor: 500,
		}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 50, 800, template.ID)
	carrier := seedQuoteCarrier(t, db, "International Line", "INTL")
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:           "INTL-500G",
		ServiceName:           "International 500g",
		Countries:             `["US"]`,
		BillingMode:           "actual_weight",
		FirstWeightGrams:      500,
		AdditionalWeightGrams: 500,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items:    []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	require.NoError(t, err)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, 1000, leg.BillableWeightGrams)
	assert.Equal(t, "20.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "20.00", quote.ShippingFeeDecimal)
}

func TestQuoteCartRejectsAdditionalWeightRuleWithoutCarrierScale(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Unconfigured weight scale", Type: "weight", Currency: "USD", DefaultFeeMinor: 9900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 1500, AdditionalMinor: 500}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 50, 800, template.ID)

	_, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items:    []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	require.ErrorIs(t, err, ErrShippingRateConfigurationInvalid)
}

func TestCreateCarrierServiceRejectsMissingWeightScaleForAdditionalRule(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Carrier scale required", Type: "weight", Currency: "USD", DefaultFeeMinor: 9900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 1500, AdditionalMinor: 500}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	carrier := seedQuoteCarrier(t, db, "Configuration Carrier", "CONFIG")
	service := shippingdomain.CarrierService{
		CarrierID: carrier.ID, TemplateID: &template.ID, ServiceCode: "CONFIG-MISSING-SCALE",
		ServiceName: "Missing scale", Currency: "USD", BillingMode: "actual_weight", Enabled: true,
	}

	err := shippingService.CreateCarrierService(&service)
	require.ErrorIs(t, err, ErrShippingRateConfigurationInvalid)
}

func TestCalculateShippingRejectsUnmatchedCountryWithoutPositiveDefaultFee(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Core markets only", Type: "weight", DefaultFeeMinor: 0, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 2000}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	quote, err := shippingService.CalculateShipping(ShippingCalculationInput{
		TemplateID: template.ID,
		Weight:     1,
		Country:    "BR",
	})

	assert.Nil(t, quote)
	require.ErrorIs(t, err, ErrCountryNotSupported)
	assert.Contains(t, err.Error(), `shipping template "Core markets only"`)
	assert.Contains(t, err.Error(), "country BR")
}

func TestQuoteCartRejectsUnmatchedCountryWithoutPositiveDefaultFee(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Core markets only", Type: "weight", DefaultFeeMinor: 0, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 2000}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 1200, 900, template.ID)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "BR",
		Currency: "USD",
		Items:    []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	assert.Nil(t, quote)
	require.ErrorIs(t, err, ErrCountryNotSupported)
	assert.Contains(t, err.Error(), "country BR")
}

func TestQuoteCartUsesPositiveDefaultFeeForUnmatchedCountry(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Worldwide fallback", Type: "weight", DefaultFeeMinor: 7500, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 2000}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 1200, 900, template.ID)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "BR",
		Currency: "USD",
		Items:    []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Equal(t, "75.00", quote.ShippingFeeDecimal)
	assert.False(t, quote.FreeShipping)
}

func TestQuoteCartAllowsExplicitZeroFeeRule(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Explicit Brazil promotion", Type: "weight", DefaultFeeMinor: 0, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "BR", MinValue: 0, MaxValue: 10, FeeMinor: 0}},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 1200, 900, template.ID)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "BR",
		Currency: "USD",
		Items:    []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	assert.Equal(t, "0.00", quote.ShippingFeeDecimal)
	assert.True(t, quote.FreeShipping)
}

func TestShippingServicePublicCarrierServicesRequireVisibleAssociations(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)

	enabledTemplate := shippingdomain.ShippingTemplate{Name: "Enabled", Type: "weight", DefaultFeeMinor: 1000, Enabled: true}
	disabledTemplate := shippingdomain.ShippingTemplate{Name: "Disabled", Type: "weight", DefaultFeeMinor: 2000, Enabled: false}
	require.NoError(t, shippingService.CreateTemplate(&enabledTemplate))
	require.NoError(t, shippingService.CreateTemplate(&disabledTemplate))

	enabledCarrier := shippingdomain.Carrier{Name: "Enabled Carrier", Code: "enabled", Enabled: true}
	disabledCarrier := shippingdomain.Carrier{Name: "Disabled Carrier", Code: "disabled", Enabled: false}
	require.NoError(t, shippingService.CreateCarrier(&enabledCarrier))
	require.NoError(t, shippingService.CreateCarrier(&disabledCarrier))

	visibleService := shippingdomain.CarrierService{
		CarrierID:   enabledCarrier.ID,
		TemplateID:  &enabledTemplate.ID,
		ServiceCode: "visible",
		ServiceName: "Visible",
		Enabled:     true,
	}
	hiddenByCarrier := shippingdomain.CarrierService{
		CarrierID:   disabledCarrier.ID,
		TemplateID:  &enabledTemplate.ID,
		ServiceCode: "hidden-carrier",
		ServiceName: "Hidden Carrier",
		Enabled:     true,
	}
	hiddenByTemplate := shippingdomain.CarrierService{
		CarrierID:   enabledCarrier.ID,
		TemplateID:  &disabledTemplate.ID,
		ServiceCode: "hidden-template",
		ServiceName: "Hidden Template",
		Enabled:     true,
	}
	require.NoError(t, shippingService.CreateCarrierService(&visibleService))
	require.NoError(t, shippingService.CreateCarrierService(&hiddenByCarrier))
	require.NoError(t, shippingService.CreateCarrierService(&hiddenByTemplate))

	services, err := shippingService.ListPublicCarrierServices()
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, visibleService.ID, services[0].ID)

	_, err = shippingService.GetPublicCarrierService(hiddenByCarrier.ID)
	require.ErrorIs(t, err, ErrShippingNotFound)

	_, err = shippingService.GetPublicCarrierService(hiddenByTemplate.ID)
	require.ErrorIs(t, err, ErrShippingNotFound)
}

func TestShippingServicePersistsDisplayPriceSnapshotsByMoneyField(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)

	template := shippingdomain.ShippingTemplate{
		Name:               "Display priced shipping",
		Type:               "price",
		DefaultFeeMinor:    2000,
		FreeThresholdMinor: 10000,
		Enabled:            true,
		DisplayPriceData: shippingdomain.TemplateDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
			shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee: {
				{AmountDecimal: "2.80", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
			},
			shippingdomain.ShippingTemplateDisplayPriceFieldFreeThreshold: {
				{AmountDecimal: "14.00", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
			},
		}),
		Rules: []shippingdomain.ShippingRule{
			{
				Region:          "US",
				MinValueMinor:   10000,
				MaxValue:        300,
				FeeMinor:        1500,
				AdditionalMinor: 200,
				DisplayPriceData: shippingdomain.RuleDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
					shippingdomain.ShippingRuleDisplayPriceFieldMinValue: {
						{AmountDecimal: "14.00", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
					},
					shippingdomain.ShippingRuleDisplayPriceFieldFee: {
						{AmountDecimal: "2.10", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
					},
				}),
			},
		},
	}

	require.NoError(t, shippingService.CreateTemplate(&template))

	found, err := shippingService.GetTemplate(template.ID)
	require.NoError(t, err)
	templateSnapshots := currency.ParseDisplayPriceSnapshotMap(found.DisplayPriceData, shippingdomain.ShippingTemplateDisplayPriceFields...)
	require.Len(t, templateSnapshots, 2)
	require.Len(t, templateSnapshots[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee], 1)
	assert.Equal(t, "2.80", templateSnapshots[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee][0].AmountDecimal)

	require.Len(t, found.Rules, 1)
	ruleSnapshots := currency.ParseDisplayPriceSnapshotMap(found.Rules[0].DisplayPriceData, shippingdomain.ShippingRuleDisplayPriceFields...)
	require.Len(t, ruleSnapshots, 2)
	require.Len(t, ruleSnapshots[shippingdomain.ShippingRuleDisplayPriceFieldFee], 1)
	assert.Equal(t, "2.10", ruleSnapshots[shippingdomain.ShippingRuleDisplayPriceFieldFee][0].AmountDecimal)

	found.DefaultFeeMinor = 2500
	found.DisplayPriceData = shippingdomain.TemplateDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
		shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee: {
			{AmountDecimal: "3.50", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
		},
	})
	found.Rules = []shippingdomain.ShippingRule{
		{
			Region:   "US",
			FeeMinor: 1800,
			DisplayPriceData: shippingdomain.RuleDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
				shippingdomain.ShippingRuleDisplayPriceFieldFee: {
					{AmountDecimal: "2.52", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
				},
			}),
		},
	}
	require.NoError(t, shippingService.UpdateTemplate(found))

	updated, err := shippingService.GetTemplate(template.ID)
	require.NoError(t, err)
	updatedTemplateSnapshots := currency.ParseDisplayPriceSnapshotMap(updated.DisplayPriceData, shippingdomain.ShippingTemplateDisplayPriceFields...)
	require.Len(t, updatedTemplateSnapshots[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee], 1)
	assert.Equal(t, "3.50", updatedTemplateSnapshots[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee][0].AmountDecimal)
	require.Len(t, updated.Rules, 1)
	updatedRuleSnapshots := currency.ParseDisplayPriceSnapshotMap(updated.Rules[0].DisplayPriceData, shippingdomain.ShippingRuleDisplayPriceFields...)
	require.Len(t, updatedRuleSnapshots[shippingdomain.ShippingRuleDisplayPriceFieldFee], 1)
	assert.Equal(t, "2.52", updatedRuleSnapshots[shippingdomain.ShippingRuleDisplayPriceFieldFee][0].AmountDecimal)
}

func TestQuoteCartReturnsDisplayPriceFromStoredShippingSnapshots(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)

	template := shippingdomain.ShippingTemplate{
		Name:            "Stored display quote template",
		Type:            "weight",
		Currency:        "CNY",
		DefaultFeeMinor: 9900,
		Enabled:         true,
		Rules: []shippingdomain.ShippingRule{
			{
				Region:          "US",
				Currency:        "CNY",
				MinValue:        0,
				MaxValue:        10,
				FeeMinor:        500,
				AdditionalMinor: 200,
				DisplayPriceData: shippingdomain.RuleDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
					shippingdomain.ShippingRuleDisplayPriceFieldFee: {
						{AmountDecimal: "0.70", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
					},
					shippingdomain.ShippingRuleDisplayPriceFieldAdditional: {
						{AmountDecimal: "0.28", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
					},
				}),
			},
		},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 50, 2500, template.ID)
	carrier := seedQuoteCarrier(t, db, "Display Carrier", "DISPLAY")
	seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "DISPLAY-500G", ServiceName: "Display 500g", Countries: `["US"]`, BillingMode: "actual_weight",
		FirstWeightGrams: 1000, AdditionalWeightGrams: 1000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:         "US",
		Currency:        "CNY",
		DisplayCurrency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "9.00", quote.ShippingFeeDecimal)
	require.NotNil(t, quote.DisplayPrice)
	assert.Equal(t, "USD", quote.DisplayPrice.Currency)
	assert.Equal(t, "1.26", quote.DisplayPrice.AmountDecimal)
	require.Len(t, quote.DisplayPrices, 1)
	assert.Equal(t, "1.26", quote.DisplayPrices[0].AmountDecimal)
}

func TestQuoteCartOmitsIncompleteShippingDisplayPrice(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)

	template := shippingdomain.ShippingTemplate{
		Name:            "Incomplete display quote template",
		Type:            "weight",
		Currency:        "CNY",
		DefaultFeeMinor: 9900,
		Enabled:         true,
		Rules: []shippingdomain.ShippingRule{
			{
				Region:          "US",
				Currency:        "CNY",
				MinValue:        0,
				MaxValue:        10,
				FeeMinor:        500,
				AdditionalMinor: 200,
				DisplayPriceData: shippingdomain.RuleDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
					shippingdomain.ShippingRuleDisplayPriceFieldFee: {
						{AmountDecimal: "0.70", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
					},
				}),
			},
		},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	record, variant := seedQuoteProduct(t, db, 50, 2500, template.ID)
	carrier := seedQuoteCarrier(t, db, "Incomplete Display Carrier", "DISPLAY-INCOMPLETE")
	seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "DISPLAY-INCOMPLETE-500G", ServiceName: "Display incomplete 500g", Countries: `["US"]`, BillingMode: "actual_weight",
		FirstWeightGrams: 1000, AdditionalWeightGrams: 1000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:         "US",
		Currency:        "CNY",
		DisplayCurrency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "9.00", quote.ShippingFeeDecimal)
	assert.Nil(t, quote.DisplayPrice)
	assert.Empty(t, quote.DisplayPrices)
}

func TestQuoteCartUsesSkuWeightWhenNoPackagingRule(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Items, 1)
	assert.Equal(t, "5.00", quote.ShippingFeeDecimal)
	assert.Equal(t, 900, quote.Items[0].WeightGrams)
	assert.Equal(t, 0, quote.Items[0].PackagingWeightGrams)
	assert.Equal(t, 900, quote.Items[0].ChargeWeightGrams)
	assert.Nil(t, quote.Items[0].PackagingRuleID)
}

func TestQuoteCartAddsPackagingWeightToChargeWeight(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)

	packagingRule := shippingdomain.PackagingRule{
		RuleName:  "Bike frame carton",
		BoxWeight: 0.2,
		IsActive:  true,
	}
	require.NoError(t, db.Create(&packagingRule).Error)
	require.NoError(t, db.Create(&shippingdomain.PackagingRuleApply{
		RuleID:    packagingRule.ID,
		ProductID: record.ID,
	}).Error)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Items, 1)
	item := quote.Items[0]
	require.NotNil(t, item.PackagingRuleID)
	assert.Equal(t, packagingRule.ID, *item.PackagingRuleID)
	assert.Equal(t, "Bike frame carton", item.PackagingRuleName)
	assert.Equal(t, 900, item.WeightGrams)
	assert.Equal(t, 200, item.PackagingWeightGrams)
	assert.Equal(t, 1100, item.ChargeWeightGrams)
	assert.Equal(t, "9.00", quote.ShippingFeeDecimal)
	assert.Equal(t, "9.00", item.ShippingFeeDecimal)
}

func TestQuoteResolvedItemsAddsConfiguredPackagingWeightDelta(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)

	quote, err := shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{
			ProductID: record.ID, VariantID: &variant.ID, ShippingTemplateID: &template.ID,
			Quantity: 1, UnitPriceMinor: 5000, WeightGrams: 900, PackagingWeightDeltaGrams: 75,
		}},
	})
	require.NoError(t, err)
	require.Len(t, quote.Items, 1)
	assert.Equal(t, 75, quote.Items[0].PackagingWeightGrams)
	assert.Equal(t, 975, quote.Items[0].ChargeWeightGrams)
}

func TestQuoteCartUsesLowestCarrierServiceOptionWhenAvailable(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")

	expensiveService := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:                 "DHL-EXP",
		ServiceName:                 "DHL Express",
		Countries:                   `["US"]`,
		BillingMode:                 "actual_weight",
		FirstWeightGrams:            500,
		AdditionalWeightGrams:       500,
		FuelSurchargePercentDecimal: "10",
		RemoteSurchargeMinor:        100,
		SortOrder:                   2,
	})
	cheapService := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:           "DHL-STD",
		ServiceName:           "DHL Standard",
		Countries:             `["US"]`,
		BillingMode:           "actual_weight",
		FirstWeightGrams:      500,
		AdditionalWeightGrams: 500,
		SortOrder:             1,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 2)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, cheapService.ID, leg.CarrierServiceID)
	assert.Equal(t, "5.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "5.00", quote.ShippingFeeDecimal)
	assert.Equal(t, "5.00", quote.Items[0].ShippingFeeDecimal)

	require.Len(t, quote.Plans[1].Legs, 1)
	expensiveLeg := quote.Plans[1].Legs[0]
	assert.Equal(t, expensiveService.ID, expensiveLeg.CarrierServiceID)
	assert.Equal(t, 1000, expensiveLeg.BillableWeightGrams)
	assert.Equal(t, "5.00", expensiveLeg.BaseFeeDecimal)
	assert.Equal(t, "0.50", expensiveLeg.FuelSurchargeDecimal)
	assert.Equal(t, "1.00", expensiveLeg.RemoteSurchargeDecimal)
	assert.Equal(t, "6.50", expensiveLeg.ShippingFeeDecimal)
}

func TestQuoteCartUsesRequestedCarrierServiceOption(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")
	cheapService := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "DHL-STD", ServiceName: "DHL Standard", Countries: `["US"]`, BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})
	expressService := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "DHL-EXP", ServiceName: "DHL Express", Countries: `["US"]`, BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500, RemoteSurchargeMinor: 10000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 2)
	expressPlan := quote.Plans[1]
	selectedQuote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD", ShippingQuoteID: quote.ID, SelectedQuotePlanID: expressPlan.ID,
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})
	require.NoError(t, err)
	leg := requireSelectedQuoteLeg(t, selectedQuote)
	assert.Equal(t, expressService.ID, leg.CarrierServiceID)
	assert.Equal(t, "105.00", selectedQuote.ShippingFeeDecimal)
	assert.Equal(t, "5.00", quote.Plans[0].ShippingFeeDecimal)
	assert.Equal(t, cheapService.ID, quote.Plans[0].Legs[0].CarrierServiceID)
}

func TestQuoteCartRejectsUnavailableRequestedCarrierService(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")
	seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "DHL-STD", ServiceName: "DHL Standard", Countries: `["US"]`, BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})
	require.NoError(t, err)

	_, err = shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD", ShippingQuoteID: quote.ID, SelectedQuotePlanID: "999999",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})

	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
}

func TestQuoteCartRestoresLockedPlanAfterRateChange(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "Lock Carrier", "LOCK")
	seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "LOCK-STD", ServiceName: "Locked Standard", Countries: `["US"]`,
		BillingMode: "actual_weight",
	})
	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	}

	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	require.NotEmpty(t, quote.ID)
	require.NotEmpty(t, quote.RateVersion)
	require.True(t, quote.ExpiresAt.After(time.Now()))
	require.NotNil(t, quote.SelectedPlan)
	lockedFee := quote.ShippingFeeDecimal

	require.NoError(t, db.Model(&shippingdomain.ShippingRule{}).
		Where("template_id = ? AND min_value = ?", template.ID, 0).
		Update("fee_minor", 5000).Error)
	input.ShippingQuoteID = quote.ID
	input.SelectedQuotePlanID = quote.SelectedPlan.ID
	restored, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, quote.ID, restored.ID)
	assert.Equal(t, quote.RateVersion, restored.RateVersion)
	assert.Equal(t, lockedFee, restored.ShippingFeeDecimal)

	input.ShippingQuoteID = ""
	input.SelectedQuotePlanID = ""
	fresh, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, "50.00", fresh.ShippingFeeDecimal)
}

func TestQuoteCartRejectsSnapshotForChangedItems(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)

	input.ShippingQuoteID = quote.ID
	input.SelectedQuotePlanID = quote.SelectedPlan.ID
	input.Items[0].Quantity = 2
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuoteStale)
}

func TestQuoteCartRejectsExpiredSnapshot(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	input := ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	require.NoError(t, db.Model(&shippingdomain.QuoteSnapshot{}).
		Where("id = ?", quote.ID).
		Update("expires_at", time.Now().Add(-time.Minute)).Error)

	input.ShippingQuoteID = quote.ID
	input.SelectedQuotePlanID = quote.SelectedPlan.ID
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuoteExpired)
}

func TestQuoteCartBuildsCompleteCarrierPlanCombinations(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	firstTemplate := seedWeightQuoteTemplate(t, db)
	secondTemplate := seedWeightQuoteTemplate(t, db)
	firstProduct, firstVariant := seedQuoteProduct(t, db, 50, 900, firstTemplate.ID)
	secondProduct, secondVariant := seedQuoteProduct(t, db, 50, 900, secondTemplate.ID)
	carrier := seedQuoteCarrier(t, db, "Plan Carrier", "PLAN")
	for _, serviceInput := range []struct {
		templateID uint
		code       string
		surcharge  float64
	}{
		{templateID: firstTemplate.ID, code: "FIRST-A", surcharge: 0},
		{templateID: firstTemplate.ID, code: "FIRST-B", surcharge: 1},
		{templateID: secondTemplate.ID, code: "SECOND-A", surcharge: 0},
		{templateID: secondTemplate.ID, code: "SECOND-B", surcharge: 2},
	} {
		seedQuoteCarrierService(t, db, carrier.ID, serviceInput.templateID, shippingdomain.CarrierService{
			ServiceCode: serviceInput.code, ServiceName: serviceInput.code, Countries: `["US"]`,
			BillingMode: "actual_weight", RemoteSurchargeMinor: int64(serviceInput.surcharge * 100),
		})
	}

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: firstProduct.ID, VariantID: &firstVariant.ID, Quantity: 1},
			{ProductID: secondProduct.ID, VariantID: &secondVariant.ID, Quantity: 1},
		},
	})
	require.NoError(t, err)
	require.Len(t, quote.Plans, 4)
	assert.Equal(t, []string{"10.00", "11.00", "12.00", "13.00"}, []string{
		quote.Plans[0].ShippingFeeDecimal,
		quote.Plans[1].ShippingFeeDecimal,
		quote.Plans[2].ShippingFeeDecimal,
		quote.Plans[3].ShippingFeeDecimal,
	})
	for _, plan := range quote.Plans {
		require.Len(t, plan.Legs, 2)
		assert.NotEqual(t, plan.Legs[0].TemplateID, plan.Legs[1].TemplateID)
	}
}

func TestQuoteCartKeepsCarrierOptionsForMultipleShippingTemplates(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	largeTemplate := seedWeightQuoteTemplate(t, db)
	largeTemplate.Name = "Large item template"
	require.NoError(t, db.Save(&largeTemplate).Error)
	smallTemplate := seedWeightQuoteTemplate(t, db)
	smallTemplate.Name = "Accessory template"
	require.NoError(t, db.Save(&smallTemplate).Error)
	largeProduct, largeVariant := seedQuoteProduct(t, db, 50, 900, largeTemplate.ID)
	smallProduct, smallVariant := seedQuoteProduct(t, db, 10, 100, smallTemplate.ID)
	require.NoError(t, db.Model(&productdomain.ProductVariant{}).Where("id = ?", smallVariant.ID).Update("sku", "SKU-QUOTE-ACCESSORY").Error)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")
	largeService := seedQuoteCarrierService(t, db, carrier.ID, largeTemplate.ID, shippingdomain.CarrierService{
		ServiceCode: "DHL-LARGE", ServiceName: "DHL Large", Countries: `["US"]`, BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500, RemoteSurchargeMinor: 400,
	})
	smallService := seedQuoteCarrierService(t, db, carrier.ID, smallTemplate.ID, shippingdomain.CarrierService{
		ServiceCode: "DHL-SMALL", ServiceName: "DHL Small", Countries: `["US"]`, BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500, RemoteSurchargeMinor: 100,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: largeProduct.ID, VariantID: &largeVariant.ID, Quantity: 1},
			{ProductID: smallProduct.ID, VariantID: &smallVariant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 1)
	require.NotNil(t, quote.SelectedPlan)
	require.Len(t, quote.SelectedPlan.Legs, 2)
	assert.Equal(t, largeService.ID, quote.SelectedPlan.Legs[0].CarrierServiceID)
	assert.Equal(t, smallService.ID, quote.SelectedPlan.Legs[1].CarrierServiceID)
	assert.Equal(t, "15.00", quote.ShippingFeeDecimal)
	assert.Equal(t, "9.00", quote.Items[0].ShippingFeeDecimal)
	assert.Equal(t, "6.00", quote.Items[1].ShippingFeeDecimal)

	selectedQuote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:             "US",
		Currency:            "USD",
		ShippingQuoteID:     quote.ID,
		SelectedQuotePlanID: quote.SelectedPlan.ID,
		Items: []ShippingQuoteItemInput{
			{ProductID: largeProduct.ID, VariantID: &largeVariant.ID, Quantity: 1},
			{ProductID: smallProduct.ID, VariantID: &smallVariant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, selectedQuote.SelectedPlan)
	assert.Equal(t, quote.SelectedPlan.ID, selectedQuote.SelectedPlan.ID)
	assert.Equal(t, "15.00", selectedQuote.ShippingFeeDecimal)
	assert.Equal(t, "9.00", selectedQuote.Items[0].ShippingFeeDecimal)
	assert.Equal(t, "6.00", selectedQuote.Items[1].ShippingFeeDecimal)
}

func TestQuoteCartOmitsPartialCarrierDisplayPriceAcrossTemplates(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	largeTemplate := shippingdomain.ShippingTemplate{
		Name: "Large item display template", Type: "weight", Currency: "CNY", DefaultFeeMinor: 9900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{
			Region: "US", Currency: "CNY", MinValue: 0, MaxValue: 1, FeeMinor: 500,
			DisplayPriceData: shippingdomain.RuleDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
				shippingdomain.ShippingRuleDisplayPriceFieldFee: {
					{AmountDecimal: "0.70", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
				},
			}),
		}},
	}
	smallTemplate := shippingdomain.ShippingTemplate{
		Name: "Accessory display template", Type: "weight", Currency: "CNY", DefaultFeeMinor: 9900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{
			Region: "US", Currency: "CNY", MinValue: 0, MaxValue: 1, FeeMinor: 500,
			DisplayPriceData: shippingdomain.RuleDisplayPriceSnapshotsJSON(map[string][]currency.DisplayPriceSnapshot{
				shippingdomain.ShippingRuleDisplayPriceFieldFee: {
					{AmountDecimal: "0.70", Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true},
				},
			}),
		}},
	}
	shippingRepo := repository.NewShippingRepository(db)
	largeTemplateRules := append([]shippingdomain.ShippingRule(nil), largeTemplate.Rules...)
	smallTemplateRules := append([]shippingdomain.ShippingRule(nil), smallTemplate.Rules...)
	require.NoError(t, shippingRepo.CreateTemplateWithRules(&largeTemplate, largeTemplateRules))
	require.NoError(t, shippingRepo.CreateTemplateWithRules(&smallTemplate, smallTemplateRules))
	largeProduct, largeVariant := seedQuoteProduct(t, db, 50, 900, largeTemplate.ID)
	smallProduct, smallVariant := seedQuoteProduct(t, db, 10, 100, smallTemplate.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")
	seedQuoteCarrierService(t, db, carrier.ID, largeTemplate.ID, shippingdomain.CarrierService{
		ServiceCode: "DHL-REMOTE", ServiceName: "DHL Remote", Countries: `["US"]`, Currency: "CNY",
		BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500, RemoteSurchargeMinor: 100,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "CNY", DisplayCurrency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: largeProduct.ID, VariantID: &largeVariant.ID, Quantity: 1},
			{ProductID: smallProduct.ID, VariantID: &smallVariant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 1)
	assert.Equal(t, "11.00", quote.ShippingFeeDecimal)
	// Remote surcharge is now included in the stored display snapshot instead
	// of suppressing all multi-currency prices.
	require.NotNil(t, quote.DisplayPrice)
	assert.Equal(t, "1.54", quote.DisplayPrice.AmountDecimal)
	require.Len(t, quote.DisplayPrices, 1)
	require.NotNil(t, quote.Plans[0].DisplayPrice)
	assert.Equal(t, "1.54", quote.Plans[0].DisplayPrice.AmountDecimal)
	require.Len(t, quote.Plans[0].DisplayPrices, 1)
}

func TestQuoteCartUsesVariantPackagingRuleBeforeProductDefault(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, defaultVariant := seedQuoteProduct(t, db, 50, 900, template.ID)
	secondVariant := productdomain.ProductVariant{
		ProductID:    record.ID,
		SKU:          "SKU-QUOTE-VARIANT-2",
		Title:        "Second variant",
		OptionValues: `{"size":"large"}`,
		Currency:     "USD",
		PriceMinor:   5000,
		Weight:       900,
		Stock:        10,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&secondVariant).Error)

	defaultRule := shippingdomain.PackagingRule{RuleName: "Product default", BoxWeight: 0.1, IsActive: true}
	variantRule := shippingdomain.PackagingRule{RuleName: "Variant carton", BoxWeight: 0.5, IsActive: true}
	require.NoError(t, db.Create(&defaultRule).Error)
	require.NoError(t, db.Create(&variantRule).Error)
	require.NoError(t, shippingService.CreatePackagingRuleApply(&shippingdomain.PackagingRuleApply{
		RuleID: defaultRule.ID, ProductID: record.ID,
	}))
	require.NoError(t, shippingService.CreatePackagingRuleApply(&shippingdomain.PackagingRuleApply{
		RuleID: variantRule.ID, ProductID: record.ID, VariantID: &secondVariant.ID,
	}))

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &defaultVariant.ID, Quantity: 1},
			{ProductID: record.ID, VariantID: &secondVariant.ID, Quantity: 1},
		},
	})
	require.NoError(t, err)
	require.Len(t, quote.Items, 2)
	assert.Equal(t, 100, quote.Items[0].PackagingWeightGrams)
	assert.Equal(t, defaultRule.ID, *quote.Items[0].PackagingRuleID)
	assert.Equal(t, 500, quote.Items[1].PackagingWeightGrams)
	assert.Equal(t, variantRule.ID, *quote.Items[1].PackagingRuleID)
}

func TestQuoteCartAppliesRemoteSurchargeOnlyForConfiguredPostalRange(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "Postal Carrier", "POSTAL")
	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "POSTAL-REMOTE", ServiceName: "Postal Remote", Countries: `["US"]`,
		RemoteSurchargeMinor: 700, RemotePostalCodes: `["10000-10009"]`,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", PostalCode: "10005", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})
	require.NoError(t, err)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, "12.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "7.00", leg.RemoteSurchargeDecimal)

	quote, err = shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", PostalCode: "10020", Currency: "USD",
		Items: []ShippingQuoteItemInput{{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1}},
	})
	require.NoError(t, err)
	leg = requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, "5.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "0.00", leg.RemoteSurchargeDecimal)
}

func TestQuoteResolvedItemsKeepsFreeShippingThresholdWithinTemplateGroup(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	primaryTemplate := shippingdomain.ShippingTemplate{
		Name: "Primary free template", Type: "weight", Currency: "USD", DefaultFeeMinor: 2000,
		FreeShipping: true, FreeThresholdMinor: 10000, Enabled: true,
	}
	accessoryTemplate := shippingdomain.ShippingTemplate{
		Name: "Accessory free template", Type: "weight", Currency: "USD", DefaultFeeMinor: 500,
		FreeShipping: true, FreeThresholdMinor: 5000, Enabled: true,
	}
	require.NoError(t, db.Create(&primaryTemplate).Error)
	require.NoError(t, db.Create(&accessoryTemplate).Error)
	primary, primaryVariant := seedQuoteProduct(t, db, 200, 900, primaryTemplate.ID)
	accessory, accessoryVariant := seedQuoteProduct(t, db, 10, 100, accessoryTemplate.ID)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country: "US", Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: primary.ID, VariantID: &primaryVariant.ID, Quantity: 1},
			{ProductID: accessory.ID, VariantID: &accessoryVariant.ID, Quantity: 1},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "5.00", quote.ShippingFeeDecimal)
	assert.True(t, quote.Items[0].FreeShipping)
	assert.False(t, quote.Items[1].FreeShipping)
	assert.Equal(t, "5.00", quote.Items[1].ShippingFeeDecimal)
}

func TestQuoteCartUsesPackagingDimensionsForVolumetricCarrierService(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "YunExpress", "YUN")

	packagingRule := shippingdomain.PackagingRule{
		RuleName:  "Large wheelset carton",
		BoxWeight: 0.2,
		BoxLength: 60,
		BoxWidth:  40,
		BoxHeight: 30,
		IsActive:  true,
	}
	require.NoError(t, db.Create(&packagingRule).Error)
	require.NoError(t, db.Create(&shippingdomain.PackagingRuleApply{
		RuleID:    packagingRule.ID,
		ProductID: record.ID,
	}).Error)

	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:       "YUN-VOL",
		ServiceName:       "YunExpress Volumetric",
		Countries:         `["US"]`,
		BillingMode:       "greater_of_actual_and_volumetric",
		VolumetricDivisor: 6000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 1)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, 1100, leg.ActualWeightGrams)
	assert.Equal(t, 12000, leg.VolumetricWeightGrams)
	assert.Equal(t, 12000, leg.ChargeWeightGrams)
	assert.Equal(t, 12000, leg.BillableWeightGrams)
	assert.Equal(t, "99.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "99.00", quote.ShippingFeeDecimal)
}

func TestQuoteCartKeepsKnownVolumetricWeightWhenAccessoryLacksPackagingRule(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "DHL wheelset rates", Type: "weight", Currency: "USD", DefaultFeeMinor: 99900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "US", MinValue: 0, MaxValue: 2, FeeMinor: 1800},
			{Region: "US", MinValue: 20, MaxValue: 21, FeeMinor: 16000},
		},
	}
	require.NoError(t, db.Create(&template).Error)
	wheelset, wheelsetVariant := seedQuoteProduct(t, db, 1200, 1350, template.ID)
	accessory, accessoryVariant := seedQuoteProduct(t, db, 2, 20, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")

	packagingRule := shippingdomain.PackagingRule{
		RuleName:  "Wheelset carton",
		BoxLength: 83,
		BoxWidth:  68,
		BoxHeight: 18,
		IsActive:  true,
	}
	require.NoError(t, db.Create(&packagingRule).Error)
	require.NoError(t, db.Create(&shippingdomain.PackagingRuleApply{
		RuleID:    packagingRule.ID,
		ProductID: wheelset.ID,
	}).Error)

	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:       "DHL-WHEELSET",
		ServiceName:       "DHL Wheelset Express",
		Countries:         `["US"]`,
		BillingMode:       "greater_of_actual_and_volumetric",
		VolumetricDivisor: 5000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: wheelset.ID, VariantID: &wheelsetVariant.ID, Quantity: 1},
			{ProductID: accessory.ID, VariantID: &accessoryVariant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 1)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, 1370, leg.ActualWeightGrams)
	assert.Equal(t, 20319, leg.VolumetricWeightGrams)
	assert.Equal(t, 20339, leg.ChargeWeightGrams)
	assert.Equal(t, 20339, leg.BillableWeightGrams)
	assert.Equal(t, "160.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "160.00", quote.ShippingFeeDecimal)
}

func TestQuoteCartOmitsVolumetricOnlyCarrierServiceWhenAccessoryLacksPackagingRule(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := shippingdomain.ShippingTemplate{
		Name: "Volumetric-only rates", Type: "weight", Currency: "USD", DefaultFeeMinor: 99900, Enabled: true,
		Rules: []shippingdomain.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 2, FeeMinor: 1800}},
	}
	require.NoError(t, db.Create(&template).Error)
	wheelset, wheelsetVariant := seedQuoteProduct(t, db, 1200, 1350, template.ID)
	accessory, accessoryVariant := seedQuoteProduct(t, db, 2, 20, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")

	packagingRule := shippingdomain.PackagingRule{
		RuleName:  "Wheelset carton",
		BoxLength: 83,
		BoxWidth:  68,
		BoxHeight: 18,
		IsActive:  true,
	}
	require.NoError(t, db.Create(&packagingRule).Error)
	require.NoError(t, db.Create(&shippingdomain.PackagingRuleApply{
		RuleID:    packagingRule.ID,
		ProductID: wheelset.ID,
	}).Error)
	seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:       "DHL-VOLUMETRIC-ONLY",
		ServiceName:       "DHL Volumetric Only",
		Countries:         `["US"]`,
		BillingMode:       "volumetric_weight",
		VolumetricDivisor: 5000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: wheelset.ID, VariantID: &wheelsetVariant.ID, Quantity: 1},
			{ProductID: accessory.ID, VariantID: &accessoryVariant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, quote)
	require.Len(t, quote.Plans, 1)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Zero(t, leg.CarrierServiceID)
	assert.Equal(t, "template", leg.BillingMode)
	assert.Equal(t, "18.00", quote.ShippingFeeDecimal)
}

func TestQuoteCartFallsBackToActualWeightWhenGreaterOfLacksVolumetricWeight(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")

	service := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:       "DHL-GREATER-OF",
		ServiceName:       "DHL Greater Of",
		Countries:         `["US"]`,
		BillingMode:       "greater_of_actual_and_volumetric",
		VolumetricDivisor: 6000,
	})

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Plans, 1)
	leg := requireSelectedQuoteLeg(t, quote)
	assert.Equal(t, service.ID, leg.CarrierServiceID)
	assert.Equal(t, 900, leg.ActualWeightGrams)
	assert.Equal(t, 0, leg.VolumetricWeightGrams)
	assert.Equal(t, 900, leg.ChargeWeightGrams)
	assert.Equal(t, 900, leg.BillableWeightGrams)
	assert.Equal(t, "5.00", leg.ShippingFeeDecimal)
	assert.Equal(t, "5.00", quote.ShippingFeeDecimal)
}

func TestCreatePackagingRuleApplyRejectsSecondRuleForProduct(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	record, _ := seedQuoteProduct(t, db, 50, 900)

	firstRule := shippingdomain.PackagingRule{RuleName: "Small carton", BoxWeight: 0.1, IsActive: true}
	secondRule := shippingdomain.PackagingRule{RuleName: "Large carton", BoxWeight: 0.5, IsActive: true}
	require.NoError(t, db.Create(&firstRule).Error)
	require.NoError(t, db.Create(&secondRule).Error)

	require.NoError(t, shippingService.CreatePackagingRuleApply(&shippingdomain.PackagingRuleApply{
		RuleID:    firstRule.ID,
		ProductID: record.ID,
	}))

	err := shippingService.CreatePackagingRuleApply(&shippingdomain.PackagingRuleApply{
		RuleID:    secondRule.ID,
		ProductID: record.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "product already has a packaging rule")
}

func TestQuoteCartUsesVariantShippingTemplateOverride(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	productTemplate := seedWeightQuoteTemplate(t, db)
	variantTemplate := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, productTemplate.ID)
	require.NoError(t, db.Model(&productdomain.ProductVariant{}).
		Where("id = ?", variant.ID).
		Update("shipping_template_id", variantTemplate.ID).Error)

	quote, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.NoError(t, err)
	require.Len(t, quote.Items, 1)
	assert.Equal(t, variantTemplate.ID, quote.Items[0].TemplateID)
}

func TestQuoteCartRejectsMissingShippingTemplate(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	record, variant := seedQuoteProduct(t, db, 50, 900)

	_, err := shippingService.QuoteCart(ShippingQuoteInput{
		Country:  "US",
		Currency: "USD",
		Items: []ShippingQuoteItemInput{
			{ProductID: record.ID, VariantID: &variant.ID, Quantity: 1},
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "shipping template is missing")
}

func newTestShippingQuoteService(t *testing.T) (*gorm.DB, *ShippingService) {
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

	require.NoError(t, db.AutoMigrate(
		&productdomain.ProductSpecificationTemplate{},
		&productdomain.SpecDefinition{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{},
		&productdomain.ProductVariant{},
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.ShippingRule{},
		&shippingdomain.ShippingDisplayPriceSnapshot{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.QuoteSnapshot{},
		&shippingdomain.PackagingRule{},
		&shippingdomain.PackagingRuleApply{},
	))

	shippingRepo := repository.NewShippingRepository(db)
	productRepo := repository.NewProductRepository(db)
	return db, NewShippingService(shippingRepo, productRepo)
}

func seedQuoteProduct(t *testing.T, db *gorm.DB, price float64, weightGrams int, shippingTemplateIDs ...uint) (productdomain.Product, productdomain.ProductVariant) {
	t.Helper()
	priceMoney, err := domainmoney.FromMajorFloat(price, "USD")
	require.NoError(t, err)

	var shippingTemplateID *uint
	if len(shippingTemplateIDs) > 0 {
		shippingTemplateID = &shippingTemplateIDs[0]
	}
	var productCount int64
	require.NoError(t, db.Model(&productdomain.Product{}).Count(&productCount).Error)
	suffix := ""
	if productCount > 0 {
		suffix = fmt.Sprintf("-%d", productCount+1)
	}
	record := productdomain.Product{
		ShippingTemplateID: shippingTemplateID,
		SKU:                "SKU-QUOTE" + suffix,
		Name:               "Quote Product",
		Slug:               "quote-product" + suffix,
		PriceMinor:         priceMoney.AmountMinor(),
		Stock:              10,
		Status:             "active",
	}
	require.NoError(t, db.Create(&record).Error)

	variant := productdomain.ProductVariant{
		ProductID:    record.ID,
		SKU:          "SKU-QUOTE-DEFAULT" + suffix,
		Title:        "Default",
		OptionValues: "{}",
		PriceMinor:   priceMoney.AmountMinor(),
		Stock:        10,
		Weight:       weightGrams,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)

	return record, variant
}

func seedWeightQuoteTemplate(t *testing.T, db *gorm.DB) shippingdomain.ShippingTemplate {
	t.Helper()

	template := shippingdomain.ShippingTemplate{
		Name:            "Weight quote template",
		Type:            "weight",
		DefaultFeeMinor: 9900,
		Enabled:         true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "US", MinValue: 0, MaxValue: 1, FeeMinor: 500},
			{Region: "US", MinValue: 1, MaxValue: 2, FeeMinor: 900},
		},
	}
	require.NoError(t, db.Create(&template).Error)
	return template
}

func seedQuoteCarrier(t *testing.T, db *gorm.DB, name string, code string) shippingdomain.Carrier {
	t.Helper()

	carrier := shippingdomain.Carrier{
		Name:    name,
		Code:    code,
		Enabled: true,
	}
	require.NoError(t, db.Create(&carrier).Error)
	return carrier
}

func requireSelectedQuoteLeg(t *testing.T, quote *ShippingQuote) ShippingQuoteLeg {
	t.Helper()
	require.NotNil(t, quote)
	require.NotNil(t, quote.SelectedPlan)
	require.Len(t, quote.SelectedPlan.Legs, 1)
	return quote.SelectedPlan.Legs[0]
}

func seedQuoteCarrierService(t *testing.T, db *gorm.DB, carrierID uint, templateID uint, input shippingdomain.CarrierService) shippingdomain.CarrierService {
	t.Helper()

	input.CarrierID = carrierID
	input.TemplateID = &templateID
	if input.Countries == "" {
		input.Countries = "[]"
	}
	if input.Currency == "" {
		input.Currency = "USD"
	}
	if input.BillingMode == "" {
		input.BillingMode = "actual_weight"
	}
	if input.VolumetricDivisor == 0 {
		input.VolumetricDivisor = 6000
	}
	input.Enabled = true
	require.NoError(t, db.Create(&input).Error)
	return input
}

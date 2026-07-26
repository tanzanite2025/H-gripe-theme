package service

import (
	"testing"

	productdomain "tanzanite/internal/domain/product"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShippingServicePublicCatalogFiltersDisabledResources(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)

	enabledTemplate := shippingdomain.ShippingTemplate{Name: "Enabled", Type: "weight", DefaultFee: 10, Enabled: true}
	disabledTemplate := shippingdomain.ShippingTemplate{Name: "Disabled", Type: "weight", DefaultFee: 20, Enabled: false}
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

func TestShippingServicePublicCarrierServicesRequireVisibleAssociations(t *testing.T) {
	_, shippingService := newTestShippingQuoteService(t)

	enabledTemplate := shippingdomain.ShippingTemplate{Name: "Enabled", Type: "weight", DefaultFee: 10, Enabled: true}
	disabledTemplate := shippingdomain.ShippingTemplate{Name: "Disabled", Type: "weight", DefaultFee: 20, Enabled: false}
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
	assert.Equal(t, 5.0, quote.ShippingFee)
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
	assert.Equal(t, 9.0, quote.ShippingFee)
	assert.Equal(t, 9.0, item.ShippingFee)
}

func TestQuoteCartUsesLowestCarrierServiceOptionWhenAvailable(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, db)
	record, variant := seedQuoteProduct(t, db, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, db, "DHL", "DHL")

	expensiveService := seedQuoteCarrierService(t, db, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode:           "DHL-EXP",
		ServiceName:           "DHL Express",
		Countries:             `["US"]`,
		BillingMode:           "actual_weight",
		FirstWeightGrams:      500,
		AdditionalWeightGrams: 500,
		FuelSurchargePercent:  10,
		RemoteSurcharge:       1,
		SortOrder:             2,
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
	require.Len(t, quote.Options, 2)
	require.NotNil(t, quote.SelectedOption)
	assert.Equal(t, "carrier_service", quote.Source)
	assert.Equal(t, cheapService.ID, quote.SelectedOption.CarrierServiceID)
	assert.Equal(t, 5.0, quote.SelectedOption.ShippingFee)
	assert.Equal(t, 5.0, quote.ShippingFee)
	assert.Equal(t, 5.0, quote.Items[0].ShippingFee)

	assert.Equal(t, expensiveService.ID, quote.Options[1].CarrierServiceID)
	assert.Equal(t, 1000, quote.Options[1].BillableWeightGrams)
	assert.Equal(t, 5.0, quote.Options[1].BaseFee)
	assert.Equal(t, 0.5, quote.Options[1].FuelSurcharge)
	assert.Equal(t, 1.0, quote.Options[1].RemoteSurcharge)
	assert.Equal(t, 6.5, quote.Options[1].ShippingFee)
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
	require.Len(t, quote.Options, 1)
	require.NotNil(t, quote.SelectedOption)
	assert.Equal(t, service.ID, quote.SelectedOption.CarrierServiceID)
	assert.Equal(t, 1100, quote.SelectedOption.ActualWeightGrams)
	assert.Equal(t, 12000, quote.SelectedOption.VolumetricWeightGrams)
	assert.Equal(t, 12000, quote.SelectedOption.ChargeWeightGrams)
	assert.Equal(t, 12000, quote.SelectedOption.BillableWeightGrams)
	assert.Equal(t, 99.0, quote.SelectedOption.ShippingFee)
	assert.Equal(t, 99.0, quote.ShippingFee)
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
		&productdomain.ProductType{},
		&productdomain.SpecDefinition{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{},
		&productdomain.ProductVariant{},
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.ShippingRule{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.PackagingRule{},
		&shippingdomain.PackagingRuleApply{},
	))

	shippingRepo := repository.NewShippingRepository(db)
	productRepo := repository.NewProductRepository(db)
	return db, NewShippingService(shippingRepo, productRepo)
}

func seedQuoteProduct(t *testing.T, db *gorm.DB, price float64, weightGrams int, shippingTemplateIDs ...uint) (productdomain.Product, productdomain.ProductVariant) {
	t.Helper()

	var shippingTemplateID *uint
	if len(shippingTemplateIDs) > 0 {
		shippingTemplateID = &shippingTemplateIDs[0]
	}
	record := productdomain.Product{
		ShippingTemplateID: shippingTemplateID,
		SKU:                "SKU-QUOTE",
		Name:               "Quote Product",
		Slug:               "quote-product",
		Price:              price,
		Stock:              10,
		Status:             "active",
	}
	require.NoError(t, db.Create(&record).Error)

	variant := productdomain.ProductVariant{
		ProductID:    record.ID,
		SKU:          "SKU-QUOTE-DEFAULT",
		Title:        "Default",
		OptionValues: "{}",
		Price:        price,
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
		Name:       "Weight quote template",
		Type:       "weight",
		DefaultFee: 99,
		Enabled:    true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "US", MinValue: 0, MaxValue: 1, Fee: 5},
			{Region: "US", MinValue: 1, MaxValue: 2, Fee: 9},
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

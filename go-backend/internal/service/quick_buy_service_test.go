package service

import (
	"encoding/json"
	"testing"
	"time"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/quickbuy"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestQuickBuySessionTotalsUsesCurrencyMinorUnits(t *testing.T) {
	total, weight, err := quickBuySessionTotals([]quickbuy.SessionItem{{
		Quantity:               2,
		UnitPriceSnapshotMinor: 10000,
		CurrencySnapshot:       "JPY",
		WeightSnapshotG:        300,
	}})
	require.NoError(t, err)
	require.Equal(t, int64(20000), total)
	require.Equal(t, 600, weight)
}

func TestQuickBuyRejectsSpokeRepairKitWithoutWheelsetSelection(t *testing.T) {
	service := &QuickBuyService{}
	item := productdomain.Product{
		ID: 42,
		ProductCategory: &productdomain.ProductCategory{
			Slug: productdomain.SpokeRepairKitProductCategorySlug,
		},
	}

	err := service.validateQuickBuyProductAllowedForStep(quickbuy.Step{StepKey: "product"}, item)
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	require.ErrorContains(t, err, "compatible wheelset model selection")
}

func TestQuickBuySessionTotalsRejectsInvalidMoney(t *testing.T) {
	total, weight, err := quickBuySessionTotals([]quickbuy.SessionItem{{
		ID:                     9,
		Quantity:               1,
		UnitPriceSnapshotMinor: 10,
		CurrencySnapshot:       "XXX",
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "price")
	require.Zero(t, total)
	require.Zero(t, weight)
}

func TestQuickBuySessionTotalsRejectsNonPositiveQuantity(t *testing.T) {
	_, _, err := quickBuySessionTotals([]quickbuy.SessionItem{{
		ID:                     10,
		Quantity:               0,
		UnitPriceSnapshotMinor: 10,
		CurrencySnapshot:       "USD",
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "quantity")
}

func TestNormalizeQuickBuyCurrencyRejectsUnsupportedValues(t *testing.T) {
	currency, err := normalizeQuickBuyCurrency("")
	require.NoError(t, err)
	require.Equal(t, productdomain.DefaultPriceCurrency, currency)

	currency, err = normalizeQuickBuyCurrency("jpy")
	require.NoError(t, err)
	require.Equal(t, "JPY", currency)

	currency, err = normalizeQuickBuyCurrency("XXX")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	require.Empty(t, currency)
}

func TestQuickBuySelectionRejectsCurrencyMismatch(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	template := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "currency-guard",
		Name:         "Currency Guard",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{Steps: []QuickBuyStepInput{{
			StepKey:            "rim",
			Name:               "Rim",
			ProductCategoryIDs: []uint{template.ID},
		}}},
	})
	require.NoError(t, err)
	published, err := quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	productRecord := seedQuickBuyProductWithDetails(t, db, template.ID, "CURRENCY-GUARD", "Currency Guard Product", "currency-guard-product", 10)
	var variant productdomain.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)

	session, err := quickBuyService.CreateSession(QuickBuySessionInput{
		FlowVersionID: published.Version.ID,
		Currency:      "EUR",
	})
	require.NoError(t, err)
	_, err = quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{Selections: []QuickBuySelectionInput{{
		StepKey:   "rim",
		ProductID: productRecord.ID,
		VariantID: &variant.ID,
		Quantity:  1,
	}}})
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	require.Contains(t, err.Error(), "does not match")
}

func TestQuickBuyServiceCreatesPublishesAndReturnsCurrentFlow(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wheelset-build",
		Name:         "Wheelset Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{productSpecificationTemplate.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "draft", created.Version.Status)
	require.Len(t, created.Steps, 1)

	published, err := quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)
	require.Equal(t, quickbuy.FlowVersionStatusPublished, published.Version.Status)

	current, err := quickBuyService.CurrentFlow("dock", "en")
	require.NoError(t, err)
	require.NotNil(t, current)
	require.Len(t, current.Steps, 1)
	assert.Equal(t, "rim", current.Steps[0].StepKey)
	assert.Equal(t, "rim", current.Steps[0].Slug)
	require.Len(t, current.Steps[0].ProductCategories, 1)
	assert.Equal(t, "rim", current.Steps[0].ProductCategories[0].Slug)
}

func TestQuickBuyServiceProtectsDefaultQuickBuildSteps(t *testing.T) {
	_, quickBuyService := newQuickBuyTestService(t)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "quick-build",
		Name:         "QUICK Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{StepKey: "product-search", Name: "Rim"},
				{StepKey: "specifications", Name: "Hub"},
				{StepKey: "quantity", Name: "Spokes"},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Steps, 3)
	assert.Equal(t, "Rim", created.Steps[0].Name)

	_, err = quickBuyService.UpdateDraftVersion(created.Version.ID, QuickBuyVersionInput{
		Steps: []QuickBuyStepInput{
			{StepKey: "product-search", Name: "Custom Step A"},
			{StepKey: "specifications", Name: "Custom Step B"},
		},
	})
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	assert.Contains(t, err.Error(), "quantity")
}

func TestQuickBuyServicePublishesDefaultFlowWithoutConfiguredProductSpecificationTemplates(t *testing.T) {
	_, quickBuyService := newQuickBuyTestService(t)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "quick-build",
		Name:         "QUICK Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{StepKey: "product-search", Name: "Wheelset custom configuration"},
				{StepKey: "specifications", Name: "Specifications"},
				{StepKey: "quantity", Name: "Quantity"},
			},
		},
	})
	require.NoError(t, err)

	published, err := quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)
	require.Equal(t, quickbuy.FlowVersionStatusPublished, published.Version.Status)
}

func TestQuickBuyServiceAllowsAdditionalQuickBuildSteps(t *testing.T) {
	_, quickBuyService := newQuickBuyTestService(t)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "quick-build",
		Name:         "QUICK Build With Extra Step",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{StepKey: "product-search", Name: "Custom Step A"},
				{StepKey: "specifications", Name: "Custom Step B"},
				{StepKey: "quantity", Name: "Custom Step C"},
				{StepKey: "fit-check", Name: "Fit check"},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Steps, 4)
	assert.Equal(t, "fit-check", created.Steps[3].StepKey)
}

func TestQuickBuyServiceFiltersDefaultFlowCandidatesByProductCategory(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	wheelsets := seedQuickBuyProductCategory(t, db, "Wheelsets", "wheelset", nil)
	carbonWheelsets := seedQuickBuyProductCategory(t, db, "Carbon Wheelsets", "carbon-wheelset", &wheelsets.ID)
	tires := seedQuickBuyProductCategory(t, db, "Tires", "tires", nil)
	inScope := seedQuickBuyProductWithDetails(t, db, productSpecificationTemplate.ID, "QB-WHEEL-001", "Quick Buy Wheel", "quick-buy-wheel", 100, carbonWheelsets.ID)
	outOfScope := seedQuickBuyProductWithDetails(t, db, productSpecificationTemplate.ID, "QB-TIRE-001", "Quick Buy Tire", "quick-buy-tire", 50, tires.ID)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "quick-build",
		Name:         "QUICK Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{StepKey: "product-search", Name: "Wheelset custom configuration", ProductCategoryIDs: []uint{wheelsets.ID}},
				{StepKey: "specifications", Name: "Specifications", ProductCategoryIDs: []uint{wheelsets.ID}},
				{StepKey: "quantity", Name: "Quantity", ProductCategoryIDs: []uint{wheelsets.ID}},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Steps, 3)
	require.Len(t, created.Steps[0].ProductCategories, 1)
	assert.Equal(t, "wheelset", created.Steps[0].ProductCategories[0].Slug)

	result, err := quickBuyService.PreviewVersionStepCandidates(created.Version.ID, QuickBuyCandidateInput{
		StepKey:  "product-search",
		Locale:   "en",
		PageSize: 12,
	})
	require.NoError(t, err)
	require.Len(t, result.Products, 1)
	assert.Equal(t, inScope.ID, result.Products[0].ID)

	published, err := quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)
	session, err := quickBuyService.CreateSession(QuickBuySessionInput{
		FlowVersionID: published.Version.ID,
		Surface:       "dock",
		Locale:        "en",
		MarketCountry: "US",
		Currency:      "USD",
	})
	require.NoError(t, err)

	_, err = quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{
		Selections: []QuickBuySelectionInput{
			{StepKey: "product-search", ProductID: outOfScope.ID, Quantity: 1},
		},
	})
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	assert.Contains(t, err.Error(), "allowed product category")
}

func TestQuickBuyServiceLocalizesFlowHelpTextWithBaseFallback(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Localized Rim", "localized_rim")

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:     "localized-build",
		Name:     "Localized Build",
		HelpText: "Base QUICK flow help",
		Translations: []QuickBuyFlowTranslationInput{
			{Locale: "zh-CN", HelpText: "中文 QUICK 说明"},
			{Locale: "fr", HelpText: "Aide QUICK française"},
		},
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{productSpecificationTemplate.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	chinese, err := quickBuyService.CurrentFlow("dock", "zh-CN")
	require.NoError(t, err)
	require.Len(t, chinese.Steps, 1)
	assert.Equal(t, "中文 QUICK 说明", chinese.HelpText)
	assert.NotContains(t, string(mustMarshalQuickBuyTestValue(t, chinese)), `"translations"`)

	french, err := quickBuyService.CurrentFlow("dock", "fr")
	require.NoError(t, err)
	require.Len(t, french.Steps, 1)
	assert.Equal(t, "Aide QUICK française", french.HelpText)

	german, err := quickBuyService.CurrentFlow("dock", "de")
	require.NoError(t, err)
	require.Len(t, german.Steps, 1)
	assert.Equal(t, "Base QUICK flow help", german.HelpText)
}

func mustMarshalQuickBuyTestValue(t *testing.T, value interface{}) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return encoded
}

func TestQuickBuyServiceRejectsPublishWithoutRequiredProductSpecificationTemplate(t *testing.T) {
	_, quickBuyService := newQuickBuyTestService(t)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "empty-build",
		Name:         "Empty Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{StepKey: "rim", Name: "Rims"},
			},
		},
	})
	require.NoError(t, err)

	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
}

func TestQuickBuyServiceValidationFlagsDisabledProductCategory(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productCategory := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	require.NoError(t, db.Model(&productdomain.ProductCategory{}).
		Where("id = ?", productCategory.ID).
		Update("is_enabled", false).Error)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "disabled-type-build",
		Name:         "Disabled Type Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{productCategory.ID},
				},
			},
		},
	})
	require.NoError(t, err)

	result, err := quickBuyService.ValidateVersion(created.Version.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Issues)
	assert.Equal(t, "disabled_product_category", result.Issues[0].Code)

	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
}

func TestQuickBuyServiceCreatesSessionAndStoresSelectionSnapshot(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	productRecord := seedQuickBuyProduct(t, db, productSpecificationTemplate.ID)
	require.NoError(t, db.Create(&productdomain.ProductMedia{
		ProductID:    productRecord.ID,
		MediaType:    "image",
		URL:          "/uploads/quick-buy-rim.webp",
		ThumbnailURL: "/uploads/quick-buy-rim-thumb.webp",
		IsPrimary:    true,
		IsVisible:    true,
	}).Error)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "session-build",
		Name:         "Session Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{productSpecificationTemplate.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	session, err := quickBuyService.CreateSession(QuickBuySessionInput{
		Surface:       "dock",
		Locale:        "en",
		MarketCountry: "US",
		Currency:      "USD",
	})
	require.NoError(t, err)
	require.NotEmpty(t, session.SessionToken)
	require.Equal(t, quickbuy.ValidationStatusInvalid, session.ValidationStatus)

	updated, err := quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{
		Selections: []QuickBuySelectionInput{
			{StepKey: "rim", ProductID: productRecord.ID, Quantity: 2},
		},
	})
	require.NoError(t, err)
	require.Len(t, updated.Items, 1)
	assert.Equal(t, productRecord.ID, updated.Items[0].ProductID)
	assert.Equal(t, 2, updated.Items[0].Quantity)
	assert.Equal(t, "180.00", updated.SubtotalSnapshot)
	assert.Equal(t, 820, updated.WeightSnapshotG)
	assert.Equal(t, quickbuy.ValidationStatusValid, updated.ValidationStatus)
	require.NotNil(t, updated.Validation)
	assert.True(t, updated.Validation.Valid)
	assert.NotEmpty(t, updated.Items[0].ProductSnapshot)
	assert.NotEmpty(t, updated.Items[0].VariantSnapshot)
	assert.Contains(t, string(updated.Items[0].ProductSnapshot), "/uploads/quick-buy-rim-thumb.webp")
}

func TestQuickBuyProductSnapshotCanonicalizesThumbnailURL(t *testing.T) {
	resolver := NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	snapshot := quickBuyProductSnapshot(productdomain.Product{
		ID:   91,
		Name: "Quick Media Product",
		Slug: "quick-media-product",
		Media: []productdomain.ProductMedia{
			{
				MediaType:    "image",
				URL:          "http://media.internal:8080/uploads/quick-buy/full.webp",
				ThumbnailURL: "http://media.internal:8080/uploads/quick-buy/thumb.webp",
				IsPrimary:    true,
				IsVisible:    true,
			},
		},
	}, resolver)

	assert.Contains(t, string(snapshot), `"thumbnail":"https://shop.example.test/uploads/quick-buy/thumb.webp"`)
	assert.NotContains(t, string(snapshot), "media.internal")
}

func TestQuickBuyProductSnapshotUsesSelectedVariantSummary(t *testing.T) {
	variant := productdomain.ProductVariant{
		ID:         12,
		SKU:        "QB-VARIANT-12",
		PriceMinor: 12500,
		Currency:   "USD",
		IsActive:   true,
		IsDefault:  true,
	}
	snapshot := quickBuyProductSnapshotForVariant(productdomain.Product{
		ID:       91,
		Name:     "Quick product",
		Slug:     "quick-product",
		Variants: []productdomain.ProductVariant{variant},
	}, &variant)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(snapshot, &payload))
	require.Equal(t, "QB-VARIANT-12", payload["sku"])
	require.Equal(t, "125.00", payload["price"])
	require.NotContains(t, payload, "sale_price")
	require.Equal(t, 12.0, payload["variant_id"])
}

func TestQuickBuyServiceAllowsClearingAndReselectingStep(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	firstProduct := seedQuickBuyProductWithDetails(t, db, productSpecificationTemplate.ID, "QB-RIM-001", "First Rim", "first-rim", 100)
	secondProduct := seedQuickBuyProductWithDetails(t, db, productSpecificationTemplate.ID, "QB-RIM-002", "Second Rim", "second-rim", 120)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "reselect-build",
		Name:         "Reselect Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Wheelset custom configuration",
					ProductCategoryIDs: []uint{productSpecificationTemplate.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	session, err := quickBuyService.CreateSession(QuickBuySessionInput{
		Surface:       "dock",
		Locale:        "en",
		MarketCountry: "US",
		Currency:      "USD",
	})
	require.NoError(t, err)

	selected, err := quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{
		Selections: []QuickBuySelectionInput{
			{StepKey: "rim", ProductID: firstProduct.ID, Quantity: 1},
		},
	})
	require.NoError(t, err)
	require.Len(t, selected.Items, 1)
	assert.Equal(t, firstProduct.ID, selected.Items[0].ProductID)

	cleared, err := quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{
		Selections: []QuickBuySelectionInput{
			{StepKey: "rim", ProductID: 0, Quantity: 0},
		},
	})
	require.NoError(t, err)
	assert.Empty(t, cleared.Items)

	reselected, err := quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{
		Selections: []QuickBuySelectionInput{
			{StepKey: "rim", ProductID: secondProduct.ID, Quantity: 1},
		},
	})
	require.NoError(t, err)
	require.Len(t, reselected.Items, 1)
	assert.Equal(t, secondProduct.ID, reselected.Items[0].ProductID)
}

func TestQuickBuyServiceListsSessionStepCandidatesByBoundProductSpecificationTemplate(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	rimType := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	handlebarType := seedQuickBuyCatalogCategory(t, db, "Handlebar", "handlebar")
	rimProduct := seedQuickBuyProduct(t, db, rimType.ID)
	_ = seedQuickBuyProductWithDetails(t, db, handlebarType.ID, "QB-HB-001", "Quick Buy Handlebar", "quick-buy-handlebar", 70)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "candidate-build",
		Name:         "Candidate Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{rimType.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	session, err := quickBuyService.CreateSession(QuickBuySessionInput{
		Surface:       "dock",
		Locale:        "en",
		MarketCountry: "US",
		Currency:      "USD",
	})
	require.NoError(t, err)

	result, err := quickBuyService.ListSessionStepCandidates(session.SessionToken, QuickBuyCandidateInput{
		StepKey:  "rim",
		Locale:   "en",
		PageSize: 12,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Products, 1)
	assert.Equal(t, rimProduct.ID, result.Products[0].ID)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "rim", result.Step.StepKey)
	assert.Equal(t, "rim", result.Step.ProductCategories[0].Slug)
	assert.False(t, result.HasMore)
}

func TestQuickBuyServiceRejectsRetiredTemplateFilterConfiguration(t *testing.T) {
	_, quickBuyService := newQuickBuyTestService(t)
	_, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "retired-template-filter",
		Name:         "Retired Template Filter",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{Steps: []QuickBuyStepInput{{
			StepKey: "product", Name: "Products", ProductSpecificationTemplateIDs: []uint{1},
		}}},
	})
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	require.ErrorContains(t, err, "select products by category")
}

func TestQuickBuyServiceFiltersDedicatedWheelsetsByCategoryWithoutTemplateBinding(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	wheelsetCategory := seedQuickBuyProductCategory(t, db, "Wheelsets", productdomain.ProductTypeWheelset, nil)
	productCategoryID := wheelsetCategory.ID
	wheelset := productdomain.Product{
		ProductType:       productdomain.ProductTypeWheelset,
		ProductCategoryID: &productCategoryID,
		WheelsetSpecification: &productdomain.WheelsetSpecification{
			SchemaVersion: 1,
			FrontWheel: productdomain.WheelsetProductWheelSpecification{
				WheelSize:       "29",
				SpokeHoleCounts: []int{24},
				LacingPatterns:  []string{"2-cross"},
				AxleStandard:    "15x110",
				BrakeInterface:  "center_lock",
			},
			RearWheel: productdomain.WheelsetProductWheelSpecification{
				WheelSize:       "27.5",
				SpokeHoleCounts: []int{24},
				LacingPatterns:  []string{"2-cross"},
				AxleStandard:    "12x148",
				BrakeInterface:  "center_lock",
				FreehubBody:     "micro_spline",
			},
		},
		Name:   "Mullet Wheelset",
		Slug:   "quick-buy-mullet-wheelset",
		Status: "active",
		Locale: "en",
	}
	require.NoError(t, db.Create(&wheelset).Error)
	require.NoError(t, db.Create(&productdomain.ProductVariant{
		ProductID:  wheelset.ID,
		SKU:        "QUICK-BUY-MULLET-WHEELSET-VARIANT",
		IsActive:   true,
		IsDefault:  true,
		Stock:      1,
		PriceMinor: 10000,
	}).Error)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wheelset-category-build",
		Name:         "Wheelset Category Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{{
				StepKey:            "wheelset",
				Name:               "Wheelsets",
				ProductCategoryIDs: []uint{wheelsetCategory.ID},
			}},
		},
	})
	require.NoError(t, err)

	candidates, err := quickBuyService.PreviewVersionStepCandidates(created.Version.ID, QuickBuyCandidateInput{
		StepKey: "wheelset",
		Locale:  "en",
	})
	require.NoError(t, err)
	require.Len(t, candidates.Products, 1)
	require.Len(t, candidates.Step.Filters, len(productdomain.WheelsetProductLevelSpecificationFilterDefinitions()))
	filterValuesBySlug := make(map[string][]string, len(candidates.Step.Filters))
	for _, filter := range candidates.Step.Filters {
		filterValuesBySlug[filter.Slug] = filter.Values
	}
	require.Equal(t, []string{"27.5"}, filterValuesBySlug[productdomain.WheelsetFilterKeyRearWheelSize])

	filtered, err := quickBuyService.PreviewVersionStepCandidates(created.Version.ID, QuickBuyCandidateInput{
		StepKey: "wheelset",
		Locale:  "en",
		SpecFilters: map[string][]string{
			productdomain.WheelsetFilterKeyRearWheelSize: {"27.5"},
		},
	})
	require.NoError(t, err)
	require.Len(t, filtered.Products, 1)
	require.Equal(t, wheelset.ID, filtered.Products[0].ID)
}

func TestQuickBuyServiceRejectsSessionSelectionFromWrongProductCategory(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	rimType := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	handlebarType := seedQuickBuyCatalogCategory(t, db, "Handlebar", "handlebar")
	wrongProduct := seedQuickBuyProductWithDetails(t, db, handlebarType.ID, "QB-HB-002", "Wrong Handlebar", "wrong-handlebar", 80)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wrong-type-build",
		Name:         "Wrong Type Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{rimType.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	session, err := quickBuyService.CreateSession(QuickBuySessionInput{
		Surface:       "dock",
		Locale:        "en",
		MarketCountry: "US",
		Currency:      "USD",
	})
	require.NoError(t, err)

	_, err = quickBuyService.UpdateSessionSelections(session.SessionToken, QuickBuySelectionUpdateInput{
		Selections: []QuickBuySelectionInput{
			{StepKey: "rim", ProductID: wrongProduct.ID, Quantity: 1},
		},
	})
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
	assert.Contains(t, err.Error(), "not in an allowed product category")
}

func TestQuickBuyServicePreviewsDraftVersionCandidates(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	productRecord := seedQuickBuyProduct(t, db, productSpecificationTemplate.ID)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "draft-preview-build",
		Name:         "Draft Preview Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{productSpecificationTemplate.ID},
				},
			},
		},
	})
	require.NoError(t, err)

	result, err := quickBuyService.PreviewVersionStepCandidates(created.Version.ID, QuickBuyCandidateInput{StepKey: "rim", Locale: "en"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Products, 1)
	assert.Equal(t, productRecord.ID, result.Products[0].ID)
	assert.Equal(t, quickbuy.FlowVersionStatusDraft, created.Version.Status)
}

func TestQuickBuyServiceDoesNotAllowExplicitInactiveVersion(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyCatalogCategory(t, db, "Rim", "rim")
	startsAt := time.Now().UTC().Add(time.Hour)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "future-build",
		Name:         "Future Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			StartsAt: &startsAt,
			Steps: []QuickBuyStepInput{
				{
					StepKey:            "rim",
					Name:               "Rims",
					ProductCategoryIDs: []uint{productSpecificationTemplate.ID},
				},
			},
		},
	})
	require.NoError(t, err)
	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.NoError(t, err)

	_, err = quickBuyService.CreateSession(QuickBuySessionInput{
		FlowVersionID: created.Version.ID,
		Surface:       "dock",
		Locale:        "en",
		MarketCountry: "US",
		Currency:      "USD",
	})
	require.ErrorIs(t, err, ErrQuickBuyNotFound)
}

func newQuickBuyTestService(t *testing.T) (*gorm.DB, *QuickBuyService) {
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
		&productdomain.ProductInformationTemplate{},
		&productdomain.ProductSpecificationTemplate{},
		&productdomain.ProductCategory{},
		&productdomain.ProductCategoryTranslation{},
		&productdomain.SpecDefinition{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{},
		&productdomain.ProductVariant{},
		&productdomain.ProductShippingTemplateBinding{},
		&productdomain.ProductVariantOptionValue{},
		&quickbuy.Flow{},
		&quickbuy.FlowTranslation{},
		&quickbuy.Version{},
		&quickbuy.Step{},
		&quickbuy.StepProductCategory{},
		&quickbuy.StepProductSpecificationTemplate{},
		&quickbuy.StepFilter{},
		&quickbuy.Rule{},
		&quickbuy.Session{},
		&quickbuy.SessionItem{},
	))

	productRepo := repository.NewProductRepository(db)
	return db, NewQuickBuyService(repository.NewQuickBuyRepository(db), productRepo, repository.NewProductCategoryRepository(db))
}

func seedQuickBuyCatalogCategory(t *testing.T, db *gorm.DB, name, slug string) productdomain.ProductCategory {
	t.Helper()
	return seedQuickBuyProductCategory(t, db, name, slug, nil)
}

func seedQuickBuyProductCategory(t *testing.T, db *gorm.DB, name, slug string, parentID *uint) productdomain.ProductCategory {
	t.Helper()
	depth := 1
	if parentID != nil {
		depth = 2
	}
	category := productdomain.ProductCategory{
		ParentID:  parentID,
		Name:      name,
		Slug:      slug,
		Depth:     depth,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&category).Error)
	return category
}

func seedQuickBuyProduct(t *testing.T, db *gorm.DB, productCategoryID uint) productdomain.Product {
	t.Helper()
	return seedQuickBuyProductWithDetails(t, db, productCategoryID, "QB-RIM-001", "Quick Buy Rim", "quick-buy-rim", 100)
}

func seedQuickBuyProductWithDetails(t *testing.T, db *gorm.DB, productCategoryID uint, sku, name, slug string, price float64, additionalCategoryIDs ...uint) productdomain.Product {
	t.Helper()
	if len(additionalCategoryIDs) > 0 && additionalCategoryIDs[0] > 0 {
		productCategoryID = additionalCategoryIDs[0]
	}
	productRecord := productdomain.Product{
		ProductCategoryID: &productCategoryID,
		SKU:               sku,
		Name:              name,
		Slug:              slug,
		Currency:          "USD",
		PriceMinor:        int64(price * 100),
		Stock:             5,
		Status:            "active",
		Locale:            "en",
	}
	require.NoError(t, db.Create(&productRecord).Error)
	variant := productdomain.ProductVariant{
		ProductID:  productRecord.ID,
		SKU:        sku + "-DEFAULT",
		Title:      "Default",
		Currency:   "USD",
		PriceMinor: int64((price - 10) * 100),
		Stock:      5,
		Weight:     410,
		IsDefault:  true,
		IsActive:   true,
		SortOrder:  10,
	}
	require.NoError(t, db.Create(&variant).Error)
	return productRecord
}

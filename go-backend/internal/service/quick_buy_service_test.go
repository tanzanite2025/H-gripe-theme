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

func TestQuickBuyServiceCreatesPublishesAndReturnsCurrentFlow(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wheelset-build",
		Name:         "Wheelset Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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
	require.Len(t, current.Steps[0].ProductSpecificationTemplates, 1)
	assert.Equal(t, "rim", current.Steps[0].ProductSpecificationTemplates[0].Slug)
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
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
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
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Localized Rim", "localized_rim")

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
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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

func TestQuickBuyServiceValidationFlagsDisabledProductSpecificationTemplate(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	require.NoError(t, db.Model(&productdomain.ProductSpecificationTemplate{}).
		Where("id = ?", productSpecificationTemplate.ID).
		Update("is_enabled", false).Error)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "disabled-type-build",
		Name:         "Disabled Type Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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
	assert.Equal(t, "disabled_product_specification_template", result.Issues[0].Code)

	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
}

func TestQuickBuyServiceCreatesSessionAndStoresSelectionSnapshot(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
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
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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
	assert.Equal(t, 180.0, updated.SubtotalSnapshot)
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

func TestQuickBuyServiceAllowsClearingAndReselectingStep(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	firstProduct := seedQuickBuyProductWithDetails(t, db, productSpecificationTemplate.ID, "QB-RIM-001", "First Rim", "first-rim", 100)
	secondProduct := seedQuickBuyProductWithDetails(t, db, productSpecificationTemplate.ID, "QB-RIM-002", "Second Rim", "second-rim", 120)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "reselect-build",
		Name:         "Reselect Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Wheelset custom configuration",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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
	rimType := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	handlebarType := seedQuickBuyProductSpecificationTemplate(t, db, "Handlebar", "handlebar")
	rimProduct := seedQuickBuyProduct(t, db, rimType.ID)
	_ = seedQuickBuyProductWithDetails(t, db, handlebarType.ID, "QB-HB-001", "Quick Buy Handlebar", "quick-buy-handlebar", 70)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "candidate-build",
		Name:         "Candidate Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{rimType.ID},
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
	assert.Equal(t, "rim", result.Step.ProductSpecificationTemplates[0].Slug)
	assert.False(t, result.HasMore)
}

func TestQuickBuyServiceUsesTemplateFilterableSpecificationsAndDynamicValues(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	rimType := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	specDefinition := productdomain.SpecDefinition{
		ProductSpecificationTemplateID: rimType.ID,
		Group:                          "规格",
		Name:                           "Rim Depth",
		Slug:                           "rim_depth",
		FieldType:                      "number",
		Unit:                           "mm",
		IsVisible:                      true,
		IsFilterable:                   true,
		SortOrder:                      10,
	}
	require.NoError(t, db.Create(&specDefinition).Error)

	first := seedQuickBuyProductWithDetails(t, db, rimType.ID, "QB-RIM-033", "Rim 33", "quick-buy-rim-33", 100)
	second := seedQuickBuyProductWithDetails(t, db, rimType.ID, "QB-RIM-045", "Rim 45", "quick-buy-rim-45", 110)
	require.NoError(t, db.Create(&productdomain.ProductSpecValue{
		ProductID:        first.ID,
		SpecDefinitionID: specDefinition.ID,
		Value:            "33",
	}).Error)
	require.NoError(t, db.Create(&productdomain.ProductSpecValue{
		ProductID:        second.ID,
		SpecDefinitionID: specDefinition.ID,
		Value:            "45",
	}).Error)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "dynamic-filter-build",
		Name:         "Dynamic Filter Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{rimType.ID},
				},
			},
		},
	})
	require.NoError(t, err)

	result, err := quickBuyService.PreviewVersionStepCandidates(created.Version.ID, QuickBuyCandidateInput{
		StepKey:  "rim",
		Locale:   "en",
		PageSize: 12,
	})
	require.NoError(t, err)
	require.Len(t, result.Products, 2)
	require.Len(t, result.Step.Filters, 1)
	assert.Equal(t, "rim_depth", result.Step.Filters[0].Slug)
	assert.Equal(t, []string{"33", "45"}, result.Step.Filters[0].Values)

	filtered, err := quickBuyService.PreviewVersionStepCandidates(created.Version.ID, QuickBuyCandidateInput{
		StepKey: "rim",
		Locale:  "en",
		SpecFilters: map[string][]string{
			"rim_depth": {"45"},
		},
		PageSize: 12,
	})
	require.NoError(t, err)
	require.Len(t, filtered.Products, 1)
	assert.Equal(t, second.ID, filtered.Products[0].ID)
	assert.NotEqual(t, first.ID, filtered.Products[0].ID)
}

func TestQuickBuyServiceRejectsSessionSelectionFromWrongProductSpecificationTemplate(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	rimType := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	handlebarType := seedQuickBuyProductSpecificationTemplate(t, db, "Handlebar", "handlebar")
	wrongProduct := seedQuickBuyProductWithDetails(t, db, handlebarType.ID, "QB-HB-002", "Wrong Handlebar", "wrong-handlebar", 80)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wrong-type-build",
		Name:         "Wrong Type Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{rimType.ID},
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
	assert.Contains(t, err.Error(), "not allowed")
}

func TestQuickBuyServicePreviewsDraftVersionCandidates(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	productRecord := seedQuickBuyProduct(t, db, productSpecificationTemplate.ID)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "draft-preview-build",
		Name:         "Draft Preview Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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
	productSpecificationTemplate := seedQuickBuyProductSpecificationTemplate(t, db, "Rim", "rim")
	startsAt := time.Now().UTC().Add(time.Hour)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "future-build",
		Name:         "Future Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			StartsAt: &startsAt,
			Steps: []QuickBuyStepInput{
				{
					StepKey:                         "rim",
					Name:                            "Rims",
					ProductSpecificationTemplateIDs: []uint{productSpecificationTemplate.ID},
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
		&productdomain.ProductSpecificationTemplateTranslation{},
		&productdomain.ProductCategory{},
		&productdomain.ProductCategoryTranslation{},
		&productdomain.SpecDefinition{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{},
		&productdomain.ProductVariant{},
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

func seedQuickBuyProductSpecificationTemplate(t *testing.T, db *gorm.DB, name, slug string) productdomain.ProductSpecificationTemplate {
	t.Helper()

	productSpecificationTemplate := productdomain.ProductSpecificationTemplate{
		Name:      name,
		Slug:      slug,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&productSpecificationTemplate).Error)
	return productSpecificationTemplate
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

func seedQuickBuyProduct(t *testing.T, db *gorm.DB, productSpecificationTemplateID uint) productdomain.Product {
	t.Helper()
	return seedQuickBuyProductWithDetails(t, db, productSpecificationTemplateID, "QB-RIM-001", "Quick Buy Rim", "quick-buy-rim", 100)
}

func seedQuickBuyProductWithDetails(t *testing.T, db *gorm.DB, productSpecificationTemplateID uint, sku, name, slug string, price float64, productCategoryIDs ...uint) productdomain.Product {
	t.Helper()
	var productCategoryID *uint
	if len(productCategoryIDs) > 0 && productCategoryIDs[0] > 0 {
		productCategoryID = &productCategoryIDs[0]
	}
	productRecord := productdomain.Product{
		ProductSpecificationTemplateID: &productSpecificationTemplateID,
		ProductCategoryID:              productCategoryID,
		SKU:                            sku,
		Name:                           name,
		Slug:                           slug,
		Currency:                       "USD",
		Price:                          price,
		Stock:                          5,
		Status:                         "active",
		Locale:                         "en",
	}
	require.NoError(t, db.Create(&productRecord).Error)
	variant := productdomain.ProductVariant{
		ProductID: productRecord.ID,
		SKU:       sku + "-DEFAULT",
		Title:     "Default",
		Currency:  "USD",
		Price:     price - 10,
		Stock:     5,
		Weight:    410,
		IsDefault: true,
		IsActive:  true,
		SortOrder: 10,
	}
	require.NoError(t, db.Create(&variant).Error)
	return productRecord
}

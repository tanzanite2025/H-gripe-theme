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
	productType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wheelset-build",
		Name:         "Wheelset Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{productType.ID},
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
	assert.Equal(t, "carbon_rim", current.Steps[0].Slug)
	require.Len(t, current.Steps[0].ProductTypes, 1)
	assert.Equal(t, "carbon_rim", current.Steps[0].ProductTypes[0].Slug)
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

func TestQuickBuyServicePublishesDefaultFlowWithoutConfiguredProductTypes(t *testing.T) {
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

func TestQuickBuyServiceLocalizesFlowHelpTextWithBaseFallback(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productType := seedQuickBuyProductType(t, db, "Localized Rim", "localized_rim")

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
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{productType.ID},
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

func TestQuickBuyServiceRejectsPublishWithoutRequiredProductType(t *testing.T) {
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

func TestQuickBuyServiceValidationFlagsDisabledProductType(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	require.NoError(t, db.Model(&productdomain.ProductType{}).
		Where("id = ?", productType.ID).
		Update("is_enabled", false).Error)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "disabled-type-build",
		Name:         "Disabled Type Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{productType.ID},
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
	assert.Equal(t, "disabled_product_type", result.Issues[0].Code)

	_, err = quickBuyService.PublishVersion(created.Version.ID, nil)
	require.ErrorIs(t, err, ErrQuickBuyInvalid)
}

func TestQuickBuyServiceCreatesSessionAndStoresSelectionSnapshot(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	productRecord := seedQuickBuyProduct(t, db, productType.ID)
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
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{productType.ID},
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

func TestQuickBuyServiceAllowsClearingAndReselectingStep(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	productType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	firstProduct := seedQuickBuyProductWithDetails(t, db, productType.ID, "QB-RIM-001", "First Rim", "first-rim", 100)
	secondProduct := seedQuickBuyProductWithDetails(t, db, productType.ID, "QB-RIM-002", "Second Rim", "second-rim", 120)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "reselect-build",
		Name:         "Reselect Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Wheelset custom configuration",
					ProductTypeIDs: []uint{productType.ID},
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

func TestQuickBuyServiceListsSessionStepCandidatesByBoundProductType(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	rimType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	handlebarType := seedQuickBuyProductType(t, db, "Handlebar", "handlebar")
	rimProduct := seedQuickBuyProduct(t, db, rimType.ID)
	_ = seedQuickBuyProductWithDetails(t, db, handlebarType.ID, "QB-HB-001", "Quick Buy Handlebar", "quick-buy-handlebar", 70)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "candidate-build",
		Name:         "Candidate Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{rimType.ID},
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
	assert.Equal(t, "carbon_rim", result.Step.ProductTypes[0].Slug)
	assert.False(t, result.HasMore)
}

func TestQuickBuyServiceRejectsSessionSelectionFromWrongProductType(t *testing.T) {
	db, quickBuyService := newQuickBuyTestService(t)
	rimType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	handlebarType := seedQuickBuyProductType(t, db, "Handlebar", "handlebar")
	wrongProduct := seedQuickBuyProductWithDetails(t, db, handlebarType.ID, "QB-HB-002", "Wrong Handlebar", "wrong-handlebar", 80)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "wrong-type-build",
		Name:         "Wrong Type Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{rimType.ID},
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
	productType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	productRecord := seedQuickBuyProduct(t, db, productType.ID)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "draft-preview-build",
		Name:         "Draft Preview Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{productType.ID},
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
	productType := seedQuickBuyProductType(t, db, "Carbon Rim", "carbon_rim")
	startsAt := time.Now().UTC().Add(time.Hour)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "future-build",
		Name:         "Future Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			StartsAt: &startsAt,
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					ProductTypeIDs: []uint{productType.ID},
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
		&productdomain.ProductType{},
		&productdomain.ProductTypeTranslation{},
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
		&quickbuy.StepProductType{},
		&quickbuy.StepFilter{},
		&quickbuy.Rule{},
		&quickbuy.Session{},
		&quickbuy.SessionItem{},
	))

	productRepo := repository.NewProductRepository(db)
	return db, NewQuickBuyService(repository.NewQuickBuyRepository(db), productRepo)
}

func seedQuickBuyProductType(t *testing.T, db *gorm.DB, name, slug string) productdomain.ProductType {
	t.Helper()

	productType := productdomain.ProductType{
		Name:      name,
		Slug:      slug,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&productType).Error)
	return productType
}

func seedQuickBuyProduct(t *testing.T, db *gorm.DB, productTypeID uint) productdomain.Product {
	t.Helper()
	return seedQuickBuyProductWithDetails(t, db, productTypeID, "QB-RIM-001", "Quick Buy Rim", "quick-buy-rim", 100)
}

func seedQuickBuyProductWithDetails(t *testing.T, db *gorm.DB, productTypeID uint, sku, name, slug string, price float64) productdomain.Product {
	t.Helper()
	productRecord := productdomain.Product{
		ProductTypeID: &productTypeID,
		SKU:           sku,
		Name:          name,
		Slug:          slug,
		Currency:      "USD",
		Price:         price,
		Stock:         5,
		Status:        "active",
		Locale:        "en",
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

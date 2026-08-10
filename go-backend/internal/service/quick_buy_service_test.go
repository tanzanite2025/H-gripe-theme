package service

import (
	"testing"
	"time"

	productdomain "tanzanite/internal/domain/product"
	"tanzanite/internal/domain/quickbuy"
	"tanzanite/internal/repository"

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
					SelectionMode:  quickbuy.SelectionModeSingle,
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

func TestQuickBuyServiceRejectsPublishWithoutRequiredProductType(t *testing.T) {
	_, quickBuyService := newQuickBuyTestService(t)

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "empty-build",
		Name:         "Empty Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{StepKey: "rim", Name: "Rims", SelectionMode: quickbuy.SelectionModeSingle},
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
					SelectionMode:  quickbuy.SelectionModeSingle,
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

	created, err := quickBuyService.CreateFlow(QuickBuyFlowInput{
		Slug:         "session-build",
		Name:         "Session Build",
		EntrySurface: "dock",
		Version: QuickBuyVersionInput{
			Steps: []QuickBuyStepInput{
				{
					StepKey:        "rim",
					Name:           "Rims",
					SelectionMode:  quickbuy.SelectionModeSingle,
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
					SelectionMode:  quickbuy.SelectionModeSingle,
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
					SelectionMode:  quickbuy.SelectionModeSingle,
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
					SelectionMode:  quickbuy.SelectionModeSingle,
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
					SelectionMode:  quickbuy.SelectionModeSingle,
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

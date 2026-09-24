package service

import (
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductQualityRequirementServiceUpsertsAndResolvesRules(t *testing.T) {
	db := newProductQualityRequirementServiceTestDB(t)
	service := NewProductQualityRequirementService(
		repository.NewProductQualityRequirementRepository(db),
	)
	productRecord, variantID := seedProductQualityRequirementServiceProduct(t, db)

	created, wasCreated, err := service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:              productRecord.ID,
		SpokeTensionQCRequired: true,
		Status:                 " ACTIVE ",
		RuleVersion:            " product-v1 ",
		Reason:                 " product default ",
		CreatedBy:              42,
	})
	require.NoError(t, err)
	require.True(t, wasCreated)
	require.NotNil(t, created)
	assert.Equal(t, productrequirement.RuleStatusActive, created.Status)
	assert.Equal(t, "product-v1", created.RuleVersion)
	assert.Equal(t, "product default", created.Reason)
	productRuleID := created.ID

	updated, wasCreated, err := service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:              productRecord.ID,
		SpokeTensionQCRequired: false,
		RuleVersion:            "product-v2",
		Reason:                 "disabled by default",
		CreatedBy:              99,
	})
	require.NoError(t, err)
	require.False(t, wasCreated)
	require.NotNil(t, updated)
	assert.Equal(t, productRuleID, updated.ID)
	assert.False(t, updated.SpokeTensionQCRequired)
	assert.Equal(t, "product-v2", updated.RuleVersion)
	assert.Equal(t, "disabled by default", updated.Reason)
	assert.Equal(t, uint(42), updated.CreatedBy)

	resolution, err := service.ResolveSpokeTensionQC(productRecord.ID, &variantID)
	require.NoError(t, err)
	assert.False(t, resolution.Required)
	assert.True(t, resolution.Matched)
	assert.Equal(t, productrequirement.ResolutionSourceProduct, resolution.Source)
	require.NotNil(t, resolution.RuleID)
	assert.Equal(t, productRuleID, *resolution.RuleID)
	assert.Equal(t, "product-v2", resolution.RuleVersion)

	variantRule, wasCreated, err := service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:              productRecord.ID,
		VariantID:              &variantID,
		SpokeTensionQCRequired: true,
		RuleVersion:            "variant-v1",
		Reason:                 "selected variant requires QC",
		CreatedBy:              42,
	})
	require.NoError(t, err)
	require.True(t, wasCreated)
	require.NotNil(t, variantRule)

	resolution, err = service.ResolveSpokeTensionQC(productRecord.ID, &variantID)
	require.NoError(t, err)
	assert.True(t, resolution.Required)
	assert.True(t, resolution.Matched)
	assert.Equal(t, productrequirement.ResolutionSourceVariant, resolution.Source)
	require.NotNil(t, resolution.RuleID)
	assert.Equal(t, variantRule.ID, *resolution.RuleID)
	assert.Equal(t, "variant-v1", resolution.RuleVersion)

	otherVariantID := uint(999999)
	resolution, err = service.ResolveSpokeTensionQC(productRecord.ID, &otherVariantID)
	require.ErrorIs(t, err, ErrProductQualityRequirementVariantNotFound)
}

func TestProductQualityRequirementServiceRejectsInvalidScopesAndInputs(t *testing.T) {
	db := newProductQualityRequirementServiceTestDB(t)
	service := NewProductQualityRequirementService(
		repository.NewProductQualityRequirementRepository(db),
	)
	productRecord, variantID := seedProductQualityRequirementServiceProduct(t, db)

	_, _, err := service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:   900000,
		RuleVersion: "product-v1",
	})
	require.ErrorIs(t, err, ErrProductQualityRequirementProductNotFound)

	_, otherVariantID := seedProductQualityRequirementServiceProductWithSKU(t, db, "SKU-QUALITY-OTHER")
	_, _, err = service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:   productRecord.ID,
		VariantID:   &otherVariantID,
		RuleVersion: "variant-v1",
	})
	require.ErrorIs(t, err, ErrProductQualityRequirementVariantNotFound)

	zeroVariantID := uint(0)
	_, _, err = service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:   productRecord.ID,
		VariantID:   &zeroVariantID,
		RuleVersion: "variant-v1",
	})
	require.ErrorIs(t, err, ErrProductQualityRequirementVariantNotFound)

	_, _, err = service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID: productRecord.ID,
		VariantID: &variantID,
		Status:    "deleted",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported product quality requirement status")
}

func TestProductQualityRequirementServiceReturnsNoRequirementWhenRuleIsInactive(t *testing.T) {
	db := newProductQualityRequirementServiceTestDB(t)
	service := NewProductQualityRequirementService(
		repository.NewProductQualityRequirementRepository(db),
	)
	productRecord, variantID := seedProductQualityRequirementServiceProduct(t, db)

	_, _, err := service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:              productRecord.ID,
		VariantID:              &variantID,
		SpokeTensionQCRequired: true,
		Status:                 productrequirement.RuleStatusInactive,
		RuleVersion:            "variant-v1",
		Reason:                 "temporarily disabled",
		CreatedBy:              42,
	})
	require.NoError(t, err)

	resolution, err := service.ResolveSpokeTensionQC(productRecord.ID, &variantID)
	require.NoError(t, err)
	assert.False(t, resolution.Required)
	assert.False(t, resolution.Matched)
	assert.Equal(t, productrequirement.ResolutionSourceNone, resolution.Source)
	assert.Equal(t, "no_active_rule", resolution.Reason)
}

func TestProductQualityRequirementServiceRequiresStore(t *testing.T) {
	service := NewProductQualityRequirementService(nil)

	_, _, err := service.UpsertSpokeTensionQC(ProductQualityRequirementInput{
		ProductID:   1,
		RuleVersion: "product-v1",
	})
	require.ErrorIs(t, err, ErrProductQualityRequirementStoreUnavailable)

	_, err = service.ResolveSpokeTensionQC(1, nil)
	require.ErrorIs(t, err, ErrProductQualityRequirementStoreUnavailable)
}

func newProductQualityRequirementServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&productdomain.Product{},
		&productdomain.ProductVariant{},
		&productrequirement.ProductQualityRequirementRule{},
	))
	return db
}

func seedProductQualityRequirementServiceProduct(t *testing.T, db *gorm.DB) (productdomain.Product, uint) {
	return seedProductQualityRequirementServiceProductWithSKU(t, db, "SKU-QUALITY-FIRST")
}

func seedProductQualityRequirementServiceProductWithSKU(
	t *testing.T,
	db *gorm.DB,
	sku string,
) (productdomain.Product, uint) {
	t.Helper()

	record := productdomain.Product{
		SKU:        sku,
		Name:       sku,
		Slug:       "slug-" + strings.ToLower(sku),
		Currency:   "USD",
		PriceMinor: 10000,
		Stock:      5,
	}
	require.NoError(t, db.Create(&record).Error)
	variantID := seedProductQualityRequirementServiceVariant(t, db, record.ID, sku+"-VARIANT")
	return record, variantID
}

func seedProductQualityRequirementServiceVariant(
	t *testing.T,
	db *gorm.DB,
	productID uint,
	sku string,
) uint {
	t.Helper()

	variant := productdomain.ProductVariant{
		ProductID:    productID,
		SKU:          sku,
		Title:        "Default",
		OptionValues: "{}",
		Currency:     "USD",
		PriceMinor:   10000,
		Stock:        5,
		Weight:       9000,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)
	return variant.ID
}

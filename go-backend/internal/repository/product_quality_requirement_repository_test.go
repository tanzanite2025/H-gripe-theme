package repository

import (
	productdomain "commerce-platform/internal/domain/product"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"strconv"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductQualityRequirementRepositorySeparatesProductAndVariantScopes(t *testing.T) {
	db := newProductQualityRequirementRepositoryTestDB(t)
	repo := NewProductQualityRequirementRepository(db)

	productRecord := seedProductQualityRequirementRepositoryProduct(t, db)
	variantID := seedProductQualityRequirementRepositoryVariant(t, db, productRecord.ID)

	productRule := &productrequirement.ProductQualityRequirementRule{
		ProductID:              productRecord.ID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: true,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "product-v1",
	}
	require.NoError(t, repo.Create(productRule))

	variantRule := &productrequirement.ProductQualityRequirementRule{
		ProductID:              productRecord.ID,
		VariantID:              &variantID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: false,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "variant-v1",
	}
	require.NoError(t, repo.Create(variantRule))

	foundProductRule, err := repo.FindByScope(
		productRecord.ID,
		nil,
		productrequirement.RequirementTypeSpokeTensionQC,
	)
	require.NoError(t, err)
	require.Equal(t, "product-v1", foundProductRule.RuleVersion)

	foundVariantRule, err := repo.FindByScope(
		productRecord.ID,
		&variantID,
		productrequirement.RequirementTypeSpokeTensionQC,
	)
	require.NoError(t, err)
	require.Equal(t, "variant-v1", foundVariantRule.RuleVersion)
}

func TestProductQualityRequirementRepositoryEnsureProductScopeRejectsForeignVariant(t *testing.T) {
	db := newProductQualityRequirementRepositoryTestDB(t)
	repo := NewProductQualityRequirementRepository(db)

	firstProduct := seedProductQualityRequirementRepositoryProduct(t, db)
	secondProduct := seedProductQualityRequirementRepositoryProductWithSKU(t, db, "SKU-REQUIREMENT-SECOND")
	variantID := seedProductQualityRequirementRepositoryVariant(t, db, secondProduct.ID)

	require.NoError(t, repo.EnsureProductScope(firstProduct.ID, nil))
	require.ErrorIs(t, repo.EnsureProductScope(firstProduct.ID, &variantID), gorm.ErrRecordNotFound)
}

func newProductQualityRequirementRepositoryTestDB(t *testing.T) *gorm.DB {
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

func seedProductQualityRequirementRepositoryProduct(t *testing.T, db *gorm.DB) productdomain.Product {
	return seedProductQualityRequirementRepositoryProductWithSKU(t, db, "SKU-REQUIREMENT-FIRST")
}

func seedProductQualityRequirementRepositoryProductWithSKU(t *testing.T, db *gorm.DB, sku string) productdomain.Product {
	t.Helper()

	record := productdomain.Product{
		SKU:      sku,
		Name:     sku,
		Slug:     strings.ToLower(sku),
		Currency: "USD",
		PriceMinor: 10000,
		Stock:    5,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedProductQualityRequirementRepositoryVariant(t *testing.T, db *gorm.DB, productID uint) uint {
	t.Helper()

	variant := productdomain.ProductVariant{
		ProductID:    productID,
		SKU:          "VARIANT-" + strconv.FormatUint(uint64(productID), 10),
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

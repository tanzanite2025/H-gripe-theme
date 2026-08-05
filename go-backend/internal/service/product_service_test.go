package service

import (
	"errors"
	"testing"

	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductServiceCreateAdminProductPersistsTemplateSpecs(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Carbon Rim 45",
		Slug:          "carbon-rim-45",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30.5",
			"tubeless_ready": "yes",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-001-24H-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, createdProduct)
	assert.Equal(t, productType.ID, *createdProduct.ProductTypeID)
	require.Len(t, createdProduct.SpecValues, 2)
	require.Len(t, createdProduct.Variants, 1)
	assert.Equal(t, "RIM-001-24H-DISC", createdProduct.SKU)
	assert.Equal(t, 5, createdProduct.Stock)
	assert.Equal(t, "true", findSavedSpecValue(t, createdProduct, "tubeless_ready"))
	assert.Equal(t, "30.5", findSavedSpecValue(t, createdProduct, "outer_width_mm"))
	assert.Equal(t, "RIM-001-24H-DISC", createdProduct.Variants[0].SKU)
	assert.JSONEq(t, `{"brake_type":"disc"}`, createdProduct.Variants[0].OptionValues)
}

func TestProductServiceCreateAdminProductRejectsInvalidTemplateSpec(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Invalid Carbon Rim",
		Slug:          "invalid-carbon-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-INVALID-CANTI",
				OptionValues: map[string]string{"brake_type": "cantilever"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrProductSpecInvalid) || errors.Is(err, ErrProductVariantInvalid))
	assert.Nil(t, createdProduct)

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	assert.Equal(t, int64(0), productCount)
}

func TestProductServiceUpdateAdminProductPreservesInactiveVariantWhenAnotherVariantIsActive(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-INACTIVE", "inactive-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.Len(t, createdProduct.Variants, 1)

	inactive := false
	variantID := createdProduct.Variants[0].ID
	updatedProduct, err := productService.UpdateAdminProduct(createdProduct.ID, ProductUpdateInput{
		Variants: []ProductVariantInput{
			{
				ID:           &variantID,
				SKU:          createdProduct.Variants[0].SKU,
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     &inactive,
			},
			{
				SKU:          "RIM-INACTIVE-ACTIVE-VAR",
				OptionValues: map[string]string{"brake_type": "rim"},
				Price:        419,
				Stock:        7,
				IsDefault:    false,
				IsActive:     boolPtr(true),
			},
		},
		UpdateVariants: true,
	})

	require.NoError(t, err)
	require.Len(t, updatedProduct.Variants, 2)
	assert.Equal(t, 7, updatedProduct.TotalVariantStock())

	var inactiveVariant, activeVariant *product.ProductVariant
	for i := range updatedProduct.Variants {
		switch updatedProduct.Variants[i].SKU {
		case createdProduct.Variants[0].SKU:
			inactiveVariant = &updatedProduct.Variants[i]
		case "RIM-INACTIVE-ACTIVE-VAR":
			activeVariant = &updatedProduct.Variants[i]
		}
	}
	require.NotNil(t, inactiveVariant)
	require.NotNil(t, activeVariant)
	assert.False(t, inactiveVariant.IsActive)
	assert.False(t, inactiveVariant.IsDefault)
	assert.True(t, activeVariant.IsActive)
	assert.True(t, activeVariant.IsDefault)
}

func TestProductServiceCreateAdminProductRejectsAllInactiveVariants(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	inactive := false

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Inactive SKU Rim",
		Slug:          "inactive-sku-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-ALL-INACTIVE-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        15,
				IsDefault:    true,
				IsActive:     &inactive,
			},
		},
	})

	require.ErrorIs(t, err, ErrProductVariantInvalid)
	assert.Nil(t, createdProduct)

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	assert.Equal(t, int64(0), productCount)
}

func TestProductServiceCreateAdminProductRequiresVariant(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "No Variant Rim",
		Slug:          "no-variant-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
	})

	require.ErrorIs(t, err, ErrProductVariantInvalid)
	assert.Nil(t, createdProduct)

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	assert.Equal(t, int64(0), productCount)
}

func TestProductServiceSearchPublicFiltersByTemplateSpec(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	discRim := createProductWithSpecs(t, productService, productType.ID, "RIM-DISC", "disc-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	createProductWithSpecs(t, productService, productType.ID, "RIM-RIM", "rim-brake-rim", map[string]string{
		"outer_width_mm": "25",
	}, map[string]string{"brake_type": "rim"})

	results, total, err := productService.SearchPublic(ProductSearchInput{
		Locale: "en",
		Status: "active",
		SpecFilters: map[string][]string{
			"brake_type": {"disc"},
		},
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, results, 1)
	assert.Equal(t, discRim.ID, results[0].ID)
}

func TestProductServicePublicAccessOnlyReturnsActiveProducts(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	activeProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-PUBLIC-ACTIVE", "public-active", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	inactiveProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-PUBLIC-INACTIVE", "public-inactive", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", inactiveProduct.ID).Update("status", "inactive").Error)

	_, err := productService.GetPublicByID(inactiveProduct.ID)
	require.ErrorIs(t, err, ErrProductNotFound)

	products, total, err := productService.ListPublic("en", false, 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, products, 1)
	assert.Equal(t, activeProduct.ID, products[0].ID)

	results, total, err := productService.SearchPublic(ProductSearchInput{
		Locale:   "en",
		Status:   "inactive",
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	assert.Equal(t, activeProduct.ID, results[0].ID)
}

func TestProductRepositoryPurchasableVariantRejectsInactiveProduct(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	inactiveProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-NOT-BUYABLE", "not-buyable", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", inactiveProduct.ID).Update("status", "inactive").Error)

	_, _, err := repository.NewProductRepository(db).FindPurchasableVariant(inactiveProduct.ID, nil)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductServicePublicCatalogIgnoresProductLocale(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Unified Rim",
		Slug:          "global-rim",
		Status:        "active",
		Locale:        "zh_cn",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-GLOBAL-VAR",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createdProduct)

	publicProduct, err := productService.GetPublicBySlug("global-rim", "en")
	require.NoError(t, err)
	assert.Equal(t, createdProduct.ID, publicProduct.ID)

	listed, listTotal, err := productService.ListPublic("en", false, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), listTotal)
	require.Len(t, listed, 1)
	assert.Equal(t, createdProduct.ID, listed[0].ID)

	searched, searchTotal, err := productService.SearchPublic(ProductSearchInput{
		Locale:   "en",
		Status:   "active",
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), searchTotal)
	require.Len(t, searched, 1)
	assert.Equal(t, createdProduct.ID, searched[0].ID)

	available, availableTotal, err := productService.ListPublicAvailable("en", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), availableTotal)
	require.Len(t, available, 1)
	assert.Equal(t, createdProduct.ID, available[0].ID)
}

func newTestProductService(t *testing.T) (*gorm.DB, *ProductService) {
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
		&product.ProductType{},
		&product.SpecDefinition{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
	))

	return db, NewProductService(repository.NewProductRepository(db), nil, 0)
}

func seedCarbonRimType(t *testing.T, db *gorm.DB) product.ProductType {
	t.Helper()

	productType := product.ProductType{
		Name:      "Carbon Rim",
		Slug:      "carbon_rim",
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&productType).Error)

	specDefinitions := []product.SpecDefinition{
		{
			ProductTypeID: productType.ID,
			Group:         "Dimensions",
			Name:          "Outer Width",
			Slug:          "outer_width_mm",
			FieldType:     "number",
			Unit:          "mm",
			IsRequired:    true,
			IsFilterable:  true,
			IsVisible:     true,
			SortOrder:     10,
		},
		{
			ProductTypeID:   productType.ID,
			Group:           "Compatibility",
			Name:            "Brake Type",
			Slug:            "brake_type",
			FieldType:       "select",
			IsRequired:      true,
			IsFilterable:    true,
			IsVisible:       true,
			IsVariantOption: true,
			SortOrder:       20,
			Options:         `["disc","rim"]`,
		},
		{
			ProductTypeID: productType.ID,
			Group:         "Compatibility",
			Name:          "Tubeless Ready",
			Slug:          "tubeless_ready",
			FieldType:     "boolean",
			IsFilterable:  true,
			IsVisible:     true,
			SortOrder:     30,
		},
	}
	require.NoError(t, db.Create(&specDefinitions).Error)

	return productType
}

func createProductWithSpecs(t *testing.T, productService *ProductService, productTypeID uint, sku, slug string, specs map[string]string, variantOptions map[string]string) *product.Product {
	t.Helper()

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productTypeID,
		Name:          sku,
		Slug:          slug,
		Status:        "active",
		Locale:        "en",
		SpecValues:    specs,
		Variants: []ProductVariantInput{
			{
				SKU:          sku + "-VAR",
				OptionValues: variantOptions,
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)
	return createdProduct
}

func findSavedSpecValue(t *testing.T, productRecord *product.Product, slug string) string {
	t.Helper()

	for _, specValue := range productRecord.SpecValues {
		if specValue.SpecDefinition != nil && specValue.SpecDefinition.Slug == slug {
			return specValue.Value
		}
	}

	t.Fatalf("spec value %q not found", slug)
	return ""
}

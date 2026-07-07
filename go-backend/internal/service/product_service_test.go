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
		SKU:           "RIM-001",
		Name:          "Carbon Rim 45",
		Slug:          "carbon-rim-45",
		Price:         399,
		Stock:         5,
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30.5",
			"brake_type":     "disc",
			"tubeless_ready": "yes",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, createdProduct)
	assert.Equal(t, productType.ID, *createdProduct.ProductTypeID)
	require.Len(t, createdProduct.SpecValues, 3)
	assert.Equal(t, "true", findSavedSpecValue(t, createdProduct, "tubeless_ready"))
	assert.Equal(t, "30.5", findSavedSpecValue(t, createdProduct, "outer_width_mm"))
}

func TestProductServiceCreateAdminProductRejectsInvalidTemplateSpec(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		SKU:           "RIM-INVALID",
		Name:          "Invalid Carbon Rim",
		Slug:          "invalid-carbon-rim",
		Price:         399,
		Stock:         5,
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
			"brake_type":     "cantilever",
		},
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrProductSpecInvalid))
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
		"brake_type":     "disc",
	})
	createProductWithSpecs(t, productService, productType.ID, "RIM-RIM", "rim-brake-rim", map[string]string{
		"outer_width_mm": "25",
		"brake_type":     "rim",
	})

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
		&product.ProductImage{},
		&product.ProductSpecValue{},
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
			ProductTypeID: productType.ID,
			Group:         "Compatibility",
			Name:          "Brake Type",
			Slug:          "brake_type",
			FieldType:     "select",
			IsRequired:    true,
			IsFilterable:  true,
			IsVisible:     true,
			SortOrder:     20,
			Options:       `["disc","rim"]`,
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

func createProductWithSpecs(t *testing.T, productService *ProductService, productTypeID uint, sku, slug string, specs map[string]string) *product.Product {
	t.Helper()

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productTypeID,
		SKU:           sku,
		Name:          sku,
		Slug:          slug,
		Price:         399,
		Stock:         5,
		Status:        "active",
		Locale:        "en",
		SpecValues:    specs,
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

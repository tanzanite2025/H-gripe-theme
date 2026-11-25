package service

import (
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createBicycleTireProductServiceTestInput(t *testing.T, db *gorm.DB) ProductCreateInput {
	t.Helper()
	wheelComponentsCategory := seedProductCategoryForProductServiceTest(t, db, "Wheel Components", SystemProductCategoryWheelComponentsSlug, nil)
	tireCategory := seedProductCategoryForProductServiceTest(t, db, "Tires", product.BicycleTireProductCategorySlug, &wheelComponentsCategory.ID)
	productType := product.ProductTypeBicycleTire
	return ProductCreateInput{
		MerchantProductIdentifier:     "TIRE-MERCHANT-001",
		ManufacturerProductIdentifier: "TIRE-MANUFACTURER-001",
		ProductType:                   &productType,
		ProductCategoryID:             &tireCategory.ID,
		Name:                          "Road Tire 40-622",
		Slug:                          "road-tire-40-622",
		Locale:                        "en",
		Status:                        "active",
		Currency:                      "USD",
		BicycleTireSpecification:      &product.BicycleTireSpecification{SchemaVersion: 1},
		Variants: []ProductVariantInput{{
			SKU:        "ROAD-TIRE-40-622",
			PriceMinor: 4500,
			Stock:      3,
			IsDefault:  true,
		}},
	}
}

func TestDedicatedBicycleTireCreationAndSpecificationUpdateRoundTrip(t *testing.T) {
	db, productService := newTestProductService(t)
	input := createBicycleTireProductServiceTestInput(t, db)

	created, err := productService.CreateAdminProduct(input)
	require.NoError(t, err)
	require.Equal(t, product.ProductTypeBicycleTire, created.ProductType)
	require.Nil(t, created.ProductSpecificationTemplateID)
	require.Equal(t, input.ProductCategoryID, created.ProductCategoryID)
	require.Equal(t, &product.BicycleTireSpecification{SchemaVersion: 1}, created.BicycleTireSpecification)
	require.Len(t, created.Variants, 1)
	require.Equal(t, "ROAD-TIRE-40-622", created.Variants[0].SKU)

	updatedSpecification := *created.BicycleTireSpecification
	updatedSpecification.ETRTO = "40-622"
	maxPressureBar := 5.0
	updatedSpecification.MaxPressureBar = &maxPressureBar
	updated, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateBicycleTireSpecification: true,
		BicycleTireSpecification:       &updatedSpecification,
	})
	require.NoError(t, err)
	require.Equal(t, "40-622", updated.BicycleTireSpecification.ETRTO)
	require.Equal(t, 5.0, *updated.BicycleTireSpecification.MaxPressureBar)

	updatedSpecification.ETRTO = ""
	updated, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateBicycleTireSpecification: true,
		BicycleTireSpecification:       &updatedSpecification,
	})
	require.NoError(t, err)
	require.Empty(t, updated.BicycleTireSpecification.ETRTO)
	require.Equal(t, 5.0, *updated.BicycleTireSpecification.MaxPressureBar)
}

func TestDedicatedBicycleTireCreationRejectsTemplateGenericOptionsAndCarbonRimSpecifications(t *testing.T) {
	for _, scenario := range []string{"template", "generic-variant-option", "article-number-without-brand", "carbon-rim-product-specification", "carbon-rim-variant-specification"} {
		t.Run(scenario, func(t *testing.T) {
			db, productService := newTestProductService(t)
			input := createBicycleTireProductServiceTestInput(t, db)
			switch scenario {
			case "template":
				templateID := uint(123)
				input.ProductSpecificationTemplateID = &templateID
			case "generic-variant-option":
				input.Variants[0].OptionValues = map[string]string{"size": "40-622"}
			case "article-number-without-brand":
				input.BicycleTireSpecification.ArticleNo = "TIRE-ARTICLE-001"
			case "carbon-rim-product-specification":
				input.CarbonRimSpecification = &product.CarbonRimSpecification{SchemaVersion: 1}
			case "carbon-rim-variant-specification":
				input.Variants[0].CarbonRimSpecification = &product.CarbonRimVariantSpecification{}
			}

			_, err := productService.CreateAdminProduct(input)
			switch scenario {
			case "article-number-without-brand":
				require.ErrorIs(t, err, ErrProductBrandInvalid)
			case "carbon-rim-variant-specification", "generic-variant-option":
				require.ErrorIs(t, err, ErrProductVariantInvalid)
			default:
				require.ErrorIs(t, err, ErrProductSpecInvalid)
			}
		})
	}
}

func TestDedicatedBicycleTireUpdateRejectsProductSpecificationTemplateBinding(t *testing.T) {
	db, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(createBicycleTireProductServiceTestInput(t, db))
	require.NoError(t, err)

	templateID := uint(123)
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateProductSpecificationTemplateID: true,
		ProductSpecificationTemplateID:       &templateID,
	})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "not a template")
}

func TestDedicatedBicycleTireProductTypeAndTranslationParentCannotChangeThroughAdminUpdate(t *testing.T) {
	db, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(createBicycleTireProductServiceTestInput(t, db))
	require.NoError(t, err)

	carbonRimType := product.ProductTypeCarbonRim
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{ProductType: &carbonRimType})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "product_type is immutable")

	parentID := created.ID
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateParentID: true, ParentID: &parentID})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "manual translation parent")
}

func TestDedicatedBicycleTireCreationRejectsDuplicateBrandAndArticleNumber(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&product.ProductBrand{}))
	brand := product.ProductBrand{Name: "Schwalbe", Slug: "schwalbe", IsEnabled: true}
	require.NoError(t, db.Create(&brand).Error)
	productService.ConfigureProductBrandRepository(repository.NewProductBrandRepository(db))

	firstInput := createBicycleTireProductServiceTestInput(t, db)
	firstInput.BrandID = &brand.ID
	firstInput.BicycleTireSpecification.ArticleNo = "11100568"
	_, err := productService.CreateAdminProduct(firstInput)
	require.NoError(t, err)

	duplicateInput := firstInput
	duplicateInput.Name = "Second Road Tire"
	duplicateInput.Slug = "second-road-tire"
	duplicateInput.MerchantProductIdentifier = "TIRE-MERCHANT-002"
	duplicateInput.ManufacturerProductIdentifier = "TIRE-MANUFACTURER-002"
	duplicateInput.Variants = append([]ProductVariantInput(nil), firstInput.Variants...)
	duplicateInput.Variants[0].SKU = "SECOND-ROAD-TIRE"
	duplicateSpecification := *firstInput.BicycleTireSpecification
	duplicateSpecification.ArticleNo = " 11100568 "
	duplicateInput.BicycleTireSpecification = &duplicateSpecification
	_, err = productService.CreateAdminProduct(duplicateInput)
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "already belongs to another tire product")
}

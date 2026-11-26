package service

import (
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createWheelsetProductServiceTestInput(t *testing.T, db *gorm.DB) ProductCreateInput {
	t.Helper()
	wheelsetCategory := seedProductCategoryForProductServiceTest(t, db, "Wheelsets", SystemProductCategoryWheelsetSlug, nil)
	return wheelsetProductServiceTestInput(wheelsetCategory)
}

func wheelsetProductServiceTestInput(wheelsetCategory product.ProductCategory) ProductCreateInput {
	productType := product.ProductTypeWheelset
	return ProductCreateInput{
		MerchantProductIdentifier:     "WHEELSET-MERCHANT-001",
		ManufacturerProductIdentifier: "WHEELSET-MANUFACTURER-001",
		ProductType:                   &productType,
		ProductCategoryID:             &wheelsetCategory.ID,
		WheelsetSpecification: &product.WheelsetSpecification{
			SchemaVersion: 1,
			FrontWheel: product.WheelsetProductWheelSpecification{
				WheelSize: "700C", SpokeHoleCounts: []int{24, 28}, LacingPatterns: []string{"2x", "3x"},
			},
			RearWheel: product.WheelsetProductWheelSpecification{
				WheelSize: "650B", SpokeHoleCounts: []int{28, 32}, LacingPatterns: []string{"3x"},
			},
		},
		Name:     "Mixed Diameter Wheelset",
		Slug:     "mixed-diameter-wheelset",
		Locale:   "en",
		Status:   "active",
		Currency: "USD",
		Variants: []ProductVariantInput{{
			SKU: "MIXED-DIAMETER-WHEELSET", PriceMinor: 25000, Stock: 2, IsDefault: true,
		}},
	}
}

func TestDedicatedWheelsetCreationAndSpecificationUpdateRoundTrip(t *testing.T) {
	db, productService := newTestProductService(t)
	input := createWheelsetProductServiceTestInput(t, db)
	input.StandardProductIdentifier = " 00012345678905 "
	input.WheelsetSpecification.FrontWheel.WheelSize = " 700C "
	input.WheelsetSpecification.FrontWheel.LacingPatterns[0] = " 2x "
	input.WheelsetSpecification.RearWheel.Material = " Carbon fiber "
	input.WheelsetSpecification.Certifications = []string{" UCI approved "}

	created, err := productService.CreateAdminProduct(input)
	require.NoError(t, err)
	require.Equal(t, product.ProductTypeWheelset, created.ProductType)
	require.Equal(t, "00012345678905", created.StandardProductIdentifier)
	require.Nil(t, created.ProductSpecificationTemplateID)
	require.Equal(t, input.ProductCategoryID, created.ProductCategoryID)
	require.Equal(t, input.WheelsetSpecification, created.WheelsetSpecification)
	require.Equal(t, "700C", created.WheelsetSpecification.FrontWheel.WheelSize)
	require.Equal(t, "650B", created.WheelsetSpecification.RearWheel.WheelSize)
	require.Equal(t, []string{"2x", "3x"}, created.WheelsetSpecification.FrontWheel.LacingPatterns)
	require.Equal(t, "Carbon fiber", created.WheelsetSpecification.RearWheel.Material)
	require.Equal(t, []string{"UCI approved"}, created.WheelsetSpecification.Certifications)

	updatedSpecification := *created.WheelsetSpecification
	updatedSpecification.RearWheel.RimDepthMM = wheelsetFloat64Pointer(45)
	updatedSpecification.MaxSystemWeightKG = wheelsetFloat64Pointer(120)
	updated, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateWheelsetSpecification: true,
		WheelsetSpecification:       &updatedSpecification,
	})
	require.NoError(t, err)
	require.Equal(t, 45.0, *updated.WheelsetSpecification.RearWheel.RimDepthMM)
	require.Equal(t, 120.0, *updated.WheelsetSpecification.MaxSystemWeightKG)
	require.Equal(t, "700C", updated.WheelsetSpecification.FrontWheel.WheelSize)
	require.Equal(t, "650B", updated.WheelsetSpecification.RearWheel.WheelSize)
}

func TestDedicatedWheelsetGTINRejectsInvalidCheckDigitWithoutPartialCreateOrUpdate(t *testing.T) {
	db, productService := newTestProductService(t)
	input := createWheelsetProductServiceTestInput(t, db)
	input.StandardProductIdentifier = "4006381333932"
	_, err := productService.CreateAdminProduct(input)
	require.ErrorIs(t, err, ErrProductCustomsInfoInvalid)
	require.ErrorContains(t, err, "correct check digit")
	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	require.Zero(t, productCount)

	input.StandardProductIdentifier = "4006381333931"
	created, err := productService.CreateAdminProduct(input)
	require.NoError(t, err)

	invalidGTIN := "4006381333932"
	updatedName := "Must not persist with invalid GTIN"
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		Name:                            &updatedName,
		StandardProductIdentifier:       &invalidGTIN,
		UpdateStandardProductIdentifier: true,
	})
	require.ErrorIs(t, err, ErrProductCustomsInfoInvalid)

	reloaded, err := productService.GetAdminProduct(created.ID)
	require.NoError(t, err)
	require.Equal(t, created.Name, reloaded.Name)
	require.Equal(t, "4006381333931", reloaded.StandardProductIdentifier)
}

func TestDedicatedWheelsetCreationRejectsMissingRequiredFactsAndCrossTypeFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *gorm.DB, *ProductCreateInput)
	}{
		{"missing front size", func(_ *testing.T, _ *gorm.DB, input *ProductCreateInput) {
			input.WheelsetSpecification.FrontWheel.WheelSize = ""
		}},
		{"missing rear hole counts", func(_ *testing.T, _ *gorm.DB, input *ProductCreateInput) {
			input.WheelsetSpecification.RearWheel.SpokeHoleCounts = nil
		}},
		{"missing front lacing patterns", func(_ *testing.T, _ *gorm.DB, input *ProductCreateInput) {
			input.WheelsetSpecification.FrontWheel.LacingPatterns = nil
		}},
		{"generic template", func(_ *testing.T, _ *gorm.DB, input *ProductCreateInput) {
			templateID := uint(123)
			input.ProductSpecificationTemplateID = &templateID
		}},
		{"carbon rim fields", func(_ *testing.T, _ *gorm.DB, input *ProductCreateInput) {
			input.CarbonRimSpecification = &product.CarbonRimSpecification{SchemaVersion: 1}
		}},
		{"wrong category", func(t *testing.T, db *gorm.DB, input *ProductCreateInput) {
			wrongCategory := seedProductCategoryForProductServiceTest(t, db, "Other", "other", nil)
			input.ProductCategoryID = &wrongCategory.ID
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, productService := newTestProductService(t)
			input := createWheelsetProductServiceTestInput(t, db)
			test.mutate(t, db, &input)
			_, err := productService.CreateAdminProduct(input)
			if test.name == "wrong category" {
				require.ErrorIs(t, err, ErrProductCategoryInvalid)
			} else {
				require.ErrorIs(t, err, ErrProductSpecInvalid)
			}
		})
	}
}

func TestDedicatedWheelsetCategoryMayRemainNestedAndCannotChangeDuringUpdate(t *testing.T) {
	db, productService := newTestProductService(t)
	rootCategory := seedProductCategoryForProductServiceTest(t, db, "Root", "root", nil)
	wheelsetCategory := seedProductCategoryForProductServiceTest(t, db, "Wheelsets", SystemProductCategoryWheelsetSlug, &rootCategory.ID)
	input := wheelsetProductServiceTestInput(wheelsetCategory)
	created, err := productService.CreateAdminProduct(input)
	require.NoError(t, err)

	otherCategory := seedProductCategoryForProductServiceTest(t, db, "Other", "other", nil)
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateProductCategoryID: true,
		ProductCategoryID:       &otherCategory.ID,
	})
	require.ErrorIs(t, err, ErrProductCategoryInvalid)
}

func TestDedicatedWheelsetTypeAndSpecificationAreExclusive(t *testing.T) {
	productService := &ProductService{}
	ordinaryProductType := ""
	_, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:           &ordinaryProductType,
		WheelsetSpecification: &product.WheelsetSpecification{SchemaVersion: 1},
	})
	require.ErrorIs(t, err, ErrProductSpecInvalid)

	db, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(createWheelsetProductServiceTestInput(t, db))
	require.NoError(t, err)
	carbonRimType := product.ProductTypeCarbonRim
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{ProductType: &carbonRimType})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "product_type is immutable")
}

func TestGenericProductCannotSelectWheelsetSystemCategoryAsItsType(t *testing.T) {
	db, productService := newTestProductService(t)
	wheelsetCategory := seedProductCategoryForProductServiceTest(t, db, "Wheelsets", SystemProductCategoryWheelsetSlug, nil)
	ordinaryProductType := ""
	_, err := productService.CreateAdminProduct(ProductCreateInput{
		MerchantProductIdentifier:     "GENERIC-MERCHANT-001",
		ManufacturerProductIdentifier: "GENERIC-MANUFACTURER-001",
		ProductType:                   &ordinaryProductType,
		ProductCategoryID:             &wheelsetCategory.ID,
		Name:                          "Generic Product",
		Slug:                          "generic-product-in-wheelset-category",
		Locale:                        "en",
		Status:                        "active",
		Currency:                      "USD",
	})
	require.ErrorIs(t, err, ErrProductCategoryInvalid)
	require.ErrorContains(t, err, "reserved for wheelset products")
}

func TestDedicatedWheelsetUpdateRequiresSpecificationMarkerAndCannotClearRequiredFacts(t *testing.T) {
	db, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(createWheelsetProductServiceTestInput(t, db))
	require.NoError(t, err)

	specification := *created.WheelsetSpecification
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{WheelsetSpecification: &specification})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "update marker")

	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateWheelsetSpecification: true})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
}

func wheelsetFloat64Pointer(value float64) *float64 {
	return &value
}

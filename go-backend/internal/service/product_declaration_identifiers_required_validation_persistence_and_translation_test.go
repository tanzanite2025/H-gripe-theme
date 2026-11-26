package service

import (
	"strings"
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func createProductInputWithRequiredDeclarationIdentifiers() ProductCreateInput {
	return ProductCreateInput{
		Name: "Declaration product", Slug: "declaration-product", Locale: "en", Status: "active",
		MerchantProductIdentifier: "  SHOP-001  ", ManufacturerProductIdentifier: "  FACTORY-002  ",
		StandardProductIdentifier: "  00012345678905  ",
		Variants:                  []ProductVariantInput{{SKU: "SEPARATE-SKU", PriceMinor: 1500, Stock: 2, IsDefault: true}},
	}
}

func TestProductDeclarationIdentifiersRejectInvalidCreationWithoutPersistingProduct(t *testing.T) {
	for _, invalid := range []struct {
		name   string
		change func(*ProductCreateInput)
		label  string
	}{
		{"missing merchant", func(input *ProductCreateInput) { input.MerchantProductIdentifier = "" }, "MPID"},
		{"blank manufacturer", func(input *ProductCreateInput) { input.ManufacturerProductIdentifier = " \t\n " }, "NSPID"},
		{"long merchant", func(input *ProductCreateInput) { input.MerchantProductIdentifier = strings.Repeat("商", 71) }, "MPID"},
		{"long manufacturer", func(input *ProductCreateInput) { input.ManufacturerProductIdentifier = strings.Repeat("M", 71) }, "NSPID"},
		{"long optional standard", func(input *ProductCreateInput) { input.StandardProductIdentifier = strings.Repeat("0", 71) }, "SPID"},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			db, svc := newTestProductService(t)
			input := createProductInputWithRequiredDeclarationIdentifiers()
			invalid.change(&input)
			_, err := svc.CreateAdminProduct(input)
			require.ErrorIs(t, err, ErrProductCustomsInfoInvalid)
			require.ErrorContains(t, err, invalid.label)
			var count int64
			require.NoError(t, db.Model(&product.Product{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestProductDeclarationIdentifiersPersistUpdatePreserveOmittedAndCopyTranslation(t *testing.T) {
	_, svc := newTestProductService(t)
	created, err := svc.CreateAdminProduct(createProductInputWithRequiredDeclarationIdentifiers())
	require.NoError(t, err)
	require.Equal(t, "SHOP-001", created.MerchantProductIdentifier)
	require.Equal(t, "FACTORY-002", created.ManufacturerProductIdentifier)
	require.Equal(t, "00012345678905", created.StandardProductIdentifier)

	merchant := strings.Repeat("商", 70)
	manufacturer := " FACTORY-NEW "
	updated, err := svc.UpdateAdminProduct(created.ID, ProductUpdateInput{
		MerchantProductIdentifier: &merchant, UpdateMerchantProductIdentifier: true,
		ManufacturerProductIdentifier: &manufacturer, UpdateManufacturerProductIdentifier: true,
		UpdateStandardProductIdentifier: true, // Explicit null clears the optional value.
	})
	require.NoError(t, err)
	require.Equal(t, merchant, updated.MerchantProductIdentifier)
	require.Equal(t, "FACTORY-NEW", updated.ManufacturerProductIdentifier)
	require.Empty(t, updated.StandardProductIdentifier)

	name := "Updated name"
	updated, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &name})
	require.NoError(t, err)
	require.Equal(t, merchant, updated.MerchantProductIdentifier)
	require.Equal(t, "FACTORY-NEW", updated.ManufacturerProductIdentifier)

	translated, _, err := svc.CopyAdminProductTranslation(created.ID, "fr")
	require.NoError(t, err)
	require.Equal(t, merchant, translated.MerchantProductIdentifier)
	require.Equal(t, "FACTORY-NEW", translated.ManufacturerProductIdentifier)
	require.Empty(t, translated.StandardProductIdentifier)
}

func TestProductDeclarationIdentifiersRejectClearingRequiredValuesWithoutPartialUpdate(t *testing.T) {
	_, svc := newTestProductService(t)
	created, err := svc.CreateAdminProduct(createProductInputWithRequiredDeclarationIdentifiers())
	require.NoError(t, err)
	blank := "  "
	changedName := "Must not persist"
	for _, update := range []ProductUpdateInput{
		{Name: &changedName, UpdateMerchantProductIdentifier: true},
		{Name: &changedName, ManufacturerProductIdentifier: &blank, UpdateManufacturerProductIdentifier: true},
	} {
		_, err = svc.UpdateAdminProduct(created.ID, update)
		require.ErrorIs(t, err, ErrProductCustomsInfoInvalid)
		reloaded, err := svc.GetAdminProduct(created.ID)
		require.NoError(t, err)
		require.Equal(t, created.Name, reloaded.Name)
		require.Equal(t, created.MerchantProductIdentifier, reloaded.MerchantProductIdentifier)
		require.Equal(t, created.ManufacturerProductIdentifier, reloaded.ManufacturerProductIdentifier)
	}
}

func TestProductDeclarationIdentifiersAllowMissingOptionalStandardIdentifier(t *testing.T) {
	_, svc := newTestProductService(t)
	input := createProductInputWithRequiredDeclarationIdentifiers()
	input.StandardProductIdentifier = ""
	created, err := svc.CreateAdminProduct(input)
	require.NoError(t, err)
	require.Empty(t, created.StandardProductIdentifier)
}

func TestDedicatedSpokeGTINRejectsInvalidCheckDigitWithoutPartialCreateOrUpdate(t *testing.T) {
	db, productService := newTestProductService(t)
	parent := seedProductCategoryForProductServiceTest(t, db, "Wheel Components", SystemProductCategoryWheelComponentsSlug, nil)
	category := seedProductCategoryForProductServiceTest(t, db, "Spokes", SystemProductCategorySpokesSlug, &parent.ID)
	spokeType := product.ProductTypeSpoke
	input := ProductCreateInput{
		ProductType:               &spokeType,
		SpokeSpecification:        &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray"},
		ProductCategoryID:         &category.ID,
		Name:                      "SAPIM CX-Ray",
		Slug:                      "sapim-cx-ray-gtin-validation",
		Locale:                    "en",
		Status:                    "inactive",
		StandardProductIdentifier: "4006381333932",
	}
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
	changedName := "Must not persist with invalid GTIN"
	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		Name:                            &changedName,
		StandardProductIdentifier:       &invalidGTIN,
		UpdateStandardProductIdentifier: true,
	})
	require.ErrorIs(t, err, ErrProductCustomsInfoInvalid)

	reloaded, err := productService.GetAdminProduct(created.ID)
	require.NoError(t, err)
	require.Equal(t, created.Name, reloaded.Name)
	require.Equal(t, "4006381333931", reloaded.StandardProductIdentifier)
}

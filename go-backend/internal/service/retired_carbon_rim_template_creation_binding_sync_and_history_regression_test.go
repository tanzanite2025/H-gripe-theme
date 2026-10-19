package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/product"
)

func TestRetiredCarbonRimTemplateRejectsNewEntryAndMutationButPreservesHistoricalProduct(t *testing.T) {
	db, svc := newTestProductService(t)
	template := product.ProductSpecificationTemplate{Name: "Historical rim", Slug: "rim", IsEnabled: false}
	require.NoError(t, db.Create(&template).Error)
	_, err := svc.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{Name: "New rim", Slug: "carbon_rim", IsEnabled: true})
	require.ErrorIs(t, err, ErrProductSpecificationTemplateInvalid)
	_, err = svc.UpdateProductSpecificationTemplate(template.ID, ProductSpecificationTemplateInput{Name: "Renamed", Slug: "generic"})
	require.ErrorIs(t, err, ErrProductSpecificationTemplateInvalid)
	input := ProductCreateInput{MerchantProductIdentifier: "TEST-MERCHANT-PRODUCT", ManufacturerProductIdentifier: "TEST-MANUFACTURER-PRODUCT", Name: "Historical", Slug: "historical", Locale: "en", Status: "active", Currency: "USD", ProductSpecificationTemplateID: &template.ID}
	_, err = svc.CreateAdminProduct(input)
	require.ErrorIs(t, err, ErrProductSpecificationTemplateInvalid)
	legacy := product.Product{MerchantProductIdentifier: "TEST-MERCHANT-PRODUCT", ManufacturerProductIdentifier: "TEST-MANUFACTURER-PRODUCT", Name: "Historical", Slug: "historical", Locale: "en", Currency: "USD", ProductSpecificationTemplateID: &template.ID}
	require.NoError(t, db.Create(&legacy).Error)
	title := "Historical name corrected"
	updated, err := svc.UpdateAdminProduct(legacy.ID, ProductUpdateInput{Name: &title})
	require.NoError(t, err)
	require.Equal(t, title, updated.Name)
	_, err = svc.PreviewProductTemplateSync(legacy.ID)
	require.ErrorIs(t, err, ErrProductSpecificationTemplateInvalid)
	_, _, err = svc.CopyAdminProductTranslation(legacy.ID, "fr")
	require.ErrorIs(t, err, ErrProductSpecificationTemplateInvalid)
	dedicated, err := svc.CreateAdminProduct(createCarbonRimProductServiceTestInput(t, db))
	require.NoError(t, err)
	_, err = svc.UpdateAdminProduct(dedicated.ID, ProductUpdateInput{UpdateProductSpecificationTemplateID: true, ProductSpecificationTemplateID: &template.ID})
	require.Error(t, err)
}

package service

import (
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func TestOrdinaryProductGenericSchemaRoundTripsWithoutTemplateBinding(t *testing.T) {
	db, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		MerchantProductIdentifier:     "GENERIC-SCHEMA-MERCHANT",
		ManufacturerProductIdentifier: "GENERIC-SCHEMA-MANUFACTURER",
		Name:                          "Aluminum Bottle Cage",
		Slug:                          "aluminum-bottle-cage",
		Status:                        "active",
		Locale:                        "en",
		GenericProductSchema: &product.GenericProductSchema{
			SchemaVersion: 1,
			Model:         "BC-100",
			Material:      "Aluminum",
			Color:         "Black",
			Size:          "Standard",
			WeightG:       float64Pointer(42.25),
		},
		Variants: []ProductVariantInput{{
			SKU:        "BC-100-BLK",
			PriceMinor: 1999,
			Stock:      5,
			Weight:     42,
			IsDefault:  true,
			IsActive:   boolPtr(true),
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, created.GenericProductSchema)
	require.Nil(t, created.ProductSpecificationTemplateID)
	require.Equal(t, "Aluminum", created.GenericProductSchema.Material)

	loaded, err := productService.GetAdminProduct(created.ID)
	require.NoError(t, err)
	require.Equal(t, created.GenericProductSchema, loaded.GenericProductSchema)

	var stored product.Product
	require.NoError(t, db.First(&stored, created.ID).Error)
	require.Equal(t, created.GenericProductSchema, stored.GenericProductSchema)

	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		GenericProductSchema: &product.GenericProductSchema{SchemaVersion: 1, Material: "Aluminum"},
	})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	require.ErrorContains(t, err, "update marker is required")

	cleared, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateGenericProductSchema: true,
		GenericProductSchema:       nil,
	})
	require.NoError(t, err)
	require.Nil(t, cleared.GenericProductSchema)

	var clearedStored product.Product
	require.NoError(t, db.First(&clearedStored, created.ID).Error)
	require.Nil(t, clearedStored.GenericProductSchema)
}

func TestDedicatedProductRejectsGenericProductSchema(t *testing.T) {
	_, productService := newTestProductService(t)
	productType := product.ProductTypeSpoke
	_, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType: &productType,
		GenericProductSchema: &product.GenericProductSchema{
			SchemaVersion: 1,
			Material:      "Steel",
		},
		Name:   "Dedicated spoke",
		Slug:   "dedicated-spoke-generic-schema-rejection",
		Status: "inactive",
		Locale: "en",
	})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
}

func TestGenericProductSchemaRejectsUnknownFields(t *testing.T) {
	var specification product.GenericProductSchema
	err := json.Unmarshal([]byte(`{"schema_version":1,"material":"Steel","materail":"typo"}`), &specification)
	require.Error(t, err)
}

func TestOrdinaryProductTranslationCopyKeepsGenericSchema(t *testing.T) {
	_, productService := newTestProductService(t)
	source, err := productService.CreateAdminProduct(ProductCreateInput{
		MerchantProductIdentifier:     "ORDINARY-TRANSLATED-MERCHANT",
		ManufacturerProductIdentifier: "ORDINARY-TRANSLATED-MANUFACTURER",
		Name:                          "Ordinary translated product",
		Slug:                          "ordinary-translated-product",
		Status:                        "active",
		Locale:                        "en",
		Currency:                      "USD",
		GenericProductSchema: &product.GenericProductSchema{
			SchemaVersion: 1,
			Model:         "OT-100",
			Material:      "Steel",
		},
		Variants: []ProductVariantInput{{
			SKU:        "OT-100",
			PriceMinor: 1000,
			IsDefault:  true,
			IsActive:   boolPtr(true),
		}},
	})
	require.NoError(t, err)

	translated, _, err := productService.CopyAdminProductTranslation(source.ID, "zh_cn")
	require.NoError(t, err)
	require.NotNil(t, translated.GenericProductSchema)
	require.Equal(t, source.GenericProductSchema, translated.GenericProductSchema)
}

func float64Pointer(value float64) *float64 { return &value }

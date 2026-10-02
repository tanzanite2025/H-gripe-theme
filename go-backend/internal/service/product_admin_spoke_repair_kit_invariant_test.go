package service

import (
	productdomain "commerce-platform/internal/domain/product"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericProductCannotBeConvertedIntoSpokeRepairKitByGenericUpdate(t *testing.T) {
	db, productService := newTestProductService(t)
	template := seedCarbonRimType(t, db)
	created := createProductWithSpecs(
		t,
		productService,
		template.ID,
		"GENERIC-PRODUCT",
		"generic-product",
		map[string]string{"outer_width_mm": "30"},
		map[string]string{"brake_type": "disc"},
	)
	repairKitCategory := seedProductCategoryForProductServiceTest(t, db, "Spoke Repair Kits", productdomain.SpokeRepairKitProductCategorySlug, nil)

	_, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		ProductCategoryID:       &repairKitCategory.ID,
		UpdateProductCategoryID: true,
	})

	require.ErrorIs(t, err, ErrSpokeRepairKitProductTypeImmutable)
	var persisted productdomain.Product
	require.NoError(t, db.First(&persisted, created.ID).Error)
	assert.Nil(t, persisted.ProductCategoryID)
}

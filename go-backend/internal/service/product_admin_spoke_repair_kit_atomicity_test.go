package service

import (
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func TestCreateAdminSpokeRepairKitRollsBackWhenModelBindingFails(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&product.SpokeRepairKitModel{}))
	repairKitCategory := seedProductCategoryForProductServiceTest(
		t,
		db,
		"Spoke Repair Kits",
		product.SpokeRepairKitProductCategorySlug,
		nil,
	)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER reject_spoke_repair_kit_model_insert
		BEFORE INSERT ON product_spoke_repair_kit_models
		BEGIN
			SELECT RAISE(ABORT, 'forced repair-kit model relation failure');
		END;
	`).Error)

	_, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductCategoryID:       &repairKitCategory.ID,
		Name:                    "DT Swiss spoke repair kit",
		Slug:                    "dt-swiss-spoke-repair-kit-atomicity",
		Status:                  "active",
		Locale:                  "en",
		SpokeRepairKitModelKeys: []string{"dt-swiss:arc-1100-dicut-db-38"},
		Variants: []ProductVariantInput{{
			SKU:        "DT-REPAIR-KIT-ATOMICITY",
			PriceMinor: 2500,
			Stock:      4,
			IsDefault:  true,
			IsActive:   boolPtr(true),
		}},
	})
	require.ErrorContains(t, err, "forced repair-kit model relation failure")

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Where("slug = ?", "dt-swiss-spoke-repair-kit-atomicity").Count(&productCount).Error)
	require.Zero(t, productCount, "a failed model binding must roll back the product row")
}

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductServiceReusesSoftDeletedProductAndVariantIdentifiers(t *testing.T) {
	db, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		Name:   "Soft Delete Reuse",
		Slug:   "soft-delete-reuse",
		Status: "active",
		Locale: "en",
		Variants: []ProductVariantInput{{
			SKU:        "SOFT-DELETE-REUSE-VAR",
			PriceMinor: 99,
			Stock:      2,
			IsDefault:  true,
			IsActive:   boolPtr(true),
		}},
	})
	require.NoError(t, err)
	require.Len(t, created.Variants, 1)

	require.NoError(t, db.Delete(&created.Variants[0]).Error)
	require.NoError(t, db.Delete(created).Error)

	recreated, err := productService.CreateAdminProduct(ProductCreateInput{
		Name:   "Soft Delete Reuse Again",
		Slug:   "soft-delete-reuse",
		Status: "active",
		Locale: "en",
		Variants: []ProductVariantInput{{
			SKU:        "SOFT-DELETE-REUSE-VAR",
			PriceMinor: 99,
			Stock:      3,
			IsDefault:  true,
			IsActive:   boolPtr(true),
		}},
	})
	require.NoError(t, err)
	require.NotEqual(t, created.ID, recreated.ID)
	require.Equal(t, "soft-delete-reuse", recreated.Slug)
	require.Equal(t, "SOFT-DELETE-REUSE-VAR", recreated.Variants[0].SKU)
}

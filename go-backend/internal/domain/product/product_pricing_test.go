package product

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductStartingPriceVariantUsesLowestEffectiveActivePrice(t *testing.T) {
	salePrice := 80.0
	item := Product{
		Price:    999,
		Currency: "USD",
		Variants: []ProductVariant{
			{ID: 1, Price: 100, Currency: "USD", IsDefault: true, IsActive: true, SortOrder: 1},
			{ID: 2, Price: 120, SalePrice: &salePrice, Currency: "USD", IsActive: true, SortOrder: 2},
			{ID: 3, Price: 20, Currency: "USD", IsActive: false},
		},
	}

	startingVariant := item.StartingPriceVariant()
	require.NotNil(t, startingVariant)
	require.Equal(t, uint(2), startingVariant.ID)

	price, sale := item.DisplayPrices()
	require.Equal(t, 120.0, price)
	require.NotNil(t, sale)
	require.Equal(t, 80.0, *sale)
	require.Equal(t, "USD", item.DisplayPriceCurrency())
}

func TestProductStartingPriceVariantPrefersDefaultVariantOnPriceTie(t *testing.T) {
	item := Product{
		Variants: []ProductVariant{
			{ID: 1, Price: 100, IsActive: true, SortOrder: 0},
			{ID: 2, Price: 100, IsDefault: true, IsActive: true, SortOrder: 10},
		},
	}

	startingVariant := item.StartingPriceVariant()
	require.NotNil(t, startingVariant)
	require.Equal(t, uint(2), startingVariant.ID)
}

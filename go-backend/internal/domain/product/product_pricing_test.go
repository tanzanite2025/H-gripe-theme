package product

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductStartingPriceVariantUsesLowestActiveCatalogPrice(t *testing.T) {
	item := Product{
		PriceMinor: 99900,
		Currency:   "USD",
		Variants: []ProductVariant{
			{ID: 1, PriceMinor: 10000, Currency: "USD", IsDefault: true, IsActive: true, SortOrder: 1},
			{ID: 2, PriceMinor: 12000, Currency: "USD", IsActive: true, SortOrder: 2},
			{ID: 3, PriceMinor: 2000, Currency: "USD", IsActive: false},
		},
	}

	startingVariant := item.StartingPriceVariant()
	require.NotNil(t, startingVariant)
	require.Equal(t, uint(1), startingVariant.ID)
	require.Equal(t, "100.00", item.DisplayPriceDecimal())
	require.Equal(t, "USD", item.DisplayPriceCurrency())
}

func TestProductStartingPriceVariantPrefersDefaultVariantOnPriceTie(t *testing.T) {
	item := Product{
		Variants: []ProductVariant{
			{ID: 1, PriceMinor: 10000, Currency: "USD", IsActive: true, SortOrder: 0},
			{ID: 2, PriceMinor: 10000, Currency: "USD", IsDefault: true, IsActive: true, SortOrder: 10},
		},
	}

	startingVariant := item.StartingPriceVariant()
	require.NotNil(t, startingVariant)
	require.Equal(t, uint(2), startingVariant.ID)
}

func TestProductJSONDerivesLegacySummaryFieldsFromVariant(t *testing.T) {
	raw, err := json.Marshal(Product{
		SKU:        "legacy-product-sku",
		PriceMinor: 99900,
		Stock:      999,
		Variants: []ProductVariant{{
			SKU:        "VARIANT-SKU",
			PriceMinor: 10000,
			Stock:      7,
			IsActive:   true,
			IsDefault:  true,
			Currency:   "USD",
		}},
	})
	require.NoError(t, err)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &payload))
	require.Equal(t, "VARIANT-SKU", payload["sku"])
	require.Equal(t, "100.00", payload["price_decimal"])
	require.Equal(t, 7.0, payload["stock"])
}

func TestProductMoneyAccessorsRequireCanonicalMinorFields(t *testing.T) {
	item := Product{Currency: "USD", PriceMinor: 1234}
	price, err := item.PriceMoney()
	require.NoError(t, err)
	require.Equal(t, int64(1234), price.AmountMinor())
	variant := ProductVariant{Currency: "USD", PriceMinor: 1234}
	price, err = variant.PriceMoney()
	require.NoError(t, err)
	require.Equal(t, int64(1234), price.AmountMinor())
}

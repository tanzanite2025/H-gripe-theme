package product

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductStartingPriceVariantUsesLowestEffectiveActivePrice(t *testing.T) {
	salePrice := int64(8000)
	item := Product{
		PriceMinor: 99900,
		Currency:   "USD",
		Variants: []ProductVariant{
			{ID: 1, PriceMinor: 10000, Currency: "USD", IsDefault: true, IsActive: true, SortOrder: 1},
			{ID: 2, PriceMinor: 12000, SalePriceMinor: &salePrice, Currency: "USD", IsActive: true, SortOrder: 2},
			{ID: 3, PriceMinor: 2000, Currency: "USD", IsActive: false},
		},
	}

	startingVariant := item.StartingPriceVariant()
	require.NotNil(t, startingVariant)
	require.Equal(t, uint(2), startingVariant.ID)

	price, sale := item.DisplayPrices()
	require.Equal(t, "120.00", price)
	require.NotNil(t, sale)
	require.Equal(t, "80.00", *sale)
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
	sale := int64(8000)
	raw, err := json.Marshal(Product{
		SKU:        "legacy-product-sku",
		PriceMinor: 99900,
		Stock:      999,
		Variants: []ProductVariant{{
			SKU:            "VARIANT-SKU",
			PriceMinor:     10000,
			SalePriceMinor: &sale,
			Stock:          7,
			IsActive:       true,
			IsDefault:      true,
			Currency:       "USD",
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
	saleMinor := int64(1234)
	item := Product{Currency: "USD", PriceMinor: 1234, SalePriceMinor: &saleMinor}
	price, err := item.PriceMoney()
	require.NoError(t, err)
	require.Equal(t, int64(1234), price.AmountMinor())
	sale, err := item.SalePriceMoney()
	require.NoError(t, err)
	require.NotNil(t, sale)
	require.Equal(t, int64(1234), sale.AmountMinor())

	variant := ProductVariant{Currency: "USD", PriceMinor: 1234, SalePriceMinor: &saleMinor}
	price, err = variant.PriceMoney()
	require.NoError(t, err)
	require.Equal(t, int64(1234), price.AmountMinor())
	sale, err = variant.SalePriceMoney()
	require.NoError(t, err)
	require.NotNil(t, sale)
	require.Equal(t, int64(1234), sale.AmountMinor())
}

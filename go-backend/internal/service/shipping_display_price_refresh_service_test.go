package service

import (
	"math/big"
	"testing"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	shippingdomain "commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshShippingDisplayPriceMapConvertsMoneyWithExactRate(t *testing.T) {
	// This rate is just below the half-unit boundary. Converting it to float64
	// first produces 149.5 and would incorrectly round the JPY amount upward.
	rate, ok := new(big.Rat).SetString("149.499999999999999")
	require.True(t, ok)
	amounts, err := shippingTemplateDisplayPriceAmounts(&shippingdomain.ShippingTemplate{
		Currency:        "USD",
		DefaultFeeMinor: 100,
	})
	require.NoError(t, err)
	require.Equal(t, domainmoney.MustNew(100, "USD"), amounts[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee])

	raw := refreshShippingDisplayPriceMap(
		nil,
		amounts,
		"USD",
		[]string{"JPY"},
		map[string]*big.Rat{"JPY": rate},
		shippingdomain.ShippingTemplateDisplayPriceFields,
	)

	snapshots := currency.ParseDisplayPriceSnapshotMap(raw, shippingdomain.ShippingTemplateDisplayPriceFields...)
	require.Len(t, snapshots[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee], 1)
	snapshot := snapshots[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee][0]
	assert.Equal(t, "JPY", snapshot.Currency)
	assert.Equal(t, "149", snapshot.AmountDecimal)
	// Rate is part of the display JSON contract and is intentionally serialized
	// as float64 only after the exact Money conversion has completed.
	assert.Equal(t, 149.5, snapshot.Rate)
}

func TestShippingDisplayPriceAmountsRejectInvalidCurrency(t *testing.T) {
	_, err := shippingTemplateDisplayPriceAmounts(&shippingdomain.ShippingTemplate{
		Currency:        "XXX",
		DefaultFeeMinor: 100,
	})
	require.Error(t, err)
}

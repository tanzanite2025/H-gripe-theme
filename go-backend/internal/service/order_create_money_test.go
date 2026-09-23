package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildOrderMoneyFieldsUsesPricingPipelineMinorSnapshot(t *testing.T) {
	quote := &CheckoutQuote{
		Currency:            "USD",
		SubtotalMinor:       1001,
		ShippingFeeMinor:    99,
		TaxMinor:            17,
		DiscountMinor:       113,
		TotalMinor:          1004,
		PointsDiscountMinor: 3,
	}

	fields, err := buildOrderMoneyFields(quote, "USD", "USD", 1004)
	require.NoError(t, err)
	require.Equal(t, orderMoneyFields{
		PaymentAmountMinor:  1004,
		SubtotalAmountMinor: 1001,
		ShippingFeeMinor:    99,
		TaxAmountMinor:      17,
		DiscountAmountMinor: 113,
		TotalAmountMinor:    1004,
		PointsValueMinor:    3,
	}, fields)
}

func TestResolveExactMinorRejectsNegativeSnapshot(t *testing.T) {
	_, err := resolveExactMinor(-1, "USD")
	require.Error(t, err)
}

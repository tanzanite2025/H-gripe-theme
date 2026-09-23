package payment

import (
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
	"github.com/stretchr/testify/require"
)

func TestPaymentMethodFeeRatePreservesDecimalPrecision(t *testing.T) {
	method := PaymentMethod{
		FeeType:        "percentage",
		FeeRateDecimal: "2.555555555555555",
	}

	require.NoError(t, method.BeforeSave(nil))
	require.Equal(t, "2.555555555555555", method.FeeRateDecimal)

	fee, err := method.CalculateFeeMoney(domainmoney.MustNew(333, "USD"))
	require.NoError(t, err)
	require.Equal(t, int64(9), fee.AmountMinor())
}

func TestPaymentMethodFeeRateRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"-1", "100.000000000000001", "5/2"} {
		method := PaymentMethod{FeeType: "percentage", FeeRateDecimal: value}
		require.Error(t, method.BeforeSave(nil), value)
	}
}

func TestPaymentMethodFixedFeeUsesMinorUnits(t *testing.T) {
	method := PaymentMethod{FeeType: "fixed", FeeValueMinor: 125}

	require.NoError(t, method.BeforeSave(nil))
	fee, err := method.CalculateFeeMoney(domainmoney.MustNew(999, "JPY"))
	require.NoError(t, err)
	require.Equal(t, int64(125), fee.AmountMinor())
	require.Equal(t, "0", method.FeeRateDecimal)
}

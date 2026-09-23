package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefundAmountMoneyPrefersMinorSnapshot(t *testing.T) {
	refund := Refund{
		AmountMinor: 1234,
		Currency:    "USD",
	}

	amount, err := refund.AmountMoney()

	require.NoError(t, err)
	require.Equal(t, int64(1234), amount.AmountMinor())
}

func TestRefundBeforeSaveRequiresCanonicalMinorAmounts(t *testing.T) {
	refund := Refund{
		AmountMinor:                 1234,
		RequestedAmountMinor:        1500,
		DiscountClawbackAmountMinor: 266,
		Currency:                    "USD",
	}

	require.NoError(t, refund.BeforeSave(nil))
	require.Equal(t, int64(1234), refund.AmountMinor)
	require.Equal(t, int64(1500), refund.RequestedAmountMinor)
	require.Equal(t, int64(266), refund.DiscountClawbackAmountMinor)
}

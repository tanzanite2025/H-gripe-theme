package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefundAmountMoneyPrefersMinorSnapshot(t *testing.T) {
	refund := Refund{
		AmountMinor: 1234,
		Amount:      999,
		Currency:    "USD",
	}

	amount, err := refund.AmountMoney()

	require.NoError(t, err)
	require.Equal(t, int64(1234), amount.AmountMinor())
}

func TestRefundBeforeSaveBackfillsAllCoreMinorAmounts(t *testing.T) {
	refund := Refund{
		Amount:                 12.34,
		GiftCardRefundAmount:   1.25,
		RequestedAmount:        15,
		DiscountClawbackAmount: 2.66,
		Currency:               "USD",
	}

	require.NoError(t, refund.BeforeSave(nil))
	require.Equal(t, int64(1234), refund.AmountMinor)
	require.Equal(t, int64(125), refund.GiftCardRefundAmountMinor)
	require.Equal(t, int64(1500), refund.RequestedAmountMinor)
	require.Equal(t, int64(266), refund.DiscountClawbackAmountMinor)
}

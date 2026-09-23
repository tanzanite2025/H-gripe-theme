package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionAmountMoneyUsesMinorSnapshot(t *testing.T) {
	transaction := Transaction{AmountMinor: 12345, Currency: "USD"}

	amount, err := transaction.AmountMoney()

	require.NoError(t, err)
	require.Equal(t, int64(12345), amount.AmountMinor())
}

func TestTransactionBeforeSaveRequiresMinorSnapshot(t *testing.T) {
	transaction := Transaction{Currency: "USD"}

	require.NoError(t, transaction.BeforeSave(nil))
	require.Zero(t, transaction.AmountMinor)

	canonical := Transaction{AmountMinor: 1234, Currency: "USD"}
	require.NoError(t, canonical.BeforeSave(nil))
	require.Equal(t, int64(1234), canonical.AmountMoneyMustMinor())
}

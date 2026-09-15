package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionAmountMoneyUsesMinorSnapshot(t *testing.T) {
	transaction := Transaction{AmountMinor: 12345, Amount: 999, Currency: "USD"}

	amount, err := transaction.AmountMoney()

	require.NoError(t, err)
	require.Equal(t, int64(12345), amount.AmountMinor())
}

func TestTransactionBeforeSaveBackfillsMinorSnapshot(t *testing.T) {
	transaction := Transaction{Amount: 12.34, Currency: "USD"}

	require.NoError(t, transaction.BeforeSave(nil))
	require.Equal(t, int64(1234), transaction.AmountMinor)
}

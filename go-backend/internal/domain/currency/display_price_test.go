package currency

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestDisplayPriceSnapshotsUseDecimalOnlyAmounts(t *testing.T) {
	raw := DisplayPriceSnapshotsJSON([]DisplayPriceSnapshot{
		{AmountDecimal: "91.5", Currency: "USD", QuoteCurrency: "USD"},
		{AmountDecimal: "149", Currency: "JPY", QuoteCurrency: "JPY"},
	}, "CNY")

	var decoded []DisplayPriceSnapshot
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, "91.50", decoded[0].AmountDecimal)
	require.Equal(t, "149", decoded[1].AmountDecimal)
	require.NotContains(t, string(raw), `"amount":`)
}

func TestParseDisplayPriceSnapshotsRejectsLegacyNumericAmount(t *testing.T) {
	parsed := ParseDisplayPriceSnapshots(datatypes.JSON([]byte(`[{"amount":91.5,"currency":"USD"}]`)))
	require.Empty(t, parsed)
}

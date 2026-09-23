package currency

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestOrderFXSnapshotUsesExactDecimalRate(t *testing.T) {
	snapshot := OrderFXSnapshot{
		Version:       OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "EUR",
		RateDecimal:   "0.9000000000000001",
		Source:        "test",
		CapturedAt:    time.Now().UTC(),
	}
	require.NoError(t, snapshot.Validate("EUR"))
	rate, err := snapshot.RateRat()
	require.NoError(t, err)
	require.Equal(t, "9000000000000001/10000000000000000", rate.RatString())

	raw := OrderFXSnapshotJSON(snapshot)
	require.Contains(t, string(raw), `"rate_decimal":"0.9000000000000001"`)
	require.NotContains(t, string(raw), "base_to_order_rate")
	parsed, err := ParseOrderFXSnapshot(raw)
	require.NoError(t, err)
	require.Equal(t, snapshot.RateDecimal, parsed.RateDecimal)
}

func TestOrderFXSnapshotRejectsLegacyFloatRatePayload(t *testing.T) {
	raw := []byte(`{"version":1,"base_currency":"USD","order_currency":"EUR","base_to_order_rate":0.9,"source":"test","captured_at":"2026-09-20T00:00:00Z"}`)
	_, err := ParseOrderFXSnapshot(datatypes.JSON(raw))
	require.Error(t, err)
}

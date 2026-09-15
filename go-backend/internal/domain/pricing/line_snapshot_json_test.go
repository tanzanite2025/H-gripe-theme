package pricing

import (
	"testing"

	"commerce-platform/internal/domain/money"

	"github.com/stretchr/testify/require"
)

func TestLineSnapshotJSONRoundTripPreservesExactAmounts(t *testing.T) {
	snapshot, err := NewSnapshot([]LineInput{
		{Key: "line-a", ProductID: 10, VariantID: 20, Quantity: 2, UnitPrice: money.MustNew(125, "USD")},
	})
	require.NoError(t, err)
	snapshot, err = snapshot.AllocateDiscount(DiscountInput{
		Kind:      DiscountKindCoupon,
		Reference: "SAVE",
		Amount:    money.MustNew(25, "USD"),
	})
	require.NoError(t, err)
	snapshot, err = snapshot.AllocateTax(money.MustNew(10, "USD"))
	require.NoError(t, err)

	raw, err := MarshalLineSnapshot(snapshot.Lines()[0])
	require.NoError(t, err)
	require.Contains(t, string(raw), `"unit_price_minor":125`)
	require.Contains(t, string(raw), `"tax_minor":10`)
	require.NotContains(t, string(raw), ".")

	parsed, err := ParseLineSnapshot(raw)
	require.NoError(t, err)
	require.Equal(t, int64(250), parsed.BaseSubtotal().AmountMinor())
	require.Equal(t, int64(25), parsed.DiscountTotal().AmountMinor())
	require.Equal(t, int64(225), parsed.NetSubtotal().AmountMinor())
	require.Equal(t, int64(10), parsed.Tax().AmountMinor())
	require.Equal(t, DiscountKindCoupon, parsed.DiscountAllocations()[0].Kind())
}

func TestParseLineSnapshotRejectsTamperedTotals(t *testing.T) {
	_, err := ParseLineSnapshot([]byte(`{"schema_version":1,"key":"a","product_id":1,"variant_id":2,"quantity":1,"currency":"USD","unit_price_minor":100,"base_subtotal_minor":100,"discount_total_minor":20,"net_subtotal_minor":90,"discounts":[{"kind":"coupon","amount_minor":20}]}`))
	require.ErrorIs(t, err, ErrInvalidLineSnapshot)
}

func TestParseLineSnapshotRejectsNegativeOrExcessiveDiscounts(t *testing.T) {
	for _, raw := range []string{
		`{"schema_version":1,"key":"a","quantity":1,"currency":"USD","unit_price_minor":-1,"base_subtotal_minor":0,"discount_total_minor":0,"net_subtotal_minor":0}`,
		`{"schema_version":1,"key":"a","quantity":1,"currency":"USD","unit_price_minor":100,"base_subtotal_minor":100,"discount_total_minor":101,"net_subtotal_minor":-1}`,
	} {
		_, err := ParseLineSnapshot([]byte(raw))
		require.ErrorIs(t, err, ErrInvalidLineSnapshot)
	}
}

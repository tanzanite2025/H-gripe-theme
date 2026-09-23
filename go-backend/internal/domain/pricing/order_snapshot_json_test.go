package pricing

import (
	"testing"

	"commerce-platform/internal/domain/money"

	"github.com/stretchr/testify/require"
)

func TestOrderPricingSnapshotRoundTripUsesMinorUnits(t *testing.T) {
	input := orderSnapshotTestInput()
	raw, err := MarshalOrderPricingSnapshot(input)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"total_minor":1080`)
	require.NotContains(t, string(raw), ".")
	payload, err := ParseOrderPricingSnapshot(raw)
	require.NoError(t, err)
	require.Equal(t, int64(1080), payload.TotalMinor)
	require.Equal(t, int64(120), payload.DiscountTotalMinor)
}

func TestOrderPricingSnapshotRejectsInconsistentTotals(t *testing.T) {
	input := orderSnapshotTestInput()
	input.Total = money.MustNew(1081, "USD")
	_, err := MarshalOrderPricingSnapshot(input)
	require.ErrorIs(t, err, ErrInvalidOrderSnapshot)
}

func orderSnapshotTestInput() OrderPricingSnapshotInput {
	return OrderPricingSnapshotInput{
		Currency:         "USD",
		Subtotal:         money.MustNew(1000, "USD"),
		Shipping:         money.MustNew(100, "USD"),
		Tax:              money.MustNew(100, "USD"),
		MemberDiscount:   money.MustNew(50, "USD"),
		CouponDiscount:   money.MustNew(40, "USD"),
		PointsDiscount:   money.MustNew(30, "USD"),
		DiscountTotal:    money.MustNew(120, "USD"),
		Total:            money.MustNew(1080, "USD"),
	}
}

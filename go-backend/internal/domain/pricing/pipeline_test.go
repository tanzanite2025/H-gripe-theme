package pricing

import (
	"testing"

	"commerce-platform/internal/domain/money"

	"github.com/stretchr/testify/require"
)

func TestPipelineAppliesCanonicalStagesUsingPreviousNetAmounts(t *testing.T) {
	pipeline, err := NewPipeline(
		DiscountInput{
			Kind:             DiscountKindMember,
			Reference:        "tier-gold",
			Amount:           money.MustNew(20, "USD"),
			EligibleLineKeys: []string{"a"},
		},
		DiscountInput{
			Kind:      DiscountKindCoupon,
			Reference: "SAVE30",
			Amount:    money.MustNew(30, "USD"),
		},
		DiscountInput{
			Kind:      DiscountKindPoints,
			Reference: "redemption-1",
			Amount:    money.MustNew(10, "USD"),
		},
	)
	require.NoError(t, err)

	snapshot, err := pipeline.Calculate([]LineInput{
		{Key: "a", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		{Key: "b", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
	})
	require.NoError(t, err)
	require.Equal(t, []int64{37, 23}, lineAmounts(snapshot.Lines(), LineSnapshot.DiscountTotal))
	require.Equal(t, []int64{63, 77}, lineAmounts(snapshot.Lines(), LineSnapshot.NetSubtotal))
	require.Equal(t, int64(60), snapshot.DiscountTotal().AmountMinor())
	require.Equal(t, int64(140), snapshot.NetTotal().AmountMinor())
	require.Equal(t,
		[]DiscountKind{DiscountKindMember, DiscountKindCoupon, DiscountKindPoints},
		allocationKinds(snapshot.Lines()[0].DiscountAllocations()),
	)
}

func TestNewPipelineRejectsDuplicateOrOutOfOrderStages(t *testing.T) {
	_, err := NewPipeline(
		DiscountInput{Kind: DiscountKindCoupon},
		DiscountInput{Kind: DiscountKindCoupon},
	)
	require.ErrorIs(t, err, ErrDuplicateDiscountStage)

	_, err = NewPipeline(
		DiscountInput{Kind: DiscountKindCoupon},
		DiscountInput{Kind: DiscountKindMember},
	)
	require.ErrorIs(t, err, ErrDiscountStageOrder)

	_, err = NewPipeline(DiscountInput{Kind: DiscountKind("shipping")})
	require.ErrorIs(t, err, ErrInvalidPipeline)
}

func TestPipelineDefensivelyCopiesStageScope(t *testing.T) {
	eligible := []string{"a"}
	pipeline, err := NewPipeline(DiscountInput{
		Kind:             DiscountKindCoupon,
		Reference:        "original",
		Amount:           money.MustNew(10, "USD"),
		EligibleLineKeys: eligible,
	})
	require.NoError(t, err)
	eligible[0] = "b"

	stages := pipeline.Stages()
	stages[0].EligibleLineKeys[0] = "b"

	snapshot, err := pipeline.Calculate([]LineInput{
		{Key: "a", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		{Key: "b", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
	})
	require.NoError(t, err)
	require.Equal(t, []int64{10, 0}, lineAmounts(snapshot.Lines(), LineSnapshot.DiscountTotal))
	require.Equal(t, "original", snapshot.Lines()[0].DiscountAllocations()[0].Reference())
}

func allocationKinds(allocations []DiscountAllocation) []DiscountKind {
	result := make([]DiscountKind, len(allocations))
	for i, allocation := range allocations {
		result[i] = allocation.Kind()
	}
	return result
}

func TestPricingSnapshotAllocatesTaxAcrossDiscountedLines(t *testing.T) {
	snapshot, err := NewSnapshot([]LineInput{
		{Key: "a", ProductID: 1, Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		{Key: "b", ProductID: 2, Quantity: 1, UnitPrice: money.MustNew(200, "USD")},
	})
	require.NoError(t, err)
	snapshot, err = snapshot.AllocateTax(money.MustNew(10, "USD"))
	require.NoError(t, err)
	lines := snapshot.Lines()
	require.Equal(t, int64(3), lines[0].Tax().AmountMinor())
	require.Equal(t, int64(7), lines[1].Tax().AmountMinor())
	require.Equal(t, "USD", lines[0].Tax().Currency().String())
}

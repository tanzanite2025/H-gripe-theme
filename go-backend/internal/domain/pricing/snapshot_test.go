package pricing

import (
	"math"
	"testing"

	"commerce-platform/internal/domain/money"

	"github.com/stretchr/testify/require"
)

func TestNewSnapshotBuildsExactLineAndOrderTotals(t *testing.T) {
	snapshot, err := NewSnapshot([]LineInput{
		{Key: "line-a", ProductID: 1, Quantity: 2, UnitPrice: money.MustNew(125, "USD")},
		{Key: "line-b", ProductID: 2, Quantity: 1, UnitPrice: money.MustNew(250, "USD")},
	})
	require.NoError(t, err)
	require.Equal(t, SnapshotVersion, snapshot.Version())
	require.Equal(t, int64(500), snapshot.BaseTotal().AmountMinor())
	require.Equal(t, int64(500), snapshot.NetTotal().AmountMinor())
	require.Equal(t, []int64{250, 250}, lineAmounts(snapshot.Lines(), LineSnapshot.BaseSubtotal))
}

func TestAllocateDiscountUsesLargestRemainderAndPreservesTotal(t *testing.T) {
	snapshot := mustSnapshot(t,
		LineInput{Key: "a", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		LineInput{Key: "b", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		LineInput{Key: "c", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
	)

	discounted, err := snapshot.AllocateDiscount(DiscountInput{
		Kind:      DiscountKindCoupon,
		Reference: "SAVE-1",
		Amount:    money.MustNew(100, "USD"),
	})
	require.NoError(t, err)
	require.Equal(t, []int64{34, 33, 33}, lineAmounts(discounted.Lines(), LineSnapshot.DiscountTotal))
	require.Equal(t, int64(100), discounted.DiscountTotal().AmountMinor())
	require.Equal(t, int64(200), discounted.NetTotal().AmountMinor())

	var allocated int64
	for _, line := range discounted.Lines() {
		allocated += line.DiscountTotal().AmountMinor()
	}
	require.Equal(t, discounted.DiscountTotal().AmountMinor(), allocated)
}

func TestAllocateDiscountHonorsEligibleLines(t *testing.T) {
	snapshot := mustSnapshot(t,
		LineInput{Key: "eligible-a", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		LineInput{Key: "excluded", Quantity: 1, UnitPrice: money.MustNew(200, "USD")},
		LineInput{Key: "eligible-b", Quantity: 1, UnitPrice: money.MustNew(50, "USD")},
	)

	discounted, err := snapshot.AllocateDiscount(DiscountInput{
		Kind:             DiscountKindCoupon,
		Reference:        "SCOPED",
		Amount:           money.MustNew(51, "USD"),
		EligibleLineKeys: []string{"eligible-a", "eligible-b"},
	})
	require.NoError(t, err)
	require.Equal(t, []int64{34, 0, 17}, lineAmounts(discounted.Lines(), LineSnapshot.DiscountTotal))
	require.Empty(t, discounted.Lines()[1].DiscountAllocations())
}

func TestAllocateDiscountIsImmutableAndNextStageUsesCurrentNetWeights(t *testing.T) {
	original := mustSnapshot(t,
		LineInput{Key: "a", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		LineInput{Key: "b", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
	)
	member, err := original.AllocateDiscount(DiscountInput{
		Kind:             DiscountKindMember,
		Amount:           money.MustNew(50, "USD"),
		EligibleLineKeys: []string{"a"},
	})
	require.NoError(t, err)
	coupon, err := member.AllocateDiscount(DiscountInput{
		Kind:   DiscountKindCoupon,
		Amount: money.MustNew(30, "USD"),
	})
	require.NoError(t, err)

	require.Equal(t, int64(200), original.NetTotal().AmountMinor())
	require.Equal(t, []int64{0, 0}, lineAmounts(original.Lines(), LineSnapshot.DiscountTotal))
	require.Equal(t, []int64{60, 20}, lineAmounts(coupon.Lines(), LineSnapshot.DiscountTotal))
	require.Equal(t, []int64{40, 80}, lineAmounts(coupon.Lines(), LineSnapshot.NetSubtotal))
}

func TestSnapshotRejectsInvalidLinesAndDiscounts(t *testing.T) {
	_, err := NewSnapshot([]LineInput{
		{Key: "same", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		{Key: "same", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
	})
	require.ErrorIs(t, err, ErrDuplicateLineKey)

	_, err = NewSnapshot([]LineInput{
		{Key: "usd", Quantity: 1, UnitPrice: money.MustNew(100, "USD")},
		{Key: "jpy", Quantity: 1, UnitPrice: money.MustNew(100, "JPY")},
	})
	require.ErrorIs(t, err, money.ErrCurrencyMismatch)

	_, err = NewSnapshot([]LineInput{
		{Key: "overflow", Quantity: 2, UnitPrice: money.MustNew(math.MaxInt64, "USD")},
	})
	require.ErrorIs(t, err, ErrInvalidLine)

	snapshot := mustSnapshot(t, LineInput{Key: "a", Quantity: 1, UnitPrice: money.MustNew(100, "USD")})
	_, err = snapshot.AllocateDiscount(DiscountInput{Kind: DiscountKindCoupon, Amount: money.MustNew(101, "USD")})
	require.ErrorIs(t, err, ErrDiscountExceedsEligible)
	_, err = snapshot.AllocateDiscount(DiscountInput{Kind: DiscountKindCoupon, Amount: money.MustNew(1, "EUR")})
	require.ErrorIs(t, err, ErrInvalidDiscount)
	_, err = snapshot.AllocateDiscount(DiscountInput{
		Kind: DiscountKindCoupon, Amount: money.MustNew(1, "USD"), EligibleLineKeys: []string{"missing"},
	})
	require.ErrorIs(t, err, ErrInvalidDiscount)
	_, err = snapshot.AllocateDiscount(DiscountInput{
		Kind: DiscountKindCoupon, Amount: money.MustNew(0, "USD"), EligibleLineKeys: []string{"missing"},
	})
	require.ErrorIs(t, err, ErrInvalidDiscount)
}

func mustSnapshot(t *testing.T, inputs ...LineInput) Snapshot {
	t.Helper()
	snapshot, err := NewSnapshot(inputs)
	require.NoError(t, err)
	return snapshot
}

func lineAmounts(lines []LineSnapshot, amount func(LineSnapshot) money.Money) []int64 {
	result := make([]int64, len(lines))
	for i, line := range lines {
		result[i] = amount(line).AmountMinor()
	}
	return result
}

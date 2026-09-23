package money

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNormalizesAndValidatesCurrency(t *testing.T) {
	value, err := New(1234, " usd ")
	require.NoError(t, err)
	require.Equal(t, int64(1234), value.AmountMinor())
	require.Equal(t, "USD", value.Currency().String())

	_, err = New(1, "XXX")
	require.Error(t, err)
}

func TestParseMajorAvoidsFloatingPoint(t *testing.T) {
	value, err := ParseMajor(" 1.23 ", "USD")
	require.NoError(t, err)
	require.Equal(t, int64(123), value.AmountMinor())

	jpy, err := ParseMajor("123.00", "JPY")
	require.NoError(t, err)
	require.Equal(t, int64(123), jpy.AmountMinor())

	_, err = ParseMajor("1.001", "USD")
	require.ErrorIs(t, err, ErrExcessPrecision)
}

func TestParseMajorRejectsMalformedAndOverflowInput(t *testing.T) {
	for _, input := range []string{"", ".50", "1.", "1.2.3", "USD 1"} {
		_, err := ParseMajor(input, "USD")
		require.ErrorIs(t, err, ErrInvalidAmount, input)
	}
	_, err := ParseMajor("92233720368547758.08", "USD")
	require.Error(t, err)
}

func TestFormatMajorUsesCurrencyScale(t *testing.T) {
	value := MustNew(1234, "USD")
	formatted, err := value.FormatMajor()
	require.NoError(t, err)
	require.Equal(t, "12.34", formatted)

	zeroDecimal := MustNew(1234, "JPY")
	formatted, err = zeroDecimal.FormatMajor()
	require.NoError(t, err)
	require.Equal(t, "1234", formatted)

	negative := MustNew(-5, "USD")
	formatted, err = negative.FormatMajor()
	require.NoError(t, err)
	require.Equal(t, "-0.05", formatted)
}

func TestAddAndSubtractRequireSameCurrency(t *testing.T) {
	left := MustNew(100, "USD")
	right := MustNew(25, "USD")

	sum, err := left.Add(right)
	require.NoError(t, err)
	require.Equal(t, int64(125), sum.AmountMinor())

	difference, err := left.Subtract(right)
	require.NoError(t, err)
	require.Equal(t, int64(75), difference.AmountMinor())

	_, err = left.Add(MustNew(1, "JPY"))
	require.ErrorIs(t, err, ErrCurrencyMismatch)
}

func TestAddRejectsOverflow(t *testing.T) {
	_, err := MustNew(math.MaxInt64, "USD").Add(MustNew(1, "USD"))
	require.Error(t, err)
}

func TestMultiplyIntRetainsCurrencyAndRejectsOverflow(t *testing.T) {
	product, err := MustNew(125, "USD").MultiplyInt(3)
	require.NoError(t, err)
	require.Equal(t, int64(375), product.AmountMinor())
	require.Equal(t, "USD", product.Currency().String())

	_, err = MustNew(math.MaxInt64, "USD").MultiplyInt(2)
	require.Error(t, err)
}

func TestMultiplyRatioRoundsHalfAwayFromZero(t *testing.T) {
	value, err := MustNew(100, "USD").MultiplyRatio(1, 3)
	require.NoError(t, err)
	require.Equal(t, int64(33), value.AmountMinor())

	value, err = MustNew(1, "USD").MultiplyRatio(1, 2)
	require.NoError(t, err)
	require.Equal(t, int64(1), value.AmountMinor())

	value, err = MustNew(-1, "USD").MultiplyRatio(1, 2)
	require.NoError(t, err)
	require.Equal(t, int64(-1), value.AmountMinor())

	_, err = MustNew(1, "USD").MultiplyRatio(1, 0)
	require.Error(t, err)
}

func TestMultiplyRatRoundsDecimalRateInMinorUnits(t *testing.T) {
	rate := new(big.Rat).SetFrac(big.NewInt(55), big.NewInt(1000))
	value, err := MustNew(333, "USD").MultiplyRat(rate)
	require.NoError(t, err)
	require.Equal(t, int64(18), value.AmountMinor())

	negative, err := MustNew(-333, "USD").MultiplyRat(rate)
	require.NoError(t, err)
	require.Equal(t, int64(-18), negative.AmountMinor())
}

func TestConvertAtRatUsesTargetMinorUnits(t *testing.T) {
	converted, err := MustNew(100, "USD").ConvertAtRat(new(big.Rat).SetInt64(150), "JPY")
	require.NoError(t, err)
	require.Equal(t, int64(150), converted.AmountMinor())

	converted, err = MustNew(1, "JPY").ConvertAtRat(new(big.Rat).SetFrac64(67, 10000), "USD")
	require.NoError(t, err)
	require.Equal(t, int64(1), converted.AmountMinor())
}

func TestAbsRejectsMinimumInt64Overflow(t *testing.T) {
	absolute, err := MustNew(-25, "USD").Abs()
	require.NoError(t, err)
	require.Equal(t, int64(25), absolute.AmountMinor())

	_, err = MustNew(math.MinInt64, "USD").Abs()
	require.Error(t, err)
}

func TestSubtractHandlesMinimumInt64WithoutNegationOverflow(t *testing.T) {
	result, err := MustNew(math.MinInt64, "USD").Subtract(MustNew(math.MinInt64, "USD"))
	require.NoError(t, err)
	require.Zero(t, result.AmountMinor())

	_, err = MustNew(0, "USD").Subtract(MustNew(math.MinInt64, "USD"))
	require.Error(t, err)
}

func TestAllocateUsesLargestRemainderAndPreservesTotal(t *testing.T) {
	allocations, err := MustNew(100, "USD").Allocate([]int64{1, 1, 1})
	require.NoError(t, err)
	require.Equal(t, []int64{34, 33, 33}, minorUnits(allocations))

	negative, err := MustNew(-100, "USD").Allocate([]int64{1, 1, 1})
	require.NoError(t, err)
	require.Equal(t, []int64{-34, -33, -33}, minorUnits(negative))
}

func TestAllocateRejectsInvalidRatios(t *testing.T) {
	for _, ratios := range [][]int64{nil, {0, 0}, {1, -1}} {
		_, err := MustNew(100, "USD").Allocate(ratios)
		require.ErrorIs(t, err, ErrInvalidAllocation)
	}
}

func TestZeroValueIsNotValidMoney(t *testing.T) {
	require.Error(t, (Money{}).Validate())
}

func minorUnits(values []Money) []int64 {
	result := make([]int64, len(values))
	for i, value := range values {
		result[i] = value.AmountMinor()
	}
	return result
}

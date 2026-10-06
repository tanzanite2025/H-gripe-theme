package service

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeSpokeTensionRatioMatchesReportReference(t *testing.T) {
	result := computeSpokeTensionRatio(26.2, 20.3, 270, 270)

	require.NotNil(t, result)
	require.InDelta(t, 20.3/26.2, result.LeftToRight, 0.0001)
	require.InDelta(t, 20.3/26.2, result.LowerToHigher, 0.0001)
	require.Equal(t, "left", result.LowerSide)
}

func TestComputeSpokeTensionRatioDetectsSymmetricGeometry(t *testing.T) {
	result := computeSpokeTensionRatio(35, 35, 270, 270)

	require.NotNil(t, result)
	require.Equal(t, "balanced", result.LowerSide)
	require.InDelta(t, 1, result.LeftToRight, 0.0001)
	require.InDelta(t, 1, result.LowerToHigher, 0.0001)
}

func TestComputeSpokeTensionRatioFromBracingSumsAccountsForUnequalSpokeCounts(t *testing.T) {
	// A 2:1 layout has 8 left spokes and 16 right spokes. Summed lateral
	// components, rather than one representative spoke per side, determine the
	// per-spoke tension ratio required for static equilibrium.
	leftPerSpokeSin := 0.117
	rightPerSpokeSin := 0.0585
	result, err := computeSpokeTensionRatioFromBracingSums(
		8*leftPerSpokeSin,
		16*rightPerSpokeSin,
		8,
		16,
	)

	require.NoError(t, err)
	require.InDelta(t, (16*rightPerSpokeSin)/(8*leftPerSpokeSin), result.LeftToRight, 0.0001)
	require.InDelta(t, 1, result.LowerToHigher, 0.0001)
	require.Equal(t, "balanced", result.LowerSide)
	require.InDelta(t, math.Asin(leftPerSpokeSin)*180/math.Pi, result.LeftBracingAngleDeg, 0.0001)
	require.InDelta(t, math.Asin(rightPerSpokeSin)*180/math.Pi, result.RightBracingAngleDeg, 0.0001)
}

func TestComputeSpokeTensionRatioFromBracingSumsRejectsMissingSide(t *testing.T) {
	_, err := computeSpokeTensionRatioFromBracingSums(0, 1, 0, 8)
	require.ErrorIs(t, err, ErrInvalidSpokeCalculation)
}

func TestEffectiveSpokeFlangeDistanceAppliesSignedRimOffset(t *testing.T) {
	require.InDelta(t, 29, effectiveSpokeFlangeDistance(26.2, 2.8, "left"), 0.0001)
	require.InDelta(t, 17.5, effectiveSpokeFlangeDistance(20.3, 2.8, "right"), 0.0001)
}

func TestAlternatingDrillingOffsetPositiveMovesRimHoleTowardPhysicalLeft(t *testing.T) {
	positiveLeftOffset := 1.0
	leftDistance := effectiveSpokeFlangeDistance(22.5, 0, "left") - positiveLeftOffset
	rightDistance := effectiveSpokeFlangeDistance(35.6, 0, "right") + positiveLeftOffset

	require.InDelta(t, 21.5, leftDistance, 0.0001)
	require.InDelta(t, 36.6, rightDistance, 0.0001)
}

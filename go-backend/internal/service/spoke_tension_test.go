package service

import (
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

func TestEffectiveSpokeFlangeDistanceAppliesSignedRimOffset(t *testing.T) {
	require.InDelta(t, 29, effectiveSpokeFlangeDistance(26.2, 2.8, "left"), 0.0001)
	require.InDelta(t, 17.5, effectiveSpokeFlangeDistance(20.3, 2.8, "right"), 0.0001)
}

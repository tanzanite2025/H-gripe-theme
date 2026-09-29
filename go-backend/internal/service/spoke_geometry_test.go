package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSpokeGeometryInput() spokeGeometryInput {
	return spokeGeometryInput{
		Left: spokeGeometrySideInput{
			RimRadiusMM:      299,
			FlangeRadiusMM:   22,
			FlangeDistanceMM: 22.5,
			PhaseRad:         1.0471975512,
		},
		Right: spokeGeometrySideInput{
			RimRadiusMM:      299,
			FlangeRadiusMM:   22,
			FlangeDistanceMM: 35.6,
			PhaseRad:         1.0471975512,
		},
		SpokeHoleDiameterMM:         2.5,
		StraightPullTangentOffsetMM: 2,
	}
}

func TestCalculateSpokeGeometryDispatchesJBendAndStraightPull(t *testing.T) {
	input := testSpokeGeometryInput()

	jBend, err := calculateSpokeGeometry(spokeHeadTypeJBend, input)
	require.NoError(t, err)
	assert.InDelta(t, 1.25, jBend.SpokeHoleCorrectionMM, 0.0001)
	assert.Zero(t, jBend.StraightPullTangentOffsetMM)

	straightPull, err := calculateSpokeGeometry(spokeHeadTypeStraightPull, input)
	require.NoError(t, err)
	assert.Zero(t, straightPull.SpokeHoleCorrectionMM)
	assert.Equal(t, input.StraightPullTangentOffsetMM, straightPull.StraightPullTangentOffsetMM)
	assert.NotEqual(t, jBend.LeftLengthMM, straightPull.LeftLengthMM)
}

func TestCalculateSpokeGeometryRejectsUnknownHeadType(t *testing.T) {
	_, err := calculateSpokeGeometry("unknown", testSpokeGeometryInput())
	require.ErrorIs(t, err, ErrInvalidSpokeCalculation)
}

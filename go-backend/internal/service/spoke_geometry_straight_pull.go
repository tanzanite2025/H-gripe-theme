package service

import (
	"fmt"
	"math"
)

type straightPullSpokeGeometryCalculator struct{}

func (straightPullSpokeGeometryCalculator) calculate(input spokeGeometryInput) (spokeGeometryResult, error) {
	left := straightPullSpokeLength(input.Left, input.StraightPullTangentOffsetMM)
	right := straightPullSpokeLength(input.Right, input.StraightPullTangentOffsetMM)
	if !isFinite(left) || !isFinite(right) || left <= 0 || right <= 0 {
		return spokeGeometryResult{}, fmt.Errorf("%w: spoke geometry produced an invalid straight-pull vector", ErrInvalidSpokeCalculation)
	}

	correction := input.SpokeHoleDiameterMM / 2
	left -= correction
	right -= correction
	if left <= 0 || right <= 0 {
		return spokeGeometryResult{}, fmt.Errorf("%w: spoke geometry became invalid after hole-edge correction", ErrInvalidSpokeCalculation)
	}

	return spokeGeometryResult{
		LeftLengthMM:                left,
		RightLengthMM:               right,
		SpokeHoleCorrectionMM:       correction,
		StraightPullTangentOffsetMM: input.StraightPullTangentOffsetMM,
	}, nil
}

func straightPullSpokeLength(input spokeGeometrySideInput, tangentOffsetMM float64) float64 {
	// The slot point is (flange radius, tangential offset) in the local
	// flange frame. Resolve the rim point into that frame and add the axial
	// component separately; no J-bend polar hub-hole angle is synthesized.
	radialDelta := input.RimRadiusMM*math.Cos(input.PhaseRad) - input.FlangeRadiusMM
	tangentialDelta := input.RimRadiusMM*math.Sin(input.PhaseRad) - tangentOffsetMM
	squared := radialDelta*radialDelta + tangentialDelta*tangentialDelta + input.FlangeDistanceMM*input.FlangeDistanceMM
	if !isFinite(squared) || squared < 0 {
		return math.NaN()
	}
	return math.Sqrt(squared)
}

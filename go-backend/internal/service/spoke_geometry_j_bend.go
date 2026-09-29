package service

import (
	"fmt"
	"math"
)

const defaultSpokeHoleDiameterMM = 2.5

type jBendSpokeGeometryCalculator struct{}

func (jBendSpokeGeometryCalculator) calculate(input spokeGeometryInput) (spokeGeometryResult, error) {
	left := jBendSpokeLength(input.Left)
	right := jBendSpokeLength(input.Right)
	if !isFinite(left) || !isFinite(right) || left <= 0 || right <= 0 {
		return spokeGeometryResult{}, fmt.Errorf("%w: spoke geometry produced an invalid triangle", ErrInvalidSpokeCalculation)
	}

	correction := input.SpokeHoleDiameterMM / 2
	left -= correction
	right -= correction
	return spokeGeometryResult{
		LeftLengthMM:                left,
		RightLengthMM:               right,
		SpokeHoleCorrectionMM:       correction,
		StraightPullTangentOffsetMM: 0,
	}, nil
}

func jBendSpokeLength(input spokeGeometrySideInput) float64 {
	squared := input.RimRadiusMM*input.RimRadiusMM +
		input.FlangeRadiusMM*input.FlangeRadiusMM +
		input.FlangeDistanceMM*input.FlangeDistanceMM -
		2*input.RimRadiusMM*input.FlangeRadiusMM*math.Cos(input.PhaseRad)
	if !isFinite(squared) || squared < 0 {
		return math.NaN()
	}
	return math.Sqrt(squared)
}

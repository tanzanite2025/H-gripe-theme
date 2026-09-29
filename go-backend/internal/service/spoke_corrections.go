package service

import (
	"fmt"
	"math"
)

const (
	defaultElasticModulusNMM2 = 200000.0
	defaultInterlaceOffsetMM  = 0.45
)

var spokeProfileAreasMM2 = map[string]float64{
	"round_2_0":      math.Pi * 2.0 * 2.0 / 4.0,
	"round_1_8":      math.Pi * 1.8 * 1.8 / 4.0,
	"bladed_0_9x2_2": 0.9 * 2.2,
}

type spokePhysicalCorrectionInput struct {
	LeftLengthMM  float64
	RightLengthMM float64

	NippleType              string
	NippleLengthMM          *float64
	Crossing                int
	Interlacing             bool
	InterlaceCompensationMM *float64
	SpokeProfile            string
	TargetTensionN          float64
}

type spokePhysicalCorrectionResult struct {
	LeftLengthMM  float64
	RightLengthMM float64

	InterlaceCompensationMM float64
	StretchLeftMM           float64
	StretchRightMM          float64
}

func applySpokePhysicalCorrections(input spokePhysicalCorrectionInput) (spokePhysicalCorrectionResult, error) {
	left := input.LeftLengthMM
	right := input.RightLengthMM
	if !isFinite(left) || !isFinite(right) || left <= 0 || right <= 0 {
		return spokePhysicalCorrectionResult{}, fmt.Errorf("%w: spoke geometry produced an invalid length", ErrInvalidSpokeCalculation)
	}

	if input.NippleType == "hidden" && input.NippleLengthMM != nil {
		correction := *input.NippleLengthMM - 3
		if correction < 0 {
			return spokePhysicalCorrectionResult{}, fmt.Errorf("%w: hidden nipple length must be at least 3mm", ErrInvalidSpokeCalculation)
		}
		left += correction
		right += correction
	}

	interlaceCompensationMM := 0.0
	if input.Interlacing && input.Crossing > 0 {
		if input.InterlaceCompensationMM != nil {
			interlaceCompensationMM = *input.InterlaceCompensationMM
		} else {
			interlaceCompensationMM = defaultInterlaceCompensation(input.Crossing)
		}
		left += interlaceCompensationMM
		right += interlaceCompensationMM
	}

	stretchLeftMM, stretchRightMM := 0.0, 0.0
	if input.TargetTensionN > 0 {
		areaMM2 := spokeProfileAreasMM2[input.SpokeProfile]
		stretchLeftMM = spokeElasticStretchMM(input.TargetTensionN, left, areaMM2)
		stretchRightMM = spokeElasticStretchMM(input.TargetTensionN, right, areaMM2)
		left -= stretchLeftMM
		right -= stretchRightMM
	}
	if !isFinite(left) || !isFinite(right) || left <= 0 || right <= 0 {
		return spokePhysicalCorrectionResult{}, fmt.Errorf("%w: physical spoke corrections produced a non-positive length", ErrInvalidSpokeCalculation)
	}

	return spokePhysicalCorrectionResult{
		LeftLengthMM:            left,
		RightLengthMM:           right,
		InterlaceCompensationMM: interlaceCompensationMM,
		StretchLeftMM:           stretchLeftMM,
		StretchRightMM:          stretchRightMM,
	}, nil
}

func defaultInterlaceCompensation(crossing int) float64 {
	switch {
	case crossing >= 3:
		return defaultInterlaceOffsetMM
	case crossing == 2:
		return 0.4
	case crossing == 1:
		return 0.25
	default:
		return 0
	}
}

func spokeElasticStretchMM(targetTensionN, lengthMM, areaMM2 float64) float64 {
	if targetTensionN <= 0 || lengthMM <= 0 || areaMM2 <= 0 {
		return 0
	}
	return targetTensionN * lengthMM / (defaultElasticModulusNMM2 * areaMM2)
}

// computeSpokeTensionRatio is retained for package-level compatibility. The
// service path uses the error-returning variant so invalid values fail loudly.
func computeSpokeTensionRatio(leftBracingDistance, rightBracingDistance, leftLength, rightLength float64) *SpokeTensionRatio {
	result, _ := computeSpokeTensionRatioSafe(leftBracingDistance, rightBracingDistance, leftLength, rightLength)
	return result
}

func computeSpokeTensionRatioSafe(leftBracingDistance, rightBracingDistance, leftLength, rightLength float64) (*SpokeTensionRatio, error) {
	if leftBracingDistance <= 0 || rightBracingDistance <= 0 || leftLength <= 0 || rightLength <= 0 {
		return nil, fmt.Errorf("%w: bracing geometry must be positive", ErrInvalidSpokeCalculation)
	}
	if !isFinite(leftBracingDistance) || !isFinite(rightBracingDistance) || !isFinite(leftLength) || !isFinite(rightLength) {
		return nil, fmt.Errorf("%w: tension ratio calculation received a non-finite value", ErrInvalidSpokeCalculation)
	}

	leftSin := math.Min(1, leftBracingDistance/leftLength)
	rightSin := math.Min(1, rightBracingDistance/rightLength)
	if leftSin <= 0 || rightSin <= 0 || !isFinite(leftSin) || !isFinite(rightSin) {
		return nil, fmt.Errorf("%w: tension ratio calculation diverged", ErrInvalidSpokeCalculation)
	}

	leftToRight := rightSin / leftSin
	rightToLeft := leftSin / rightSin
	lowerToHigher := math.Min(leftToRight, rightToLeft)

	lowerSide := "balanced"
	switch {
	case leftToRight < 0.995:
		lowerSide = "left"
	case leftToRight > 1.005:
		lowerSide = "right"
	}

	result := &SpokeTensionRatio{
		LeftToRight:          roundSpokeRatio(leftToRight),
		RightToLeft:          roundSpokeRatio(rightToLeft),
		LowerToHigher:        roundSpokeRatio(lowerToHigher),
		LowerSide:            lowerSide,
		LeftBracingAngleDeg:  roundSpokeRatio(math.Asin(leftSin) * 180 / math.Pi),
		RightBracingAngleDeg: roundSpokeRatio(math.Asin(rightSin) * 180 / math.Pi),
	}
	if !isFinite(result.LeftToRight) || !isFinite(result.RightToLeft) || !isFinite(result.LowerToHigher) {
		return nil, fmt.Errorf("%w: tension ratio calculation diverged", ErrInvalidSpokeCalculation)
	}
	return result, nil
}

func roundSpokeLength(value float64) float64 { return math.Round(value*100) / 100 }

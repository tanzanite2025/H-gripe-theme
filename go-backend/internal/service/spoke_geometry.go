package service

import (
	"commerce-platform/internal/domain/spoke"
	"fmt"
	"math"
)

const (
	spokeHeadTypeJBend        = "j_bend"
	spokeHeadTypeStraightPull = "straight_pull"
)

// spokeGeometrySideInput describes one flange side. Keeping the phase on the
// side input lets a future asymmetric layout (for example 2:1 lacing) supply
// different spoke phases without changing the J-bend/Straight Pull interface.
type spokeGeometrySideInput struct {
	RimRadiusMM      float64
	FlangeRadiusMM   float64
	FlangeDistanceMM float64
	PhaseRad         float64
}

type spokeGeometryInput struct {
	Left  spokeGeometrySideInput
	Right spokeGeometrySideInput

	// SpokeHoleDiameterMM supplies the flange-hole inner-edge correction for
	// both head geometries. The straight-pull calculator applies the same
	// radius deduction after resolving its tangent slot vector.
	SpokeHoleDiameterMM float64
	// StraightPullTangentOffsetMM is consumed by the Straight Pull calculator
	// only. It is measured in the local flange frame in millimeters.
	StraightPullTangentOffsetMM float64
}

type spokeGeometryResult struct {
	LeftLengthMM  float64
	RightLengthMM float64

	SpokeHoleCorrectionMM       float64
	StraightPullTangentOffsetMM float64
}

type spokeGeometryCalculator interface {
	calculate(input spokeGeometryInput) (spokeGeometryResult, error)
}

func calculateSpokeGeometry(headType string, input spokeGeometryInput) (spokeGeometryResult, error) {
	var calculator spokeGeometryCalculator
	switch headType {
	case spokeHeadTypeJBend:
		calculator = jBendSpokeGeometryCalculator{}
	case spokeHeadTypeStraightPull:
		calculator = straightPullSpokeGeometryCalculator{}
	default:
		return spokeGeometryResult{}, fmt.Errorf("%w: unsupported spoke head type %q", ErrInvalidSpokeCalculation, headType)
	}
	return calculator.calculate(input)
}

// calculateSingleSpokeGeometryLength reuses the existing J-bend and
// straight-pull geometry calculators for one mapped spoke. Passing the same
// side input twice keeps the established correction and validation path while
// allowing each topology mapping to supply its own phase.
func calculateSingleSpokeGeometryLength(headType string, sideInput spokeGeometrySideInput, spokeHoleDiameterMM, straightPullTangentOffsetMM float64) (spokeGeometryResult, error) {
	return calculateSpokeGeometry(headType, spokeGeometryInput{
		Left:                        sideInput,
		Right:                       sideInput,
		SpokeHoleDiameterMM:         spokeHoleDiameterMM,
		StraightPullTangentOffsetMM: straightPullTangentOffsetMM,
	})
}

func effectiveSpokeFlangeDistance(flangeDistance, rimOffset float64, side string) float64 {
	if side == "left" {
		return flangeDistance + rimOffset
	}
	return flangeDistance - rimOffset
}

func finiteGeometry(geometry *spoke.HubGeometry) bool {
	if geometry == nil || geometry.LeftFlange == nil || geometry.RightFlange == nil || geometry.LeftFlangePCD == nil || geometry.RightFlangePCD == nil {
		return false
	}
	return isFinite(*geometry.LeftFlange) && isFinite(*geometry.RightFlange) &&
		isFinite(*geometry.LeftFlangePCD) && isFinite(*geometry.RightFlangePCD) &&
		*geometry.LeftFlange > 0 && *geometry.LeftFlange <= 100 &&
		*geometry.RightFlange > 0 && *geometry.RightFlange <= 100 &&
		*geometry.LeftFlangePCD >= 10 && *geometry.LeftFlangePCD <= 150 &&
		*geometry.RightFlangePCD >= 10 && *geometry.RightFlangePCD <= 150 &&
		(geometry.SpokeHoleDiameter == nil || (isFinite(*geometry.SpokeHoleDiameter) && *geometry.SpokeHoleDiameter >= 0 && *geometry.SpokeHoleDiameter <= 10))
}

func isFinite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

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

	// SpokeHoleDiameterMM is consumed by the J-bend calculator only.
	SpokeHoleDiameterMM float64
	// StraightPullTangentOffsetMM is consumed by the Straight Pull calculator
	// only. It is measured in the local flange frame in millimetres.
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

func spokeLacingPhaseRadians(crossing, spokeCount int) float64 {
	return (720.0 * float64(crossing) / float64(spokeCount)) * math.Pi / 180.0
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

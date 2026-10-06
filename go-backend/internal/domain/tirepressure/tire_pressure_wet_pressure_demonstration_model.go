package tirepressure

import "math"

const (
	WetRoadDemoDefaultWaterFilmDepthMm       = 1.0
	WetRoadDemoReferenceSpeedKmh             = 30.0
	WetRoadDemoReferenceFrictionLoss         = 0.20
	WetRoadDemoMinimumFrictionRetentionRatio = 0.45
	WetRoadDemoMaximumWaterFilmDepthMm       = 5.0
	WetRoadDemoSurfaceTextureBaseline        = "INDOOR_FLAT_BASELINE"
	WetRoadDemoRubberBaseline                = "GENERIC_NON_RADIAL_RUBBER"
)

// WetPressureCompensation contains an explicit equivalent-pressure comparison.
// It uses a bounded demonstration proxy for wet-friction retention; it is not a
// calibrated tire-road friction law or a pressure recommendation.
type WetPressureCompensation struct {
	WaterFilmDepthMm                 float64  `json:"water_film_depth_mm"`
	SpeedKmh                         float64  `json:"speed_kmh"`
	LateralDemandRatio               float64  `json:"lateral_demand_ratio"`
	FrictionRetentionRatio           float64  `json:"friction_retention_ratio"`
	ReferencePressurePsi             float64  `json:"reference_pressure_psi"`
	EquivalentPressurePsi            float64  `json:"equivalent_pressure_psi"`
	PressureReductionPsi             float64  `json:"pressure_reduction_psi"`
	PressureReductionPct             float64  `json:"pressure_reduction_pct"`
	ReferenceContactAreaCm2          float64  `json:"reference_contact_area_cm2"`
	EquivalentContactAreaCm2         float64  `json:"equivalent_contact_area_cm2"`
	ContactAreaChangePct             float64  `json:"contact_area_change_pct"`
	WetGripLimitAtReferencePressureN float64  `json:"wet_grip_limit_at_reference_pressure_n"`
	WetGripMarginPct                 float64  `json:"wet_grip_margin_pct"`
	PressureClampedToMinimum         bool     `json:"pressure_clamped_to_minimum"`
	MinimumPressurePsi               *float64 `json:"minimum_pressure_psi,omitempty"`
	SurfaceTextureBaseline           string   `json:"surface_texture_baseline"`
	RubberBaseline                   string   `json:"rubber_baseline"`
}

// CalculateWetRoadFrictionRetentionRatio applies a transparent first-order
// proxy: dynamic water pressure grows with speed squared, the water film
// scales the loss linearly, and higher lateral demand increases sensitivity.
// The coefficients are demonstration assumptions and must not be presented as
// measured values for a specific tire, rubber, or road.
func CalculateWetRoadFrictionRetentionRatio(speedKmh, waterFilmDepthMm, lateralDemandRatio float64) float64 {
	if speedKmh <= 0 || waterFilmDepthMm <= 0 {
		return 1
	}
	normalizedSpeed := speedKmh / WetRoadDemoReferenceSpeedKmh
	normalizedWaterFilm := waterFilmDepthMm / WetRoadDemoDefaultWaterFilmDepthMm
	boundedLateralDemandRatio := math.Max(0, math.Min(1, lateralDemandRatio))
	demandFactor := 0.5 + 0.5*boundedLateralDemandRatio
	loss := WetRoadDemoReferenceFrictionLoss * normalizedSpeed * normalizedSpeed * normalizedWaterFilm * demandFactor
	return math.Max(WetRoadDemoMinimumFrictionRetentionRatio, math.Min(1, 1-loss))
}

func resolveWetRoadWaterFilmDepthMm(value *float64) float64 {
	if value == nil {
		return WetRoadDemoDefaultWaterFilmDepthMm
	}
	return *value
}

func attachWetPressureCompensation(wheel *WheelDynamics, referencePressurePsi, minimumPressurePsi *float64, speedKmh, waterFilmDepthMm, tireWidthMm float64) {
	if wheel == nil || wheel.EstimatedStaticContactCm2 == nil || referencePressurePsi == nil {
		return
	}
	if !isFinitePositiveTirePressureCalculationNumber(*referencePressurePsi) || !isFinitePositiveTirePressureCalculationNumber(*wheel.EstimatedStaticContactCm2) {
		return
	}
	lateralDemandRatio := 0.0
	if wheel.IdealizedGripLimitN > 0 {
		lateralDemandRatio = math.Abs(wheel.LateralDemandN) / wheel.IdealizedGripLimitN
	}
	retentionRatio := CalculateWetRoadFrictionRetentionRatio(speedKmh, waterFilmDepthMm, lateralDemandRatio)
	equivalentPressurePsi := *referencePressurePsi * retentionRatio
	pressureClampedToMinimum := false
	if minimumPressurePsi != nil && *minimumPressurePsi > equivalentPressurePsi {
		equivalentPressurePsi = *minimumPressurePsi
		pressureClampedToMinimum = true
	}
	if !isFinitePositiveTirePressureCalculationNumber(equivalentPressurePsi) {
		return
	}
	equivalentAreaCm2 := *wheel.EstimatedStaticContactCm2 * *referencePressurePsi / equivalentPressurePsi
	if !isFinitePositiveTirePressureCalculationNumber(equivalentAreaCm2) {
		return
	}
	pressureReductionPsi := *referencePressurePsi - equivalentPressurePsi
	pressureReductionPct := pressureReductionPsi / *referencePressurePsi * 100
	if !isFiniteTirePressureCalculationNumber(pressureReductionPsi) || !isFiniteTirePressureCalculationNumber(pressureReductionPct) {
		return
	}
	wetGripAtReferencePressureN := wheel.IdealizedGripLimitN * retentionRatio
	comparison := &WetPressureCompensation{
		WaterFilmDepthMm:                 roundTirePressureEngineeringValue(waterFilmDepthMm, 2),
		SpeedKmh:                         roundTirePressureEngineeringValue(speedKmh, 1),
		LateralDemandRatio:               roundTirePressureEngineeringValue(lateralDemandRatio, 3),
		FrictionRetentionRatio:           roundTirePressureEngineeringValue(retentionRatio, 3),
		ReferencePressurePsi:             roundTirePressureEngineeringValue(*referencePressurePsi, 1),
		EquivalentPressurePsi:            roundTirePressureEngineeringValue(equivalentPressurePsi, 1),
		PressureReductionPsi:             roundTirePressureEngineeringValue(pressureReductionPsi, 1),
		PressureReductionPct:             roundTirePressureEngineeringValue(pressureReductionPct, 1),
		ReferenceContactAreaCm2:          roundTirePressureEngineeringValue(*wheel.EstimatedStaticContactCm2, 2),
		EquivalentContactAreaCm2:         roundTirePressureEngineeringValue(equivalentAreaCm2, 2),
		ContactAreaChangePct:             roundTirePressureEngineeringValue((equivalentAreaCm2 / *wheel.EstimatedStaticContactCm2 - 1)*100, 1),
		WetGripLimitAtReferencePressureN: roundTirePressureEngineeringValue(wetGripAtReferencePressureN, 1),
		WetGripMarginPct:                 roundTirePressureEngineeringValue(calculateGripMarginPercentage(wetGripAtReferencePressureN, wheel.LateralDemandN), 1),
		PressureClampedToMinimum:         pressureClampedToMinimum,
		SurfaceTextureBaseline:           WetRoadDemoSurfaceTextureBaseline,
		RubberBaseline:                   WetRoadDemoRubberBaseline,
	}
	if minimumPressurePsi != nil {
		minimumPressure := roundTirePressureEngineeringValue(*minimumPressurePsi, 1)
		comparison.MinimumPressurePsi = &minimumPressure
	}
	wheel.WetPressureCompensation = comparison
	attachPressureContactAreaComparison(wheel, referencePressurePsi, &equivalentPressurePsi, tireWidthMm)
}

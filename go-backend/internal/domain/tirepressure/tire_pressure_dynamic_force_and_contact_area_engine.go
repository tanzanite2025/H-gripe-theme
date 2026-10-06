package tirepressure

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	ModelVersion                                    = "pressure-baseline-v1-unapproved"
	DynamicsModelVersion                            = "cornering-demo-v7-pressure-friction-status"
	PsiPerBar                                       = 14.5037738
	LoadToleranceKg                                 = 0.5
	GravityMps2                                     = 9.80665
	GenericTireBodyNormalizationBaseline            = 1.0
	GenericTireBodyNormalizationSourceFixedBaseline = "FIXED_GENERIC_BASELINE"
	FlatRoadDemoFrictionCoefficient                 = 0.76
	PressureFrictionDataStatus                      = "INSUFFICIENT_MEASUREMENT_SUPPORT"
	PressureFrictionCalculationMethod               = "fixed_flat_road_baseline_without_pressure_adjustment"
	TireWidthSourceMeasured                         = "MEASURED"
	TireWidthSourceNominalUncorrected               = "NOMINAL_UNCORRECTED"
)

var (
	ErrInvalidField      = errors.New("invalid field")
	ErrOutOfRange        = errors.New("field out of range")
	ErrLoadMismatch      = errors.New("front and rear loads do not match system mass")
	ErrLimitUnverified   = errors.New("pressure limit is unverified")
	ErrModelUncalibrated = errors.New("pressure model is not calibrated")
)

type RimSystem string

const (
	RimHookless RimSystem = "HOOKLESS"
	RimHooked   RimSystem = "HOOKED"
	RimUnknown  RimSystem = "UNKNOWN"
)

type RidingPosition string

const (
	PositionAggressiveRace RidingPosition = "AGGRESSIVE_RACE"
	PositionEndurance      RidingPosition = "ENDURANCE"
	PositionTTAero         RidingPosition = "TT_AERO"
	PositionGravelTrail    RidingPosition = "GRAVEL_TRAIL"
)

type SurfaceCondition string

const (
	SurfaceFlatRoad SurfaceCondition = "FLAT_ROAD"
)

type PressureLimitSource struct {
	Name      string  `json:"name"`
	Version   string  `json:"version"`
	AppliesTo string  `json:"applies_to"`
	LimitType string  `json:"limit_type"`
	LimitBar  float64 `json:"limit_bar"`
}

type Request struct {
	RiderWeightKg                   float64               `json:"rider_weight_kg"`
	BikeWeightKg                    float64               `json:"bike_weight_kg"`
	FrontLoadKg                     *float64              `json:"front_load_kg,omitempty"`
	RearLoadKg                      *float64              `json:"rear_load_kg,omitempty"`
	NominalTireWidthMm              float64               `json:"nominal_tire_width_mm"`
	MeasuredTireWidthMm             *float64              `json:"measured_tire_width_mm,omitempty"`
	InnerRimWidthMm                 float64               `json:"inner_rim_width_mm"`
	RimSystem                       RimSystem             `json:"rim_system"`
	TireMaxPressureBar              *float64              `json:"tire_max_pressure_bar,omitempty"`
	RimMaxPressureBar               *float64              `json:"rim_max_pressure_bar,omitempty"`
	WheelMaxPressureBar             *float64              `json:"wheel_max_pressure_bar,omitempty"`
	LimitSources                    []PressureLimitSource `json:"limit_sources,omitempty"`
	Position                        RidingPosition        `json:"riding_position"`
	Surface                         SurfaceCondition      `json:"surface_condition"`
	LeanAngleDeg                    *float64              `json:"lean_angle_deg,omitempty"`
	SpeedKmh                        float64               `json:"speed_kmh"`
	FrontOperatingPsi               *float64              `json:"front_operating_pressure_psi,omitempty"`
	RearOperatingPsi                *float64              `json:"rear_operating_pressure_psi,omitempty"`
	FrontComparisonPressurePsi      *float64              `json:"front_comparison_pressure_psi,omitempty"`
	RearComparisonPressurePsi       *float64              `json:"rear_comparison_pressure_psi,omitempty"`
	FrontMinimumPressurePsi         *float64              `json:"front_minimum_pressure_psi,omitempty"`
	RearMinimumPressurePsi          *float64              `json:"rear_minimum_pressure_psi,omitempty"`
	WetPressureDemonstrationEnabled bool                  `json:"wet_pressure_demonstration_enabled,omitempty"`
	WaterFilmDepthMm                *float64              `json:"water_film_depth_mm,omitempty"`
}

type ValidationError struct {
	Code   string
	Field  string
	Reason string
}

func (e *ValidationError) Error() string { return e.Reason }

type Loads struct {
	FrontKg float64
	RearKg  float64
	Source  string
}

// WheelDynamics is an intentionally explicit, ground-frame force decomposition.
// It is a demo estimate, not a tire safety or pressure recommendation.
type WheelDynamics struct {
	LoadKg                                       float64                                  `json:"load_kg"`
	VerticalLoadN                                float64                                  `json:"vertical_load_n"`
	LateralDemandN                               float64                                  `json:"lateral_demand_n"`
	ResultantContactForceN                       float64                                  `json:"resultant_contact_force_n"`
	EstimatedStaticContactCm2                    *float64                                 `json:"estimated_static_contact_area_cm2,omitempty"`
	EstimatedEquivalentCircularContactDiameterMm *float64                                 `json:"estimated_equivalent_circular_contact_diameter_mm,omitempty"`
	EstimatedContactPatchWidthMm                 *float64                                 `json:"estimated_contact_patch_width_mm,omitempty"`
	EstimatedContactPatchLengthMm                *float64                                 `json:"estimated_contact_patch_length_mm,omitempty"`
	VerticalDeformation                          *TirePressureVerticalDeformationEstimate `json:"vertical_deformation,omitempty"`
	MuNominal                                    float64                                  `json:"mu_nominal"`
	IdealizedGripLimitN                          float64                                  `json:"idealized_grip_limit_n"`
	GripMarginPct                                float64                                  `json:"grip_margin_pct"`
	PressureContactAreaComparison                *PressureContactAreaComparison           `json:"pressure_contact_area_comparison,omitempty"`
	PressureFrictionCoefficient                  *PressureFrictionCoefficientEstimate     `json:"pressure_friction_coefficient,omitempty"`
	WetPressureCompensation                      *WetPressureCompensation                 `json:"wet_pressure_compensation,omitempty"`
}

// PressureFrictionCoefficientEstimate makes the missing pressure-to-friction
// link explicit. A complete same-tire, same-load, same-surface and same-speed
// pressure sweep is not available, so the calculation retains the fixed flat
// road coefficient and reports that pressure adjustment was not applied.
type PressureFrictionCoefficientEstimate struct {
	DataStatus               string  `json:"data_status"`
	OperatingPressurePsi     float64 `json:"operating_pressure_psi"`
	NominalCoefficient       float64 `json:"nominal_coefficient"`
	EstimatedCoefficient     float64 `json:"estimated_coefficient"`
	RelativeCoefficientIndex float64 `json:"relative_coefficient_index"`
	PressureEffectApplied    bool    `json:"pressure_effect_applied"`
	CalculationMethod        string  `json:"calculation_method"`
}

// TirePressureVerticalDeformationEstimate is a transparent reference-tire
// estimate. It is intentionally separate from the selected Schwalbe product:
// the local dataset provides one measured-pressure reference and the pressure
// scaling used here is a first-order application assumption.
type TirePressureVerticalDeformationEstimate struct {
	DataStatus                       string  `json:"data_status"`
	DatasetVersion                   string  `json:"dataset_version"`
	ReferencePressureBar             float64 `json:"reference_pressure_bar"`
	OperatingPressurePsi             float64 `json:"operating_pressure_psi"`
	OperatingPressureBar             float64 `json:"operating_pressure_bar"`
	VerticalLoadN                    float64 `json:"vertical_load_n"`
	ReferenceDeflectionMm            float64 `json:"reference_deflection_mm"`
	EstimatedDeflectionMm            float64 `json:"estimated_deflection_mm"`
	RelativeDeformationIndex         float64 `json:"relative_deformation_index"`
	ReferenceVerticalStiffnessNPerMm float64 `json:"reference_vertical_stiffness_n_per_mm"`
	EstimatedVerticalStiffnessNPerMm float64 `json:"estimated_vertical_stiffness_n_per_mm"`
	CalculationMethod                string  `json:"calculation_method"`
	PressureScalingAssumption        string  `json:"pressure_scaling_assumption"`
}

// PressureContactAreaComparison reports the deterministic pressure-only
// comparison for the same wheel load. It deliberately does not infer a
// friction or grip change from the area difference.
type PressureContactAreaComparison struct {
	ReferencePressurePsi    float64  `json:"reference_pressure_psi"`
	ComparisonPressurePsi   float64  `json:"comparison_pressure_psi"`
	ReferenceAreaCm2        float64  `json:"reference_contact_area_cm2"`
	ComparisonAreaCm2       float64  `json:"comparison_contact_area_cm2"`
	AreaChangePct           float64  `json:"area_change_pct"`
	ReferencePatchWidthMm   *float64 `json:"reference_contact_patch_width_mm,omitempty"`
	ReferencePatchLengthMm  *float64 `json:"reference_contact_patch_length_mm,omitempty"`
	ComparisonPatchWidthMm  *float64 `json:"comparison_contact_patch_width_mm,omitempty"`
	ComparisonPatchLengthMm *float64 `json:"comparison_contact_patch_length_mm,omitempty"`
}

type Dynamics struct {
	ModelVersion                string           `json:"model_version"`
	ModelStatus                 string           `json:"model_status"`
	ForceFrame                  string           `json:"force_frame"`
	LeanAngleDeg                float64          `json:"lean_angle_deg"`
	SpeedKmh                    float64          `json:"speed_kmh"`
	EquivalentTurnRadiusM       *float64         `json:"equivalent_turn_radius_m,omitempty"`
	LateralAccelerationMps2     float64          `json:"lateral_acceleration_mps2"`
	LateralAccelerationG        float64          `json:"lateral_acceleration_g"`
	MuSource                    string           `json:"mu_source"`
	Surface                     SurfaceCondition `json:"surface_condition"`
	TireBodyNormalizationFactor float64          `json:"tire_body_normalization_factor"`
	TireBodyNormalizationSource string           `json:"tire_body_normalization_source"`
	TireWidthMm                 float64          `json:"tire_width_mm"`
	TireWidthSource             string           `json:"tire_width_source"`
	Front                       WheelDynamics    `json:"front"`
	Rear                        WheelDynamics    `json:"rear"`
}

func ValidateTirePressureDynamicCalculationRequest(req Request) error {
	checks := []struct {
		field    string
		value    float64
		min, max float64
	}{
		{"rider_weight_kg", req.RiderWeightKg, 35, 150},
		{"bike_weight_kg", req.BikeWeightKg, 4.5, 30},
		{"nominal_tire_width_mm", req.NominalTireWidthMm, 20, 80},
	}
	for _, check := range checks {
		if !isFiniteTirePressureCalculationNumber(check.value) {
			return &ValidationError{Code: "INVALID_FIELD", Field: check.field, Reason: check.field + " must be finite"}
		}
		if check.value < check.min || check.value > check.max {
			return &ValidationError{Code: "OUT_OF_RANGE", Field: check.field, Reason: fmt.Sprintf("%s must be between %.1f and %.1f", check.field, check.min, check.max)}
		}
	}
	// Inner-rim width remains an optional compatibility field for the older
	// metadata endpoint. The force demonstration intentionally does not use it:
	// its contact-patch width is limited by the supplied tire width only.
	if req.InnerRimWidthMm != 0 && (req.InnerRimWidthMm < 12 || req.InnerRimWidthMm > 40 || !isFiniteTirePressureCalculationNumber(req.InnerRimWidthMm)) {
		return &ValidationError{Code: "OUT_OF_RANGE", Field: "inner_rim_width_mm", Reason: "inner_rim_width_mm must be between 12.0 and 40.0 when provided"}
	}
	// Rim system is retained only for older metadata callers. The force demo
	// does not select a rim system; an omitted value is treated as unknown.
	if req.RimSystem != "" && req.RimSystem != RimHooked && req.RimSystem != RimHookless && req.RimSystem != RimUnknown {
		return &ValidationError{Code: "INVALID_FIELD", Field: "rim_system", Reason: "rim_system must be HOOKED, HOOKLESS, or UNKNOWN"}
	}
	if req.Position != PositionAggressiveRace && req.Position != PositionEndurance && req.Position != PositionTTAero && req.Position != PositionGravelTrail {
		return &ValidationError{Code: "INVALID_FIELD", Field: "riding_position", Reason: "riding_position is required and must be a supported enum"}
	}
	if req.Surface != "" && req.Surface != SurfaceFlatRoad {
		return &ValidationError{Code: "INVALID_FIELD", Field: "surface_condition", Reason: "surface_condition must be FLAT_ROAD"}
	}
	for field, value := range map[string]*float64{
		"measured_tire_width_mm":        req.MeasuredTireWidthMm,
		"tire_max_pressure_bar":         req.TireMaxPressureBar,
		"rim_max_pressure_bar":          req.RimMaxPressureBar,
		"wheel_max_pressure_bar":        req.WheelMaxPressureBar,
		"front_operating_pressure_psi":  req.FrontOperatingPsi,
		"rear_operating_pressure_psi":   req.RearOperatingPsi,
		"front_comparison_pressure_psi": req.FrontComparisonPressurePsi,
		"rear_comparison_pressure_psi":  req.RearComparisonPressurePsi,
		"front_minimum_pressure_psi":    req.FrontMinimumPressurePsi,
		"rear_minimum_pressure_psi":     req.RearMinimumPressurePsi,
	} {
		if value != nil && (!isFiniteTirePressureCalculationNumber(*value) || *value <= 0) {
			return &ValidationError{Code: "INVALID_FIELD", Field: field, Reason: field + " must be a positive finite number"}
		}
	}
	if req.LeanAngleDeg != nil && (!isFiniteTirePressureCalculationNumber(*req.LeanAngleDeg) || *req.LeanAngleDeg < 0 || *req.LeanAngleDeg > 60) {
		return &ValidationError{Code: "OUT_OF_RANGE", Field: "lean_angle_deg", Reason: "lean_angle_deg must be between 0 and 60 degrees"}
	}
	if !isFiniteTirePressureCalculationNumber(req.SpeedKmh) || req.SpeedKmh < 0 || req.SpeedKmh > 120 {
		return &ValidationError{Code: "OUT_OF_RANGE", Field: "speed_kmh", Reason: "speed_kmh must be between 0 and 120 km/h"}
	}
	if (req.FrontOperatingPsi == nil) != (req.RearOperatingPsi == nil) {
		return &ValidationError{Code: "INVALID_FIELD", Field: "front_operating_pressure_psi/rear_operating_pressure_psi", Reason: "both operating pressures must be supplied together"}
	}
	if (req.FrontComparisonPressurePsi == nil) != (req.RearComparisonPressurePsi == nil) {
		return &ValidationError{Code: "INVALID_FIELD", Field: "front_comparison_pressure_psi/rear_comparison_pressure_psi", Reason: "both comparison pressures must be supplied together"}
	}
	if req.FrontComparisonPressurePsi != nil && req.FrontOperatingPsi == nil {
		return &ValidationError{Code: "INVALID_FIELD", Field: "front_comparison_pressure_psi/rear_comparison_pressure_psi", Reason: "comparison pressures require both operating pressures"}
	}
	if (req.FrontMinimumPressurePsi == nil) != (req.RearMinimumPressurePsi == nil) {
		return &ValidationError{Code: "INVALID_FIELD", Field: "front_minimum_pressure_psi/rear_minimum_pressure_psi", Reason: "both minimum pressures must be supplied together"}
	}
	if req.WetPressureDemonstrationEnabled {
		if req.FrontOperatingPsi == nil || req.RearOperatingPsi == nil {
			return &ValidationError{Code: "INVALID_FIELD", Field: "front_operating_pressure_psi/rear_operating_pressure_psi", Reason: "wet pressure demonstration requires both operating pressures"}
		}
		if req.FrontMinimumPressurePsi == nil || req.RearMinimumPressurePsi == nil {
			return &ValidationError{Code: "INVALID_FIELD", Field: "front_minimum_pressure_psi/rear_minimum_pressure_psi", Reason: "wet pressure demonstration requires both minimum pressures"}
		}
	}
	if req.FrontMinimumPressurePsi != nil && req.FrontOperatingPsi != nil {
		if *req.FrontMinimumPressurePsi > *req.FrontOperatingPsi || *req.RearMinimumPressurePsi > *req.RearOperatingPsi {
			return &ValidationError{Code: "OUT_OF_RANGE", Field: "front_minimum_pressure_psi/rear_minimum_pressure_psi", Reason: "minimum pressure must not exceed the operating pressure"}
		}
	}
	if req.WaterFilmDepthMm != nil && (!isFiniteTirePressureCalculationNumber(*req.WaterFilmDepthMm) || *req.WaterFilmDepthMm < 0 || *req.WaterFilmDepthMm > WetRoadDemoMaximumWaterFilmDepthMm) {
		return &ValidationError{Code: "OUT_OF_RANGE", Field: "water_film_depth_mm", Reason: fmt.Sprintf("water_film_depth_mm must be between 0.0 and %.1f", WetRoadDemoMaximumWaterFilmDepthMm)}
	}
	seenLimitTypes := make(map[string]struct{}, len(req.LimitSources))
	for i, source := range req.LimitSources {
		if strings.TrimSpace(source.Name) == "" || strings.TrimSpace(source.Version) == "" || strings.TrimSpace(source.AppliesTo) == "" || !validLimitType(source.LimitType) || !isFiniteTirePressureCalculationNumber(source.LimitBar) || source.LimitBar <= 0 {
			return &ValidationError{Code: "INVALID_FIELD", Field: fmt.Sprintf("limit_sources[%d]", i), Reason: "limit source requires name, version, applies_to, supported limit_type, and positive limit_bar"}
		}
		if _, exists := seenLimitTypes[source.LimitType]; exists {
			return &ValidationError{Code: "INVALID_FIELD", Field: fmt.Sprintf("limit_sources[%d].limit_type", i), Reason: "limit_type must not be duplicated"}
		}
		seenLimitTypes[source.LimitType] = struct{}{}
	}
	for _, provided := range []struct {
		limitType string
		value     *float64
	}{
		{limitType: "TIRE", value: req.TireMaxPressureBar},
		{limitType: "RIM", value: req.RimMaxPressureBar},
		{limitType: "WHEEL", value: req.WheelMaxPressureBar},
	} {
		if provided.value == nil {
			continue
		}
		if _, exists := seenLimitTypes[provided.limitType]; !exists {
			return &ValidationError{Code: "INVALID_FIELD", Field: "limit_sources", Reason: "each provided pressure limit must have a matching source record"}
		}
	}
	return nil
}

func validLimitType(value string) bool {
	return value == "TIRE" || value == "RIM" || value == "WHEEL" || value == "SYSTEM"
}

func ResolveTirePressureWheelLoads(req Request) (Loads, error) {
	if err := ValidateTirePressureDynamicCalculationRequest(req); err != nil {
		return Loads{}, err
	}
	total := req.RiderWeightKg + req.BikeWeightKg
	if req.FrontLoadKg != nil || req.RearLoadKg != nil {
		if req.FrontLoadKg == nil || req.RearLoadKg == nil || !isFinitePositiveTirePressureCalculationNumber(*req.FrontLoadKg) || !isFinitePositiveTirePressureCalculationNumber(*req.RearLoadKg) {
			return Loads{}, &ValidationError{Code: "INVALID_FIELD", Field: "front_load_kg/rear_load_kg", Reason: "both front and rear loads must be positive when either is supplied"}
		}
		if *req.FrontLoadKg > total || *req.RearLoadKg > total {
			return Loads{}, &ValidationError{Code: "OUT_OF_RANGE", Field: "front_load_kg/rear_load_kg", Reason: "each wheel load must not exceed total system mass"}
		}
		if math.Abs((*req.FrontLoadKg+*req.RearLoadKg)-total) > LoadToleranceKg {
			return Loads{}, ErrLoadMismatch
		}
		return Loads{FrontKg: *req.FrontLoadKg, RearKg: *req.RearLoadKg, Source: "MEASURED"}, nil
	}
	ratio := map[RidingPosition]float64{PositionAggressiveRace: .47, PositionEndurance: .45, PositionTTAero: .48, PositionGravelTrail: .44}
	front, ok := ratio[req.Position]
	if !ok {
		return Loads{}, &ValidationError{Code: "INVALID_FIELD", Field: "riding_position", Reason: "unknown riding_position"}
	}
	return Loads{FrontKg: total * front, RearKg: total * (1 - front), Source: "ESTIMATED"}, nil
}

func isFinitePositiveTirePressureCalculationNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}

func isFiniteTirePressureCalculationNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func CalculateEffectiveProvidedTirePressureLimitBar(req Request) (float64, string) {
	limits := make([]float64, 0, len(req.LimitSources))
	for _, source := range req.LimitSources {
		limits = append(limits, source.LimitBar)
	}
	if len(limits) == 0 {
		return 0, "LIMITS_NOT_PROVIDED"
	}
	min := limits[0]
	for _, value := range limits[1:] {
		if value < min {
			min = value
		}
	}
	return min, "PROVIDED_UNVERIFIED"
}

// CalculateTirePressureGroundFrameCorneringDynamics evaluates the transparent
// first-order force demonstration. The road is fixed to a flat-road baseline;
// no weather or road-condition selector is part of this model. Operating
// pressure affects the static contact-area and width-limited footprint
// estimates. Speed is paired with lean to report the equivalent corner radius:
// R = v² / (g tan(lean)). At a fixed lean angle, speed changes the radius while
// the required lateral acceleration remains g tan(lean). When explicitly
// enabled, the wet demonstration adds a bounded water-film proxy and derives
// an equivalent pressure for the same reference contact area; it does not
// claim that lowering pressure restores the lost grip or infer a recommendation.
func CalculateTirePressureGroundFrameCorneringDynamics(req Request, loads Loads) (Dynamics, error) {
	if err := ValidateTirePressureDynamicCalculationRequest(req); err != nil {
		return Dynamics{}, err
	}
	angle := 0.0
	if req.LeanAngleDeg != nil {
		angle = *req.LeanAngleDeg
	}
	speedKmh := req.SpeedKmh
	speedMps := speedKmh / 3.6
	lateralAcceleration := GravityMps2 * math.Tan(angle*math.Pi/180)
	var equivalentTurnRadiusM *float64
	if speedMps > 0 && lateralAcceleration > 0 {
		radius := speedMps * speedMps / lateralAcceleration
		radius = roundTirePressureEngineeringValue(radius, 2)
		equivalentTurnRadiusM = &radius
	}
	muNominal := FlatRoadDemoFrictionCoefficient
	tireWidthMm, tireWidthSource := resolveTireWidthForDisplay(req)
	front := CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(loads.FrontKg, angle, muNominal, req.FrontOperatingPsi, tireWidthMm)
	rear := CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(loads.RearKg, angle, muNominal, req.RearOperatingPsi, tireWidthMm)
	if req.WetPressureDemonstrationEnabled {
		waterFilmDepthMm := resolveWetRoadWaterFilmDepthMm(req.WaterFilmDepthMm)
		attachWetPressureCompensation(&front, req.FrontOperatingPsi, req.FrontMinimumPressurePsi, speedKmh, waterFilmDepthMm, tireWidthMm)
		attachWetPressureCompensation(&rear, req.RearOperatingPsi, req.RearMinimumPressurePsi, speedKmh, waterFilmDepthMm, tireWidthMm)
	} else {
		attachPressureContactAreaComparison(&front, req.FrontOperatingPsi, req.FrontComparisonPressurePsi, tireWidthMm)
		attachPressureContactAreaComparison(&rear, req.RearOperatingPsi, req.RearComparisonPressurePsi, tireWidthMm)
	}
	return Dynamics{
		ModelVersion:                DynamicsModelVersion,
		ModelStatus:                 "DEMO_ESTIMATE_UNCALIBRATED",
		ForceFrame:                  "GROUND",
		LeanAngleDeg:                roundTirePressureEngineeringValue(angle, 1),
		SpeedKmh:                    roundTirePressureEngineeringValue(speedKmh, 1),
		EquivalentTurnRadiusM:       equivalentTurnRadiusM,
		LateralAccelerationMps2:     roundTirePressureEngineeringValue(lateralAcceleration, 2),
		LateralAccelerationG:        roundTirePressureEngineeringValue(lateralAcceleration/GravityMps2, 3),
		MuSource:                    "flat_road_demo_baseline",
		Surface:                     SurfaceFlatRoad,
		TireBodyNormalizationFactor: roundTirePressureEngineeringValue(GenericTireBodyNormalizationBaseline, 3),
		TireBodyNormalizationSource: GenericTireBodyNormalizationSourceFixedBaseline,
		TireWidthMm:                 roundTirePressureEngineeringValue(tireWidthMm, 1),
		TireWidthSource:             tireWidthSource,
		Front:                       front,
		Rear:                        rear,
	}, nil
}

func attachPressureContactAreaComparison(wheel *WheelDynamics, referencePressurePsi, comparisonPressurePsi *float64, tireWidthMm float64) {
	if wheel == nil || wheel.EstimatedStaticContactCm2 == nil || referencePressurePsi == nil || comparisonPressurePsi == nil {
		return
	}
	if !isFinitePositiveTirePressureCalculationNumber(*referencePressurePsi) || !isFinitePositiveTirePressureCalculationNumber(*comparisonPressurePsi) {
		return
	}
	comparisonAreaCm2 := *wheel.EstimatedStaticContactCm2 * *referencePressurePsi / *comparisonPressurePsi
	if !isFinitePositiveTirePressureCalculationNumber(comparisonAreaCm2) {
		return
	}
	areaChangePct := (comparisonAreaCm2 / *wheel.EstimatedStaticContactCm2 - 1) * 100
	if !isFiniteTirePressureCalculationNumber(areaChangePct) {
		return
	}
	comparison := &PressureContactAreaComparison{
		ReferencePressurePsi:  roundTirePressureEngineeringValue(*referencePressurePsi, 1),
		ComparisonPressurePsi: roundTirePressureEngineeringValue(*comparisonPressurePsi, 1),
		ReferenceAreaCm2:      roundTirePressureEngineeringValue(*wheel.EstimatedStaticContactCm2, 2),
		ComparisonAreaCm2:     roundTirePressureEngineeringValue(comparisonAreaCm2, 2),
		AreaChangePct:         roundTirePressureEngineeringValue(areaChangePct, 1),
	}
	comparison.ReferencePatchWidthMm = wheel.EstimatedContactPatchWidthMm
	comparison.ReferencePatchLengthMm = wheel.EstimatedContactPatchLengthMm
	if tireWidthMm > 0 {
		_, comparisonWidthMm, comparisonLengthMm, ok := CalculateWidthLimitedEquivalentCircularContactPatch(comparisonAreaCm2, tireWidthMm)
		if ok {
			roundedComparisonWidthMm := roundTirePressureEngineeringValue(comparisonWidthMm, 1)
			roundedComparisonLengthMm := roundTirePressureEngineeringValue(comparisonLengthMm, 1)
			comparison.ComparisonPatchWidthMm = &roundedComparisonWidthMm
			comparison.ComparisonPatchLengthMm = &roundedComparisonLengthMm
		}
	}
	wheel.PressureContactAreaComparison = comparison
}

func CalculateSingleWheelGroundFrameCorneringDynamics(loadKg, angleDeg, muNominal float64, pressurePsi *float64) WheelDynamics {
	return CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(loadKg, angleDeg, muNominal, pressurePsi, 0)
}

// CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth adds a
// width-limited equivalent circular footprint and a generic reference-tire
// vertical-deformation estimate to the transparent force model. The footprint
// dimensions are geometric estimates only, and the deformation estimate is
// not a selected Schwalbe tire measurement.
func CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(loadKg, angleDeg, muNominal float64, pressurePsi *float64, tireWidthMm float64) WheelDynamics {
	fz := loadKg * GravityMps2
	fy := fz * math.Tan(angleDeg*math.Pi/180)
	resultant := math.Hypot(fz, fy)
	grip := muNominal * fz * GenericTireBodyNormalizationBaseline
	margin := calculateGripMarginPercentage(grip, fy)
	result := WheelDynamics{
		LoadKg:                 roundTirePressureEngineeringValue(loadKg, 1),
		VerticalLoadN:          roundTirePressureEngineeringValue(fz, 1),
		LateralDemandN:         roundTirePressureEngineeringValue(fy, 1),
		ResultantContactForceN: roundTirePressureEngineeringValue(resultant, 1),
		MuNominal:              roundTirePressureEngineeringValue(muNominal, 3),
		IdealizedGripLimitN:    roundTirePressureEngineeringValue(grip, 1),
		GripMarginPct:          roundTirePressureEngineeringValue(margin, 1),
	}
	if pressurePsi != nil && *pressurePsi > 0 {
		pressurePa := *pressurePsi * 6894.757293
		areaCm2 := fz / pressurePa * 10000
		if isFiniteTirePressureCalculationNumber(areaCm2) && areaCm2 > 0 {
			roundedAreaCm2 := roundTirePressureEngineeringValue(areaCm2, 2)
			result.EstimatedStaticContactCm2 = &roundedAreaCm2
			if tireWidthMm > 0 {
				circularDiameterMm, patchWidthMm, patchLengthMm, ok := CalculateWidthLimitedEquivalentCircularContactPatch(areaCm2, tireWidthMm)
				if ok {
					roundedCircularDiameterMm := roundTirePressureEngineeringValue(circularDiameterMm, 1)
					roundedPatchWidthMm := roundTirePressureEngineeringValue(patchWidthMm, 1)
					roundedPatchLengthMm := roundTirePressureEngineeringValue(patchLengthMm, 1)
					result.EstimatedEquivalentCircularContactDiameterMm = &roundedCircularDiameterMm
					result.EstimatedContactPatchWidthMm = &roundedPatchWidthMm
					result.EstimatedContactPatchLengthMm = &roundedPatchLengthMm
				}
			}
		}
	}
	result.VerticalDeformation = CalculateTirePressureVerticalDeformationEstimate(fz, pressurePsi)
	result.PressureFrictionCoefficient = CalculateTirePressureFrictionCoefficientEstimate(muNominal, pressurePsi)
	return result
}

// CalculateTirePressureFrictionCoefficientEstimate returns the fixed friction
// baseline together with its data status. It intentionally does not invent a
// pressure-dependent coefficient until a complete matched measurement set is
// available for the same tire, load, surface, and speed.
func CalculateTirePressureFrictionCoefficientEstimate(muNominal float64, operatingPressurePsi *float64) *PressureFrictionCoefficientEstimate {
	if operatingPressurePsi == nil || !isFinitePositiveTirePressureCalculationNumber(*operatingPressurePsi) || !isFinitePositiveTirePressureCalculationNumber(muNominal) {
		return nil
	}

	return &PressureFrictionCoefficientEstimate{
		DataStatus:               PressureFrictionDataStatus,
		OperatingPressurePsi:     roundTirePressureEngineeringValue(*operatingPressurePsi, 1),
		NominalCoefficient:       roundTirePressureEngineeringValue(muNominal, 3),
		EstimatedCoefficient:     roundTirePressureEngineeringValue(muNominal, 3),
		RelativeCoefficientIndex: 1,
		PressureEffectApplied:    false,
		CalculationMethod:        PressureFrictionCalculationMethod,
	}
}

// CalculateTirePressureVerticalDeformationEstimate interpolates the local
// 4.75-bar reference points and scales the result by reference pressure over
// operating pressure. The scaling is a clearly labelled first-order assumption
// because the source paper does not publish a complete pressure sweep.
func CalculateTirePressureVerticalDeformationEstimate(verticalLoadN float64, operatingPressurePsi *float64) *TirePressureVerticalDeformationEstimate {
	if operatingPressurePsi == nil || !isFinitePositiveTirePressureCalculationNumber(verticalLoadN) || !isFinitePositiveTirePressureCalculationNumber(*operatingPressurePsi) {
		return nil
	}

	referenceData := GetTirePressureVerticalDeformationReferenceData()
	referencePressureBar := referenceData.VerticalStiffnessReference.InflationPressureBar
	referenceStiffnessNPerMm := referenceData.VerticalStiffnessReference.FittedVerticalStiffnessNPerMm
	operatingPressureBar := *operatingPressurePsi / PsiPerBar
	if !isFinitePositiveTirePressureCalculationNumber(referencePressureBar) || !isFinitePositiveTirePressureCalculationNumber(referenceStiffnessNPerMm) || !isFinitePositiveTirePressureCalculationNumber(operatingPressureBar) {
		return nil
	}

	referenceDeflectionMm := interpolateTirePressureReferenceDeflectionMm(verticalLoadN, referenceData.VerticalForceDeflectionPoints, referenceStiffnessNPerMm)
	if !isFinitePositiveTirePressureCalculationNumber(referenceDeflectionMm) {
		return nil
	}
	deformationIndex := referencePressureBar / operatingPressureBar
	estimatedDeflectionMm := referenceDeflectionMm * deformationIndex
	estimatedStiffnessNPerMm := referenceStiffnessNPerMm / deformationIndex
	if !isFinitePositiveTirePressureCalculationNumber(deformationIndex) || !isFinitePositiveTirePressureCalculationNumber(estimatedDeflectionMm) || !isFinitePositiveTirePressureCalculationNumber(estimatedStiffnessNPerMm) {
		return nil
	}

	return &TirePressureVerticalDeformationEstimate{
		DataStatus:                       referenceData.DataStatus,
		DatasetVersion:                   referenceData.DatasetVersion,
		ReferencePressureBar:             roundTirePressureEngineeringValue(referencePressureBar, 2),
		OperatingPressurePsi:             roundTirePressureEngineeringValue(*operatingPressurePsi, 1),
		OperatingPressureBar:             roundTirePressureEngineeringValue(operatingPressureBar, 3),
		VerticalLoadN:                    roundTirePressureEngineeringValue(verticalLoadN, 1),
		ReferenceDeflectionMm:            roundTirePressureEngineeringValue(referenceDeflectionMm, 2),
		EstimatedDeflectionMm:            roundTirePressureEngineeringValue(estimatedDeflectionMm, 2),
		RelativeDeformationIndex:         roundTirePressureEngineeringValue(deformationIndex, 3),
		ReferenceVerticalStiffnessNPerMm: roundTirePressureEngineeringValue(referenceStiffnessNPerMm, 1),
		EstimatedVerticalStiffnessNPerMm: roundTirePressureEngineeringValue(estimatedStiffnessNPerMm, 1),
		CalculationMethod:                "piecewise_linear_reference_points_with_zero_load_origin",
		PressureScalingAssumption:        "estimated_deflection_scales with reference_pressure_bar / operating_pressure_bar; reference stiffness scales inversely",
	}
}

func interpolateTirePressureReferenceDeflectionMm(verticalLoadN float64, points []TirePressureVerticalForceDeflectionPoint, fallbackStiffnessNPerMm float64) float64 {
	if !isFinitePositiveTirePressureCalculationNumber(verticalLoadN) {
		return 0
	}
	if !isFinitePositiveTirePressureCalculationNumber(fallbackStiffnessNPerMm) {
		return 0
	}
	if len(points) == 0 {
		return verticalLoadN / fallbackStiffnessNPerMm
	}

	first := points[0]
	if verticalLoadN <= first.VerticalLoadN {
		if !isFinitePositiveTirePressureCalculationNumber(first.VerticalLoadN) || !isFinitePositiveTirePressureCalculationNumber(first.DeflectionMm) {
			return verticalLoadN / fallbackStiffnessNPerMm
		}
		// The zero-load origin is a transparent geometric anchor, not a
		// measured point in the paper.
		return verticalLoadN * first.DeflectionMm / first.VerticalLoadN
	}

	for index := 1; index < len(points); index++ {
		previous := points[index-1]
		current := points[index]
		if verticalLoadN > current.VerticalLoadN {
			continue
		}
		loadSpan := current.VerticalLoadN - previous.VerticalLoadN
		if loadSpan <= 0 {
			return verticalLoadN / fallbackStiffnessNPerMm
		}
		loadRatio := (verticalLoadN - previous.VerticalLoadN) / loadSpan
		return previous.DeflectionMm + loadRatio*(current.DeflectionMm-previous.DeflectionMm)
	}

	if len(points) == 1 {
		last := points[0]
		if !isFinitePositiveTirePressureCalculationNumber(last.VerticalLoadN) || !isFinitePositiveTirePressureCalculationNumber(last.DeflectionMm) {
			return verticalLoadN / fallbackStiffnessNPerMm
		}
		return verticalLoadN * last.DeflectionMm / last.VerticalLoadN
	}

	last := points[len(points)-1]
	previous := points[len(points)-2]
	loadSpan := last.VerticalLoadN - previous.VerticalLoadN
	if loadSpan <= 0 {
		return verticalLoadN / fallbackStiffnessNPerMm
	}
	return last.DeflectionMm + (verticalLoadN-last.VerticalLoadN)*(last.DeflectionMm-previous.DeflectionMm)/loadSpan
}

func calculateGripMarginPercentage(gripLimitN, lateralDemandN float64) float64 {
	if gripLimitN <= 0 {
		return 0
	}
	return (gripLimitN - math.Abs(lateralDemandN)) / gripLimitN * 100
}

// CalculateWidthLimitedEquivalentCircularContactPatch converts the estimated
// load-bearing area into an equivalent circular diameter, then limits the
// footprint width to the supplied tire width. The remaining area becomes the
// estimated longitudinal patch length. Inputs and outputs use cm² and mm.
func CalculateWidthLimitedEquivalentCircularContactPatch(areaCm2, tireWidthMm float64) (circularDiameterMm, patchWidthMm, patchLengthMm float64, ok bool) {
	if !isFinitePositiveTirePressureCalculationNumber(areaCm2) || !isFinitePositiveTirePressureCalculationNumber(tireWidthMm) {
		return 0, 0, 0, false
	}
	areaMm2 := areaCm2 * 100
	circularDiameterMm = 2 * math.Sqrt(areaMm2/math.Pi)
	patchWidthMm = math.Min(tireWidthMm, circularDiameterMm)
	if patchWidthMm <= 0 || !isFinitePositiveTirePressureCalculationNumber(patchWidthMm) {
		return 0, 0, 0, false
	}
	patchLengthMm = areaMm2 / patchWidthMm
	if !isFinitePositiveTirePressureCalculationNumber(patchLengthMm) {
		return 0, 0, 0, false
	}
	return circularDiameterMm, patchWidthMm, patchLengthMm, true
}

func resolveTireWidthForDisplay(req Request) (float64, string) {
	if req.MeasuredTireWidthMm != nil {
		return *req.MeasuredTireWidthMm, TireWidthSourceMeasured
	}
	// A nominal width is shown as a reference only. The 0.40 rim-width
	// expansion coefficient is intentionally not applied without calibration.
	return req.NominalTireWidthMm, TireWidthSourceNominalUncorrected
}

func roundTirePressureEngineeringValue(value float64, decimals int) float64 {
	scale := math.Pow10(decimals)
	return math.Round(value*scale) / scale
}

package tirepressure

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	ModelVersion                                    = "pressure-baseline-v1-unapproved"
	DynamicsModelVersion                            = "cornering-demo-v3-speed-radius-and-contact-shape"
	PsiPerBar                                       = 14.5037738
	LoadToleranceKg                                 = 0.5
	GravityMps2                                     = 9.80665
	GenericTireBodyNormalizationBaseline            = 1.0
	GenericTireBodyNormalizationSourceFixedBaseline = "FIXED_GENERIC_BASELINE"
	FlatRoadDemoFrictionCoefficient                 = 0.76
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
	RiderWeightKg       float64               `json:"rider_weight_kg"`
	BikeWeightKg        float64               `json:"bike_weight_kg"`
	FrontLoadKg         *float64              `json:"front_load_kg,omitempty"`
	RearLoadKg          *float64              `json:"rear_load_kg,omitempty"`
	NominalTireWidthMm  float64               `json:"nominal_tire_width_mm"`
	MeasuredTireWidthMm *float64              `json:"measured_tire_width_mm,omitempty"`
	InnerRimWidthMm     float64               `json:"inner_rim_width_mm"`
	RimSystem           RimSystem             `json:"rim_system"`
	TireMaxPressureBar  *float64              `json:"tire_max_pressure_bar,omitempty"`
	RimMaxPressureBar   *float64              `json:"rim_max_pressure_bar,omitempty"`
	WheelMaxPressureBar *float64              `json:"wheel_max_pressure_bar,omitempty"`
	LimitSources        []PressureLimitSource `json:"limit_sources,omitempty"`
	Position            RidingPosition        `json:"riding_position"`
	Surface             SurfaceCondition      `json:"surface_condition"`
	LeanAngleDeg        *float64              `json:"lean_angle_deg,omitempty"`
	SpeedKmh            float64               `json:"speed_kmh"`
	FrontOperatingPsi   *float64              `json:"front_operating_pressure_psi,omitempty"`
	RearOperatingPsi    *float64              `json:"rear_operating_pressure_psi,omitempty"`
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
	LoadKg                                       float64  `json:"load_kg"`
	VerticalLoadN                                float64  `json:"vertical_load_n"`
	LateralDemandN                               float64  `json:"lateral_demand_n"`
	ResultantContactForceN                       float64  `json:"resultant_contact_force_n"`
	EstimatedStaticContactCm2                    *float64 `json:"estimated_static_contact_area_cm2,omitempty"`
	EstimatedEquivalentCircularContactDiameterMm *float64 `json:"estimated_equivalent_circular_contact_diameter_mm,omitempty"`
	EstimatedContactPatchWidthMm                 *float64 `json:"estimated_contact_patch_width_mm,omitempty"`
	EstimatedContactPatchLengthMm                *float64 `json:"estimated_contact_patch_length_mm,omitempty"`
	MuNominal                                    float64  `json:"mu_nominal"`
	IdealizedGripLimitN                          float64  `json:"idealized_grip_limit_n"`
	GripMarginPct                                float64  `json:"grip_margin_pct"`
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
		"measured_tire_width_mm":       req.MeasuredTireWidthMm,
		"tire_max_pressure_bar":        req.TireMaxPressureBar,
		"rim_max_pressure_bar":         req.RimMaxPressureBar,
		"wheel_max_pressure_bar":       req.WheelMaxPressureBar,
		"front_operating_pressure_psi": req.FrontOperatingPsi,
		"rear_operating_pressure_psi":  req.RearOperatingPsi,
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
// the required lateral acceleration remains g tan(lean). It deliberately does
// not infer a pressure recommendation.
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

func CalculateSingleWheelGroundFrameCorneringDynamics(loadKg, angleDeg, muNominal float64, pressurePsi *float64) WheelDynamics {
	return CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(loadKg, angleDeg, muNominal, pressurePsi, 0)
}

// CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth adds a
// width-limited equivalent circular footprint to the transparent force model.
// The footprint dimensions are geometric estimates only; the generic tire-body
// baseline is fixed at 1.0 and no vertical stiffness or displacement is inferred.
func CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(loadKg, angleDeg, muNominal float64, pressurePsi *float64, tireWidthMm float64) WheelDynamics {
	fz := loadKg * GravityMps2
	fy := fz * math.Tan(angleDeg*math.Pi/180)
	resultant := math.Hypot(fz, fy)
	grip := muNominal * fz * GenericTireBodyNormalizationBaseline
	margin := 0.0
	if grip > 0 {
		margin = (grip - math.Abs(fy)) / grip * 100
	}
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
	return result
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

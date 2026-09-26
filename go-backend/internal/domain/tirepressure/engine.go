package tirepressure

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	ModelVersion         = "pressure-baseline-v1-unapproved"
	DynamicsModelVersion = "cornering-demo-v1"
	PsiPerBar            = 14.5037738
	LoadToleranceKg      = 0.5
	GravityMps2          = 9.80665
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
type CasingType string
type WeatherCondition string

const (
	SurfaceSmoothTrack   SurfaceCondition = "SMOOTH_TRACK"
	SurfaceRoughChip     SurfaceCondition = "ROUGH_CHIP"
	SurfaceCobbles       SurfaceCondition = "COBBLES"
	SurfaceUnpavedGravel SurfaceCondition = "UNPAVED_GRAVEL"
	CasingTubeless       CasingType       = "TUBELESS"
	CasingTPUTube        CasingType       = "TPU_TUBE"
	CasingButylTube      CasingType       = "BUTYL_TUBE"
	WeatherDry           WeatherCondition = "DRY"
	WeatherWet           WeatherCondition = "WET"
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
	Casing              CasingType            `json:"casing_type"`
	Weather             WeatherCondition      `json:"weather_condition"`
	LeanAngleDeg        *float64              `json:"lean_angle_deg,omitempty"`
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
	LoadKg                    float64  `json:"load_kg"`
	VerticalLoadN             float64  `json:"vertical_load_n"`
	LateralDemandN            float64  `json:"lateral_demand_n"`
	ResultantContactForceN    float64  `json:"resultant_contact_force_n"`
	EstimatedStaticContactCm2 *float64 `json:"estimated_static_contact_area_cm2,omitempty"`
	MuNominal                 float64  `json:"mu_nominal"`
	MuLow                     float64  `json:"mu_low"`
	MuHigh                    float64  `json:"mu_high"`
	IdealizedGripLimitN       float64  `json:"idealized_grip_limit_n"`
	GripMarginPct             float64  `json:"grip_margin_pct"`
}

type Dynamics struct {
	ModelVersion string        `json:"model_version"`
	ModelStatus  string        `json:"model_status"`
	ForceFrame   string        `json:"force_frame"`
	LeanAngleDeg float64       `json:"lean_angle_deg"`
	MuSource     string        `json:"mu_source"`
	Front        WheelDynamics `json:"front"`
	Rear         WheelDynamics `json:"rear"`
}

func Validate(req Request) error {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	checks := []struct {
		field    string
		value    float64
		min, max float64
	}{
		{"rider_weight_kg", req.RiderWeightKg, 35, 150},
		{"bike_weight_kg", req.BikeWeightKg, 4.5, 30},
		{"nominal_tire_width_mm", req.NominalTireWidthMm, 20, 80},
		{"inner_rim_width_mm", req.InnerRimWidthMm, 12, 40},
	}
	for _, check := range checks {
		if !finite(check.value) {
			return &ValidationError{Code: "INVALID_FIELD", Field: check.field, Reason: check.field + " must be finite"}
		}
		if check.value < check.min || check.value > check.max {
			return &ValidationError{Code: "OUT_OF_RANGE", Field: check.field, Reason: fmt.Sprintf("%s must be between %.1f and %.1f", check.field, check.min, check.max)}
		}
	}
	if req.RimSystem != RimHooked && req.RimSystem != RimHookless && req.RimSystem != RimUnknown {
		return &ValidationError{Code: "INVALID_FIELD", Field: "rim_system", Reason: "rim_system must be HOOKED, HOOKLESS, or UNKNOWN"}
	}
	if req.Position != PositionAggressiveRace && req.Position != PositionEndurance && req.Position != PositionTTAero && req.Position != PositionGravelTrail {
		return &ValidationError{Code: "INVALID_FIELD", Field: "riding_position", Reason: "riding_position is required and must be a supported enum"}
	}
	if req.Surface != "" && req.Surface != SurfaceSmoothTrack && req.Surface != SurfaceRoughChip && req.Surface != SurfaceCobbles && req.Surface != SurfaceUnpavedGravel {
		return &ValidationError{Code: "INVALID_FIELD", Field: "surface_condition", Reason: "surface_condition is required and must be a supported enum"}
	}
	if req.Casing != "" && req.Casing != CasingTubeless && req.Casing != CasingTPUTube && req.Casing != CasingButylTube {
		return &ValidationError{Code: "INVALID_FIELD", Field: "casing_type", Reason: "casing_type is required and must be a supported enum"}
	}
	if req.Weather != "" && req.Weather != WeatherDry && req.Weather != WeatherWet {
		return &ValidationError{Code: "INVALID_FIELD", Field: "weather_condition", Reason: "weather_condition is required and must be a supported enum"}
	}
	for field, value := range map[string]*float64{
		"measured_tire_width_mm":       req.MeasuredTireWidthMm,
		"tire_max_pressure_bar":        req.TireMaxPressureBar,
		"rim_max_pressure_bar":         req.RimMaxPressureBar,
		"wheel_max_pressure_bar":       req.WheelMaxPressureBar,
		"front_operating_pressure_psi": req.FrontOperatingPsi,
		"rear_operating_pressure_psi":  req.RearOperatingPsi,
	} {
		if value != nil && (!finite(*value) || *value <= 0) {
			return &ValidationError{Code: "INVALID_FIELD", Field: field, Reason: field + " must be a positive finite number"}
		}
	}
	if req.LeanAngleDeg != nil && (!finite(*req.LeanAngleDeg) || *req.LeanAngleDeg < 0 || *req.LeanAngleDeg > 60) {
		return &ValidationError{Code: "OUT_OF_RANGE", Field: "lean_angle_deg", Reason: "lean_angle_deg must be between 0 and 60 degrees"}
	}
	if (req.FrontOperatingPsi == nil) != (req.RearOperatingPsi == nil) {
		return &ValidationError{Code: "INVALID_FIELD", Field: "front_operating_pressure_psi/rear_operating_pressure_psi", Reason: "both operating pressures must be supplied together"}
	}
	seenLimitTypes := make(map[string]struct{}, len(req.LimitSources))
	for i, source := range req.LimitSources {
		if strings.TrimSpace(source.Name) == "" || strings.TrimSpace(source.Version) == "" || strings.TrimSpace(source.AppliesTo) == "" || !validLimitType(source.LimitType) || !finite(source.LimitBar) || source.LimitBar <= 0 {
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

func ResolveLoads(req Request) (Loads, error) {
	if err := Validate(req); err != nil {
		return Loads{}, err
	}
	total := req.RiderWeightKg + req.BikeWeightKg
	if req.FrontLoadKg != nil || req.RearLoadKg != nil {
		if req.FrontLoadKg == nil || req.RearLoadKg == nil || !finitePositive(*req.FrontLoadKg) || !finitePositive(*req.RearLoadKg) {
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

func finitePositive(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func EffectiveMaxPressureBar(req Request) (float64, string) {
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

// ComputeDynamics evaluates only the transparent first-order force model described
// in the specification. It deliberately does not infer a pressure recommendation.
func ComputeDynamics(req Request, loads Loads) (Dynamics, error) {
	if err := Validate(req); err != nil {
		return Dynamics{}, err
	}
	angle := 0.0
	if req.LeanAngleDeg != nil {
		angle = *req.LeanAngleDeg
	}
	muNominal, muLow, muHigh := demoMu(req.Surface, req.Weather)
	front := wheelDynamics(loads.FrontKg, angle, muNominal, muLow, muHigh, req.FrontOperatingPsi, req)
	rear := wheelDynamics(loads.RearKg, angle, muNominal, muLow, muHigh, req.RearOperatingPsi, req)
	return Dynamics{
		ModelVersion: DynamicsModelVersion,
		ModelStatus:  "DEMO_ESTIMATE_UNCALIBRATED",
		ForceFrame:   "GROUND",
		LeanAngleDeg: angle,
		MuSource:     "surface_prior_unvalidated",
		Front:        front,
		Rear:         rear,
	}, nil
}

func wheelDynamics(loadKg, angleDeg, muNominal, muLow, muHigh float64, pressurePsi *float64, req Request) WheelDynamics {
	fz := loadKg * GravityMps2
	fy := fz * math.Tan(angleDeg*math.Pi/180)
	resultant := math.Hypot(fz, fy)
	grip := muNominal * fz
	margin := 0.0
	if grip > 0 {
		margin = (grip - math.Abs(fy)) / grip * 100
	}
	result := WheelDynamics{
		LoadKg: loadKg, VerticalLoadN: fz, LateralDemandN: fy,
		ResultantContactForceN: resultant, MuNominal: muNominal,
		MuLow: muLow, MuHigh: muHigh, IdealizedGripLimitN: grip, GripMarginPct: margin,
	}
	if pressurePsi != nil && *pressurePsi > 0 {
		widthMm := req.NominalTireWidthMm
		if req.MeasuredTireWidthMm != nil {
			widthMm = *req.MeasuredTireWidthMm
		} else {
			widthMm += 0.4 * (req.InnerRimWidthMm - 19)
		}
		pressurePa := *pressurePsi * 6894.757293
		areaCm2 := fz / pressurePa * 10000
		if widthMm > 0 && finite(areaCm2) {
			result.EstimatedStaticContactCm2 = &areaCm2
		}
	}
	return result
}

func demoMu(surface SurfaceCondition, weather WeatherCondition) (nominal, low, high float64) {
	switch surface {
	case SurfaceSmoothTrack:
		nominal, low, high = 0.88, 0.70, 1.00
	case SurfaceRoughChip:
		nominal, low, high = 0.76, 0.58, 0.90
	case SurfaceCobbles:
		nominal, low, high = 0.64, 0.42, 0.82
	case SurfaceUnpavedGravel:
		nominal, low, high = 0.52, 0.30, 0.70
	default:
		nominal, low, high = 0.70, 0.35, 0.90
	}
	if weather == WeatherWet {
		nominal *= 0.65
		low *= 0.45
		high *= 0.80
	}
	return nominal, low, high
}

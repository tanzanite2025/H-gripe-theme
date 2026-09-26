package tirepressure

import (
	"errors"
	"math"
	"testing"
)

func baseRequest() Request {
	return Request{
		RiderWeightKg: 72, BikeWeightKg: 8.5, NominalTireWidthMm: 28,
		InnerRimWidthMm: 23, RimSystem: RimHookless,
		Position: PositionAggressiveRace, Surface: SurfaceRoughChip,
		Casing: CasingTubeless, Weather: WeatherDry,
	}
}

func TestResolveLoadsUsesMeasuredLoads(t *testing.T) {
	front, rear := 37.8, 42.7
	req := baseRequest()
	req.FrontLoadKg, req.RearLoadKg = &front, &rear
	loads, err := ResolveLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	if loads.Source != "MEASURED" || loads.FrontKg != front || loads.RearKg != rear {
		t.Fatalf("loads = %+v", loads)
	}
}

func TestResolveLoadsRejectsMismatchedMeasuredLoads(t *testing.T) {
	front, rear := 20.0, 20.0
	req := baseRequest()
	req.FrontLoadKg, req.RearLoadKg = &front, &rear
	if !errors.Is(func() error { _, err := ResolveLoads(req); return err }(), ErrLoadMismatch) {
		t.Fatal("expected load mismatch")
	}
}

func TestResolveLoadsEstimatesWhenMeasurementsMissing(t *testing.T) {
	loads, err := ResolveLoads(baseRequest())
	if err != nil {
		t.Fatal(err)
	}
	if loads.Source != "ESTIMATED" || loads.FrontKg+loads.RearKg != 80.5 {
		t.Fatalf("loads = %+v", loads)
	}
}

func TestEffectiveMaxPressureUsesLowestKnownLimit(t *testing.T) {
	limit := 5.0
	req := baseRequest()
	req.TireMaxPressureBar = &limit
	rim := 4.5
	req.RimMaxPressureBar = &rim
	req.LimitSources = []PressureLimitSource{
		{Name: "tire", Version: "v1", AppliesTo: "configuration", LimitType: "TIRE", LimitBar: limit},
		{Name: "rim", Version: "v1", AppliesTo: "configuration", LimitType: "RIM", LimitBar: rim},
	}
	got, status := EffectiveMaxPressureBar(req)
	if got != 4.5 || status != "PROVIDED_UNVERIFIED" {
		t.Fatalf("got %v/%s", got, status)
	}
}

func TestEffectiveMaxPressureDoesNotGuessWhenMissing(t *testing.T) {
	got, status := EffectiveMaxPressureBar(baseRequest())
	if got != 0 || status != "LIMITS_NOT_PROVIDED" {
		t.Fatalf("got %v/%s", got, status)
	}
}

func TestEffectiveMaxPressureAcceptsExplicitSystemSource(t *testing.T) {
	req := baseRequest()
	req.LimitSources = []PressureLimitSource{{
		Name: "system policy", Version: "policy-v1", AppliesTo: "selected setup", LimitType: "SYSTEM", LimitBar: 4.2,
	}}
	got, status := EffectiveMaxPressureBar(req)
	if got != 4.2 || status != "PROVIDED_UNVERIFIED" {
		t.Fatalf("got %v/%s", got, status)
	}
}

func TestValidateRequiresSourceRecordForProvidedLimit(t *testing.T) {
	limit := 5.0
	req := baseRequest()
	req.TireMaxPressureBar = &limit
	if err := Validate(req); err == nil {
		t.Fatal("expected source record requirement")
	}
}

func TestValidateRejectsIncompleteLimitSource(t *testing.T) {
	req := baseRequest()
	req.LimitSources = []PressureLimitSource{{Name: "rim", AppliesTo: "configuration", LimitType: "RIM", LimitBar: 4.5}}
	if err := Validate(req); err == nil {
		t.Fatal("expected incomplete source validation error")
	}
}

func TestValidateRejectsInvalidRimSystem(t *testing.T) {
	req := baseRequest()
	req.RimSystem = "MAYBE"
	if err := Validate(req); err == nil {
		t.Fatal("expected invalid rim system")
	}
}

func TestValidateRequiresSupportedConditionEnums(t *testing.T) {
	tests := []struct {
		name  string
		field string
		edit  func(*Request)
	}{
		{name: "position", field: "riding_position", edit: func(req *Request) { req.Position = "" }},
		{name: "surface", field: "surface_condition", edit: func(req *Request) { req.Surface = "MUD" }},
		{name: "casing", field: "casing_type", edit: func(req *Request) { req.Casing = "LATEX" }},
		{name: "weather", field: "weather_condition", edit: func(req *Request) { req.Weather = "SNOW" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := baseRequest()
			tt.edit(&req)
			err := Validate(req)
			validationErr, ok := err.(*ValidationError)
			if !ok || validationErr.Field != tt.field {
				t.Fatalf("expected validation error for %s, got %v", tt.field, err)
			}
		})
	}
}

func TestResolveLoadsRejectsNonFiniteMeasuredLoads(t *testing.T) {
	tests := []struct {
		name        string
		front, rear float64
	}{
		{name: "nan front", front: math.NaN(), rear: 80},
		{name: "infinite rear", front: 38, rear: math.Inf(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := baseRequest()
			req.FrontLoadKg, req.RearLoadKg = &tt.front, &tt.rear
			if _, err := ResolveLoads(req); err == nil {
				t.Fatal("expected measured load validation error")
			}
		})
	}
}

func TestResolveLoadsRejectsWheelLoadAboveSystemMass(t *testing.T) {
	front, rear := 81.0, 1.0
	req := baseRequest()
	req.FrontLoadKg, req.RearLoadKg = &front, &rear
	err := func() error { _, err := ResolveLoads(req); return err }()
	validationErr, ok := err.(*ValidationError)
	if !ok || validationErr.Code != "OUT_OF_RANGE" {
		t.Fatalf("expected out-of-range validation error, got %v", err)
	}
}

func TestComputeDynamicsUsesGroundFrameAndTransparentForceDecomposition(t *testing.T) {
	req := baseRequest()
	angle := 30.0
	frontPsi, rearPsi := 50.0, 55.0
	req.LeanAngleDeg = &angle
	req.FrontOperatingPsi, req.RearOperatingPsi = &frontPsi, &rearPsi
	loads, err := ResolveLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	dynamics, err := ComputeDynamics(req, loads)
	if err != nil {
		t.Fatal(err)
	}
	if dynamics.ForceFrame != "GROUND" || dynamics.ModelStatus != "DEMO_ESTIMATE_UNCALIBRATED" {
		t.Fatalf("unexpected dynamics metadata: %+v", dynamics)
	}
	wheel := dynamics.Front
	if math.Abs(wheel.VerticalLoadN-loads.FrontKg*GravityMps2) > 1e-9 {
		t.Fatalf("vertical load = %v", wheel.VerticalLoadN)
	}
	expectedLateral := wheel.VerticalLoadN * math.Tan(math.Pi/6)
	if math.Abs(wheel.LateralDemandN-expectedLateral) > 1e-9 {
		t.Fatalf("lateral demand = %v, want %v", wheel.LateralDemandN, expectedLateral)
	}
	if wheel.ResultantContactForceN <= wheel.VerticalLoadN || wheel.EstimatedStaticContactCm2 == nil {
		t.Fatalf("expected resultant and contact estimate: %+v", wheel)
	}
}

func TestValidateRejectsUnpairedOperatingPressure(t *testing.T) {
	req := baseRequest()
	pressure := 45.0
	req.FrontOperatingPsi = &pressure
	if err := Validate(req); err == nil {
		t.Fatal("expected paired operating pressure validation")
	}
}

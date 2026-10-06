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
		Position: PositionAggressiveRace, Surface: SurfaceFlatRoad,
	}
}

func TestResolveTirePressureWheelLoadsUsesMeasuredLoads(t *testing.T) {
	front, rear := 37.8, 42.7
	req := baseRequest()
	req.FrontLoadKg, req.RearLoadKg = &front, &rear
	loads, err := ResolveTirePressureWheelLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	if loads.Source != "MEASURED" || loads.FrontKg != front || loads.RearKg != rear {
		t.Fatalf("loads = %+v", loads)
	}
}

func TestResolveTirePressureWheelLoadsRejectsMismatchedMeasuredLoads(t *testing.T) {
	front, rear := 20.0, 20.0
	req := baseRequest()
	req.FrontLoadKg, req.RearLoadKg = &front, &rear
	if !errors.Is(func() error { _, err := ResolveTirePressureWheelLoads(req); return err }(), ErrLoadMismatch) {
		t.Fatal("expected load mismatch")
	}
}

func TestResolveTirePressureWheelLoadsEstimatesWhenMeasurementsMissing(t *testing.T) {
	loads, err := ResolveTirePressureWheelLoads(baseRequest())
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
	got, status := CalculateEffectiveProvidedTirePressureLimitBar(req)
	if got != 4.5 || status != "PROVIDED_UNVERIFIED" {
		t.Fatalf("got %v/%s", got, status)
	}
}

func TestEffectiveMaxPressureDoesNotGuessWhenMissing(t *testing.T) {
	got, status := CalculateEffectiveProvidedTirePressureLimitBar(baseRequest())
	if got != 0 || status != "LIMITS_NOT_PROVIDED" {
		t.Fatalf("got %v/%s", got, status)
	}
}

func TestEffectiveMaxPressureAcceptsExplicitSystemSource(t *testing.T) {
	req := baseRequest()
	req.LimitSources = []PressureLimitSource{{
		Name: "system policy", Version: "policy-v1", AppliesTo: "selected setup", LimitType: "SYSTEM", LimitBar: 4.2,
	}}
	got, status := CalculateEffectiveProvidedTirePressureLimitBar(req)
	if got != 4.2 || status != "PROVIDED_UNVERIFIED" {
		t.Fatalf("got %v/%s", got, status)
	}
}

func TestValidateTirePressureDynamicCalculationRequestRequiresSourceRecordForProvidedLimit(t *testing.T) {
	limit := 5.0
	req := baseRequest()
	req.TireMaxPressureBar = &limit
	if err := ValidateTirePressureDynamicCalculationRequest(req); err == nil {
		t.Fatal("expected source record requirement")
	}
}

func TestValidateTirePressureDynamicCalculationRequestRejectsIncompleteLimitSource(t *testing.T) {
	req := baseRequest()
	req.LimitSources = []PressureLimitSource{{Name: "rim", AppliesTo: "configuration", LimitType: "RIM", LimitBar: 4.5}}
	if err := ValidateTirePressureDynamicCalculationRequest(req); err == nil {
		t.Fatal("expected incomplete source validation error")
	}
}

func TestValidateTirePressureDynamicCalculationRequestRequiresReferencePressureForComparison(t *testing.T) {
	comparisonPressurePsi := 40.0
	req := baseRequest()
	req.FrontComparisonPressurePsi = &comparisonPressurePsi
	req.RearComparisonPressurePsi = &comparisonPressurePsi
	validationErr, ok := ValidateTirePressureDynamicCalculationRequest(req).(*ValidationError)
	if !ok || validationErr.Field != "front_comparison_pressure_psi/rear_comparison_pressure_psi" {
		t.Fatalf("expected comparison/reference pressure validation error, got %v", validationErr)
	}
}

func TestValidateTirePressureDynamicCalculationRequestRejectsInvalidRimSystem(t *testing.T) {
	req := baseRequest()
	req.RimSystem = "MAYBE"
	if err := ValidateTirePressureDynamicCalculationRequest(req); err == nil {
		t.Fatal("expected invalid rim system")
	}
}

func TestValidateTirePressureDynamicCalculationRequestRequiresSupportedConditionEnums(t *testing.T) {
	tests := []struct {
		name  string
		field string
		edit  func(*Request)
	}{
		{name: "position", field: "riding_position", edit: func(req *Request) { req.Position = "" }},
		{name: "surface", field: "surface_condition", edit: func(req *Request) { req.Surface = "MUD" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := baseRequest()
			tt.edit(&req)
			err := ValidateTirePressureDynamicCalculationRequest(req)
			validationErr, ok := err.(*ValidationError)
			if !ok || validationErr.Field != tt.field {
				t.Fatalf("expected validation error for %s, got %v", tt.field, err)
			}
		})
	}
}

func TestResolveTirePressureWheelLoadsRejectsNonFiniteMeasuredLoads(t *testing.T) {
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
			if _, err := ResolveTirePressureWheelLoads(req); err == nil {
				t.Fatal("expected measured load validation error")
			}
		})
	}
}

func TestResolveTirePressureWheelLoadsRejectsWheelLoadAboveSystemMass(t *testing.T) {
	front, rear := 81.0, 1.0
	req := baseRequest()
	req.FrontLoadKg, req.RearLoadKg = &front, &rear
	err := func() error { _, err := ResolveTirePressureWheelLoads(req); return err }()
	validationErr, ok := err.(*ValidationError)
	if !ok || validationErr.Code != "OUT_OF_RANGE" {
		t.Fatalf("expected out-of-range validation error, got %v", err)
	}
}

func TestCalculateTirePressureGroundFrameCorneringDynamicsUsesGroundFrameAndTransparentForceDecomposition(t *testing.T) {
	req := baseRequest()
	angle := 30.0
	req.SpeedKmh = 36
	frontPsi, rearPsi := 50.0, 55.0
	req.LeanAngleDeg = &angle
	req.FrontOperatingPsi, req.RearOperatingPsi = &frontPsi, &rearPsi
	loads, err := ResolveTirePressureWheelLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	dynamics, err := CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		t.Fatal(err)
	}
	if dynamics.ForceFrame != "GROUND" || dynamics.ModelStatus != "DEMO_ESTIMATE_UNCALIBRATED" {
		t.Fatalf("unexpected dynamics metadata: %+v", dynamics)
	}
	if dynamics.TireBodyNormalizationFactor != GenericTireBodyNormalizationBaseline || dynamics.TireBodyNormalizationSource != GenericTireBodyNormalizationSourceFixedBaseline {
		t.Fatalf("unexpected tire-body normalization metadata: %+v", dynamics)
	}
	wheel := dynamics.Front
	expectedVerticalLoad := math.Round(loads.FrontKg*GravityMps2*10) / 10
	if math.Abs(wheel.VerticalLoadN-expectedVerticalLoad) > 1e-9 {
		t.Fatalf("vertical load = %v", wheel.VerticalLoadN)
	}
	expectedLateral := math.Round(loads.FrontKg*GravityMps2*math.Tan(math.Pi/6)*10) / 10
	if math.Abs(wheel.LateralDemandN-expectedLateral) > 1e-9 {
		t.Fatalf("lateral demand = %v, want %v", wheel.LateralDemandN, expectedLateral)
	}
	if dynamics.TireWidthMm != 28 || dynamics.TireWidthSource != TireWidthSourceNominalUncorrected || dynamics.Surface != SurfaceFlatRoad || dynamics.MuSource != "flat_road_demo_baseline" {
		t.Fatalf("unexpected nominal width provenance: %+v", dynamics)
	}
	if dynamics.SpeedKmh != 36 || dynamics.EquivalentTurnRadiusM == nil || math.Abs(*dynamics.EquivalentTurnRadiusM-17.66) > 0.01 || math.Abs(dynamics.LateralAccelerationG-math.Tan(math.Pi/6)) > 0.001 {
		t.Fatalf("unexpected speed and corner geometry: %+v", dynamics)
	}
	if wheel.ResultantContactForceN <= wheel.VerticalLoadN || wheel.EstimatedStaticContactCm2 == nil {
		t.Fatalf("expected resultant and contact estimate: %+v", wheel)
	}
	if wheel.EstimatedEquivalentCircularContactDiameterMm == nil || wheel.EstimatedContactPatchWidthMm == nil || wheel.EstimatedContactPatchLengthMm == nil {
		t.Fatalf("expected width-limited circular footprint estimate: %+v", wheel)
	}
}

func TestValidateTirePressureDynamicCalculationRequestRejectsOutOfRangeSpeed(t *testing.T) {
	req := baseRequest()
	req.SpeedKmh = 121
	validationErr, ok := ValidateTirePressureDynamicCalculationRequest(req).(*ValidationError)
	if !ok || validationErr.Field != "speed_kmh" {
		t.Fatalf("expected speed validation error, got %v", validationErr)
	}
}

func TestGenericTireBodyNormalizationBaselineIsUsedInGripCalculation(t *testing.T) {
	wheel := CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(70, 20, FlatRoadDemoFrictionCoefficient, nil, 28)
	expectedGrip := math.Round(FlatRoadDemoFrictionCoefficient*70*GravityMps2*GenericTireBodyNormalizationBaseline*10) / 10
	if wheel.IdealizedGripLimitN != expectedGrip {
		t.Fatalf("grip=%v want=%v", wheel.IdealizedGripLimitN, expectedGrip)
	}
}

func TestOperatingPressureChangesContactAreaWithoutChangingFixedGripBaseline(t *testing.T) {
	lowPressurePsi, highPressurePsi := 40.0, 80.0
	lowPressureWheel := CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(70, 20, FlatRoadDemoFrictionCoefficient, &lowPressurePsi, 28)
	highPressureWheel := CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(70, 20, FlatRoadDemoFrictionCoefficient, &highPressurePsi, 28)
	if lowPressureWheel.EstimatedStaticContactCm2 == nil || highPressureWheel.EstimatedStaticContactCm2 == nil {
		t.Fatal("expected contact-area estimates at both operating pressures")
	}
	if lowPressureWheel.IdealizedGripLimitN != highPressureWheel.IdealizedGripLimitN {
		t.Fatalf("fixed grip baseline changed with pressure: low=%v high=%v", lowPressureWheel.IdealizedGripLimitN, highPressureWheel.IdealizedGripLimitN)
	}
	areaRatio := *lowPressureWheel.EstimatedStaticContactCm2 / *highPressureWheel.EstimatedStaticContactCm2
	if math.Abs(areaRatio-2) > 0.03 {
		t.Fatalf("contact area did not follow inverse pressure relationship: low=%v high=%v ratio=%v", *lowPressureWheel.EstimatedStaticContactCm2, *highPressureWheel.EstimatedStaticContactCm2, areaRatio)
	}
}

func TestPressureFrictionCoefficientEstimateExposesUnsupportedPressureEffect(t *testing.T) {
	pressurePsi := 48.0
	estimate := CalculateTirePressureFrictionCoefficientEstimate(FlatRoadDemoFrictionCoefficient, &pressurePsi)
	if estimate == nil {
		t.Fatal("expected pressure-friction coefficient status")
	}
	if estimate.DataStatus != PressureFrictionDataStatus || estimate.PressureEffectApplied {
		t.Fatalf("unexpected pressure-friction status: %+v", estimate)
	}
	if estimate.EstimatedCoefficient != FlatRoadDemoFrictionCoefficient || estimate.RelativeCoefficientIndex != 1 {
		t.Fatalf("pressure should retain fixed baseline until matched data exists: %+v", estimate)
	}
	if CalculateTirePressureFrictionCoefficientEstimate(FlatRoadDemoFrictionCoefficient, nil) != nil {
		t.Fatal("expected no pressure-friction status without operating pressure")
	}
}

func TestValidateTirePressureDynamicCalculationRequestRejectsMinimumPressureAboveOperatingPressure(t *testing.T) {
	operatingPressurePsi := 40.0
	minimumPressurePsi := 45.0
	req := baseRequest()
	req.FrontOperatingPsi, req.RearOperatingPsi = &operatingPressurePsi, &operatingPressurePsi
	req.FrontMinimumPressurePsi, req.RearMinimumPressurePsi = &minimumPressurePsi, &minimumPressurePsi
	validationErr, ok := ValidateTirePressureDynamicCalculationRequest(req).(*ValidationError)
	if !ok || validationErr.Code != "OUT_OF_RANGE" || validationErr.Field != "front_minimum_pressure_psi/rear_minimum_pressure_psi" {
		t.Fatalf("expected minimum-pressure range validation, got %v", validationErr)
	}
}

func TestValidateTirePressureDynamicCalculationRequestRequiresMinimumPressureForWetDemonstration(t *testing.T) {
	operatingPressurePsi := 48.0
	req := baseRequest()
	req.FrontOperatingPsi, req.RearOperatingPsi = &operatingPressurePsi, &operatingPressurePsi
	req.WetPressureDemonstrationEnabled = true
	validationErr, ok := ValidateTirePressureDynamicCalculationRequest(req).(*ValidationError)
	if !ok || validationErr.Code != "INVALID_FIELD" || validationErr.Field != "front_minimum_pressure_psi/rear_minimum_pressure_psi" {
		t.Fatalf("expected wet minimum-pressure validation, got %v", validationErr)
	}
}

func TestPressureContactAreaComparisonUsesOneWheelReferenceLoad(t *testing.T) {
	referencePressurePsi, comparisonPressurePsi := 50.0, 40.0
	req := baseRequest()
	req.FrontOperatingPsi = &referencePressurePsi
	req.RearOperatingPsi = &referencePressurePsi
	req.FrontComparisonPressurePsi = &comparisonPressurePsi
	req.RearComparisonPressurePsi = &comparisonPressurePsi
	loads, err := ResolveTirePressureWheelLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	dynamics, err := CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		t.Fatal(err)
	}
	comparison := dynamics.Front.PressureContactAreaComparison
	if comparison == nil {
		t.Fatal("expected pressure contact-area comparison")
	}
	if comparison.ReferencePressurePsi != referencePressurePsi || comparison.ComparisonPressurePsi != comparisonPressurePsi {
		t.Fatalf("unexpected pressure comparison: %+v", comparison)
	}
	if math.Abs(comparison.AreaChangePct-25) > 0.1 {
		t.Fatalf("area change = %v, want 25%%", comparison.AreaChangePct)
	}
	if math.Abs(comparison.ComparisonAreaCm2/comparison.ReferenceAreaCm2-1.25) > 0.01 {
		t.Fatalf("area ratio = %v, want 1.25", comparison.ComparisonAreaCm2/comparison.ReferenceAreaCm2)
	}
	if comparison.ReferencePatchWidthMm == nil || comparison.ReferencePatchLengthMm == nil || comparison.ComparisonPatchWidthMm == nil || comparison.ComparisonPatchLengthMm == nil {
		t.Fatalf("expected reference and comparison footprint dimensions: %+v", comparison)
	}
	if *comparison.ComparisonPatchLengthMm <= *comparison.ReferencePatchLengthMm {
		t.Fatalf("lower pressure should increase the width-limited footprint length: reference=%v comparison=%v", *comparison.ReferencePatchLengthMm, *comparison.ComparisonPatchLengthMm)
	}
	baselineRequest := req
	baselineRequest.FrontComparisonPressurePsi = nil
	baselineRequest.RearComparisonPressurePsi = nil
	baselineDynamics, err := CalculateTirePressureGroundFrameCorneringDynamics(baselineRequest, loads)
	if err != nil {
		t.Fatal(err)
	}
	if dynamics.Front.IdealizedGripLimitN != baselineDynamics.Front.IdealizedGripLimitN || dynamics.Rear.IdealizedGripLimitN != baselineDynamics.Rear.IdealizedGripLimitN {
		t.Fatalf("comparison should not alter fixed grip model: with=%v/%v without=%v/%v", dynamics.Front.IdealizedGripLimitN, dynamics.Rear.IdealizedGripLimitN, baselineDynamics.Front.IdealizedGripLimitN, baselineDynamics.Rear.IdealizedGripLimitN)
	}
}

func TestCalculateWidthLimitedEquivalentCircularContactPatchUsesTireWidthAsFootprintLimit(t *testing.T) {
	diameter, width, length, ok := CalculateWidthLimitedEquivalentCircularContactPatch(13.12, 28)
	if !ok {
		t.Fatal("expected valid footprint estimate")
	}
	if diameter <= width || width != 28 {
		t.Fatalf("expected equivalent diameter to be limited by tire width: diameter=%v width=%v", diameter, width)
	}
	if math.Abs(width*length-13.12*100) > 0.1 {
		t.Fatalf("estimated footprint does not preserve area: width=%v length=%v", width, length)
	}
}

func TestCalculateTirePressureGroundFrameCorneringDynamicsUsesMeasuredWidthWithoutUncalibratedRimExpansion(t *testing.T) {
	req := baseRequest()
	measuredWidth := 29.6
	angle := 0.0
	frontPsi, rearPsi := 48.0, 52.0
	req.MeasuredTireWidthMm = &measuredWidth
	req.LeanAngleDeg = &angle
	req.FrontOperatingPsi, req.RearOperatingPsi = &frontPsi, &rearPsi
	loads, err := ResolveTirePressureWheelLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	dynamics, err := CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		t.Fatal(err)
	}
	if dynamics.TireWidthMm != measuredWidth || dynamics.TireWidthSource != TireWidthSourceMeasured {
		t.Fatalf("unexpected measured width provenance: %+v", dynamics)
	}
}

func TestTirePressureVerticalDeformationReferenceDataLoadsFromLocalVersionedDataset(t *testing.T) {
	data := GetTirePressureVerticalDeformationReferenceData()
	if data.DataStatus != "REFERENCE_ONLY" || data.DatasetVersion == "" {
		t.Fatalf("unexpected reference-data metadata: %+v", data)
	}
	if data.VerticalStiffnessReference.InflationPressureBar != 4.75 || data.VerticalStiffnessReference.FittedVerticalStiffnessNPerMm != 173 {
		t.Fatalf("unexpected reference stiffness: %+v", data.VerticalStiffnessReference)
	}
	if len(data.VerticalForceDeflectionPoints) < 3 {
		t.Fatalf("expected digitized force-deflection points: %+v", data.VerticalForceDeflectionPoints)
	}
}

func TestCalculateTirePressureVerticalDeformationEstimateInterpolatesReferenceLoad(t *testing.T) {
	referencePressurePsi := 4.75 * PsiPerBar
	estimate := CalculateTirePressureVerticalDeformationEstimate(1000, &referencePressurePsi)
	if estimate == nil {
		t.Fatal("expected vertical deformation estimate")
	}
	if math.Abs(estimate.ReferenceDeflectionMm-6.0) > 0.01 || math.Abs(estimate.EstimatedDeflectionMm-6.0) > 0.01 {
		t.Fatalf("unexpected reference-pressure displacement: %+v", estimate)
	}
	if math.Abs(estimate.RelativeDeformationIndex-1) > 0.001 || math.Abs(estimate.EstimatedVerticalStiffnessNPerMm-173) > 0.1 {
		t.Fatalf("unexpected reference-pressure scaling: %+v", estimate)
	}
}

func TestCalculateTirePressureVerticalDeformationEstimateIncreasesWithLoadAndLowerPressure(t *testing.T) {
	highPressurePsi, lowPressurePsi := 70.0, 40.0
	lowLoad := CalculateTirePressureVerticalDeformationEstimate(500, &highPressurePsi)
	highLoad := CalculateTirePressureVerticalDeformationEstimate(1000, &highPressurePsi)
	lowerPressure := CalculateTirePressureVerticalDeformationEstimate(500, &lowPressurePsi)
	if lowLoad == nil || highLoad == nil || lowerPressure == nil {
		t.Fatal("expected vertical deformation estimates")
	}
	if highLoad.EstimatedDeflectionMm <= lowLoad.EstimatedDeflectionMm {
		t.Fatalf("higher load should increase estimated displacement: low=%+v high=%+v", lowLoad, highLoad)
	}
	if lowerPressure.EstimatedDeflectionMm <= lowLoad.EstimatedDeflectionMm || lowerPressure.RelativeDeformationIndex <= lowLoad.RelativeDeformationIndex {
		t.Fatalf("lower pressure should increase displacement and deformation index: high=%+v low=%+v", lowLoad, lowerPressure)
	}
}

func TestCalculateSingleWheelGroundFrameCorneringDynamicsOmitsVerticalDeformationWithoutPressure(t *testing.T) {
	wheel := CalculateSingleWheelGroundFrameCorneringDynamicsWithTireWidth(40, 0, FlatRoadDemoFrictionCoefficient, nil, 28)
	if wheel.VerticalDeformation != nil {
		t.Fatalf("vertical deformation should be omitted without operating pressure: %+v", wheel.VerticalDeformation)
	}
}

func TestValidateTirePressureDynamicCalculationRequestRejectsUnpairedOperatingPressure(t *testing.T) {
	req := baseRequest()
	pressure := 45.0
	req.FrontOperatingPsi = &pressure
	if err := ValidateTirePressureDynamicCalculationRequest(req); err == nil {
		t.Fatal("expected paired operating pressure validation")
	}
}

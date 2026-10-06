package tirepressure

import (
	"math"
	"testing"
)

func TestCalculateWetRoadFrictionRetentionRatioUsesSpeedWaterAndLateralDemand(t *testing.T) {
	if got := CalculateWetRoadFrictionRetentionRatio(0, 1, 1); got != 1 {
		t.Fatalf("zero speed retention = %v, want 1", got)
	}
	lowDemand := CalculateWetRoadFrictionRetentionRatio(30, 1, 0)
	highDemand := CalculateWetRoadFrictionRetentionRatio(30, 1, 1)
	if !(lowDemand > highDemand && highDemand < 1) {
		t.Fatalf("expected speed/water/lateral-demand loss: low=%v high=%v", lowDemand, highDemand)
	}
	if got := CalculateWetRoadFrictionRetentionRatio(120, 5, 1); got != WetRoadDemoMinimumFrictionRetentionRatio {
		t.Fatalf("retention lower bound = %v, want %v", got, WetRoadDemoMinimumFrictionRetentionRatio)
	}
}

func TestWetPressureCompensationDerivesEquivalentPressureAndArea(t *testing.T) {
	referencePressurePsi, rearPressurePsi := 50.0, 55.0
	angle := 30.0
	req := baseRequest()
	req.SpeedKmh = 30
	req.LeanAngleDeg = &angle
	req.FrontOperatingPsi, req.RearOperatingPsi = &referencePressurePsi, &rearPressurePsi
	minimumFrontPressurePsi, minimumRearPressurePsi := 20.0, 20.0
	req.FrontMinimumPressurePsi, req.RearMinimumPressurePsi = &minimumFrontPressurePsi, &minimumRearPressurePsi
	req.WetPressureDemonstrationEnabled = true
	waterFilmDepthMm := 1.0
	req.WaterFilmDepthMm = &waterFilmDepthMm
	loads, err := ResolveTirePressureWheelLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	dynamics, err := CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		t.Fatal(err)
	}
	compensation := dynamics.Front.WetPressureCompensation
	if compensation == nil {
		t.Fatal("expected wet pressure compensation")
	}
	if compensation.FrictionRetentionRatio >= 1 || compensation.EquivalentPressurePsi >= compensation.ReferencePressurePsi {
		t.Fatalf("expected wet equivalent pressure reduction: %+v", compensation)
	}
	if math.Abs(compensation.PressureReductionPsi-(compensation.ReferencePressurePsi-compensation.EquivalentPressurePsi)) > 0.05 {
		t.Fatalf("pressure reduction does not match equivalent pressure delta: %+v", compensation)
	}
	if math.Abs(compensation.PressureReductionPct-(compensation.PressureReductionPsi/compensation.ReferencePressurePsi*100)) > 0.05 {
		t.Fatalf("pressure reduction percentage does not match pressure delta: %+v", compensation)
	}
	if math.Abs(compensation.EquivalentContactAreaCm2/compensation.ReferenceContactAreaCm2-1/compensation.FrictionRetentionRatio) > 0.02 {
		t.Fatalf("area ratio does not compensate retention ratio: %+v", compensation)
	}
	if compensation.WetGripLimitAtReferencePressureN >= dynamics.Front.IdealizedGripLimitN {
		t.Fatalf("expected wet grip below dry baseline: %+v", compensation)
	}
	if compensation.WetGripMarginPct >= dynamics.Front.GripMarginPct {
		t.Fatalf("expected wet margin below dry baseline: %+v", compensation)
	}
	if compensation.SurfaceTextureBaseline != WetRoadDemoSurfaceTextureBaseline || compensation.RubberBaseline != WetRoadDemoRubberBaseline {
		t.Fatalf("unexpected fixed baselines: %+v", compensation)
	}
	if dynamics.Front.PressureContactAreaComparison == nil {
		t.Fatal("expected wet scenario to expose pressure-area comparison")
	}
}

func TestWetPressureCompensationClampsEquivalentPressureToProvidedMinimum(t *testing.T) {
	referencePressurePsi, rearPressurePsi := 50.0, 55.0
	minimumFrontPressurePsi, minimumRearPressurePsi := 45.0, 45.0
	angle := 30.0
	req := baseRequest()
	req.SpeedKmh = 30
	req.LeanAngleDeg = &angle
	req.FrontOperatingPsi, req.RearOperatingPsi = &referencePressurePsi, &rearPressurePsi
	req.FrontMinimumPressurePsi, req.RearMinimumPressurePsi = &minimumFrontPressurePsi, &minimumRearPressurePsi
	req.WetPressureDemonstrationEnabled = true
	loads, err := ResolveTirePressureWheelLoads(req)
	if err != nil {
		t.Fatal(err)
	}
	dynamics, err := CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		t.Fatal(err)
	}
	compensation := dynamics.Front.WetPressureCompensation
	if compensation == nil || !compensation.PressureClampedToMinimum || compensation.EquivalentPressurePsi != minimumFrontPressurePsi {
		t.Fatalf("expected minimum-pressure clamp: %+v", compensation)
	}
	if compensation.PressureReductionPsi != 5 || compensation.PressureReductionPct != 10 {
		t.Fatalf("expected reduction after minimum-pressure clamp: %+v", compensation)
	}
	if compensation.WetGripLimitAtReferencePressureN >= dynamics.Front.IdealizedGripLimitN {
		t.Fatalf("clamped wet grip should remain below dry target: %+v", compensation)
	}
}

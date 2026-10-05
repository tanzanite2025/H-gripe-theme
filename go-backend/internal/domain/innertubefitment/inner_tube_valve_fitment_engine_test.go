package innertubefitment

import (
	"errors"
	"testing"
)

func TestSolveAutomaticInnerTubeValveFitmentSupports100MillimetreRim(t *testing.T) {
	result, err := SolveInnerTubeValveFitment(SolveInput{RimDepthMM: 100, Mode: FitmentModeAutomatic})
	if err != nil {
		t.Fatal(err)
	}
	if result.RimDepthMM != 100 || result.PassageDepthMM != 93.5 {
		t.Fatalf("unexpected 100 mm geometry: %+v", result)
	}
	if result.ValveLengthMM != 80 || result.ExtenderLengthMM != 40 {
		t.Fatalf("100 mm rim should use the 80 mm valve and 40 mm extender: %+v", result)
	}
	if result.TotalAssemblyLengthMM != 120 || result.EffectiveExposureMM != 26.5 || result.Status != FitmentStatusOptimal {
		t.Fatalf("unexpected 100 mm result: %+v", result)
	}
	if result.MinimumRequiredLengthMM != 115 || result.PreferredMinimumLengthMM != 120 {
		t.Fatalf("unexpected length thresholds: %+v", result)
	}
	if result.MinimumLengthMarginMM != 5 || result.PreferredLengthMarginMM != 0 {
		t.Fatalf("unexpected length margins: %+v", result)
	}
	if result.RecommendationReasonKey != RecommendationReasonShortestExtenderAssembly || result.Recommendation.RecommendationReasonKey != RecommendationReasonShortestExtenderAssembly {
		t.Fatalf("unexpected automatic recommendation reason: %+v", result)
	}
	if len(result.Alternatives) != 2 || !result.Alternatives[0].IsAutomaticRecommendation || result.Alternatives[0].ValveLengthMM != 80 || result.Alternatives[0].ExtenderLengthMM != 40 {
		t.Fatalf("unexpected 100 mm alternatives: %+v", result.Alternatives)
	}
}

func TestGetInnerTubeValveFitmentMatrixMetadataIncludes100MillimetrePresetAndResult(t *testing.T) {
	metadata := GetInnerTubeValveFitmentMatrixMetadata()
	if metadata.RimDepthMaxMM != 100 {
		t.Fatalf("maximum rim depth = %d, want 100", metadata.RimDepthMaxMM)
	}
	if metadata.PumpHeadGripDepthMinMM != 10 || metadata.PumpHeadGripDepthMaxMM != 30 || metadata.DefaultPumpHeadGripDepthMM != 15 {
		t.Fatalf("unexpected pump-head grip depth metadata: %+v", metadata)
	}
	if metadata.RimDepthUncertaintyMaxMM != 5 || metadata.DefaultRimDepthUncertaintyMM != 2 {
		t.Fatalf("unexpected rim-depth uncertainty metadata: %+v", metadata)
	}
	has100MillimetrePreset := false
	for _, preset := range metadata.PresetInnerTubeRimDepthsMM {
		if preset == 100 {
			has100MillimetrePreset = true
			break
		}
	}
	if !has100MillimetrePreset {
		t.Fatal("100 mm preset is missing")
	}
	var row *MatrixRow
	for index := range metadata.Rows {
		if metadata.Rows[index].RimDepthMM == 100 {
			row = &metadata.Rows[index]
			break
		}
	}
	if row == nil {
		t.Fatal("100 mm matrix row is missing")
	}
	if row.RecommendedResult.ValveLengthMM != 80 || row.RecommendedResult.ExtenderLengthMM != 40 {
		t.Fatalf("unexpected 100 mm matrix recommendation: %+v", row.RecommendedResult.Recommendation)
	}
}

func TestSolveManualInnerTubeValveFitmentUsesBackendSuppliedLengths(t *testing.T) {
	valveLengthMM := 60
	extenderLengthMM := 40
	result, err := SolveInnerTubeValveFitment(SolveInput{
		RimDepthMM: 80, Mode: FitmentModeManual,
		ValveLengthMM: &valveLengthMM, ExtenderLengthMM: &extenderLengthMM,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalAssemblyLengthMM != 100 || result.EffectiveExposureMM != 26.5 || result.Status != FitmentStatusOptimal {
		t.Fatalf("unexpected manual result: %+v", result)
	}
	if result.RecommendationReasonKey != RecommendationReasonManualCombination || result.MinimumLengthMarginMM != 5 || result.PreferredLengthMarginMM != 0 {
		t.Fatalf("unexpected manual recommendation explanation: %+v", result)
	}
}

func TestSolveInnerTubeValveFitmentUsesPumpHeadGripDepthParameter(t *testing.T) {
	result, err := SolveInnerTubeValveFitment(SolveInput{
		RimDepthMM: 50, Mode: FitmentModeAutomatic, PumpHeadGripDepthMM: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.PumpHeadGripDepthMM != 20 || result.MinimumRequiredLengthMM != 70 || result.PreferredMinimumLengthMM != 75 {
		t.Fatalf("unexpected pump-head grip depth thresholds: %+v", result)
	}
	if result.PreferredExposureMM != 23 {
		t.Fatalf("preferred exposure = %v, want 23", result.PreferredExposureMM)
	}
}

func TestSolveInnerTubeValveFitmentReturnsUncertaintyIntervalReview(t *testing.T) {
	valveLengthMM := 60
	extenderLengthMM := 0
	result, err := SolveInnerTubeValveFitment(SolveInput{
		RimDepthMM: 50, Mode: FitmentModeManual,
		ValveLengthMM: &valveLengthMM, ExtenderLengthMM: &extenderLengthMM,
		PumpHeadGripDepthMM: 15, RimDepthUncertaintyMM: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.UncertaintyReview == nil {
		t.Fatal("uncertainty review is missing")
	}
	review := result.UncertaintyReview
	if review.LowerBound.RimDepthMM != 47 || review.UpperBound.RimDepthMM != 53 {
		t.Fatalf("unexpected uncertainty bounds: %+v", review)
	}
	if review.LowerBound.Status != FitmentStatusOptimal || review.UpperBound.Status != FitmentStatusUnsafe {
		t.Fatalf("unexpected boundary statuses: %+v", review)
	}
	if review.IsStatusStable || review.WorstCaseStatus != FitmentStatusUnsafe {
		t.Fatalf("unexpected uncertainty stability result: %+v", review)
	}
}

func TestSolveInnerTubeValveFitmentRejectsUnsupportedInputs(t *testing.T) {
	valveLengthMM := 100
	extenderLengthMM := 0
	tests := []struct {
		name  string
		input SolveInput
		err   error
	}{
		{name: "rim below range", input: SolveInput{RimDepthMM: 19, Mode: FitmentModeAutomatic}, err: ErrInvalidRimDepth},
		{name: "rim above range", input: SolveInput{RimDepthMM: 101, Mode: FitmentModeAutomatic}, err: ErrInvalidRimDepth},
		{name: "mode", input: SolveInput{RimDepthMM: 50, Mode: "other"}, err: ErrInvalidFitmentMode},
		{name: "missing manual values", input: SolveInput{RimDepthMM: 50, Mode: FitmentModeManual}, err: ErrMissingManualLength},
		{name: "pump head grip depth", input: SolveInput{RimDepthMM: 50, Mode: FitmentModeAutomatic, PumpHeadGripDepthMM: 31}, err: ErrInvalidPumpHeadGripDepth},
		{name: "rim depth uncertainty", input: SolveInput{RimDepthMM: 50, Mode: FitmentModeAutomatic, RimDepthUncertaintyMM: 6}, err: ErrInvalidRimDepthUncertainty},
		{name: "valve", input: SolveInput{RimDepthMM: 50, Mode: FitmentModeManual, ValveLengthMM: &valveLengthMM, ExtenderLengthMM: &extenderLengthMM}, err: ErrInvalidValveLength},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := SolveInnerTubeValveFitment(test.input)
			if !errors.Is(err, test.err) {
				t.Fatalf("error = %v, want %v", err, test.err)
			}
		})
	}
}

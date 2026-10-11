package tirerim

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestKnownWideTireReturnsOneBestRimSeparatelyFromNormalFitmentRange(t *testing.T) {
	tests := []struct {
		name                 string
		system               RimSystem
		tireWidthMM          int
		expectedInnerWidthMM int
		expectedProfile      TireRimApplicationProfile
	}{
		{"gravel lower boundary", RimSystemHookless, 31, 25, TireRimApplicationProfileGravelAllRoad},
		{"gravel upper boundary", RimSystemHookless, 50, 30, TireRimApplicationProfileGravelAllRoad},
		{"MTB hookless", RimSystemHookless, 60, 28, TireRimApplicationProfileMTBXC},
		{"MTB hooked", RimSystemHooked, 60, 31, TireRimApplicationProfileMTBXC},
		{"reviewed fat snow suggestion", RimSystemHookless, 105, 85, TireRimApplicationProfileFatSnow},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ResolveTireRimWidthReference(test.tireWidthMM, test.system)
			if err != nil {
				t.Fatal(err)
			}
			best := result.BestFitment
			if best == nil || best.RimInnerWidthMM != test.expectedInnerWidthMM || best.TireWidthMM != test.tireWidthMM || best.ApplicationProfile != test.expectedProfile {
				t.Fatalf("unexpected single best pairing: %#v", best)
			}
			if best.SelectionMethod != "recommended_window_center" || best.ModelStatus != "pending_profile_calibration" {
				t.Fatalf("selection must identify its provisional model basis: %#v", best)
			}
			if result.AeroRuleApplies || result.DerivedMetrics.AeroTargetOuterWidthMM != nil || !reflect.DeepEqual(result.RimWidthRanges, result.NormalFitmentRimWidthRanges) {
				t.Fatalf("legacy normal ranges and non-aero path changed: %#v", result)
			}
		})
	}
}

func TestKnownWideRimReturnsOneBestTireSeparatelyFromNormalFitmentRange(t *testing.T) {
	tests := []struct {
		name                      string
		system                    RimSystem
		innerWidthMM              int
		expectedTireWidthMM       int
		expectedTargetTireWidthMM float64
	}{
		{"gravel half millimetre tie", RimSystemHookless, 28, 43, 42.5},
		{"MTB", RimSystemHookless, 35, 67, 67.31},
		{"wide hooked", RimSystemHooked, 45, 78, 78},
		{"reviewed fat snow suggestion", RimSystemHookless, 85, 105, 105.41},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ResolveTireRimWidthReferenceForRim(test.system, test.innerWidthMM)
			if err != nil {
				t.Fatal(err)
			}
			best := result.BestFitment
			if best == nil || best.RimInnerWidthMM != test.innerWidthMM || best.TireWidthMM != test.expectedTireWidthMM || best.TargetTireWidthMM != test.expectedTargetTireWidthMM {
				t.Fatalf("unexpected single best tire: %#v", best)
			}
			if result.AeroRuleApplies || result.AeroTargetTireWidthMM != nil || result.BestMatch != nil {
				t.Fatalf("wide-rim selection must not use aero fields: %#v", result)
			}
			calculation, err := CalculateTireRimEngineeringReference(test.system, test.innerWidthMM, test.expectedTireWidthMM+1, result.ResolvedApplicationProfile)
			if err == nil {
				exact := BuildTireRimEngineeringSuggestion(calculation)
				if !reflect.DeepEqual(exact.BestFitment, best) || exact.Engineering.TireWidthMM != test.expectedTireWidthMM+1 {
					t.Fatalf("entered pairing must remain separate from the rim recommendation: %#v", exact)
				}
			}
		})
	}
}

func TestExistingBestWideFitmentStaysInsideItsProjectRecommendationWindowAfterISONormalRangeIntegration(t *testing.T) {
	for _, system := range []RimSystem{RimSystemHookless, RimSystemHooked} {
		for tireWidthMM := 31; tireWidthMM <= 127; tireWidthMM++ {
			result, err := ResolveTireRimWidthReference(tireWidthMM, system)
			if err != nil {
				t.Fatalf("%s / %d mm: %v", system, tireWidthMM, err)
			}
			if result.BestFitment == nil {
				if len(result.RecommendedRimWidthRanges) > 0 {
					t.Fatalf("missing best value for recommended %s / %d mm", system, tireWidthMM)
				}
				continue
			}
			best := result.BestFitment
			calculation, err := CalculateTireRimEngineeringReference(system, best.RimInnerWidthMM, best.TireWidthMM)
			projectRule := tireRimEngineeringRules[system][best.ApplicationProfile][best.RimInnerWidthMM]
			if err != nil || best.TireWidthMM < projectRule.RecommendedTireWidthRange.Min || best.TireWidthMM > projectRule.RecommendedTireWidthRange.Max || best.TireWidthMM != tireWidthMM || calculation.AeroRuleApplies {
				t.Fatalf("selected an invalid or adjusted tire pairing for %s / %d mm: %#v error=%v", system, tireWidthMM, best, err)
			}
		}
		for innerWidthMM := 19; innerWidthMM <= 100; innerWidthMM++ {
			result, err := ResolveTireRimWidthReferenceForRim(system, innerWidthMM)
			if err != nil {
				t.Fatal(err)
			}
			if result.BestFitment == nil {
				continue
			}
			best := result.BestFitment
			calculation, err := CalculateTireRimEngineeringReference(system, innerWidthMM, best.TireWidthMM, result.ResolvedApplicationProfile)
			if err != nil || best.RimInnerWidthMM != innerWidthMM {
				t.Fatalf("selected an invalid rim pairing for %s / %d mm: %#v error=%v", system, innerWidthMM, best, err)
			}
			projectRule := tireRimEngineeringRules[system][best.ApplicationProfile][innerWidthMM]
			if !result.AeroRuleApplies && (best.TireWidthMM <= 30 || best.TireWidthMM < projectRule.RecommendedTireWidthRange.Min || best.TireWidthMM > projectRule.RecommendedTireWidthRange.Max || calculation.AeroRuleApplies) {
				t.Fatalf("wide path selected a narrow or non-recommended tire: %#v", best)
			}
		}
	}
}

func TestNewBestFitmentDoesNotLabelBlockedOrTransitionalAeroTargetAsBest(t *testing.T) {
	for _, innerWidthMM := range []int{19, 20} {
		result, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, innerWidthMM)
		if err != nil {
			t.Fatal(err)
		}
		if result.BestFitment != nil || result.BestMatch == nil {
			t.Fatalf("blocked/transitional target must remain a legacy geometric value: %#v", result)
		}
	}
	result, err := ResolveTireRimWidthReferenceForRim(RimSystemHooked, 19)
	if err != nil {
		t.Fatal(err)
	}
	if result.BestFitment == nil || result.BestFitment.TireWidthMM != result.BestMatch.TireWidthMM {
		t.Fatalf("valid aero target should still produce its existing match: %#v", result)
	}
}

func TestResolveTireRimWidthReferenceReturnsSingleAeroBestInnerWidth(t *testing.T) {
	result, err := ResolveTireRimWidthReference(30, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "engineering_recommended" {
		t.Fatalf("expected engineering recommendation, got %q", result.ResultKind)
	}
	if !result.AeroRuleApplies {
		t.Fatal("expected the inclusive 30C boundary to use the aero rule")
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 23, Max: 23}) {
		t.Fatalf("unexpected exact ranges: %#v", result.RimWidthRanges)
	}
	if result.Calculation != nil {
		t.Fatal("exact row should not include interpolation details")
	}
	if result.ApplicationProfile != TireRimApplicationProfileRoadAero {
		t.Fatalf("expected road aero application profile, got %q", result.ApplicationProfile)
	}
	if len(result.NormalFitmentRimWidthRanges) != 1 || result.NormalFitmentRimWidthRanges[0] != (WidthRange{Min: 16, Max: 25}) {
		t.Fatalf("unexpected normal fitment ranges: %#v", result.NormalFitmentRimWidthRanges)
	}
	if result.StandardReference.ISO5775Part1 != "ISO 5775-1:2023" || result.StandardReference.ISO5775Part2 != "ISO 5775-2:2021" || result.StandardReference.ETRTOEdition != "" || result.StandardReference.DTSwissUsedByCalculation {
		t.Fatalf("unexpected standard reference context: %#v", result.StandardReference)
	}
}

func TestResolveTireRimWidthReferenceClassifiesWideTireWithoutAeroRule(t *testing.T) {
	result, err := ResolveTireRimWidthReference(60, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ApplicationProfile != TireRimApplicationProfileMTBXC {
		t.Fatalf("expected MTB/XC application profile, got %q", result.ApplicationProfile)
	}
	if result.AeroRuleApplies {
		t.Fatal("60C must not use the 105% aero rule")
	}
	if len(result.NormalFitmentRimWidthRanges) == 0 {
		t.Fatal("expected normal fitment ranges for 60C")
	}
}

func TestResolveTireRimWidthReferenceReturnsSingleAeroBestInnerWidthAcrossWidths(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		system     RimSystem
		expected   []WidthRange
		resultKind string
	}{
		{
			name:       "hookless 30 millimetres",
			width:      30,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 23, Max: 23}},
			resultKind: "engineering_recommended",
		},
		{
			name:       "hooked 28 millimetres",
			width:      28,
			system:     RimSystemHooked,
			expected:   []WidthRange{{Min: 21, Max: 21}},
			resultKind: "engineering_recommended",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := ResolveTireRimWidthReference(testCase.width, testCase.system)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if result.ResultKind != testCase.resultKind {
				t.Fatalf("expected result kind %q, got %q", testCase.resultKind, result.ResultKind)
			}
			if !result.AeroRuleApplies {
				t.Fatal("expected aero rule to apply")
			}
			if len(result.RimWidthRanges) != len(testCase.expected) {
				t.Fatalf("expected ranges %#v, got %#v", testCase.expected, result.RimWidthRanges)
			}
			for index, expectedRange := range testCase.expected {
				if result.RimWidthRanges[index] != expectedRange {
					t.Fatalf("range %d: expected %#v, got %#v", index, expectedRange, result.RimWidthRanges[index])
				}
			}
		})
	}
}

func TestResolveTireRimWidthReferenceReturnsNormalFitmentRangeFor31MmTire(t *testing.T) {
	result, err := ResolveTireRimWidthReference(31, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "engineering_normal_fitment" {
		t.Fatalf("expected normal fitment result, got %q", result.ResultKind)
	}
	if result.AeroRuleApplies {
		t.Fatal("31C must not use the 105% aero rule")
	}
	if result.DerivedMetrics.AeroTargetOuterWidthMM != nil {
		t.Fatalf("31C must not return an aero target: %#v", result.DerivedMetrics.AeroTargetOuterWidthMM)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 16, Max: 25}) {
		t.Fatalf("unexpected normal fitment ranges: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceUsesNormalFitmentRangeFor50MmTire(t *testing.T) {
	result, err := ResolveTireRimWidthReference(50, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.AeroRuleApplies || result.DerivedMetrics.AeroTargetOuterWidthMM != nil {
		t.Fatalf("50C must bypass the 105%% aero rule: %#v", result)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 17, Max: 30}) {
		t.Fatalf("unexpected 50C normal fitment range: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceDoesNotReturnSeveralEquallySuitableInnerWidths(t *testing.T) {
	result, err := ResolveTireRimWidthReference(30, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "engineering_recommended" {
		t.Fatalf("expected engineering recommendation, got %q", result.ResultKind)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 23, Max: 23}) {
		t.Fatalf("unexpected engineering range: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceRejectsWidthsOutsideTheEngineeringModel(t *testing.T) {
	for _, width := range []int{17, 133, 158} {
		result, err := ResolveTireRimWidthReference(width, RimSystemHookless)
		if result != nil || !errors.Is(err, ErrInvalidTireWidth) {
			t.Fatalf("expected unsupported width error for %d mm, result=%#v error=%v", width, result, err)
		}
	}
}

func TestResolveTireRimWidthReferenceReportsTheInputWidthAsTireWidth(t *testing.T) {
	result, err := ResolveTireRimWidthReference(30, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.DerivedMetrics.Status != "engineering" {
		t.Fatalf("unexpected derived metric status: %q", result.DerivedMetrics.Status)
	}
	if result.DerivedMetrics.TireWidthMM.Min != 30 || result.DerivedMetrics.TireWidthMM.Max != 30 {
		t.Fatalf("unexpected direct tire-width value: %#v", result.DerivedMetrics.TireWidthMM)
	}
	if result.DerivedMetrics.AeroTargetOuterWidthMM == nil || result.DerivedMetrics.AeroTargetOuterWidthMM.Min != 28.6 || result.DerivedMetrics.AeroTargetOuterWidthMM.Max != 28.6 {
		t.Fatalf("unexpected inverse 105%% outer-width projection: %#v", result.DerivedMetrics.AeroTargetOuterWidthMM)
	}
	if len(result.RimOuterWidthEstimates) != 1 {
		t.Fatalf("expected one outer-width estimate for the best-fit rim, got %#v", result.RimOuterWidthEstimates)
	}
	if result.RimOuterWidthEstimates[0] != (RimOuterWidthEstimate{
		RimInnerWidthMM:                23,
		WallThicknessPerSideEstimateMM: 3,
		RimOuterWidthEstimateMM:        29,
	}) {
		t.Fatalf("unexpected best-fit rim outer-width estimates: %#v", result.RimOuterWidthEstimates)
	}
}

func TestStandardCoveredNarrowTiresReturnNormalRangesWithoutInventingAnExistingBestRecommendation(t *testing.T) {
	for width := 18; width <= 25; width++ {
		result, err := ResolveTireRimWidthReference(width, RimSystemHookless)
		if err != nil || result.BestFitment != nil || len(result.NormalFitmentRimWidthRanges) != 1 || len(result.RimWidthRanges) != 0 || len(result.RimOuterWidthEstimates) != 0 {
			t.Fatalf("standard-covered %d mm tire lost normal range or acquired a new optimum: %#v error=%v", width, result, err)
		}
	}
}

func TestCalculateTireRimEngineeringReferenceUsesInputWidthForThe105PercentAeroProjection(t *testing.T) {
	result, err := CalculateTireRimEngineeringReference(RimSystemHookless, 25, 30)
	if err != nil {
		t.Fatalf("calculate engineering reference: %v", err)
	}
	if result.Verdict != TireRimEngineeringVerdictRecommended {
		t.Fatalf("expected recommended verdict, got %q", result.Verdict)
	}
	if result.TireWidthMM != 30 || result.AeroTargetOuterWidthMM == nil || *result.AeroTargetOuterWidthMM != 28.6 || result.AeroTargetTireWidthMM == nil || *result.AeroTargetTireWidthMM != 32.6 {
		t.Fatalf("unexpected inverse aero projection: tire=%v target=%v", result.TireWidthMM, result.AeroTargetOuterWidthMM)
	}
	if estimates := BuildTireRimEngineeringSuggestion(result).RimOuterWidthEstimates; len(estimates) != 1 || estimates[0] != (RimOuterWidthEstimate{
		RimInnerWidthMM:                25,
		WallThicknessPerSideEstimateMM: 3,
		RimOuterWidthEstimateMM:        31,
	}) {
		t.Fatalf("unexpected exact rim outer-width estimate: %#v", estimates)
	}
	if result.MaximumPressure.Bar != 4.5 || result.MaximumPressure.PSI != 65 {
		t.Fatalf("unexpected pressure limit: %#v", result.MaximumPressure)
	}
}

func TestCalculateTireRimEngineeringReferenceDoesNotAdjustThirtyFiveMillimetreInputByRimWidth(t *testing.T) {
	result, err := CalculateTireRimEngineeringReference(RimSystemHooked, 25, 35)
	if err != nil {
		t.Fatalf("calculate engineering reference: %v", err)
	}
	if result.TireWidthMM != 35 || result.AeroRuleApplies || result.AeroTargetOuterWidthMM != nil || result.AeroTargetTireWidthMM != nil {
		t.Fatalf("expected 35C to bypass the aero rule, got tire=%v outer=%v tireTarget=%v", result.TireWidthMM, result.AeroTargetOuterWidthMM, result.AeroTargetTireWidthMM)
	}
}

func TestCalculateTireRimEngineeringReferenceBlocksHooklessTwentyFiveMillimetreRimWithTwentyEightMillimetreTire(t *testing.T) {
	result, err := CalculateTireRimEngineeringReference(RimSystemHookless, 25, 28)
	if err != nil {
		t.Fatalf("calculate engineering reference: %v", err)
	}
	if result.Verdict != TireRimEngineeringVerdictCritical || result.VerdictReason != TireRimEngineeringReasonBelowMinimum {
		t.Fatalf("expected below-minimum critical result, got verdict=%q reason=%q", result.Verdict, result.VerdictReason)
	}
}

func TestCalculateTireRimEngineeringReferenceMarksNineteenMillimetreHooklessRimAsTransitional(t *testing.T) {
	result, err := CalculateTireRimEngineeringReference(RimSystemHookless, 19, 28)
	if err != nil {
		t.Fatalf("calculate engineering reference: %v", err)
	}
	if result.Verdict != TireRimEngineeringVerdictReference || result.VerdictReason != TireRimEngineeringReasonTransitionalRim {
		t.Fatalf("expected transitional reference result, got verdict=%q reason=%q", result.Verdict, result.VerdictReason)
	}
}

func TestCalculateTireRimEngineeringReferenceRejectsUnsupportedInnerWidth(t *testing.T) {
	_, err := CalculateTireRimEngineeringReference(RimSystemHookless, 101, 127)
	if !errors.Is(err, ErrInvalidTireRimEngineeringInnerWidth) {
		t.Fatalf("expected unsupported inner-width error, got %v", err)
	}
}

func TestTireRimEngineeringModelSupportsOneMillimetreWidthsThroughSnowBikeBoundary(t *testing.T) {
	for _, innerWidthMM := range []int{20, 24, 31, 50, 76, 99} {
		if err := ValidateTireRimEngineeringInnerWidth(innerWidthMM); err != nil {
			t.Fatalf("expected %d mm to be supported: %v", innerWidthMM, err)
		}
	}
	for _, system := range []RimSystem{RimSystemHookless, RimSystemHooked} {
		result, err := CalculateTireRimEngineeringReference(system, 100, 127)
		if err != nil {
			t.Fatalf("calculate %s snow-bike boundary: %v", system, err)
		}
		if result.AllowedTireWidthRange != (WidthRange{Min: 114, Max: 132}) || result.RecommendedTireWidthRange != nil || BuildTireRimEngineeringSuggestion(result).BestFitment != nil {
			t.Fatalf("unexpected %s 100 mm boundary rule: %#v", system, result)
		}
		if result.Verdict != TireRimEngineeringVerdictReference || result.VerdictReason != TireRimEngineeringReasonNoRecommendation || result.MaximumPressure != nil {
			t.Fatalf("model boundary must not become an optimum or pressure rating for %s: %#v", system, result)
		}
		if result.TubeRecommendation.Band != "fat" || result.TubeRecommendation.WidthMinMM != 90 || result.TubeRecommendation.WidthMaxMM != 127 {
			t.Fatalf("expected %s fat-bike tube band at 127 mm, got %#v", system, result.TubeRecommendation)
		}
	}
}

func TestResolveTireRimWidthReferenceForRimSelectsClosestTireWidth(t *testing.T) {
	result, err := ResolveTireRimWidthReferenceForRim(RimSystemHooked, 19)
	if err != nil {
		t.Fatalf("resolve reverse tire guidance: %v", err)
	}
	if result.RimOuterWidthEstimate.RimOuterWidthEstimateMM != 25 {
		t.Fatalf("unexpected rim outer-width estimate: %#v", result.RimOuterWidthEstimate)
	}
	if result.AeroTargetTireWidthMM == nil || *result.AeroTargetTireWidthMM != 26.3 {
		t.Fatalf("unexpected target tire width: %v", result.AeroTargetTireWidthMM)
	}
	if result.BestMatch.TireWidthMM != 26 || result.BestMatch.TargetTireWidthDifferenceMM != 0.3 {
		t.Fatalf("unexpected closest tire-width match: %#v", result.BestMatch)
	}
}

func TestResolveTireRimWidthReferenceForRimUsesFullSearchForHooklessRim(t *testing.T) {
	result, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 19)
	if err != nil {
		t.Fatalf("resolve reverse tire guidance: %v", err)
	}
	if result.BestMatch.TireWidthMM != 26 || result.BestMatch.TargetTireWidthDifferenceMM != 0.3 {
		t.Fatalf("unexpected full-range tire-width match: %#v", result.BestMatch)
	}
}

func TestResolveTireRimWidthReferenceForRimSearchesAllSupportedTireWidths(t *testing.T) {
	result, err := ResolveTireRimWidthReferenceForRim(RimSystemHooked, 45)
	if err != nil {
		t.Fatalf("resolve wide-rim reverse tire guidance: %v", err)
	}
	if result.RimOuterWidthEstimate.RimOuterWidthEstimateMM != 51 {
		t.Fatalf("unexpected wide-rim outer-width estimate: %#v", result.RimOuterWidthEstimate)
	}
	if result.AeroRuleApplies || result.AeroTargetTireWidthMM != nil || result.BestMatch != nil {
		t.Fatalf("wide rim must bypass the aero target: target=%v match=%v", result.AeroTargetTireWidthMM, result.BestMatch)
	}
	if result.NormalFitmentTireWidthRange != (WidthRange{Min: 72, Max: 95}) || result.RecommendedFitmentTireWidthRange == nil || *result.RecommendedFitmentTireWidthRange != (WidthRange{Min: 71, Max: 85}) {
		t.Fatalf("unexpected wide-rim normal fitment ranges: normal=%#v recommended=%#v", result.NormalFitmentTireWidthRange, result.RecommendedFitmentTireWidthRange)
	}
	if result.ApplicationProfile != TireRimApplicationProfileAutoByWidth || result.StandardReference.DTSwissReferenceRole != "display_only" {
		t.Fatalf("unexpected reverse lookup result context: profile=%q standard=%#v", result.ApplicationProfile, result.StandardReference)
	}
}

func TestReviewedMountainAndPlusWindowsProducePurposeSpecificTargetsAndVerdicts(t *testing.T) {
	tests := []struct {
		innerWidthMM    int
		profile         TireRimApplicationProfile
		recommended     WidthRange
		targetMM        float64
		bestTireWidthMM int
	}{
		{25, TireRimApplicationProfileMTBXC, WidthRange{53, 61}, 57.15, 57},
		{28, TireRimApplicationProfileMTBXC, WidthRange{56, 64}, 59.69, 60},
		{30, TireRimApplicationProfileMTBXC, WidthRange{58, 66}, 62.23, 62},
		{35, TireRimApplicationProfileMTBXC, WidthRange{64, 71}, 67.31, 67},
		{35, TireRimApplicationProfileMTBPlus, WidthRange{66, 76}, 71.12, 71},
		{40, TireRimApplicationProfileMTBPlus, WidthRange{71, 81}, 76.2, 76},
	}
	for _, test := range tests {
		result, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, test.innerWidthMM, test.profile)
		if err != nil {
			t.Fatal(err)
		}
		best := result.BestFitment
		if best == nil || best.TireWidthMM != test.bestTireWidthMM || best.TargetTireWidthMM != test.targetMM || best.ApplicationProfile != test.profile {
			t.Fatalf("%d mm / %s: unexpected purpose-specific best: %#v", test.innerWidthMM, test.profile, best)
		}
		if result.RecommendedFitmentTireWidthRange == nil || *result.RecommendedFitmentTireWidthRange != test.recommended || result.ResolvedApplicationProfile != test.profile || result.ApplicationProfileSelection != "explicit" {
			t.Fatalf("unexpected purpose window or selection: %#v", result)
		}
		calculation, err := CalculateTireRimEngineeringReference(RimSystemHookless, test.innerWidthMM, test.bestTireWidthMM, test.profile)
		if err != nil {
			t.Fatal(err)
		}
		if calculation.Verdict != TireRimEngineeringVerdictRecommended || !reflect.DeepEqual(calculation.RecommendedTireWidthRange, result.RecommendedFitmentTireWidthRange) || !reflect.DeepEqual(BuildTireRimEngineeringSuggestion(calculation).BestFitment, best) {
			t.Fatalf("same purpose must drive recommendation and exact verdict: %#v", calculation)
		}
		if len(best.FitmentRuleReference.ManufacturerSourceURLs) == 0 || best.FitmentRuleReference.Interpolated || len(result.FitmentRuleReference.ManufacturerSourceURLs) != 0 || result.AeroRuleApplies || calculation.AeroTargetTireWidthMM != nil {
			t.Fatalf("reviewed wide window must retain sources without aero targets: %#v", result)
		}
	}
	trail, _ := CalculateTireRimEngineeringReference(RimSystemHookless, 35, 74, TireRimApplicationProfileMTBXC)
	plus, _ := CalculateTireRimEngineeringReference(RimSystemHookless, 35, 74, TireRimApplicationProfileMTBPlus)
	if trail.Verdict != TireRimEngineeringVerdictReference || plus.Verdict != TireRimEngineeringVerdictRecommended || trail.AllowedTireWidthRange != plus.AllowedTireWidthRange {
		t.Fatalf("purpose must change the recommendation window without changing ISO normal bounds: trail=%#v plus=%#v", trail, plus)
	}
}

func TestManufacturerGuidanceRemainsScopedToSystemAndPurposeAndInterpolationIsIdentified(t *testing.T) {
	hooked, err := ResolveTireRimWidthReferenceForRim(RimSystemHooked, 30, TireRimApplicationProfileMTBXC)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooked.BestFitment.FitmentRuleReference.ManufacturerSourceURLs) != 2 || hooked.BestFitment.FitmentRuleReference.ManufacturerSourceURLs[0] != "https://www.wtb.com/pages/tire-rim-fit-chart" || hooked.BestFitment.TargetTireWidthMM == 62.23 {
		t.Fatalf("hookless guidance leaked into hooked rules: %#v", hooked)
	}
	gravel, err := CalculateTireRimEngineeringReference(RimSystemHookless, 28, 60, TireRimApplicationProfileGravelAllRoad)
	if err != nil {
		t.Fatal(err)
	}
	if gravel.Verdict != TireRimEngineeringVerdictReference || len(gravel.FitmentRuleReference.ManufacturerSourceURLs) != 0 {
		t.Fatalf("MTB guidance leaked into gravel rules: %#v", gravel)
	}
	interpolated, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 29, TireRimApplicationProfileMTBXC)
	if err != nil {
		t.Fatal(err)
	}
	if interpolated.BestFitment == nil || interpolated.BestFitment.TargetTireWidthMM != 60.96 || !interpolated.BestFitment.FitmentRuleReference.Interpolated || interpolated.BestFitment.FitmentRuleReference.RecommendedRangeBasis != "project_profile_window_interpolation" || len(interpolated.BestFitment.FitmentRuleReference.ManufacturerSourceURLs) != 3 {
		t.Fatalf("interpolated row must identify project derivation and both anchor sources: %#v", interpolated)
	}
	interpolated.BestFitment.FitmentRuleReference.ManufacturerSourceURLs[0] = "changed"
	again, _ := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 29, TireRimApplicationProfileMTBXC)
	if again.BestFitment.FitmentRuleReference.ManufacturerSourceURLs[0] == "changed" {
		t.Fatal("response mutation changed the shared rule sources")
	}
}

func TestHookedMountainRecommendationsFollowPublishedOptimalCellsAndKeepProjectGapEstimatesExplicit(t *testing.T) {
	// These source row bounds are independently transcribed from WTB's solid
	// optimal-performance cells, including grouped inner-width columns.
	for _, source := range []struct {
		tireWidthMM         int
		minimumInnerWidthMM int
		maximumInnerWidthMM int
	}{
		{51, 25, 30}, {52, 25, 30}, {54, 25, 30}, {55, 25, 30}, {56, 25, 30},
		{57, 26, 30}, {58, 26, 32}, {59, 26, 32}, {60, 28, 33}, {61, 28, 33},
		{62, 28, 33}, {64, 28, 35}, {65, 30, 35}, {66, 32, 40}, {67, 32, 40},
		{68, 35, 43}, {69, 35, 43}, {70, 35, 43}, {71, 35, 43},
	} {
		result, err := ResolveTireRimWidthReference(source.tireWidthMM, RimSystemHooked, TireRimApplicationProfileMTBXC)
		if err != nil {
			t.Fatal(err)
		}
		best := result.BestFitment
		if best == nil || best.RimInnerWidthMM < source.minimumInnerWidthMM || best.RimInnerWidthMM > source.maximumInnerWidthMM || best.FitmentRuleReference.RecommendedRangeBasis != "reviewed_manufacturer_optimal_bands_project_envelope" {
			t.Fatalf("%d mm recommendation left WTB optimal cells: %#v", source.tireWidthMM, best)
		}
		if best.ReasonCode != "closest_reviewed_recommended_tire_window_center" || result.AeroRuleApplies || result.DerivedMetrics.AeroTargetOuterWidthMM != nil {
			t.Fatalf("wrong reviewed mountain selection path: %#v", result)
		}
	}
	for _, expected := range []struct {
		innerWidthMM    int
		window          WidthRange
		targetMM        float64
		bestTireWidthMM int
	}{
		{25, WidthRange{50, 56}, 53, 53},
		{28, WidthRange{50, 64}, 57, 57},
		{30, WidthRange{50, 65}, 57.5, 58},
		{31, WidthRange{58, 65}, 61.5, 62},
		{32, WidthRange{58, 67}, 62.5, 63},
		{33, WidthRange{60, 67}, 63.5, 64},
		{35, WidthRange{64, 71}, 67.5, 68},
		{40, WidthRange{66, 71}, 68.5, 69},
		{43, WidthRange{68, 71}, 69.5, 70},
	} {
		result, err := ResolveTireRimWidthReferenceForRim(RimSystemHooked, expected.innerWidthMM, TireRimApplicationProfileMTBXC)
		if err != nil {
			t.Fatal(err)
		}
		if result.RecommendedFitmentTireWidthRange == nil || *result.RecommendedFitmentTireWidthRange != expected.window || result.BestFitment == nil || result.BestFitment.TargetTireWidthMM != expected.targetMM || result.BestFitment.TireWidthMM != expected.bestTireWidthMM {
			t.Fatalf("%d mm: wrong calibrated project target: %#v", expected.innerWidthMM, result)
		}
		if result.BestFitment.ModelStatus != "pending_profile_calibration" {
			t.Fatalf("source windows do not prove physical performance: %#v", result)
		}
	}
	for _, missingSourceTireWidthMM := range []int{53, 63} {
		result, err := ResolveTireRimWidthReference(missingSourceTireWidthMM, RimSystemHooked)
		if err != nil {
			t.Fatal(err)
		}
		if result.BestFitment == nil || result.BestFitment.FitmentRuleReference.RecommendedRangeBasis != "reviewed_manufacturer_optimal_bands_project_envelope" {
			t.Fatalf("missing source rows must remain identified project envelope estimates: %#v", result)
		}
	}
}

func TestHookedMountainCompatibleOnlyCellsDoNotCreateAnOptimalRecommendation(t *testing.T) {
	result, err := CalculateTireRimEngineeringReference(RimSystemHooked, 45, 70, TireRimApplicationProfileMTBXC)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != TireRimEngineeringVerdictCritical || result.VerdictReason != TireRimEngineeringReasonBelowMinimum || result.RecommendedTireWidthRange != nil || BuildTireRimEngineeringSuggestion(result).BestFitment != nil {
		t.Fatalf("unreviewed old window survived outside reviewed optimal rim widths: %#v", result)
	}
	for _, tireWidthMM := range []int{72, 80, 96, 127} {
		result, err := ResolveTireRimWidthReference(tireWidthMM, RimSystemHooked, TireRimApplicationProfileMTBXC)
		if err != nil {
			t.Fatal(err)
		}
		if result.BestFitment != nil || len(result.RecommendedRimWidthRanges) != 0 || len(result.NormalFitmentRimWidthRanges) == 0 {
			t.Fatalf("compatible-only chart rows became a best or lost normal ranges: %#v", result)
		}
	}
	result, err = CalculateTireRimEngineeringReference(RimSystemHooked, 30, 65, TireRimApplicationProfileMTBXC)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != TireRimEngineeringVerdictRecommended || result.AllowedTireWidthRange.Max < 65 {
		t.Fatalf("new recommendation was blocked by the old normal ceiling: %#v", result)
	}
}

func TestReviewedPurposeWindowsTakePriorityOverCloserUnreviewedCentersWithoutInventingPerformanceWeights(t *testing.T) {
	for _, expected := range []struct {
		tireWidthMM  int
		profile      TireRimApplicationProfile
		innerWidthMM int
		targetMM     float64
	}{
		{70, TireRimApplicationProfileMTBXC, 35, 67.31},
		{78, TireRimApplicationProfileMTBPlus, 40, 76.2},
	} {
		result, err := ResolveTireRimWidthReference(expected.tireWidthMM, RimSystemHookless, expected.profile)
		if err != nil {
			t.Fatal(err)
		}
		best := result.BestFitment
		if best == nil || best.RimInnerWidthMM != expected.innerWidthMM || best.TargetTireWidthMM != expected.targetMM || best.ReasonCode != "closest_reviewed_recommended_tire_window_center" || len(best.FitmentRuleReference.ManufacturerSourceURLs) == 0 {
			t.Fatalf("unreviewed center displaced reviewed evidence: %#v", result)
		}
	}
	fallback, err := ResolveTireRimWidthReference(84, RimSystemHookless, TireRimApplicationProfileMTBPlus)
	if err != nil {
		t.Fatal(err)
	}
	if fallback.BestFitment == nil || fallback.BestFitment.ReasonCode != "closest_recommended_tire_window_center" || fallback.BestFitment.ModelStatus != "pending_profile_calibration" {
		t.Fatalf("out-of-source legacy fallback was hidden or misrepresented: %#v", fallback)
	}
	oneReviewedEnd, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 44, TireRimApplicationProfileMTBPlus)
	if err != nil {
		t.Fatal(err)
	}
	if oneReviewedEnd.BestFitment == nil || tireRimEngineeringRules[RimSystemHookless][TireRimApplicationProfileMTBPlus][44].RecommendedWindowUsesReviewedGuidance {
		t.Fatalf("one reviewed interpolation endpoint certified the whole mixed window: %#v", oneReviewedEnd)
	}
}

func TestReviewedMountainAndPlusTargetsRemainNondecreasingWithinEachReviewedInnerWidthSpan(t *testing.T) {
	for _, span := range []struct {
		system              RimSystem
		profile             TireRimApplicationProfile
		minimumInnerWidthMM int
		maximumInnerWidthMM int
	}{
		{RimSystemHooked, TireRimApplicationProfileMTBXC, 25, 43},
		{RimSystemHookless, TireRimApplicationProfileMTBXC, 25, 35},
		{RimSystemHookless, TireRimApplicationProfileMTBPlus, 35, 40},
	} {
		previousTargetMM := 0.0
		for innerWidthMM := span.minimumInnerWidthMM; innerWidthMM <= span.maximumInnerWidthMM; innerWidthMM++ {
			result, err := ResolveTireRimWidthReferenceForRim(span.system, innerWidthMM, span.profile)
			if err != nil {
				t.Fatal(err)
			}
			best := result.BestFitment
			if best == nil || best.TargetTireWidthMM < previousTargetMM {
				t.Fatalf("reviewed %s / %s / %d mm target moved backwards: %#v previous=%v", span.system, span.profile, innerWidthMM, best, previousTargetMM)
			}
			previousTargetMM = best.TargetTireWidthMM
			calculation, err := CalculateTireRimEngineeringReference(span.system, innerWidthMM, best.TireWidthMM, span.profile)
			if err != nil || calculation.Verdict != TireRimEngineeringVerdictRecommended || !reflect.DeepEqual(BuildTireRimEngineeringSuggestion(calculation).BestFitment, best) {
				t.Fatalf("calibrated target failed its own rule: %#v err=%v", best, err)
			}
		}
	}
}

func TestExplicitApplicationProfileValidationPreservesTheInclusiveAeroBoundary(t *testing.T) {
	for _, tireWidthMM := range []int{30, 31} {
		result, err := ResolveTireRimWidthReference(tireWidthMM, RimSystemHookless)
		if err != nil {
			t.Fatal(err)
		}
		if result.AeroRuleApplies != (tireWidthMM == 30) {
			t.Fatalf("wrong aero scope: %#v", result)
		}
	}
	for _, test := range []struct {
		width    int
		profile  TireRimApplicationProfile
		expected error
	}{
		{30, TireRimApplicationProfileMTBXC, ErrTireRimApplicationProfileOutsideAeroScope},
		{31, TireRimApplicationProfileRoadAero, ErrTireRimApplicationProfileOutsideAeroScope},
		{60, "unknown", ErrInvalidTireRimApplicationProfile},
	} {
		_, knownErr := ResolveTireRimWidthReference(test.width, RimSystemHookless, test.profile)
		_, exactErr := CalculateTireRimEngineeringReference(RimSystemHookless, 30, test.width, test.profile)
		if !errors.Is(knownErr, test.expected) || !errors.Is(exactErr, test.expected) {
			t.Fatalf("profile validation differs by input mode: %v / %v", knownErr, exactErr)
		}
	}
	_, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 30, TireRimApplicationProfileRoadAero)
	if !errors.Is(err, ErrTireRimApplicationProfileOutsideAeroScope) {
		t.Fatalf("wide reverse target entered aero scope: %v", err)
	}
	_, err = ResolveTireRimWidthReferenceForRim(RimSystemHookless, 30, "unknown")
	if !errors.Is(err, ErrInvalidTireRimApplicationProfile) {
		t.Fatalf("unknown reverse profile accepted: %v", err)
	}
	if NormalizeTireRimApplicationProfile(" MTB_PLUS ") != TireRimApplicationProfileMTBPlus {
		t.Fatal("profile normalization failed")
	}
}

func TestEmptyPurposeWindowRetainsOriginalVerdictAndDoesNotInventAnotherPurposeRecommendation(t *testing.T) {
	calculation, err := CalculateTireRimEngineeringReference(RimSystemHookless, 35, 127)
	if err != nil {
		t.Fatal(err)
	}
	result := BuildTireRimEngineeringSuggestion(calculation)
	if calculation.ApplicationProfile != TireRimApplicationProfileFatSnow || calculation.Verdict != TireRimEngineeringVerdictCritical || calculation.VerdictReason != TireRimEngineeringReasonAboveMaximum || calculation.RecommendedTireWidthRange != nil || result.BestFitment != nil {
		t.Fatalf("fat input must not receive an unrelated mountain best: %#v", result)
	}
	calculation, err = CalculateTireRimEngineeringReference(RimSystemHookless, 40, 60, TireRimApplicationProfileGravelAllRoad)
	if err != nil {
		t.Fatal(err)
	}
	if calculation.Verdict != TireRimEngineeringVerdictCritical || calculation.VerdictReason != TireRimEngineeringReasonBelowMinimum || calculation.RecommendedTireWidthRange != nil {
		t.Fatalf("an empty purpose window must not suppress the ISO below-minimum verdict: %#v", calculation)
	}
	for _, width := range []int{26, 27} {
		reverse, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, width, TireRimApplicationProfileMTBPlus)
		if err != nil {
			t.Fatal(err)
		}
		projectRule := tireRimEngineeringRules[RimSystemHookless][TireRimApplicationProfileMTBPlus][width]
		if reverse.RecommendedFitmentTireWidthRange != nil || reverse.BestFitment != nil || !projectRule.FitmentRuleReference.Interpolated || reverse.FitmentRuleReference.Interpolated {
			t.Fatalf("empty neighboring windows must not produce a spurious interpolation: %#v", reverse)
		}
	}
}

func TestEveryExplicitWideProfileKeepsRecommendationsWithinItsOwnRuleAcrossAllSupportedWidths(t *testing.T) {
	for _, system := range []RimSystem{RimSystemHooked, RimSystemHookless} {
		for _, profile := range []TireRimApplicationProfile{TireRimApplicationProfileGravelAllRoad, TireRimApplicationProfileMTBXC, TireRimApplicationProfileMTBPlus, TireRimApplicationProfileFatSnow} {
			for innerWidthMM := 19; innerWidthMM <= 100; innerWidthMM++ {
				result, err := ResolveTireRimWidthReferenceForRim(system, innerWidthMM, profile)
				if err != nil {
					t.Fatal(err)
				}
				if result.AeroRuleApplies || result.AeroTargetTireWidthMM != nil || result.BestMatch != nil {
					t.Fatalf("wide purpose produced an aero value: %#v", result)
				}
				if result.BestFitment == nil {
					continue
				}
				best := result.BestFitment
				calculation, err := CalculateTireRimEngineeringReference(system, innerWidthMM, best.TireWidthMM, profile)
				projectRule := tireRimEngineeringRules[system][profile][innerWidthMM]
				if err != nil || best.TireWidthMM < projectRule.RecommendedTireWidthRange.Min || best.TireWidthMM > projectRule.RecommendedTireWidthRange.Max || calculation.ApplicationProfile != profile || best.ApplicationProfile != profile || best.RimInnerWidthMM != innerWidthMM {
					t.Fatalf("recommendation failed same-profile validation: %#v err=%v", best, err)
				}
			}
			for tireWidthMM := 31; tireWidthMM <= 127; tireWidthMM++ {
				result, err := ResolveTireRimWidthReference(tireWidthMM, system, profile)
				if err != nil {
					t.Fatal(err)
				}
				if result.BestFitment == nil {
					continue
				}
				best := result.BestFitment
				calculation, err := CalculateTireRimEngineeringReference(system, best.RimInnerWidthMM, tireWidthMM, profile)
				projectRule := tireRimEngineeringRules[system][profile][best.RimInnerWidthMM]
				if err != nil || tireWidthMM < projectRule.RecommendedTireWidthRange.Min || tireWidthMM > projectRule.RecommendedTireWidthRange.Max || calculation.ApplicationProfile != profile || best.ApplicationProfile != profile || best.TireWidthMM != tireWidthMM {
					t.Fatalf("known tire failed same-profile validation: %#v err=%v", best, err)
				}
			}
		}
	}
}

func TestAutomaticApplicationProfileTransitionsIdentifyDefaultsWithoutBecomingCompatibilityBounds(t *testing.T) {
	for _, test := range []struct {
		tireWidthMM     int
		expectedProfile TireRimApplicationProfile
	}{
		{30, TireRimApplicationProfileRoadAero},
		{31, TireRimApplicationProfileGravelAllRoad},
		{50, TireRimApplicationProfileGravelAllRoad},
		{51, TireRimApplicationProfileMTBXC},
		{70, TireRimApplicationProfileMTBXC},
		{71, TireRimApplicationProfileMTBPlus},
		{85, TireRimApplicationProfileMTBPlus},
		{86, TireRimApplicationProfileFatSnow},
	} {
		result, err := ResolveTireRimWidthReference(test.tireWidthMM, RimSystemHookless)
		if err != nil {
			t.Fatal(err)
		}
		if result.ApplicationProfile != test.expectedProfile || result.ApplicationProfileSelection != "automatic_by_tire_width" {
			t.Fatalf("wrong default at %d mm: %#v", test.tireWidthMM, result)
		}
	}
	result, err := CalculateTireRimEngineeringReference(RimSystemHookless, 35, 66, TireRimApplicationProfileMTBPlus)
	if err != nil {
		t.Fatal(err)
	}
	if result.ApplicationProfile != TireRimApplicationProfileMTBPlus || result.Verdict != TireRimEngineeringVerdictRecommended {
		t.Fatalf("default 71 mm boundary incorrectly blocked explicit Plus guidance: %#v", result)
	}
	reverse, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 35)
	if err != nil {
		t.Fatal(err)
	}
	if reverse.ApplicationProfile != TireRimApplicationProfileAutoByWidth || reverse.ResolvedApplicationProfile != TireRimApplicationProfileMTBXC || reverse.ApplicationProfileSelection != "automatic_by_rim_default_window" {
		t.Fatalf("reverse lookup did not identify its actual default: %#v", reverse)
	}
}

func TestFatSnowUsesOnlyReviewedSizeSuggestionsWithoutCompatibilityMidpointsOrNeighboringOptima(t *testing.T) {
	for _, system := range []RimSystem{RimSystemHookless, RimSystemHooked} {
		for innerWidthMM := TireRimEngineeringMinInnerWidthMM; innerWidthMM <= TireRimEngineeringMaxInnerWidthMM; innerWidthMM++ {
			result, err := ResolveTireRimWidthReferenceForRim(system, innerWidthMM, TireRimApplicationProfileFatSnow)
			if err != nil {
				t.Fatal(err)
			}
			if result.AeroRuleApplies || result.AeroTargetTireWidthMM != nil || result.BestMatch != nil {
				t.Fatalf("Fat purpose used an aero projection: %#v", result)
			}
			if system == RimSystemHookless && innerWidthMM == 85 {
				best := result.BestFitment
				if best == nil || best.TireWidthMM != 105 || best.TargetTireWidthMM != 105.41 || best.TargetTireWidthDifferenceMM != 0.41 || best.ModelStatus != "pending_profile_calibration" || result.RecommendedFitmentTireWidthRange == nil || *result.RecommendedFitmentTireWidthRange != (WidthRange{97, 114}) || result.NormalFitmentTireWidthRange != (WidthRange{96, 132}) {
					t.Fatalf("ENVE M685 suggestion did not create its single project default: %#v", result)
				}
				if best.FitmentRuleReference.RecommendedRangeBasis != "reviewed_manufacturer_guidance_project_window" || !best.FitmentRuleReference.Interpolated || !reflect.DeepEqual(best.FitmentRuleReference.ManufacturerSourceURLs, []string{"https://enve.com/pages/tire-compatibility"}) || result.FitmentRuleReference.Interpolated {
					t.Fatalf("exact recommendation lost source or legacy normal interpolation identity: %#v", result)
				}
				continue
			}
			projectRule := tireRimEngineeringRules[system][TireRimApplicationProfileFatSnow][innerWidthMM]
			if result.BestFitment != nil || result.RecommendedFitmentTireWidthRange != nil || projectRule.FitmentRuleReference.RecommendedRangeBasis != "no_reviewed_tire_size_suggestion_for_fat_snow" || len(result.FitmentRuleReference.ManufacturerSourceURLs) != 0 {
				t.Fatalf("unreviewed Fat width acquired an optimum or manufacturer identity: %#v", result)
			}
			if result.NormalFitmentTireWidthRange != findISO5775Part12023NormalTireWidthRangeForRimInnerWidth(innerWidthMM) {
				t.Fatalf("purpose-specific product evidence rewrote the ISO normal range: %#v", result)
			}
			calculation, err := CalculateTireRimEngineeringReference(system, innerWidthMM, result.NormalFitmentTireWidthRange.Max, TireRimApplicationProfileFatSnow)
			if err != nil || calculation.Verdict == TireRimEngineeringVerdictRecommended || calculation.RecommendedTireWidthRange != nil || BuildTireRimEngineeringSuggestion(calculation).BestFitment != nil || calculation.MaximumPressure != nil {
				t.Fatalf("Fat normal reference was promoted to an optimum or pressure rating: %#v error=%v", calculation, err)
			}
		}
		for tireWidthMM := 86; tireWidthMM <= MaxInputTireWidthMM; tireWidthMM++ {
			result, err := ResolveTireRimWidthReference(tireWidthMM, system)
			if err != nil {
				t.Fatal(err)
			}
			hasReviewedSuggestion := system == RimSystemHookless && tireWidthMM >= 97 && tireWidthMM <= 114
			if hasReviewedSuggestion {
				if result.BestFitment == nil || result.BestFitment.RimInnerWidthMM != 85 || result.BestFitment.TireWidthMM != tireWidthMM || !reflect.DeepEqual(result.RecommendedRimWidthRanges, []WidthRange{{85, 85}}) {
					t.Fatalf("known Fat tire escaped the reviewed suggestion or changed input width: %#v", result)
				}
			} else if result.BestFitment != nil || len(result.RecommendedRimWidthRanges) != 0 || len(result.NormalFitmentRimWidthRanges) == 0 {
				t.Fatalf("compatibility-only Fat tire acquired an optimum or lost its reference range: %#v", result)
			}
		}
	}
}

func TestFatSnowExactPairingRetainsEnteredWidthAndDistinguishesRecommendationFromNormalEnvelope(t *testing.T) {
	for _, expected := range []struct {
		innerWidthMM int
		tireWidthMM  int
		verdict      TireRimEngineeringVerdict
		reason       TireRimEngineeringVerdictReason
		hasBest      bool
	}{
		{76, 115, TireRimEngineeringVerdictReference, TireRimEngineeringReasonNoRecommendation, false},
		{76, 97, TireRimEngineeringVerdictReference, TireRimEngineeringReasonNoRecommendation, false},
		{85, 96, TireRimEngineeringVerdictReference, TireRimEngineeringReasonReference, true},
		{85, 97, TireRimEngineeringVerdictRecommended, TireRimEngineeringReasonRecommended, true},
		{85, 114, TireRimEngineeringVerdictRecommended, TireRimEngineeringReasonRecommended, true},
		{85, 115, TireRimEngineeringVerdictReference, TireRimEngineeringReasonReference, true},
		{100, 126, TireRimEngineeringVerdictReference, TireRimEngineeringReasonNoRecommendation, false},
		{100, 127, TireRimEngineeringVerdictReference, TireRimEngineeringReasonNoRecommendation, false},
	} {
		calculation, err := CalculateTireRimEngineeringReference(RimSystemHookless, expected.innerWidthMM, expected.tireWidthMM)
		if err != nil {
			t.Fatal(err)
		}
		result := BuildTireRimEngineeringSuggestion(calculation)
		if calculation.Verdict != expected.verdict || calculation.VerdictReason != expected.reason || calculation.TireWidthMM != expected.tireWidthMM || result.TireWidthMM != expected.tireWidthMM || result.RimWidthRanges[0] != (WidthRange{expected.innerWidthMM, expected.innerWidthMM}) || (result.BestFitment != nil) != expected.hasBest {
			t.Fatalf("original Fat pairing was replaced or its verdict was lost: %#v", result)
		}
		if expected.hasBest && result.BestFitment.TireWidthMM != 105 {
			t.Fatalf("fixed-rim alternative must remain distinct from the entered tire: %#v", result)
		}
		if calculation.MaximumPressure != nil || calculation.AeroTargetTireWidthMM != nil || calculation.AeroTargetOuterWidthMM != nil {
			t.Fatalf("Fat pairing received generic pressure or aero targets: %#v", calculation)
		}
	}
}

func TestFatSnowPressureRequiresProductConditionsAndUnsupportedWidthsAreNeverClamped(t *testing.T) {
	for _, system := range []RimSystem{RimSystemHookless, RimSystemHooked} {
		for _, requested := range []struct {
			tireWidthMM int
			profile     TireRimApplicationProfile
		}{
			{105, TireRimApplicationProfileAutoByWidth},
			{60, TireRimApplicationProfileFatSnow},
			{127, TireRimApplicationProfileMTBPlus},
		} {
			calculation, err := CalculateTireRimEngineeringReference(system, 76, requested.tireWidthMM, requested.profile)
			if err != nil || calculation.MaximumPressure != nil {
				t.Fatalf("unsupported pressure conditions received a generic maximum: %#v error=%v", calculation, err)
			}
		}
		_, err := ResolveTireRimWidthReference(158, system)
		if !errors.Is(err, ErrInvalidTireWidth) {
			t.Fatalf("6.2 inch dedicated system was clamped into the model: %v", err)
		}
		_, err = CalculateTireRimEngineeringReference(system, 100, 158)
		if !errors.Is(err, ErrInvalidTireWidth) {
			t.Fatalf("exact dedicated pairing was clamped into the model: %v", err)
		}
		_, err = ResolveTireRimWidthReferenceForRim(system, 101)
		if !errors.Is(err, ErrInvalidTireRimEngineeringInnerWidth) {
			t.Fatalf("unsupported inner width was clamped into the model: %v", err)
		}
	}
	first, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 85)
	if err != nil {
		t.Fatal(err)
	}
	first.BestFitment.FitmentRuleReference.ManufacturerSourceURLs[0] = "changed"
	again, err := ResolveTireRimWidthReferenceForRim(RimSystemHookless, 85)
	if err != nil || len(again.FitmentRuleReference.ManufacturerSourceURLs) != 0 || again.BestFitment.FitmentRuleReference.ManufacturerSourceURLs[0] != "https://enve.com/pages/tire-compatibility" {
		t.Fatalf("Fat source response mutation changed the unique engine: %#v error=%v", again, err)
	}
}

func TestISONormalWidthMatrixMatchesEveryTireWidthAndRemainsIndependentOfProjectBestProfiles(t *testing.T) {
	// Expected cells are independently transcribed from the supplied matrix.
	// Do not derive this fixture from the production table or project windows.
	expectedRows := []struct {
		tireMinimum, tireMaximum, rimMinimum, rimMaximum int
	}{
		{18, 19, 13, 15}, {20, 21, 13, 17}, {22, 24, 13, 20},
		{25, 27, 13, 22}, {28, 28, 15, 23}, {29, 34, 16, 25},
		{35, 46, 17, 27}, {47, 57, 17, 30}, {58, 65, 21, 35},
		{66, 71, 25, 43}, {72, 83, 31, 53}, {84, 95, 41, 64},
		{96, 113, 59, 89}, {114, 132, 72, 100},
	}
	for _, system := range []RimSystem{RimSystemHooked, RimSystemHookless} {
		for _, row := range expectedRows {
			for tireWidthMM := row.tireMinimum; tireWidthMM <= row.tireMaximum; tireWidthMM++ {
				profiles := []TireRimApplicationProfile{TireRimApplicationProfileAutoByWidth}
				if tireWidthMM <= 30 {
					profiles = append(profiles, TireRimApplicationProfileRoadAero)
				} else {
					profiles = append(profiles, TireRimApplicationProfileGravelAllRoad, TireRimApplicationProfileMTBXC, TireRimApplicationProfileMTBPlus, TireRimApplicationProfileFatSnow)
				}
				for _, profile := range profiles {
					result, err := ResolveTireRimWidthReference(tireWidthMM, system, profile)
					if err != nil || !reflect.DeepEqual(result.NormalFitmentRimWidthRanges, []WidthRange{{row.rimMinimum, row.rimMaximum}}) {
						t.Fatalf("%s / %s / %d mm: ISO cells changed: %#v error=%v", system, profile, tireWidthMM, result, err)
					}
					if result.NormalFitmentReference != (TireRimNormalFitmentTableReference{"ISO 5775-1:2023", "4.4", "7", "user_provided_standard_width_matrix_image", "nominal_width_pairing_only"}) || result.StandardReference.ETRTOEdition != "" || result.StandardReference.DTSwissUsedByCalculation {
						t.Fatalf("normal result misidentified its applied source: %#v", result)
					}
					for rimWidthMM := 13; rimWidthMM <= 100; rimWidthMM++ {
						calculation, err := CalculateTireRimEngineeringReference(system, rimWidthMM, tireWidthMM, profile)
						outsideNormalRange := rimWidthMM < row.rimMinimum || rimWidthMM > row.rimMaximum
						if err != nil || (calculation.Verdict == TireRimEngineeringVerdictCritical) != outsideNormalRange {
							t.Fatalf("%s / %s / %d mm rim / %d mm tire: entered verdict does not match ISO membership: %#v error=%v", system, profile, rimWidthMM, tireWidthMM, calculation, err)
						}
						if calculation.NormalFitmentReference != result.NormalFitmentReference || calculation.FitmentRuleReference.NormalRangeBasis != "iso_5775_1_2023_clause_4_4_table_7" || calculation.FitmentRuleReference.Interpolated || len(calculation.FitmentRuleReference.ManufacturerSourceURLs) != 0 || calculation.FitmentRuleReference.RecommendedRangeBasis != "" {
							t.Fatalf("normal source was mixed with optimization guidance: %#v", calculation)
						}
					}
				}
			}
		}
	}
}

func TestISONormalReverseRangesCoverEveryIntegerRimWidthWithoutInterpolation(t *testing.T) {
	expectedRows := []struct {
		rimMinimum, rimMaximum, tireMinimum, tireMaximum int
	}{
		{13, 14, 18, 27}, {15, 15, 18, 28}, {16, 16, 20, 34},
		{17, 17, 20, 57}, {18, 20, 22, 57}, {21, 22, 25, 65},
		{23, 23, 28, 65}, {24, 24, 29, 65}, {25, 25, 29, 71},
		{26, 27, 35, 71}, {28, 30, 47, 71}, {31, 35, 58, 83},
		{36, 40, 66, 83}, {41, 43, 66, 95}, {44, 53, 72, 95},
		{54, 58, 84, 95}, {59, 64, 84, 113}, {65, 71, 96, 113},
		{72, 89, 96, 132}, {90, 100, 114, 132},
	}
	for _, system := range []RimSystem{RimSystemHooked, RimSystemHookless} {
		for _, row := range expectedRows {
			for rimWidthMM := row.rimMinimum; rimWidthMM <= row.rimMaximum; rimWidthMM++ {
				t.Run(fmt.Sprintf("%s_%d_mm", system, rimWidthMM), func(t *testing.T) {
					for _, profile := range []TireRimApplicationProfile{TireRimApplicationProfileAutoByWidth, TireRimApplicationProfileMTBXC, TireRimApplicationProfileFatSnow} {
						result, err := ResolveTireRimWidthReferenceForRim(system, rimWidthMM, profile)
						if err != nil || result.NormalFitmentTireWidthRange != (WidthRange{row.tireMinimum, row.tireMaximum}) || result.FitmentRuleReference.Interpolated {
							t.Fatalf("ISO reverse range changed with system, purpose or interpolation: %#v error=%v", result, err)
						}
						calculation, err := CalculateTireRimEngineeringReference(system, rimWidthMM, row.tireMinimum)
						if err != nil || calculation.AllowedTireWidthRange != result.NormalFitmentTireWidthRange {
							t.Fatalf("exact and reverse normal ranges disagree: %#v error=%v", calculation, err)
						}
					}
				})
			}
		}
		for _, invalidRimWidthMM := range []int{12, 101} {
			result, err := ResolveTireRimWidthReferenceForRim(system, invalidRimWidthMM)
			if result != nil || !errors.Is(err, ErrInvalidTireRimEngineeringInnerWidth) {
				t.Fatalf("unsupported normal input was clamped: %#v error=%v", result, err)
			}
		}
	}
}

func TestISONormalRangeIntegrationPreservesExistingBestConflictForTheDeferredAlgorithmStep(t *testing.T) {
	result, err := ResolveTireRimWidthReference(34, RimSystemHookless)
	if err != nil || result.BestFitment == nil || result.BestFitment.RimInnerWidthMM != 26 || !reflect.DeepEqual(result.NormalFitmentRimWidthRanges, []WidthRange{{16, 25}}) {
		t.Fatalf("standard integration silently changed the deferred best algorithm: %#v error=%v", result, err)
	}
	calculation, err := CalculateTireRimEngineeringReference(RimSystemHookless, 26, 34)
	if err != nil || calculation.Verdict != TireRimEngineeringVerdictCritical || calculation.VerdictReason != TireRimEngineeringReasonBelowMinimum {
		t.Fatalf("preserving a project best must not broaden ISO normal limits: %#v error=%v", calculation, err)
	}
}

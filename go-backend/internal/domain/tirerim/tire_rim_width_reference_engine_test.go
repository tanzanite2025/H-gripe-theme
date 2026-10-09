package tirerim

import (
	"errors"
	"testing"
)

func TestResolveTireRimWidthReferenceReturnsExactRecommendedRow(t *testing.T) {
	result, err := ResolveTireRimWidthReference(32, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "engineering_recommended" {
		t.Fatalf("expected engineering recommendation, got %q", result.ResultKind)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 23, Max: 25}) {
		t.Fatalf("unexpected exact ranges: %#v", result.RimWidthRanges)
	}
	if result.Calculation != nil {
		t.Fatal("exact row should not include interpolation details")
	}
}

func TestResolveTireRimWidthReferenceMatchesThePublishedRowsShownByTheGuide(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		system     RimSystem
		expected   []WidthRange
		resultKind string
	}{
		{
			name:       "hookless 34 millimetres",
			width:      34,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 25, Max: 25}},
			resultKind: "engineering_reference",
		},
		{
			name:       "hookless 47 millimetres",
			width:      47,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 28, Max: 30}},
			resultKind: "engineering_reference",
		},
		{
			name:       "hookless 60 millimetres",
			width:      60,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 30, Max: 30}},
			resultKind: "engineering_reference",
		},
		{
			name:       "hooked 28 millimetres",
			width:      28,
			system:     RimSystemHooked,
			expected:   []WidthRange{{Min: 19, Max: 25}},
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

func TestGetTireRimWidthReferenceMatrixPreservesPossibleHooklessRowsAtSixtyAndSixtyFourMillimetres(t *testing.T) {
	rows, err := GetTireRimWidthReferenceMatrix(RimSystemHookless)
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	expectedPossibleRangesByTireWidth := map[int][]WidthRange{
		60: {{Min: 23, Max: 25}, {Min: 31, Max: 35}},
		64: {{Min: 23, Max: 25}, {Min: 26, Max: 27}, {Min: 36, Max: 40}},
	}
	for tireWidthMM, expectedRanges := range expectedPossibleRangesByTireWidth {
		var found *MatrixRow
		for index := range rows {
			if rows[index].TireWidthMM == tireWidthMM {
				found = &rows[index]
				break
			}
		}
		if found == nil {
			t.Fatalf("missing hookless %d mm row", tireWidthMM)
		}
		if len(found.Possible) != len(expectedRanges) {
			t.Fatalf("hookless %d mm possible ranges: expected %#v, got %#v", tireWidthMM, expectedRanges, found.Possible)
		}
		for index, expectedRange := range expectedRanges {
			if found.Possible[index] != expectedRange {
				t.Fatalf("hookless %d mm possible range %d: expected %#v, got %#v", tireWidthMM, index, expectedRange, found.Possible[index])
			}
		}
	}
}

func TestResolveTireRimWidthReferenceUsesEngineeringWindowInsteadOfInterpolation(t *testing.T) {
	result, err := ResolveTireRimWidthReference(31, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "engineering_recommended" {
		t.Fatalf("expected engineering recommendation, got %q", result.ResultKind)
	}
	if result.Calculation != nil || len(result.SourceRows) != 0 {
		t.Fatalf("engineering result must not include DT interpolation metadata: calculation=%#v source_rows=%#v", result.Calculation, result.SourceRows)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 23, Max: 25}) {
		t.Fatalf("unexpected engineering ranges: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceUsesTheEngineeringRecommendationWhenSeveralInnerWidthsMatch(t *testing.T) {
	result, err := ResolveTireRimWidthReference(30, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "engineering_recommended" {
		t.Fatalf("expected engineering recommendation, got %q", result.ResultKind)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 21, Max: 25}) {
		t.Fatalf("unexpected engineering range: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceRejectsWidthsOutsideTheEngineeringModel(t *testing.T) {
	result, err := ResolveTireRimWidthReference(96, RimSystemHookless)
	if result != nil || !errors.Is(err, ErrNoPublishedBracket) {
		t.Fatalf("expected no engineering bracket, result=%#v error=%v", result, err)
	}
}

func TestResolveTireRimWidthReferenceReportsDisplayOnlyPhysicalProjectionFromTheSelectedRange(t *testing.T) {
	result, err := ResolveTireRimWidthReference(32, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.DerivedMetrics.Status != "engineering" {
		t.Fatalf("unexpected derived metric status: %q", result.DerivedMetrics.Status)
	}
	if result.DerivedMetrics.InflatedTireWidthMM.Min != 33.6 || result.DerivedMetrics.InflatedTireWidthMM.Max != 34.4 {
		t.Fatalf("unexpected inflated-width projection: %#v", result.DerivedMetrics.InflatedTireWidthMM)
	}
	if result.DerivedMetrics.AeroTargetOuterWidthMM.Min != 35.3 || result.DerivedMetrics.AeroTargetOuterWidthMM.Max != 36.1 {
		t.Fatalf("unexpected 105%% projection: %#v", result.DerivedMetrics.AeroTargetOuterWidthMM)
	}
}

func TestResolveTireRimWidthReferenceRejectsWidthsWithoutAChartBracket(t *testing.T) {
	_, err := ResolveTireRimWidthReference(18, RimSystemHookless)
	if !errors.Is(err, ErrNoPublishedBracket) {
		t.Fatalf("expected no bracket error, got %v", err)
	}
}

func TestGetTireRimWidthReferenceMetadataClonesRows(t *testing.T) {
	metadata := GetTireRimWidthReferenceMetadata()
	if metadata.ModelVersion != ModelVersion || len(metadata.Rows) == 0 {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	metadata.Rows[0].Recommended = append(metadata.Rows[0].Recommended, WidthRange{1, 1})
	rows, err := GetTireRimWidthReferenceMatrix(RimSystemHookless)
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	for _, row := range rows {
		for _, value := range row.Recommended {
			if value == (WidthRange{1, 1}) {
				t.Fatal("metadata mutation leaked into matrix")
			}
		}
	}
}

func TestCalculateTireRimEngineeringReferenceIncludesThe105PercentAeroProjection(t *testing.T) {
	result, err := CalculateTireRimEngineeringReference(RimSystemHookless, 25, 32)
	if err != nil {
		t.Fatalf("calculate engineering reference: %v", err)
	}
	if result.Verdict != TireRimEngineeringVerdictRecommended {
		t.Fatalf("expected recommended verdict, got %q", result.Verdict)
	}
	if result.InflatedTireWidthMM != 34.4 || result.AeroTargetOuterWidthMM != 36.1 {
		t.Fatalf("unexpected aero projection: inflated=%v target=%v", result.InflatedTireWidthMM, result.AeroTargetOuterWidthMM)
	}
	if result.MaximumPressure.Bar != 4.5 || result.MaximumPressure.PSI != 65 {
		t.Fatalf("unexpected pressure limit: %#v", result.MaximumPressure)
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
	_, err := CalculateTireRimEngineeringReference(RimSystemHookless, 24, 28)
	if !errors.Is(err, ErrInvalidTireRimEngineeringInnerWidth) {
		t.Fatalf("expected unsupported inner-width error, got %v", err)
	}
}

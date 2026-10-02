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
	if result.ResultKind != "exact_recommended" {
		t.Fatalf("expected exact recommendation, got %q", result.ResultKind)
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
			expected:   []WidthRange{{Min: 23, Max: 25}},
			resultKind: "exact_recommended",
		},
		{
			name:       "hookless 47 millimetres",
			width:      47,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 23, Max: 25}, {Min: 26, Max: 27}},
			resultKind: "exact_recommended",
		},
		{
			name:       "hookless 60 millimetres",
			width:      60,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 26, Max: 27}, {Min: 28, Max: 30}},
			resultKind: "exact_recommended",
		},
		{
			name:       "hookless 64 millimetres",
			width:      64,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 28, Max: 30}, {Min: 31, Max: 35}},
			resultKind: "exact_recommended",
		},
		{
			name:       "hookless 102 millimetres",
			width:      102,
			system:     RimSystemHookless,
			expected:   []WidthRange{{Min: 76, Max: 76}},
			resultKind: "exact_recommended",
		},
		{
			name:       "hooked 28 millimetres",
			width:      28,
			system:     RimSystemHooked,
			expected:   []WidthRange{{Min: 18, Max: 20}, {Min: 21, Max: 22}},
			resultKind: "exact_recommended",
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

func TestResolveTireRimWidthReferenceInterpolatesTheGapInsteadOfRejectingIt(t *testing.T) {
	result, err := ResolveTireRimWidthReference(31, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "interpolated" {
		t.Fatalf("expected interpolation, got %q", result.ResultKind)
	}
	if result.Calculation == nil || result.Calculation.LowerTireWidthMM != 30 || result.Calculation.UpperTireWidthMM != 32 {
		t.Fatalf("unexpected interpolation metadata: %#v", result.Calculation)
	}
	if len(result.SourceRows) != 2 || result.SourceRows[0].Kind != "possible_reference" || result.SourceRows[1].Kind != "recommended" {
		t.Fatalf("unexpected interpolation source rows: %#v", result.SourceRows)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 23, Max: 25}) {
		t.Fatalf("unexpected interpolated ranges: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceKeepsPossibleOnlyRowsLabeledAsReference(t *testing.T) {
	result, err := ResolveTireRimWidthReference(30, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.ResultKind != "possible_reference" {
		t.Fatalf("expected possible reference, got %q", result.ResultKind)
	}
	if result.SourceRows[0].Kind != "possible_reference" {
		t.Fatalf("possible row was relabeled: %#v", result.SourceRows)
	}
}

func TestResolveTireRimWidthReferenceDoesNotApplyAnArbitraryTenMillimetreCutoff(t *testing.T) {
	result, err := ResolveTireRimWidthReference(96, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(result.RimWidthRanges) != 1 || result.RimWidthRanges[0] != (WidthRange{Min: 76, Max: 76}) {
		t.Fatalf("unexpected wide-rim result: %#v", result.RimWidthRanges)
	}
}

func TestResolveTireRimWidthReferenceReportsDisplayOnlyPhysicalProjectionFromTheSelectedRange(t *testing.T) {
	result, err := ResolveTireRimWidthReference(32, RimSystemHookless)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.DerivedMetrics.Status != "display_only" {
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
	_, err := ResolveTireRimWidthReference(29, RimSystemHookless)
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

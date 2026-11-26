package product

import (
	"encoding/json"
	"testing"
)

func TestWheelsetSpecificationAllowsDifferentFrontAndRearSizesAndMissingOptionalFields(t *testing.T) {
	specification := &WheelsetSpecification{
		SchemaVersion: 1,
		FrontWheel: WheelsetProductWheelSpecification{
			WheelSize:       "700C",
			SpokeHoleCounts: []int{24, 28},
			LacingPatterns:  []string{"2x", "3x"},
		},
		RearWheel: WheelsetProductWheelSpecification{
			WheelSize:       "650B",
			SpokeHoleCounts: []int{28, 32},
			LacingPatterns:  []string{"3x"},
		},
	}

	if err := ValidateWheelsetSpecification(specification); err != nil {
		t.Fatalf("expected valid wheelset specification: %v", err)
	}
	if specification.FrontWheel.WheelSize == specification.RearWheel.WheelSize {
		t.Fatal("test specification must keep front and rear wheel sizes distinct")
	}
}

func TestValidateWheelsetSpecificationRequiresFrontAndRearProductFacts(t *testing.T) {
	valid := WheelsetSpecification{
		SchemaVersion: 1,
		FrontWheel: WheelsetProductWheelSpecification{
			WheelSize: "700C", SpokeHoleCounts: []int{24}, LacingPatterns: []string{"2x"},
		},
		RearWheel: WheelsetProductWheelSpecification{
			WheelSize: "700C", SpokeHoleCounts: []int{28}, LacingPatterns: []string{"3x"},
		},
	}
	tests := []struct {
		name   string
		mutate func(*WheelsetSpecification)
	}{
		{"front wheel size", func(spec *WheelsetSpecification) { spec.FrontWheel.WheelSize = " " }},
		{"rear wheel size", func(spec *WheelsetSpecification) { spec.RearWheel.WheelSize = "" }},
		{"front spoke hole counts", func(spec *WheelsetSpecification) { spec.FrontWheel.SpokeHoleCounts = nil }},
		{"rear spoke hole counts", func(spec *WheelsetSpecification) { spec.RearWheel.SpokeHoleCounts = nil }},
		{"front lacing patterns", func(spec *WheelsetSpecification) { spec.FrontWheel.LacingPatterns = nil }},
		{"rear lacing patterns", func(spec *WheelsetSpecification) { spec.RearWheel.LacingPatterns = nil }},
		{"unsupported schema version", func(spec *WheelsetSpecification) { spec.SchemaVersion = 2 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			specification := valid
			specification.FrontWheel.SpokeHoleCounts = append([]int(nil), valid.FrontWheel.SpokeHoleCounts...)
			specification.FrontWheel.LacingPatterns = append([]string(nil), valid.FrontWheel.LacingPatterns...)
			specification.RearWheel.SpokeHoleCounts = append([]int(nil), valid.RearWheel.SpokeHoleCounts...)
			specification.RearWheel.LacingPatterns = append([]string(nil), valid.RearWheel.LacingPatterns...)
			test.mutate(&specification)
			if err := ValidateWheelsetSpecification(&specification); err == nil {
				t.Fatal("expected invalid wheelset specification")
			}
		})
	}
}

func TestValidateWheelsetSpecificationRejectsInvalidOptionalValues(t *testing.T) {
	negativeWidth := -1.0
	tooNarrowExternalWidth := 20.0
	internalWidth := 25.0
	valid := func() WheelsetSpecification {
		return WheelsetSpecification{
			SchemaVersion: 1,
			FrontWheel: WheelsetProductWheelSpecification{
				WheelSize: "700C", SpokeHoleCounts: []int{24}, LacingPatterns: []string{"2x"},
			},
			RearWheel: WheelsetProductWheelSpecification{
				WheelSize: "700C", SpokeHoleCounts: []int{28}, LacingPatterns: []string{"3x"},
			},
		}
	}
	tests := []struct {
		name   string
		mutate func(*WheelsetSpecification)
	}{
		{"non-positive optional measurement", func(spec *WheelsetSpecification) { spec.FrontWheel.RimDepthMM = &negativeWidth }},
		{"freehub body belongs only to rear wheel", func(spec *WheelsetSpecification) { spec.FrontWheel.FreehubBody = "XDR" }},
		{"external width must exceed internal width", func(spec *WheelsetSpecification) {
			spec.FrontWheel.RimInternalWidthMM = &internalWidth
			spec.FrontWheel.RimExternalWidthMM = &tooNarrowExternalWidth
		}},
		{"duplicate hole counts", func(spec *WheelsetSpecification) { spec.FrontWheel.SpokeHoleCounts = []int{24, 24} }},
		{"blank lacing pattern", func(spec *WheelsetSpecification) { spec.RearWheel.LacingPatterns = []string{" "} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			specification := valid()
			test.mutate(&specification)
			if err := ValidateWheelsetSpecification(&specification); err == nil {
				t.Fatal("expected invalid wheelset specification")
			}
		})
	}
}

func TestWheelsetSpecificationRejectsUnknownFieldsAtEveryLevel(t *testing.T) {
	for _, raw := range []string{
		`{"schema_version":1,"unknown":true}`,
		`{"schema_version":1,"front_wheel":{"unknown":true}}`,
	} {
		var specification WheelsetSpecification
		if err := json.Unmarshal([]byte(raw), &specification); err == nil {
			t.Fatalf("expected unknown field to be rejected: %s", raw)
		}
	}
}

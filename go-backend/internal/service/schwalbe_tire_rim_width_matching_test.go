package service

import (
	"testing"

	"commerce-platform/internal/repository"
)

func TestSchwalbeTireRimWidthGuidanceIncludesRuleBoundaries(t *testing.T) {
	tireWidth := 40
	innerWidth := 17.0
	rules := []repository.SchwalbeTireRimWidthCombinationRule{{
		TireWidthMinMM:     35,
		TireWidthMaxMM:     46,
		InnerRimWidthMinMM: 17,
		InnerRimWidthMaxMM: 27,
	}}

	if matches := schwalbeTireRimWidthGuidanceFor(&tireWidth, &innerWidth, rules); len(matches) != 1 {
		t.Fatalf("expected the lower inner-width boundary to match, got %+v", matches)
	}
	innerWidth = 27
	if matches := schwalbeTireRimWidthGuidanceFor(&tireWidth, &innerWidth, rules); len(matches) != 1 {
		t.Fatalf("expected the upper inner-width boundary to match, got %+v", matches)
	}
	innerWidth = 27.01
	if matches := schwalbeTireRimWidthGuidanceFor(&tireWidth, &innerWidth, rules); len(matches) != 0 {
		t.Fatalf("expected a value outside the inclusive range to be excluded, got %+v", matches)
	}
}

func TestSchwalbeTireRimWidthGuidanceRequiresParsedETRTO(t *testing.T) {
	innerWidth := 23.0
	rules := []repository.SchwalbeTireRimWidthCombinationRule{{
		TireWidthMinMM:     28,
		TireWidthMaxMM:     28,
		InnerRimWidthMinMM: 16,
		InnerRimWidthMaxMM: 23,
	}}

	if matches := schwalbeTireRimWidthGuidanceFor(nil, &innerWidth, rules); len(matches) != 0 {
		t.Fatalf("expected a missing ETRTO width to be excluded, got %+v", matches)
	}
	if matches := schwalbeTireRimWidthGuidanceFor(new(int), &innerWidth, rules); len(matches) != 0 {
		t.Fatalf("expected a non-positive ETRTO width to be excluded, got %+v", matches)
	}
}

func TestSchwalbeTireRimWidthGuidanceForTireWidthReturnsReferenceRange(t *testing.T) {
	tireWidth := 40
	rules := []repository.SchwalbeTireRimWidthCombinationRule{
		{
			TireWidthMinMM:     35,
			TireWidthMaxMM:     46,
			InnerRimWidthMinMM: 17,
			InnerRimWidthMaxMM: 27,
		},
		{
			TireWidthMinMM:     47,
			TireWidthMaxMM:     57,
			InnerRimWidthMinMM: 17,
			InnerRimWidthMaxMM: 30,
		},
	}

	matches := schwalbeTireRimWidthGuidanceForTireWidth(&tireWidth, rules)
	if len(matches) != 1 || matches[0].TireWidthMinMM != 35 ||
		matches[0].TireWidthMaxMM != 46 || matches[0].InnerRimWidthMinMM != 17 ||
		matches[0].InnerRimWidthMaxMM != 27 {
		t.Fatalf("expected the reference range for the parsed tire width, got %+v", matches)
	}
}

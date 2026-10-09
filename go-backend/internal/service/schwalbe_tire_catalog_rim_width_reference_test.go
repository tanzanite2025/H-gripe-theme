package service

import (
	"testing"

	"commerce-platform/internal/domain/tirerim"
	"commerce-platform/internal/repository"
)

func TestSchwalbeTireCatalogRimWidthReferenceForRowsUsesBackendEngineeringWindow(t *testing.T) {
	nominalWidthMM := 31
	rows := []schwalbeTireCatalogSelectorRow{{
		item: repository.SchwalbeTireCatalogItem{
			ArticleNo: "TEST-31",
			ModelName: "Reviewed model",
			ETRTO:     "31-622",
		},
		dimensions: schwalbeTireCatalogSelectorDimensions{
			NominalTireWidthMM: &nominalWidthMM,
		},
	}}
	compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
		"TEST-31": {{RimSystem: "hooked", Status: "supported"}},
	}

	guidanceByArticle := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility)
	guidance := guidanceByArticle["TEST-31"]
	if len(guidance) != 1 {
		t.Fatalf("expected one hooked reference result, got %#v", guidance)
	}
	if guidance[0].RimSystem != tirerim.RimSystemHooked || guidance[0].ResultKind != "engineering_recommended" {
		t.Fatalf("expected backend hooked engineering recommendation, got %#v", guidance[0])
	}
	if len(guidance[0].RimWidthRanges) != 1 || guidance[0].RimWidthRanges[0] != (tirerim.WidthRange{Min: 23, Max: 25}) {
		t.Fatalf("unexpected engineering hooked range: %#v", guidance[0].RimWidthRanges)
	}
}

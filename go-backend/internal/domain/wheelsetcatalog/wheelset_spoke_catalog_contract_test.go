package wheelsetcatalog

import "testing"

func TestValidateWheelsetCatalogIdentifiersRejectsDuplicateIdentifiers(t *testing.T) {
	duplicateBrand := []Brand{
		{BrandSlug: "DT-Swiss"},
		{BrandSlug: "dt-swiss"},
	}
	if err := validateWheelsetCatalogIdentifiers(duplicateBrand); err == nil {
		t.Fatal("expected duplicate brand slug to be rejected")
	}

	duplicateModel := []Brand{
		{
			BrandSlug: "brand-a",
			Wheelsets: []Wheelset{{Slug: "Model-A"}},
		},
		{
			BrandSlug: "brand-b",
			Wheelsets: []Wheelset{{Slug: "model-a"}},
		},
	}
	if err := validateWheelsetCatalogIdentifiers(duplicateModel); err == nil {
		t.Fatal("expected duplicate wheelset slug to be rejected")
	}
}

func TestPublishedModelsCarryBrandSourceCheckedAt(t *testing.T) {
	models, err := ListPublishedWheelsetSpokeCatalogModels()
	if err != nil {
		t.Fatalf("load published models: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected at least one published wheelset model")
	}
	for _, model := range models {
		if model.SourceCheckedAt == nil || *model.SourceCheckedAt == "" {
			t.Fatalf("published model %q has no source checked date", model.Slug)
		}
	}
}

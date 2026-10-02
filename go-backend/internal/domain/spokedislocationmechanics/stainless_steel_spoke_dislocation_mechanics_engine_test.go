package spokedislocationmechanics

import (
	"errors"
	"testing"
)

func TestCalculateStainlessSteelSpokeDislocationMechanicsUsesBackendModelPreset(t *testing.T) {
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "cx-ray",
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   135,
	})
	if err != nil {
		t.Fatalf("expected calculation to succeed, got %v", err)
	}
	if result.EffectiveAreaMM2 != 1.56 {
		t.Fatalf("expected backend CX-Ray area 1.56, got %.2f", result.EffectiveAreaMM2)
	}
	if result.WorkStressMPA != 769 || result.WorkYieldRatioPercent != 54.9 {
		t.Fatalf("unexpected working stress result: %+v", result)
	}
	if result.OverloadForceN != 1620 || result.OverloadStressMPA != 1038 || result.OverloadYieldRatioPercent != 74.2 {
		t.Fatalf("unexpected overload result: %+v", result)
	}
	if result.WorkElasticElongationMM != 1.156 || result.OverloadElasticElongationMM != 1.56 || result.OverloadDeltaElasticElongationMM != 0.405 {
		t.Fatalf("unexpected elastic elongation result: %+v", result)
	}
	if result.SafetyLevel != StainlessSteelSpokeDislocationMechanicsSafetyReference {
		t.Fatalf("expected reference safety level, got %s", result.SafetyLevel)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsSupportsCustomArea(t *testing.T) {
	customArea := 2.0
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "custom",
		CustomEffectiveAreaMM2: &customArea,
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   135,
	})
	if err != nil {
		t.Fatalf("expected custom calculation to succeed, got %v", err)
	}
	if result.EffectiveAreaMM2 != customArea || result.OverloadStressMPA != 810 {
		t.Fatalf("unexpected custom area result: %+v", result)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsAcceptsTwoHundredMillimeterLoadedLength(t *testing.T) {
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "cx-ray",
		EffectiveSpokeLengthMM: 200,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   135,
	})
	if err != nil {
		t.Fatalf("the screenshot values should be accepted by the current backend model, got %v", err)
	}
	if result.EffectiveSpokeLengthMM != 200 || result.WorkElasticElongationMM != 0.797 || result.OverloadElasticElongationMM != 1.076 || result.OverloadDeltaElasticElongationMM != 0.279 {
		t.Fatalf("unexpected 200 mm elongation result: %+v", result)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsReturnsCriticalSafetyState(t *testing.T) {
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "leader",
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1400,
		OverloadRatioPercent:   150,
	})
	if err != nil {
		t.Fatalf("expected unsafe state to be returned for display, got %v", err)
	}
	if result.OverloadForceN != 2100 || result.SafetyLevel != StainlessSteelSpokeDislocationMechanicsSafetyCritical {
		t.Fatalf("expected critical unsafe result, got %+v", result)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsReturnsWarningSafetyState(t *testing.T) {
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "leader",
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   140,
	})
	if err != nil {
		t.Fatalf("expected warning state to be returned, got %v", err)
	}
	if result.OverloadForceN != 1680 || result.SafetyLevel != StainlessSteelSpokeDislocationMechanicsSafetyWarning {
		t.Fatalf("expected warning result, got %+v", result)
	}
}

func TestValidateStainlessSteelSpokeDislocationMechanicsRejectsInvalidPhysicalInputs(t *testing.T) {
	invalidRequests := []StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		{ModelID: "cx-ray", EffectiveSpokeLengthMM: 290, NominalWorkingTensionN: 0, OverloadRatioPercent: 135},
		{ModelID: "cx-ray", EffectiveSpokeLengthMM: 0, NominalWorkingTensionN: 1200, OverloadRatioPercent: 135},
		{ModelID: "cx-ray", EffectiveSpokeLengthMM: 290, NominalWorkingTensionN: 1200, OverloadRatioPercent: 99},
		{ModelID: "unknown", EffectiveSpokeLengthMM: 290, NominalWorkingTensionN: 1200, OverloadRatioPercent: 135},
	}
	for index, request := range invalidRequests {
		if err := ValidateStainlessSteelSpokeDislocationMechanicsCalculationRequest(request); err == nil {
			t.Fatalf("request %d should fail validation", index)
		}
	}

	if err := ValidateStainlessSteelSpokeDislocationMechanicsCalculationRequest(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "custom",
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   135,
	}); !errors.Is(err, ErrCustomSpokeAreaRequired) {
		t.Fatalf("expected custom area error, got %v", err)
	}
}

func TestGetStainlessSteelSpokeDislocationMechanicsMetadataExposesInputEnvelope(t *testing.T) {
	metadata := GetStainlessSteelSpokeDislocationMechanicsMetadata()
	if metadata.ModelVersion != StainlessSteelSpokeDislocationMechanicsModelVersion {
		t.Fatalf("unexpected model version %q", metadata.ModelVersion)
	}
	if len(metadata.Models) != 12 {
		t.Fatalf("expected twelve backend model presets, got %d", len(metadata.Models))
	}
	if metadata.PhysicalReferences.CarbonRimSpokeHoleWarningForceN != 1850 || metadata.PhysicalReferences.ElasticModulusMPA != 193000 || metadata.InputBounds.EffectiveSpokeLengthMaxMM != 1000 || metadata.InputBounds.OverloadRatioMaxPercent != 200 {
		t.Fatalf("unexpected backend metadata: %+v", metadata)
	}
}

func TestStainlessSteelSpokeMaterialReferenceCatalogIsVersionedAndTraceable(t *testing.T) {
	catalog := GetStainlessSteelSpokeMaterialReferenceCatalog()
	if catalog.CatalogID != "stainless-steel-spoke-material-reference-catalog" || catalog.CatalogVersion != "1.0.2" {
		t.Fatalf("unexpected material catalog identity: %+v", catalog)
	}
	material, found := GetStainlessSteelSpokeMaterialReferenceByID(catalog.DefaultMaterialReferenceID)
	if !found {
		t.Fatalf("expected default material reference %q", catalog.DefaultMaterialReferenceID)
	}
	if material.ElasticModulusMPA != 193000 || material.MacroYieldReferenceMPA != 1400 || material.UltimateTensileStrengthMPA != 1600 {
		t.Fatalf("unexpected material reference values: %+v", material)
	}
	if material.DataStatus != "generic-materials-science-reference" || material.DataVersion == "" || len(material.SourceReferences) < 2 {
		t.Fatalf("material reference must expose revision status and sources: %+v", material)
	}
}

func TestStainlessSteelSpokeModelsShareOnlyTheCurrentGenericMaterialReference(t *testing.T) {
	defaultMaterialReferenceID := GetStainlessSteelSpokeMaterialReferenceCatalog().DefaultMaterialReferenceID
	for _, model := range ListStainlessSteelSpokeDislocationMechanicsModelReferences() {
		if model.MaterialReferenceID != defaultMaterialReferenceID {
			t.Fatalf("model %s unexpectedly maps to material reference %q", model.ID, model.MaterialReferenceID)
		}
	}
}

func TestStainlessSteelSpokeModelReferenceCatalogContainsCalculationGeometryOnly(t *testing.T) {
	model, found := GetStainlessSteelSpokeDislocationMechanicsModelReferenceByID("cx-ray")
	if !found {
		t.Fatal("expected SAPIM CX-Ray model reference")
	}
	if model.DataStatus != "catalog-geometry-reference" || model.DataVersion != "1.2.0" {
		t.Fatalf("unexpected SAPIM model data revision: %+v", model)
	}
	if model.EffectiveAreaMM2 != 1.56 || model.MaterialReferenceID == "" {
		t.Fatalf("expected SAPIM CX-Ray calculation geometry, got %+v", model)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsReturnsSAPIMModelReferenceWithResult(t *testing.T) {
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "cx-ray",
		EffectiveSpokeLengthMM: 200,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   135,
	})
	if err != nil {
		t.Fatalf("expected SAPIM calculation to succeed, got %v", err)
	}
	if result.ModelReference.ID != "cx-ray" || result.ModelReference.EffectiveAreaMM2 != 1.56 {
		t.Fatalf("expected result to carry the SAPIM backend model reference, got %+v", result.ModelReference)
	}
	if result.WorkElasticElongationMM != 0.797 || result.OverloadElasticElongationMM != 1.076 {
		t.Fatalf("unexpected SAPIM elastic elongation result: %+v", result)
	}
}

func TestStainlessSteelSpokeDislocationMechanicsIncludesDTSwissPublishedCentreGaugePresets(t *testing.T) {
	expectedAreas := map[string]float64{
		"dt-aerolite":    1.77,
		"dt-aero-comp":   2.54,
		"dt-revolite":    1.94,
		"dt-revolution":  1.77,
		"dt-competition": 2.54,
		"dt-champion":    3.14,
	}

	models := ListStainlessSteelSpokeDislocationMechanicsModelReferences()
	seenDTModels := make(map[string]bool, len(expectedAreas))
	for _, model := range models {
		expectedArea, isDTModel := expectedAreas[model.ID]
		if !isDTModel {
			continue
		}
		seenDTModels[model.ID] = true
		if model.EffectiveAreaMM2 != expectedArea {
			t.Fatalf("expected DT Swiss model %s to expose %.2f mm², got %.2f", model.ID, expectedArea, model.EffectiveAreaMM2)
		}
	}
	if len(seenDTModels) != len(expectedAreas) {
		t.Fatalf("expected all six DT Swiss presets, found %d", len(seenDTModels))
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsAcceptsDTAerolitePreset(t *testing.T) {
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "dt-aerolite",
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   135,
	})
	if err != nil {
		t.Fatalf("expected DT Aerolite calculation to succeed, got %v", err)
	}
	if result.EffectiveAreaMM2 != 1.77 || result.WorkStressMPA != 678 || result.OverloadStressMPA != 915 {
		t.Fatalf("unexpected DT Aerolite calculation result: %+v", result)
	}
}

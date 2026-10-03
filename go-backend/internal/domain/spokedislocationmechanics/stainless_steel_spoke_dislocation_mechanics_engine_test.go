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
	if result.WorkStressMPA != 769 || result.WorkYieldRatioPercent != 67.7 {
		t.Fatalf("unexpected working stress result: %+v", result)
	}
	if result.OverloadForceN != 1620 || result.OverloadStressMPA != 1038 || result.OverloadYieldRatioPercent != 91.3 {
		t.Fatalf("unexpected overload result: %+v", result)
	}
	if result.WorkTotalElongationMM != 1.156 || result.WorkElasticElongationMM != 1.156 || result.WorkPermanentElongationMM != 0 || result.OverloadTotalElongationMM != 1.56 || result.OverloadElasticElongationMM != 1.56 || result.OverloadPermanentElongationMM != 0 || result.OverloadDeltaElasticElongationMM != 0.405 || result.OverloadDeltaTotalElongationMM != 0.405 || result.OverloadDeltaPermanentElongationMM != 0 || result.FractureReferenceTotalElongationMM != 14.5 {
		t.Fatalf("unexpected elastic elongation result: %+v", result)
	}
	if result.SafetyLevel != StainlessSteelSpokeDislocationMechanicsSafetyWarning {
		t.Fatalf("expected warning safety level from the AISI 304 curve yield ratio, got %s", result.SafetyLevel)
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
	if result.EffectiveSpokeLengthMM != 200 || result.WorkTotalElongationMM != 0.797 || result.WorkElasticElongationMM != 0.797 || result.OverloadTotalElongationMM != 1.076 || result.OverloadElasticElongationMM != 1.076 || result.OverloadDeltaTotalElongationMM != 0.279 || result.FractureReferenceTotalElongationMM != 10 {
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

func TestGetStainlessSteelSpokeDislocationMechanicsMetadataIncludesFixedReferenceResults(t *testing.T) {
	metadata := GetStainlessSteelSpokeDislocationMechanicsMetadata()
	reference := metadata.ReferenceCalculation
	if reference.EffectiveSpokeLengthMM != 270 || reference.ReferenceCalculationTensionKGF != 180 || reference.ReferenceCalculationTensionN != 1765.197 || reference.YieldStrengthMPA != 1137 || reference.CurveEndpointStressMPA != 1346 || reference.CurveEndpointTotalStrainPercent != 3 || reference.CurveEndpointTotalElongationMM != 8.1 || reference.TotalElongationToFailurePercent != 5 || reference.FractureReferenceTotalElongationMM != 13.5 {
		t.Fatalf("unexpected public reference baseline: %+v", reference)
	}
	if len(reference.Results) != len(metadata.Models) {
		t.Fatalf("expected one reference result per model, got %d results for %d models", len(reference.Results), len(metadata.Models))
	}
	expectedReferenceByModelID := map[string][3]float64{
		"cx-ray":         {1774, 2100, 1.56},
		"pillar-wing-20": {1842, 2181, 1.62},
		"laser":          {2012, 2382, 1.77},
		"d-light":        {2433, 2880, 2.14},
		"race":           {2888, 3419, 2.54},
		"leader":         {3570, 4226, 3.14},
		"dt-aerolite":    {2012, 2382, 1.77},
		"dt-aero-comp":   {2888, 3419, 2.54},
		"dt-revolite":    {2206, 2611, 1.94},
		"dt-revolution":  {2012, 2382, 1.77},
		"dt-competition": {2888, 3419, 2.54},
		"dt-champion":    {3570, 4226, 3.14},
	}
	for _, result := range reference.Results {
		expected, exists := expectedReferenceByModelID[result.ModelID]
		if !exists {
			t.Fatalf("unexpected reference model %q", result.ModelID)
		}
		if result.YieldReferenceForceN != expected[0] || result.CurveEndpointForceN != expected[1] || result.EffectiveAreaMM2 != expected[2] || result.FractureReferenceTotalElongationMM != 13.5 {
			t.Fatalf("unexpected fixed reference result for %s: %+v", result.ModelID, result)
		}
	}
	expectedWorkingElongationByModelID := map[string]float64{
		"cx-ray":         1.583,
		"pillar-wing-20": 1.524,
		"laser":          1.395,
		"d-light":        1.154,
		"race":           0.972,
		"leader":         0.786,
		"dt-aerolite":    1.395,
		"dt-aero-comp":   0.972,
		"dt-revolite":    1.273,
		"dt-revolution":  1.395,
		"dt-competition": 0.972,
		"dt-champion":    0.786,
	}
	for _, result := range reference.Results {
		if result.ReferenceCalculationTotalElongationMM != expectedWorkingElongationByModelID[result.ModelID] {
			t.Fatalf("expected 180 kgf allowance-reference elongation to vary by model section for %s, got %+v", result.ModelID, result)
		}
		if result.ReferenceCalculationPermanentElongationMM != 0 {
			t.Fatalf("expected the 180 kgf reference to remain on the elastic branch for %s, got %+v", result.ModelID, result)
		}
	}
}

func TestStainlessSteelSpokeMaterialReferenceCatalogIsVersionedAndTraceable(t *testing.T) {
	catalog := GetStainlessSteelSpokeMaterialReferenceCatalog()
	if catalog.CatalogID != "stainless-steel-spoke-material-reference-catalog" || catalog.CatalogVersion != "2.1.0" {
		t.Fatalf("unexpected material catalog identity: %+v", catalog)
	}
	material, found := GetStainlessSteelSpokeMaterialReferenceByID(catalog.DefaultMaterialReferenceID)
	if !found {
		t.Fatalf("expected default material reference %q", catalog.DefaultMaterialReferenceID)
	}
	if material.ID != "aisi-304-cold-drawn-wire-true-strain-0-585" || material.ElasticModulusMPA != 193000 || material.YieldStrengthMPA != 1137 || material.UltimateTensileStrengthMPA != 1346 || material.TotalElongationToFailurePercent != 5 {
		t.Fatalf("unexpected material reference values: %+v", material)
	}
	if material.DataStatus != "published-curve-reference-digitized-from-figure" || material.DataVersion != "2.1.0" || len(material.SourceReferences) < 3 || len(material.StressStrainCurve) != 10 {
		t.Fatalf("material reference must expose revision status and sources: %+v", material)
	}
	if material.StressStrainCurve[0].EngineeringStrain != 0 || material.StressStrainCurve[0].EngineeringStressMPA != 0 {
		t.Fatalf("material curve must start at zero: %+v", material.StressStrainCurve[0])
	}
	lastCurvePoint := material.StressStrainCurve[len(material.StressStrainCurve)-1]
	if lastCurvePoint.EngineeringStressMPA != material.UltimateTensileStrengthMPA {
		t.Fatalf("material curve endpoint must match the published ultimate strength: %+v", lastCurvePoint)
	}
}

func TestStainlessSteelSpokeModelsUseTheSelectedAISI304MaterialReference(t *testing.T) {
	defaultMaterialReferenceID := GetStainlessSteelSpokeMaterialReferenceCatalog().DefaultMaterialReferenceID
	if defaultMaterialReferenceID != "aisi-304-cold-drawn-wire-true-strain-0-585" {
		t.Fatalf("expected the selected AISI 304 material reference, got %q", defaultMaterialReferenceID)
	}
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
	if model.DataStatus != "catalog-geometry-reference" || model.DataVersion != "1.3.0" {
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
	if result.WorkTotalElongationMM != 0.797 || result.WorkElasticElongationMM != 0.797 || result.OverloadTotalElongationMM != 1.076 || result.OverloadElasticElongationMM != 1.076 {
		t.Fatalf("unexpected SAPIM elastic elongation result: %+v", result)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsInterpolatesAISI304PlasticStrain(t *testing.T) {
	customArea := 1.0
	result, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "custom",
		CustomEffectiveAreaMM2: &customArea,
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1210,
		OverloadRatioPercent:   100,
	})
	if err != nil {
		t.Fatalf("expected the in-curve AISI 304 calculation to succeed, got %v", err)
	}
	if result.WorkStressMPA != 1210 || result.WorkTotalStrainPercent != 1 || result.WorkElasticStrainPercent != 0.627 || result.WorkPlasticStrainPercent != 0.373 {
		t.Fatalf("unexpected AISI 304 curve state: %+v", result)
	}
	if result.WorkTotalElongationMM != 2.9 || result.WorkElasticElongationMM != 1.818 || result.WorkPermanentElongationMM != 1.082 {
		t.Fatalf("expected total, elastic, and permanent elongation from the curve, got %+v", result)
	}
	if result.OverloadPermanentElongationMM != result.WorkPermanentElongationMM || result.OverloadDeltaPermanentElongationMM != 0 {
		t.Fatalf("expected identical working and 100%% overload permanent elongation, got %+v", result)
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsRejectsStressOutsideAISI304Curve(t *testing.T) {
	customArea := 1.0
	_, err := CalculateStainlessSteelSpokeDislocationMechanics(StainlessSteelSpokeDislocationMechanicsCalculationRequest{
		ModelID:                "custom",
		CustomEffectiveAreaMM2: &customArea,
		EffectiveSpokeLengthMM: 290,
		NominalWorkingTensionN: 1200,
		OverloadRatioPercent:   120,
	})
	if !errors.Is(err, ErrStainlessSteelSpokeStressOutsideMaterialCurve) {
		t.Fatalf("expected stress outside the AISI 304 curve to be rejected, got %v", err)
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

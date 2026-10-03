package spokedislocationmechanics

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// StainlessSteelSpokeDislocationMechanicsModelVersion identifies the backend
// calculation model. Material values are versioned independently in the
// stainless steel spoke material reference catalog.
const StainlessSteelSpokeDislocationMechanicsModelVersion = "stainless-steel-spoke-dislocation-mechanics-v3.0"

const (
	minimumNominalWorkingTensionN = 0.0
	maximumNominalWorkingTensionN = 3000.0
	minimumEffectiveSpokeAreaMM2  = 0.2
	maximumEffectiveSpokeAreaMM2  = 10.0
	minimumEffectiveSpokeLengthMM = 1.0
	maximumEffectiveSpokeLengthMM = 1000.0
	minimumOverloadRatioPercent   = 100.0
	maximumOverloadRatioPercent   = 200.0
	// The public model table uses one fixed loaded length and one fixed
	// calculation-load allowance. Each model applies F = sigma x A with its own
	// effective area, so the displayed elongation changes by model.
	spokeReferenceCalculationLengthMM   = 270.0
	spokeReferenceCalculationTensionKGF = 180.0
	standardGravityForcePerKilogram     = 9.80665
	spokeReferenceCalculationTensionN   = spokeReferenceCalculationTensionKGF * standardGravityForcePerKilogram
	// This page-level comparison baseline is deliberately not a universal rim
	// failure limit, factory test standard, or manufacturer certification value.
	carbonRimSpokeHoleWarningForceN = 1850.0
	highStressWarningForceN         = 1650.0
	criticalYieldRatioPercent       = 95.0
	highStressYieldRatioPercent     = 80.0
)

var (
	ErrInvalidStainlessSteelSpokeDislocationMechanicsField    = errors.New("invalid stainless steel spoke dislocation mechanics field")
	ErrStainlessSteelSpokeDislocationMechanicsFieldOutOfRange = errors.New("stainless steel spoke dislocation mechanics field out of range")
	ErrStainlessSteelSpokeStressOutsideMaterialCurve          = errors.New("spoke stress is outside the selected material curve")
	ErrUnknownStainlessSteelSpokeModel                        = errors.New("unknown stainless steel spoke model")
	ErrCustomSpokeAreaRequired                                = errors.New("custom effective spoke area is required")
)

// StainlessSteelSpokeDislocationMechanicsPhysicalReferences contains the
// material and rim references used by the calculation and safety gate.
type StainlessSteelSpokeDislocationMechanicsPhysicalReferences struct {
	MaterialReferenceID             string  `json:"material_reference_id"`
	MaterialCatalogID               string  `json:"material_catalog_id"`
	MaterialCatalogVersion          string  `json:"material_catalog_version"`
	MaterialDataVersion             string  `json:"material_data_version"`
	MaterialDataStatus              string  `json:"material_data_status"`
	YieldStrengthMPA                float64 `json:"yield_strength_mpa"`
	UltimateTensileStrengthMPA      float64 `json:"ultimate_tensile_strength_mpa"`
	ElasticModulusMPA               float64 `json:"elastic_modulus_mpa"`
	CarbonRimSpokeHoleWarningForceN float64 `json:"carbon_rim_spoke_hole_warning_force_n"`
	HighStressWarningForceN         float64 `json:"high_stress_warning_force_n"`
	CriticalYieldRatioPercent       float64 `json:"critical_yield_ratio_percent"`
	HighStressYieldRatioPercent     float64 `json:"high_stress_yield_ratio_percent"`
}

// StainlessSteelSpokeDislocationMechanicsMetadata is the complete public
// reference contract needed to render the calculator controls and notes.
type StainlessSteelSpokeDislocationMechanicsMetadata struct {
	ModelVersion           string                                                      `json:"model_version"`
	ModelCatalogID         string                                                      `json:"model_catalog_id"`
	ModelCatalogVersion    string                                                      `json:"model_catalog_version"`
	MaterialCatalogID      string                                                      `json:"material_catalog_id"`
	MaterialCatalogVersion string                                                      `json:"material_catalog_version"`
	MaterialReference      StainlessSteelSpokeMaterialReference                        `json:"material_reference"`
	Models                 []StainlessSteelSpokeModelReference                         `json:"models"`
	ReferenceCalculation   StainlessSteelSpokeDislocationMechanicsReferenceCalculation `json:"reference_calculation"`
	PhysicalReferences     StainlessSteelSpokeDislocationMechanicsPhysicalReferences   `json:"physical_references"`
	InputBounds            StainlessSteelSpokeDislocationMechanicsInputBounds          `json:"input_bounds"`
}

// StainlessSteelSpokeDislocationMechanicsReferenceCalculation contains the
// fixed, crawlable comparison values shown in the public model table. Force
// boundary references are direct material calculations (F = sigma x A), while
// the main elongation reference applies the fixed 180 kgf calculation-load
// allowance to each model's effective area at the fixed loaded length. It is
// above the common 140 kgf working-tension reference to leave room for effects
// such as stress relief and service loading that this axial model cannot
// reproduce individually. The total-failure value is a sourced tensile-test
// elongation reference, not a fracture-force model.
type StainlessSteelSpokeDislocationMechanicsReferenceCalculation struct {
	EffectiveSpokeLengthMM             float64                                                             `json:"effective_spoke_length_mm"`
	ReferenceCalculationTensionKGF     float64                                                             `json:"reference_calculation_tension_kgf"`
	ReferenceCalculationTensionN       float64                                                             `json:"reference_calculation_tension_n"`
	YieldStrengthMPA                   float64                                                             `json:"yield_strength_mpa"`
	CurveEndpointStressMPA             float64                                                             `json:"curve_endpoint_stress_mpa"`
	CurveEndpointTotalStrainPercent    float64                                                             `json:"curve_endpoint_total_strain_percent"`
	CurveEndpointTotalElongationMM     float64                                                             `json:"curve_endpoint_total_elongation_mm"`
	TotalElongationToFailurePercent    float64                                                             `json:"total_elongation_to_failure_percent"`
	FractureReferenceTotalElongationMM float64                                                             `json:"fracture_reference_total_elongation_mm"`
	Results                            []StainlessSteelSpokeDislocationMechanicsReferenceCalculationResult `json:"results"`
}

// StainlessSteelSpokeDislocationMechanicsReferenceCalculationResult contains
// the fixed 180 kgf calculation-allowance, 270 mm elongation and
// material-boundary force values for one model row.
type StainlessSteelSpokeDislocationMechanicsReferenceCalculationResult struct {
	ModelID                                   string  `json:"model_id"`
	EffectiveAreaMM2                          float64 `json:"effective_area_mm2"`
	ReferenceCalculationStressMPA             float64 `json:"reference_calculation_stress_mpa"`
	ReferenceCalculationTotalStrainPercent    float64 `json:"reference_calculation_total_strain_percent"`
	ReferenceCalculationTotalElongationMM     float64 `json:"reference_calculation_total_elongation_mm"`
	ReferenceCalculationElasticElongationMM   float64 `json:"reference_calculation_elastic_elongation_mm"`
	ReferenceCalculationPermanentElongationMM float64 `json:"reference_calculation_permanent_elongation_mm"`
	YieldReferenceForceN                      float64 `json:"yield_reference_force_n"`
	CurveEndpointForceN                       float64 `json:"curve_endpoint_force_n"`
	FractureReferenceTotalElongationMM        float64 `json:"fracture_reference_total_elongation_mm"`
}

// StainlessSteelSpokeDislocationMechanicsInputBounds documents the accepted
// physical input envelope so the client can configure controls without owning
// the validation rules.
type StainlessSteelSpokeDislocationMechanicsInputBounds struct {
	NominalWorkingTensionMinExclusiveN float64 `json:"nominal_working_tension_min_exclusive_n"`
	NominalWorkingTensionMaxN          float64 `json:"nominal_working_tension_max_n"`
	EffectiveSpokeAreaMinExclusiveMM2  float64 `json:"effective_spoke_area_min_exclusive_mm2"`
	EffectiveSpokeAreaMaxMM2           float64 `json:"effective_spoke_area_max_mm2"`
	EffectiveSpokeLengthMinExclusiveMM float64 `json:"effective_spoke_length_min_exclusive_mm"`
	EffectiveSpokeLengthMaxMM          float64 `json:"effective_spoke_length_max_mm"`
	OverloadRatioMinPercent            float64 `json:"overload_ratio_min_percent"`
	OverloadRatioMaxPercent            float64 `json:"overload_ratio_max_percent"`
}

// StainlessSteelSpokeDislocationMechanicsCalculationRequest contains only
// user-selected values. For a catalog model the backend resolves the area;
// custom_effective_area_mm2 is used only when model_id is custom.
type StainlessSteelSpokeDislocationMechanicsCalculationRequest struct {
	ModelID                string   `json:"model_id"`
	CustomEffectiveAreaMM2 *float64 `json:"custom_effective_area_mm2,omitempty"`
	EffectiveSpokeLengthMM float64  `json:"effective_spoke_length_mm"`
	NominalWorkingTensionN float64  `json:"nominal_working_tension_n"`
	OverloadRatioPercent   float64  `json:"overload_ratio_percent"`
}

// StainlessSteelSpokeDislocationMechanicsSafetyLevel is deliberately small
// and machine-readable; wording remains localized in the presentation layer.
type StainlessSteelSpokeDislocationMechanicsSafetyLevel string

const (
	StainlessSteelSpokeDislocationMechanicsSafetyReference StainlessSteelSpokeDislocationMechanicsSafetyLevel = "reference"
	StainlessSteelSpokeDislocationMechanicsSafetyWarning   StainlessSteelSpokeDislocationMechanicsSafetyLevel = "warning"
	StainlessSteelSpokeDislocationMechanicsSafetyCritical  StainlessSteelSpokeDislocationMechanicsSafetyLevel = "critical"
)

// StainlessSteelSpokeDislocationMechanicsCalculationResult contains rounded
// display values. No raw IEEE-754 calculation tails are exposed to clients.
type StainlessSteelSpokeDislocationMechanicsCalculationResult struct {
	ModelVersion                       string                                             `json:"model_version"`
	ModelID                            string                                             `json:"model_id"`
	ModelReference                     StainlessSteelSpokeModelReference                  `json:"model_reference"`
	ModelCatalogID                     string                                             `json:"model_catalog_id"`
	ModelCatalogVersion                string                                             `json:"model_catalog_version"`
	MaterialReferenceID                string                                             `json:"material_reference_id"`
	MaterialCatalogID                  string                                             `json:"material_catalog_id"`
	MaterialCatalogVersion             string                                             `json:"material_catalog_version"`
	MaterialReference                  StainlessSteelSpokeMaterialReference               `json:"material_reference"`
	EffectiveAreaMM2                   float64                                            `json:"effective_area_mm2"`
	EffectiveSpokeLengthMM             float64                                            `json:"effective_spoke_length_mm"`
	ElasticModulusMPA                  float64                                            `json:"elastic_modulus_mpa"`
	NominalWorkingTensionN             int                                                `json:"nominal_working_tension_n"`
	OverloadRatioPercent               float64                                            `json:"overload_ratio_percent"`
	WorkStressMPA                      int                                                `json:"work_stress_mpa"`
	WorkYieldRatioPercent              float64                                            `json:"work_yield_ratio_percent"`
	WorkTotalStrainPercent             float64                                            `json:"work_total_strain_percent"`
	WorkElasticStrainPercent           float64                                            `json:"work_elastic_strain_percent"`
	WorkPlasticStrainPercent           float64                                            `json:"work_plastic_strain_percent"`
	WorkTotalElongationMM              float64                                            `json:"work_total_elongation_mm"`
	WorkElasticElongationMM            float64                                            `json:"work_elastic_elongation_mm"`
	WorkPermanentElongationMM          float64                                            `json:"work_permanent_elongation_mm"`
	OverloadForceN                     int                                                `json:"overload_force_n"`
	OverloadDeltaForceN                int                                                `json:"overload_delta_force_n"`
	OverloadStressMPA                  int                                                `json:"overload_stress_mpa"`
	OverloadYieldRatioPercent          float64                                            `json:"overload_yield_ratio_percent"`
	OverloadTotalStrainPercent         float64                                            `json:"overload_total_strain_percent"`
	OverloadElasticStrainPercent       float64                                            `json:"overload_elastic_strain_percent"`
	OverloadPlasticStrainPercent       float64                                            `json:"overload_plastic_strain_percent"`
	OverloadTotalElongationMM          float64                                            `json:"overload_total_elongation_mm"`
	OverloadElasticElongationMM        float64                                            `json:"overload_elastic_elongation_mm"`
	OverloadPermanentElongationMM      float64                                            `json:"overload_permanent_elongation_mm"`
	OverloadDeltaElasticElongationMM   float64                                            `json:"overload_delta_elastic_elongation_mm"`
	OverloadDeltaTotalElongationMM     float64                                            `json:"overload_delta_total_elongation_mm"`
	OverloadDeltaPermanentElongationMM float64                                            `json:"overload_delta_permanent_elongation_mm"`
	FractureReferenceTotalElongationMM float64                                            `json:"fracture_reference_total_elongation_mm"`
	YieldMarginPercent                 float64                                            `json:"yield_margin_percent"`
	SafetyLevel                        StainlessSteelSpokeDislocationMechanicsSafetyLevel `json:"safety_level"`
	RimWarningMarginN                  int                                                `json:"rim_warning_margin_n"`
	RimWarningMarginPercent            float64                                            `json:"rim_warning_margin_percent"`
	RimWarningExcessN                  int                                                `json:"rim_warning_excess_n"`
}

// GetStainlessSteelSpokeDislocationMechanicsMetadata returns the complete
// backend reference set used by the page.
func GetStainlessSteelSpokeDislocationMechanicsMetadata() StainlessSteelSpokeDislocationMechanicsMetadata {
	modelCatalog := GetStainlessSteelSpokeModelReferenceCatalog()
	materialCatalog := GetStainlessSteelSpokeMaterialReferenceCatalog()
	materialReference := GetDefaultStainlessSteelSpokeMaterialReference()
	return StainlessSteelSpokeDislocationMechanicsMetadata{
		ModelVersion:           StainlessSteelSpokeDislocationMechanicsModelVersion,
		ModelCatalogID:         modelCatalog.CatalogID,
		ModelCatalogVersion:    modelCatalog.CatalogVersion,
		MaterialCatalogID:      materialCatalog.CatalogID,
		MaterialCatalogVersion: materialCatalog.CatalogVersion,
		MaterialReference:      materialReference,
		Models:                 ListStainlessSteelSpokeDislocationMechanicsModelReferences(),
		ReferenceCalculation:   buildStainlessSteelSpokeDislocationMechanicsReferenceCalculation(),
		PhysicalReferences: StainlessSteelSpokeDislocationMechanicsPhysicalReferences{
			MaterialReferenceID:             materialReference.ID,
			MaterialCatalogID:               materialCatalog.CatalogID,
			MaterialCatalogVersion:          materialCatalog.CatalogVersion,
			MaterialDataVersion:             materialReference.DataVersion,
			MaterialDataStatus:              materialReference.DataStatus,
			YieldStrengthMPA:                materialReference.YieldStrengthMPA,
			UltimateTensileStrengthMPA:      materialReference.UltimateTensileStrengthMPA,
			ElasticModulusMPA:               materialReference.ElasticModulusMPA,
			CarbonRimSpokeHoleWarningForceN: carbonRimSpokeHoleWarningForceN,
			HighStressWarningForceN:         highStressWarningForceN,
			CriticalYieldRatioPercent:       criticalYieldRatioPercent,
			HighStressYieldRatioPercent:     highStressYieldRatioPercent,
		},
		InputBounds: StainlessSteelSpokeDislocationMechanicsInputBounds{
			NominalWorkingTensionMinExclusiveN: minimumNominalWorkingTensionN,
			NominalWorkingTensionMaxN:          maximumNominalWorkingTensionN,
			EffectiveSpokeAreaMinExclusiveMM2:  minimumEffectiveSpokeAreaMM2,
			EffectiveSpokeAreaMaxMM2:           maximumEffectiveSpokeAreaMM2,
			EffectiveSpokeLengthMinExclusiveMM: minimumEffectiveSpokeLengthMM,
			EffectiveSpokeLengthMaxMM:          maximumEffectiveSpokeLengthMM,
			OverloadRatioMinPercent:            minimumOverloadRatioPercent,
			OverloadRatioMaxPercent:            maximumOverloadRatioPercent,
		},
	}
}

func buildStainlessSteelSpokeDislocationMechanicsReferenceCalculation() StainlessSteelSpokeDislocationMechanicsReferenceCalculation {
	materialReference := GetDefaultStainlessSteelSpokeMaterialReference()
	baseline := StainlessSteelSpokeDislocationMechanicsReferenceCalculation{
		EffectiveSpokeLengthMM:             spokeReferenceCalculationLengthMM,
		ReferenceCalculationTensionKGF:     spokeReferenceCalculationTensionKGF,
		ReferenceCalculationTensionN:       spokeReferenceCalculationTensionN,
		YieldStrengthMPA:                   materialReference.YieldStrengthMPA,
		CurveEndpointStressMPA:             materialReference.UltimateTensileStrengthMPA,
		CurveEndpointTotalStrainPercent:    0,
		CurveEndpointTotalElongationMM:     0,
		TotalElongationToFailurePercent:    materialReference.TotalElongationToFailurePercent,
		FractureReferenceTotalElongationMM: roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(materialReference.TotalElongationToFailurePercent/100*spokeReferenceCalculationLengthMM, 3),
		Results:                            make([]StainlessSteelSpokeDislocationMechanicsReferenceCalculationResult, 0),
	}
	curveEndpointState, err := resolveStainlessSteelSpokeCurveState(materialReference, baseline.CurveEndpointStressMPA)
	if err != nil {
		panic(fmt.Sprintf("resolve stainless steel spoke curve endpoint state: %v", err))
	}
	baseline.CurveEndpointTotalStrainPercent = roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(curveEndpointState.TotalStrain*100, 3)
	baseline.CurveEndpointTotalElongationMM = roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(curveEndpointState.TotalStrain*baseline.EffectiveSpokeLengthMM, 3)
	models := ListStainlessSteelSpokeDislocationMechanicsModelReferences()
	baseline.Results = make([]StainlessSteelSpokeDislocationMechanicsReferenceCalculationResult, 0, len(models))
	for _, model := range models {
		modelMaterialReference, found := GetStainlessSteelSpokeMaterialReferenceByID(model.MaterialReferenceID)
		if !found {
			panic(fmt.Sprintf("material reference %q for model %q is unavailable", model.MaterialReferenceID, model.ID))
		}
		referenceCalculationStressMPA := spokeReferenceCalculationTensionN / model.EffectiveAreaMM2
		referenceCalculationState, err := resolveStainlessSteelSpokeCurveState(modelMaterialReference, referenceCalculationStressMPA)
		if err != nil {
			panic(fmt.Sprintf("resolve stainless steel spoke 180 kgf reference state for %q: %v", model.ID, err))
		}
		yieldReferenceForceN := modelMaterialReference.YieldStrengthMPA * model.EffectiveAreaMM2
		curveEndpointForceN := modelMaterialReference.UltimateTensileStrengthMPA * model.EffectiveAreaMM2
		modelFractureReferenceTotalElongationMM := roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(modelMaterialReference.TotalElongationToFailurePercent/100*baseline.EffectiveSpokeLengthMM, 3)
		baseline.Results = append(baseline.Results, StainlessSteelSpokeDislocationMechanicsReferenceCalculationResult{
			ModelID:                                   model.ID,
			EffectiveAreaMM2:                          model.EffectiveAreaMM2,
			ReferenceCalculationStressMPA:             roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(referenceCalculationStressMPA, 3),
			ReferenceCalculationTotalStrainPercent:    roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(referenceCalculationState.TotalStrain*100, 3),
			ReferenceCalculationTotalElongationMM:     roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(referenceCalculationState.TotalStrain*baseline.EffectiveSpokeLengthMM, 3),
			ReferenceCalculationElasticElongationMM:   roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(referenceCalculationState.ElasticStrain*baseline.EffectiveSpokeLengthMM, 3),
			ReferenceCalculationPermanentElongationMM: roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(referenceCalculationState.PlasticStrain*baseline.EffectiveSpokeLengthMM, 3),
			YieldReferenceForceN:                      roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(yieldReferenceForceN, 0),
			CurveEndpointForceN:                       roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(curveEndpointForceN, 0),
			FractureReferenceTotalElongationMM:        modelFractureReferenceTotalElongationMM,
		})
	}
	return baseline
}

// ValidateStainlessSteelSpokeDislocationMechanicsCalculationRequest applies
// fail-loudly finite and physical boundary checks before any division.
func ValidateStainlessSteelSpokeDislocationMechanicsCalculationRequest(request StainlessSteelSpokeDislocationMechanicsCalculationRequest) error {
	if strings.TrimSpace(request.ModelID) == "" {
		return fmt.Errorf("%w: model_id is required", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}
	if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(request.NominalWorkingTensionN) {
		return fmt.Errorf("%w: nominal_working_tension_n must be finite", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}
	if request.NominalWorkingTensionN <= minimumNominalWorkingTensionN || request.NominalWorkingTensionN > maximumNominalWorkingTensionN {
		return fmt.Errorf("%w: nominal_working_tension_n must be greater than %.1f and at most %.1f", ErrStainlessSteelSpokeDislocationMechanicsFieldOutOfRange, minimumNominalWorkingTensionN, maximumNominalWorkingTensionN)
	}
	if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(request.EffectiveSpokeLengthMM) {
		return fmt.Errorf("%w: effective_spoke_length_mm must be finite", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}
	if request.EffectiveSpokeLengthMM <= minimumEffectiveSpokeLengthMM || request.EffectiveSpokeLengthMM > maximumEffectiveSpokeLengthMM {
		return fmt.Errorf("%w: effective_spoke_length_mm must be greater than %.1f and at most %.1f", ErrStainlessSteelSpokeDislocationMechanicsFieldOutOfRange, minimumEffectiveSpokeLengthMM, maximumEffectiveSpokeLengthMM)
	}
	if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(request.OverloadRatioPercent) {
		return fmt.Errorf("%w: overload_ratio_percent must be finite", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}
	if request.OverloadRatioPercent < minimumOverloadRatioPercent || request.OverloadRatioPercent > maximumOverloadRatioPercent {
		return fmt.Errorf("%w: overload_ratio_percent must be between %.1f and %.1f", ErrStainlessSteelSpokeDislocationMechanicsFieldOutOfRange, minimumOverloadRatioPercent, maximumOverloadRatioPercent)
	}

	modelID := strings.ToLower(strings.TrimSpace(request.ModelID))
	if modelID == "custom" {
		if request.CustomEffectiveAreaMM2 == nil {
			return ErrCustomSpokeAreaRequired
		}
		if err := validateEffectiveStainlessSteelSpokeAreaSquareMillimeters(*request.CustomEffectiveAreaMM2); err != nil {
			return err
		}
		return nil
	}
	for _, model := range ListStainlessSteelSpokeDislocationMechanicsModelReferences() {
		if model.ID == modelID {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrUnknownStainlessSteelSpokeModel, modelID)
}

// CalculateStainlessSteelSpokeDislocationMechanics resolves a catalog or
// custom section and performs the complete backend-authoritative calculation.
func CalculateStainlessSteelSpokeDislocationMechanics(request StainlessSteelSpokeDislocationMechanicsCalculationRequest) (StainlessSteelSpokeDislocationMechanicsCalculationResult, error) {
	if err := ValidateStainlessSteelSpokeDislocationMechanicsCalculationRequest(request); err != nil {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, err
	}
	modelID := strings.ToLower(strings.TrimSpace(request.ModelID))
	areaMM2, err := resolveEffectiveStainlessSteelSpokeAreaSquareMillimeters(modelID, request.CustomEffectiveAreaMM2)
	if err != nil {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, err
	}
	modelReference, modelReferenceFound := GetStainlessSteelSpokeDislocationMechanicsModelReferenceByID(modelID)
	materialReferenceID := getDefaultStainlessSteelSpokeMaterialReferenceID()
	if modelReferenceFound {
		materialReferenceID = modelReference.MaterialReferenceID
	}
	materialReference, found := GetStainlessSteelSpokeMaterialReferenceByID(materialReferenceID)
	if !found {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, fmt.Errorf("%w: default material reference is unavailable", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}
	modelCatalog := GetStainlessSteelSpokeModelReferenceCatalog()
	materialCatalog := GetStainlessSteelSpokeMaterialReferenceCatalog()

	workStressMPA := request.NominalWorkingTensionN / areaMM2
	workYieldRatioPercent := (workStressMPA / materialReference.YieldStrengthMPA) * 100
	overloadForceN := request.NominalWorkingTensionN * (request.OverloadRatioPercent / 100)
	overloadDeltaForceN := overloadForceN - request.NominalWorkingTensionN
	overloadStressMPA := overloadForceN / areaMM2
	overloadYieldRatioPercent := (overloadStressMPA / materialReference.YieldStrengthMPA) * 100
	workCurveState, err := resolveStainlessSteelSpokeCurveState(materialReference, workStressMPA)
	if err != nil {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, fmt.Errorf("%w: working tension stress %.1f MPa", err, workStressMPA)
	}
	overloadCurveState, err := resolveStainlessSteelSpokeCurveState(materialReference, overloadStressMPA)
	if err != nil {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, fmt.Errorf("%w: overload stress %.1f MPa", err, overloadStressMPA)
	}
	workTotalElongationMM := workCurveState.TotalStrain * request.EffectiveSpokeLengthMM
	workElasticElongationMM := workCurveState.ElasticStrain * request.EffectiveSpokeLengthMM
	workPermanentElongationMM := workCurveState.PlasticStrain * request.EffectiveSpokeLengthMM
	overloadTotalElongationMM := overloadCurveState.TotalStrain * request.EffectiveSpokeLengthMM
	overloadElasticElongationMM := overloadCurveState.ElasticStrain * request.EffectiveSpokeLengthMM
	overloadPermanentElongationMM := overloadCurveState.PlasticStrain * request.EffectiveSpokeLengthMM
	overloadDeltaElasticElongationMM := overloadElasticElongationMM - workElasticElongationMM
	overloadDeltaTotalElongationMM := overloadTotalElongationMM - workTotalElongationMM
	overloadDeltaPermanentElongationMM := overloadPermanentElongationMM - workPermanentElongationMM
	fractureReferenceTotalElongationMM := materialReference.TotalElongationToFailurePercent / 100 * request.EffectiveSpokeLengthMM
	yieldMarginPercent := math.Max(0, 100-overloadYieldRatioPercent)

	if !areFiniteStainlessSteelSpokeDislocationMechanicsNumbers(workStressMPA, workYieldRatioPercent, overloadForceN, overloadDeltaForceN, overloadStressMPA, overloadYieldRatioPercent, workCurveState.TotalStrain, workCurveState.ElasticStrain, workCurveState.PlasticStrain, overloadCurveState.TotalStrain, overloadCurveState.ElasticStrain, overloadCurveState.PlasticStrain, workTotalElongationMM, workElasticElongationMM, workPermanentElongationMM, overloadTotalElongationMM, overloadElasticElongationMM, overloadPermanentElongationMM, overloadDeltaElasticElongationMM, overloadDeltaTotalElongationMM, overloadDeltaPermanentElongationMM, fractureReferenceTotalElongationMM, yieldMarginPercent) {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, fmt.Errorf("%w: calculation produced a non-finite value", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}

	safetyLevel := StainlessSteelSpokeDislocationMechanicsSafetyReference
	if overloadForceN > carbonRimSpokeHoleWarningForceN || overloadYieldRatioPercent >= criticalYieldRatioPercent {
		safetyLevel = StainlessSteelSpokeDislocationMechanicsSafetyCritical
	} else if overloadForceN > highStressWarningForceN || overloadYieldRatioPercent > highStressYieldRatioPercent {
		safetyLevel = StainlessSteelSpokeDislocationMechanicsSafetyWarning
	}

	return StainlessSteelSpokeDislocationMechanicsCalculationResult{
		ModelVersion:                       StainlessSteelSpokeDislocationMechanicsModelVersion,
		ModelID:                            modelID,
		ModelReference:                     modelReference,
		ModelCatalogID:                     modelCatalog.CatalogID,
		ModelCatalogVersion:                modelCatalog.CatalogVersion,
		MaterialReferenceID:                materialReference.ID,
		MaterialCatalogID:                  materialCatalog.CatalogID,
		MaterialCatalogVersion:             materialCatalog.CatalogVersion,
		MaterialReference:                  materialReference,
		EffectiveAreaMM2:                   roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(areaMM2, 2),
		EffectiveSpokeLengthMM:             roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(request.EffectiveSpokeLengthMM, 1),
		ElasticModulusMPA:                  materialReference.ElasticModulusMPA,
		NominalWorkingTensionN:             int(math.Round(request.NominalWorkingTensionN)),
		OverloadRatioPercent:               roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(request.OverloadRatioPercent, 1),
		WorkStressMPA:                      int(math.Round(workStressMPA)),
		WorkYieldRatioPercent:              roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workYieldRatioPercent, 1),
		WorkTotalStrainPercent:             roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workCurveState.TotalStrain*100, 3),
		WorkElasticStrainPercent:           roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workCurveState.ElasticStrain*100, 3),
		WorkPlasticStrainPercent:           roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workCurveState.PlasticStrain*100, 3),
		WorkTotalElongationMM:              roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workTotalElongationMM, 3),
		WorkElasticElongationMM:            roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workElasticElongationMM, 3),
		WorkPermanentElongationMM:          roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workPermanentElongationMM, 3),
		OverloadForceN:                     int(math.Round(overloadForceN)),
		OverloadDeltaForceN:                int(math.Round(overloadDeltaForceN)),
		OverloadStressMPA:                  int(math.Round(overloadStressMPA)),
		OverloadYieldRatioPercent:          roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadYieldRatioPercent, 1),
		OverloadTotalStrainPercent:         roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadCurveState.TotalStrain*100, 3),
		OverloadElasticStrainPercent:       roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadCurveState.ElasticStrain*100, 3),
		OverloadPlasticStrainPercent:       roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadCurveState.PlasticStrain*100, 3),
		OverloadTotalElongationMM:          roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadTotalElongationMM, 3),
		OverloadElasticElongationMM:        roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadElasticElongationMM, 3),
		OverloadPermanentElongationMM:      roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadPermanentElongationMM, 3),
		OverloadDeltaElasticElongationMM:   roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadDeltaElasticElongationMM, 3),
		OverloadDeltaTotalElongationMM:     roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadDeltaTotalElongationMM, 3),
		OverloadDeltaPermanentElongationMM: roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadDeltaPermanentElongationMM, 3),
		FractureReferenceTotalElongationMM: roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(fractureReferenceTotalElongationMM, 3),
		YieldMarginPercent:                 roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(yieldMarginPercent, 1),
		SafetyLevel:                        safetyLevel,
		RimWarningMarginN:                  int(math.Round(carbonRimSpokeHoleWarningForceN - overloadForceN)),
		RimWarningMarginPercent:            roundStainlessSteelSpokeDislocationMechanicsDisplayNumber((carbonRimSpokeHoleWarningForceN-overloadForceN)/carbonRimSpokeHoleWarningForceN*100, 1),
		RimWarningExcessN:                  int(math.Round(math.Max(0, overloadForceN-carbonRimSpokeHoleWarningForceN))),
	}, nil
}

type stainlessSteelSpokeCurveState struct {
	TotalStrain   float64
	ElasticStrain float64
	PlasticStrain float64
}

// resolveStainlessSteelSpokeCurveState interpolates the selected material's
// stored engineering curve. The selected profile is already a cold-drawn
// material state, so its post-yield ascending branch contains the additional
// monotonic work-hardening response of that state. The calculator does not add
// a separate dislocation threshold or invent a hardening coefficient. Stress
// above the stored curve endpoint is rejected so the backend never invents a
// post-UTS stress, force, necking, or fracture-load estimate. The separately
// sourced total elongation-to-failure value is exposed only as a loaded-length
// reference.
func resolveStainlessSteelSpokeCurveState(material StainlessSteelSpokeMaterialReference, stressMPA float64) (stainlessSteelSpokeCurveState, error) {
	if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(stressMPA) || stressMPA < 0 || stressMPA > material.UltimateTensileStrengthMPA {
		return stainlessSteelSpokeCurveState{}, fmt.Errorf("%w: %.1f MPa is outside 0–%.1f MPa", ErrStainlessSteelSpokeStressOutsideMaterialCurve, stressMPA, material.UltimateTensileStrengthMPA)
	}
	curve := material.StressStrainCurve
	if len(curve) < 2 {
		return stainlessSteelSpokeCurveState{}, fmt.Errorf("%w: selected material curve is empty", ErrStainlessSteelSpokeStressOutsideMaterialCurve)
	}
	for index := 1; index < len(curve); index++ {
		left := curve[index-1]
		right := curve[index]
		if stressMPA > right.EngineeringStressMPA {
			continue
		}
		stressSpan := right.EngineeringStressMPA - left.EngineeringStressMPA
		strainRatio := 0.0
		if stressSpan > 0 {
			strainRatio = (stressMPA - left.EngineeringStressMPA) / stressSpan
		}
		totalStrain := left.EngineeringStrain + strainRatio*(right.EngineeringStrain-left.EngineeringStrain)
		elasticStrain := stressMPA / material.ElasticModulusMPA
		// The difference between the cold-drawn curve's total strain and the
		// ideal elastic strain is the permanent strain after unloading.
		plasticStrain := math.Max(0, totalStrain-elasticStrain)
		return stainlessSteelSpokeCurveState{TotalStrain: totalStrain, ElasticStrain: elasticStrain, PlasticStrain: plasticStrain}, nil
	}
	return stainlessSteelSpokeCurveState{}, fmt.Errorf("%w: %.1f MPa is outside the stored curve", ErrStainlessSteelSpokeStressOutsideMaterialCurve, stressMPA)
}

func resolveEffectiveStainlessSteelSpokeAreaSquareMillimeters(modelID string, customAreaMM2 *float64) (float64, error) {
	if modelID == "custom" {
		if customAreaMM2 == nil {
			return 0, ErrCustomSpokeAreaRequired
		}
		return *customAreaMM2, nil
	}
	for _, model := range ListStainlessSteelSpokeDislocationMechanicsModelReferences() {
		if model.ID == modelID {
			return model.EffectiveAreaMM2, nil
		}
	}
	return 0, fmt.Errorf("%w: %s", ErrUnknownStainlessSteelSpokeModel, modelID)
}

func validateEffectiveStainlessSteelSpokeAreaSquareMillimeters(areaMM2 float64) error {
	if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(areaMM2) {
		return fmt.Errorf("%w: custom_effective_area_mm2 must be finite", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}
	if areaMM2 <= minimumEffectiveSpokeAreaMM2 || areaMM2 > maximumEffectiveSpokeAreaMM2 {
		return fmt.Errorf("%w: custom_effective_area_mm2 must be greater than %.1f and at most %.1f", ErrStainlessSteelSpokeDislocationMechanicsFieldOutOfRange, minimumEffectiveSpokeAreaMM2, maximumEffectiveSpokeAreaMM2)
	}
	return nil
}

func isFiniteStainlessSteelSpokeDislocationMechanicsNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func areFiniteStainlessSteelSpokeDislocationMechanicsNumbers(values ...float64) bool {
	for _, value := range values {
		if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(value) {
			return false
		}
	}
	return true
}

func roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(value float64, decimalPlaces int) float64 {
	factor := math.Pow10(decimalPlaces)
	return math.Round(value*factor) / factor
}

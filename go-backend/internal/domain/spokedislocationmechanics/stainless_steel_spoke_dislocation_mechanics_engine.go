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
const StainlessSteelSpokeDislocationMechanicsModelVersion = "stainless-steel-spoke-dislocation-mechanics-v2.1"

const (
	minimumNominalWorkingTensionN = 0.0
	maximumNominalWorkingTensionN = 3000.0
	minimumEffectiveSpokeAreaMM2  = 0.2
	maximumEffectiveSpokeAreaMM2  = 10.0
	minimumEffectiveSpokeLengthMM = 1.0
	maximumEffectiveSpokeLengthMM = 1000.0
	minimumOverloadRatioPercent   = 100.0
	maximumOverloadRatioPercent   = 200.0
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
	ErrUnknownStainlessSteelSpokeModel                        = errors.New("unknown stainless steel spoke model")
	ErrCustomSpokeAreaRequired                                = errors.New("custom effective spoke area is required")
)

// StainlessSteelSpokeDislocationMechanicsPhysicalReferences contains the
// material and rim references used by the calculation and safety gate.
type StainlessSteelSpokeDislocationMechanicsPhysicalReferences struct {
	MaterialReferenceID                 string  `json:"material_reference_id"`
	MaterialCatalogID                   string  `json:"material_catalog_id"`
	MaterialCatalogVersion              string  `json:"material_catalog_version"`
	MaterialDataVersion                 string  `json:"material_data_version"`
	MaterialDataStatus                  string  `json:"material_data_status"`
	MacroYieldReferenceMPA              float64 `json:"macro_yield_reference_mpa"`
	UltimateTensileStrengthReferenceMPA float64 `json:"ultimate_tensile_strength_reference_mpa"`
	ElasticModulusMPA                   float64 `json:"elastic_modulus_mpa"`
	CarbonRimSpokeHoleWarningForceN     float64 `json:"carbon_rim_spoke_hole_warning_force_n"`
	HighStressWarningForceN             float64 `json:"high_stress_warning_force_n"`
	CriticalYieldRatioPercent           float64 `json:"critical_yield_ratio_percent"`
	HighStressYieldRatioPercent         float64 `json:"high_stress_yield_ratio_percent"`
}

// StainlessSteelSpokeDislocationMechanicsMetadata is the complete public
// reference contract needed to render the calculator controls and notes.
type StainlessSteelSpokeDislocationMechanicsMetadata struct {
	ModelVersion           string                                                    `json:"model_version"`
	ModelCatalogID         string                                                    `json:"model_catalog_id"`
	ModelCatalogVersion    string                                                    `json:"model_catalog_version"`
	MaterialCatalogID      string                                                    `json:"material_catalog_id"`
	MaterialCatalogVersion string                                                    `json:"material_catalog_version"`
	MaterialReference      StainlessSteelSpokeMaterialReference                      `json:"material_reference"`
	Models                 []StainlessSteelSpokeModelReference                       `json:"models"`
	PhysicalReferences     StainlessSteelSpokeDislocationMechanicsPhysicalReferences `json:"physical_references"`
	InputBounds            StainlessSteelSpokeDislocationMechanicsInputBounds        `json:"input_bounds"`
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
	ModelVersion                     string                                             `json:"model_version"`
	ModelID                          string                                             `json:"model_id"`
	ModelReference                   StainlessSteelSpokeModelReference                  `json:"model_reference"`
	ModelCatalogID                   string                                             `json:"model_catalog_id"`
	ModelCatalogVersion              string                                             `json:"model_catalog_version"`
	MaterialReferenceID              string                                             `json:"material_reference_id"`
	MaterialCatalogID                string                                             `json:"material_catalog_id"`
	MaterialCatalogVersion           string                                             `json:"material_catalog_version"`
	MaterialReference                StainlessSteelSpokeMaterialReference               `json:"material_reference"`
	EffectiveAreaMM2                 float64                                            `json:"effective_area_mm2"`
	EffectiveSpokeLengthMM           float64                                            `json:"effective_spoke_length_mm"`
	ElasticModulusMPA                float64                                            `json:"elastic_modulus_mpa"`
	NominalWorkingTensionN           int                                                `json:"nominal_working_tension_n"`
	OverloadRatioPercent             float64                                            `json:"overload_ratio_percent"`
	WorkStressMPA                    int                                                `json:"work_stress_mpa"`
	WorkYieldRatioPercent            float64                                            `json:"work_yield_ratio_percent"`
	WorkElasticElongationMM          float64                                            `json:"work_elastic_elongation_mm"`
	OverloadForceN                   int                                                `json:"overload_force_n"`
	OverloadDeltaForceN              int                                                `json:"overload_delta_force_n"`
	OverloadStressMPA                int                                                `json:"overload_stress_mpa"`
	OverloadYieldRatioPercent        float64                                            `json:"overload_yield_ratio_percent"`
	OverloadElasticElongationMM      float64                                            `json:"overload_elastic_elongation_mm"`
	OverloadDeltaElasticElongationMM float64                                            `json:"overload_delta_elastic_elongation_mm"`
	YieldMarginPercent               float64                                            `json:"yield_margin_percent"`
	SafetyLevel                      StainlessSteelSpokeDislocationMechanicsSafetyLevel `json:"safety_level"`
	RimWarningMarginN                int                                                `json:"rim_warning_margin_n"`
	RimWarningMarginPercent          float64                                            `json:"rim_warning_margin_percent"`
	RimWarningExcessN                int                                                `json:"rim_warning_excess_n"`
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
		PhysicalReferences: StainlessSteelSpokeDislocationMechanicsPhysicalReferences{
			MaterialReferenceID:                 materialReference.ID,
			MaterialCatalogID:                   materialCatalog.CatalogID,
			MaterialCatalogVersion:              materialCatalog.CatalogVersion,
			MaterialDataVersion:                 materialReference.DataVersion,
			MaterialDataStatus:                  materialReference.DataStatus,
			MacroYieldReferenceMPA:              materialReference.MacroYieldReferenceMPA,
			UltimateTensileStrengthReferenceMPA: materialReference.UltimateTensileStrengthMPA,
			ElasticModulusMPA:                   materialReference.ElasticModulusMPA,
			CarbonRimSpokeHoleWarningForceN:     carbonRimSpokeHoleWarningForceN,
			HighStressWarningForceN:             highStressWarningForceN,
			CriticalYieldRatioPercent:           criticalYieldRatioPercent,
			HighStressYieldRatioPercent:         highStressYieldRatioPercent,
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
	workYieldRatioPercent := (workStressMPA / materialReference.MacroYieldReferenceMPA) * 100
	overloadForceN := request.NominalWorkingTensionN * (request.OverloadRatioPercent / 100)
	overloadDeltaForceN := overloadForceN - request.NominalWorkingTensionN
	overloadStressMPA := overloadForceN / areaMM2
	overloadYieldRatioPercent := (overloadStressMPA / materialReference.MacroYieldReferenceMPA) * 100
	workElasticElongationMM := calculateStainlessSteelSpokeElasticElongationMillimeters(request.NominalWorkingTensionN, request.EffectiveSpokeLengthMM, areaMM2, materialReference.ElasticModulusMPA)
	overloadElasticElongationMM := calculateStainlessSteelSpokeElasticElongationMillimeters(overloadForceN, request.EffectiveSpokeLengthMM, areaMM2, materialReference.ElasticModulusMPA)
	overloadDeltaElasticElongationMM := overloadElasticElongationMM - workElasticElongationMM
	yieldMarginPercent := math.Max(0, 100-overloadYieldRatioPercent)

	if !areFiniteStainlessSteelSpokeDislocationMechanicsNumbers(workStressMPA, workYieldRatioPercent, overloadForceN, overloadDeltaForceN, overloadStressMPA, overloadYieldRatioPercent, workElasticElongationMM, overloadElasticElongationMM, overloadDeltaElasticElongationMM, yieldMarginPercent) {
		return StainlessSteelSpokeDislocationMechanicsCalculationResult{}, fmt.Errorf("%w: calculation produced a non-finite value", ErrInvalidStainlessSteelSpokeDislocationMechanicsField)
	}

	safetyLevel := StainlessSteelSpokeDislocationMechanicsSafetyReference
	if overloadForceN > carbonRimSpokeHoleWarningForceN || overloadYieldRatioPercent >= criticalYieldRatioPercent {
		safetyLevel = StainlessSteelSpokeDislocationMechanicsSafetyCritical
	} else if overloadForceN > highStressWarningForceN || overloadYieldRatioPercent > highStressYieldRatioPercent {
		safetyLevel = StainlessSteelSpokeDislocationMechanicsSafetyWarning
	}

	return StainlessSteelSpokeDislocationMechanicsCalculationResult{
		ModelVersion:                     StainlessSteelSpokeDislocationMechanicsModelVersion,
		ModelID:                          modelID,
		ModelReference:                   modelReference,
		ModelCatalogID:                   modelCatalog.CatalogID,
		ModelCatalogVersion:              modelCatalog.CatalogVersion,
		MaterialReferenceID:              materialReference.ID,
		MaterialCatalogID:                materialCatalog.CatalogID,
		MaterialCatalogVersion:           materialCatalog.CatalogVersion,
		MaterialReference:                materialReference,
		EffectiveAreaMM2:                 roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(areaMM2, 2),
		EffectiveSpokeLengthMM:           roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(request.EffectiveSpokeLengthMM, 1),
		ElasticModulusMPA:                materialReference.ElasticModulusMPA,
		NominalWorkingTensionN:           int(math.Round(request.NominalWorkingTensionN)),
		OverloadRatioPercent:             roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(request.OverloadRatioPercent, 1),
		WorkStressMPA:                    int(math.Round(workStressMPA)),
		WorkYieldRatioPercent:            roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workYieldRatioPercent, 1),
		WorkElasticElongationMM:          roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(workElasticElongationMM, 3),
		OverloadForceN:                   int(math.Round(overloadForceN)),
		OverloadDeltaForceN:              int(math.Round(overloadDeltaForceN)),
		OverloadStressMPA:                int(math.Round(overloadStressMPA)),
		OverloadYieldRatioPercent:        roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadYieldRatioPercent, 1),
		OverloadElasticElongationMM:      roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadElasticElongationMM, 3),
		OverloadDeltaElasticElongationMM: roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(overloadDeltaElasticElongationMM, 3),
		YieldMarginPercent:               roundStainlessSteelSpokeDislocationMechanicsDisplayNumber(yieldMarginPercent, 1),
		SafetyLevel:                      safetyLevel,
		RimWarningMarginN:                int(math.Round(carbonRimSpokeHoleWarningForceN - overloadForceN)),
		RimWarningMarginPercent:          roundStainlessSteelSpokeDislocationMechanicsDisplayNumber((carbonRimSpokeHoleWarningForceN-overloadForceN)/carbonRimSpokeHoleWarningForceN*100, 1),
		RimWarningExcessN:                int(math.Round(math.Max(0, overloadForceN-carbonRimSpokeHoleWarningForceN))),
	}, nil
}

// calculateStainlessSteelSpokeElasticElongationMillimeters applies Hooke's law
// in N/mm² and mm so the returned elastic elongation is expressed in mm.
func calculateStainlessSteelSpokeElasticElongationMillimeters(forceN, lengthMM, areaMM2, elasticModulusMPA float64) float64 {
	return forceN * lengthMM / (areaMM2 * elasticModulusMPA)
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

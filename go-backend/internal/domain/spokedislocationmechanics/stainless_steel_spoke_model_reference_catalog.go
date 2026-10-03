package spokedislocationmechanics

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed stainless_steel_spoke_model_reference_catalog.json
var embeddedStainlessSteelSpokeModelReferenceCatalogFile []byte

// StainlessSteelSpokeModelReferenceCatalog is the backend-owned model and
// geometry catalog used by both metadata and elastic-elongation calculation.
// Keeping this data in a versioned JSON file allows reviewed values to change
// without rewriting the calculation formula.
type StainlessSteelSpokeModelReferenceCatalog struct {
	CatalogID      string                              `json:"catalog_id"`
	CatalogVersion string                              `json:"catalog_version"`
	Models         []StainlessSteelSpokeModelReference `json:"models"`
}

// StainlessSteelSpokeModelReference contains the geometry used by ΔL and
// the material profile selected by the calculation catalogue.
type StainlessSteelSpokeModelReference struct {
	ID                  string  `json:"id"`
	DisplayName         string  `json:"display_name"`
	GeometryDescription string  `json:"geometry_description"`
	EffectiveAreaMM2    float64 `json:"effective_area_mm2"`
	MaterialReferenceID string  `json:"material_reference_id"`
	DataVersion         string  `json:"data_version"`
	DataStatus          string  `json:"data_status"`
}

var loadedStainlessSteelSpokeModelReferenceCatalog = loadEmbeddedStainlessSteelSpokeModelReferenceCatalogOrPanic()

func loadEmbeddedStainlessSteelSpokeModelReferenceCatalogOrPanic() StainlessSteelSpokeModelReferenceCatalog {
	var catalog StainlessSteelSpokeModelReferenceCatalog
	if err := json.Unmarshal(embeddedStainlessSteelSpokeModelReferenceCatalogFile, &catalog); err != nil {
		panic(fmt.Sprintf("load stainless steel spoke model reference catalog: %v", err))
	}
	if err := validateStainlessSteelSpokeModelReferenceCatalog(catalog); err != nil {
		panic(fmt.Sprintf("validate stainless steel spoke model reference catalog: %v", err))
	}
	return catalog
}

func validateStainlessSteelSpokeModelReferenceCatalog(catalog StainlessSteelSpokeModelReferenceCatalog) error {
	if catalog.CatalogID == "" || catalog.CatalogVersion == "" {
		return fmt.Errorf("model catalog_id and catalog_version are required")
	}
	if len(catalog.Models) == 0 {
		return fmt.Errorf("at least one spoke model reference is required")
	}
	seenIDs := make(map[string]struct{}, len(catalog.Models))
	for _, model := range catalog.Models {
		if model.ID == "" || model.DisplayName == "" || model.GeometryDescription == "" || model.MaterialReferenceID == "" || model.DataVersion == "" || model.DataStatus == "" {
			return fmt.Errorf("model id, display_name, geometry_description, material_reference_id, data_version, and data_status are required")
		}
		if _, exists := seenIDs[model.ID]; exists {
			return fmt.Errorf("duplicate spoke model id %q", model.ID)
		}
		seenIDs[model.ID] = struct{}{}
		if !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(model.EffectiveAreaMM2) || model.EffectiveAreaMM2 <= minimumEffectiveSpokeAreaMM2 || model.EffectiveAreaMM2 > maximumEffectiveSpokeAreaMM2 {
			return fmt.Errorf("model %q effective_area_mm2 is outside the accepted physical envelope", model.ID)
		}
	}
	return nil
}

// GetStainlessSteelSpokeModelReferenceCatalog returns a defensive copy of the
// complete backend model catalogue.
func GetStainlessSteelSpokeModelReferenceCatalog() StainlessSteelSpokeModelReferenceCatalog {
	catalog := loadedStainlessSteelSpokeModelReferenceCatalog
	catalog.Models = cloneStainlessSteelSpokeModelReferences(catalog.Models)
	return catalog
}

// ListStainlessSteelSpokeDislocationMechanicsModelReferences returns fresh
// model values so API callers cannot mutate the embedded catalogue.
func ListStainlessSteelSpokeDislocationMechanicsModelReferences() []StainlessSteelSpokeModelReference {
	return cloneStainlessSteelSpokeModelReferences(loadedStainlessSteelSpokeModelReferenceCatalog.Models)
}

// GetStainlessSteelSpokeDislocationMechanicsModelReferenceByID resolves one
// backend model reference by its stable ID.
func GetStainlessSteelSpokeDislocationMechanicsModelReferenceByID(modelID string) (StainlessSteelSpokeModelReference, bool) {
	for _, model := range loadedStainlessSteelSpokeModelReferenceCatalog.Models {
		if model.ID == modelID {
			return cloneStainlessSteelSpokeModelReference(model), true
		}
	}
	return StainlessSteelSpokeModelReference{}, false
}

func cloneStainlessSteelSpokeModelReferences(models []StainlessSteelSpokeModelReference) []StainlessSteelSpokeModelReference {
	clonedModels := make([]StainlessSteelSpokeModelReference, len(models))
	for index, model := range models {
		clonedModels[index] = cloneStainlessSteelSpokeModelReference(model)
	}
	return clonedModels
}

func cloneStainlessSteelSpokeModelReference(model StainlessSteelSpokeModelReference) StainlessSteelSpokeModelReference {
	return model
}

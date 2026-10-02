package spokedislocationmechanics

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
)

//go:embed stainless_steel_spoke_material_reference_catalog.json
var embeddedStainlessSteelSpokeMaterialReferenceCatalogFile []byte

// StainlessSteelSpokeMaterialReferenceSource identifies the document or
// standard behind a material profile without implying that the source proves
// every numeric value in the profile.
type StainlessSteelSpokeMaterialReferenceSource struct {
	Title         string `json:"title"`
	ReferenceType string `json:"reference_type"`
	Location      string `json:"location,omitempty"`
	URL           string `json:"url,omitempty"`
	Role          string `json:"role"`
}

// StainlessSteelSpokeMaterialReference is a versioned, reviewable material
// profile used by the spoke stress and elastic-elongation calculation.
type StainlessSteelSpokeMaterialReference struct {
	ID                         string                                       `json:"id"`
	MaterialName               string                                       `json:"material_name"`
	MaterialFamily             string                                       `json:"material_family"`
	ManufacturingCondition     string                                       `json:"manufacturing_condition"`
	ElasticModulusMPA          float64                                      `json:"elastic_modulus_mpa"`
	MacroYieldReferenceMPA     float64                                      `json:"macro_yield_reference_mpa"`
	UltimateTensileStrengthMPA float64                                      `json:"ultimate_tensile_strength_reference_mpa"`
	DataVersion                string                                       `json:"data_version"`
	DataStatus                 string                                       `json:"data_status"`
	SourceReferences           []StainlessSteelSpokeMaterialReferenceSource `json:"source_references"`
	Applicability              string                                       `json:"applicability"`
	Notes                      string                                       `json:"notes"`
}

// StainlessSteelSpokeMaterialReferenceCatalog is the complete backend-owned
// material data set. New profiles can be added without changing the formula.
type StainlessSteelSpokeMaterialReferenceCatalog struct {
	CatalogID                  string                                 `json:"catalog_id"`
	CatalogVersion             string                                 `json:"catalog_version"`
	DefaultMaterialReferenceID string                                 `json:"default_material_reference_id"`
	Materials                  []StainlessSteelSpokeMaterialReference `json:"materials"`
}

var loadedStainlessSteelSpokeMaterialReferenceCatalog = loadEmbeddedStainlessSteelSpokeMaterialReferenceCatalogOrPanic()

// loadEmbeddedStainlessSteelSpokeMaterialReferenceCatalogOrPanic parses and
// validates the checked-in catalog once during package initialization. A
// malformed embedded catalog is a deploy-time programming error, so starting
// the backend is safer than silently calculating with incomplete material data.
func loadEmbeddedStainlessSteelSpokeMaterialReferenceCatalogOrPanic() StainlessSteelSpokeMaterialReferenceCatalog {
	var catalog StainlessSteelSpokeMaterialReferenceCatalog
	if err := json.Unmarshal(embeddedStainlessSteelSpokeMaterialReferenceCatalogFile, &catalog); err != nil {
		panic(fmt.Sprintf("load stainless steel spoke material reference catalog: %v", err))
	}
	if err := validateStainlessSteelSpokeMaterialReferenceCatalog(catalog); err != nil {
		panic(fmt.Sprintf("validate stainless steel spoke material reference catalog: %v", err))
	}
	return catalog
}

// validateStainlessSteelSpokeMaterialReferenceCatalog rejects incomplete data
// before it can become part of an API response or a physical calculation.
func validateStainlessSteelSpokeMaterialReferenceCatalog(catalog StainlessSteelSpokeMaterialReferenceCatalog) error {
	if catalog.CatalogID == "" || catalog.CatalogVersion == "" || catalog.DefaultMaterialReferenceID == "" {
		return fmt.Errorf("catalog_id, catalog_version, and default_material_reference_id are required")
	}
	if len(catalog.Materials) == 0 {
		return fmt.Errorf("at least one material profile is required")
	}
	seenIDs := make(map[string]struct{}, len(catalog.Materials))
	defaultProfileFound := false
	for _, material := range catalog.Materials {
		if material.ID == "" || material.MaterialName == "" || material.DataVersion == "" || material.DataStatus == "" {
			return fmt.Errorf("material id, material_name, data_version, and data_status are required")
		}
		if _, exists := seenIDs[material.ID]; exists {
			return fmt.Errorf("duplicate material profile id %q", material.ID)
		}
		seenIDs[material.ID] = struct{}{}
		if material.ID == catalog.DefaultMaterialReferenceID {
			defaultProfileFound = true
		}
		if !areFiniteStainlessSteelSpokeMaterialReferenceNumbers(material.ElasticModulusMPA, material.MacroYieldReferenceMPA, material.UltimateTensileStrengthMPA) {
			return fmt.Errorf("material %q contains a non-finite numeric value", material.ID)
		}
		if material.ElasticModulusMPA <= 0 || material.MacroYieldReferenceMPA <= 0 || material.UltimateTensileStrengthMPA <= 0 {
			return fmt.Errorf("material %q numeric values must be positive", material.ID)
		}
		if material.MacroYieldReferenceMPA >= material.UltimateTensileStrengthMPA {
			return fmt.Errorf("material %q must order yield reference < UTS reference", material.ID)
		}
		if len(material.SourceReferences) == 0 {
			return fmt.Errorf("material %q must include at least one source reference", material.ID)
		}
		for _, source := range material.SourceReferences {
			if source.Title == "" || source.ReferenceType == "" || source.Role == "" || (source.Location == "" && source.URL == "") {
				return fmt.Errorf("material %q contains an incomplete source reference", material.ID)
			}
		}
	}
	if !defaultProfileFound {
		return fmt.Errorf("default material reference %q was not found", catalog.DefaultMaterialReferenceID)
	}
	return nil
}

// GetStainlessSteelSpokeMaterialReferenceCatalog returns a defensive copy of
// the complete catalog so API callers cannot mutate package-owned slices.
func GetStainlessSteelSpokeMaterialReferenceCatalog() StainlessSteelSpokeMaterialReferenceCatalog {
	catalog := loadedStainlessSteelSpokeMaterialReferenceCatalog
	catalog.Materials = cloneStainlessSteelSpokeMaterialReferences(catalog.Materials)
	return catalog
}

// GetDefaultStainlessSteelSpokeMaterialReference returns the profile currently
// assigned to all models until verified brand/model-specific data is available.
func GetDefaultStainlessSteelSpokeMaterialReference() StainlessSteelSpokeMaterialReference {
	material, found := GetStainlessSteelSpokeMaterialReferenceByID(loadedStainlessSteelSpokeMaterialReferenceCatalog.DefaultMaterialReferenceID)
	if !found {
		panic("default stainless steel spoke material reference is missing")
	}
	return material
}

// getDefaultStainlessSteelSpokeMaterialReferenceID is used while every spoke
// model still shares the generic reference profile. A future verified mapping
// can replace this single lookup without changing the calculation formula.
func getDefaultStainlessSteelSpokeMaterialReferenceID() string {
	return loadedStainlessSteelSpokeMaterialReferenceCatalog.DefaultMaterialReferenceID
}

// GetStainlessSteelSpokeMaterialReferenceByID returns a defensive copy of a
// selected profile for future model-to-material mappings.
func GetStainlessSteelSpokeMaterialReferenceByID(materialReferenceID string) (StainlessSteelSpokeMaterialReference, bool) {
	for _, material := range loadedStainlessSteelSpokeMaterialReferenceCatalog.Materials {
		if material.ID == materialReferenceID {
			material.SourceReferences = append([]StainlessSteelSpokeMaterialReferenceSource(nil), material.SourceReferences...)
			return material, true
		}
	}
	return StainlessSteelSpokeMaterialReference{}, false
}

func cloneStainlessSteelSpokeMaterialReferences(materials []StainlessSteelSpokeMaterialReference) []StainlessSteelSpokeMaterialReference {
	clonedMaterials := make([]StainlessSteelSpokeMaterialReference, len(materials))
	copy(clonedMaterials, materials)
	for index := range clonedMaterials {
		clonedMaterials[index].SourceReferences = append([]StainlessSteelSpokeMaterialReferenceSource(nil), materials[index].SourceReferences...)
	}
	return clonedMaterials
}

func areFiniteStainlessSteelSpokeMaterialReferenceNumbers(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

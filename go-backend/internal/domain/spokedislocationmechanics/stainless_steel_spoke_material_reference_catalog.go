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

// StainlessSteelSpokeEngineeringStressStrainPoint is one published or
// explicitly digitized point on a monotonic engineering stress–strain curve.
// The curve is kept with the material profile so the axial calculator never
// silently replaces a missing curve with a generic yield or dislocation value.
type StainlessSteelSpokeEngineeringStressStrainPoint struct {
	EngineeringStrain    float64 `json:"engineering_strain"`
	EngineeringStressMPA float64 `json:"engineering_stress_mpa"`
}

// StainlessSteelSpokeMaterialReference is a versioned, reviewable material
// profile used by the spoke stress and elastic-elongation calculation.
type StainlessSteelSpokeMaterialReference struct {
	ID                              string                                            `json:"id"`
	MaterialName                    string                                            `json:"material_name"`
	MaterialFamily                  string                                            `json:"material_family"`
	ManufacturingCondition          string                                            `json:"manufacturing_condition"`
	ElasticModulusMPA               float64                                           `json:"elastic_modulus_mpa"`
	YieldStrengthMPA                float64                                           `json:"yield_strength_mpa"`
	UltimateTensileStrengthMPA      float64                                           `json:"ultimate_tensile_strength_mpa"`
	TotalElongationToFailurePercent float64                                           `json:"total_elongation_to_failure_percent"`
	CurveType                       string                                            `json:"curve_type"`
	StressStrainCurve               []StainlessSteelSpokeEngineeringStressStrainPoint `json:"stress_strain_curve"`
	DataVersion                     string                                            `json:"data_version"`
	DataStatus                      string                                            `json:"data_status"`
	SourceReferences                []StainlessSteelSpokeMaterialReferenceSource      `json:"source_references"`
	Applicability                   string                                            `json:"applicability"`
	Notes                           string                                            `json:"notes"`
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
		if !areFiniteStainlessSteelSpokeMaterialReferenceNumbers(material.ElasticModulusMPA, material.YieldStrengthMPA, material.UltimateTensileStrengthMPA, material.TotalElongationToFailurePercent) {
			return fmt.Errorf("material %q contains a non-finite numeric value", material.ID)
		}
		if material.ElasticModulusMPA <= 0 || material.YieldStrengthMPA <= 0 || material.UltimateTensileStrengthMPA <= 0 || material.TotalElongationToFailurePercent <= 0 {
			return fmt.Errorf("material %q numeric values must be positive", material.ID)
		}
		if material.YieldStrengthMPA >= material.UltimateTensileStrengthMPA {
			return fmt.Errorf("material %q must order yield strength < ultimate tensile strength", material.ID)
		}
		if material.CurveType == "" || len(material.StressStrainCurve) < 2 {
			return fmt.Errorf("material %q must include a stress-strain curve", material.ID)
		}
		if err := validateStainlessSteelSpokeEngineeringStressStrainCurve(material); err != nil {
			return err
		}
		lastCurvePoint := material.StressStrainCurve[len(material.StressStrainCurve)-1]
		if material.TotalElongationToFailurePercent < lastCurvePoint.EngineeringStrain*100 {
			return fmt.Errorf("material %q total elongation to failure must reach the curve endpoint strain", material.ID)
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

func validateStainlessSteelSpokeEngineeringStressStrainCurve(material StainlessSteelSpokeMaterialReference) error {
	previousStrain := -1.0
	previousStress := -1.0
	for index, point := range material.StressStrainCurve {
		if !areFiniteStainlessSteelSpokeMaterialReferenceNumbers(point.EngineeringStrain, point.EngineeringStressMPA) {
			return fmt.Errorf("material %q curve point %d contains a non-finite value", material.ID, index)
		}
		if point.EngineeringStrain < 0 || point.EngineeringStressMPA < 0 {
			return fmt.Errorf("material %q curve point %d must be non-negative", material.ID, index)
		}
		if point.EngineeringStrain <= previousStrain || point.EngineeringStressMPA <= previousStress {
			return fmt.Errorf("material %q curve must increase in strain and stress", material.ID)
		}
		previousStrain = point.EngineeringStrain
		previousStress = point.EngineeringStressMPA
	}
	firstPoint := material.StressStrainCurve[0]
	lastPoint := material.StressStrainCurve[len(material.StressStrainCurve)-1]
	if firstPoint.EngineeringStrain != 0 || firstPoint.EngineeringStressMPA != 0 {
		return fmt.Errorf("material %q curve must start at zero strain and stress", material.ID)
	}
	if math.Abs(lastPoint.EngineeringStressMPA-material.UltimateTensileStrengthMPA) > 1e-9 {
		return fmt.Errorf("material %q curve endpoint must equal ultimate tensile strength", material.ID)
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
// geometry model still references the selected AISI 304 material state. A
// future verified product mapping can replace this lookup without changing
// the curve calculation formula.
func getDefaultStainlessSteelSpokeMaterialReferenceID() string {
	return loadedStainlessSteelSpokeMaterialReferenceCatalog.DefaultMaterialReferenceID
}

// GetStainlessSteelSpokeMaterialReferenceByID returns a defensive copy of a
// selected profile for future model-to-material mappings.
func GetStainlessSteelSpokeMaterialReferenceByID(materialReferenceID string) (StainlessSteelSpokeMaterialReference, bool) {
	for _, material := range loadedStainlessSteelSpokeMaterialReferenceCatalog.Materials {
		if material.ID == materialReferenceID {
			material.SourceReferences = append([]StainlessSteelSpokeMaterialReferenceSource(nil), material.SourceReferences...)
			material.StressStrainCurve = append([]StainlessSteelSpokeEngineeringStressStrainPoint(nil), material.StressStrainCurve...)
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
		clonedMaterials[index].StressStrainCurve = append([]StainlessSteelSpokeEngineeringStressStrainPoint(nil), materials[index].StressStrainCurve...)
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

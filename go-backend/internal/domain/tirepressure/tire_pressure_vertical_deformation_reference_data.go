package tirepressure

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed tire_pressure_vertical_deformation_reference_data_v1.json
var tirePressureVerticalDeformationReferenceDataJSON []byte

// TirePressureVerticalDeformationReferenceData is the versioned, local source
// record used by the generic vertical deformation reference model.
type TirePressureVerticalDeformationReferenceData struct {
	SchemaVersion                 string                                     `json:"schema_version"`
	DatasetVersion                string                                     `json:"dataset_version"`
	DataStatus                    string                                     `json:"data_status"`
	Source                        TirePressureVerticalDeformationSource      `json:"source"`
	VerticalStiffnessReference    TirePressureVerticalStiffnessReference     `json:"vertical_stiffness_reference"`
	VerticalForceDeflectionPoints []TirePressureVerticalForceDeflectionPoint `json:"vertical_force_deflection_points"`
	ContactPatchGeometryPoints    []TirePressureContactPatchGeometryPoint    `json:"contact_patch_geometry_points"`
	CalculationLimits             []string                                   `json:"calculation_limits"`
}

type TirePressureVerticalDeformationSource struct {
	Authors             []string `json:"authors"`
	Title               string   `json:"title"`
	Journal             string   `json:"journal"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	Pages               string   `json:"pages"`
	PublicationYear     int      `json:"publication_year"`
	DOI                 string   `json:"doi"`
	SourceURL           string   `json:"source_url"`
	PublisherArticleURL string   `json:"publisher_article_url"`
	TestTireDescription string   `json:"test_tire_description"`
	TestMethod          string   `json:"test_method"`
	SourceAccessedOn    string   `json:"source_accessed_on"`
}

type TirePressureVerticalStiffnessReference struct {
	SourceFigure                  string  `json:"source_figure"`
	InflationPressureBar          float64 `json:"inflation_pressure_bar"`
	FittedVerticalStiffnessNPerMm float64 `json:"fitted_vertical_stiffness_n_per_mm"`
	FitDescription                string  `json:"fit_description"`
	PressureScope                 string  `json:"pressure_scope"`
	ExtractionMethod              string  `json:"extraction_method"`
}

type TirePressureVerticalForceDeflectionPoint struct {
	SourceFigure                       string  `json:"source_figure"`
	InflationPressureBar               float64 `json:"inflation_pressure_bar"`
	VerticalLoadN                      float64 `json:"vertical_load_n"`
	DeflectionMm                       float64 `json:"deflection_mm"`
	ExtractionMethod                   string  `json:"extraction_method"`
	EstimatedDigitizationUncertaintyMm float64 `json:"estimated_digitization_uncertainty_mm"`
}

type TirePressureContactPatchGeometryPoint struct {
	SourceFigure                       string  `json:"source_figure"`
	InflationPressureBar               float64 `json:"inflation_pressure_bar"`
	WheelLoadN                         float64 `json:"wheel_load_n"`
	PatchLengthMm                      float64 `json:"patch_length_mm"`
	PatchWidthMm                       float64 `json:"patch_width_mm"`
	ExtractionMethod                   string  `json:"extraction_method"`
	EstimatedDigitizationUncertaintyMm float64 `json:"estimated_digitization_uncertainty_mm"`
	FigureCell                         string  `json:"figure_cell"`
}

func loadTirePressureVerticalDeformationReferenceData() (TirePressureVerticalDeformationReferenceData, error) {
	var data TirePressureVerticalDeformationReferenceData
	if err := json.Unmarshal(tirePressureVerticalDeformationReferenceDataJSON, &data); err != nil {
		return TirePressureVerticalDeformationReferenceData{}, fmt.Errorf("decode vertical deformation reference data: %w", err)
	}
	if data.SchemaVersion == "" || data.DatasetVersion == "" || data.VerticalStiffnessReference.FittedVerticalStiffnessNPerMm <= 0 || data.VerticalStiffnessReference.InflationPressureBar <= 0 || len(data.VerticalForceDeflectionPoints) == 0 {
		return TirePressureVerticalDeformationReferenceData{}, fmt.Errorf("vertical deformation reference data is incomplete")
	}
	for index, point := range data.VerticalForceDeflectionPoints {
		if point.VerticalLoadN <= 0 || point.DeflectionMm <= 0 {
			return TirePressureVerticalDeformationReferenceData{}, fmt.Errorf("vertical deformation reference point %d is invalid", index)
		}
	}
	return data, nil
}

var tirePressureVerticalDeformationReferenceData = mustLoadTirePressureVerticalDeformationReferenceData()

func mustLoadTirePressureVerticalDeformationReferenceData() TirePressureVerticalDeformationReferenceData {
	data, err := loadTirePressureVerticalDeformationReferenceData()
	if err != nil {
		panic(err)
	}
	return data
}

// GetTirePressureVerticalDeformationReferenceData returns the immutable local
// dataset metadata used by the calculation and API response.
func GetTirePressureVerticalDeformationReferenceData() TirePressureVerticalDeformationReferenceData {
	return tirePressureVerticalDeformationReferenceData
}

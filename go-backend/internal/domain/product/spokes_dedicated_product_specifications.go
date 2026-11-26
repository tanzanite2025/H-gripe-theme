package product

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"
)

const (
	ProductTypeSpoke             = "spoke"
	SpokesProductCategorySlug    = "spokes"
	SpokeMiddleStrengthUnitLabel = "N/mm²"
)

type SpokesProductSpecification struct {
	SchemaVersion           int                       `json:"schema_version"`
	ModelName               string                    `json:"model_name"`
	ManufacturerPartNumber  string                    `json:"manufacturer_part_number,omitempty"`
	ManufacturerFamily      string                    `json:"manufacturer_family,omitempty"`
	Material                string                    `json:"material,omitempty"`
	ManufacturerDescription string                    `json:"manufacturer_description,omitempty"`
	GaugeOptions            []SpokeGaugeOption        `json:"gauge_options,omitempty"`
	MiddleStrength          *SpokeMiddleStrengthRange `json:"middle_strength_n_per_mm2,omitempty"`
	ColorOptions            []string                  `json:"color_options,omitempty"`
	ManufacturerOptions     []string                  `json:"manufacturer_options,omitempty"`
	RecommendedUses         []string                  `json:"recommended_uses,omitempty"`
	ManufacturerClaims      []string                  `json:"manufacturer_claims,omitempty"`
	SourceURLs              []string                  `json:"source_urls,omitempty"`
}

type SpokeGaugeOption struct {
	Gauge               string                `json:"gauge"`
	DiameterMM          *float64              `json:"diameter_mm,omitempty"`
	HubEndDiameterMM    *float64              `json:"hub_end_diameter_mm,omitempty"`
	MiddleDiameterMM    *float64              `json:"middle_diameter_mm,omitempty"`
	MiddleWidthMM       *float64              `json:"middle_width_mm,omitempty"`
	MiddleThicknessMM   *float64              `json:"middle_thickness_mm,omitempty"`
	ThreadEndDiameterMM *float64              `json:"thread_end_diameter_mm,omitempty"`
	LengthMinMM         *float64              `json:"length_min_mm,omitempty"`
	LengthMaxMM         *float64              `json:"length_max_mm,omitempty"`
	WeightReference     *SpokeWeightReference `json:"weight_reference,omitempty"`
}

type SpokeWeightReference struct {
	SpokeCount   int     `json:"spoke_count"`
	LengthMM     float64 `json:"length_mm"`
	TotalWeightG float64 `json:"total_weight_g"`
}

type SpokeMiddleStrengthRange struct {
	MinimumNPerMM2 float64 `json:"minimum_n_per_mm2"`
	MaximumNPerMM2 float64 `json:"maximum_n_per_mm2"`
}

func (spec *SpokesProductSpecification) UnmarshalJSON(data []byte) error {
	type spokesProductSpecificationJSON SpokesProductSpecification
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*spokesProductSpecificationJSON)(spec))
}

func (option *SpokeGaugeOption) UnmarshalJSON(data []byte) error {
	type spokeGaugeOptionJSON SpokeGaugeOption
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*spokeGaugeOptionJSON)(option))
}

func (reference *SpokeWeightReference) UnmarshalJSON(data []byte) error {
	type spokeWeightReferenceJSON SpokeWeightReference
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*spokeWeightReferenceJSON)(reference))
}

func (strength *SpokeMiddleStrengthRange) UnmarshalJSON(data []byte) error {
	type spokeMiddleStrengthRangeJSON SpokeMiddleStrengthRange
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*spokeMiddleStrengthRangeJSON)(strength))
}

func (spec *SpokesProductSpecification) Normalize() {
	if spec == nil {
		return
	}
	spec.ModelName = strings.TrimSpace(spec.ModelName)
	spec.ManufacturerPartNumber = strings.TrimSpace(spec.ManufacturerPartNumber)
	spec.ManufacturerFamily = strings.TrimSpace(spec.ManufacturerFamily)
	spec.Material = strings.TrimSpace(spec.Material)
	spec.ManufacturerDescription = strings.TrimSpace(spec.ManufacturerDescription)
	for index := range spec.GaugeOptions {
		spec.GaugeOptions[index].Gauge = strings.TrimSpace(spec.GaugeOptions[index].Gauge)
	}
	for _, values := range [][]string{spec.ColorOptions, spec.ManufacturerOptions, spec.RecommendedUses, spec.ManufacturerClaims, spec.SourceURLs} {
		for index := range values {
			values[index] = strings.TrimSpace(values[index])
		}
	}
}

func ValidateSpokesProductSpecification(spec *SpokesProductSpecification) error {
	if spec == nil || spec.SchemaVersion != 1 {
		return fmt.Errorf("spoke_specification requires schema_version 1")
	}
	if strings.TrimSpace(spec.ModelName) == "" || len(spec.ModelName) > 120 {
		return fmt.Errorf("model_name is required and must not exceed 120 bytes")
	}
	for _, field := range []struct {
		name  string
		value string
		limit int
	}{
		{"manufacturer_part_number", spec.ManufacturerPartNumber, 120},
		{"manufacturer_family", spec.ManufacturerFamily, 120},
		{"material", spec.Material, 160},
		{"manufacturer_description", spec.ManufacturerDescription, 4000},
	} {
		if len(field.value) > field.limit {
			return fmt.Errorf("%s exceeds %d bytes", field.name, field.limit)
		}
	}

	if spec.MiddleStrength != nil {
		if !isFinitePositiveSpokeMeasurement(spec.MiddleStrength.MinimumNPerMM2) || !isFinitePositiveSpokeMeasurement(spec.MiddleStrength.MaximumNPerMM2) {
			return fmt.Errorf("middle_strength_n_per_mm2 values must be finite positive numbers")
		}
		if spec.MiddleStrength.MinimumNPerMM2 > spec.MiddleStrength.MaximumNPerMM2 {
			return fmt.Errorf("middle_strength_n_per_mm2 minimum cannot exceed maximum")
		}
	}

	seenGauges := make(map[string]struct{}, len(spec.GaugeOptions))
	for index := range spec.GaugeOptions {
		option := &spec.GaugeOptions[index]
		if strings.TrimSpace(option.Gauge) == "" || len(option.Gauge) > 32 {
			return fmt.Errorf("gauge_options[%d].gauge is required and must not exceed 32 bytes", index)
		}
		gaugeKey := strings.ToLower(option.Gauge)
		if _, exists := seenGauges[gaugeKey]; exists {
			return fmt.Errorf("gauge_options cannot contain duplicate gauge values")
		}
		seenGauges[gaugeKey] = struct{}{}
		for _, measurement := range []struct {
			name  string
			value *float64
		}{
			{"diameter_mm", option.DiameterMM},
			{"hub_end_diameter_mm", option.HubEndDiameterMM},
			{"middle_diameter_mm", option.MiddleDiameterMM},
			{"middle_width_mm", option.MiddleWidthMM},
			{"middle_thickness_mm", option.MiddleThicknessMM},
			{"thread_end_diameter_mm", option.ThreadEndDiameterMM},
			{"length_min_mm", option.LengthMinMM},
			{"length_max_mm", option.LengthMaxMM},
		} {
			if measurement.value != nil && !isFinitePositiveSpokeMeasurement(*measurement.value) {
				return fmt.Errorf("gauge_options[%d].%s must be a finite positive number", index, measurement.name)
			}
		}
		if option.LengthMinMM != nil && option.LengthMaxMM != nil && *option.LengthMinMM > *option.LengthMaxMM {
			return fmt.Errorf("gauge_options[%d].length_min_mm cannot exceed length_max_mm", index)
		}
		if option.WeightReference != nil {
			reference := option.WeightReference
			if reference.SpokeCount <= 0 || !isFinitePositiveSpokeMeasurement(reference.LengthMM) || !isFinitePositiveSpokeMeasurement(reference.TotalWeightG) {
				return fmt.Errorf("gauge_options[%d].weight_reference requires a positive spoke_count, length_mm and total_weight_g", index)
			}
		}
	}

	for _, field := range []struct {
		name     string
		values   []string
		maxBytes int
	}{
		{"color_options", spec.ColorOptions, 120},
		{"manufacturer_options", spec.ManufacturerOptions, 240},
		{"recommended_uses", spec.RecommendedUses, 120},
		{"manufacturer_claims", spec.ManufacturerClaims, 500},
		{"source_urls", spec.SourceURLs, 1000},
	} {
		if err := validateSpokeTextValues(field.name, field.values, field.maxBytes); err != nil {
			return err
		}
	}
	for _, source := range spec.SourceURLs {
		parsed, err := url.Parse(source)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return fmt.Errorf("source_urls values must be absolute HTTP(S) URLs")
		}
	}
	return nil
}

func validateSpokeTextValues(name string, values []string, maxBytes int) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" || len(value) > maxBytes {
			return fmt.Errorf("%s values must be non-empty and not exceed %d bytes", name, maxBytes)
		}
		if _, exists := seen[normalized]; exists {
			return fmt.Errorf("%s cannot contain duplicate values", name)
		}
		seen[normalized] = struct{}{}
	}
	return nil
}

func isFinitePositiveSpokeMeasurement(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

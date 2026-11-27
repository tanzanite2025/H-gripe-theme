package product

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// GenericProductSchema is the optional product-level structured-data contract
// for ordinary products. It is deliberately separate from dedicated product
// type specifications and from SKU fulfillment facts.
type GenericProductSchema struct {
	SchemaVersion int      `json:"schema_version"`
	Model         string   `json:"model,omitempty"`
	Material      string   `json:"material,omitempty"`
	Color         string   `json:"color,omitempty"`
	Size          string   `json:"size,omitempty"`
	WeightG       *float64 `json:"weight_g,omitempty"`
}

func (spec *GenericProductSchema) UnmarshalJSON(data []byte) error {
	type genericProductSchemaJSON GenericProductSchema
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*genericProductSchemaJSON)(spec))
}

const GenericProductSchemaVersion = 1

func (spec *GenericProductSchema) Normalize() {
	if spec == nil {
		return
	}
	spec.Model = strings.TrimSpace(spec.Model)
	spec.Material = strings.TrimSpace(spec.Material)
	spec.Color = strings.TrimSpace(spec.Color)
	spec.Size = strings.TrimSpace(spec.Size)
	if spec.WeightG != nil {
		weight := math.Round(*spec.WeightG*100) / 100
		spec.WeightG = &weight
	}
}

func ValidateGenericProductSchema(spec *GenericProductSchema) error {
	if spec == nil {
		return nil
	}
	spec.Normalize()
	if spec.SchemaVersion != GenericProductSchemaVersion {
		return fmt.Errorf("generic product schema_version must be %d", GenericProductSchemaVersion)
	}
	for field, value := range map[string]string{
		"model":    spec.Model,
		"material": spec.Material,
		"color":    spec.Color,
		"size":     spec.Size,
	} {
		limit := limitForGenericProductSchemaField(field, 120)
		if len([]byte(value)) > limit {
			return fmt.Errorf("generic product %s exceeds %d bytes", field, limit)
		}
	}
	if spec.WeightG != nil && (math.IsNaN(*spec.WeightG) || math.IsInf(*spec.WeightG, 0) || *spec.WeightG <= 0 || *spec.WeightG > 1000000) {
		return fmt.Errorf("generic product weight_g must be greater than 0 and at most 1000000")
	}
	return nil
}

func limitForGenericProductSchemaField(field string, fallback int) int {
	switch field {
	case "model":
		return 160
	case "material":
		return 160
	case "color":
		return 120
	case "size":
		return 120
	default:
		return fallback
	}
}

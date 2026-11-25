package product

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const (
	ProductTypeBicycleTire          = "bicycle_tire"
	BicycleTireProductCategorySlug = "tire"
)

type BicycleTireSpecification struct {
	SchemaVersion   int      `json:"schema_version"`
	ModelName       string   `json:"model_name,omitempty"`
	ArticleNo       string   `json:"article_no,omitempty"`
	ETRTO           string   `json:"etrto,omitempty"`
	InchDesignation string   `json:"inch_designation,omitempty"`
	WeightG         *float64 `json:"weight_g,omitempty"`
	VersionLabel    string   `json:"version_label,omitempty"`
	Compound        string   `json:"compound,omitempty"`
	Color           string   `json:"color,omitempty"`
	Bead            string   `json:"bead,omitempty"`
	EBikeRating     string   `json:"e_bike_rating,omitempty"`
	EPI             *int     `json:"epi,omitempty"`
	LoadKG          *float64 `json:"load_kg,omitempty"`
	Seal            string   `json:"seal,omitempty"`
	Tread           string   `json:"tread,omitempty"`
	MinPressureBar  *float64 `json:"min_pressure_bar,omitempty"`
	MaxPressureBar  *float64 `json:"max_pressure_bar,omitempty"`
	MinPressurePSI  *float64 `json:"min_pressure_psi,omitempty"`
	MaxPressurePSI  *float64 `json:"max_pressure_psi,omitempty"`
}

func (spec *BicycleTireSpecification) UnmarshalJSON(data []byte) error {
	type bicycleTireSpecificationJSON BicycleTireSpecification
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*bicycleTireSpecificationJSON)(spec))
}

func (spec *BicycleTireSpecification) Normalize() {
	if spec == nil {
		return
	}
	spec.ModelName = strings.TrimSpace(spec.ModelName)
	spec.ArticleNo = strings.TrimSpace(spec.ArticleNo)
	spec.ETRTO = strings.TrimSpace(spec.ETRTO)
	spec.InchDesignation = strings.TrimSpace(spec.InchDesignation)
	spec.VersionLabel = strings.TrimSpace(spec.VersionLabel)
	spec.Compound = strings.TrimSpace(spec.Compound)
	spec.Color = strings.TrimSpace(spec.Color)
	spec.Bead = strings.TrimSpace(spec.Bead)
	spec.EBikeRating = strings.TrimSpace(spec.EBikeRating)
	spec.Seal = strings.TrimSpace(spec.Seal)
	spec.Tread = strings.TrimSpace(spec.Tread)
}

func ValidateBicycleTireSpecification(spec *BicycleTireSpecification) error {
	if spec == nil || spec.SchemaVersion != 1 {
		return fmt.Errorf("bicycle_tire_specification requires schema_version 1")
	}
	for _, field := range []struct {
		name  string
		value string
		limit int
	}{
		{"model_name", spec.ModelName, 180},
		{"article_no", spec.ArticleNo, 70},
		{"etrto", spec.ETRTO, 32},
		{"inch_designation", spec.InchDesignation, 32},
		{"version_label", spec.VersionLabel, 120},
		{"compound", spec.Compound, 120},
		{"color", spec.Color, 120},
		{"bead", spec.Bead, 64},
		{"e_bike_rating", spec.EBikeRating, 64},
		{"seal", spec.Seal, 120},
		{"tread", spec.Tread, 120},
	} {
		if len(field.value) > field.limit {
			return fmt.Errorf("%s exceeds %d bytes", field.name, field.limit)
		}
	}
	for _, field := range []struct {
		name  string
		value *float64
	}{
		{"weight_g", spec.WeightG},
		{"load_kg", spec.LoadKG},
		{"min_pressure_bar", spec.MinPressureBar},
		{"max_pressure_bar", spec.MaxPressureBar},
		{"min_pressure_psi", spec.MinPressurePSI},
		{"max_pressure_psi", spec.MaxPressurePSI},
	} {
		if field.value != nil && (*field.value <= 0 || math.IsNaN(*field.value) || math.IsInf(*field.value, 0)) {
			return fmt.Errorf("%s must be a finite positive number when provided", field.name)
		}
	}
	if spec.EPI != nil && *spec.EPI <= 0 {
		return fmt.Errorf("epi must be a positive integer when provided")
	}
	if spec.MinPressureBar != nil && spec.MaxPressureBar != nil && *spec.MinPressureBar > *spec.MaxPressureBar {
		return fmt.Errorf("min_pressure_bar cannot exceed max_pressure_bar")
	}
	if spec.MinPressurePSI != nil && spec.MaxPressurePSI != nil && *spec.MinPressurePSI > *spec.MaxPressurePSI {
		return fmt.Errorf("min_pressure_psi cannot exceed max_pressure_psi")
	}
	return nil
}

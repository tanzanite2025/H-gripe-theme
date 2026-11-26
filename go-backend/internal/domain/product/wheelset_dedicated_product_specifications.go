package product

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const ProductTypeWheelset = "wheelset"

const (
	WheelsetFilterKeyWheelSize      = "wheel_size"
	WheelsetFilterKeyFrontWheelSize = "front_wheel_size"
	WheelsetFilterKeyRearWheelSize  = "rear_wheel_size"
	WheelsetFilterKeyFrontAxle      = "front_axle"
	WheelsetFilterKeyRearAxle       = "rear_axle"
	WheelsetFilterKeyBrakeMount     = "brake_mount"
	WheelsetFilterKeyFreehub        = "freehub"
)

type WheelsetSpecificationFilterDefinition struct {
	Slug                     string
	Name                     string
	FieldType                string
	Presentation             string
	Unit                     string
	JSONPaths                [][]string
	RequiresSameValueInPaths bool
}

func WheelsetProductLevelSpecificationFilterDefinitions() []WheelsetSpecificationFilterDefinition {
	return []WheelsetSpecificationFilterDefinition{
		{
			Slug:                     WheelsetFilterKeyWheelSize,
			Name:                     "Wheel size",
			FieldType:                "select",
			Presentation:             "select",
			JSONPaths:                [][]string{{"front_wheel", "wheel_size"}, {"rear_wheel", "wheel_size"}},
			RequiresSameValueInPaths: true,
		},
		{
			Slug:         WheelsetFilterKeyFrontWheelSize,
			Name:         "Front wheel size",
			FieldType:    "select",
			Presentation: "select",
			JSONPaths:    [][]string{{"front_wheel", "wheel_size"}},
		},
		{
			Slug:         WheelsetFilterKeyRearWheelSize,
			Name:         "Rear wheel size",
			FieldType:    "select",
			Presentation: "select",
			JSONPaths:    [][]string{{"rear_wheel", "wheel_size"}},
		},
		{
			Slug:         WheelsetFilterKeyFrontAxle,
			Name:         "Front axle",
			FieldType:    "select",
			Presentation: "select",
			JSONPaths:    [][]string{{"front_wheel", "axle_standard"}},
		},
		{
			Slug:         WheelsetFilterKeyRearAxle,
			Name:         "Rear axle",
			FieldType:    "select",
			Presentation: "select",
			JSONPaths:    [][]string{{"rear_wheel", "axle_standard"}},
		},
		{
			Slug:                     WheelsetFilterKeyBrakeMount,
			Name:                     "Brake interface",
			FieldType:                "select",
			Presentation:             "select",
			JSONPaths:                [][]string{{"front_wheel", "brake_interface"}, {"rear_wheel", "brake_interface"}},
			RequiresSameValueInPaths: true,
		},
		{
			Slug:         WheelsetFilterKeyFreehub,
			Name:         "Freehub body",
			FieldType:    "select",
			Presentation: "select",
			JSONPaths:    [][]string{{"rear_wheel", "freehub_body"}},
		},
	}
}

type WheelsetSpecification struct {
	SchemaVersion     int                               `json:"schema_version"`
	FrontWheel        WheelsetProductWheelSpecification `json:"front_wheel"`
	RearWheel         WheelsetProductWheelSpecification `json:"rear_wheel"`
	NetWeightG        *float64                          `json:"net_weight_g,omitempty"`
	MaxSystemWeightKG *float64                          `json:"max_system_weight_kg,omitempty"`
	Certifications    []string                          `json:"certifications,omitempty"`
}

type WheelsetProductWheelSpecification struct {
	WheelSize          string   `json:"wheel_size"`
	SpokeHoleCounts    []int    `json:"spoke_hole_counts"`
	LacingPatterns     []string `json:"lacing_patterns"`
	BeadSeatDiameterMM *float64 `json:"bead_seat_diameter_mm,omitempty"`
	Material           string   `json:"material,omitempty"`
	BrakeInterface     string   `json:"brake_interface,omitempty"`
	RimProfile         string   `json:"rim_profile,omitempty"`
	TireSystem         string   `json:"tire_system,omitempty"`
	RimDepthMM         *float64 `json:"rim_depth_mm,omitempty"`
	RimInternalWidthMM *float64 `json:"rim_internal_width_mm,omitempty"`
	RimExternalWidthMM *float64 `json:"rim_external_width_mm,omitempty"`
	HubInterface       string   `json:"hub_interface,omitempty"`
	AxleStandard       string   `json:"axle_standard,omitempty"`
	FreehubBody        string   `json:"freehub_body,omitempty"`
	SpokeType          string   `json:"spoke_type,omitempty"`
	WeightG            *float64 `json:"weight_g,omitempty"`
	MaxTirePressureBar *float64 `json:"max_tire_pressure_bar,omitempty"`
}

func (spec *WheelsetSpecification) UnmarshalJSON(data []byte) error {
	type wheelsetSpecificationJSON WheelsetSpecification
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*wheelsetSpecificationJSON)(spec))
}

func (spec *WheelsetProductWheelSpecification) UnmarshalJSON(data []byte) error {
	type wheelsetProductWheelSpecificationJSON WheelsetProductWheelSpecification
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*wheelsetProductWheelSpecificationJSON)(spec))
}

func (spec *WheelsetSpecification) Normalize() {
	if spec == nil {
		return
	}
	spec.FrontWheel.normalize()
	spec.RearWheel.normalize()
	for index := range spec.Certifications {
		spec.Certifications[index] = strings.TrimSpace(spec.Certifications[index])
	}
}

func (spec *WheelsetProductWheelSpecification) normalize() {
	spec.WheelSize = strings.TrimSpace(spec.WheelSize)
	for index := range spec.LacingPatterns {
		spec.LacingPatterns[index] = strings.TrimSpace(spec.LacingPatterns[index])
	}
	spec.Material = strings.TrimSpace(spec.Material)
	spec.BrakeInterface = strings.TrimSpace(spec.BrakeInterface)
	spec.RimProfile = strings.TrimSpace(spec.RimProfile)
	spec.TireSystem = strings.TrimSpace(spec.TireSystem)
	spec.HubInterface = strings.TrimSpace(spec.HubInterface)
	spec.AxleStandard = strings.TrimSpace(spec.AxleStandard)
	spec.FreehubBody = strings.TrimSpace(spec.FreehubBody)
	spec.SpokeType = strings.TrimSpace(spec.SpokeType)
}

func ValidateWheelsetSpecification(spec *WheelsetSpecification) error {
	if spec == nil || spec.SchemaVersion != 1 {
		return fmt.Errorf("wheelset_specification requires schema_version 1")
	}
	if strings.TrimSpace(spec.FrontWheel.FreehubBody) != "" {
		return fmt.Errorf("front_wheel.freehub_body is only supported on rear_wheel")
	}
	if err := validateWheelsetProductWheelSpecification("front_wheel", &spec.FrontWheel); err != nil {
		return err
	}
	if err := validateWheelsetProductWheelSpecification("rear_wheel", &spec.RearWheel); err != nil {
		return err
	}
	for _, field := range []struct {
		name  string
		value *float64
	}{
		{"net_weight_g", spec.NetWeightG},
		{"max_system_weight_kg", spec.MaxSystemWeightKG},
	} {
		if err := validateOptionalPositiveWheelsetMeasurement(field.name, field.value); err != nil {
			return err
		}
	}
	return validateWheelsetTextValues("certifications", spec.Certifications, 160)
}

func validateWheelsetProductWheelSpecification(position string, spec *WheelsetProductWheelSpecification) error {
	if spec == nil || strings.TrimSpace(spec.WheelSize) == "" || len(spec.WheelSize) > 32 {
		return fmt.Errorf("%s.wheel_size is required and must not exceed 32 bytes", position)
	}
	if len(spec.SpokeHoleCounts) == 0 {
		return fmt.Errorf("%s.spoke_hole_counts requires at least one value", position)
	}
	seenHoleCounts := make(map[int]struct{}, len(spec.SpokeHoleCounts))
	for _, count := range spec.SpokeHoleCounts {
		if count < 1 || count > 128 {
			return fmt.Errorf("%s.spoke_hole_counts values must be between 1 and 128", position)
		}
		if _, exists := seenHoleCounts[count]; exists {
			return fmt.Errorf("%s.spoke_hole_counts cannot contain duplicates", position)
		}
		seenHoleCounts[count] = struct{}{}
	}
	if err := validateWheelsetTextValues(position+".lacing_patterns", spec.LacingPatterns, 64); err != nil {
		return err
	}
	if len(spec.LacingPatterns) == 0 {
		return fmt.Errorf("%s.lacing_patterns requires at least one value", position)
	}
	for _, field := range []struct {
		name  string
		value *float64
	}{
		{"bead_seat_diameter_mm", spec.BeadSeatDiameterMM},
		{"rim_depth_mm", spec.RimDepthMM},
		{"rim_internal_width_mm", spec.RimInternalWidthMM},
		{"rim_external_width_mm", spec.RimExternalWidthMM},
		{"weight_g", spec.WeightG},
		{"max_tire_pressure_bar", spec.MaxTirePressureBar},
	} {
		if err := validateOptionalPositiveWheelsetMeasurement(position+"."+field.name, field.value); err != nil {
			return err
		}
	}
	if spec.RimInternalWidthMM != nil && spec.RimExternalWidthMM != nil && *spec.RimExternalWidthMM <= *spec.RimInternalWidthMM {
		return fmt.Errorf("%s.rim_external_width_mm must exceed rim_internal_width_mm", position)
	}
	for _, field := range []struct {
		name  string
		value string
		limit int
	}{
		{"material", spec.Material, 120},
		{"brake_interface", spec.BrakeInterface, 80},
		{"rim_profile", spec.RimProfile, 80},
		{"tire_system", spec.TireSystem, 80},
		{"hub_interface", spec.HubInterface, 120},
		{"axle_standard", spec.AxleStandard, 120},
		{"freehub_body", spec.FreehubBody, 120},
		{"spoke_type", spec.SpokeType, 120},
	} {
		if len(field.value) > field.limit {
			return fmt.Errorf("%s.%s exceeds %d bytes", position, field.name, field.limit)
		}
	}
	return nil
}

func validateWheelsetTextValues(name string, values []string, maxLength int) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" || len(value) > maxLength {
			return fmt.Errorf("%s values must be non-empty and not exceed %d bytes", name, maxLength)
		}
		if _, exists := seen[normalized]; exists {
			return fmt.Errorf("%s cannot contain duplicate values", name)
		}
		seen[normalized] = struct{}{}
	}
	return nil
}

func validateOptionalPositiveWheelsetMeasurement(name string, value *float64) error {
	if value != nil && (*value <= 0 || math.IsNaN(*value) || math.IsInf(*value, 0)) {
		return fmt.Errorf("%s must be a finite positive number when provided", name)
	}
	return nil
}

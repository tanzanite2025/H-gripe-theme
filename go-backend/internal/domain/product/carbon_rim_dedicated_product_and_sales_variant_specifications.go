package product

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

const ProductTypeCarbonRim = "carbon_rim"

// CarbonRimSpecification is a versioned, typed product-owned document. It is
// stored atomically with the product, independently of specification templates.
type CarbonRimSpecification struct {
	WheelSize           string   `json:"wheel_size,omitempty"`
	Material            string   `json:"material,omitempty"`
	MaterialDetails     string   `json:"material_details,omitempty"`
	RidingUses          []string `json:"riding_uses,omitempty"`
	SchemaVersion       int      `json:"schema_version"`
	Model               string   `json:"model"`
	BeadSeatDiameterMM  float64  `json:"bead_seat_diameter_mm"`
	DepthMM             float64  `json:"depth_mm"`
	InternalWidthMM     float64  `json:"internal_width_mm"`
	ExternalWidthMM     float64  `json:"external_width_mm"`
	BeadStructure       string   `json:"bead_structure"`
	TubelessReady       *bool    `json:"tubeless_ready"`
	RimTapeRequired     *bool    `json:"rim_tape_required,omitempty"`
	ValveHoleDiameterMM *float64 `json:"valve_hole_diameter_mm,omitempty"`
	AsymmetricOffsetMM  float64  `json:"asymmetric_offset_mm"`
	TechnicalNotes      string   `json:"technical_notes,omitempty"`
}

type CarbonRimVariantSpecification struct {
	SpokeHoleCount      int      `json:"spoke_hole_count"`
	DrillingPattern     string   `json:"drilling_pattern"`
	FinishCode          string   `json:"finish_code"`
	ERDMM               *float64 `json:"erd_mm,omitempty"`
	ERDMeasurementNotes string   `json:"erd_measurement_notes,omitempty"`
	NetWeightGrams      float64  `json:"net_weight_grams"`
}

// Reject misspelled or unimplemented engineering fields rather than accepting
// a successful save which silently discarded a manufacturer's specification.
func (spec *CarbonRimSpecification) UnmarshalJSON(data []byte) error {
	type carbonRimSpecificationJSON CarbonRimSpecification
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*carbonRimSpecificationJSON)(spec))
}

func (spec *CarbonRimVariantSpecification) UnmarshalJSON(data []byte) error {
	type carbonRimVariantSpecificationJSON CarbonRimVariantSpecification
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*carbonRimVariantSpecificationJSON)(spec))
}

func ValidateCarbonRimSpecification(spec *CarbonRimSpecification) error {
	if spec == nil || spec.SchemaVersion != 1 || strings.TrimSpace(spec.Model) == "" || len(spec.Model) > 120 {
		return fmt.Errorf("carbon_rim_specification requires schema_version 1 and a model (max 120 characters)")
	}
	for name, value := range map[string]float64{"bead_seat_diameter_mm": spec.BeadSeatDiameterMM, "depth_mm": spec.DepthMM, "internal_width_mm": spec.InternalWidthMM, "external_width_mm": spec.ExternalWidthMM} {
		if !isFinitePositiveCarbonRimMeasurement(value) {
			return fmt.Errorf("%s must be a finite positive number", name)
		}
	}
	if spec.ExternalWidthMM <= spec.InternalWidthMM || spec.DepthMM >= spec.BeadSeatDiameterMM/2 {
		return fmt.Errorf("external_width_mm must exceed internal_width_mm; depth_mm must be less than the bead seat radius")
	}
	if spec.BeadStructure != "hooked" && spec.BeadStructure != "hookless" {
		return fmt.Errorf("bead_structure must be hooked or hookless")
	}
	if spec.TubelessReady == nil {
		return fmt.Errorf("tubeless_ready must explicitly be true or false")
	}
	if math.IsNaN(spec.AsymmetricOffsetMM) || math.IsInf(spec.AsymmetricOffsetMM, 0) || spec.AsymmetricOffsetMM < 0 || spec.AsymmetricOffsetMM >= spec.ExternalWidthMM/2 {
		return fmt.Errorf("asymmetric_offset_mm must be non-negative and less than half the external width")
	}
	if spec.ValveHoleDiameterMM != nil && !isFinitePositiveCarbonRimMeasurement(*spec.ValveHoleDiameterMM) {
		return fmt.Errorf("valve_hole_diameter_mm must be positive when provided")
	}
	if len(spec.TechnicalNotes) > 4000 {
		return fmt.Errorf("technical_notes exceeds 4000 bytes")
	}
	return ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, false)
}

var carbonRimFinishCodePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateCarbonRimVariantSpecification(spec *CarbonRimVariantSpecification) error {
	if spec == nil {
		return fmt.Errorf("each carbon rim SKU requires carbon_rim_specification")
	}
	if spec.SpokeHoleCount <= 0 || spec.SpokeHoleCount > 128 {
		return fmt.Errorf("spoke_hole_count must be between 1 and 128")
	}
	switch spec.DrillingPattern {
	case "uniform", "paired", "2-to-1", "g3":
	default:
		return fmt.Errorf("unsupported drilling_pattern")
	}
	if !carbonRimFinishCodePattern.MatchString(spec.FinishCode) || len(spec.FinishCode) > 80 {
		return fmt.Errorf("finish_code must be a lowercase hyphen-separated code (max 80 characters)")
	}
	if !isFinitePositiveCarbonRimMeasurement(spec.NetWeightGrams) {
		return fmt.Errorf("net_weight_grams must be positive")
	}
	if spec.ERDMM != nil && (!isFinitePositiveCarbonRimMeasurement(*spec.ERDMM) || strings.TrimSpace(spec.ERDMeasurementNotes) == "") {
		return fmt.Errorf("erd_mm requires a positive value and erd_measurement_notes")
	}
	if len(spec.ERDMeasurementNotes) > 1000 {
		return fmt.Errorf("erd_measurement_notes exceeds 1000 bytes")
	}
	return nil
}

func isFinitePositiveCarbonRimMeasurement(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

// One complete actual SKU choice avoids generating unavailable combinations.
func CarbonRimSalesConfigurationLabel(spec *CarbonRimVariantSpecification) string {
	if spec == nil {
		return ""
	}
	return fmt.Sprintf("%dH / %s / %s", spec.SpokeHoleCount, spec.DrillingPattern, spec.FinishCode)
}

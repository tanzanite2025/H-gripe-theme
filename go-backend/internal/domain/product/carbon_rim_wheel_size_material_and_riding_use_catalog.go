package product

import (
	"fmt"
	"strings"
)

type CarbonRimWheelSizeDefinition struct {
	BeadSeatDiameterMM float64
	Label              string
}

var carbonRimWheelSizeDefinitions = map[string]CarbonRimWheelSizeDefinition{
	"700c": {622, "700C"}, "29-inch": {622, "29 inch"},
	"650b": {584, "650B"}, "27.5-inch": {584, "27.5 inch"},
	"650c": {571, "650C"}, "26-inch": {559, "26 inch"},
	"24-inch": {507, "24 inch"}, "20-inch-406": {406, "20 inch (406 mm BSD)"},
	"20-inch-451": {451, "20 inch (451 mm BSD)"},
}

var carbonRimRidingUseLabels = map[string][2]string{
	"road": {"Road cycling", "公路骑行"}, "gravel": {"Gravel cycling", "Gravel 碎石骑行"},
	"cyclocross": {"Cyclocross", "公路越野 CX"}, "cross-country": {"Cross-country MTB", "山地越野 XC"},
	"trail": {"Trail MTB", "山地林道 Trail"}, "enduro": {"Enduro MTB", "山地耐力 Enduro"},
	"downhill": {"Downhill MTB", "山地下坡 DH"}, "triathlon-time-trial": {"Triathlon / time trial", "铁人三项 / 计时赛"},
	"touring-bikepacking": {"Touring / bikepacking", "旅行 / Bikepacking"}, "urban-commuting": {"Urban commuting", "城市通勤"},
}

func CarbonRimWheelSizeDisplayLabel(wheelSize string) string {
	return carbonRimWheelSizeDefinitions[wheelSize].Label
}

func CarbonRimRidingUseDisplayLabels(ridingUses []string, locale string) []string {
	labels := make([]string, 0, len(ridingUses))
	for _, ridingUse := range ridingUses {
		if names, exists := carbonRimRidingUseLabels[ridingUse]; exists {
			if locale == "zh_cn" {
				labels = append(labels, names[1])
			} else {
				labels = append(labels, names[0])
			}
		}
	}
	return labels
}

// New products require these facts. Older version-1 documents may omit them
// until explicitly completed; a 622 mm BSD must not invent 700C vs 29-inch use.
func ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec *CarbonRimSpecification, requireComplete bool) error {
	if spec == nil {
		return fmt.Errorf("carbon rim specification is required")
	}
	if requireComplete && (spec.WheelSize == "" || spec.Material == "" || len(spec.RidingUses) == 0) {
		return fmt.Errorf("wheel_size, material and at least one riding_uses value are required")
	}
	if spec.WheelSize != "" {
		definition, exists := carbonRimWheelSizeDefinitions[spec.WheelSize]
		if !exists || definition.BeadSeatDiameterMM != spec.BeadSeatDiameterMM {
			return fmt.Errorf("wheel_size must be supported and agree with bead_seat_diameter_mm")
		}
	}
	if spec.Material != "" && spec.Material != "carbon_fiber" {
		return fmt.Errorf("carbon rim material must be carbon_fiber")
	}
	if strings.TrimSpace(spec.MaterialDetails) != "" && spec.Material != "carbon_fiber" {
		return fmt.Errorf("material_details requires explicit carbon_fiber material")
	}
	if len(spec.MaterialDetails) > 500 {
		return fmt.Errorf("material_details exceeds 500 bytes")
	}
	seen := make(map[string]bool)
	for _, ridingUse := range spec.RidingUses {
		if _, exists := carbonRimRidingUseLabels[ridingUse]; !exists || seen[ridingUse] {
			return fmt.Errorf("riding_uses contains an unknown or duplicate riding discipline")
		}
		seen[ridingUse] = true
	}
	return nil
}

package product

import (
	"fmt"
	"strconv"
	"strings"

	productdomain "commerce-platform/internal/domain/product"
)

func publicSpokesProductSpecificationValues(specification *productdomain.SpokesProductSpecification, locale string) []PublicProductSpecValue {
	if specification == nil {
		return nil
	}
	rows := make([]PublicProductSpecValue, 0, 32)
	appendRow := func(slug, english, chinese, value, unit string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		name, group := english, "Spoke specifications"
		if locale == "zh_cn" {
			name, group = chinese, "辐条规格"
		}
		rows = append(rows, PublicProductSpecValue{
			Value: value,
			Definition: &PublicSpecDefinition{
				Slug: slug, Name: name, Group: group, Unit: unit,
				FieldType: "text", Role: "attribute", IsVisible: true,
			},
		})
	}
	appendRow("model_name", "Model", "型号", specification.ModelName, "")
	appendRow("manufacturer_family", "Manufacturer family", "制造商系列", specification.ManufacturerFamily, "")
	appendRow("material", "Material", "材质", specification.Material, "")
	appendRow("manufacturer_description", "Manufacturer description", "制造商说明", specification.ManufacturerDescription, "")
	if specification.MiddleStrength != nil {
		minimum := formatSpokesMeasurement(specification.MiddleStrength.MinimumNPerMM2)
		maximum := formatSpokesMeasurement(specification.MiddleStrength.MaximumNPerMM2)
		value := minimum
		if minimum != maximum {
			value = minimum + " - " + maximum
		}
		appendRow("middle_strength_n_per_mm2", "Middle strength", "中段强度", value, productdomain.SpokeMiddleStrengthUnitLabel)
	}

	for index, gauge := range specification.GaugeOptions {
		prefix := fmt.Sprintf("gauge-%d", index+1)
		label := strings.TrimSpace(gauge.Gauge)
		appendRow(prefix+"-label", "Gauge", "线规", label, "")
		appendGaugeMeasurement := func(suffix, english, chinese string, value *float64) {
			if value == nil {
				return
			}
			appendRow(prefix+"-"+suffix, label+" "+english, label+chinese, formatSpokesMeasurement(*value), "mm")
		}
		appendGaugeMeasurement("diameter-mm", "diameter", "直径", gauge.DiameterMM)
		appendGaugeMeasurement("hub-end-diameter-mm", "hub-end diameter", "花鼓端直径", gauge.HubEndDiameterMM)
		appendGaugeMeasurement("middle-diameter-mm", "middle diameter", "中段直径", gauge.MiddleDiameterMM)
		appendGaugeMeasurement("middle-width-mm", "middle width", "中段宽度", gauge.MiddleWidthMM)
		appendGaugeMeasurement("middle-thickness-mm", "middle thickness", "中段厚度", gauge.MiddleThicknessMM)
		appendGaugeMeasurement("thread-end-diameter-mm", "thread-end diameter", "螺纹端直径", gauge.ThreadEndDiameterMM)
		if gauge.LengthMinMM != nil || gauge.LengthMaxMM != nil {
			minimum, maximum := "", ""
			if gauge.LengthMinMM != nil {
				minimum = formatSpokesMeasurement(*gauge.LengthMinMM)
			}
			if gauge.LengthMaxMM != nil {
				maximum = formatSpokesMeasurement(*gauge.LengthMaxMM)
			}
			value := minimum
			if minimum != "" && maximum != "" && minimum != maximum {
				value = minimum + " - " + maximum
			} else if maximum != "" {
				value = maximum
			}
			appendRow(prefix+"-length-range-mm", label+" length range", label+"长度范围", value, "mm")
		}
		if gauge.WeightReference != nil {
			reference := gauge.WeightReference
			unit := fmt.Sprintf("g (%d spokes, %s mm each)", reference.SpokeCount, formatSpokesMeasurement(reference.LengthMM))
			appendRow(prefix+"-weight-reference", label+" weight reference", label+"重量参考", formatSpokesMeasurement(reference.TotalWeightG), unit)
		}
	}
	for _, value := range specification.ColorOptions {
		appendRow("color-option", "Color option", "颜色选项", value, "")
	}
	for _, value := range specification.ManufacturerOptions {
		appendRow("manufacturer-option", "Manufacturer option", "制造商选项", value, "")
	}
	for _, value := range specification.RecommendedUses {
		appendRow("recommended-use", "Manufacturer recommended use", "制造商推荐用途", value, "")
	}
	for _, value := range specification.ManufacturerClaims {
		appendRow("manufacturer-claim", "Manufacturer claim", "制造商宣称", value, "")
	}
	return rows
}

func formatSpokesMeasurement(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

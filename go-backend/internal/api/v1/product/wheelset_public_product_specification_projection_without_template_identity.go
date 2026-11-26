package product

import (
	"strconv"
	"strings"

	productdomain "commerce-platform/internal/domain/product"
)

func publicWheelsetSpecificationValues(specification *productdomain.WheelsetSpecification, locale string) []PublicProductSpecValue {
	if specification == nil {
		return nil
	}

	rows := make([]PublicProductSpecValue, 0, 36)
	appendRow := func(slug, english, chinese, value, unit string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		name, group := english, "Wheelset specifications"
		if locale == "zh_cn" {
			name, group = chinese, "轮组规格"
		}
		rows = append(rows, PublicProductSpecValue{
			Value: value,
			Definition: &PublicSpecDefinition{
				Slug: slug, Name: name, Group: group, Unit: unit,
				FieldType: "text", Role: "attribute", IsVisible: true,
			},
		})
	}
	appendMeasurement := func(slug, english, chinese string, value *float64, unit string) {
		if value == nil || *value <= 0 {
			return
		}
		appendRow(slug, english, chinese, strconv.FormatFloat(*value, 'f', -1, 64), unit)
	}

	appendWheelRows := func(position, englishPrefix, chinesePrefix string, wheel productdomain.WheelsetProductWheelSpecification, includeFreehub bool) {
		appendRow(position+"-wheel-size", englishPrefix+" wheel size", chinesePrefix+"轮轮径", wheel.WheelSize, "")
		if len(wheel.SpokeHoleCounts) > 0 {
			counts := make([]string, 0, len(wheel.SpokeHoleCounts))
			for _, count := range wheel.SpokeHoleCounts {
				if count > 0 {
					counts = append(counts, strconv.Itoa(count))
				}
			}
			appendRow(position+"-spoke-hole-count", englishPrefix+" spoke hole counts", chinesePrefix+"轮支持孔数", strings.Join(counts, ", "), "holes")
		}
		appendRow(position+"-lacing-pattern", englishPrefix+" lacing patterns", chinesePrefix+"轮支持编法", strings.Join(wheel.LacingPatterns, ", "), "")
		appendMeasurement(position+"-bead-seat-diameter", englishPrefix+" bead seat diameter (BSD)", chinesePrefix+"轮 BSD", wheel.BeadSeatDiameterMM, "mm")
		appendRow(position+"-material", englishPrefix+" material", chinesePrefix+"轮材质", wheel.Material, "")
		appendRow(position+"-brake-interface", englishPrefix+" brake interface", chinesePrefix+"轮制动接口", wheel.BrakeInterface, "")
		appendRow(position+"-rim-profile", englishPrefix+" rim profile", chinesePrefix+"轮胎圈结构", wheel.RimProfile, "")
		appendRow(position+"-tire-system", englishPrefix+" tire system", chinesePrefix+"轮轮胎系统", wheel.TireSystem, "")
		appendMeasurement(position+"-rim-depth", englishPrefix+" rim depth", chinesePrefix+"轮框高", wheel.RimDepthMM, "mm")
		appendMeasurement(position+"-rim-internal-width", englishPrefix+" rim internal width", chinesePrefix+"轮内宽", wheel.RimInternalWidthMM, "mm")
		appendMeasurement(position+"-rim-external-width", englishPrefix+" rim external width", chinesePrefix+"轮外宽", wheel.RimExternalWidthMM, "mm")
		appendRow(position+"-hub-interface", englishPrefix+" hub interface", chinesePrefix+"轮花鼓接口", wheel.HubInterface, "")
		appendRow(position+"-axle-standard", englishPrefix+" axle standard", chinesePrefix+"轮轴规格", wheel.AxleStandard, "")
		if includeFreehub {
			appendRow("rear-freehub-body", "Rear freehub body", "后轮塔基", wheel.FreehubBody, "")
		}
		appendRow(position+"-spoke-type", englishPrefix+" spoke type", chinesePrefix+"轮辐条类型", wheel.SpokeType, "")
		appendMeasurement(position+"-weight", englishPrefix+" wheel weight", chinesePrefix+"轮重量", wheel.WeightG, "g")
		appendMeasurement(position+"-maximum-tire-pressure", englishPrefix+" maximum tire pressure", chinesePrefix+"轮最大胎压", wheel.MaxTirePressureBar, "bar")
	}

	appendWheelRows("front", "Front", "前", specification.FrontWheel, false)
	appendWheelRows("rear", "Rear", "后", specification.RearWheel, true)
	appendMeasurement("net-weight", "Wheelset net weight", "整组净重", specification.NetWeightG, "g")
	appendMeasurement("maximum-system-weight", "Maximum system weight", "最大系统重量", specification.MaxSystemWeightKG, "kg")
	if len(specification.Certifications) > 0 {
		appendRow("certifications", "Certifications", "认证", strings.Join(specification.Certifications, ", "), "")
	}
	return rows
}

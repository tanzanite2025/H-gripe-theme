package product

import (
	productdomain "commerce-platform/internal/domain/product"
	"strconv"
	"strings"
)

// Project the dedicated facts into the existing read-only specification rows;
// these labels have no database definition IDs or template ownership.
func publicCarbonRimSpecificationValues(spec *productdomain.CarbonRimSpecification, locale string) []PublicProductSpecValue {
	if spec == nil {
		return nil
	}
	rows := make([]PublicProductSpecValue, 0, 14)
	appendCarbonRimSpecificationRow := func(slug, english, chinese, value, unit string) {
		name, group := english, "Carbon rim specifications"
		if locale == "zh_cn" {
			name, group = chinese, "碳纤维车圈规格"
		}
		rows = append(rows, PublicProductSpecValue{Value: value, Definition: &PublicSpecDefinition{Slug: slug, Name: name, Group: group, Unit: unit, FieldType: "text", Role: "attribute", IsVisible: true}})
	}
	formatCarbonRimMeasurement := func(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
	appendCarbonRimSpecificationRow("model", "Model", "型号", spec.Model, "")
	material := "Carbon fiber"
	if locale == "zh_cn" {
		material = "碳纤维"
	}
	if details := strings.TrimSpace(spec.MaterialDetails); details != "" {
		material += " · " + details
	}
	appendCarbonRimSpecificationRow("material", "Material", "材质", material, "")
	if label := productdomain.CarbonRimWheelSizeDisplayLabel(spec.WheelSize); label != "" {
		appendCarbonRimSpecificationRow("wheel_size", "Wheel size", "标称轮径", label, "")
	}
	if labels := productdomain.CarbonRimRidingUseDisplayLabels(spec.RidingUses, locale); len(labels) > 0 {
		appendCarbonRimSpecificationRow("riding_uses", "Intended riding disciplines", "适用骑行场景", strings.Join(labels, " / "), "")
	}
	appendCarbonRimSpecificationRow("sales_unit", "Sales unit", "销售单位", "1 rim", "")
	appendCarbonRimSpecificationRow("brake_type", "Brake use", "制动用途", "Disc brake", "")
	appendCarbonRimSpecificationRow("bead_seat_diameter_mm", "Bead seat diameter (BSD)", "胎圈座直径 BSD", formatCarbonRimMeasurement(spec.BeadSeatDiameterMM), "mm")
	appendCarbonRimSpecificationRow("rim_depth_mm", "Rim depth", "框高", formatCarbonRimMeasurement(spec.DepthMM), "mm")
	appendCarbonRimSpecificationRow("rim_internal_width_mm", "Internal width", "内宽", formatCarbonRimMeasurement(spec.InternalWidthMM), "mm")
	appendCarbonRimSpecificationRow("rim_external_width_mm", "External width", "外宽", formatCarbonRimMeasurement(spec.ExternalWidthMM), "mm")
	appendCarbonRimSpecificationRow("bead_structure", "Bead structure", "胎圈结构", spec.BeadStructure, "")
	if spec.TubelessReady != nil {
		appendCarbonRimSpecificationRow("tubeless_ready", "Tubeless ready", "支持真空胎", strconv.FormatBool(*spec.TubelessReady), "")
	}
	if spec.RimTapeRequired != nil {
		appendCarbonRimSpecificationRow("rim_tape_required", "Rim tape required", "需要胎垫", strconv.FormatBool(*spec.RimTapeRequired), "")
	}
	if spec.ValveHoleDiameterMM != nil {
		appendCarbonRimSpecificationRow("valve_hole_diameter_mm", "Valve hole diameter", "气嘴孔直径", formatCarbonRimMeasurement(*spec.ValveHoleDiameterMM), "mm")
	}
	appendCarbonRimSpecificationRow("asymmetric_offset_mm", "Asymmetric offset magnitude", "偏心量绝对值", formatCarbonRimMeasurement(spec.AsymmetricOffsetMM), "mm")
	if spec.TechnicalNotes != "" {
		appendCarbonRimSpecificationRow("technical_notes", "Manufacturer technical notes", "制造商技术说明", spec.TechnicalNotes, "")
	}
	return rows
}

import { carbonRimWheelSizeChoices, carbonRimRidingUseChoices } from './carbonRimWheelSizeAndRidingUseFormChoices'

export interface CarbonRimSpecificationForm {
  wheel_size: string
  material: string
  material_details: string
  riding_uses: string[]
  schema_version: number
  model: string
  bead_seat_diameter_mm: number | string
  depth_mm: number | string
  internal_width_mm: number | string
  external_width_mm: number | string
  bead_structure: string
  tubeless_ready: boolean | null
  rim_tape_required: boolean | null
  valve_hole_diameter_mm: number | string | null
  asymmetric_offset_mm: number | string
  technical_notes: string
}

export interface CarbonRimVariantSpecificationForm {
  spoke_hole_count: number | string
  drilling_pattern: string
  finish_code: string
  erd_mm: number | string | null
  erd_measurement_notes: string
  net_weight_grams: number | string
}

export const createEmptyCarbonRimSpecificationForm = (): CarbonRimSpecificationForm => ({
  wheel_size: '', material: 'carbon_fiber', material_details: '', riding_uses: [],
  schema_version: 1, model: '', bead_seat_diameter_mm: '', depth_mm: '',
  internal_width_mm: '', external_width_mm: '', bead_structure: '',
  tubeless_ready: null, rim_tape_required: null, valve_hole_diameter_mm: null,
  asymmetric_offset_mm: 0, technical_notes: '',
})

export const createEmptyCarbonRimVariantSpecificationForm = (): CarbonRimVariantSpecificationForm => ({
  spoke_hole_count: '', drilling_pattern: '', finish_code: '', erd_mm: null,
  erd_measurement_notes: '', net_weight_grams: '',
})

export const normalizeCarbonRimSpecificationForSubmission = (spec: CarbonRimSpecificationForm) => ({
  // These optional additions may be absent in previously saved version-1 documents.
  ...spec, model: spec.model.trim(), bead_seat_diameter_mm: Number(spec.bead_seat_diameter_mm),
  material_details: (spec.material_details || '').trim(), riding_uses: [...(spec.riding_uses || [])],
  depth_mm: Number(spec.depth_mm), internal_width_mm: Number(spec.internal_width_mm),
  external_width_mm: Number(spec.external_width_mm), asymmetric_offset_mm: Number(spec.asymmetric_offset_mm),
  valve_hole_diameter_mm: spec.valve_hole_diameter_mm === '' || spec.valve_hole_diameter_mm == null ? null : Number(spec.valve_hole_diameter_mm),
})

export const normalizeCarbonRimVariantSpecificationForSubmission = (spec: CarbonRimVariantSpecificationForm) => ({
  ...spec, spoke_hole_count: Number(spec.spoke_hole_count), finish_code: spec.finish_code.trim(),
  net_weight_grams: Number(spec.net_weight_grams),
  erd_mm: spec.erd_mm === '' || spec.erd_mm == null ? null : Number(spec.erd_mm),
})

export const validateCarbonRimDedicatedProductSubmission = (payload: any): Record<string, string> => {
  const errors: Record<string, string> = {}
  const spec = payload.carbon_rim_specification
  if (!spec?.model?.trim()) errors.carbon_rim_specification = '请填写车圈型号'
  else if (['bead_seat_diameter_mm', 'depth_mm', 'internal_width_mm', 'external_width_mm'].some(key => !Number.isFinite(spec[key]) || spec[key] <= 0)) errors.carbon_rim_specification = 'BSD、框高、内宽和外宽必须为正数'
  else if (spec.external_width_mm <= spec.internal_width_mm) errors.carbon_rim_specification = '外宽必须大于内宽'
  else if (spec.schema_version !== 1 || new TextEncoder().encode(spec.model).length > 120) errors.carbon_rim_specification = '规格版本必须为 1，型号长度不能超过 120 字节'
  else if (spec.depth_mm >= spec.bead_seat_diameter_mm / 2) errors.carbon_rim_specification = '框高必须小于胎圈座半径'
  else if (!Number.isFinite(spec.asymmetric_offset_mm) || spec.asymmetric_offset_mm < 0 || spec.asymmetric_offset_mm >= spec.external_width_mm / 2) errors.carbon_rim_specification = '偏心量必须为非负数，且小于外宽的一半'
  else if (spec.valve_hole_diameter_mm != null && (!Number.isFinite(spec.valve_hole_diameter_mm) || spec.valve_hole_diameter_mm <= 0)) errors.carbon_rim_specification = '气嘴孔直径填写时必须为正数'
  else if (new TextEncoder().encode(spec.technical_notes || '').length > 4000) errors.carbon_rim_specification = '技术说明不能超过 4000 字节'
  else if (!['hooked', 'hookless'].includes(spec.bead_structure) || typeof spec.tubeless_ready !== 'boolean') errors.carbon_rim_specification = '请选择胎圈结构，并明确是否支持真空胎'
  else if (!carbonRimWheelSizeChoices.some(choice => choice.value === spec.wheel_size && choice.beadSeatDiameterMM === spec.bead_seat_diameter_mm)) errors.carbon_rim_specification = '请选择标称轮径，并填写与之对应的 BSD（例如 700C 和 29 英寸均为 622 mm）'
  else if (spec.material !== 'carbon_fiber') errors.carbon_rim_specification = '碳纤维车圈的材质必须为碳纤维'
  else if (new TextEncoder().encode(spec.material_details || '').length > 500) errors.carbon_rim_specification = '材质牌号与铺层说明不能超过 500 字节'
  else if (!Array.isArray(spec.riding_uses) || !spec.riding_uses.length || spec.riding_uses.some((value: string) => !carbonRimRidingUseChoices.some(choice => choice.value === value)) || new Set(spec.riding_uses).size !== spec.riding_uses.length) errors.carbon_rim_specification = '请至少选择一种实际适用的骑行场景，不要重复选择'
  const configurations = new Set<string>()
  for (const [index, variant] of payload.variants.entries()) {
    const values = variant.carbon_rim_specification
    const prefix = `第 ${index + 1} 行 SKU：`
    if (!values || !Number.isInteger(values.spoke_hole_count) || values.spoke_hole_count < 1 || values.spoke_hole_count > 128) errors.variants = prefix + '请填写有效孔数'
    else if (!['uniform', 'paired', '2-to-1', 'g3'].includes(values.drilling_pattern)) errors.variants = prefix + '请选择实际钻孔形式'
    else if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(values.finish_code) || values.finish_code.length > 80) errors.variants = prefix + '涂装标识最多 80 个字符，使用小写字母、数字和短横线，例如 matte-black'
    else if (!Number.isFinite(values.net_weight_grams) || values.net_weight_grams <= 0 || !Number.isInteger(variant.weight_grams) || variant.weight_grams <= 0) errors.variants = prefix + '请填写净重和履约重量（克）'
    else if (values.erd_mm != null && (!Number.isFinite(values.erd_mm) || values.erd_mm <= 0 || !values.erd_measurement_notes.trim())) errors.variants = prefix + '填写 ERD 时需同时填写测量约定'
    else if (new TextEncoder().encode(values.erd_measurement_notes || '').length > 1000) errors.variants = prefix + 'ERD 测量约定不能超过 1000 字节'
    if (!Number.isSafeInteger(variant.price_minor) || variant.price_minor <= 0 || !Number.isSafeInteger(variant.stock) || variant.stock < 0) errors.variants = prefix + '价格必须有效，库存必须为非负整数'
    const key = JSON.stringify([values?.spoke_hole_count, values?.drilling_pattern, values?.finish_code])
    if (configurations.has(key)) errors.variants = prefix + '孔数、钻孔形式和涂装组合重复'
    configurations.add(key)
  }
  return errors
}

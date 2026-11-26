export interface WheelsetProductWheelSpecificationForm {
  wheel_size: string
  spoke_hole_counts: string
  lacing_patterns: string
  bead_seat_diameter_mm: number | string | null
  material: string
  brake_interface: string
  rim_profile: string
  tire_system: string
  rim_depth_mm: number | string | null
  rim_internal_width_mm: number | string | null
  rim_external_width_mm: number | string | null
  hub_interface: string
  axle_standard: string
  freehub_body: string
  spoke_type: string
  weight_g: number | string | null
  max_tire_pressure_bar: number | string | null
}

export interface WheelsetSpecificationForm {
  front_wheel: WheelsetProductWheelSpecificationForm
  rear_wheel: WheelsetProductWheelSpecificationForm
  net_weight_g: number | string | null
  max_system_weight_kg: number | string | null
  certifications: string
}

const createEmptyWheelSpecificationForm = (): WheelsetProductWheelSpecificationForm => ({
  wheel_size: '',
  spoke_hole_counts: '',
  lacing_patterns: '',
  bead_seat_diameter_mm: '',
  material: '',
  brake_interface: '',
  rim_profile: '',
  tire_system: '',
  rim_depth_mm: '',
  rim_internal_width_mm: '',
  rim_external_width_mm: '',
  hub_interface: '',
  axle_standard: '',
  freehub_body: '',
  spoke_type: '',
  weight_g: '',
  max_tire_pressure_bar: '',
})

export const createEmptyWheelsetSpecificationForm = (): WheelsetSpecificationForm => ({
  front_wheel: createEmptyWheelSpecificationForm(),
  rear_wheel: createEmptyWheelSpecificationForm(),
  net_weight_g: '',
  max_system_weight_kg: '',
  certifications: '',
})

const optionalText = (value: string | null | undefined): string | undefined => {
  const normalized = String(value || '').trim()
  return normalized || undefined
}

const optionalNumber = (value: number | string | null | undefined): number | undefined => {
  if (value == null || String(value).trim() === '') return undefined
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : Number.NaN
}

const splitList = (value: unknown): string[] => (Array.isArray(value) ? value.map(String) : String(value || '').split(/[,;\uFF0C\uFF1B]/))
  .map(item => item.trim())
  .filter(Boolean)

const normalizeWheelsetWheelSpecificationForSubmission = (specification: WheelsetProductWheelSpecificationForm) => {
  const spokeHoleCounts = splitList(specification.spoke_hole_counts).map(Number)
  const lacingPatterns = splitList(specification.lacing_patterns)
  return {
    wheel_size: String(specification.wheel_size || '').trim(),
    spoke_hole_counts: spokeHoleCounts,
    lacing_patterns: lacingPatterns,
    bead_seat_diameter_mm: optionalNumber(specification.bead_seat_diameter_mm),
    material: optionalText(specification.material),
    brake_interface: optionalText(specification.brake_interface),
    rim_profile: optionalText(specification.rim_profile),
    tire_system: optionalText(specification.tire_system),
    rim_depth_mm: optionalNumber(specification.rim_depth_mm),
    rim_internal_width_mm: optionalNumber(specification.rim_internal_width_mm),
    rim_external_width_mm: optionalNumber(specification.rim_external_width_mm),
    hub_interface: optionalText(specification.hub_interface),
    axle_standard: optionalText(specification.axle_standard),
    freehub_body: optionalText(specification.freehub_body),
    spoke_type: optionalText(specification.spoke_type),
    weight_g: optionalNumber(specification.weight_g),
    max_tire_pressure_bar: optionalNumber(specification.max_tire_pressure_bar),
  }
}

export const normalizeWheelsetSpecificationForSubmission = (specification: WheelsetSpecificationForm) => ({
  schema_version: 1,
  front_wheel: normalizeWheelsetWheelSpecificationForSubmission(specification.front_wheel),
  rear_wheel: normalizeWheelsetWheelSpecificationForSubmission(specification.rear_wheel),
  net_weight_g: optionalNumber(specification.net_weight_g),
  max_system_weight_kg: optionalNumber(specification.max_system_weight_kg),
  certifications: splitList(specification.certifications),
})

const formTextFromList = (value: unknown): string => Array.isArray(value) ? value.join(', ') : ''

const wheelSpecificationFormFromProduct = (value: Record<string, any> | undefined): WheelsetProductWheelSpecificationForm => ({
  ...createEmptyWheelSpecificationForm(),
  ...(value || {}),
  spoke_hole_counts: formTextFromList(value?.spoke_hole_counts),
  lacing_patterns: formTextFromList(value?.lacing_patterns),
})

export const createWheelsetSpecificationFormFromProduct = (value: Record<string, any> | null | undefined): WheelsetSpecificationForm => ({
  front_wheel: wheelSpecificationFormFromProduct(value?.front_wheel),
  rear_wheel: wheelSpecificationFormFromProduct(value?.rear_wheel),
  net_weight_g: value?.net_weight_g ?? '',
  max_system_weight_kg: value?.max_system_weight_kg ?? '',
  certifications: formTextFromList(value?.certifications),
})

const validateTextList = (value: string, fieldLabel: string, maximumBytes: number, errors: string[]): string[] => {
  const items = splitList(value)
  if (items.some(item => new TextEncoder().encode(item).length > maximumBytes)) errors.push(`${fieldLabel}单项不能超过 ${maximumBytes} 个字节`)
  if (new Set(items.map(item => item.toLowerCase())).size !== items.length) errors.push(`${fieldLabel}不能重复`)
  return items
}

const validateOptionalPositiveNumber = (value: number | string | null, fieldLabel: string, errors: string[]): void => {
  if (value == null || String(value).trim() === '') return
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) errors.push(`${fieldLabel}填写时必须是大于 0 的有限数值`)
}

const validateWheelsetWheel = (
  position: 'front_wheel' | 'rear_wheel',
  specification: WheelsetProductWheelSpecificationForm,
  errors: string[],
): void => {
  const positionLabel = position === 'front_wheel' ? '前轮' : '后轮'
  if (!String(specification.wheel_size || '').trim()) errors.push(`${positionLabel}必须填写轮径`)
  if (new TextEncoder().encode(String(specification.wheel_size || '').trim()).length > 32) errors.push(`${positionLabel}轮径不能超过 32 个字节`)

  const holeCounts = validateTextList(specification.spoke_hole_counts, `${positionLabel}孔数`, 8, errors)
  if (!holeCounts.length) errors.push(`${positionLabel}至少填写一个支持孔数`)
  for (const countText of holeCounts) {
    const count = Number(countText)
    if (!Number.isInteger(count) || count < 1 || count > 128) errors.push(`${positionLabel}孔数必须是 1 到 128 之间的整数`)
  }

  const lacingPatterns = validateTextList(specification.lacing_patterns, `${positionLabel}编法`, 64, errors)
  if (!lacingPatterns.length) errors.push(`${positionLabel}至少填写一个支持编法`)

  const optionalMeasurements: Array<[keyof WheelsetProductWheelSpecificationForm, string]> = [
    ['bead_seat_diameter_mm', `${positionLabel}BSD`],
    ['rim_depth_mm', `${positionLabel}框高`],
    ['rim_internal_width_mm', `${positionLabel}内宽`],
    ['rim_external_width_mm', `${positionLabel}外宽`],
    ['weight_g', `${positionLabel}重量`],
    ['max_tire_pressure_bar', `${positionLabel}最大胎压`],
  ]
  for (const [key, label] of optionalMeasurements) validateOptionalPositiveNumber(specification[key], label, errors)

  const internalWidth = optionalNumber(specification.rim_internal_width_mm)
  const externalWidth = optionalNumber(specification.rim_external_width_mm)
  if (internalWidth != null && externalWidth != null && Number.isFinite(internalWidth) && Number.isFinite(externalWidth) && externalWidth <= internalWidth) {
    errors.push(`${positionLabel}外宽必须大于内宽`)
  }

  const optionalTextLimits: Array<[keyof WheelsetProductWheelSpecificationForm, string, number]> = [
    ['material', `${positionLabel}材质`, 120],
    ['brake_interface', `${positionLabel}制动接口`, 80],
    ['rim_profile', `${positionLabel}胎圈结构`, 80],
    ['tire_system', `${positionLabel}轮胎系统`, 80],
    ['hub_interface', `${positionLabel}花鼓接口`, 120],
    ['axle_standard', `${positionLabel}轴规格`, 120],
    ['freehub_body', `${positionLabel}塔基`, 120],
    ['spoke_type', `${positionLabel}辐条类型`, 120],
  ]
  for (const [key, label, maximumBytes] of optionalTextLimits) {
    const value = String(specification[key] || '').trim()
    if (new TextEncoder().encode(value).length > maximumBytes) errors.push(`${label}不能超过 ${maximumBytes} 个字节`)
  }
}

export const validateWheelsetDedicatedProductSubmission = (payload: Record<string, any>): Record<string, string> => {
  const errors: string[] = []
  const specification = payload.wheelset_specification as WheelsetSpecificationForm | undefined
  if (!specification) return { wheelset_specification: '缺少 WHEELSET 产品级规格' }

  validateWheelsetWheel('front_wheel', specification.front_wheel || createEmptyWheelSpecificationForm(), errors)
  validateWheelsetWheel('rear_wheel', specification.rear_wheel || createEmptyWheelSpecificationForm(), errors)
  validateOptionalPositiveNumber(specification.net_weight_g, '整组净重', errors)
  validateOptionalPositiveNumber(specification.max_system_weight_kg, '最大系统重量', errors)
  validateTextList(specification.certifications, '认证', 160, errors)

  return errors.length ? { wheelset_specification: errors[0] } : {}
}

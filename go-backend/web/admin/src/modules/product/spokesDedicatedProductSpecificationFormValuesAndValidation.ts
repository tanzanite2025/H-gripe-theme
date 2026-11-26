export interface SpokeWeightReferenceForm {
  spoke_count: number | string | null
  length_mm: number | string | null
  total_weight_g: number | string | null
}

export interface SpokeGaugeOptionForm {
  gauge: string
  diameter_mm: number | string | null
  hub_end_diameter_mm: number | string | null
  middle_diameter_mm: number | string | null
  middle_width_mm: number | string | null
  middle_thickness_mm: number | string | null
  thread_end_diameter_mm: number | string | null
  length_min_mm: number | string | null
  length_max_mm: number | string | null
  weight_reference: SpokeWeightReferenceForm
}

export interface SpokesProductSpecificationForm {
  model_name: string
  manufacturer_part_number: string
  manufacturer_family: string
  material: string
  manufacturer_description: string
  gauge_options: SpokeGaugeOptionForm[]
  middle_strength_minimum_n_per_mm2: number | string | null
  middle_strength_maximum_n_per_mm2: number | string | null
  color_options: string
  manufacturer_options: string
  recommended_uses: string
  manufacturer_claims: string
  source_urls: string
}

const emptySpokeWeightReference = (): SpokeWeightReferenceForm => ({
  spoke_count: '',
  length_mm: '',
  total_weight_g: '',
})

export const createEmptySpokeGaugeOptionForm = (): SpokeGaugeOptionForm => ({
  gauge: '',
  diameter_mm: '',
  hub_end_diameter_mm: '',
  middle_diameter_mm: '',
  middle_width_mm: '',
  middle_thickness_mm: '',
  thread_end_diameter_mm: '',
  length_min_mm: '',
  length_max_mm: '',
  weight_reference: emptySpokeWeightReference(),
})

export const createEmptySpokesProductSpecificationForm = (): SpokesProductSpecificationForm => ({
  model_name: '',
  manufacturer_part_number: '',
  manufacturer_family: '',
  material: '',
  manufacturer_description: '',
  gauge_options: [createEmptySpokeGaugeOptionForm()],
  middle_strength_minimum_n_per_mm2: '',
  middle_strength_maximum_n_per_mm2: '',
  color_options: '',
  manufacturer_options: '',
  recommended_uses: '',
  manufacturer_claims: '',
  source_urls: '',
})

const optionalText = (value: unknown): string | undefined => {
  const normalized = String(value ?? '').trim()
  return normalized || undefined
}

const optionalNumber = (value: number | string | null | undefined): number | undefined => {
  if (value == null || String(value).trim() === '') return undefined
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : Number.NaN
}

const splitSpokeSpecificationList = (value: unknown): string[] => (
  (Array.isArray(value) ? value.map(String) : String(value || '').split(/[,;，；\n]/))
    .map(item => item.trim())
    .filter(Boolean)
)

const normalizeSpokeGaugeOption = (option: SpokeGaugeOptionForm) => {
  const weightFields = option.weight_reference
  const weightReference = [weightFields.spoke_count, weightFields.length_mm, weightFields.total_weight_g].some(value => value != null && String(value).trim() !== '')
    ? {
        spoke_count: optionalNumber(weightFields.spoke_count),
        length_mm: optionalNumber(weightFields.length_mm),
        total_weight_g: optionalNumber(weightFields.total_weight_g),
      }
    : undefined
  return {
    gauge: String(option.gauge || '').trim(),
    diameter_mm: optionalNumber(option.diameter_mm),
    hub_end_diameter_mm: optionalNumber(option.hub_end_diameter_mm),
    middle_diameter_mm: optionalNumber(option.middle_diameter_mm),
    middle_width_mm: optionalNumber(option.middle_width_mm),
    middle_thickness_mm: optionalNumber(option.middle_thickness_mm),
    thread_end_diameter_mm: optionalNumber(option.thread_end_diameter_mm),
    length_min_mm: optionalNumber(option.length_min_mm),
    length_max_mm: optionalNumber(option.length_max_mm),
    weight_reference: weightReference,
  }
}

const spokeGaugeOptionContainsValues = (option: SpokeGaugeOptionForm): boolean => (
  String(option.gauge || '').trim() !== ''
  || [
    option.diameter_mm,
    option.hub_end_diameter_mm,
    option.middle_diameter_mm,
    option.middle_width_mm,
    option.middle_thickness_mm,
    option.thread_end_diameter_mm,
    option.length_min_mm,
    option.length_max_mm,
    option.weight_reference.spoke_count,
    option.weight_reference.length_mm,
    option.weight_reference.total_weight_g,
  ].some(value => value != null && String(value).trim() !== '')
)

export const normalizeSpokesProductSpecificationForSubmission = (specification: SpokesProductSpecificationForm) => {
  const minimumStrength = optionalNumber(specification.middle_strength_minimum_n_per_mm2)
  const maximumStrength = optionalNumber(specification.middle_strength_maximum_n_per_mm2)
  const hasStrength = [specification.middle_strength_minimum_n_per_mm2, specification.middle_strength_maximum_n_per_mm2]
    .some(value => value != null && String(value).trim() !== '')
  return {
    schema_version: 1,
    model_name: String(specification.model_name || '').trim(),
    manufacturer_part_number: optionalText(specification.manufacturer_part_number),
    manufacturer_family: optionalText(specification.manufacturer_family),
    material: optionalText(specification.material),
    manufacturer_description: optionalText(specification.manufacturer_description),
    gauge_options: specification.gauge_options.filter(spokeGaugeOptionContainsValues).map(normalizeSpokeGaugeOption),
    middle_strength_n_per_mm2: hasStrength
      ? { minimum_n_per_mm2: minimumStrength, maximum_n_per_mm2: maximumStrength }
      : undefined,
    color_options: splitSpokeSpecificationList(specification.color_options),
    manufacturer_options: splitSpokeSpecificationList(specification.manufacturer_options),
    recommended_uses: splitSpokeSpecificationList(specification.recommended_uses),
    manufacturer_claims: splitSpokeSpecificationList(specification.manufacturer_claims),
    source_urls: splitSpokeSpecificationList(specification.source_urls),
  }
}

const byteLength = (value: unknown): number => new TextEncoder().encode(String(value || '')).length
const optionalMeasurements = [
  'diameter_mm',
  'hub_end_diameter_mm',
  'middle_diameter_mm',
  'middle_width_mm',
  'middle_thickness_mm',
  'thread_end_diameter_mm',
  'length_min_mm',
  'length_max_mm',
] as const

export const validateSpokesDedicatedProductSubmission = (payload: Record<string, any>): Record<string, string> => {
  const errors: Record<string, string> = {}
  const specification = payload.spoke_specification as ReturnType<typeof normalizeSpokesProductSpecificationForSubmission> | undefined
  if (!specification || specification.schema_version !== 1) return { spoke_specification: '辐条规格版本无效' }
  if (!specification.model_name || byteLength(specification.model_name) > 120) {
    errors.spoke_specification = '型号名称为必填项且不能超过 120 个字节'
  }
  for (const [field, maximum] of [
    ['manufacturer_part_number', 120],
    ['manufacturer_family', 120],
    ['material', 160],
    ['manufacturer_description', 4000],
  ] as const) {
    if (byteLength(specification[field]) > maximum) errors.spoke_specification = `${field} 不能超过 ${maximum} 个字节`
  }
  const middleStrength = specification.middle_strength_n_per_mm2
  if (middleStrength) {
    const minimum = Number(middleStrength.minimum_n_per_mm2)
    const maximum = Number(middleStrength.maximum_n_per_mm2)
    if (!Number.isFinite(minimum) || !Number.isFinite(maximum) || minimum <= 0 || maximum <= 0 || minimum > maximum) {
      errors.spoke_specification = 'Middle strength 必须填写有效的最小值和最大值（N/mm²）'
    }
  }
  const gaugeKeys = new Set<string>()
  for (const [index, option] of specification.gauge_options.entries()) {
    const gauge = String(option.gauge || '').trim()
    if (!gauge || byteLength(gauge) > 32) errors.spoke_specification = `第 ${index + 1} 个线规必须填写且不能超过 32 个字节`
    const gaugeKey = gauge.toLowerCase()
    if (gaugeKeys.has(gaugeKey)) errors.spoke_specification = '线规不能重复'
    gaugeKeys.add(gaugeKey)
    for (const field of optionalMeasurements) {
      const value = option[field]
      if (value != null && (!Number.isFinite(Number(value)) || Number(value) <= 0)) errors.spoke_specification = `${field} 填写时必须为正数`
    }
    if (option.length_min_mm != null && option.length_max_mm != null && Number(option.length_min_mm) > Number(option.length_max_mm)) {
      errors.spoke_specification = '辐条最短长度不能大于最长长度'
    }
    const weight = option.weight_reference
    if (weight) {
      const count = Number(weight.spoke_count)
      if (!Number.isInteger(count) || count <= 0 || !Number.isFinite(Number(weight.length_mm)) || Number(weight.length_mm) <= 0 || !Number.isFinite(Number(weight.total_weight_g)) || Number(weight.total_weight_g) <= 0) {
        errors.spoke_specification = '重量参考需同时填写正数支数、长度和总重'
      }
    }
  }
  for (const [field, values, maximum] of [
    ['color_options', specification.color_options, 120],
    ['manufacturer_options', specification.manufacturer_options, 240],
    ['recommended_uses', specification.recommended_uses, 120],
    ['manufacturer_claims', specification.manufacturer_claims, 500],
    ['source_urls', specification.source_urls, 1000],
  ] as const) {
    const seen = new Set<string>()
    for (const value of values) {
      if (!value || byteLength(value) > maximum) errors.spoke_specification = `${field} 单项不能为空且不能超过 ${maximum} 个字节`
      if (seen.has(value.toLowerCase())) errors.spoke_specification = `${field} 不能重复`
      seen.add(value.toLowerCase())
    }
  }
  for (const source of specification.source_urls) {
    try {
      const parsed = new URL(source)
      if (!['http:', 'https:'].includes(parsed.protocol)) errors.spoke_specification = '来源必须使用 HTTP 或 HTTPS URL'
    } catch {
      errors.spoke_specification = '来源必须是完整的 HTTP 或 HTTPS URL'
    }
  }
  return errors
}

export const createSpokesProductSpecificationFormFromProduct = (value: Record<string, any> | null | undefined): SpokesProductSpecificationForm => {
  const specification = value || {}
  const options = Array.isArray(specification.gauge_options) ? specification.gauge_options : []
  const middleStrength = specification.middle_strength_n_per_mm2 || {}
  const listText = (values: unknown): string => Array.isArray(values) ? values.join(', ') : ''
  return {
    ...createEmptySpokesProductSpecificationForm(),
    ...specification,
    gauge_options: options.length
      ? options.map((option: any) => ({ ...createEmptySpokeGaugeOptionForm(), ...option, weight_reference: { ...emptySpokeWeightReference(), ...(option.weight_reference || {}) } }))
      : [createEmptySpokeGaugeOptionForm()],
    middle_strength_minimum_n_per_mm2: middleStrength.minimum_n_per_mm2 ?? '',
    middle_strength_maximum_n_per_mm2: middleStrength.maximum_n_per_mm2 ?? '',
    color_options: listText(specification.color_options),
    manufacturer_options: listText(specification.manufacturer_options),
    recommended_uses: listText(specification.recommended_uses),
    manufacturer_claims: listText(specification.manufacturer_claims),
    source_urls: listText(specification.source_urls),
  }
}

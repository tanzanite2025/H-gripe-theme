import type { SchwalbeTireCatalogItem } from '@/api/schwalbeTireCatalog'

export interface BicycleTireSpecificationForm {
  schema_version: number
  model_name: string
  article_no: string
  etrto: string
  inch_designation: string
  weight_g: number | string | null
  version_label: string
  compound: string
  color: string
  bead: string
  e_bike_rating: string
  epi: number | string | null
  load_kg: number | string | null
  seal: string
  tread: string
  min_pressure_bar: number | string | null
  max_pressure_bar: number | string | null
  min_pressure_psi: number | string | null
  max_pressure_psi: number | string | null
}

export const createBicycleTireProductFormValuesFromSchwalbeCatalogItem = (item: SchwalbeTireCatalogItem) => ({
  bicycle_tire_specification: {
    schema_version: 1,
    model_name: item.model_name,
    article_no: item.article_no,
    etrto: item.etrto,
    inch_designation: item.inch_designation ?? '',
    weight_g: item.weight_g ?? '',
    version_label: item.version_label ?? '',
    compound: item.compound ?? '',
    color: item.color ?? '',
    bead: item.bead ?? '',
    e_bike_rating: item.e_bike_rating ?? '',
    epi: item.epi ?? '',
    load_kg: item.load_kg ?? '',
    seal: item.seal ?? '',
    tread: item.tread ?? '',
    min_pressure_bar: item.min_pressure_bar ?? '',
    max_pressure_bar: item.max_pressure_bar ?? '',
    min_pressure_psi: item.min_pressure_psi ?? '',
    max_pressure_psi: item.max_pressure_psi ?? '',
  } satisfies BicycleTireSpecificationForm,
  standard_product_identifier: item.ean ?? '',
})

export const createEmptyBicycleTireSpecificationForm = (): BicycleTireSpecificationForm => ({
  schema_version: 1,
  model_name: '',
  article_no: '',
  etrto: '',
  inch_designation: '',
  weight_g: '',
  version_label: '',
  compound: '',
  color: '',
  bead: '',
  e_bike_rating: '',
  epi: '',
  load_kg: '',
  seal: '',
  tread: '',
  min_pressure_bar: '',
  max_pressure_bar: '',
  min_pressure_psi: '',
  max_pressure_psi: '',
})

const optionalNumber = (value: number | string | null | undefined): number | undefined => {
  if (value == null || String(value).trim() === '') return undefined
  return Number(value)
}

const optionalText = (value: string | null | undefined): string | undefined => {
  const normalized = String(value || '').trim()
  return normalized || undefined
}

const utf8ByteLength = (value: string): number => new TextEncoder().encode(value).length

export const normalizeBicycleTireSpecificationForSubmission = (spec: BicycleTireSpecificationForm) => ({
  schema_version: 1,
  model_name: optionalText(spec.model_name),
  article_no: optionalText(spec.article_no),
  etrto: optionalText(spec.etrto),
  inch_designation: optionalText(spec.inch_designation),
  weight_g: optionalNumber(spec.weight_g),
  version_label: optionalText(spec.version_label),
  compound: optionalText(spec.compound),
  color: optionalText(spec.color),
  bead: optionalText(spec.bead),
  e_bike_rating: optionalText(spec.e_bike_rating),
  epi: optionalNumber(spec.epi),
  load_kg: optionalNumber(spec.load_kg),
  seal: optionalText(spec.seal),
  tread: optionalText(spec.tread),
  min_pressure_bar: optionalNumber(spec.min_pressure_bar),
  max_pressure_bar: optionalNumber(spec.max_pressure_bar),
  min_pressure_psi: optionalNumber(spec.min_pressure_psi),
  max_pressure_psi: optionalNumber(spec.max_pressure_psi),
})

export const validateBicycleTireDedicatedProductSubmission = (payload: any): Record<string, string> => {
  const errors: Record<string, string> = {}
  const spec = payload.bicycle_tire_specification
  if (!spec || spec.schema_version !== 1) {
    errors.bicycle_tire_specification = '外胎规格版本无效'
    return errors
  }

  const textLimits: Array<[string, number]> = [
    ['model_name', 180], ['article_no', 70], ['etrto', 32], ['inch_designation', 32],
    ['version_label', 120], ['compound', 120], ['color', 120], ['bead', 64],
    ['e_bike_rating', 64], ['seal', 120], ['tread', 120],
  ]
  for (const [key, limit] of textLimits) {
    if (utf8ByteLength(String(spec[key] || '')) > limit) errors.bicycle_tire_specification = `${key} 不能超过 ${limit} 个字节`
  }

  for (const key of ['weight_g', 'load_kg', 'min_pressure_bar', 'max_pressure_bar', 'min_pressure_psi', 'max_pressure_psi']) {
    const value = spec[key]
    if (value != null && (!Number.isFinite(value) || value <= 0)) errors.bicycle_tire_specification = `${key} 填写时必须为正数`
  }
  if (spec.epi != null && (!Number.isInteger(spec.epi) || spec.epi <= 0)) errors.bicycle_tire_specification = 'EPI 填写时必须为正整数'
  if (spec.min_pressure_bar != null && spec.max_pressure_bar != null && spec.min_pressure_bar > spec.max_pressure_bar) errors.bicycle_tire_specification = '最小 Bar 胎压不能大于最大 Bar 胎压'
  if (spec.min_pressure_psi != null && spec.max_pressure_psi != null && spec.min_pressure_psi > spec.max_pressure_psi) errors.bicycle_tire_specification = '最小 PSI 胎压不能大于最大 PSI 胎压'
  if (String(spec.article_no || '').trim() && !Number(payload.brand_id)) errors.bicycle_tire_specification = '填写 Article No. 时必须选择品牌'
  return errors
}

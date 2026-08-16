export interface CustomsClassificationRecord {
  id: number
  product_specification_template_id?: number | null
  name: string
  slug: string
  component_kind: string
  material: string
  hs_code: string
  cn_code: string
  country_of_origin: string
  customs_description: string
  source: string
  source_code: string
  source_url: string
  notes: string
  status: 'draft' | 'active' | 'paused'
  product_specification_template?: { id: number; name: string }
}

export type CustomsClassificationForm = Omit<CustomsClassificationRecord, 'id'> & { id?: number }

export interface LookupCandidate {
  provider: string
  source_code: string
  hs_code: string
  cn_code?: string
  description: string
  customs_description: string
  duty?: string
  source_url: string
}

export interface CustomsProductFilters {
  search: string
  product_specification_template_id: string
  customs_status: string
}

export interface CustomsFieldDefinition {
  key: 'hs_code' | 'cn_code' | 'country_of_origin' | 'customs_description'
  label: string
}

export const customsFieldDefinitions: CustomsFieldDefinition[] = [
  { key: 'hs_code', label: 'HS Code' },
  { key: 'cn_code', label: 'CN Code' },
  { key: 'country_of_origin', label: '原产国' },
  { key: 'customs_description', label: '英文品名' },
]

export const missingCustomsFields = (product: Record<string, any>): string[] => (
  customsFieldDefinitions
    .filter((field) => !String(product[field.key] || '').trim())
    .map((field) => field.label)
)

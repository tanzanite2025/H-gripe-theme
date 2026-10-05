export type CustomsTradeRemedyRiskLevel = 'none' | 'low' | 'medium' | 'high' | 'critical'

export type CustomsTradeRemedyRiskTag =
  | 'eu_anti_dumping_attention'
  | 'eu_complete_wheelset_anti_circumvention'
  | 'us_section_301_list_3_review'

export interface CustomsClassificationRecord {
  id: number
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
  source_url_us: string
  source_url_eu: string
  source_url_uk: string
  notes: string
  verified_at: string | null
  review_due_at: string | null
  trade_remedy_risk_level: CustomsTradeRemedyRiskLevel
  trade_remedy_risk_tags: CustomsTradeRemedyRiskTag[]
  trade_remedy_declaration_advice: string
  status: 'draft' | 'active' | 'paused'
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
  customs_status: string
}

export interface CustomsFieldDefinition {
  key: 'hs_code' | 'cn_code' | 'country_of_origin' | 'customs_description'
  label: string
}

export const customsTradeRemedyRiskLevelLabels: Record<CustomsTradeRemedyRiskLevel, string> = {
  none: '无特别标记',
  low: '低敏感',
  medium: '中敏感',
  high: '高敏感',
  critical: '极高敏感',
}

export const customsTradeRemedyRiskTagLabels: Record<CustomsTradeRemedyRiskTag, string> = {
  eu_anti_dumping_attention: '欧盟反倾销关注',
  eu_complete_wheelset_anti_circumvention: '欧盟整轮反规避关注',
  us_section_301_list_3_review: '美线 301 / List 3 复核',
}

export const customsTradeRemedyRiskTagOptions: Array<{ value: CustomsTradeRemedyRiskTag; label: string }> = Object.entries(customsTradeRemedyRiskTagLabels).map(([value, label]) => ({
  value: value as CustomsTradeRemedyRiskTag,
  label,
}))

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

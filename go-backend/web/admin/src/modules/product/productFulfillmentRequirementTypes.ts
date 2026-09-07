export type ProductFulfillmentRuleStatus = 'active' | 'inactive' | string

export interface ProductFulfillmentRequirementRule {
  id: number
  product_id: number
  variant_id?: number | null
  requirement_type: string
  spoke_tension_qc_required: boolean
  status: ProductFulfillmentRuleStatus
  rule_version: string
  reason: string
  created_by?: number | null
  created_at?: string | null
  updated_at?: string | null
}

export interface ProductFulfillmentRequirementListResult {
  product_id: number
  rules: ProductFulfillmentRequirementRule[]
}

export interface ProductFulfillmentRequirementUpsertInput {
  variant_id?: number | null
  spoke_tension_qc_required: boolean
  status: 'active' | 'inactive'
  rule_version: string
  reason: string
}

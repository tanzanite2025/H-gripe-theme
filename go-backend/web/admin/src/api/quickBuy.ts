import axios from '@/utils/axios'

const unwrapPayload = (response: any) => {
  const payload = response.data?.data ?? response.data ?? {}
  return payload?.data ?? payload
}

const unwrapList = (response: any) => {
  const payload = unwrapPayload(response)
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (Array.isArray(payload.flows)) return payload.flows
  return []
}

export type QuickBuyVersionStatus = 'draft' | 'published' | 'archived' | string
export type QuickBuySelectionMode = 'single' | 'multiple' | 'quantity' | 'auto' | string

export interface QuickBuyProductTypeRef {
  id: number
  slug: string
  name: string
  image_url?: string
  primary?: boolean
}

export interface QuickBuyStep {
  id?: number
  step_key: string
  slug?: string
  name: string
  description?: string
  help_text?: string
  sort_order: number
  selection_mode: QuickBuySelectionMode
  is_required: boolean
  min_select: number
  max_select: number
  default_quantity: number
  allow_skip: boolean
  product_types: QuickBuyProductTypeRef[]
}

export interface QuickBuyVersion {
  id?: number
  version_number?: number
  status?: QuickBuyVersionStatus
  published_at?: string
  starts_at?: string
  ends_at?: string
}

export interface QuickBuyFlowSummary {
  id: number
  slug: string
  name: string
  description?: string
  entry_surface: string
  is_enabled: boolean
  sort_order: number
  versions?: QuickBuyVersion[]
  created_at?: string
  updated_at?: string
}

export interface QuickBuyFlow extends QuickBuyFlowSummary {
  version: QuickBuyVersion
  steps: QuickBuyStep[]
}

export interface QuickBuyFlowPayload {
  slug: string
  name: string
  description?: string
  entry_surface: string
  is_enabled: boolean
  sort_order: number
  version: QuickBuyVersionPayload
}

export interface QuickBuyVersionPayload {
  starts_at?: string | null
  ends_at?: string | null
  steps: QuickBuyStepPayload[]
}

export interface QuickBuyStepPayload {
  step_key: string
  name: string
  description?: string
  help_text?: string
  sort_order: number
  selection_mode: QuickBuySelectionMode
  is_required: boolean
  min_select: number
  max_select: number
  default_quantity: number
  allow_skip: boolean
  product_type_ids: number[]
}

export interface QuickBuyValidationIssue {
  severity: 'error' | 'warning' | 'info' | string
  code: string
  message: string
  step_key?: string
  rule_key?: string
  product_type_id?: number
}

export interface QuickBuyValidationResult {
  valid: boolean
  issues: QuickBuyValidationIssue[]
}

export interface QuickBuyPreviewProduct {
  id: number
  name?: string
  title?: string
  slug?: string
  sku?: string
  price?: number
  sale_price?: number | null
  currency?: string
  display_price?: {
    amount?: number
    currency?: string
  } | null
  media?: Array<{
    url?: string
    thumbnail_url?: string
    is_primary?: boolean
    media_type?: string
  }>
  product_type?: {
    name?: string
    slug?: string
  } | null
  availability?: string
}

export interface QuickBuyPreviewResult {
  step?: QuickBuyStep
  products: QuickBuyPreviewProduct[]
  flow_id?: number
  flow_version_id?: number
  locale?: string
  currency?: string
  page: number
  page_size: number
  total: number
  has_more: boolean
}

export const quickBuyApi = {
  async listFlows(): Promise<QuickBuyFlowSummary[]> {
    return unwrapList(await axios.get('/api/admin/quick-buy/flows'))
  },

  async getFlow(id: number | string, params: Record<string, any> = {}): Promise<QuickBuyFlow> {
    return unwrapPayload(await axios.get(`/api/admin/quick-buy/flows/${id}`, { params }))
  },

  async createFlow(payload: QuickBuyFlowPayload): Promise<QuickBuyFlow> {
    return unwrapPayload(await axios.post('/api/admin/quick-buy/flows', payload))
  },

  async updateFlow(id: number | string, payload: QuickBuyFlowPayload): Promise<QuickBuyFlowSummary> {
    return unwrapPayload(await axios.put(`/api/admin/quick-buy/flows/${id}`, payload))
  },

  async createDraftVersion(flowId: number | string, payload: QuickBuyVersionPayload): Promise<QuickBuyFlow> {
    return unwrapPayload(await axios.post(`/api/admin/quick-buy/flows/${flowId}/draft`, payload))
  },

  async updateDraftVersion(versionId: number | string, payload: QuickBuyVersionPayload): Promise<QuickBuyFlow> {
    return unwrapPayload(await axios.put(`/api/admin/quick-buy/flow-versions/${versionId}`, payload))
  },

  async validateVersion(versionId: number | string): Promise<QuickBuyValidationResult> {
    return unwrapPayload(await axios.post(`/api/admin/quick-buy/flow-versions/${versionId}/validate`))
  },

  async previewVersion(versionId: number | string, payload: { step_key: string, keyword?: string, locale?: string, currency?: string, page?: number, page_size?: number }): Promise<QuickBuyPreviewResult> {
    return unwrapPayload(await axios.post(`/api/admin/quick-buy/flow-versions/${versionId}/preview`, payload))
  },

  async publishVersion(versionId: number | string): Promise<QuickBuyFlow> {
    return unwrapPayload(await axios.post(`/api/admin/quick-buy/flow-versions/${versionId}/publish`))
  },
}

export default quickBuyApi

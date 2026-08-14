import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type QuickBuyVersionStatus = 'draft' | 'published' | 'archived' | string
export interface QuickBuyProductTypeRef {
  id: number
  slug: string
  name: string
  image_url?: string
  primary?: boolean
}

export interface QuickBuyFlowTranslation {
  id?: number | string | null
  locale: string
  help_text?: string | null
}

export interface QuickBuyStep {
  id?: number
  step_key: string
  slug?: string
  name: string
  sort_order: number
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
  help_text?: string
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
  translations?: QuickBuyFlowTranslation[]
}

export interface QuickBuyFlowPayload {
  slug: string
  name: string
  description?: string
  help_text?: string
  translations?: QuickBuyFlowTranslation[]
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

const readDataEnvelope = (response: unknown, endpoint: string) => (
  requireApiObject(unwrapApiPayload(response, endpoint), endpoint, 'response payload')
)

const readFlowSummary = (response: unknown, endpoint: string): any => {
  const flow = requireApiObjectField(readDataEnvelope(response, endpoint), 'data', endpoint)
  requireApiNumberField(flow, 'id', endpoint)
  requireApiStringField(flow, 'slug', endpoint)
  requireApiStringField(flow, 'name', endpoint)
  requireApiStringField(flow, 'entry_surface', endpoint)
  requireApiBooleanField(flow, 'is_enabled', endpoint)
  requireApiNumberField(flow, 'sort_order', endpoint)
  if (flow.versions !== undefined) {
    if (!Array.isArray(flow.versions)) {
      requireApiArrayField(flow, 'versions', endpoint)
    }
  }
  return flow
}

const readFlow = (response: unknown, endpoint: string): any => {
  const flow = readFlowSummary(response, endpoint)
  requireApiObjectField(flow, 'version', endpoint)
  requireApiArrayField(flow, 'steps', endpoint)
  if (flow.translations !== undefined) {
    requireApiArrayField(flow, 'translations', endpoint)
  }
  return flow
}

const readValidationResult = (response: unknown, endpoint: string): any => {
  const result = requireApiObjectField(readDataEnvelope(response, endpoint), 'data', endpoint)
  requireApiBooleanField(result, 'valid', endpoint)
  requireApiArrayField(result, 'issues', endpoint)
  return result
}

const readPreviewResult = (response: unknown, endpoint: string): any => {
  const result = requireApiObjectField(readDataEnvelope(response, endpoint), 'data', endpoint)
  requireApiObjectField(result, 'step', endpoint)
  requireApiArrayField(result, 'products', endpoint)
  requireApiNumberField(result, 'page', endpoint)
  requireApiNumberField(result, 'page_size', endpoint)
  requireApiNumberField(result, 'total', endpoint)
  requireApiBooleanField(result, 'has_more', endpoint)
  return result
}

export const quickBuyApi = {
  async listFlows(): Promise<QuickBuyFlowSummary[]> {
    const endpoint = '/api/admin/quick-buy/flows'
    const payload = readDataEnvelope(await axios.get(endpoint), endpoint)
    return requireApiArrayField<QuickBuyFlowSummary>(payload, 'data', endpoint)
  },

  async getFlow(id: number | string, params: Record<string, any> = {}): Promise<QuickBuyFlow> {
    const endpoint = `/api/admin/quick-buy/flows/${id}`
    return readFlow(await axios.get(endpoint, { params }), endpoint)
  },

  async createFlow(payload: QuickBuyFlowPayload): Promise<QuickBuyFlow> {
    const endpoint = '/api/admin/quick-buy/flows'
    return readFlow(await axios.post(endpoint, payload), endpoint)
  },

  async updateFlow(id: number | string, payload: QuickBuyFlowPayload): Promise<QuickBuyFlowSummary> {
    const endpoint = `/api/admin/quick-buy/flows/${id}`
    return readFlowSummary(await axios.put(endpoint, payload), endpoint)
  },

  async saveFlowConfiguration(id: number | string, payload: QuickBuyFlowPayload): Promise<QuickBuyFlow> {
    const endpoint = `/api/admin/quick-buy/flows/${id}/configuration`
    return readFlow(await axios.put(endpoint, payload), endpoint)
  },

  async createDraftVersion(flowId: number | string, payload: QuickBuyVersionPayload): Promise<QuickBuyFlow> {
    const endpoint = `/api/admin/quick-buy/flows/${flowId}/draft`
    return readFlow(await axios.post(endpoint, payload), endpoint)
  },

  async updateDraftVersion(versionId: number | string, payload: QuickBuyVersionPayload): Promise<QuickBuyFlow> {
    const endpoint = `/api/admin/quick-buy/flow-versions/${versionId}`
    return readFlow(await axios.put(endpoint, payload), endpoint)
  },

  async validateVersion(versionId: number | string): Promise<QuickBuyValidationResult> {
    const endpoint = `/api/admin/quick-buy/flow-versions/${versionId}/validate`
    return readValidationResult(await axios.post(endpoint), endpoint)
  },

  async previewVersion(versionId: number | string, payload: { step_key: string, keyword?: string, locale?: string, currency?: string, page?: number, page_size?: number }): Promise<QuickBuyPreviewResult> {
    const endpoint = `/api/admin/quick-buy/flow-versions/${versionId}/preview`
    return readPreviewResult(await axios.post(endpoint, payload), endpoint)
  },

  async publishVersion(versionId: number | string): Promise<QuickBuyFlow> {
    const endpoint = `/api/admin/quick-buy/flow-versions/${versionId}/publish`
    return readFlow(await axios.post(endpoint), endpoint)
  },
}

export default quickBuyApi

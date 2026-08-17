import axios from '@/utils/axios'

export type StorefrontRedirectRuleState = 'draft' | 'published' | 'disabled'

export interface StorefrontRedirectRule {
  id: number
  source_path: string
  target_path: string
  status_code: 301 | 308
  state: StorefrontRedirectRuleState
  reason: string
  created_by_id: number
  published_by_id?: number | null
  published_at?: string | null
  disabled_at?: string | null
  created_at: string
  updated_at: string
}

export interface StorefrontRedirectRuleInput {
  source_path: string
  target_path: string
  status_code: 301 | 308
  reason: string
}

export const storefrontRedirectRulesApi = {
  async list(): Promise<StorefrontRedirectRule[]> {
    const response = await axios.get('/api/admin/urls/redirects')
    return Array.isArray(response.data?.items) ? response.data.items : []
  },

  async create(input: StorefrontRedirectRuleInput): Promise<StorefrontRedirectRule> {
    const response = await axios.post('/api/admin/urls/redirects', input)
    return response.data?.data || {}
  },

  async publish(id: number): Promise<StorefrontRedirectRule> {
    const response = await axios.post(`/api/admin/urls/redirects/${id}/publish`)
    return response.data?.data || {}
  },

  async disable(id: number): Promise<StorefrontRedirectRule> {
    const response = await axios.post(`/api/admin/urls/redirects/${id}/disable`)
    return response.data?.data || {}
  },
}

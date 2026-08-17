import axios from '@/utils/axios'
import type { SEOResourcePagination } from '@/modules/seo/types'
import type { StorefrontRouteCatalogEntry, StorefrontRouteCheckResult } from './routeCatalogTypes'

export type StorefrontURLIssueSeverity = 'low' | 'medium' | 'high' | 'critical'
export type StorefrontURLIssueState = 'open' | 'acknowledged' | 'resolved' | 'verified' | 'suppressed'
export type StorefrontURLIssueStateFilter = StorefrontURLIssueState | 'active' | 'all'

export interface StorefrontURLIssue {
  id: number
  route_entry_id: number
  issue_type: string
  severity: StorefrontURLIssueSeverity
  state: StorefrontURLIssueState
  assignee_id?: number | null
  resolution_type?: string | null
  resolution_note?: string | null
  linked_redirect_rule_id?: number | null
  latest_check_result_id?: number | null
  first_detected_at: string
  last_detected_at: string
  resolved_at?: string | null
  verified_at?: string | null
  suppressed_until?: string | null
  suppression_reason?: string | null
  created_at: string
  updated_at: string
  route_entry?: StorefrontRouteCatalogEntry | null
}

export interface StorefrontURLIssueEvent {
  id: number
  issue_id: number
  event_type: string
  actor_user_id: number
  note?: string | null
  metadata?: string | null
  created_at: string
}

export interface StorefrontURLIssueListResponse {
  items: StorefrontURLIssue[]
  pagination: SEOResourcePagination
}

export interface StorefrontURLIssueEventsResponse {
  items: StorefrontURLIssueEvent[]
  pagination: SEOResourcePagination
}

export interface StorefrontURLIssueStats {
  active: number
  open: number
  acknowledged: number
  resolved: number
  verified: number
  suppressed: number
  critical: number
  high: number
}

export interface StorefrontURLIssueResolutionInput {
  resolution_type: string
  resolution_note: string
  linked_redirect_rule_id?: number
}

export const storefrontURLIssuesApi = {
  async summary(): Promise<StorefrontURLIssueStats> {
    const response = await axios.get('/api/admin/urls/issues/summary')
    return response.data?.data || {}
  },

  async list(params: {
    page: number
    page_size: number
    state?: StorefrontURLIssueStateFilter
    severity?: StorefrontURLIssueSeverity
    issue_type?: string
  }): Promise<StorefrontURLIssueListResponse> {
    const response = await axios.get('/api/admin/urls/issues', { params })
    return {
      items: Array.isArray(response.data?.items) ? response.data.items : [],
      pagination: {
        page: params.page,
        page_size: params.page_size,
        total: Number(response.data?.pagination?.total || 0),
        total_pages: Number(response.data?.pagination?.total_pages || 0),
      },
    }
  },

  async get(id: number): Promise<StorefrontURLIssue> {
    const response = await axios.get(`/api/admin/urls/issues/${id}`)
    return response.data?.data || {}
  },

  async events(id: number, params: { page: number; page_size: number }): Promise<StorefrontURLIssueEventsResponse> {
    const response = await axios.get(`/api/admin/urls/issues/${id}/events`, { params })
    return {
      items: Array.isArray(response.data?.items) ? response.data.items : [],
      pagination: {
        page: params.page,
        page_size: params.page_size,
        total: Number(response.data?.pagination?.total || 0),
        total_pages: Number(response.data?.pagination?.total_pages || 0),
      },
    }
  },

  async acknowledge(id: number, note = ''): Promise<StorefrontURLIssue> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/acknowledge`, { note })
    return response.data?.data || {}
  },

  async claim(id: number): Promise<StorefrontURLIssue> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/claim`)
    return response.data?.data || {}
  },

  async comment(id: number, note: string): Promise<StorefrontURLIssue> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/comments`, { note })
    return response.data?.data || {}
  },

  async linkRedirect(id: number, redirectRuleID: number): Promise<StorefrontURLIssue> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/link-redirect`, {
      redirect_rule_id: redirectRuleID,
    })
    return response.data?.data || {}
  },

  async resolve(id: number, input: StorefrontURLIssueResolutionInput): Promise<StorefrontURLIssue> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/resolve`, input)
    return response.data?.data || {}
  },

  async suppress(id: number, input: { reason: string; suppressed_until: string }): Promise<StorefrontURLIssue> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/suppress`, input)
    return response.data?.data || {}
  },

  async recheck(id: number): Promise<{ issue: StorefrontURLIssue; check_result: StorefrontRouteCheckResult }> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/recheck`)
    return response.data?.data || {}
  },

  async verify(id: number): Promise<{ issue: StorefrontURLIssue; check_result?: StorefrontRouteCheckResult }> {
    const response = await axios.post(`/api/admin/urls/issues/${id}/verify`)
    return response.data?.data || {}
  },
}

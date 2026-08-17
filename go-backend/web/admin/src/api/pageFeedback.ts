import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiObject,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type PageFeedbackStatus = 'pending' | 'approved' | 'rejected' | 'hidden'
export type PageFeedbackStatusFilter = PageFeedbackStatus | 'all'
export type PageFeedbackRiskLevel = 'normal' | 'warning' | 'critical'

export interface PageFeedbackItem {
  id: number
  thread_key: string
  page_path: string
  page_title: string
  user_id: number
  name: string
  email: string
  source_hash_preview: string
  content: string
  status: PageFeedbackStatus
  locale: string
  reply_content: string
  replied_at?: string | null
  replied_by: number
  reviewed_at?: string | null
  reviewed_by: number
  created_at: string
  updated_at: string
}

export interface PageFeedbackPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface PageFeedbackListResult {
  data: PageFeedbackItem[]
  pagination: PageFeedbackPagination
}

export interface PageFeedbackRiskRateLimit {
  window_hours: number
  total: number
  read_ip: number
  write_ip: number
  write_user: number
  fallback_total: number
  fallback_read_ip: number
  fallback_write_ip: number
  fallback_write_user: number
  redis_unavailable: number
  unavailable: boolean
}

export interface PageFeedbackRiskTotals {
  pending_total: number
  pending_over_24h: number
  window_total: number
  last_hour_total: number
  by_status: Record<string, number>
}

export interface PageFeedbackRiskPage {
  page_path: string
  page_title: string
  thread_key: string
  filter_kind: 'page_path' | 'thread_key'
  filter_value: string
  feedback_count: number
  pending_count: number
  last_feedback_at: string
}

export interface PageFeedbackRiskSource {
  source_hash_preview: string
  feedback_count: number
  page_count: number
  pending_count: number
  last_feedback_at: string
}

export interface PageFeedbackRiskOverview {
  window_hours: number
  generated_at: string
  level: PageFeedbackRiskLevel
  totals: PageFeedbackRiskTotals
  rate_limit: PageFeedbackRiskRateLimit
  hot_pages: PageFeedbackRiskPage[]
  source_bursts: PageFeedbackRiskSource[]
}

const readPaged = (response: unknown, path: string): PageFeedbackListResult => {
  const responseBody = requireApiObject((response as { data?: unknown }).data, path, 'response body')
  const payload = unwrapApiPayload(response, path)
  return {
    data: requireApiArray<PageFeedbackItem>(payload, path, 'data'),
    pagination: requireApiPagination(responseBody, payload, path),
  }
}

const readObjectPayload = (response: unknown, path: string): PageFeedbackItem => (
  requireApiObject(unwrapApiPayload(response, path), path) as PageFeedbackItem
)

export const pageFeedbackApi = {
  async list(params: {
    status: PageFeedbackStatusFilter
    thread_key?: string
    page_path?: string
    search?: string
    page: number
    page_size: number
  }): Promise<PageFeedbackListResult> {
    const path = '/api/admin/content/feedback'
    return readPaged(await axios.get(path, { params }), path)
  },

  async get(id: number): Promise<PageFeedbackItem> {
    const path = `/api/admin/content/feedback/${id}`
    return readObjectPayload(await axios.get(path), path)
  },

  async update(
    id: number,
    payload: {
      status: PageFeedbackStatus
      reply_content: string
    },
  ): Promise<PageFeedbackItem> {
    const path = `/api/admin/content/feedback/${id}`
    return readObjectPayload(await axios.patch(path, payload), path)
  },

  async riskOverview(windowHours = 24): Promise<PageFeedbackRiskOverview> {
    const path = '/api/admin/content/feedback/risk-overview'
    return requireApiObject(
      unwrapApiPayload(
        await axios.get(path, { params: { window_hours: windowHours } }),
        path,
      ),
      path,
    ) as PageFeedbackRiskOverview
  },
}

export default pageFeedbackApi

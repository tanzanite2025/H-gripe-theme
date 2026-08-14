import axios from '@/utils/axios'
import {
  readApiBody,
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiObject,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type ShowcaseStatus = 'pending' | 'approved' | 'rejected'
export type ShowcaseStatusFilter = ShowcaseStatus | 'all'

export interface ShowcaseImageFile {
  index: number
  file_url: string
}

export interface ShowcaseRecord {
  id: number
  user_id: number
  kind: string
  title?: string
  order_id?: number | null
  region?: string
  location?: string
  nickname?: string
  notes?: string
  gallery_images: string[]
  image_files: ShowcaseImageFile[]
  image_count: number
  status: ShowcaseStatus
  rejected_reason?: string
  approved_at?: string | null
  created_at: string
  updated_at: string
}

export interface ShowcasePagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface ShowcaseListResult {
  items: ShowcaseRecord[]
  pagination: ShowcasePagination
}

export const showcaseApi = {
  async list(params: {
    status: ShowcaseStatusFilter
    page: number
    page_size: number
  }): Promise<ShowcaseListResult> {
    const endpoint = '/api/admin/showcase'
    const response = await axios.get(endpoint, {
      params: {
        type: 'user',
        status: params.status,
        page: params.page,
        page_size: params.page_size,
      },
    })
    const body = requireApiObject(readApiBody(response, endpoint), endpoint, 'response body')
    const payload = requireApiObject(unwrapApiPayload(response, endpoint), endpoint, 'response payload')
    const items = requireApiArrayField<ShowcaseRecord>(payload, 'items', endpoint)
    const pagination = requireApiPagination(body, payload, endpoint)

    return {
      items,
      pagination,
    }
  },

  async approve(id: number): Promise<void> {
    const endpoint = `/api/admin/showcase/${id}/approve`
    requireApiAcknowledgement(await axios.put(endpoint), endpoint)
  },

  async reject(id: number, reason: string): Promise<void> {
    const endpoint = `/api/admin/showcase/${id}/reject`
    requireApiAcknowledgement(await axios.put(endpoint, { reason }), endpoint)
  },
}

export default showcaseApi

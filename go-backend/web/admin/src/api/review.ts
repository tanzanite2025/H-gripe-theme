import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiObject,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type ReviewStatus = 'pending' | 'approved' | 'rejected'
export type ReviewStatusFilter = ReviewStatus | 'all'

export interface AdminReviewProduct {
  id: number
  name: string
  sku: string
}

export interface AdminReviewUser {
  id: number
  username: string
  email: string
  display_name: string
}

export interface AdminReview {
  id: number
  product?: AdminReviewProduct | null
  user?: AdminReviewUser | null
  order_id: number
  rating: number
  title: string
  content: string
  images: string[]
  pros: string
  cons: string
  status: ReviewStatus
  featured: boolean
  verified: boolean
  helpful_count: number
  reply_content: string
  replied_at?: string | null
  replied_by: number
  moderated_at?: string | null
  moderated_by?: number | null
  moderation_reason: string
  created_at: string
  updated_at: string
}

export interface AdminReviewPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface AdminReviewListResult {
  data: AdminReview[]
  pagination: AdminReviewPagination
}

const readPaged = (response: unknown, path: string): AdminReviewListResult => {
  const responseBody = requireApiObject((response as { data?: unknown }).data, path, 'response body')
  const payload = unwrapApiPayload(response, path)
  return {
    data: requireApiArray<AdminReview>(payload, path, 'data'),
    pagination: requireApiPagination(responseBody, payload, path),
  }
}

const readObjectPayload = (response: unknown, path: string): AdminReview => (
  requireApiObject(unwrapApiPayload(response, path), path) as AdminReview
)

export const reviewApi = {
  async list(params: {
    status: ReviewStatusFilter
    search?: string
    product_id?: number
    page: number
    page_size: number
  }): Promise<AdminReviewListResult> {
    const path = '/api/admin/reviews'
    return readPaged(await axios.get(path, { params }), path)
  },

  async get(id: number): Promise<AdminReview> {
    const path = `/api/admin/reviews/${id}`
    return readObjectPayload(await axios.get(path), path)
  },

  async updateStatus(id: number, status: ReviewStatus, reason = ''): Promise<AdminReview> {
    const path = `/api/admin/reviews/${id}/status`
    return readObjectPayload(await axios.patch(path, { status, reason }), path)
  },
}

export default reviewApi

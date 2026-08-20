import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiObject,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import { adminApiUrl } from '@/lib/adminUrl'

export type AfterSalesCaseID = number | string

export interface AfterSalesCaseItem {
  id?: AfterSalesCaseID
  order_item_id?: AfterSalesCaseID
  product_name?: string | null
  sku?: string | null
  quantity?: number | null
}

export interface AfterSalesCaseEvent {
  id?: AfterSalesCaseID
  case_id?: AfterSalesCaseID
  from_status?: string | null
  to_status?: string | null
  resolution?: string | null
  updated_by?: AfterSalesCaseID | null
  operator_name?: string | null
  created_at?: string | null
}

export interface AfterSalesCaseAttachment {
  id: AfterSalesCaseID
  case_id?: AfterSalesCaseID
  kind?: 'image' | 'video' | string | null
  filename?: string | null
  content_type?: string | null
  size_bytes?: number | null
  created_at?: string | null
  updated_at?: string | null
}

export interface AfterSalesRefundReview {
  id?: AfterSalesCaseID
  case_id?: AfterSalesCaseID
  status?: 'pending' | 'approved' | 'rejected' | 'cancelled' | string | null
  proposed_amount?: number | null
  currency?: string | null
  request_notes?: string | null
  decision_notes?: string | null
  created_by?: AfterSalesCaseID | null
  updated_by?: AfterSalesCaseID | null
  reviewed_by_id?: AfterSalesCaseID | null
  creator_name?: string | null
  reviewer_name?: string | null
  reviewed_at?: string | null
  linked_refund_id?: AfterSalesCaseID | null
  created_at?: string | null
  updated_at?: string | null
}

export interface AfterSalesCase {
  id: AfterSalesCaseID
  order_id?: AfterSalesCaseID | null
  order_number?: string | null
  type?: string | null
  status?: string | null
  reason?: string | null
  description?: string | null
  resolution?: string | null
  items?: AfterSalesCaseItem[]
  events?: AfterSalesCaseEvent[]
  attachments?: AfterSalesCaseAttachment[]
  refund_review?: AfterSalesRefundReview | null
  refund_review_maximum_amount?: number | null
  refund_review_currency?: string | null
  created_at?: string | null
  updated_at?: string | null
}

export interface SaveAfterSalesRefundReviewInput {
  proposed_amount: number
  currency: string
  request_notes: string
}

export interface AfterSalesPendingRefundResult {
  refund_review: AfterSalesRefundReview
  refund: {
    id: AfterSalesCaseID
    status?: string | null
    amount?: number | null
    requested_amount?: number | null
  }
}

export interface CreateAfterSalesCaseItemInput {
  order_item_id: AfterSalesCaseID
  quantity: number
}

export interface CreateAfterSalesCaseInput {
  type: 'return_refund' | 'exchange' | 'refund_only' | 'reshipment'
  reason: string
  description?: string
  items: CreateAfterSalesCaseItemInput[]
}

export interface AfterSalesListParams {
  page: number
  page_size: number
  status?: string
  type?: string
  search?: string
}

export interface AfterSalesListResult {
  data: AfterSalesCase[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export const afterSalesAttachmentUrl = (caseID: AfterSalesCaseID, attachmentID: AfterSalesCaseID): string => {
  return adminApiUrl(`/api/admin/after-sales/${encodeURIComponent(String(caseID))}/attachments/${encodeURIComponent(String(attachmentID))}`)
}

const readObjectPayload = <T = Record<string, unknown>>(response: unknown, path: string): T => (
  requireApiObject(unwrapApiPayload(response, path), path) as T
)

export const afterSalesApi = {
  async create(
    orderID: AfterSalesCaseID,
    input: CreateAfterSalesCaseInput,
  ): Promise<AfterSalesCase> {
    const path = `/api/admin/orders/${orderID}/after-sales`
    return readObjectPayload<AfterSalesCase>(await axios.post(path, input), path)
  },

  async list(params: AfterSalesListParams): Promise<AfterSalesListResult> {
    const path = '/api/admin/after-sales'
    const response = await axios.get(path, { params })
    const responseBody = requireApiObject((response as { data?: unknown }).data, path, 'response body')
    const payload = unwrapApiPayload(response, path)
    return {
      data: requireApiArray<AfterSalesCase>(payload, path, 'data'),
      pagination: requireApiPagination(responseBody, payload, path),
    }
  },

  async get(caseID: AfterSalesCaseID): Promise<AfterSalesCase> {
    const path = `/api/admin/after-sales/${caseID}`
    return readObjectPayload<AfterSalesCase>(await axios.get(path), path)
  },

  async updateStatus(
    caseID: AfterSalesCaseID,
    status: string,
    resolution = '',
  ): Promise<AfterSalesCase> {
    const path = `/api/admin/after-sales/${caseID}/status`
    return readObjectPayload<AfterSalesCase>(await axios.patch(path, { status, resolution }), path)
  },

  async getRefundReview(caseID: AfterSalesCaseID): Promise<AfterSalesRefundReview> {
    const path = `/api/admin/after-sales/${caseID}/refund-review`
    return readObjectPayload<AfterSalesRefundReview>(await axios.get(path), path)
  },

  async saveRefundReview(
    caseID: AfterSalesCaseID,
    input: SaveAfterSalesRefundReviewInput,
  ): Promise<AfterSalesRefundReview> {
    const path = `/api/admin/after-sales/${caseID}/refund-review`
    return readObjectPayload<AfterSalesRefundReview>(await axios.put(path, input), path)
  },

  async decideRefundReview(
    caseID: AfterSalesCaseID,
    status: 'approved' | 'rejected' | 'cancelled',
    decisionNotes: string,
  ): Promise<AfterSalesRefundReview> {
    const path = `/api/admin/after-sales/${caseID}/refund-review/decision`
    return readObjectPayload<AfterSalesRefundReview>(
      await axios.patch(path, { status, decision_notes: decisionNotes }),
      path,
    )
  },

  async createPendingRefund(caseID: AfterSalesCaseID): Promise<AfterSalesPendingRefundResult> {
    const path = `/api/admin/after-sales/${caseID}/refund-review/pending-refund`
    const result = readObjectPayload<AfterSalesPendingRefundResult>(await axios.post(path, { confirm: true }), path)
    if (!result.refund_review || !result.refund) {
      throw new Error(`${path}: missing refund review or refund payload`)
    }
    return result
  },
}

export default afterSalesApi

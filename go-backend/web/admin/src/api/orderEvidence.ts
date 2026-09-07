import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiObject,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type {
  OrderEvidenceAttachment,
  OrderEvidenceItemUpdateInput,
  OrderEvidenceListParams,
  OrderEvidenceListResult,
  OrderEvidencePackageResult,
  OrderEvidenceOrderSummary,
} from '@/modules/order/orderEvidenceTypes'

const readObjectPayload = <T = Record<string, unknown>>(response: unknown, path: string): T => (
  requireApiObject(unwrapApiPayload(response, path), path) as T
)

const readPaged = (response: unknown, path: string): OrderEvidenceListResult => {
  const responseBody = requireApiObject((response as { data?: unknown }).data, path, 'response body')
  const payload = unwrapApiPayload(response, path)
  return {
    data: requireApiArray<OrderEvidenceOrderSummary>(payload, path, 'data'),
    pagination: requireApiPagination(responseBody, payload, path),
  }
}

const orderPath = (orderID: number | string): string => (
  `/api/admin/orders/${encodeURIComponent(String(orderID))}/evidence`
)

export const orderEvidenceAttachmentUrl = (
  orderID: number | string,
  itemID: number | string,
  attachmentID: number | string,
): string => (
  `${orderPath(orderID)}/items/${encodeURIComponent(String(itemID))}/attachments/${encodeURIComponent(String(attachmentID))}`
)

export const orderEvidenceApi = {
  async listOrders(params: OrderEvidenceListParams): Promise<OrderEvidenceListResult> {
    const path = '/api/admin/order-evidence'
    return readPaged(await axios.get(path, { params }), path)
  },

  async getPackage(orderID: number | string): Promise<OrderEvidencePackageResult> {
    const path = orderPath(orderID)
    return readObjectPayload<OrderEvidencePackageResult>(await axios.get(path), path)
  },

  async exportSnapshot(orderID: number | string): Promise<Blob> {
    const path = `${orderPath(orderID)}/export`
    const response = await axios.get<Blob>(path, { responseType: 'blob' })
    return response.data
  },

  async updateItem(
    orderID: number | string,
    itemID: number | string,
    input: OrderEvidenceItemUpdateInput,
  ): Promise<OrderEvidencePackageResult> {
    const path = `${orderPath(orderID)}/items/${encodeURIComponent(String(itemID))}`
    return readObjectPayload<OrderEvidencePackageResult>(await axios.patch(path, input), path)
  },

  async lockPackage(orderID: number | string): Promise<OrderEvidencePackageResult> {
    const path = `${orderPath(orderID)}/lock`
    return readObjectPayload<OrderEvidencePackageResult>(await axios.post(path), path)
  },

  async createRevision(orderID: number | string): Promise<OrderEvidencePackageResult> {
    const path = `${orderPath(orderID)}/revisions`
    return readObjectPayload<OrderEvidencePackageResult>(await axios.post(path), path)
  },

  async uploadAttachment(
    orderID: number | string,
    itemID: number | string,
    file: File,
  ): Promise<OrderEvidenceAttachment> {
    const path = `${orderPath(orderID)}/items/${encodeURIComponent(String(itemID))}/attachments`
    const formData = new FormData()
    formData.append('file', file)
    return readObjectPayload<OrderEvidenceAttachment>(
      await axios.post(path, formData),
      path,
    )
  },

  async deleteAttachment(
    orderID: number | string,
    itemID: number | string,
    attachmentID: number | string,
  ): Promise<void> {
    const path = `${orderPath(orderID)}/items/${encodeURIComponent(String(itemID))}/attachments/${encodeURIComponent(String(attachmentID))}`
    await axios.delete(path)
  },
}

export default orderEvidenceApi

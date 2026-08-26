import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArray,
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type APIID = string | number
type APIParams = Record<string, any>
type APIPayload = Record<string, any>

export interface Pagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface PagedPayload<T = any> {
  data: T[]
  pagination: Pagination
}

export interface WarrantyStats {
  total_count?: number
  active_count?: number
  expired_count?: number
  unbound_count?: number
  [key: string]: unknown
}

export interface WarrantyOrderItem {
  id?: APIID | null
  [key: string]: unknown
}

export interface WarrantyServiceRecord {
  id?: APIID | null
  [key: string]: unknown
}

export interface WarrantyClaim {
  id?: APIID | null
  status?: string | null
  processed_at?: string | null
  resolution?: string | null
  order_item_id?: APIID | null
  order_item?: WarrantyOrderItem | null
  order_number?: string | null
  service_records?: WarrantyServiceRecord[]
  [key: string]: unknown
}

export interface ShipmentItemSnapshot {
  id?: APIID | null
  product_id?: APIID | null
  product_name?: string | null
  sku?: string | null
  variant_id?: APIID | null
  quantity?: number | string | null
  price?: number | string | null
  subtotal?: number | string | null
  total?: number | string | null
  attributes?: string | null
  [key: string]: unknown
}

export interface ShipmentRecord {
  id?: APIID | null
  order_id?: APIID | null
  order_number?: string | null
  user_id?: APIID | null
  customer_name?: string | null
  customer_email?: string | null
  tracking_shipment_id?: APIID | null
  tracking_number?: string | null
  shipped_at?: string | null
  items_snapshot?: ShipmentItemSnapshot[]
  product_codes?: string[]
  shipping_note?: string | null
  shipping_images?: string[]
  warranty_months?: number | null
  warranty_start_at?: string | null
  warranty_expires?: string | null
  status?: string | null
  record_bound?: boolean
  order_status?: string | null
  shipping_status?: string | null
  created_at?: string | null
  updated_at?: string | null
  [key: string]: unknown
}

const readObjectPayload = (response: unknown, endpoint: string): any => (
  requireApiObject(unwrapApiPayload(response, endpoint), endpoint)
)

const readPagedPayload = <T = any>(response: unknown, endpoint: string): PagedPayload<T> => {
  const responseBody = requireApiObject(
    (response as { data?: unknown })?.data,
    endpoint,
    'response body',
  )
  const payload = requireApiArray<T>(
    unwrapApiPayload(response, endpoint),
    endpoint,
    'data',
  )
  return {
    data: payload,
    pagination: requireApiPagination(responseBody, payload, endpoint),
  }
}

const readNamedArray = <T = any>(response: unknown, field: string, endpoint: string): T[] => {
  const payload = readObjectPayload(response, endpoint)
  return requireApiArrayField<T>(payload, field, endpoint)
}

const readClaim = (response: unknown, endpoint: string): any => {
  const claim = readObjectPayload(response, endpoint)
  requireApiNumberField(claim, 'id', endpoint)
  requireApiStringField(claim, 'status', endpoint)
  if (claim.service_records !== undefined) {
    requireApiArrayField(claim, 'service_records', endpoint)
  }
  return claim
}

const readServiceRecord = (response: unknown, endpoint: string): any => {
  const record = readObjectPayload(response, endpoint)
  requireApiNumberField(record, 'id', endpoint)
  requireApiNumberField(record, 'claim_id', endpoint)
  requireApiStringField(record, 'status', endpoint)
  requireApiStringField(record, 'summary', endpoint)
  return record
}

const readShipmentRecord = (response: unknown, endpoint: string): ShipmentRecord => {
  const record = readObjectPayload(response, endpoint) as ShipmentRecord
  requireApiNumberField(record, 'id', endpoint)
  requireApiStringField(record, 'order_number', endpoint)
  if (record.shipping_images !== undefined) requireApiArrayField(record, 'shipping_images', endpoint)
  if (record.items_snapshot !== undefined) requireApiArrayField(record, 'items_snapshot', endpoint)
  if (record.product_codes !== undefined) requireApiArrayField(record, 'product_codes', endpoint)
  return record
}

export const warrantyApi = {
  async getShipmentStats(): Promise<WarrantyStats> {
    const endpoint = '/api/admin/warranty/shipment-records/stats'
    const stats = readObjectPayload(await axios.get(endpoint), endpoint)
    if (stats.total_count !== undefined) requireApiNumberField(stats, 'total_count', endpoint)
    if (stats.active_count !== undefined) requireApiNumberField(stats, 'active_count', endpoint)
    if (stats.expired_count !== undefined) requireApiNumberField(stats, 'expired_count', endpoint)
    return stats
  },

  async listWarrantyClaims(params: APIParams = {}): Promise<PagedPayload<WarrantyClaim>> {
    const endpoint = '/api/admin/warranty/claims'
    return readPagedPayload<WarrantyClaim>(await axios.get(endpoint, { params }), endpoint)
  },

  async getWarrantyClaim(id: APIID): Promise<WarrantyClaim> {
    const endpoint = `/api/admin/warranty/claims/${id}`
    return readClaim(await axios.get(endpoint), endpoint)
  },

  async updateWarrantyClaimStatus(id: APIID, status: string): Promise<any> {
    const endpoint = `/api/admin/warranty/claims/${id}/status`
    return requireApiAcknowledgement(await axios.put(endpoint, { status }), endpoint)
  },

  async updateWarrantyClaimResolution(id: APIID, resolution: string): Promise<any> {
    const endpoint = `/api/admin/warranty/claims/${id}/resolution`
    return requireApiAcknowledgement(await axios.put(endpoint, { resolution }), endpoint)
  },

  async listWarrantyClaimOrderItems(id: APIID): Promise<WarrantyOrderItem[]> {
    const endpoint = `/api/admin/warranty/claims/${id}/order-items`
    return readNamedArray<WarrantyOrderItem>(await axios.get(endpoint), 'items', endpoint)
  },

  async bindWarrantyClaimOrderItem(id: APIID, orderItemId: APIID | null | undefined): Promise<any> {
    const endpoint = `/api/admin/warranty/claims/${id}/order-item`
    return requireApiAcknowledgement(await axios.put(endpoint, {
      order_item_id: orderItemId || null,
    }), endpoint)
  },

  async listWarrantyServiceRecords(id: APIID): Promise<WarrantyServiceRecord[]> {
    const endpoint = `/api/admin/warranty/claims/${id}/service-records`
    return readNamedArray<WarrantyServiceRecord>(await axios.get(endpoint), 'records', endpoint)
  },

  async createWarrantyServiceRecord(id: APIID, payload: APIPayload): Promise<WarrantyServiceRecord> {
    const endpoint = `/api/admin/warranty/claims/${id}/service-records`
    return readServiceRecord(await axios.post(endpoint, payload), endpoint)
  },

  async listShipmentRecords(params: APIParams = {}): Promise<PagedPayload<ShipmentRecord>> {
    const endpoint = '/api/admin/warranty/shipment-records'
    return readPagedPayload<ShipmentRecord>(await axios.get(endpoint, { params }), endpoint)
  },

  async getShipmentRecord(orderID: APIID): Promise<ShipmentRecord> {
    const endpoint = `/api/admin/warranty/shipment-records/${orderID}`
    return readShipmentRecord(await axios.get(endpoint), endpoint)
  },

  async updateShipmentRecord(orderID: APIID, payload: {
    shipping_note: string
    shipping_images: string[]
    product_codes: string[]
    warranty_months: number
    warranty_start_at: string
  }): Promise<ShipmentRecord> {
    const endpoint = `/api/admin/warranty/shipment-records/${orderID}`
    return readShipmentRecord(await axios.put(endpoint, payload), endpoint)
  },

  async uploadShipmentImages(orderID: APIID, files: File[]): Promise<ShipmentRecord> {
    const endpoint = `/api/admin/warranty/shipment-records/${orderID}/images`
    const formData = new FormData()
    files.forEach((file) => formData.append('images[]', file))
    const response = await axios.post(endpoint, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const payload = readObjectPayload(response, endpoint)
    return readShipmentRecord({ data: payload.record }, endpoint)
  },
}

export default warrantyApi

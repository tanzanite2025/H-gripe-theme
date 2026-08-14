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
  [key: string]: unknown
}

export interface WarrantyRegistration {
  id?: APIID | null
  status?: string | null
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

const readRegistration = (response: unknown, endpoint: string): any => {
  const registration = readObjectPayload(response, endpoint)
  requireApiNumberField(registration, 'id', endpoint)
  requireApiStringField(registration, 'status', endpoint)
  return registration
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

export const registrationApi = {
  async getStats(): Promise<WarrantyStats> {
    const endpoint = '/api/admin/registrations/stats'
    const stats = readObjectPayload(await axios.get(endpoint), endpoint)
    if (stats.total_count !== undefined) requireApiNumberField(stats, 'total_count', endpoint)
    if (stats.active_count !== undefined) requireApiNumberField(stats, 'active_count', endpoint)
    if (stats.expired_count !== undefined) requireApiNumberField(stats, 'expired_count', endpoint)
    return stats
  },

  async listRegistrations(params: APIParams = {}): Promise<PagedPayload<WarrantyRegistration>> {
    const endpoint = '/api/admin/registrations'
    return readPagedPayload<WarrantyRegistration>(await axios.get(endpoint, { params }), endpoint)
  },

  async updateRegistrationStatus(id: APIID, status: string): Promise<any> {
    const endpoint = `/api/admin/registrations/${id}/status`
    return requireApiAcknowledgement(await axios.put(endpoint, { status }), endpoint)
  },

  async listExpiringWarranties(limit = 30): Promise<WarrantyRegistration[]> {
    const endpoint = '/api/admin/registrations/expiring'
    const payload = readObjectPayload(await axios.get(endpoint, { params: { limit } }), endpoint)
    return requireApiArrayField<WarrantyRegistration>(payload, 'data', endpoint)
  },

  async listWarrantyClaims(params: APIParams = {}): Promise<PagedPayload<WarrantyClaim>> {
    const endpoint = '/api/admin/registrations/warranty-claims'
    return readPagedPayload<WarrantyClaim>(await axios.get(endpoint, { params }), endpoint)
  },

  async getWarrantyClaim(id: APIID): Promise<WarrantyClaim> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}`
    return readClaim(await axios.get(endpoint), endpoint)
  },

  async updateWarrantyClaimStatus(id: APIID, status: string): Promise<any> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}/status`
    return requireApiAcknowledgement(await axios.put(endpoint, { status }), endpoint)
  },

  async updateWarrantyClaimResolution(id: APIID, resolution: string): Promise<any> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}/resolution`
    return requireApiAcknowledgement(await axios.put(endpoint, { resolution }), endpoint)
  },

  async listWarrantyClaimOrderItems(id: APIID): Promise<WarrantyOrderItem[]> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}/order-items`
    return readNamedArray<WarrantyOrderItem>(await axios.get(endpoint), 'items', endpoint)
  },

  async bindWarrantyClaimOrderItem(id: APIID, orderItemId: APIID | null | undefined): Promise<any> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}/order-item`
    return requireApiAcknowledgement(await axios.put(endpoint, {
      order_item_id: orderItemId || null,
    }), endpoint)
  },

  async listWarrantyServiceRecords(id: APIID): Promise<WarrantyServiceRecord[]> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}/service-records`
    return readNamedArray<WarrantyServiceRecord>(await axios.get(endpoint), 'records', endpoint)
  },

  async createWarrantyServiceRecord(id: APIID, payload: APIPayload): Promise<WarrantyServiceRecord> {
    const endpoint = `/api/admin/registrations/warranty-claims/${id}/service-records`
    return readServiceRecord(await axios.post(endpoint, payload), endpoint)
  },
}

export default registrationApi

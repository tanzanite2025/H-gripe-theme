import axios from '@/utils/axios'

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

const defaultPagination = (): Pagination => ({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 0,
})

const unwrapPayload = (response: any): any => response.data?.data ?? response.data ?? {}

const unwrapPaged = <T = any>(response: any): PagedPayload<T> => ({
  data: Array.isArray(response.data?.data) ? response.data.data : [],
  pagination: response.data?.pagination ?? defaultPagination(),
})

const unwrapList = <T = any>(response: any, key?: string): T[] => {
  const payload = unwrapPayload(response)
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (key && Array.isArray(payload[key])) return payload[key]
  return []
}

export const registrationApi = {
  async getStats(): Promise<WarrantyStats> {
    const response = await axios.get('/api/admin/registrations/stats')
    return unwrapPayload(response)
  },

  async listRegistrations(params: APIParams = {}): Promise<PagedPayload<WarrantyRegistration>> {
    const response = await axios.get('/api/admin/registrations', { params })
    return unwrapPaged<WarrantyRegistration>(response)
  },

  async updateRegistrationStatus(id: APIID, status: string): Promise<any> {
    const response = await axios.put(`/api/admin/registrations/${id}/status`, { status })
    return unwrapPayload(response)
  },

  async listExpiringWarranties(limit = 30): Promise<WarrantyRegistration[]> {
    const response = await axios.get('/api/admin/registrations/expiring', { params: { limit } })
    return unwrapList<WarrantyRegistration>(response, 'data')
  },

  async listWarrantyClaims(params: APIParams = {}): Promise<PagedPayload<WarrantyClaim>> {
    const response = await axios.get('/api/admin/registrations/warranty-claims', { params })
    return unwrapPaged<WarrantyClaim>(response)
  },

  async getWarrantyClaim(id: APIID): Promise<WarrantyClaim> {
    const response = await axios.get(`/api/admin/registrations/warranty-claims/${id}`)
    return unwrapPayload(response)
  },

  async updateWarrantyClaimStatus(id: APIID, status: string): Promise<any> {
    const response = await axios.put(`/api/admin/registrations/warranty-claims/${id}/status`, { status })
    return unwrapPayload(response)
  },

  async updateWarrantyClaimResolution(id: APIID, resolution: string): Promise<any> {
    const response = await axios.put(`/api/admin/registrations/warranty-claims/${id}/resolution`, { resolution })
    return unwrapPayload(response)
  },

  async listWarrantyClaimOrderItems(id: APIID): Promise<WarrantyOrderItem[]> {
    const response = await axios.get(`/api/admin/registrations/warranty-claims/${id}/order-items`)
    return unwrapList<WarrantyOrderItem>(response, 'items')
  },

  async bindWarrantyClaimOrderItem(id: APIID, orderItemId: APIID | null | undefined): Promise<any> {
    const response = await axios.put(`/api/admin/registrations/warranty-claims/${id}/order-item`, {
      order_item_id: orderItemId || null,
    })
    return unwrapPayload(response)
  },

  async listWarrantyServiceRecords(id: APIID): Promise<WarrantyServiceRecord[]> {
    const response = await axios.get(`/api/admin/registrations/warranty-claims/${id}/service-records`)
    return unwrapList<WarrantyServiceRecord>(response, 'records')
  },

  async createWarrantyServiceRecord(id: APIID, payload: APIPayload): Promise<WarrantyServiceRecord> {
    const response = await axios.post(`/api/admin/registrations/warranty-claims/${id}/service-records`, payload)
    return unwrapPayload(response)
  },
}

export default registrationApi

import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  requireApiStringField,
  readApiBody,
} from '@/utils/apiResponse'

export interface ProcurementRecord {
  id: number
  product_code: string
  product_name: string
  purchase_price: number
  currency: string
  supplier_name: string
  supplier_contact_name: string
  supplier_phone: string
  supplier_email: string
  lead_time_days: number
  minimum_order_quantity: number
  inbound_shipping_unit_cost: number
  packaging_unit_cost: number
  other_unit_cost: number
  created_at?: string
  updated_at?: string
}

export interface ProcurementRecordDetailsPayload {
  purchase_price: number | null
  currency: string
  supplier_name: string
  supplier_contact_name: string
  supplier_phone: string
  supplier_email: string
  lead_time_days: number
  minimum_order_quantity: number
  inbound_shipping_unit_cost: number
  packaging_unit_cost: number
  other_unit_cost: number
}

export interface ProcurementRecordCreatePayload extends ProcurementRecordDetailsPayload {
  sku: string
}

export type ProcurementRecordUpdatePayload = ProcurementRecordDetailsPayload

export interface ProcurementProductOption {
  product_name: string
  variant_title: string
  sku: string
  available: boolean
}

export interface ProcurementPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface ProcurementListPayload {
  records: ProcurementRecord[]
  pagination: ProcurementPagination
}

const readRecord = (value: unknown, endpoint: string): ProcurementRecord => {
  const record = requireApiObject(value, endpoint, 'record') as ProcurementRecord
  requireApiNumberField(record, 'id', endpoint)
  requireApiStringField(record, 'product_code', endpoint)
  requireApiStringField(record, 'product_name', endpoint)
  requireApiNumberField(record, 'purchase_price', endpoint)
  requireApiStringField(record, 'currency', endpoint)
  requireApiStringField(record, 'supplier_name', endpoint)
  requireApiStringField(record, 'supplier_contact_name', endpoint)
  requireApiStringField(record, 'supplier_phone', endpoint)
  requireApiStringField(record, 'supplier_email', endpoint)
  requireApiNumberField(record, 'lead_time_days', endpoint)
  requireApiNumberField(record, 'minimum_order_quantity', endpoint)
  requireApiNumberField(record, 'inbound_shipping_unit_cost', endpoint)
  requireApiNumberField(record, 'packaging_unit_cost', endpoint)
  requireApiNumberField(record, 'other_unit_cost', endpoint)
  return record
}

const readProductOption = (value: unknown, endpoint: string): ProcurementProductOption => {
  const option = requireApiObject(value, endpoint, 'product option') as ProcurementProductOption
  requireApiStringField(option, 'product_name', endpoint)
  requireApiStringField(option, 'variant_title', endpoint)
  requireApiStringField(option, 'sku', endpoint)
  requireApiBooleanField(option, 'available', endpoint)
  return option
}

export const procurementApi = {
  async list(params: Record<string, unknown> = {}): Promise<ProcurementListPayload> {
    const endpoint = '/api/admin/procurement/records'
    const body = requireApiObject(readApiBody(await axios.get(endpoint, { params }), endpoint), endpoint, 'response body')
    const records = requireApiArrayField<ProcurementRecord>(body, 'records', endpoint).map((record) => readRecord(record, endpoint))
    const pagination = requireApiPagination(body, body, endpoint)
    return { records, pagination }
  },

  async listByCodes(codes: string[]): Promise<ProcurementRecord[]> {
    const endpoint = '/api/admin/procurement/records/by-codes'
    const normalizedCodes = codes.map((code) => code.trim()).filter(Boolean)
    if (!normalizedCodes.length) return []
    const body = requireApiObject(readApiBody(await axios.get(endpoint, {
      params: { codes: normalizedCodes.join(',') },
    }), endpoint), endpoint, 'response body')
    return requireApiArrayField<ProcurementRecord>(body, 'records', endpoint)
      .map((record) => readRecord(record, endpoint))
  },

  async listProductOptions(params: Record<string, unknown> = {}): Promise<{
    options: ProcurementProductOption[]
    pagination: ProcurementPagination
  }> {
    const endpoint = '/api/admin/procurement/product-options'
    const body = requireApiObject(readApiBody(await axios.get(endpoint, { params }), endpoint), endpoint, 'response body')
    const options = requireApiArrayField(body, 'options', endpoint)
      .map((option) => readProductOption(option, endpoint))
    const pagination = requireApiPagination(body, body, endpoint)
    return { options, pagination }
  },

  async get(id: number | string): Promise<ProcurementRecord> {
    const endpoint = `/api/admin/procurement/records/${id}`
    const body = requireApiObject(readApiBody(await axios.get(endpoint), endpoint), endpoint, 'response body')
    return readRecord(requireApiObjectField(body, 'record', endpoint), endpoint)
  },

  async create(payload: ProcurementRecordCreatePayload): Promise<ProcurementRecord> {
    const endpoint = '/api/admin/procurement/records'
    const body = requireApiObject(readApiBody(await axios.post(endpoint, payload), endpoint), endpoint, 'response body')
    return readRecord(requireApiObjectField(body, 'record', endpoint), endpoint)
  },

  async update(id: number | string, payload: ProcurementRecordUpdatePayload): Promise<ProcurementRecord> {
    const endpoint = `/api/admin/procurement/records/${id}`
    const body = requireApiObject(readApiBody(await axios.put(endpoint, payload), endpoint), endpoint, 'response body')
    return readRecord(requireApiObjectField(body, 'record', endpoint), endpoint)
  },

  async remove(id: number | string) {
    const endpoint = `/api/admin/procurement/records/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default procurementApi

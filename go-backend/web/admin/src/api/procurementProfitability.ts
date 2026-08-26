import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  readApiBody,
} from '@/utils/apiResponse'
import { procurementApi, type ProcurementRecord } from '@/api/procurement'

export interface ProfitabilityRecord {
  id: number
  product_code: string
  product_name: string
  currency: string
  list_price: number
  sale_price?: number
  effective_selling_price: number
  purchase_price: number
  inbound_shipping_unit_cost: number
  packaging_unit_cost: number
  other_unit_cost: number
  landed_cost: number
  gross_profit: number
  gross_margin_bps: number
  calculation_status: string
  formula_version: string
  warnings: string[]
  calculated_at?: string
  created_at?: string
  updated_at?: string
}

export interface ProfitabilityProcurementPayload {
  supplier_name: string
  supplier_contact_name: string
  supplier_phone: string
  supplier_email: string
  lead_time_days: number
  minimum_order_quantity: number
}

export interface ProfitabilityItemPayload {
  product_code: string
  product_name: string
  currency: string
  cost_currency?: string
  list_price: number
  sale_price?: number | null
  purchase_price?: number | null
  purchase_price_known: boolean
  inbound_shipping_unit_cost: number
  packaging_unit_cost: number
  other_unit_cost: number
  procurement?: ProfitabilityProcurementPayload
}

export interface ProfitabilityPreviewResult {
  product_code: string
  product_name: string
  currency: string
  cost_currency: string
  status: string
  formula_version: string
  warnings: string[]
  list_price: number
  sale_price?: number
  effective_selling_price: number
  purchase_price?: number
  inbound_shipping_unit_cost: number
  packaging_unit_cost: number
  other_unit_cost: number
  landed_cost?: number
  gross_profit?: number
  gross_margin_bps?: number
  gross_margin_percent?: number
}

export interface ProfitabilitySkippedItem {
  product_code: string
  status: string
  reason: string
}

export interface ProfitabilityBulkUpsertResult {
  records: ProfitabilityRecord[]
  skipped: ProfitabilitySkippedItem[]
}

const readWarnings = (value: unknown): string[] => (
  Array.isArray(value) ? value.map((item) => String(item)) : []
)

const readProfitabilityRecord = (value: unknown, endpoint: string): ProfitabilityRecord => {
  const record = requireApiObject(value, endpoint, 'record') as ProfitabilityRecord
  requireApiNumberField(record, 'id', endpoint)
  requireApiStringField(record, 'product_code', endpoint)
  requireApiStringField(record, 'product_name', endpoint)
  requireApiStringField(record, 'currency', endpoint)
  requireApiNumberField(record, 'list_price', endpoint)
  requireApiNumberField(record, 'effective_selling_price', endpoint)
  requireApiNumberField(record, 'purchase_price', endpoint)
  requireApiNumberField(record, 'inbound_shipping_unit_cost', endpoint)
  requireApiNumberField(record, 'packaging_unit_cost', endpoint)
  requireApiNumberField(record, 'other_unit_cost', endpoint)
  requireApiNumberField(record, 'landed_cost', endpoint)
  requireApiNumberField(record, 'gross_profit', endpoint)
  requireApiNumberField(record, 'gross_margin_bps', endpoint)
  requireApiStringField(record, 'calculation_status', endpoint)
  requireApiStringField(record, 'formula_version', endpoint)
  return {
    ...record,
    warnings: readWarnings(record.warnings),
  }
}

const readPreviewResult = (value: unknown, endpoint: string): ProfitabilityPreviewResult => {
  const result = requireApiObject(value, endpoint, 'item') as ProfitabilityPreviewResult
  requireApiStringField(result, 'product_code', endpoint)
  requireApiStringField(result, 'product_name', endpoint)
  requireApiStringField(result, 'currency', endpoint)
  requireApiStringField(result, 'cost_currency', endpoint)
  requireApiStringField(result, 'status', endpoint)
  requireApiStringField(result, 'formula_version', endpoint)
  requireApiNumberField(result, 'list_price', endpoint)
  requireApiNumberField(result, 'effective_selling_price', endpoint)
  requireApiNumberField(result, 'inbound_shipping_unit_cost', endpoint)
  requireApiNumberField(result, 'packaging_unit_cost', endpoint)
  requireApiNumberField(result, 'other_unit_cost', endpoint)
  return {
    ...result,
    warnings: readWarnings(result.warnings),
  }
}

const readSkippedItem = (value: unknown, endpoint: string): ProfitabilitySkippedItem => {
  const item = requireApiObject(value, endpoint, 'skipped item') as ProfitabilitySkippedItem
  requireApiStringField(item, 'product_code', endpoint)
  requireApiStringField(item, 'status', endpoint)
  requireApiStringField(item, 'reason', endpoint)
  return item
}

const readCodesBody = async (codes: string[], endpoint: string) => {
  if (!codes.length) return { records: [] as ProfitabilityRecord[] }
  const body = requireApiObject(readApiBody(await axios.get(endpoint, {
    params: { codes: codes.join(',') },
  }), endpoint), endpoint, 'response body')
  return {
    records: requireApiArrayField(body, 'records', endpoint)
      .map((record) => readProfitabilityRecord(record, endpoint)),
  }
}

export const procurementProfitabilityApi = {
  async listProcurementByCodes(codes: string[]): Promise<ProcurementRecord[]> {
    return procurementApi.listByCodes(codes)
  },

  async listProfitabilityByCodes(codes: string[]): Promise<ProfitabilityRecord[]> {
    const normalizedCodes = codes.map((code) => code.trim()).filter(Boolean)
    const endpoint = '/api/admin/procurement/profitability/by-codes'
    return (await readCodesBody(normalizedCodes, endpoint)).records
  },

  async preview(items: ProfitabilityItemPayload[]): Promise<ProfitabilityPreviewResult[]> {
    const endpoint = '/api/admin/procurement/profitability/preview'
    const body = requireApiObject(readApiBody(await axios.post(endpoint, { items }), endpoint), endpoint, 'response body')
    return requireApiArrayField(body, 'items', endpoint)
      .map((item) => readPreviewResult(item, endpoint))
  },

  async bulkUpsert(
    items: ProfitabilityItemPayload[],
    requestId = '',
  ): Promise<ProfitabilityBulkUpsertResult> {
    const endpoint = '/api/admin/procurement/profitability/bulk-upsert'
    const body = requireApiObject(readApiBody(await axios.post(endpoint, {
      request_id: requestId,
      items,
    }), endpoint), endpoint, 'response body')
    return {
      records: requireApiArrayField(body, 'records', endpoint)
        .map((record) => readProfitabilityRecord(record, endpoint)),
      skipped: requireApiArrayField(body, 'skipped', endpoint)
        .map((item) => readSkippedItem(item, endpoint)),
    }
  },
}

export default procurementProfitabilityApi

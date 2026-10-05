import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  unwrapApiPayload,
} from '@/utils/apiResponse'

type APIID = string | number
type APIParams = Record<string, any>
type APIPayload = Record<string, any>

export interface YanwenAPIConfigView {
  environment: 'fat' | 'production'
  endpoint: string
  user_id_configured: boolean
  api_token_configured: boolean
  enabled: boolean
}

export interface YanwenPingResult {
  ok: boolean
  message: string
  latency_ms: number
}

export interface YanwenOperationsOverviewDestinationFlow {
  country_code: string
  waybill_count: number
  channel_name: string
}

export interface YanwenOperationsOverview {
  generated_at: string
  environment: 'fat' | 'production'
  today_created_waybill_count: number
  official_pending_print_waybill_count: number
  active_tracking_snapshot_count: number
  customs_exception_tracking_snapshot_count: number
  destination_flows: YanwenOperationsOverviewDestinationFlow[]
  gateway: {
    environment: 'fat' | 'production'
    endpoint: string
    credentials_configured: boolean
    enabled: boolean
  }
}

export interface YanwenProductCatalogEntry {
  id: number
  environment: 'fat' | 'production'
  product_id: string
  name_ch: string
  name_en: string
  last_synced_at: string
}

export interface YanwenProductSyncSummary {
  environment: 'fat' | 'production'
  scanned: number
  added: number
  updated: number
  synced_at: string
}

export interface YanwenCountryCatalogEntry {
  id: number
  environment: 'fat' | 'production'
  country_id: string
  country_code: string
  name_ch: string
  name_en: string
  last_synced_at: string
}

export interface YanwenCountrySyncSummary {
  environment: 'fat' | 'production'
  scanned: number
  added: number
  updated: number
  synced_at: string
}

export interface YanwenWarehouseCatalogEntry {
  id: number
  environment: 'fat' | 'production'
  warehouse_code: string
  name: string
  area: string
  last_synced_at: string
}

export interface YanwenWarehouseSyncSummary {
  environment: 'fat' | 'production'
  scanned: number
  added: number
  updated: number
  synced_at: string
}

export interface YanwenKoreaPersonalCustomsClearanceCodeVerificationResult {
  environment: 'fat' | 'production'
  official_passed: boolean
  code: string
  message: string
}

export interface YanwenUnitedStatesAddressVerificationResult {
  environment: 'fat' | 'production'
  official_passed: boolean
  code: string
  message: string
  normalized_address: string
  normalized_city: string
  normalized_state: string
  normalized_zip_code4: string
  normalized_zip_code5: string
}

export interface YanwenPublishedChannel {
  id: number
  environment: 'fat' | 'production'
  product_code: string
  display_name: string
  countries: string
  package_type: string
  max_weight_grams: number
  volumetric_divisor: number
  require_receiver_tax_number: boolean
  require_ioss: boolean
  require_eori: boolean
  notes: string
  enabled: boolean
}

export interface YanwenWaybill {
  id: number
  environment: 'fat' | 'production'
  order_id: number
  order_number: string
  product_code: string
  channel_name: string
  warehouse_code: string
  destination_country: string
  consignee_name: string
  declared_description: string
  total_quantity: number
  total_weight_grams: number
  waybill_number: string
  yanwen_order_number: string
  reference_number: string
  status: string
  official_status: number
  is_printed: boolean
  last_official_synced_at?: string
  created_at: string
  updated_at: string
}

export interface YanwenWaybillLabel {
  waybill_id: number
  waybill_number: string
  file_name: string
  content_type: string
  base64_string: string
}

export interface YanwenBatchWaybillLabelDownload {
  blob: Blob
  succeeded: number
  failed: number
}

export interface YanwenWaybillCreationRequest {
  environment: 'fat' | 'production'
  order_id: number
  channel_id: number
  warehouse_code: string
  order_source?: string
  has_battery: boolean
  receiver_tax_number?: string
  ioss?: string
  eori?: string
}

export interface YanwenBatchWaybillCreationItem {
  request_index: number
  environment: 'fat' | 'production'
  order_id: number
  channel_id: number
  waybill_id?: number
  order_number?: string
  waybill_number?: string
  error?: string
}

export interface YanwenBatchWaybillCreationResult {
  items: YanwenBatchWaybillCreationItem[]
  succeeded: number
  failed: number
}

export interface YanwenWaybillCancellationResult {
  waybill: YanwenWaybill
  cancellation_accepted: boolean
  official_status_synced: boolean
  message: string
}

export interface YanwenBatchWaybillCancellationItem {
  waybill_id: number
  waybill_number: string
  cancellation_accepted: boolean
  official_status_synced: boolean
  message: string
  error?: string
}

export interface YanwenBatchWaybillCancellationResult {
  items: YanwenBatchWaybillCancellationItem[]
  succeeded: number
  failed: number
}

export interface YanwenTrackingCheckpoint {
  time_stamp: string
  time_zone: string
  tracking_status: string
  message: string
  location: string
  is_last_mile_checkpoint: boolean
  extra_properties?: Record<string, unknown> | null
}

export interface YanwenTrackingStatusWaybill {
  level1: string
  level2: string
  level3: string
}

export interface YanwenTrackingResult {
  tracking_number: string
  waybill_number: string
  exchange_number: string
  last_mile_carrier: string
  last_mile_carrier_website: string
  last_mile_carrier_contact_number: string
  checkpoints: YanwenTrackingCheckpoint[]
  tracking_status: string
  tracking_stage: 'UNKNOWN' | 'COLLECTED' | 'IN_TRANSIT_AIR' | 'CUSTOMS' | 'LAST_MILE' | 'DELIVERED'
  tracking_stage_rank: number
  tracking_has_exception: boolean
  tracking_status_waybill: YanwenTrackingStatusWaybill
  last_mile_tracking_expected: boolean
  origin_country: string
  destination_country: string
}

export interface YanwenTrackingAlert {
  tracking_number: string
  waybill_number: string
  destination_country: string
  alert_type: 'CUSTOMS_STAGNATION' | 'IN_TRANSIT_TIMEOUT'
  severity: 'ALERT' | 'CRITICAL'
  tracking_status: string
  latest_checkpoint_status: string
  checkpoint_time_stamp: string
  checkpoint_time_zone: string
  checkpoint_at: string
  elapsed_hours: number
  elapsed_business_days: number
  snapshot_last_synced_at: string
  message: string
}

const readPayload = (response: unknown, endpoint: string) => (
  unwrapApiPayload(response, endpoint)
)

const readObjectPayload = (response: unknown, endpoint: string) => (
  requireApiObject(readPayload(response, endpoint), endpoint)
)

const readDataArray = (response: unknown, endpoint: string) => (
  requireApiArrayField(readObjectPayload(response, endpoint), 'data', endpoint)
)

const readEntity = (response: unknown, endpoint: string) => {
  const entity = readObjectPayload(response, endpoint)
  requireApiNumberField(entity, 'id', endpoint)
  return entity
}

export const yanwenLogisticsAdminApi = {
  async getYanwenApiConfig(environment: 'fat' | 'production' = 'fat'): Promise<YanwenAPIConfigView> {
    const endpoint = '/api/admin/logistics/yanwen/api/config'
    return readObjectPayload(await axios.get(endpoint, { params: { environment } }), endpoint) as YanwenAPIConfigView
  },
  async saveYanwenApiConfig(payload: APIPayload) {
    const endpoint = '/api/admin/logistics/yanwen/api/config'
    return requireApiAcknowledgement(await axios.put(endpoint, payload), endpoint)
  },
  async pingYanwenApi(payload: APIPayload): Promise<YanwenPingResult> {
    const endpoint = '/api/admin/logistics/yanwen/api/ping'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as YanwenPingResult
  },
  async getYanwenOperationsOverview(environment: 'fat' | 'production' = 'production'): Promise<YanwenOperationsOverview> {
    const endpoint = '/api/admin/logistics/yanwen/overview'
    return readObjectPayload(await axios.get(endpoint, { params: { environment } }), endpoint) as YanwenOperationsOverview
  },
  async listYanwenProducts(environment: 'fat' | 'production' = 'fat'): Promise<YanwenProductCatalogEntry[]> {
    const endpoint = '/api/admin/logistics/yanwen/api/products'
    return readDataArray(await axios.get(endpoint, { params: { environment } }), endpoint) as YanwenProductCatalogEntry[]
  },
  async syncYanwenProducts(payload: APIPayload): Promise<YanwenProductSyncSummary> {
    const endpoint = '/api/admin/logistics/yanwen/api/sync-products'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as YanwenProductSyncSummary
  },
  async listYanwenCountries(environment: 'fat' | 'production' = 'fat'): Promise<YanwenCountryCatalogEntry[]> {
    const endpoint = '/api/admin/logistics/yanwen/api/countries'
    return readDataArray(await axios.get(endpoint, { params: { environment } }), endpoint) as YanwenCountryCatalogEntry[]
  },
  async syncYanwenCountries(payload: APIPayload): Promise<YanwenCountrySyncSummary> {
    const endpoint = '/api/admin/logistics/yanwen/api/sync-countries'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as YanwenCountrySyncSummary
  },
  async listYanwenWarehouses(environment: 'fat' | 'production' = 'fat'): Promise<YanwenWarehouseCatalogEntry[]> {
    const endpoint = '/api/admin/logistics/yanwen/api/warehouses'
    return readDataArray(await axios.get(endpoint, { params: { environment } }), endpoint) as YanwenWarehouseCatalogEntry[]
  },
  async syncYanwenWarehouses(payload: APIPayload): Promise<YanwenWarehouseSyncSummary> {
    const endpoint = '/api/admin/logistics/yanwen/api/sync-warehouses'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as YanwenWarehouseSyncSummary
  },
  async verifyYanwenKoreaPersonalCustomsClearanceCode(payload: APIPayload): Promise<YanwenKoreaPersonalCustomsClearanceCodeVerificationResult> {
    const endpoint = '/api/admin/logistics/yanwen/customs/korea/pccc'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as YanwenKoreaPersonalCustomsClearanceCodeVerificationResult
  },
  async verifyYanwenUnitedStatesAddress(payload: APIPayload): Promise<YanwenUnitedStatesAddressVerificationResult> {
    const endpoint = '/api/admin/logistics/yanwen/customs/united-states/address'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as YanwenUnitedStatesAddressVerificationResult
  },
  async listYanwenChannels(params: APIParams = {}): Promise<YanwenPublishedChannel[]> {
    const endpoint = '/api/admin/logistics/yanwen/channels'
    return readDataArray(await axios.get(endpoint, { params }), endpoint) as YanwenPublishedChannel[]
  },

  async createYanwenChannel(payload: APIPayload): Promise<YanwenPublishedChannel> {
    const endpoint = '/api/admin/logistics/yanwen/channels'
    return readEntity(await axios.post(endpoint, payload), endpoint) as YanwenPublishedChannel
  },

  async updateYanwenChannel(id: APIID, payload: APIPayload): Promise<YanwenPublishedChannel> {
    const endpoint = `/api/admin/logistics/yanwen/channels/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint) as YanwenPublishedChannel
  },

  async deleteYanwenChannel(id: APIID) {
    const endpoint = `/api/admin/logistics/yanwen/channels/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
  async listYanwenWaybills(params: APIParams = {}): Promise<YanwenWaybill[]> {
    const endpoint = '/api/admin/logistics/yanwen/waybills'
    return readDataArray(await axios.get(endpoint, { params }), endpoint) as YanwenWaybill[]
  },
  async createYanwenWaybill(payload: APIPayload): Promise<YanwenWaybill> {
    const endpoint = '/api/admin/logistics/yanwen/waybills'
    return readEntity(await axios.post(endpoint, payload), endpoint) as YanwenWaybill
  },
  async createYanwenWaybills(requests: YanwenWaybillCreationRequest[]): Promise<YanwenBatchWaybillCreationResult> {
    const endpoint = '/api/admin/logistics/yanwen/waybills/batch'
    return readObjectPayload(await axios.post(endpoint, { requests }), endpoint) as YanwenBatchWaybillCreationResult
  },
  async syncYanwenWaybillOfficialDetails(id: APIID): Promise<YanwenWaybill> {
    const endpoint = `/api/admin/logistics/yanwen/waybills/${id}/sync`
    return readObjectPayload(await axios.post(endpoint), endpoint) as YanwenWaybill
  },
  async syncYanwenWaybillsOfficialDetails(ids: number[]): Promise<YanwenWaybill[]> {
    const endpoint = '/api/admin/logistics/yanwen/waybills/batch-sync'
    return readDataArray(await axios.post(endpoint, { waybill_ids: ids }), endpoint) as YanwenWaybill[]
  },
  async getYanwenWaybillLabel(id: APIID): Promise<YanwenWaybillLabel> {
    const endpoint = `/api/admin/logistics/yanwen/waybills/${id}/label`
    return readObjectPayload(await axios.post(endpoint), endpoint) as YanwenWaybillLabel
  },
  async downloadYanwenWaybillLabels(ids: number[]): Promise<YanwenBatchWaybillLabelDownload> {
    const endpoint = '/api/admin/logistics/yanwen/waybills/batch-labels'
    const response = await axios.post<Blob>(endpoint, { waybill_ids: ids }, { responseType: 'blob' })
    return {
      blob: response.data,
      succeeded: Number(response.headers['x-yanwen-label-succeeded'] || 0),
      failed: Number(response.headers['x-yanwen-label-failed'] || 0),
    }
  },
  async cancelYanwenWaybill(id: APIID, note = ''): Promise<YanwenWaybillCancellationResult> {
    const endpoint = `/api/admin/logistics/yanwen/waybills/${id}/cancel`
    return readObjectPayload(await axios.post(endpoint, { note }), endpoint) as YanwenWaybillCancellationResult
  },
  async cancelYanwenWaybills(ids: number[], note = ''): Promise<YanwenBatchWaybillCancellationResult> {
    const endpoint = '/api/admin/logistics/yanwen/waybills/batch-cancel'
    return readObjectPayload(await axios.post(endpoint, { waybill_ids: ids, note }), endpoint) as YanwenBatchWaybillCancellationResult
  },
  async queryYanwenTracking(trackingNumbers: string[]): Promise<YanwenTrackingResult[]> {
    const endpoint = '/api/admin/logistics/yanwen/tracking'
    return readDataArray(await axios.get(endpoint, { params: { nums: trackingNumbers.join(',') } }), endpoint) as YanwenTrackingResult[]
  },
  async listYanwenTrackingAlerts(): Promise<YanwenTrackingAlert[]> {
    const endpoint = '/api/admin/logistics/yanwen/tracking/alerts'
    return readDataArray(await axios.get(endpoint), endpoint) as YanwenTrackingAlert[]
  },


}

export default yanwenLogisticsAdminApi

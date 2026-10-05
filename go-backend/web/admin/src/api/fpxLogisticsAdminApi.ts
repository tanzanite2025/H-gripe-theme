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

export interface FpxOverview {
  generated_at: string
  channels: {
    total: number
    enabled: number
  }
}

export interface FpxAPIConfigView {
  environment: 'production' | 'test'
  endpoint: string
  app_key_configured: boolean
  app_secret_configured: boolean
  access_token_configured: boolean
  enabled: boolean
  last_sync_status: string
  last_synced_at?: string
  last_error?: string
  last_sync_scanned: number
  last_sync_added: number
  last_sync_updated: number
  last_sync_preserved_enabled: number
}

export interface FpxPingResult {
  ok: boolean
  message: string
  latency_ms: number
}

export interface FpxChannelSyncSummary {
  scanned: number
  added: number
  updated: number
  preserved_enabled: number
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

export const fpxLogisticsAdminApi = {
  async getFpxApiConfig(environment: 'production' | 'test' = 'production'): Promise<FpxAPIConfigView> {
    const endpoint = '/api/admin/logistics/4px/api/config'
    return readObjectPayload(await axios.get(endpoint, { params: { environment } }), endpoint) as FpxAPIConfigView
  },

  async saveFpxApiConfig(payload: APIPayload) {
    const endpoint = '/api/admin/logistics/4px/api/config'
    return requireApiAcknowledgement(await axios.put(endpoint, payload), endpoint)
  },
  async pingFpxApi(payload: APIPayload): Promise<FpxPingResult> {
    const endpoint = '/api/admin/logistics/4px/api/ping'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as FpxPingResult
  },
  async syncFpxChannels(payload: APIPayload = { environment: 'production' }): Promise<FpxChannelSyncSummary> {
    const endpoint = '/api/admin/logistics/4px/api/sync-channels'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint) as FpxChannelSyncSummary
  },
  async getFpxOverview(): Promise<FpxOverview> {
    const endpoint = '/api/admin/logistics/4px/overview'
    return readObjectPayload(await axios.get(endpoint), endpoint) as FpxOverview
  },

  async listFpxChannels(params: APIParams = {}) {
    const endpoint = '/api/admin/logistics/4px/channels'
    return readDataArray(await axios.get(endpoint, { params }), endpoint)
  },

  async confirmFpxChannel(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/logistics/4px/channels/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

}

export default fpxLogisticsAdminApi

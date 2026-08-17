import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  requireApiSuccess,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type MediaID = string | number

export interface MediaAsset {
  id: MediaID
  filename?: string | null
  original_filename?: string | null
  alt?: string | null
  url?: string | null
  access_url?: string | null
  media_type?: string | null
  status?: string | null
  visibility?: string | null
  size?: number | string | null
  width?: number | string | null
  height?: number | string | null
  derivatives?: MediaAssetDerivative[]
  [key: string]: unknown
}

export interface MediaAssetDerivative {
  id?: MediaID
  media_asset_id?: MediaID
  preset?: string | null
  url?: string | null
  width?: number | string | null
  height?: number | string | null
  mime_type?: string | null
}

export interface MediaDerivativePreset {
  id: number
  code: string
  label: string
  max_width: number
  sort_order: number
  enabled: boolean
  generation_version: number
  is_system: boolean
  generated_derivatives: number
  created_at?: string
  updated_at?: string
}

export interface MediaDerivativePresetInput {
  code: string
  label: string
  max_width: number
  sort_order: number
  enabled?: boolean
}

export interface MediaDerivativeRebuildJob {
  id: number
  reason: string
  status: 'pending' | 'running' | 'succeeded'
  cursor_asset_id: number
  scanned_assets: number
  generated_assets: number
  generated_derivatives: number
  failed_assets: number
  updated_product_media_rows: number
  last_error?: string
  started_at?: string
  finished_at?: string
  created_at?: string
  updated_at?: string
}

export type MediaDerivativePresetUpdateInput = Omit<MediaDerivativePresetInput, 'code'> & {
  code?: string
}

export interface MediaPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface MediaAssetListParams {
  page?: number
  page_size?: number
  search?: string
  media_type?: string
  status?: string
  visibility?: string
}

export interface MediaAssetListResult {
  assets: MediaAsset[]
  pagination: MediaPagination
}

export interface MediaReference {
  resource_type?: string | null
  resource_id?: string | number | null
  field?: string | null
  label?: string | null
  category?: string | null
  [key: string]: unknown
}

export interface MediaReferenceReport {
  references: MediaReference[]
  total: number
}

export const mediaApi = {
  async listDerivativePresets(): Promise<MediaDerivativePreset[]> {
    const path = '/api/admin/media/derivative-presets'
    const payload = requireApiObject(unwrapApiPayload(await axios.get(path), path), path)
    return requireApiArrayField<MediaDerivativePreset>(payload, 'presets', path)
  },

  async createDerivativePreset(payload: MediaDerivativePresetInput): Promise<MediaDerivativePreset> {
    const path = '/api/admin/media/derivative-presets'
    const result = requireApiObject(unwrapApiPayload(await axios.post(path, payload), path), path)
    return requireApiObjectField<MediaDerivativePreset>(result, 'preset', path)
  },

  async updateDerivativePreset(id: number, payload: MediaDerivativePresetUpdateInput): Promise<MediaDerivativePreset> {
    const path = `/api/admin/media/derivative-presets/${id}`
    const result = requireApiObject(unwrapApiPayload(await axios.put(path, payload), path), path)
    return requireApiObjectField<MediaDerivativePreset>(result, 'preset', path)
  },

  async updateDerivativePresetEnabled(id: number, enabled: boolean): Promise<MediaDerivativePreset> {
    const path = `/api/admin/media/derivative-presets/${id}/enabled`
    const result = requireApiObject(unwrapApiPayload(await axios.patch(path, { enabled }), path), path)
    return requireApiObjectField<MediaDerivativePreset>(result, 'preset', path)
  },

  async deleteDerivativePreset(id: number): Promise<void> {
    const path = `/api/admin/media/derivative-presets/${id}`
    requireApiSuccess(await axios.delete(path), path)
  },

  async listDerivativeRebuildJobs(): Promise<MediaDerivativeRebuildJob[]> {
    const path = '/api/admin/media/derivative-rebuild-jobs'
    const payload = requireApiObject(unwrapApiPayload(await axios.get(path), path), path)
    return requireApiArrayField<MediaDerivativeRebuildJob>(payload, 'jobs', path)
  },

  async requestDerivativeRebuild(): Promise<MediaDerivativeRebuildJob> {
    const path = '/api/admin/media/derivative-rebuild-jobs'
    const result = requireApiObject(unwrapApiPayload(await axios.post(path), path), path)
    return requireApiObjectField<MediaDerivativeRebuildJob>(result, 'job', path)
  },

  async listAssets(params: MediaAssetListParams = {}): Promise<MediaAssetListResult> {
    const path = '/api/admin/media/assets'
    const response = await axios.get(path, { params })
    const body = requireApiObject(response.data, path, 'response body')
    const payload = unwrapApiPayload(response, path)
    const payloadObject = requireApiObject(payload, path)

    return {
      assets: requireApiArrayField<MediaAsset>(payloadObject, 'assets', path),
      pagination: requireApiPagination(body, payloadObject, path),
    }
  },

  async getAsset(id: MediaID): Promise<MediaAsset> {
    const path = `/api/admin/media/assets/${id}`
    const payload = requireApiObject(unwrapApiPayload(await axios.get(path), path), path)
    return requireApiObjectField<MediaAsset>(payload, 'asset', path)
  },

  async uploadAsset(formData: FormData): Promise<MediaAsset> {
    const path = '/api/admin/media/assets'
    const payload = requireApiObject(unwrapApiPayload(await axios.post(path, formData), path), path)
    return requireApiObjectField<MediaAsset>(payload, 'asset', path)
  },

  async downloadCopyrightEvidence(id: MediaID) {
    return axios.get(`/api/admin/media/assets/${id}/copyright-evidence`, {
      responseType: 'blob'
    })
  },

  async getAssetReferences(id: MediaID): Promise<MediaReferenceReport> {
    const path = `/api/admin/media/assets/${id}/references`
    const payload = requireApiObject(unwrapApiPayload(await axios.get(path), path), path)
    return {
      references: requireApiArrayField<MediaReference>(payload, 'references', path),
      total: requireApiNumberField(payload, 'total', path),
    }
  },

  async updateAsset(id: MediaID, asset: Record<string, unknown>): Promise<MediaAsset> {
    const path = `/api/admin/media/assets/${id}`
    const payload = requireApiObject(unwrapApiPayload(await axios.patch(path, asset), path), path)
    return requireApiObjectField<MediaAsset>(payload, 'asset', path)
  },

  async deleteAsset(id: MediaID, payload: Record<string, unknown> = {}): Promise<void> {
    const path = `/api/admin/media/assets/${id}`
    requireApiSuccess(await axios.delete(path, { data: payload }), path)
  }
}

export default mediaApi

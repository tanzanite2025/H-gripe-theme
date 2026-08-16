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
  [key: string]: unknown
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

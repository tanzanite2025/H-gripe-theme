import axios from '@/utils/axios'

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

const unwrapPayload = (response: any): any => response.data?.data ?? response.data ?? {}

export const mediaApi = {
  async listAssets(params: MediaAssetListParams = {}): Promise<MediaAssetListResult> {
    const response = await axios.get('/api/admin/media/assets', { params })
    const raw = response.data ?? {}
    const payload = raw.data ?? raw ?? {}
    return {
      assets: payload.assets ?? payload.data ?? [],
      pagination: raw.pagination ?? payload.pagination ?? responsePagination(payload),
    }
  },

  async getAsset(id: MediaID): Promise<MediaAsset | null> {
    const payload = unwrapPayload(await axios.get(`/api/admin/media/assets/${id}`))
    return payload.asset ?? null
  },

  async uploadAsset(formData: FormData): Promise<MediaAsset> {
    const payload = unwrapPayload(await axios.post('/api/admin/media/assets', formData))
    return payload.asset ?? payload
  },

  async downloadCopyrightEvidence(id: MediaID) {
    return axios.get(`/api/admin/media/assets/${id}/copyright-evidence`, {
      responseType: 'blob'
    })
  },

  async getAssetReferences(id: MediaID): Promise<MediaReferenceReport> {
    const payload = unwrapPayload(await axios.get(`/api/admin/media/assets/${id}/references`))
    return {
      references: payload.references ?? [],
      total: Number(payload.total ?? payload.references?.length ?? 0),
    }
  },

  async updateAsset(id: MediaID, asset: Record<string, unknown>): Promise<MediaAsset | null> {
    const payload = unwrapPayload(await axios.patch(`/api/admin/media/assets/${id}`, asset))
    return payload.asset ?? null
  },

  async deleteAsset(id: MediaID, payload: Record<string, unknown> = {}): Promise<any> {
    return unwrapPayload(await axios.delete(`/api/admin/media/assets/${id}`, { data: payload }))
  }
}

const responsePagination = (payload: Record<string, any>): MediaPagination => ({
  page: Number(payload.page || 1),
  page_size: Number(payload.page_size || 20),
  total: Number(payload.total || 0),
  total_pages: Number(payload.total_pages || 0),
})

export default mediaApi

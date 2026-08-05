import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

export const mediaApi = {
  async listAssets(params: Record<string, any> = {}) {
    const response = await axios.get('/api/admin/media/assets', { params })
    const raw = response.data ?? {}
    const payload = raw.data ?? raw ?? {}
    return {
      assets: payload.assets ?? payload.data ?? [],
      pagination: raw.pagination ?? payload.pagination ?? responsePagination(payload),
    }
  },

  async getAsset(id: number | string) {
    const payload = unwrapPayload(await axios.get(`/api/admin/media/assets/${id}`))
    return payload.asset ?? null
  },

  async uploadAsset(formData: FormData) {
    const payload = unwrapPayload(await axios.post('/api/admin/media/assets', formData))
    return payload.asset ?? payload
  },

  async downloadCopyrightEvidence(id: number | string) {
    return axios.get(`/api/admin/media/assets/${id}/copyright-evidence`, {
      responseType: 'blob'
    })
  },

  async getAssetReferences(id: number | string) {
    const payload = unwrapPayload(await axios.get(`/api/admin/media/assets/${id}/references`))
    return {
      references: payload.references ?? [],
      total: Number(payload.total ?? payload.references?.length ?? 0),
    }
  },

  async updateAsset(id: number | string, asset: Record<string, any>) {
    const payload = unwrapPayload(await axios.patch(`/api/admin/media/assets/${id}`, asset))
    return payload.asset ?? null
  },

  async deleteAsset(id: number | string, payload: Record<string, any> = {}) {
    return unwrapPayload(await axios.delete(`/api/admin/media/assets/${id}`, { data: payload }))
  }
}

const responsePagination = (payload: Record<string, any>) => ({
  page: Number(payload.page || 1),
  page_size: Number(payload.page_size || 20),
  total: Number(payload.total || 0),
  total_pages: Number(payload.total_pages || 0),
})

export default mediaApi

import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

export const mediaApi = {
  async uploadAsset(formData: FormData) {
    const payload = unwrapPayload(await axios.post('/api/admin/media/assets', formData))
    return payload.asset ?? payload
  }
}

export default mediaApi

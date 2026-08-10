import axios from '@/utils/axios'
import type { SEOHomeApi, SEOSettings, SEOUpdateRequest } from './types'

const unwrap = (response: { data?: { data?: SEOSettings } | SEOSettings }): SEOSettings => {
  const payload = response.data
  return (payload && 'data' in payload ? payload.data : payload) as SEOSettings
}

export const createSEOHomeApi = (): SEOHomeApi => ({
  async get(locale = 'en'): Promise<SEOSettings> {
    return unwrap(await axios.get('/api/admin/seo/home', { params: { locale } }))
  },

  async update(payload: SEOUpdateRequest): Promise<SEOSettings> {
    return unwrap(await axios.put('/api/admin/seo/home', payload))
  },
})

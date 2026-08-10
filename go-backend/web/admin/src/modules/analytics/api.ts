import axios from '@/utils/axios'
import type { AnalyticsSettings, AnalyticsUpdateRequest } from './types'

const unwrap = (response: { data?: { data?: AnalyticsSettings } | AnalyticsSettings }): AnalyticsSettings => {
  const payload = response.data
  return (payload && 'data' in payload ? payload.data : payload) as AnalyticsSettings
}

export const analyticsApi = {
  async get(locale = 'en'): Promise<AnalyticsSettings> {
    return unwrap(await axios.get('/api/admin/analytics', { params: { locale } }))
  },

  async update(payload: AnalyticsUpdateRequest): Promise<AnalyticsSettings> {
    return unwrap(await axios.put('/api/admin/analytics', payload))
  },
}

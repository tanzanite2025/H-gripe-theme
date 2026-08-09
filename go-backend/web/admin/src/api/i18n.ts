import axios from '@/utils/axios'

const unwrapPayload = (response: any): any => response.data ?? {}

export const i18nApi = {
  async listLanguages() {
    return unwrapPayload(await axios.get('/api/v1/i18n/languages'))
  },
}

export default i18nApi

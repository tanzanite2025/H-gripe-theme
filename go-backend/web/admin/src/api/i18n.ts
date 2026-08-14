import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export const i18nApi = {
  async listLanguages() {
    const endpoint = '/api/v1/i18n/languages'
    const payload = requireApiObject(unwrapApiPayload(await axios.get(endpoint), endpoint), endpoint)
    requireApiArrayField(payload, 'languages', endpoint)
    requireApiNumberField(payload, 'total', endpoint)
    return payload
  },
}

export default i18nApi

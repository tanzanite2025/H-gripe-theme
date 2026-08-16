import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArray,
  requireApiObject,
  requireApiObjectField,
  requireApiNumberField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

const endpoint = '/api/admin/customs-classifications'

export const customsClassificationApi = {
  async list(params: Record<string, any> = {}) {
    return requireApiArray(unwrapApiPayload(await axios.get(endpoint, { params }), endpoint), endpoint, 'data')
  },

  async lookup(params: Record<string, any> = {}) {
    const path = `${endpoint}/lookup`
    return requireApiArray(unwrapApiPayload(await axios.get(path, { params }), path), path, 'data')
  },

  async create(payload: Record<string, any>) {
    return requireApiObject(unwrapApiPayload(await axios.post(endpoint, payload), endpoint), endpoint, 'data')
  },

  async update(id: number | string, payload: Record<string, any>) {
    const path = `${endpoint}/${id}`
    return requireApiObject(unwrapApiPayload(await axios.put(path, payload), path), path, 'data')
  },

  async remove(id: number | string) {
    const path = `${endpoint}/${id}`
    return requireApiAcknowledgement(await axios.delete(path), path)
  },
}

export const customsSummaryApi = {
  async get(locale = 'en') {
    const path = '/api/admin/products/customs-summary'
    const payload = requireApiObject(unwrapApiPayload(await axios.get(path, { params: { locale } }), path), path)
    const summary = requireApiObjectField(payload, 'summary', path)
    requireApiNumberField(summary, 'total', path)
    requireApiNumberField(summary, 'complete', path)
    requireApiNumberField(summary, 'incomplete', path)
    requireApiNumberField(summary, 'missing_hs_code', path)
    requireApiNumberField(summary, 'missing_cn_code', path)
    requireApiNumberField(summary, 'missing_country_of_origin', path)
    requireApiNumberField(summary, 'missing_customs_description', path)
    return summary
  },
}

export default customsClassificationApi

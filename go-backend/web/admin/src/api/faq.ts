import axios from '@/utils/axios'

type APIID = string | number
type APIParams = Record<string, any>
type APIPayload = Record<string, any>

const unwrapPayload = (response: any): any => response.data ?? {}

export const faqAdminApi = {
  async listFAQs(params: APIParams = {}) {
    return unwrapPayload(await axios.get('/api/admin/faqs', { params }))
  },

  async listFAQGroups(params: APIParams = {}) {
    return unwrapPayload(await axios.get('/api/admin/faqs/grouped', { params }))
  },

  async getFAQ(id: APIID) {
    return unwrapPayload(await axios.get(`/api/admin/faqs/${id}`))
  },

  async createFAQ(payload: APIPayload) {
    return unwrapPayload(await axios.post('/api/admin/faqs', payload))
  },

  async updateFAQ(id: APIID, payload: APIPayload) {
    return unwrapPayload(await axios.put(`/api/admin/faqs/${id}`, payload))
  },

  async deleteFAQ(id: APIID) {
    return unwrapPayload(await axios.delete(`/api/admin/faqs/${id}`))
  },

  async deleteFAQs(ids: APIID[]) {
    return unwrapPayload(await axios.post('/api/admin/faqs/batch-delete', { faq_ids: ids }))
  },

  async listStructure(locale: string) {
    return unwrapPayload(await axios.get('/api/admin/faqs/structure', { params: { locale } }))
  },

  async updatePage(pageID: APIID, payload: APIPayload) {
    return unwrapPayload(await axios.put(`/api/admin/faqs/pages/${pageID}`, payload))
  },

  async createCategory(payload: APIPayload) {
    return unwrapPayload(await axios.post('/api/admin/faqs/categories', payload))
  },

  async updateCategory(id: APIID, payload: APIPayload) {
    return unwrapPayload(await axios.put(`/api/admin/faqs/categories/${id}`, payload))
  },

  async deleteCategory(id: APIID) {
    return unwrapPayload(await axios.delete(`/api/admin/faqs/categories/${id}`))
  },
}

export default faqAdminApi

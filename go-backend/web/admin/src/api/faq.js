import axios from '@/utils/axios'

const unwrapPayload = (response) => response.data ?? {}

export const faqAdminApi = {
  async listFAQs(params = {}) {
    return unwrapPayload(await axios.get('/api/admin/faqs', { params }))
  },

  async getFAQ(id) {
    return unwrapPayload(await axios.get(`/api/admin/faqs/${id}`))
  },

  async createFAQ(payload) {
    return unwrapPayload(await axios.post('/api/admin/faqs', payload))
  },

  async updateFAQ(id, payload) {
    return unwrapPayload(await axios.put(`/api/admin/faqs/${id}`, payload))
  },

  async deleteFAQ(id) {
    return unwrapPayload(await axios.delete(`/api/admin/faqs/${id}`))
  },

  async deleteFAQs(ids) {
    return unwrapPayload(await axios.post('/api/admin/faqs/batch-delete', { faq_ids: ids }))
  },

  async listStructure(locale) {
    return unwrapPayload(await axios.get('/api/admin/faqs/structure', { params: { locale } }))
  },

  async updatePage(pageID, payload) {
    return unwrapPayload(await axios.put(`/api/admin/faqs/pages/${pageID}`, payload))
  },

  async createCategory(payload) {
    return unwrapPayload(await axios.post('/api/admin/faqs/categories', payload))
  },

  async updateCategory(id, payload) {
    return unwrapPayload(await axios.put(`/api/admin/faqs/categories/${id}`, payload))
  },

  async deleteCategory(id) {
    return unwrapPayload(await axios.delete(`/api/admin/faqs/categories/${id}`))
  }
}

export default faqAdminApi

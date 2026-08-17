import axios from '@/utils/axios'
import type {
  StorefrontRouteCatalogCheckSummary,
  StorefrontRouteCatalogHistoryResponse,
  StorefrontRouteCatalogListParams,
  StorefrontRouteCatalogListResponse,
  StorefrontRouteCatalogStats,
  StorefrontRouteCatalogSyncSummary,
  StorefrontRouteCheckResult,
  StorefrontSitemapOverview,
  StorefrontSitemapSyncResponse,
} from './routeCatalogTypes'

export const storefrontRouteCatalogApi = {
  async stats(): Promise<StorefrontRouteCatalogStats> {
    const response = await axios.get('/api/admin/urls/stats')
    return response.data?.data || {}
  },

  async list(params: StorefrontRouteCatalogListParams): Promise<StorefrontRouteCatalogListResponse> {
    const response = await axios.get('/api/admin/urls/routes', { params })
    return {
      items: Array.isArray(response.data?.items) ? response.data.items : [],
      pagination: {
        page: params.page,
        page_size: params.page_size,
        total: Number(response.data?.pagination?.total || 0),
        total_pages: Number(response.data?.pagination?.total_pages || 0),
      },
    }
  },

  async get(id: number): Promise<StorefrontRouteCatalogListResponse['items'][number]> {
    const response = await axios.get(`/api/admin/urls/routes/${id}`)
    return response.data?.data || {}
  },

  async history(id: number, params: { page: number; page_size: number }): Promise<StorefrontRouteCatalogHistoryResponse> {
    const response = await axios.get(`/api/admin/urls/routes/${id}/history`, { params })
    return {
      items: Array.isArray(response.data?.items) ? response.data.items : [],
      pagination: {
        page: params.page,
        page_size: params.page_size,
        total: Number(response.data?.pagination?.total || 0),
        total_pages: Number(response.data?.pagination?.total_pages || 0),
      },
    }
  },

  async sync(): Promise<StorefrontRouteCatalogSyncSummary> {
    const response = await axios.post('/api/admin/urls/sync')
    return response.data?.data || {}
  },

  async sitemap(): Promise<StorefrontSitemapOverview> {
    const response = await axios.get('/api/admin/urls/sitemap')
    return response.data?.data || {}
  },

  async syncSitemap(): Promise<StorefrontSitemapSyncResponse> {
    const response = await axios.post('/api/admin/urls/sitemap/sync')
    return response.data?.data || {}
  },

  async checkOne(id: number): Promise<StorefrontRouteCheckResult> {
    const response = await axios.post(`/api/admin/urls/routes/${id}/check`)
    return response.data?.data || {}
  },

  async check(params: Omit<StorefrontRouteCatalogListParams, 'page' | 'page_size'> & { limit?: number }): Promise<StorefrontRouteCatalogCheckSummary> {
    const response = await axios.post('/api/admin/urls/check', null, { params })
    return response.data?.data || {}
  },
}

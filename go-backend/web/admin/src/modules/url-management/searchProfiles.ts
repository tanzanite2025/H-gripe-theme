import axios from '@/utils/axios'
import type { StorefrontRouteCatalogEntry } from './routeCatalogTypes'

export interface StorefrontURLSearchProfile {
  id: number
  route_entry_id: number
  enabled: boolean
  search_weight: number
  keywords: string[]
  display_title: string
  display_summary: string
  created_at: string
  updated_at: string
  route_entry?: StorefrontRouteCatalogEntry | null
}

export interface StorefrontURLSearchProfileInput {
  enabled: boolean
  search_weight: number
  keywords: string[]
  display_title: string
  display_summary: string
}

export interface StorefrontURLSearchProfileStats {
  total_routes: number
  public_indexed: number
  explicit_profiles: number
  explicit_enabled: number
  keyword_count: number
  unconfigured: number
}

export const storefrontURLSearchProfilesApi = {
  async list(locale?: string): Promise<StorefrontURLSearchProfile[]> {
    const response = await axios.get('/api/admin/urls/search-profiles', {
      params: locale ? { locale } : undefined,
    })
    return Array.isArray(response.data?.items) ? response.data.items : []
  },

  async stats(locale?: string): Promise<StorefrontURLSearchProfileStats> {
    const response = await axios.get('/api/admin/urls/search-profiles/stats', {
      params: locale ? { locale } : undefined,
    })
    return response.data?.data || {}
  },

  async get(routeEntryID: number): Promise<StorefrontURLSearchProfile> {
    const response = await axios.get(`/api/admin/urls/search-profiles/${routeEntryID}`)
    return response.data?.data || {}
  },

  async upsert(routeEntryID: number, input: StorefrontURLSearchProfileInput): Promise<StorefrontURLSearchProfile> {
    const response = await axios.put(`/api/admin/urls/search-profiles/${routeEntryID}`, input)
    return response.data?.data || {}
  },
}

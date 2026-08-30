export interface StorefrontURLSearchRouteEntry {
  id: number
  route_key: string
  path: string
  locale: string
  source_type: string
  source_key?: string | null
  title?: string | null
  summary?: string | null
  canonical_path?: string | null
  is_alias: boolean
  is_searchable: boolean
  is_checkable: boolean
  is_indexable: boolean
  entry_status: string
}

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
  route_entry?: StorefrontURLSearchRouteEntry | null
}

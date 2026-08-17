import type { SEOResourcePagination } from '@/modules/seo/types'

export interface StorefrontRouteCatalogEntry {
  id: number
  route_key: string
  path: string
  locale: string
  source_type: string
  source_id?: number | null
  source_key?: string | null
  title?: string | null
  summary?: string | null
  canonical_path?: string | null
  is_alias: boolean
  is_searchable: boolean
  is_checkable: boolean
  is_indexable: boolean
  entry_status: string
  duplicate_group_key?: string | null
  last_check_status?: string | null
  last_http_status?: number | null
  last_final_url?: string | null
  last_canonical_url?: string | null
  last_response_ms?: number | null
  last_redirect_count?: number | null
  last_check_error?: string | null
  last_checked_at?: string | null
}

export interface StorefrontRouteCatalogListParams {
  page: number
  page_size: number
  locale?: string
  source_type?: string
  entry_status?: string
  check_status?: string
  search?: string
  searchable?: string
  include_aliases?: boolean
  needs_attention?: boolean
  problem_scope?: 'canonical'
}

export interface StorefrontRouteCatalogListResponse {
  items: StorefrontRouteCatalogEntry[]
  pagination: SEOResourcePagination
}

export interface StorefrontRouteCatalogSyncSummary {
  manifest_version: string
  entries: number
  static_entries: number
  product_entries: number
  blog_entries: number
  alias_entries: number
  duplicates: number
}

export interface StorefrontSitemapOverview {
  public_path: string
  sitemap_url: string
  source: string
  dynamic_source_path: string
  entries: number
  indexable: number
  last_synced_at?: string | null
  manifest_version?: string | null
}

export interface StorefrontSitemapSyncResponse {
  sync: StorefrontRouteCatalogSyncSummary
  sitemap: StorefrontSitemapOverview
}

export interface StorefrontRouteCatalogCheckSummary {
  checked: number
  ok: number
  redirects: number
  not_found: number
  server_errors: number
  canonical_mismatch: number
  errors: number
}

export interface StorefrontRouteCatalogStats {
  total: number
  active: number
  alias: number
  duplicate: number
  stale: number
  needs_attention: number
  checked: number
  unchecked: number
  ok: number
  redirects: number
  not_found: number
  server_errors: number
  canonical_mismatch: number
  errors: number
  searchable: number
  checkable: number
  indexable: number
  sitemap_eligible: number
  last_synced_at?: string | null
  manifest_version?: string | null
}

export interface StorefrontRouteCheckResult {
  id: number
  route_entry_id: number
  checked_at: string
  http_status: number
  final_url?: string | null
  canonical_url?: string | null
  response_ms: number
  redirect_count: number
  content_hash?: string | null
  status: string
  error_message?: string | null
}

export interface StorefrontRouteCatalogHistoryResponse {
  items: StorefrontRouteCheckResult[]
  pagination: SEOResourcePagination
}

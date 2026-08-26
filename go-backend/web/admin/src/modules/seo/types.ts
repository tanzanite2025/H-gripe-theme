export interface SEOSettings {
  meta_title: string
  meta_description: string
}

export interface SEOUpdateRequest extends Partial<SEOSettings> {
  locale?: string
}

export interface SEOHomeApi {
  get(locale?: string): Promise<SEOSettings>
  update(payload: SEOUpdateRequest): Promise<SEOSettings>
}

export interface SEOArticleResource {
  id: number | string
  title?: string | null
  slug?: string | null
  route_path?: string | null
  status?: string | null
  locale?: string | null
  tags?: string | null
  meta_title?: string | null
  meta_description?: string | null
  canonical_url?: string | null
  created_at?: string | null
  published_at?: string | null
}

export interface SEOCategoryResource {
  id: number | string
  name?: string | null
  slug?: string | null
  route_path?: string | null
  status?: string | null
  locale?: string | null
  meta_title?: string | null
  meta_description?: string | null
  intro?: string | null
}

export interface SEOProductResource {
  id: number | string
  name?: string | null
  slug?: string | null
  route_path?: string | null
  status?: string | null
  locale?: string | null
  meta_title?: string | null
  meta_description?: string | null
  diagnostics?: SEOProductDiagnostics | null
  created_at?: string | null
}

export interface SEOProductMetaFieldState {
  value: string
  source: string
  is_custom: boolean
  fallback_active: boolean
  length: number
  soft_length_warning: boolean
}

export interface SEOProductStructuredDataVariant {
  '@type': string
  name: string
  sku?: string
  price?: number
  priceCurrency?: string
  availability?: string
  url: string
}

export interface SEOProductStructuredDataPreview {
  '@context': string
  '@type': string
  name: string
  brand?: {
    '@type': string
    name: string
  }
  description?: string
  image?: string[]
  sku?: string
  url: string
  offers?: {
    '@type': string
    price: number
    priceCurrency: string
    availability: string
    url: string
  }
  productGroupID?: string
  variesBy?: string[]
  hasVariant?: SEOProductStructuredDataVariant[]
}

export interface SEOProductBreadcrumbItem {
  type: string
  id?: number
  name: string
  slug?: string
  path: string
}

export interface SEOProductBreadcrumbPreview {
  status: string
  reason?: string
  items: SEOProductBreadcrumbItem[]
}

export interface SEOProductPrimaryCategory {
  id: number
  name: string
  slug: string
  path: string
}

export interface SEOProductDiagnostics {
  product_name: string
  brand: string
  brand_configured: boolean
  sku: string
  price?: number | null
  currency: string
  availability: string
  image_count: number
  has_image: boolean
  has_offer: boolean
  has_meta_title: boolean
  has_meta_description: boolean
  active_variant_count: number
  variant_count: number
  meta_title: SEOProductMetaFieldState
  meta_description: SEOProductMetaFieldState
  structured_data_type: string
  ready: boolean
  missing: string[]
  blocking_issues: string[]
  warnings: string[]
  structured_data: SEOProductStructuredDataPreview
  primary_category?: SEOProductPrimaryCategory | null
  breadcrumb?: SEOProductBreadcrumbPreview | null
  breadcrumb_path_complete?: boolean
  breadcrumb_ssr_status?: string
  breadcrumb_reason?: string
}

export interface SEOResourceListParams {
  page: number
  page_size: number
  locale?: string
  search?: string
}

export interface SEOResourcePagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface SEOResourceList<T> {
  items: T[]
  pagination: SEOResourcePagination
}

export interface SEOResourceItem {
  id: number | string
  title: string
  routePath: string
  href: string
  locale?: string
  localeLabel: string
  status: string
  metaTitle: string
  metaDescription: string
  intro?: string
  canonicalUrl?: string
  productDiagnostics?: SEOProductDiagnostics | null
}

export interface SEOResourceEditorValues {
  meta_title: string
  meta_description: string
  canonical_url: string
  intro: string
}

export interface GoogleIndexingStatus {
  enabled: boolean
  configured: boolean
  ready: boolean
  message: string
}

export interface GoogleIndexingPushResult {
  product_id: number | string
  url: string
  notification_type: string
  accepted: boolean
  http_status: number
  submitted_at: string
  metadata?: Record<string, unknown> | null
}

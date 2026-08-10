export interface ProductFormRecord {
  id: number | string | null
  product_type_id: number | string | null
  shipping_template_id: number | string | null
  after_sales_template_id: number | string | null
  packaging_template_id: number | string | null
  name: string
  slug: string
  locale: string
  currency: string
  description: string
  short_description: string
  status: string
  featured: boolean
  specs: Record<string, any>
  variants: any[]
  media: any[]
  variant_option_values: ProductVariantOptionValueForm[]
  [key: string]: any
}

export type ProductMediaType = 'image' | 'video'

export interface ProductMediaForm {
  id: number | string | null
  local_key?: string
  variant_id: number | string | null
  variant_option_value_id: number | string | null
  media_asset_id?: number | string | null
  media_type: ProductMediaType | string
  role?: string
  url: string
  thumbnail_url?: string
  poster_url?: string
  alt?: string
  title?: string
  locale?: string
  sort_order?: number | string
  is_primary?: boolean
  is_visible?: boolean
}

export interface ProductVariantForm {
  id?: number | string | null
  sku?: string
  title?: string
  option_values: Record<string, any>
  price?: number | string
  sale_price?: number | string
  weight_grams?: number | string
  shipping_template_id?: number | string | null
  stock?: number | string
  is_active?: boolean
  display_prices?: ProductDisplayPriceResult[]
}

export interface ProductVariantRecord extends ProductVariantForm {
  id?: number | string
  product_id?: number | string
  is_default?: boolean
  is_active?: boolean
  sort_order?: number | string
}

export interface ProductVariantOptionValueForm {
  id: number | string | null
  local_key?: string
  spec_definition_id: number | string
  value_key: string
  label: string
  color_hex: string
  swatch_media_asset_id: number | string | null
  swatch_url: string
  sort_order: number
  is_enabled: boolean
}

export interface ProductSpecDefinition {
  id?: number | string
  slug: string
  name: string
  field_type: string
  presentation?: string
  is_required?: boolean
  is_variant_option?: boolean
  unit?: string
  options?: string
}

export interface ShippingTemplateRecord {
  id: number | string
  name: string
  enabled?: boolean
}

export interface ProductDisplayPriceResult {
  currency?: string
  quote_currency?: string
  amount?: number | string
  fallback_reason?: string
  converted?: boolean
}

export interface ProductRecord {
  id?: number | string
  sku?: string
  name?: string
  media?: unknown
  sale_price?: number | string
  price?: number | string
  currency?: string
  stock?: number | string
  status?: string
  featured?: boolean
  locale?: string
  created_at?: string | number | Date | null
  variants?: ProductVariantRecord[]
  translation_group?: ProductTranslationGroup | null
  [key: string]: any
}

export interface ProductTranslation {
  id: number | string
  parent_id?: number | string | null
  locale?: string | null
  name?: string | null
  slug?: string | null
  sku?: string | null
  status?: string | null
  is_root?: boolean
}

export interface ProductTranslationGroup {
  root_id: number | string
  source_id: number | string
  translations: ProductTranslation[]
  missing_locales: string[]
}

export interface ProductPagination {
  page: number
  pageSize: number
  total: number
}

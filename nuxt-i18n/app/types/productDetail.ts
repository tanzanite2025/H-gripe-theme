export interface ProductMediaImage {
  id?: number | string
  url: string
  alt?: string
}

export interface ProductMediaImageVariant {
  url: string
  width?: number
  height?: number
  mime_type?: string
}

export interface ProductMedia {
  id?: number | string
  url: string
  media_type?: 'image' | 'video' | string
  role?: string
  variant_id?: number | null
  variant_option_value_id?: number | null
  thumbnail_url?: string
  poster_url?: string
  image_variants?: Record<string, ProductMediaImageVariant>
  alt?: string
  title?: string
  is_primary?: boolean
  is_visible?: boolean
}

export interface ProductSpecificationTemplate {
  id?: number
  name: string
  slug: string
  revision?: number
  spec_definitions?: SpecDefinition[]
}

export interface SpecDefinition {
  id?: number
  name: string
  slug: string
  group?: string
  field_type: string
  presentation?: 'text' | 'color' | 'image' | string
  unit?: string
  is_visible?: boolean
  role?: 'attribute' | 'variant' | 'custom_option' | string
  selection_mode?: 'single' | 'multiple' | string
  min_selections?: number
  max_selections?: number | null
  option_items?: ProductSpecOptionItem[]
  sort_order?: number
}

export interface ProductSpecOptionItem {
  id?: number
  value_key: string
  default_label?: string
  color_hex?: string
  swatch_url?: string
  is_enabled_by_default?: boolean
  is_default?: boolean
  default_price_delta_minor?: number
  default_price_currency?: string
  sort_order?: number
}

export interface ProductSpecValue {
  id?: number
  value: string
  definition?: SpecDefinition
}

export interface ProductVariant {
  id: number
  sku?: string
  title?: string
  option_values?: string | Record<string, string>
  currency?: string
  price: number
  sale_price?: number | null
  display_price?: ProductDisplayPrice
  display_prices?: ProductDisplayPrice[]
  weight_grams?: number | null
  availability: ProductAvailability
  is_default?: boolean
  is_active?: boolean
  option_group_rules?: ProductOptionGroupVariantRule[]
  option_value_rules?: ProductOptionValueVariantRule[]
}

export interface ProductOptionGroupVariantRule {
  id?: number
  variant_id?: number
  spec_definition_id?: number
  is_applicable: boolean
  min_selections_override?: number | null
  max_selections_override?: number | null
}

export interface ProductOptionValueVariantRule {
  id?: number
  variant_id?: number
  product_variant_option_value_id?: number
  is_enabled: boolean
  price_delta_minor_override?: number | null
  unavailable_reason?: string
}

export type ProductFulfillmentMode = 'stock' | 'made_to_order'
export type ProductAvailability = 'in_stock' | 'made_to_order' | 'out_of_stock'

export interface ProductVariantOptionValue {
  id: number
  spec_definition_id?: number
  spec_slug: string
  value_key: string
  label: string
  color_hex?: string
  swatch_url?: string
  sort_order?: number
  is_enabled?: boolean
  is_default?: boolean
  price_delta_minor?: number
  inventory_policy?: string
}

export interface ProductInformationTemplate {
  id: number
  kind: 'after_sales' | 'packaging' | string
  name: string
  content: string
  locale?: string
}

export interface ProductLocalizedRoute {
  locale: string
  slug: string
}

export interface ProductBrand {
  id: number
  name: string
  slug: string
  logo_url?: string
  website_url?: string
}

export interface ProductDisplayPrice {
  amount: number
  currency: string
  rate?: number
  source?: string
  converted?: boolean
  fallback_reason?: string
}

export interface ProductReviewSummary {
  product_id?: number
  total_reviews?: number
  average_rating?: number
  rating_5_count?: number
  rating_4_count?: number
  rating_3_count?: number
  rating_2_count?: number
  rating_1_count?: number
}

export interface ProductShippingDetails {
  country?: string
  amount?: number
  currency?: string
  free_shipping?: boolean
  eta_min_days?: number
  eta_max_days?: number
}

export interface ProductBreadcrumbItem {
  type: string
  id?: number
  name: string
  slug?: string
  path: string
}

export interface ProductBreadcrumb {
  status: string
  reason?: string
  items: ProductBreadcrumbItem[]
}

export interface GoProduct {
  id: number
  product_specification_template_id?: number
  product_specification_template?: ProductSpecificationTemplate
  brand?: ProductBrand | null
  name: string
  slug: string
  short_description?: string
  description?: string
  sku?: string
  currency?: string
  price: number
  sale_price?: number
  display_price?: ProductDisplayPrice
  display_prices?: ProductDisplayPrice[]
  availability?: ProductAvailability
  fulfillment_mode?: ProductFulfillmentMode
  media?: ProductMedia[]
  thumbnail?: string
  meta_title?: string
  meta_description?: string
  localized_routes?: ProductLocalizedRoute[]
  after_sales_template?: ProductInformationTemplate | null
  packaging_template?: ProductInformationTemplate | null
  spec_values?: ProductSpecValue[]
  variants?: ProductVariant[]
  variant_option_values?: ProductVariantOptionValue[]
  review_summary?: ProductReviewSummary | null
  shipping_details?: ProductShippingDetails | null
  breadcrumb?: ProductBreadcrumb | null
}

export type ProductPreviewMedia =
  | {
      kind: 'image'
      url: string
      alt: string
    }
  | {
      kind: 'video'
      url: string
      poster?: string
    }

export interface ProductGalleryItem {
  id: string
  kind: 'image' | 'video'
  url: string
  thumbnailUrl: string
  poster?: string
  alt: string
  isPrimary: boolean
  sourceIndex: number
}

export interface ProductVariantOptionGroup {
  slug: string
  name: string
  presentation: 'text' | 'color' | 'image'
  options: Array<{
    value: string
    label: string
    colorHex: string
    swatchUrl: string
    selected: boolean
    available: boolean
  }>
}

export interface ProductCustomOptionGroup {
  slug: string
  name: string
  selectionMode: 'single' | 'multiple' | string
  minSelections: number
  maxSelections: number | null
  selectedCount?: number
  isValid?: boolean
  presentation: string
  options: Array<{
    value: string
    label: string
    colorHex: string
    swatchUrl: string
    selected: boolean
    available: boolean
    unavailableReason?: string
    priceDeltaMinor: number
  }>
}

export interface ProductSpecificationGroup {
  name: string
  items: Array<{
    slug: string
    name: string
    displayValue: string
  }>
}

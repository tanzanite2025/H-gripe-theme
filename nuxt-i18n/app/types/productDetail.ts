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

export interface CarbonRimProductSpecification {
  wheel_size?: string
  material?: 'carbon_fiber'
  material_details?: string
  riding_uses?: string[]
  schema_version: 1
  model: string
  bead_seat_diameter_mm: number
  depth_mm: number
  internal_width_mm: number
  external_width_mm: number
  bead_structure: 'hooked' | 'hookless'
  tubeless_ready: boolean
  rim_tape_required?: boolean | null
  valve_hole_diameter_mm?: number | null
  asymmetric_offset_mm: number
  technical_notes?: string
}

export interface CarbonRimVariantSpecification {
  spoke_hole_count: number
  drilling_pattern: string
  finish_code: string
  erd_mm?: number | null
  erd_measurement_notes?: string
  net_weight_grams: number
}

export interface WheelsetWheelProductSpecification {
  wheel_size: string
  spoke_hole_counts: number[]
  lacing_patterns: string[]
  bead_seat_diameter_mm?: number | null
  material?: string
  brake_interface?: string
  rim_profile?: string
  tire_system?: string
  rim_depth_mm?: number | null
  rim_internal_width_mm?: number | null
  rim_external_width_mm?: number | null
  hub_interface?: string
  axle_standard?: string
  freehub_body?: string
  spoke_type?: string
  weight_g?: number | null
  max_tire_pressure_bar?: number | null
}

export interface WheelsetProductSpecification {
  schema_version: 1
  front_wheel: WheelsetWheelProductSpecification
  rear_wheel: WheelsetWheelProductSpecification
  net_weight_g?: number | null
  max_system_weight_kg?: number | null
  certifications?: string[]
}

export interface SpokeWeightReference {
  spoke_count: number
  length_mm: number
  total_weight_g: number
}

export interface SpokeGaugeOption {
  gauge: string
  diameter_mm?: number
  hub_end_diameter_mm?: number
  middle_diameter_mm?: number
  middle_width_mm?: number
  middle_thickness_mm?: number
  thread_end_diameter_mm?: number
  length_min_mm?: number
  length_max_mm?: number
  weight_reference?: SpokeWeightReference
}

export interface SpokesProductSpecification {
  schema_version: 1
  model_name: string
  manufacturer_part_number?: string
  manufacturer_family?: string
  material?: string
  manufacturer_description?: string
  gauge_options?: SpokeGaugeOption[]
  middle_strength_n_per_mm2?: {
    minimum_n_per_mm2: number
    maximum_n_per_mm2: number
  }
  color_options?: string[]
  manufacturer_options?: string[]
  recommended_uses?: string[]
  manufacturer_claims?: string[]
  source_urls?: string[]
}

export interface ProductVariant {
  carbon_rim_specification?: CarbonRimVariantSpecification
  id: number
  sku?: string
  title?: string
  option_values?: string | Record<string, string>
  currency?: string
  price_decimal: string
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

export interface ProductOptionValueRelation {
  id: number
  source_option_value_id: number
  target_option_value_id: number
  relation_type: 'requires' | 'conflicts' | string
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

export interface ProductCategory {
  id: number
  name: string
  slug: string
}

export interface ProductDisplayPrice {
  amount_decimal: string
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
  carbon_rim_specification?: CarbonRimProductSpecification
  wheelset_specification?: WheelsetProductSpecification
  spoke_specification?: SpokesProductSpecification
  product_type?: string
  mpn?: string
  gtin?: string
  id: number
  product_specification_template_id?: number
  product_specification_template?: ProductSpecificationTemplate
  brand?: ProductBrand | null
  product_category?: ProductCategory | null
  name: string
  slug: string
  short_description?: string
  description?: string
  sku?: string
  currency?: string
  price_decimal: string
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
  option_value_relations?: ProductOptionValueRelation[]
  spoke_repair_kit_models?: SpokeRepairKitModel[]
  review_summary?: ProductReviewSummary | null
  shipping_details?: ProductShippingDetails | null
  breadcrumb?: ProductBreadcrumb | null
}

export interface SpokeRepairKitModel {
  id: number
  brand_slug: string
  brand_name: string
  wheelset_model_slug: string
  wheelset_model_name: string
  lifecycle_status?: string
  sort_order?: number
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
  validationMessage?: string
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

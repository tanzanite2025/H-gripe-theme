export type ProductSpecTemplateDialogMode = 'create' | 'edit'
export type ProductSpecFieldType = 'text' | 'number' | 'select' | 'boolean'
export type ProductSpecPresentation = 'text' | 'color' | 'image'
export type ProductSpecRole = 'attribute' | 'variant' | 'custom_option'
export type ProductSpecSelectionMode = 'single' | 'multiple'

export interface ProductSpecTemplateOptionItem {
  id?: number | string | null
  value_key: string
  default_label?: string | null
  color_hex?: string | null
  swatch_media_asset_id?: number | string | null
  swatch_url?: string | null
  is_enabled_by_default?: boolean
  is_default?: boolean
  default_price_delta_minor?: number | string | null
  default_price_currency?: string | null
  sort_order?: number | string | null
  revision?: number | string | null
}

export interface ProductSpecTemplateFilters {
  search: string
  status: string
}

export interface ProductSpecTemplateSpecDefinition {
  id?: number | string | null
  group?: string | null
  name?: string | null
  slug?: string | null
  field_type?: ProductSpecFieldType | string | null
  presentation?: ProductSpecPresentation | string | null
  unit?: string | null
  is_required?: boolean
  is_filterable?: boolean
  is_visible?: boolean
  role?: ProductSpecRole | string | null
  selection_mode?: ProductSpecSelectionMode | string | null
  min_selections?: number | string | null
  max_selections?: number | string | null
  sort_order?: number | string | null
  validation?: string | null
  option_items?: ProductSpecTemplateOptionItem[]
}

export interface ProductSpecTemplateRecord {
  id: number | string
  revision?: number | string | null
  name?: string | null
  slug?: string | null
  description?: string | null
  sort_order?: number | string | null
  is_enabled?: boolean
  is_system_managed?: boolean
  updated_at?: string | null
  spec_definitions?: ProductSpecTemplateSpecDefinition[]
}

export interface ProductSpecTemplateSpecForm {
  id: number | string
  clientKey: number
  group: string
  name: string
  slug: string
  field_type: ProductSpecFieldType
  presentation: ProductSpecPresentation
  unit: string
  is_required: boolean
  is_filterable: boolean
  is_visible: boolean
  role: ProductSpecRole
  selection_mode: ProductSpecSelectionMode
  min_selections: number
  max_selections: number | null
  sort_order: number
  optionsText: string
  validation: string
  option_items: ProductSpecTemplateOptionItem[]
}

export interface ProductSpecTemplateForm {
  id: number | string | null
  is_system_managed: boolean
  name: string
  slug: string
  description: string
  sort_order: number
  is_enabled: boolean
  spec_definitions: ProductSpecTemplateSpecForm[]
}

export interface ProductSpecTemplateSpecPayload {
  id: number
  group: string
  name: string
  slug: string
  field_type: ProductSpecFieldType
  presentation: ProductSpecPresentation
  unit: string
  is_required: boolean
  is_filterable: boolean
  is_visible: boolean
  role: ProductSpecRole
  selection_mode: ProductSpecSelectionMode
  min_selections: number
  max_selections: number | null
  sort_order: number
  validation: string
  option_items: ProductSpecTemplateOptionItem[]
}

export interface ProductSpecTemplatePayload {
  name: string
  slug: string
  description: string
  sort_order: number
  is_enabled: boolean
  spec_definitions: ProductSpecTemplateSpecPayload[]
}

export type ProductSpecTemplateFormErrors = Record<string, string>
export type ProductSpecTemplateDateFormatter = (value?: string | null) => string
export type ProductSpecTemplateVariantSpecCounter = (type: ProductSpecTemplateRecord) => number
export type ProductSpecificSpecPredicate = (spec: ProductSpecTemplateSpecForm) => boolean

export type ProductTypeDialogMode = 'create' | 'edit'
export type ProductSpecFieldType = 'text' | 'number' | 'select' | 'boolean'
export type ProductSpecPresentation = 'text' | 'color' | 'image'
export const PRODUCT_TYPE_IMAGE_SIZE = 800

export interface ProductTypeFilters {
  search: string
  status: string
}

export interface ProductTypeSpecDefinition {
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
  is_variant_option?: boolean
  sort_order?: number | string | null
  options?: string | null
  validation?: string | null
}

export interface ProductTypeTranslation {
  id?: number | string | null
  locale: string
  name: string
  description?: string | null
}

export interface ProductTypeTranslationForm {
  id?: number | string | null
  locale: string
  name: string
  description: string
}

export interface ProductTypeRecord {
  id: number | string
  name?: string | null
  slug?: string | null
  description?: string | null
  image_media_asset_id?: number | string | null
  image_url?: string | null
  sort_order?: number | string | null
  is_enabled?: boolean
  updated_at?: string | null
  translations?: ProductTypeTranslation[]
  spec_definitions?: ProductTypeSpecDefinition[]
}

export interface ProductTypeSpecForm {
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
  is_variant_option: boolean
  sort_order: number
  options?: string | null
  optionsText: string
  validation: string
}

export interface ProductTypeForm {
  id: number | string | null
  name: string
  slug: string
  description: string
  image_media_asset_id: number | string | null
  image_url: string
  pending_image_file: File | null
  remove_image: boolean
  sort_order: number
  is_enabled: boolean
  translations: ProductTypeTranslationForm[]
  spec_definitions: ProductTypeSpecForm[]
}

export interface ProductTypeSpecPayload {
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
  is_variant_option: boolean
  sort_order: number
  options: string
  validation: string
}

export interface ProductTypePayload {
  name: string
  slug: string
  description: string
  sort_order: number
  is_enabled: boolean
  translations: ProductTypeTranslation[]
  spec_definitions: ProductTypeSpecPayload[]
}

export type ProductTypeFormErrors = Record<string, string>
export type ProductTypeDateFormatter = (value?: string | null) => string
export type ProductTypeVariantSpecCounter = (type: ProductTypeRecord) => number
export type ProductSpecificSpecPredicate = (spec: ProductTypeSpecForm) => boolean

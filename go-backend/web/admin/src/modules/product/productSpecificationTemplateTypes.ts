export type ProductSpecTemplateDialogMode = 'create' | 'edit'
export type ProductSpecFieldType = 'text' | 'number' | 'select' | 'boolean'
export type ProductSpecPresentation = 'text' | 'color' | 'image'

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
  is_variant_option?: boolean
  sort_order?: number | string | null
  options?: string | null
  validation?: string | null
}

export interface ProductSpecTemplateRecord {
  id: number | string
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
  is_variant_option: boolean
  sort_order: number
  options?: string | null
  optionsText: string
  validation: string
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
  is_variant_option: boolean
  sort_order: number
  options: string
  validation: string
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

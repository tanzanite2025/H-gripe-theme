import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArray,
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export interface ProductCategoryRecord {
  id: number
  parent_id: number | null
  name: string
  slug: string
  description: string
  image_media_asset_id?: number | null
  image_url?: string | null
  depth: number
  sort_order: number
  is_enabled: boolean
  created_at?: string
  updated_at?: string
  children: ProductCategoryRecord[]
  translation_completed?: number
  translation_total?: number
  translation_missing_locales?: string[]
}

export interface ProductCategoryTranslationRecord {
  id: number
  product_category_id: number
  locale: string
  name: string
  description: string
  created_at?: string
  updated_at?: string
}

export interface ProductCategoryListPayload {
  tree: ProductCategoryRecord[]
  flat: ProductCategoryRecord[]
  max_depth: number
}

const readCategory = (response: unknown, endpoint: string): ProductCategoryRecord => {
  const category = requireApiObject(unwrapApiPayload(response, endpoint), endpoint) as unknown as ProductCategoryRecord
  requireApiNumberField(category, 'id', endpoint)
  requireApiStringField(category, 'name', endpoint)
  requireApiStringField(category, 'slug', endpoint)
  requireApiNumberField(category, 'depth', endpoint)
  requireApiNumberField(category, 'sort_order', endpoint)
  requireApiBooleanField(category, 'is_enabled', endpoint)
  if (!Array.isArray(category.children)) category.children = []
  return category
}

const readTranslations = (response: unknown, endpoint: string): ProductCategoryTranslationRecord[] => {
  const translations = requireApiArray<ProductCategoryTranslationRecord>(
    unwrapApiPayload(response, endpoint),
    endpoint,
    'translations',
  )
  translations.forEach((translation) => {
    requireApiNumberField(translation, 'id', endpoint)
    requireApiNumberField(translation, 'product_category_id', endpoint)
    requireApiStringField(translation, 'locale', endpoint)
    requireApiStringField(translation, 'name', endpoint)
    requireApiStringField(translation, 'description', endpoint)
  })
  return translations
}

export const productCategoryApi = {
  async list(params: Record<string, any> = {}): Promise<ProductCategoryListPayload> {
    const endpoint = '/api/admin/product-categories'
    const payload = requireApiObject(unwrapApiPayload(await axios.get(endpoint, { params }), endpoint), endpoint)
    const tree = requireApiArrayField<ProductCategoryRecord>(payload, 'tree', endpoint)
    const flat = requireApiArrayField<ProductCategoryRecord>(payload, 'flat', endpoint)
    requireApiNumberField(payload, 'max_depth', endpoint)
    return { tree, flat, max_depth: Number(payload.max_depth || 5) }
  },

  async get(id: number | string): Promise<ProductCategoryRecord> {
    const endpoint = `/api/admin/product-categories/${id}`
    return readCategory(await axios.get(endpoint), endpoint)
  },

  async translations(id: number | string): Promise<ProductCategoryTranslationRecord[]> {
    const endpoint = `/api/admin/product-categories/${id}/translations`
    return readTranslations(await axios.get(endpoint), endpoint)
  },

  async updateTranslations(
    id: number | string,
    translations: Array<Pick<ProductCategoryTranslationRecord, 'locale' | 'name' | 'description'>>,
  ): Promise<ProductCategoryTranslationRecord[]> {
    const endpoint = `/api/admin/product-categories/${id}/translations`
    return readTranslations(await axios.put(endpoint, { translations }), endpoint)
  },

  async create(payload: Record<string, any>): Promise<ProductCategoryRecord> {
    const endpoint = '/api/admin/product-categories'
    return readCategory(await axios.post(endpoint, payload), endpoint)
  },

  async update(id: number | string, payload: Record<string, any>): Promise<ProductCategoryRecord> {
    const endpoint = `/api/admin/product-categories/${id}`
    return readCategory(await axios.put(endpoint, payload), endpoint)
  },

  async remove(id: number | string) {
    const endpoint = `/api/admin/product-categories/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default productCategoryApi

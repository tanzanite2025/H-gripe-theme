import type {
  ProductCategoryRecord,
  ProductCategoryTranslationRecord,
} from '@/api/productCategories'

export interface ProductCategoryTranslationForm {
  id: number | null
  locale: string
  name: string
  description: string
}

export type ProductCategoryTranslationPayload = Pick<
  ProductCategoryTranslationRecord,
  'locale' | 'name' | 'description'
>

export interface DraftCategoryRow {
  key: string
  id: number | null
  parent_key: string | null
  name: string
  slug: string
  description: string
  image_media_asset_id: number | null
  image_url: string
  depth: number
  sort_order: number
  is_enabled: boolean
  translation_completed: number
  translation_total: number
  translation_missing_locales: string[]
  is_new: boolean
  dirty: boolean
}

export interface ProductCategoryParentOption extends DraftCategoryRow {
  disabled: boolean
}

export interface ProductCategoryStats {
  total: number
  enabled: number
  changed: number
  translationIncomplete: number
  deepestLevel: number
  maxDepth: number
}

export const categoryToDraftRow = (category: ProductCategoryRecord): DraftCategoryRow => ({
  key: `id:${category.id}`,
  id: category.id,
  parent_key: category.parent_id ? `id:${category.parent_id}` : null,
  name: category.name || '',
  slug: category.slug || '',
  description: category.description || '',
  image_media_asset_id: category.image_media_asset_id || null,
  image_url: category.image_url || '',
  depth: Number(category.depth || 1),
  sort_order: Number(category.sort_order || 0),
  is_enabled: category.is_enabled !== false,
  translation_completed: Number(category.translation_completed || 0),
  translation_total: Number(category.translation_total || 0),
  translation_missing_locales: Array.isArray(category.translation_missing_locales)
    ? category.translation_missing_locales
    : [],
  is_new: false,
  dirty: false,
})

export const flattenProductCategoryTree = (
  items: ProductCategoryRecord[],
  result: DraftCategoryRow[] = [],
): DraftCategoryRow[] => {
  items.forEach((item) => {
    result.push(categoryToDraftRow(item))
    if (item.children?.length) flattenProductCategoryTree(item.children, result)
  })
  return result
}

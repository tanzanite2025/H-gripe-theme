import type { ProductCategoryRecord } from '@/api/productCategories'

export const wheelsetSystemCategorySlug = 'wheelset'

export const requireWheelsetSystemCategory = (categories: ProductCategoryRecord[]): ProductCategoryRecord => {
  const category = categories.find(item => item.slug === wheelsetSystemCategorySlug)
  if (!category || category.id <= 0 || !category.is_enabled) {
    throw new Error('WHEELSET 系统分类不可用，请检查商品分类配置')
  }
  return category
}

export const buildWheelsetSystemCategoryDisplayPath = (categories: ProductCategoryRecord[]): string => {
  const category = requireWheelsetSystemCategory(categories)
  const names = [category.name]
  let parentID = category.parent_id
  const visited = new Set<number>([category.id])
  while (parentID != null) {
    if (visited.has(parentID)) throw new Error('WHEELSET 系统分类路径存在循环')
    visited.add(parentID)
    const parent = categories.find(item => item.id === parentID)
    if (!parent) break
    names.unshift(parent.name)
    parentID = parent.parent_id
  }
  return names.join(' → ')
}

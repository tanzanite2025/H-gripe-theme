import type { ProductCategoryRecord } from '@/api/productCategories'

export const spokesSystemCategorySlug = 'spokes'

export const requireSpokesSystemCategory = (categories: ProductCategoryRecord[]): ProductCategoryRecord => {
  const category = categories.find(item => item.slug === spokesSystemCategorySlug)
  if (!category || category.id <= 0 || !category.is_enabled) {
    throw new Error('辐条系统分类不可用，请检查商品分类配置')
  }
  return category
}

export const buildSpokesSystemCategoryDisplayPath = (categories: ProductCategoryRecord[]): string => {
  const category = requireSpokesSystemCategory(categories)
  const names = [category.name]
  let parentID = category.parent_id
  const visited = new Set<number>([category.id])
  while (parentID != null) {
    if (visited.has(parentID)) throw new Error('辐条系统分类路径存在循环')
    visited.add(parentID)
    const parent = categories.find(item => item.id === parentID)
    if (!parent) break
    names.unshift(parent.name)
    parentID = parent.parent_id
  }
  return names.join(' → ')
}

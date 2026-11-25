import type { ProductCategoryRecord } from '@/api/productCategories'

export const bicycleTireSystemCategorySlug = 'tire'

export const requireBicycleTireSystemCategory = (categories: ProductCategoryRecord[]): ProductCategoryRecord => {
  const category = categories.find(item => item.slug === bicycleTireSystemCategorySlug)
  const parent = categories.find(item => item.id === category?.parent_id)
  if (!category || category.id <= 0 || !category.is_enabled || category.depth !== 2
    || !parent || parent.slug !== 'wheel-components' || !parent.is_enabled
    || parent.parent_id != null || parent.depth !== 1) {
    throw new Error('外胎系统分类不可用，请检查 Wheel Components → Tires 分类配置')
  }
  return category
}

export const buildBicycleTireSystemCategoryDisplayPath = (categories: ProductCategoryRecord[]): string => {
  const category = requireBicycleTireSystemCategory(categories)
  const parent = categories.find(item => item.id === category.parent_id)!
  return `${parent.name} → ${category.name}`
}

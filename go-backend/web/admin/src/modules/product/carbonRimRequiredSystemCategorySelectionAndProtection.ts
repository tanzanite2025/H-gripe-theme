import type { ProductCategoryRecord } from '@/api/productCategories'

export const carbonRimSystemCategorySlug = 'carbon-rim'

export const requireCarbonRimSystemCategory = (categories: ProductCategoryRecord[]): ProductCategoryRecord => {
  const category = categories.find(item => item.slug === carbonRimSystemCategorySlug)
  const parent = categories.find(item => item.id === category?.parent_id)
  if (!category || category.id <= 0 || !category.is_enabled || category.depth !== 2
    || !parent || parent.slug !== 'wheel-components' || !parent.is_enabled
    || parent.parent_id != null || parent.depth !== 1) {
    throw new Error('碳圈系统分类不可用，请检查数据库迁移和 Wheel Components → Carbon Rim 分类配置')
  }
  return category
}

export const isProtectedSystemProductCategory = (slug: string): boolean => (
  ['wheelset', 'wheel-components', 'tire', 'spoke-repair-kits', carbonRimSystemCategorySlug].includes(slug)
)

export const hasProtectedSystemProductCategoryParent = (slug: string): boolean => (
  isProtectedSystemProductCategory(slug) && slug !== 'wheelset'
)

import { describe, expect, it } from 'vitest'
import type { ProductCategoryRecord } from '@/api/productCategories'
import { buildBicycleTireSystemCategoryDisplayPath, requireBicycleTireSystemCategory } from './bicycleTireRequiredSystemCategorySelection'

const createTireSystemCategories = (): ProductCategoryRecord[] => [
  { id: 9, parent_id: null, name: 'Wheel Components', slug: 'wheel-components', description: '', depth: 1, sort_order: 20, is_enabled: true, children: [] },
  { id: 10, parent_id: 9, name: 'Tires', slug: 'tire', description: '', depth: 2, sort_order: 10, is_enabled: true, children: [] },
]

describe('required bicycle tire system category', () => {
  it('finds the tire category from its stable slug and builds its current label path', () => {
    const categories = createTireSystemCategories().reverse()
    expect(requireBicycleTireSystemCategory(categories).id).toBe(10)
    expect(buildBicycleTireSystemCategoryDisplayPath(categories)).toBe('Wheel Components → Tires')
  })

  it.each(['missing', 'disabled', 'wrong-parent', 'wrong-depth'])('rejects %s category configuration', scenario => {
    const categories = createTireSystemCategories()
    if (scenario === 'missing') categories.pop()
    if (scenario === 'disabled') categories[1]!.is_enabled = false
    if (scenario === 'wrong-parent') categories[1]!.parent_id = 99
    if (scenario === 'wrong-depth') categories[1]!.depth = 3
    expect(() => requireBicycleTireSystemCategory(categories)).toThrow('外胎系统分类不可用')
  })
})

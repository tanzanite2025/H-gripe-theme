import { describe, expect, it } from 'vitest'
import type { ProductCategoryRecord } from '@/api/productCategories'
import { requireCarbonRimSystemCategory, buildCarbonRimRequiredSystemCategoryDisplayPath } from './carbonRimRequiredSystemCategorySelection'

const createSystemCategoryRecords = (): ProductCategoryRecord[] => [
  { id: 317, parent_id: null, name: '车轮配件', slug: 'wheel-components', description: '', depth: 1, sort_order: 0, is_enabled: true, children: [] },
  { id: 975, parent_id: 317, name: '碳纤维车圈', slug: 'carbon-rim', description: '', depth: 2, sort_order: 30, is_enabled: true, children: [] },
]

describe('required carbon rim system category', () => {
  it('selects actual IDs by slug regardless of display language or order', () => {
    expect(requireCarbonRimSystemCategory(createSystemCategoryRecords().reverse()).id).toBe(975)
  })

  it.each(['missing', 'disabled', 'wrong-parent', 'disabled-parent', 'wrong-depth', 'name-only'])('rejects %s configuration', (scenario) => {
    const records = createSystemCategoryRecords()
    if (scenario === 'missing') records.pop()
    if (scenario === 'disabled') records[1]!.is_enabled = false
    if (scenario === 'wrong-parent') records[1]!.parent_id = 999
    if (scenario === 'disabled-parent') records[0]!.is_enabled = false
    if (scenario === 'wrong-depth') records[1]!.depth = 3
    if (scenario === 'name-only') { records[1]!.slug = 'manually-created'; records[1]!.name = 'Carbon Rim' }
    expect(() => requireCarbonRimSystemCategory(records)).toThrow('碳圈系统分类不可用')
  })

  it('uses editable category names from the same records that determine the binding', () => {
    const records = createSystemCategoryRecords()
    expect(buildCarbonRimRequiredSystemCategoryDisplayPath(records)).toBe('车轮配件 → 碳纤维车圈')
    records[1]!.name = 'Carbon Fiber Rims'
    expect(buildCarbonRimRequiredSystemCategoryDisplayPath(records)).toBe('车轮配件 → Carbon Fiber Rims')
  })
})

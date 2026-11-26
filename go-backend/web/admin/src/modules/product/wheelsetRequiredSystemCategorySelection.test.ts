import { describe, expect, it } from 'vitest'
import { buildWheelsetSystemCategoryDisplayPath, requireWheelsetSystemCategory } from './wheelsetRequiredSystemCategorySelection'

describe('WHEELSET system category selection', () => {
  it('accepts an enabled WHEELSET category whether or not it is a root category', () => {
    const categories = [
      { id: 2, name: 'Complete Wheels', slug: 'complete-wheels', description: '', parent_id: null, depth: 1, sort_order: 0, is_enabled: true, children: [] },
      { id: 8, name: 'WHEELSET', slug: 'wheelset', description: '', parent_id: 2, depth: 2, sort_order: 0, is_enabled: true, children: [] },
    ]
    expect(requireWheelsetSystemCategory(categories).id).toBe(8)
    expect(buildWheelsetSystemCategoryDisplayPath(categories)).toBe('Complete Wheels → WHEELSET')
  })

  it('rejects missing and disabled WHEELSET categories', () => {
    expect(() => requireWheelsetSystemCategory([])).toThrow('系统分类不可用')
    expect(() => requireWheelsetSystemCategory([
      { id: 8, name: 'WHEELSET', slug: 'wheelset', description: '', parent_id: null, depth: 1, sort_order: 0, is_enabled: false, children: [] },
    ])).toThrow('系统分类不可用')
  })
})

import { describe, expect, it } from 'vitest'
import { isProtectedSystemProductCategory, hasProtectedSystemProductCategoryParent } from './productSystemCategoryIdentityAndParentProtection'

describe('system product category identity and parent protection', () => {
  it.each(['wheelset', 'wheel-components', 'tire', 'spoke-repair-kits', 'carbon-rim'])('protects the %s identity', (slug) => {
    expect(isProtectedSystemProductCategory(slug)).toBe(true)
    expect(hasProtectedSystemProductCategoryParent(slug)).toBe(slug !== 'wheelset')
  })

  it('allows custom category identity and parent editing', () => {
    expect(isProtectedSystemProductCategory('custom')).toBe(false)
    expect(hasProtectedSystemProductCategoryParent('custom')).toBe(false)
  })
})

import { describe, expect, it } from 'vitest'
import {
  getCustomsReviewLifecycleStatus,
  getCustomsReviewLifecycleStatusLabel,
} from './customsReviewLifecycle'

describe('customs review lifecycle status', () => {
  const referenceDate = new Date(2026, 9, 5)

  it('classifies healthy, expiring soon, and expired review dates at the 30-day boundary', () => {
    expect(getCustomsReviewLifecycleStatus('2026-01-01', '2026-11-05', referenceDate)).toBe('healthy')
    expect(getCustomsReviewLifecycleStatus('2026-01-01', '2026-11-04', referenceDate)).toBe('expiring_soon')
    expect(getCustomsReviewLifecycleStatus('2026-01-01', '2026-10-05', referenceDate)).toBe('expiring_soon')
    expect(getCustomsReviewLifecycleStatus('2026-01-01', '2026-10-04', referenceDate)).toBe('expired')
  })

  it('marks the template as pending when either lifecycle date is missing or invalid', () => {
    expect(getCustomsReviewLifecycleStatus(null, '2026-11-04', referenceDate)).toBe('pending')
    expect(getCustomsReviewLifecycleStatus('2026-01-01', null, referenceDate)).toBe('pending')
    expect(getCustomsReviewLifecycleStatus('2026-01-01', '2026-02-30', referenceDate)).toBe('pending')
  })

  it('provides the visible Chinese status labels', () => {
    expect(getCustomsReviewLifecycleStatusLabel('healthy')).toBe('健康')
    expect(getCustomsReviewLifecycleStatusLabel('expiring_soon')).toBe('临期复核')
    expect(getCustomsReviewLifecycleStatusLabel('expired')).toBe('已过期')
    expect(getCustomsReviewLifecycleStatusLabel('pending')).toBe('待核验')
  })
})

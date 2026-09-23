import { describe, expect, it } from 'vitest'
import { isCustomerServiceWithinQuietHours } from './useCustomerServiceRealtime'

describe('customer-service quiet hours', () => {
  it('handles a normal daytime interval', () => {
    expect(isCustomerServiceWithinQuietHours(new Date(2026, 8, 21, 12, 0), true, '09:00', '18:00')).toBe(true)
    expect(isCustomerServiceWithinQuietHours(new Date(2026, 8, 21, 18, 0), true, '09:00', '18:00')).toBe(false)
  })

  it('handles an interval crossing midnight', () => {
    expect(isCustomerServiceWithinQuietHours(new Date(2026, 8, 21, 23, 30), true, '22:00', '08:00')).toBe(true)
    expect(isCustomerServiceWithinQuietHours(new Date(2026, 8, 21, 7, 59), true, '22:00', '08:00')).toBe(true)
    expect(isCustomerServiceWithinQuietHours(new Date(2026, 8, 21, 8, 0), true, '22:00', '08:00')).toBe(false)
  })

  it('does not silence when disabled or invalid', () => {
    expect(isCustomerServiceWithinQuietHours(new Date(), false, '22:00', '08:00')).toBe(false)
    expect(isCustomerServiceWithinQuietHours(new Date(), true, 'bad', '08:00')).toBe(false)
  })
})

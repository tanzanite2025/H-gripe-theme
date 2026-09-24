import { describe, expect, it } from 'vitest'
import { majorToMinor, minorToMajorDecimal } from './marketingPresentation'

describe('Merchant commercial-money presentation', () => {
  it('parses two-decimal commercial currencies exactly', () => {
    expect(majorToMinor('1', 'USD')).toBe(100)
    expect(majorToMinor('1.20', 'USD')).toBe(120)
    expect(majorToMinor('1.2300', 'USD')).toBe(123)
    expect(majorToMinor('1.001', 'USD')).toBeNull()
    expect(majorToMinor('1.005', 'USD')).toBeNull()
    expect(majorToMinor('1.23', 'ISK')).toBe(123)
  })

  it('accepts only whole amounts for zero-decimal commercial currencies', () => {
    expect(majorToMinor('123', 'JPY')).toBe(123)
    expect(majorToMinor('123.0', 'JPY')).toBe(123)
    expect(majorToMinor('123.1', 'JPY')).toBeNull()
  })

  it('rejects ambiguous or unsafe decimal input', () => {
    for (const value of ['0x10', '1e3', '.5', 'Infinity', '90071992547409.92']) {
      expect(majorToMinor(value, 'USD')).toBeNull()
    }
  })

  it('formats integer minor values without floating-point precision loss', () => {
    expect(minorToMajorDecimal('9007199254740991', 'USD')).toBe('90071992547409.91')
    expect(minorToMajorDecimal('123', 'JPY')).toBe('123')
  })
})

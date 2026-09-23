import { describe, expect, it } from 'vitest'
import { customerTimezoneDifference } from './customerServicePresentation'

describe('customerTimezoneDifference', () => {
  it('uses the current IANA offset and reports the customer lead', () => {
    const date = new Date('2026-01-15T00:00:00.000Z')

    expect(customerTimezoneDifference(date, 'Asia/Tokyo', 'Asia/Shanghai')).toBe('客户比客服快 1 小时')
  })

  it('accounts for daylight saving time at the supplied instant', () => {
    const date = new Date('2026-07-15T12:00:00.000Z')

    expect(customerTimezoneDifference(date, 'America/New_York', 'Europe/London')).toBe('客户比客服慢 5 小时')
  })

  it('does not fabricate a difference when either timezone is invalid', () => {
    const date = new Date('2026-01-15T00:00:00.000Z')

    expect(customerTimezoneDifference(date, 'not/a-timezone', 'Asia/Shanghai')).toBe('')
  })
})

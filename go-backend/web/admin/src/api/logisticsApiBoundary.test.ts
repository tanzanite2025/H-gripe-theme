import { describe, expect, it } from 'vitest'
import { fpxLogisticsAdminApi } from './fpxLogisticsAdminApi'
import { shippingServiceCollectionReferenceApi } from './shippingServiceCollectionReferenceApi'
import { shippingApi } from './shipping'
import { yanwenLogisticsAdminApi } from './yanwenLogisticsAdminApi'

describe('logistics API module boundaries', () => {
  it('keeps generic shipping API free of carrier-domain operations', () => {
    expect(Object.keys(shippingApi)).not.toEqual(expect.arrayContaining([
      'getFpxApiConfig',
      'listFpxChannels',
      'getYanwenApiConfig',
      'listYanwenWaybills',
    ]))
  })

  it('keeps full carrier operations out of the cross-domain collection reference API', () => {
    expect(Object.keys(shippingServiceCollectionReferenceApi)).toEqual([
      'listFpxPublishedCollection',
      'listYanwenPublishedCollection',
    ])
  })

  it('keeps 4PX and Yanwen admin operations in separate modules', () => {
    expect(Object.keys(fpxLogisticsAdminApi)).not.toContain('listYanwenWaybills')
    expect(Object.keys(yanwenLogisticsAdminApi)).not.toContain('listFpxChannels')
    expect(Object.keys(yanwenLogisticsAdminApi)).not.toContain('listFpxPublishedCollection')
  })
})

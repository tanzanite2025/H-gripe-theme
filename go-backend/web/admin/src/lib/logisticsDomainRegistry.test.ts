import { describe, expect, it } from 'vitest'
import { fpxLogisticsTabs, logisticsDomains, yanwenLogisticsTabs } from './logisticsDomainRegistry'

describe('logistics domain registry', () => {
  it('keeps tab identities and routes unique within each domain', () => {
    for (const domain of Object.values(logisticsDomains)) {
      const keys = domain.tabs.map((tab) => tab.key)
      const routeNames = domain.tabs.map((tab) => tab.routeName)
      const paths = domain.tabs.map((tab) => tab.path)

      expect(new Set(keys).size).toBe(keys.length)
      expect(new Set(routeNames).size).toBe(routeNames.length)
      expect(new Set(paths).size).toBe(paths.length)
      expect(domain.tabs[0]?.routeName).toBe(domain.defaultRouteName)
      expect(domain.tabs.every((tab) => tab.path.startsWith(`${domain.path}/`))).toBe(true)
    }
  })

  it('keeps configuration tabs behind manage permissions', () => {
    expect(fpxLogisticsTabs[fpxLogisticsTabs.length - 1]?.permission).toBe('logistics:fpx:manage')
    expect(yanwenLogisticsTabs[yanwenLogisticsTabs.length - 1]?.permission).toBe('logistics:yanwen:manage')
  })

  it('keeps 4PX limited to shipping tabs', () => {
    expect(fpxLogisticsTabs.map((tab) => tab.key)).toEqual([
      'overview',
      'direct',
      'collection',
      'calculator',
      'tracking',
      'rma',
      'config',
    ])
    expect(JSON.stringify(fpxLogisticsTabs)).not.toMatch(/warehouse|wms|fb4/i)
  })
})

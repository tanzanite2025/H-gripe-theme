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

  it('keeps 4PX read pages view-only and gateway configuration behind manage permissions', () => {
    expect(fpxLogisticsTabs.filter((tab) => tab.key !== 'config').every((tab) => tab.permission === 'logistics:fpx:view')).toBe(true)
    expect(fpxLogisticsTabs[fpxLogisticsTabs.length - 1]?.permission).toBe('logistics:fpx:manage')
    expect(yanwenLogisticsTabs[yanwenLogisticsTabs.length - 1]?.permission).toBe('logistics:yanwen:manage')
  })

  it('keeps 4PX limited to the catalog and gateway tabs', () => {
    expect(fpxLogisticsTabs.map((tab) => tab.key)).toEqual([
      'overview',
      'collection',
      'config',
    ])
    expect(JSON.stringify(fpxLogisticsTabs)).not.toMatch(/warehouse|wms|fb4/i)
  })
})

import { expect, test } from '@playwright/test'
import { DEFAULT_SPOKE_CATALOG } from '../app/data/spoke-calculator/database'
import { normalizeSpokeCatalogPayload } from '../app/utils/spokeCatalogNormalizer'

test('normalizes the public spoke catalog payload the frontend consumes', () => {
  const frontLeft = 282
  const rearRight = 284
  const payload = {
    ...DEFAULT_SPOKE_CATALOG,
    presets: DEFAULT_SPOKE_CATALOG.presets.map((preset, index) => {
      if (index !== 0) {
        return { ...preset }
      }

      return {
        ...preset,
        actualLengths: {
          frontLeft,
          frontRight: null,
          rearLeft: null,
          rearRight,
          notes: 'verified bench build',
        },
      }
    }),
  }

  const catalog = normalizeSpokeCatalogPayload(payload)

  expect(catalog.options.spokeCounts.length).toBeGreaterThan(0)
  expect(catalog.presets).toHaveLength(DEFAULT_SPOKE_CATALOG.presets.length)
  expect(catalog.presets[0].wheelPosition).toBe('auto')
  expect(catalog.presets[0].description).toBe(DEFAULT_SPOKE_CATALOG.presets[0].description)
  expect(catalog.presets[0].keywords).toEqual(DEFAULT_SPOKE_CATALOG.presets[0].keywords)
  expect(catalog.presets[0].actualLengths?.frontLeft).toBe(frontLeft)
  expect(catalog.presets[0].actualLengths?.frontRight).toBeNull()
  expect(catalog.presets[0].actualLengths?.rearRight).toBe(rearRight)
})

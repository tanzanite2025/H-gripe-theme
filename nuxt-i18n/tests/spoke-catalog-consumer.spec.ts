import { expect, test } from '@playwright/test'
import { DEFAULT_SPOKE_CATALOG } from '../app/data/spoke-calculator/database'
import {
  normalizeSpokeCatalogPayload,
  normalizeSpokeRecordedResultsPayload,
} from '../app/utils/spokeCatalogNormalizer'

test('normalizes the public spoke catalog payload without recorded measurements', () => {
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
  expect('actualLengths' in catalog.presets[0]).toBe(false)
})

test('normalizes the separate recorded-result projection and drops internal notes', () => {
  const results = normalizeSpokeRecordedResultsPayload({
    presets: [{
      ...DEFAULT_SPOKE_CATALOG.presets[0],
      actualLengths: {
        frontLeft: 282,
        frontRight: null,
        rearLeft: null,
        rearRight: 284,
        notes: 'internal import note',
      },
    }, {
      ...DEFAULT_SPOKE_CATALOG.presets[1],
      actualLengths: { frontLeft: null, frontRight: null, rearLeft: null, rearRight: null },
    }],
  })

  expect(results.presets).toHaveLength(1)
  expect(results.presets[0].actualLengths.frontLeft).toBe(282)
  expect(results.presets[0].actualLengths.rearRight).toBe(284)
  expect('notes' in results.presets[0].actualLengths).toBe(false)
})

import assert from 'node:assert/strict'
import {
  parseSchwalbeTireCatalogEtrtoDimensions,
} from '../app/data/tireguides/schwalbeTireCatalogDimensionNormalization'

assert.deepEqual(
  parseSchwalbeTireCatalogEtrtoDimensions('50-622'),
  {
    nominalTireWidthMm: 50,
    beadSeatDiameterMm: 622,
  },
)

assert.deepEqual(
  parseSchwalbeTireCatalogEtrtoDimensions('  37-584 '),
  {
    nominalTireWidthMm: 37,
    beadSeatDiameterMm: 584,
  },
)

for (const invalidEtrto of ['', '50 / 622', '50-622-extra', 'Pro One 50-622']) {
  assert.equal(
    parseSchwalbeTireCatalogEtrtoDimensions(invalidEtrto),
    null,
    `Expected ${invalidEtrto || '(empty)'} to be rejected`,
  )
}

console.log('Schwalbe catalog dimension normalization checks passed.')

import assert from 'node:assert/strict'
import {
  mergeSchwalbeTireCatalogFilterQuery,
  parseSchwalbeTireCatalogFilterQuery,
} from '../app/data/tireguides/schwalbeTireCatalogFilterQuery.js'

const parsed = parseSchwalbeTireCatalogFilterQuery({
  search: 'Kojak',
  page: '4',
  model: 'Kojak',
  tire_width_mm: ['35', '55', '35', '0', '-1', '1.5', 'invalid'],
  bead_seat_diameter_mm: '559',
  sort: 'etrto',
})

assert.deepEqual(parsed, {
  modelName: 'Kojak',
  nominalTireWidthsMm: [35, 55],
  beadSeatDiametersMm: [559],
  sortBy: 'etrto',
})

assert.deepEqual(
  mergeSchwalbeTireCatalogFilterQuery(
    { search: 'Kojak', page: '4', keep: 'value' },
    parsed,
  ),
  {
    search: 'Kojak',
    page: '4',
    keep: 'value',
    model: 'Kojak',
    tire_width_mm: ['35', '55'],
    bead_seat_diameter_mm: ['559'],
    sort: 'etrto',
  },
)

assert.deepEqual(
  parseSchwalbeTireCatalogFilterQuery({
    model: ['Kojak', 'Marathon'],
    tire_width_mm: ['9007199254740992', '26'],
    bead_seat_diameter_mm: ['0'],
    sort: 'unknown',
  }),
  {
    modelName: 'Kojak',
    nominalTireWidthsMm: [26],
    beadSeatDiametersMm: [],
    sortBy: 'model',
  },
)

assert.deepEqual(
  mergeSchwalbeTireCatalogFilterQuery(
    { model: 'Kojak', tire_width_mm: ['35'], bead_seat_diameter_mm: ['559'], sort: 'etrto' },
    { modelName: null, nominalTireWidthsMm: [], beadSeatDiametersMm: [], sortBy: 'model' },
  ),
  {},
)

console.log('Schwalbe tire catalog filter query checks passed.')

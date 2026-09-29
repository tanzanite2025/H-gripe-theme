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
  bead: ['WIRED', 'Folding', 'WIRED'],
  seal: ['TLR', 'Tube', 'TLR'],
  e_bike_rating: ['E-50', 'none', 'E-25', 'none'],
  color: ['Black+Reflex', 'Black', 'Black+Reflex'],
  sort: 'etrto',
})

assert.deepEqual(parsed, {
  modelName: 'Kojak',
  nominalTireWidthsMm: [35, 55],
  beadSeatDiametersMm: [559],
  beads: ['Folding', 'WIRED'],
  seals: ['TLR', 'Tube'],
  eBikeRatings: ['E-25', 'E-50', null],
  colors: ['Black', 'Black+Reflex'],
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
    bead: ['Folding', 'WIRED'],
    seal: ['TLR', 'Tube'],
    e_bike_rating: ['E-25', 'E-50', 'none'],
    color: ['Black', 'Black+Reflex'],
    sort: 'etrto',
  },
)

assert.deepEqual(
  parseSchwalbeTireCatalogFilterQuery({
    model: ['Kojak', 'Marathon'],
    tire_width_mm: ['9007199254740992', '26'],
    bead_seat_diameter_mm: ['0'],
    bead: [' WIRED ', '', 'Folding'],
    seal: ' TLR ',
    e_bike_rating: ['none', 'E-25'],
    color: ['Black+Reflex', 'Black'],
    sort: 'unknown',
  }),
  {
    modelName: 'Kojak',
    nominalTireWidthsMm: [26],
    beadSeatDiametersMm: [],
    beads: ['Folding', 'WIRED'],
    seals: ['TLR'],
    eBikeRatings: ['E-25', null],
    colors: ['Black', 'Black+Reflex'],
    sortBy: 'model',
  },
)

assert.deepEqual(
  mergeSchwalbeTireCatalogFilterQuery(
    {
      model: 'Kojak',
      tire_width_mm: ['35'],
      bead_seat_diameter_mm: ['559'],
      bead: ['WIRED'],
      seal: ['TLR'],
      e_bike_rating: ['none'],
      color: ['Black+Reflex'],
      sort: 'etrto',
    },
    {
      modelName: null,
      nominalTireWidthsMm: [],
      beadSeatDiametersMm: [],
      beads: [],
      seals: [],
      eBikeRatings: [],
      colors: [],
      sortBy: 'model',
    },
  ),
  {},
)

console.log('Schwalbe tire catalog filter query checks passed.')

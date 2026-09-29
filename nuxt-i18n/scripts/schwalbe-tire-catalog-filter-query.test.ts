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
  wheel_size: ['28-622', '26-559', '28-622', 'invalid'],
  bead_seat_diameter_mm: '559',
  min_load_kg: '90.5',
  casing: ['Super Race', 'TRAIL PRO', 'Super Race', 'not-a-casing'],
  radial: '1',
  bead: ['WIRED', 'Folding', 'WIRED'],
  seal: ['TLR', 'Tube', 'TLR'],
  e_bike_rating: ['E-50', 'none', 'E-25', 'none'],
  color: ['Black+Reflex', 'Black', 'Black+Reflex'],
  compound: ["Black'n'Roll", 'ADDIX Race', 'ADDIX Race'],
  sort: 'weight_desc',
})

assert.deepEqual(parsed, {
  modelName: 'Kojak',
  nominalTireWidthMinMm: null,
  nominalTireWidthMaxMm: null,
  nominalTireWidthsMm: [35, 55],
  wheelSizeKeys: ['26-559', '28-622'],
  beadSeatDiametersMm: [559],
  minimumLoadKg: 90.5,
  casingConstructions: ['Super Race', 'TRAIL PRO'],
  radialOnly: true,
  beads: ['Folding', 'WIRED'],
  seals: ['TLR', 'Tube'],
  eBikeRatings: ['E-25', 'E-50', null],
  colors: ['Black', 'Black+Reflex'],
  compounds: ['ADDIX Race', "Black'n'Roll"],
  sortBy: 'weight_desc',
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
    wheel_size: ['26-559', '28-622'],
    bead_seat_diameter_mm: ['559'],
    min_load_kg: '90.5',
    casing: ['Super Race', 'TRAIL PRO'],
    radial: '1',
    bead: ['Folding', 'WIRED'],
    seal: ['TLR', 'Tube'],
    e_bike_rating: ['E-25', 'E-50', 'none'],
    color: ['Black', 'Black+Reflex'],
    compound: ['ADDIX Race', "Black'n'Roll"],
    sort: 'weight_desc',
  },
)

assert.deepEqual(
  parseSchwalbeTireCatalogFilterQuery({
    model: ['Kojak', 'Marathon'],
    tire_width_mm: ['9007199254740992', '26'],
    bead_seat_diameter_mm: ['0'],
    min_load_kg: 'Infinity',
    casing: [' TRAIL PRO ', '', 'GRAVITY PRO', 'TRAIL PRO'],
    radial: 'true',
    bead: [' WIRED ', '', 'Folding'],
    seal: ' TLR ',
    e_bike_rating: ['none', 'E-25'],
    color: ['Black+Reflex', 'Black'],
    compound: ['ADDIX Race'],
    sort: 'unknown',
  }),
  {
    modelName: 'Kojak',
    nominalTireWidthMinMm: null,
    nominalTireWidthMaxMm: null,
    nominalTireWidthsMm: [26],
    wheelSizeKeys: [],
    beadSeatDiametersMm: [],
    minimumLoadKg: null,
    casingConstructions: ['GRAVITY PRO', 'TRAIL PRO'],
    radialOnly: true,
    beads: ['Folding', 'WIRED'],
    seals: ['TLR'],
    eBikeRatings: ['E-25', null],
    colors: ['Black', 'Black+Reflex'],
    compounds: ['ADDIX Race'],
    sortBy: 'weight_asc',
  },
)

assert.deepEqual(
  mergeSchwalbeTireCatalogFilterQuery(
    {
      model: 'Kojak',
      tire_width_mm: ['35'],
      bead_seat_diameter_mm: ['559'],
      min_load_kg: '90',
      casing: ['Super Race'],
      radial: '1',
      bead: ['WIRED'],
      seal: ['TLR'],
      e_bike_rating: ['none'],
      color: ['Black+Reflex'],
      compound: ["Black'n'Roll"],
      sort: 'etrto',
    },
    {
      modelName: null,
      nominalTireWidthMinMm: null,
      nominalTireWidthMaxMm: null,
      nominalTireWidthsMm: [],
      wheelSizeKeys: [],
      beadSeatDiametersMm: [],
      minimumLoadKg: null,
      casingConstructions: [],
      radialOnly: false,
      beads: [],
      seals: [],
      eBikeRatings: [],
      colors: [],
      compounds: [],
      sortBy: 'weight_asc',
    },
  ),
  {},
)

const parsedRange = parseSchwalbeTireCatalogFilterQuery({
  tire_width_min_mm: '35',
  tire_width_max_mm: '57',
  // New endpoints intentionally take precedence over old exact values.
  tire_width_mm: ['35', '100'],
})
assert.equal(parsedRange.nominalTireWidthMinMm, 35)
assert.equal(parsedRange.nominalTireWidthMaxMm, 57)
assert.deepEqual(parsedRange.nominalTireWidthsMm, [])
assert.deepEqual(
  mergeSchwalbeTireCatalogFilterQuery(
    { search: 'Kojak', tire_width_mm: ['35', '55'], page: '2' },
    parsedRange,
  ),
  {
    search: 'Kojak',
    tire_width_min_mm: '35',
    tire_width_max_mm: '57',
    page: '2',
  },
)

const reversedRange = parseSchwalbeTireCatalogFilterQuery({
  tire_width_min_mm: '57',
  tire_width_max_mm: '35',
})
assert.equal(reversedRange.nominalTireWidthMinMm, 35)
assert.equal(reversedRange.nominalTireWidthMaxMm, 57)
assert.deepEqual(
  mergeSchwalbeTireCatalogFilterQuery({}, reversedRange),
  {
    tire_width_min_mm: '35',
    tire_width_max_mm: '57',
  },
)

console.log('Schwalbe tire catalog filter query checks passed.')

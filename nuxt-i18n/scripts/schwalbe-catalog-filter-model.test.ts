import assert from 'node:assert/strict'
import type { SchwalbeTireCatalogItem } from '../app/data/tireguides/schwalbeCatalog.js'
import {
  buildSchwalbeTireCatalogFilterOptions,
  deriveSchwalbeTireCatalogFilterDimensions,
  filterSchwalbeTireCatalogItemsByInnerRimWidth,
  filterSchwalbeTireCatalogItems,
} from '../app/data/tireguides/schwalbeTireCatalogFilterModel.js'

const catalogItems: SchwalbeTireCatalogItem[] = [
  {
    article_no: '11100062.02',
    model_name: 'Kojak',
    etrto: '35-559',
    inch_designation: '26x1.35',
    load_kg: 90,
    version_label: 'RaceGuard',
    compound: 'ADDIX',
    color: 'Black',
    bead: 'WIRED',
    e_bike_rating: undefined,
    seal: 'Tube',
    source_checked_at: '2026-09-28',
    product_exists: false,
  },
  {
    article_no: '11654751',
    model_name: 'Marathon PRO Efficiency',
    etrto: '55-559',
    inch_designation: '26x2.15',
    load_kg: 110,
    version_label: 'PRO, V-Guard',
    compound: 'ADDIX Race',
    color: 'Black+Reflex',
    bead: 'Folding',
    e_bike_rating: 'E-50',
    seal: 'TLR',
    source_checked_at: '2026-09-28',
    product_exists: true,
  },
  {
    article_no: 'future-invalid-etrto',
    model_name: 'Future Tire',
    etrto: 'unknown',
    version_label: 'RaceGuard',
    compound: 'ADDIX',
    color: 'Black',
    bead: 'WIRED',
    e_bike_rating: 'E-25',
    seal: 'Tube',
    source_checked_at: '2026-09-28',
    product_exists: false,
  },
]

const possibleCombinationRules = [
  {
    tireWidthMinMm: 20,
    tireWidthMaxMm: 21,
    innerRimWidthMinMm: 15,
    innerRimWidthMaxMm: 17,
  },
  {
    tireWidthMinMm: 35,
    tireWidthMaxMm: 46,
    innerRimWidthMinMm: 17,
    innerRimWidthMaxMm: 27,
  },
  {
    tireWidthMinMm: 47,
    tireWidthMaxMm: 57,
    innerRimWidthMinMm: 17,
    innerRimWidthMaxMm: 30,
  },
]

assert.deepEqual(
  deriveSchwalbeTireCatalogFilterDimensions(catalogItems[1]),
  {
    modelName: 'Marathon PRO Efficiency',
    nominalTireWidthMm: 55,
    beadSeatDiameterMm: 559,
    wheelDiameterIn: '26',
    wheelSizeKey: '26-559',
    loadKg: 110,
    versionLabel: 'PRO, V-Guard',
    casingConstructions: [],
    isRadial: false,
    compound: 'ADDIX Race',
    color: 'Black+Reflex',
    bead: 'Folding',
    seal: 'TLR',
    eBikeRating: 'E-50',
  },
)

const options = buildSchwalbeTireCatalogFilterOptions(catalogItems)
assert.deepEqual(options.nominalTireWidthsMm.map(option => option.value), [35, 55])
assert.deepEqual(options.beadSeatDiametersMm.map(option => option.value), [559])
assert.deepEqual(options.wheelSizes, [{
  value: '26-559',
  wheelDiameterIn: '26',
  beadSeatDiameterMm: 559,
}])
assert.deepEqual(options.beads.map(option => option.value), ['Folding', 'WIRED'])
assert.deepEqual(options.seals.map(option => option.value), ['TLR', 'Tube'])
assert.deepEqual(options.casingConstructions.map(option => option.value), [])
assert.deepEqual(options.colors.map(option => option.value), ['Black', 'Black+Reflex'])
assert.deepEqual(options.compounds.map(option => option.value), ['ADDIX', 'ADDIX Race'])
assert.deepEqual(
  options.eBikeRatings.map(option => option.value),
  ['E-25', 'E-50', null],
)

assert.equal(filterSchwalbeTireCatalogItems(catalogItems).length, 3)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    beads: ['Folding', 'WIRED'],
    seals: ['TLR'],
    colors: ['Black+Reflex'],
    compounds: ['ADDIX Race'],
  }).map(item => item.article_no),
  ['11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    nominalTireWidthMm: [35, 55],
    eBikeRatings: ['E-25', 'E-50'],
  }).map(item => item.article_no),
  ['11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    nominalTireWidthMinMm: 35,
    nominalTireWidthMaxMm: 55,
  }).map(item => item.article_no),
  ['11100062.02', '11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    nominalTireWidthMinMm: 40,
  }).map(item => item.article_no),
  ['11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    nominalTireWidthMaxMm: 34,
  }).map(item => item.article_no),
  [],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    nominalTireWidthMm: [55],
    versionLabels: ['RaceGuard'],
  }),
  [],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, { minimumLoadKg: 90 }).map(item => item.article_no),
  ['11100062.02', '11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, { minimumLoadKg: 100 }).map(item => item.article_no),
  ['11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, { minimumLoadKg: 120 }),
  [],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    nominalTireWidthMm: [55],
    modelNames: ['Future Tire', 'Marathon PRO Efficiency'],
  }).map(item => item.article_no),
  ['11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(catalogItems, {
    eBikeRatings: [null],
  }).map(item => item.article_no),
  ['11100062.02'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItemsByInnerRimWidth(catalogItems, 27, possibleCombinationRules)
    .map(item => item.article_no),
  ['11100062.02', '11654751'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItemsByInnerRimWidth(catalogItems, 31, possibleCombinationRules)
    .map(item => item.article_no),
  [],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItemsByInnerRimWidth(catalogItems, 0, possibleCombinationRules),
  [],
)

const casingCatalogItems: SchwalbeTireCatalogItem[] = [
  {
    ...catalogItems[0],
    article_no: 'casing-super-race',
    version_label: 'Super Race, RaceGuard',
  },
  {
    ...catalogItems[0],
    article_no: 'casing-trail-pro-radial',
    version_label: 'TRAIL PRO, Radial',
  },
  {
    ...catalogItems[0],
    article_no: 'casing-gravity-pro-radial',
    version_label: 'GRAVITY PRO, Radial',
  },
  {
    ...catalogItems[0],
    article_no: 'radial-protection-only',
    version_label: 'DD, GreenGuard, Radial',
  },
  {
    ...catalogItems[0],
    article_no: 'not-a-casing-token',
    version_label: 'RACE PRO',
  },
  {
    ...catalogItems[0],
    article_no: 'substring-is-not-a-match',
    version_label: 'Super Racekeeper, radial',
  },
]

assert.deepEqual(
  deriveSchwalbeTireCatalogFilterDimensions(casingCatalogItems[1]),
  {
    modelName: 'Kojak',
    nominalTireWidthMm: 35,
    beadSeatDiameterMm: 559,
    wheelDiameterIn: '26',
    wheelSizeKey: '26-559',
    loadKg: 90,
    versionLabel: 'TRAIL PRO, Radial',
    casingConstructions: ['TRAIL PRO'],
    isRadial: true,
    compound: 'ADDIX',
    color: 'Black',
    bead: 'WIRED',
    seal: 'Tube',
    eBikeRating: null,
  },
)
assert.deepEqual(
  buildSchwalbeTireCatalogFilterOptions(casingCatalogItems).casingConstructions.map(option => option.value),
  ['Super Race', 'TRAIL PRO', 'GRAVITY PRO'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(casingCatalogItems, {
    casingConstructions: ['Super Race', 'TRAIL PRO'],
  }).map(item => item.article_no),
  ['casing-super-race', 'casing-trail-pro-radial'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(casingCatalogItems, { radialOnly: true })
    .map(item => item.article_no),
  ['casing-trail-pro-radial', 'casing-gravity-pro-radial', 'radial-protection-only'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(casingCatalogItems, {
    casingConstructions: ['TRAIL PRO'],
    radialOnly: true,
  }).map(item => item.article_no),
  ['casing-trail-pro-radial'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(casingCatalogItems, {
    casingConstructions: ['Race', 'super race', 'Super Racekeeper', 'RACE PRO'],
  }),
  [],
)

const wheelSizeCatalogItems: SchwalbeTireCatalogItem[] = [
  { ...catalogItems[0], article_no: 'wheel-28-622', etrto: '40-622', inch_designation: '28x1.50' },
  { ...catalogItems[0], article_no: 'wheel-29-622', etrto: '57-622', inch_designation: '29x2.25' },
  { ...catalogItems[0], article_no: 'wheel-26-559', etrto: '50-559', inch_designation: '26x2.00' },
  { ...catalogItems[0], article_no: 'wheel-26-590', etrto: '37-590', inch_designation: '26x1 3/8' },
]

assert.deepEqual(
  buildSchwalbeTireCatalogFilterOptions(wheelSizeCatalogItems).wheelSizes,
  [
    { value: '26-559', wheelDiameterIn: '26', beadSeatDiameterMm: 559 },
    { value: '26-590', wheelDiameterIn: '26', beadSeatDiameterMm: 590 },
    { value: '28-622', wheelDiameterIn: '28', beadSeatDiameterMm: 622 },
    { value: '29-622', wheelDiameterIn: '29', beadSeatDiameterMm: 622 },
  ],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(wheelSizeCatalogItems, { wheelSizeKeys: ['28-622'] })
    .map(item => item.article_no),
  ['wheel-28-622'],
)
assert.deepEqual(
  filterSchwalbeTireCatalogItems(wheelSizeCatalogItems, { wheelSizeKeys: ['26-559'] })
    .map(item => item.article_no),
  ['wheel-26-559'],
)

console.log('Schwalbe catalog filter model checks passed.')

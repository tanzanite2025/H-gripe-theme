import assert from 'node:assert/strict'
import type { SchwalbeTireCatalogItem } from '../app/data/tireguides/schwalbeCatalog'
import {
  buildSchwalbeTireCatalogFilterOptions,
  deriveSchwalbeTireCatalogFilterDimensions,
  filterSchwalbeTireCatalogItemsByInnerRimWidth,
  filterSchwalbeTireCatalogItems,
} from '../app/data/tireguides/schwalbeTireCatalogFilterModel'

const catalogItems: SchwalbeTireCatalogItem[] = [
  {
    article_no: '11100062.02',
    model_name: 'Kojak',
    etrto: '35-559',
    version_label: 'RaceGuard',
    compound: 'ADDIX',
    color: 'Black',
    bead: 'WIRED',
    e_bike_rating: undefined,
    seal: 'Tube',
    source_url: 'https://www.schwalbe.com/en/Kojak-11100062.02',
    source_checked_at: '2026-09-28',
    product_exists: false,
  },
  {
    article_no: '11654751',
    model_name: 'Marathon PRO Efficiency',
    etrto: '55-559',
    version_label: 'PRO, V-Guard',
    compound: 'ADDIX Race',
    color: 'Black+Reflex',
    bead: 'Folding',
    e_bike_rating: 'E-50',
    seal: 'TLR',
    source_url: 'https://www.schwalbe.com/en/Marathon-PRO-Efficiency-11654751',
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
    source_url: 'https://www.schwalbe.com/en/future-invalid-etrto',
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
    versionLabel: 'PRO, V-Guard',
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
assert.deepEqual(options.beads.map(option => option.value), ['Folding', 'WIRED'])
assert.deepEqual(options.seals.map(option => option.value), ['TLR', 'Tube'])
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
    nominalTireWidthMm: [55],
    versionLabels: ['RaceGuard'],
  }),
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

console.log('Schwalbe catalog filter model checks passed.')

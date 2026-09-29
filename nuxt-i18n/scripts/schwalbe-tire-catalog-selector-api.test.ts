import assert from 'node:assert/strict'
import { fetchSchwalbeTireCatalogSelectorPage } from '../app/data/tireguides/schwalbeCatalog.js'
import type { ApiRequestFunction, ApiRequestInit } from '../app/composables/useApiRequest.js'
import type { SchwalbeTireCatalogSelectorRequest } from '../app/data/tireguides/schwalbeCatalog.js'

const selectorRequest: SchwalbeTireCatalogSelectorRequest = {
  search: ' Green Marathon ',
  page: 2,
  modelName: 'Green Marathon',
  nominalTireWidthMinMm: 40,
  nominalTireWidthMaxMm: 55,
  nominalTireWidthsMm: [40, 50],
  wheelSizeKeys: ['26-559', '28-622'],
  beadSeatDiametersMm: [622],
  minimumLoadKg: 90.5,
  casingConstructions: ['Super Ground'],
  radialOnly: true,
  beads: ['Folding'],
  seals: ['TLR'],
  eBikeRatings: [null, 'E-25'],
  colors: ['Black'],
  compounds: ['ADDIX'],
  sortBy: 'weight_desc',
}

let requestedPath = ''
let requestParams: ApiRequestInit['params']

const request: ApiRequestFunction = async <T>(path, init): Promise<T> => {
  requestedPath = path
  requestParams = init?.params
  return {
    code: 0,
    data: {
      items: [{
        article_no: '11100001',
        model_name: 'Green Marathon',
        etrto: '40-622',
        source_url: 'https://www.schwalbe.com/internal-provenance',
        source_checked_at: '2026-09-28',
        product_exists: false,
      }],
      page: 2,
      page_size: 20,
      total: 21,
      total_pages: 2,
      filter_options: {
        model_names: [{ value: 'Green Marathon' }],
        nominal_tire_widths_mm: [{ value: 40 }, { value: 50 }],
        bead_seat_diameters_mm: [{ value: 622 }],
        wheel_sizes: [
          { value: '26-559', wheel_diameter_in: '26', bsd_mm: 559 },
          { value: '28-622', wheel_diameter_in: '28', bsd_mm: 622 },
        ],
        casing_constructions: [{ value: 'Super Ground' }],
        beads: [{ value: 'Folding' }],
        seals: [{ value: 'TLR' }],
        e_bike_ratings: [{ value: 'E-25' }, { value: null }],
        colors: [{ value: 'Black' }],
        compounds: [{ value: 'ADDIX' }],
      },
    },
  } as T
}

const main = async () => {
  const page = await fetchSchwalbeTireCatalogSelectorPage(request, selectorRequest)
  assert.equal(requestedPath, '/products/schwalbe-tire-catalog/selector')

  const params = requestParams as Record<string, unknown>
  assert.equal(params.search, 'Green Marathon')
  assert.equal(params.page, '2')
  assert.equal(params.sort, 'weight_desc')
  assert.equal(params.tire_width_min_mm, '40')
  assert.equal(params.tire_width_max_mm, '55')
  assert.equal('tire_width_mm' in params, false)
  assert.deepEqual(params.bead_seat_diameter_mm, ['622'])
  assert.deepEqual(params.wheel_size, ['26-559', '28-622'])
  assert.equal(params.min_load_kg, '90.5')
  assert.deepEqual(params.e_bike_rating, ['none', 'E-25'])
  assert.equal(params.radial, '1')

  assert.equal(page.items.length, 1)
  assert.equal(page.items[0].article_no, '11100001')
  assert.equal('source_url' in page.items[0], false)
  assert.equal(page.page, 2)
  assert.equal(page.page_size, 20)
  assert.equal(page.total, 21)
  assert.equal(page.total_pages, 2)
  assert.deepEqual(page.filter_options.nominalTireWidthsMm.map(option => option.value), [40, 50])
  assert.deepEqual(page.filter_options.wheelSizes, [
    { value: '26-559', wheelDiameterIn: '26', beadSeatDiameterMm: 559 },
    { value: '28-622', wheelDiameterIn: '28', beadSeatDiameterMm: 622 },
  ])
  assert.deepEqual(page.filter_options.eBikeRatings.map(option => option.value), ['E-25', null])

  await fetchSchwalbeTireCatalogSelectorPage(request, {
    ...selectorRequest,
    nominalTireWidthMinMm: null,
    nominalTireWidthMaxMm: null,
  })
  const legacyParams = requestParams as Record<string, unknown>
  assert.deepEqual(legacyParams.tire_width_mm, ['40', '50'])
  assert.equal('tire_width_min_mm' in legacyParams, false)
  assert.equal('tire_width_max_mm' in legacyParams, false)
  console.log('Schwalbe selector API adapter checks passed.')
}

main().catch((error: unknown) => {
  console.error(error)
  process.exitCode = 1
})

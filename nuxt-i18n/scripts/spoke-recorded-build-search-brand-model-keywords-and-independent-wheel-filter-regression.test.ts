import assert from 'node:assert/strict'
import { test } from 'node:test'
import type { SpokeCatalog, SpokeRecordedResult } from '../app/data/spoke-calculator/database'
import { filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections } from '../app/utils/filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections'

const recordedBuildSearchCatalog: Pick<SpokeCatalog, 'rims' | 'hubs'> = {
  rims: [{ id: 'rim-brand', name: 'Example Rims', items: [{ id: 'rim-model', name: 'Road 50', erd: null }] }],
  hubs: [
    { id: 'front-hub-brand', name: 'Front Hub Company', items: [{ id: 'front-hub-model', name: 'Model 350' }] },
    { id: 'rear-hub-brand', name: 'Rear Hub Company', items: [{ id: 'rear-hub-model', name: 'Model 240' }] },
  ],
}

const frontRecordedBuild: SpokeRecordedResult = {
  id: 'front-recorded-build', name: 'Front reference build', description: 'Disc brake wheel',
  keywords: ['aero', '轻量'], rimBrandId: 'rim-brand', rimModelId: 'rim-model',
  hubBrandId: 'front-hub-brand', hubModelId: 'front-hub-model', wheelPosition: 'front',
  spokeCount: 24, crossing: 2, nippleType: 'standard',
  actualLengths: { frontLeft: 280, frontRight: 282, rearLeft: null, rearRight: null },
}

const rearRecordedBuild: SpokeRecordedResult = {
  ...frontRecordedBuild, id: 'rear-recorded-build', name: 'Rear reference build', keywords: [],
  hubBrandId: 'rear-hub-brand', hubModelId: 'rear-hub-model', wheelPosition: 'rear',
  actualLengths: { frontLeft: null, frontRight: null, rearLeft: 278, rearRight: 280 },
}

test('searches catalog brand/model labels, descriptions, keywords, case and multiple terms', () => {
  for (const query of ['350', 'MODEL 350', '  Front Hub   350  ', 'FRONT-HUB ３５０', 'Road 50', 'disc brake', 'AERO', '轻量']) {
    const results = filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
      [frontRecordedBuild], query, {}, recordedBuildSearchCatalog,
    )
    assert.deepEqual(results.map(result => result.id), [frontRecordedBuild.id], query)
  }
})

test('does not return records for empty, short, unrelated or punctuation-only queries', () => {
  for (const query of ['', ' ', 'a', 'missing model', '!!!']) {
    assert.deepEqual(filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
      [frontRecordedBuild], query, {}, recordedBuildSearchCatalog,
    ), [], query)
  }
})

test('front and rear selections independently match their own measured builds', () => {
  const results = filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
    [frontRecordedBuild, rearRecordedBuild], 'reference', {
      front: { rimBrandId: null, rimModelId: null, hubBrandId: 'front-hub-brand', hubModelId: 'front-hub-model' },
      rear: { rimBrandId: null, rimModelId: null, hubBrandId: 'rear-hub-brand', hubModelId: 'rear-hub-model' },
    }, recordedBuildSearchCatalog,
  )
  assert.deepEqual(results.map(result => result.id), ['front-recorded-build', 'rear-recorded-build'])
})

test('catalog selections also show results before entering a query', () => {
  const results = filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
    [frontRecordedBuild], '', {
      front: { rimBrandId: null, rimModelId: null, hubBrandId: 'front-hub-brand', hubModelId: null },
    }, recordedBuildSearchCatalog,
  )
  assert.deepEqual(results.map(result => result.id), ['front-recorded-build'])
})

test('combined wheel builds can match either selected wheel without requiring conflicting IDs', () => {
  const combinedRecordedBuild: SpokeRecordedResult = { ...frontRecordedBuild, wheelPosition: 'auto' }
  const results = filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
    [combinedRecordedBuild], 'reference', {
      front: { rimBrandId: null, rimModelId: null, hubBrandId: 'front-hub-brand', hubModelId: null },
      rear: { rimBrandId: null, rimModelId: null, hubBrandId: 'rear-hub-brand', hubModelId: null },
    }, recordedBuildSearchCatalog,
  )
  assert.deepEqual(results.map(result => result.id), ['front-recorded-build'])
  assert.deepEqual(filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
    [combinedRecordedBuild], 'reference', {
      front: { rimBrandId: null, rimModelId: null, hubBrandId: 'unmatched-brand', hubModelId: null },
    }, recordedBuildSearchCatalog,
  ), [])
})

test('keeps records with missing catalog labels searchable and excludes records without lengths', () => {
  assert.deepEqual(filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
    [frontRecordedBuild], 'reference', {}, { rims: [], hubs: [] },
  ).map(result => result.id), ['front-recorded-build'])
  assert.deepEqual(filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
    [{ ...frontRecordedBuild, actualLengths: { frontLeft: null, frontRight: null, rearLeft: null, rearRight: null } }],
    'reference', {}, recordedBuildSearchCatalog,
  ), [])
})

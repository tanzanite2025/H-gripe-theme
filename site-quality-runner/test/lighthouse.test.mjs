import assert from 'node:assert/strict'
import test from 'node:test'

import { selectMedianLighthouseResult } from '../src/lighthouse.mjs'

test('selects the Lighthouse sample with the median performance score', () => {
  const samples = [
    sample(0, 0.41),
    sample(1, 0.74),
    sample(2, 0.63),
  ]

  const selected = selectMedianLighthouseResult(samples)

  assert.equal(selected.sampleIndex, 2)
  assert.equal(selected.lhr.categories.performance.score, 0.63)
})

test('selects the sample closest to the statistical median for an even run count', () => {
  const samples = [
    sample(0, 0.31),
    sample(1, 0.71),
    sample(2, 0.69),
    sample(3, 0.92),
  ]

  const selected = selectMedianLighthouseResult(samples)

  assert.equal(selected.sampleIndex, 2)
  assert.equal(selected.lhr.categories.performance.score, 0.69)
})

test('falls back to median LCP when performance score is unavailable', () => {
  const samples = [
    sampleWithLCP(0, 4200),
    sampleWithLCP(1, 1600),
    sampleWithLCP(2, 2800),
  ]

  const selected = selectMedianLighthouseResult(samples)

  assert.equal(selected.sampleIndex, 2)
  assert.equal(selected.lhr.audits['largest-contentful-paint'].numericValue, 2800)
})

function sample(sampleIndex, performanceScore) {
  return {
    sampleIndex,
    lhr: {
      categories: {
        performance: { score: performanceScore },
      },
      audits: {},
    },
  }
}

function sampleWithLCP(sampleIndex, lcpMilliseconds) {
  return {
    sampleIndex,
    lhr: {
      categories: {},
      audits: {
        'largest-contentful-paint': { numericValue: lcpMilliseconds },
      },
    },
  }
}

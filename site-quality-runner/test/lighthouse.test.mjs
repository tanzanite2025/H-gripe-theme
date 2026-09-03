import assert from 'node:assert/strict'
import test from 'node:test'

import {
  calculateRenderedAuditBudgetMilliseconds,
  LighthouseExecutionError,
  runLighthouseSamples,
  selectMedianLighthouseResult,
} from '../src/lighthouse.mjs'

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

test('continues Lighthouse sampling after one run fails', async () => {
  const calls = []
  const results = await runLighthouseSamples('https://shop.example.com/', { output: 'json' }, 3, async (url, flags) => {
    calls.push({ url, flags })
    if (calls.length === 2) {
      throw new Error('CDP protocol disconnected')
    }
    return sample(undefined, calls.length === 1 ? 0.88 : 0.84)
  })

  assert.equal(calls.length, 3)
  assert.deepEqual(results.map((result) => result.sampleIndex), [0, 2])
  assert.deepEqual(results.map((result) => result.lhr.categories.performance.score), [0.88, 0.84])
})

test('drops Lighthouse samples that do not include a report', async () => {
  const results = await runLighthouseSamples('https://shop.example.com/', {}, 2, async (_url, _flags) => {
    if (_flags.seen) {
      return sample(undefined, 0.91)
    }
    _flags.seen = true
    return {}
  })

  assert.deepEqual(results.map((result) => result.sampleIndex), [1])
  assert.equal(results[0].lhr.categories.performance.score, 0.91)
})

test('fails Lighthouse sampling only when every run fails', async () => {
  await assert.rejects(
    () => runLighthouseSamples('https://shop.example.com/', {}, 2, async () => {
      throw new Error('target closed')
    }),
    (error) => {
      assert.ok(error instanceof LighthouseExecutionError)
      assert.equal(error.statusCode, 502)
      assert.match(error.message, /All Lighthouse samples failed/)
      assert.match(error.message, /sample 1: target closed/)
      assert.match(error.message, /sample 2: target closed/)
      return true
    },
  )
})

test('reserves more than the old 15 second cap for rendered audits', () => {
  const budget = calculateRenderedAuditBudgetMilliseconds({}, {
    timeoutMilliseconds: 90_000,
    linkCheckEnabled: true,
    linkCheckMaxLinks: 80,
    linkCheckTimeoutMilliseconds: 2_500,
  })

  assert.ok(budget >= 36_000)
  assert.ok(budget > 15_000)
})

test('increases the rendered audit reserve when runtime probes are enabled', () => {
  const simpleBudget = calculateRenderedAuditBudgetMilliseconds({}, {
    timeoutMilliseconds: 120_000,
    linkCheckEnabled: false,
    softNavigationMaxLinks: 0,
  })
  const runtimeBudget = calculateRenderedAuditBudgetMilliseconds({
    interactionProbes: [{ selector: '#buy' }, { selector: '#menu' }],
    softNavigationMaxLinks: 8,
  }, {
    timeoutMilliseconds: 120_000,
    linkCheckEnabled: true,
    linkCheckMaxLinks: 80,
    linkCheckTimeoutMilliseconds: 2_500,
    softNavigationMaxDurationMilliseconds: 2_000,
  })

  assert.ok(runtimeBudget > simpleBudget)
})

test('keeps a minimum Lighthouse timeout when rendered audit estimates are high', () => {
  const budget = calculateRenderedAuditBudgetMilliseconds({
    interactionProbes: Array.from({ length: 8 }, (_, index) => ({ selector: `#probe-${index}` })),
    softNavigationMaxLinks: 8,
  }, {
    timeoutMilliseconds: 45_000,
    linkCheckEnabled: true,
    linkCheckMaxLinks: 250,
    linkCheckTimeoutMilliseconds: 15_000,
    softNavigationMaxDurationMilliseconds: 10_000,
  })

  assert.ok(budget >= 25_000)
  assert.ok(budget <= 30_000)
})

test('does not estimate disabled rendered runtime audits', () => {
  const disabledBudget = calculateRenderedAuditBudgetMilliseconds({
    linkCheckEnabled: false,
    linkCheckMaxLinks: 0,
    softNavigationMaxLinks: 0,
  }, {
    timeoutMilliseconds: 90_000,
    linkCheckEnabled: true,
    linkCheckMaxLinks: 80,
  })

  assert.equal(disabledBudget, 36_000)
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

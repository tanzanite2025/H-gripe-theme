import assert from 'node:assert/strict'
import test from 'node:test'

import {
  captureInteractionAudit,
  captureRenderedLinkAudit,
  captureSoftNavigationAudit,
  mapLimited,
  resourceBudgetViolationsFromLighthouse,
  runtimeAuditBudgetRemaining,
  runtimeAuditDeadlineReserveMilliseconds,
  runtimeAuditDeadlineReached,
} from '../src/rendered_runtime_audits.mjs'

const originalFetch = globalThis.fetch

test('uses network transfer size before decoded resource size for budgets', () => {
  const violations = resourceBudgetViolationsFromLighthouse({
    audits: {
      'network-requests': {
        details: {
          items: [
            {
              url: 'https://shop.example.com/assets/app.js',
              resourceType: 'Script',
              mimeType: 'application/javascript',
              resourceSize: 320_000,
              transferSize: 80_000,
              encodedDataLength: 90_000,
            },
          ],
        },
      },
    },
  }, {
    jsBudgetBytes: 256_000,
  })

  assert.deepEqual(violations, [])
})

test('cancels link probe response bodies when GET fallback is required', async () => {
  let cancelCount = 0
  const requests = []
  const responses = [
    () => bodyResponse({ status: 405, onCancel: () => cancelCount++ }),
    () => bodyResponse({ status: 200, onCancel: () => cancelCount++ }),
  ]
  globalThis.fetch = async (url, init) => {
    requests.push({ url, method: init.method, headers: init.headers })
    return responses.shift()()
  }

  try {
    const audit = await captureRenderedLinkAudit(fakePage([
      { href: 'https://shop.example.com/about', text: 'About', source: 'anchor' },
    ]), {
      url: 'https://shop.example.com/',
      linkCheckEnabled: true,
      linkCheckMaxLinks: 1,
    })

    assert.equal(audit.status, 'complete')
    assert.equal(audit.links[0].statusCode, 200)
    assert.equal(audit.links[0].ok, true)
    assert.deepEqual(requests.map((request) => request.method), ['HEAD', 'GET'])
    assert.equal(requests[1].headers.Range, 'bytes=0-0')
    assert.equal(cancelCount, 2)
  } finally {
    globalThis.fetch = originalFetch
  }
})

test('cancels redirect response bodies before following the next location', async () => {
  let cancelCount = 0
  const requestedURLs = []
  const responses = [
    () => bodyResponse({
      status: 302,
      headers: { location: '/next' },
      onCancel: () => cancelCount++,
    }),
    () => bodyResponse({ status: 200, onCancel: () => cancelCount++ }),
  ]
  globalThis.fetch = async (url) => {
    requestedURLs.push(url)
    return responses.shift()()
  }

  try {
    const audit = await captureRenderedLinkAudit(fakePage([
      { href: 'https://shop.example.com/start', text: 'Start', source: 'anchor' },
    ]), {
      url: 'https://shop.example.com/',
      linkCheckEnabled: true,
      linkCheckMaxLinks: 1,
    })

    assert.equal(audit.status, 'complete')
    assert.equal(audit.links[0].statusCode, 200)
    assert.equal(audit.links[0].finalUrl, 'https://shop.example.com/next')
    assert.equal(audit.links[0].redirected, true)
    assert.deepEqual(requestedURLs, ['https://shop.example.com/start', 'https://shop.example.com/next'])
    assert.equal(cancelCount, 2)
  } finally {
    globalThis.fetch = originalFetch
  }
})

test('skips configured runtime audits when the shared rendered deadline has no reserve', async () => {
  const page = fakePage([{ href: 'https://shop.example.com/about', text: 'About', source: 'anchor' }])
  const expiredOptions = { deadlineAt: Date.now() + 1_000 }

  const linkAudit = await captureRenderedLinkAudit(page, {
    url: 'https://shop.example.com/',
    linkCheckEnabled: true,
    linkCheckMaxLinks: 1,
  }, expiredOptions)
  assert.equal(linkAudit.status, 'timeout_skipped')
  assert.equal(linkAudit.configured, true)
  assert.deepEqual(linkAudit.links, [])

  const interactionAudit = await captureInteractionAudit(page, {
    interactionProbes: [{ selector: '#buy', action: 'click' }],
  }, expiredOptions)
  assert.equal(interactionAudit.status, 'timeout_skipped')
  assert.equal(interactionAudit.configured, true)
  assert.deepEqual(interactionAudit.interactions, [])

  const softNavigationAudit = await captureSoftNavigationAudit(page, {
    softNavigationSelectors: ['nav a'],
  }, expiredOptions)
  assert.equal(softNavigationAudit.status, 'timeout_skipped')
  assert.equal(softNavigationAudit.configured, true)
  assert.deepEqual(softNavigationAudit.navigations, [])
})

test('mapLimited stops starting work once the shared rendered deadline is exhausted', async () => {
  const started = []
  let deadlineAt = Date.now() + 10_000
  const result = await mapLimited([1, 2, 3, 4], 1, async (item) => {
    started.push(item)
    deadlineAt = Date.now() + 1_000
    return item * 10
  }, {
    get deadlineAt() {
      return deadlineAt
    },
  })

  assert.deepEqual(started, [1])
  assert.deepEqual(result.results, [10])
  assert.equal(result.skippedCount, 3)
  assert.equal(result.timedOut, true)
})

test('marks deadline-clipped link aborts as timeout skipped', async () => {
  globalThis.fetch = async (_url, init) => new Promise((_, reject) => {
    init.signal.addEventListener('abort', () => {
      const error = new Error('operation aborted')
      error.name = 'AbortError'
      reject(error)
    }, { once: true })
  })

  try {
    const audit = await captureRenderedLinkAudit(fakePage([
      { href: 'https://shop.example.com/slow', text: 'Slow', source: 'anchor' },
    ]), {
      url: 'https://shop.example.com/',
      linkCheckEnabled: true,
      linkCheckMaxLinks: 1,
      linkCheckTimeoutMilliseconds: 2_500,
    }, {
      deadlineAt: Date.now() + runtimeAuditDeadlineReserveMilliseconds + 100,
    })

    assert.equal(audit.status, 'timeout_skipped')
    assert.equal(audit.links.length, 1)
    assert.equal(audit.links[0].timeoutSkipped, true)
    assert.match(audit.links[0].error, /deadline/)
  } finally {
    globalThis.fetch = originalFetch
  }
})

test('runtime deadline helpers keep a reserve window before the absolute timeout', () => {
  const deadlineAt = Date.now() + 6_000

  assert.equal(runtimeAuditDeadlineReached({ deadlineAt }), false)
  assert.ok(runtimeAuditBudgetRemaining({ deadlineAt }) > 0)
  assert.ok(runtimeAuditBudgetRemaining({ deadlineAt }) <= 1_000)
  assert.equal(runtimeAuditDeadlineReached({ deadlineAt: Date.now() + 4_999 }), true)
})

test('reports unchanged soft navigation when URL wait times out', async () => {
  const audit = await captureSoftNavigationAudit(fakeSoftNavigationPage(), {
    url: 'https://shop.example.com/',
    softNavigationMaxLinks: 1,
    softNavigationMaxDurationMilliseconds: 1,
  })

  assert.equal(audit.status, 'complete')
  assert.equal(audit.navigations.length, 1)
  assert.equal(audit.navigations[0].status, 'failed')
  assert.equal(audit.navigations[0].fromUrl, 'https://shop.example.com/')
  assert.equal(audit.navigations[0].toUrl, 'https://shop.example.com/')
  assert.equal(audit.navigations[0].error, 'route did not change after clicking navigation target')
})

function fakePage(links) {
  return {
    async evaluate() {
      return links
    },
    url() {
      return 'https://shop.example.com/'
    },
  }
}

function fakeSoftNavigationPage() {
  const currentURL = 'https://shop.example.com/'
  return {
    async evaluate(fn) {
      if (fn.name === 'snapshotSoftNavigationTargets') {
        return [{
          href: 'https://shop.example.com/products',
          text: 'Products',
          selector: 'a#products',
        }]
      }
      const source = String(fn)
      if (source.includes('__siteQualitySoftNavigationMarker === value')) {
        return true
      }
      if (source.includes('performance.memory')) {
        return 0
      }
      return undefined
    },
    async click() {},
    async goto() {},
    url() {
      return currentURL
    },
    waitForFunction() {
      return Promise.reject(new Error('waiting failed: timeout exceeded'))
    },
    async waitForNetworkIdle() {},
  }
}

function bodyResponse({ status, headers, onCancel }) {
  return new Response(
    new ReadableStream({
      cancel() {
        onCancel()
      },
    }),
    { status, headers },
  )
}

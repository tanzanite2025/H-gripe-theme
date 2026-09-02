import assert from 'node:assert/strict'
import test from 'node:test'

import { RunnerInputError, isPrivateAddress, loadRunnerConfig, normalizeRunInput } from '../src/validation.mjs'

const config = loadRunnerConfig({
  SITE_QUALITY_ALLOWED_ORIGIN: 'https://shop.example.com',
  SITE_QUALITY_RUNNER_TOKEN: '01234567890123456789012345678901',
})

test('normalizes a same-origin mobile run', async () => {
  const result = await normalizeRunInput(
    {
      url: 'https://shop.example.com/support/warranty#section',
      strategy: 'mobile',
      release_id: 'release-1',
      render_wait_selector: ' #custom-wheelset-builder ',
      render_wait_timeout_ms: 12_500,
      heading_settle_ms: 5_500,
      structured_data_settle_ms: 6_000,
      throttling_method: 'devtools',
      lighthouse_run_count: 3,
    },
    config,
    async () => [{ address: '203.0.113.20', family: 4 }],
  )

  assert.equal(result.url, 'https://shop.example.com/support/warranty')
  assert.equal(result.strategy, 'mobile')
  assert.equal(result.releaseID, 'release-1')
  assert.equal(result.renderWaitSelector, '#custom-wheelset-builder')
  assert.equal(result.renderWaitTimeoutMilliseconds, 12_500)
  assert.equal(result.headingSettleMilliseconds, 5_500)
  assert.equal(result.structuredDataSettleMilliseconds, 6_000)
  assert.equal(result.throttlingMethod, 'devtools')
  assert.equal(result.lighthouseRunCount, 3)
})

test('loads a bounded worker heap limit', () => {
  const defaultConfig = loadRunnerConfig({
    SITE_QUALITY_ALLOWED_ORIGIN: 'https://shop.example.com',
    SITE_QUALITY_RUNNER_TOKEN: '01234567890123456789012345678901',
  })

  assert.equal(defaultConfig.workerHeapLimitMB, 1024)

  const overriddenConfig = loadRunnerConfig({
    SITE_QUALITY_ALLOWED_ORIGIN: 'https://shop.example.com',
    SITE_QUALITY_RUNNER_TOKEN: '01234567890123456789012345678901',
    SITE_QUALITY_RUNNER_WORKER_HEAP_LIMIT_MB: '1536',
  })

  assert.equal(overriddenConfig.workerHeapLimitMB, 1536)

  const cappedConfig = loadRunnerConfig({
    SITE_QUALITY_ALLOWED_ORIGIN: 'https://shop.example.com',
    SITE_QUALITY_RUNNER_TOKEN: '01234567890123456789012345678901',
    SITE_QUALITY_RUNNER_WORKER_HEAP_LIMIT_MB: '4096',
  })

  assert.equal(cappedConfig.workerHeapLimitMB, 1536)
})

test('loads runner accuracy controls from the environment', () => {
  const runnerConfig = loadRunnerConfig({
    SITE_QUALITY_ALLOWED_ORIGIN: 'https://shop.example.com',
    SITE_QUALITY_RUNNER_TOKEN: '01234567890123456789012345678901',
    SITE_QUALITY_THROTTLING_METHOD: 'devtools',
    SITE_QUALITY_LIGHTHOUSE_RUN_COUNT: '7',
    SITE_QUALITY_RENDER_WAIT_SELECTOR: ' #wheel-builder ',
    SITE_QUALITY_RENDER_WAIT_TIMEOUT_MS: '60000',
    SITE_QUALITY_HEADING_SETTLE_MS: '9000',
    SITE_QUALITY_STRUCTURED_DATA_SETTLE_MS: '8000',
  })

  assert.equal(runnerConfig.throttlingMethod, 'devtools')
  assert.equal(runnerConfig.lighthouseRunCount, 5)
  assert.equal(runnerConfig.renderWaitSelector, '#wheel-builder')
  assert.equal(runnerConfig.renderWaitTimeoutMilliseconds, 25_000)
  assert.equal(runnerConfig.headingSettleMilliseconds, 9_000)
  assert.equal(runnerConfig.structuredDataSettleMilliseconds, 8_000)
})

test('rejects invalid throttling mode', async () => {
  await assert.rejects(
    () => normalizeRunInput(
      { url: 'https://shop.example.com/', strategy: 'desktop', throttling_method: 'magic' },
      config,
      async () => [{ address: '203.0.113.20', family: 4 }],
    ),
    RunnerInputError,
  )
})

test('rejects a target outside the configured storefront origin', async () => {
  await assert.rejects(
    () => normalizeRunInput(
      { url: 'https://example.com/', strategy: 'desktop' },
      config,
      async () => [{ address: '203.0.113.20', family: 4 }],
    ),
    RunnerInputError,
  )
})

test('rejects private DNS answers outside explicit development mode', async () => {
  await assert.rejects(
    () => normalizeRunInput(
      { url: 'https://shop.example.com/', strategy: 'desktop' },
      config,
      async () => [{ address: '10.0.0.2', family: 4 }],
    ),
    RunnerInputError,
  )
})

test('recognizes IPv4 and IPv6 private addresses', () => {
  assert.equal(isPrivateAddress('127.0.0.1'), true)
  assert.equal(isPrivateAddress('10.12.3.4'), true)
  assert.equal(isPrivateAddress('192.168.1.1'), true)
  assert.equal(isPrivateAddress('::1'), true)
  assert.equal(isPrivateAddress('fc00::1'), true)
  assert.equal(isPrivateAddress('203.0.113.20'), false)
})

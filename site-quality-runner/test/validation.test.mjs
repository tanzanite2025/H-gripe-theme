import assert from 'node:assert/strict'
import test from 'node:test'

import { RunnerInputError, isPrivateAddress, loadRunnerConfig, normalizeRunInput } from '../src/validation.mjs'

const config = loadRunnerConfig({
  SITE_QUALITY_ALLOWED_ORIGIN: 'https://shop.example.com',
  SITE_QUALITY_RUNNER_TOKEN: '01234567890123456789012345678901',
})

test('normalizes a same-origin mobile run', async () => {
  const result = await normalizeRunInput(
    { url: 'https://shop.example.com/support/warranty#section', strategy: 'mobile', release_id: 'release-1' },
    config,
    async () => [{ address: '203.0.113.20', family: 4 }],
  )

  assert.equal(result.url, 'https://shop.example.com/support/warranty')
  assert.equal(result.strategy, 'mobile')
  assert.equal(result.releaseID, 'release-1')
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

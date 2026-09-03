import { fork } from 'node:child_process'
import { readdir, readFile } from 'node:fs/promises'
import { setTimeout as delay } from 'node:timers/promises'
import { fileURLToPath } from 'node:url'

import { launch } from 'chrome-launcher'
import lighthouse from 'lighthouse'

import { captureRenderedDocumentAudit } from './rendered_document.mjs'
import { applyResourceBudgetAudit } from './rendered_runtime_audits.mjs'

const workerScript = fileURLToPath(new URL('./lighthouse_worker.mjs', import.meta.url))
const defaultWorkerHeapLimitMB = 1024
const workerGraceMilliseconds = 5_000
const minimumLighthouseTimeoutMilliseconds = 15_000
const minimumRenderedAuditBudgetMilliseconds = 25_000
const renderedAuditBudgetRatio = 0.4
const maxRenderedAuditBudgetRatio = 0.5
const renderedAuditLinkProbeConcurrency = 4
const maxRenderedAuditLinkBudgetMilliseconds = 20_000
const maxRenderedAuditInteractionBudgetMilliseconds = 8_000
const maxRenderedAuditSoftNavigationBudgetMilliseconds = 25_000

export class LighthouseExecutionError extends Error {
  constructor(message, statusCode = 502) {
    super(message)
    this.name = 'LighthouseExecutionError'
    this.statusCode = statusCode
  }
}

// Lighthouse keeps large trace and DevTools artifacts alive during a run.
// A child-process boundary guarantees that memory is reclaimed after each job.
export function runLighthouseInWorker(input, config) {
  return new Promise((resolve, reject) => {
    const workerHeapLimitMB = normalizeWorkerHeapLimitMB(config?.workerHeapLimitMB)
    const worker = fork(workerScript, [], {
      execArgv: [`--max-old-space-size=${workerHeapLimitMB}`],
      stdio: ['ignore', 'ignore', 'inherit', 'ipc'],
    })
    let settled = false
    const timeoutMilliseconds = Math.max(
      Number(config?.timeoutMilliseconds || 0) + workerGraceMilliseconds,
      workerGraceMilliseconds,
    )
    const timeoutHandle = setTimeout(() => {
      worker.kill('SIGKILL')
      finish(new LighthouseExecutionError('Lighthouse worker timed out', 504))
    }, timeoutMilliseconds)

    function finish(error, result) {
      if (settled) {
        return
      }
      settled = true
      clearTimeout(timeoutHandle)
      if (error) {
        reject(error)
        return
      }
      resolve(result)
    }

    worker.once('message', (message) => {
      if (message?.ok) {
        finish(null, message.result)
        return
      }
      const statusCode = Number.isInteger(message?.error?.statusCode)
        ? message.error.statusCode
        : 502
      finish(new LighthouseExecutionError(
        message?.error?.message || 'Lighthouse worker failed',
        statusCode,
      ))
    })
    worker.once('error', (error) => {
      finish(new LighthouseExecutionError(`Lighthouse worker process failed: ${error.message}`))
    })
    worker.once('exit', (code, signal) => {
      if (!settled) {
        const reason = signal ? `signal ${signal}` : `exit code ${code}`
        finish(new LighthouseExecutionError(`Lighthouse worker exited unexpectedly (${reason})`))
      }
    })
    worker.send({
      input,
      config: {
        timeoutMilliseconds: config?.timeoutMilliseconds,
        headingSettleMilliseconds: config?.headingSettleMilliseconds,
        structuredDataSettleMilliseconds: config?.structuredDataSettleMilliseconds,
        renderWaitSelector: config?.renderWaitSelector,
        renderWaitTimeoutMilliseconds: config?.renderWaitTimeoutMilliseconds,
        throttlingMethod: config?.throttlingMethod,
        lighthouseRunCount: config?.lighthouseRunCount,
        interactionProbes: config?.interactionProbes,
        interactionMaxResponseMilliseconds: config?.interactionMaxResponseMilliseconds,
        softNavigationSelectors: config?.softNavigationSelectors,
        softNavigationMaxLinks: config?.softNavigationMaxLinks,
        softNavigationMaxDurationMilliseconds: config?.softNavigationMaxDurationMilliseconds,
        softNavigationMaxHeapGrowthMB: config?.softNavigationMaxHeapGrowthMB,
        jsBudgetBytes: config?.jsBudgetBytes,
        imageBudgetBytes: config?.imageBudgetBytes,
        linkCheckEnabled: config?.linkCheckEnabled,
        linkCheckMaxLinks: config?.linkCheckMaxLinks,
        linkCheckTimeoutMilliseconds: config?.linkCheckTimeoutMilliseconds,
        linkCheckExternal: config?.linkCheckExternal,
        linkCheckMaxRedirects: config?.linkCheckMaxRedirects,
      },
    }, (error) => {
      if (error) {
        finish(new LighthouseExecutionError(`Lighthouse worker request failed: ${error.message}`))
      }
    })
  })
}

export async function runLighthouse(input, config) {
  const chrome = await launch({
    chromePath: process.env.CHROME_PATH || undefined,
    chromeFlags: [
      '--headless=new',
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-dev-shm-usage',
      '--disable-gpu',
      '--disable-background-networking',
      '--disable-component-update',
      '--disable-default-apps',
      '--disable-extensions',
      '--disable-sync',
      '--metrics-recording-only',
      '--mute-audio',
      '--no-first-run',
    ],
  })

  let timeoutHandle
  try {
    const startedAt = Date.now()
    const runCount = normalizeLighthouseRunCount(input?.lighthouseRunCount ?? config?.lighthouseRunCount)
    const renderedAuditBudgetMilliseconds = calculateRenderedAuditBudgetMilliseconds(input, config)
    const lighthouseTimeoutMilliseconds = Math.max(
      minimumLighthouseTimeoutMilliseconds,
      config.timeoutMilliseconds - renderedAuditBudgetMilliseconds,
    )
    const flags = {
      port: chrome.port,
      output: 'json',
      logLevel: 'error',
      onlyCategories: ['performance', 'accessibility', 'best-practices', 'seo'],
      disableFullPageScreenshot: true,
      formFactor: input.strategy,
      screenEmulation: input.strategy === 'mobile'
        ? { mobile: true, width: 412, height: 823, deviceScaleFactor: 1.75, disabled: false }
        : { mobile: false, width: 1350, height: 940, deviceScaleFactor: 1, disabled: false },
      throttlingMethod: input.throttlingMethod || config.throttlingMethod || 'simulate',
    }
    const lighthouseResults = await Promise.race([
      runLighthouseSamples(input.url, flags, runCount),
      new Promise((_, reject) => {
        timeoutHandle = setTimeout(() => {
          void chrome.kill()
          reject(new LighthouseExecutionError('Lighthouse run timed out', 504))
        }, lighthouseTimeoutMilliseconds)
      }),
    ])
    const result = selectMedianLighthouseResult(lighthouseResults)
    if (!result?.lhr) {
      throw new LighthouseExecutionError('Lighthouse returned no report')
    }
    applyResourceBudgetAudit(result.lhr, {
      jsBudgetBytes: input.jsBudgetBytes ?? config.jsBudgetBytes,
      imageBudgetBytes: input.imageBudgetBytes ?? config.imageBudgetBytes,
    })
    annotateLighthouseConfigSettings(result.lhr, {
      lighthouseRunCount: runCount,
      selectedSampleIndex: result.sampleIndex,
      sampleSummaries: summarizeLighthouseSamples(lighthouseResults),
      throttlingMethod: flags.throttlingMethod,
      renderWaitSelector: input.renderWaitSelector || config.renderWaitSelector || '',
      renderWaitTimeoutMilliseconds: input.renderWaitTimeoutMilliseconds || config.renderWaitTimeoutMilliseconds,
      headingSettleMilliseconds: input.headingSettleMilliseconds || config.headingSettleMilliseconds,
      structuredDataSettleMilliseconds: input.structuredDataSettleMilliseconds || config.structuredDataSettleMilliseconds,
      interactionProbeCount: Array.isArray(input.interactionProbes) ? input.interactionProbes.length : 0,
      interactionMaxResponseMilliseconds: input.interactionMaxResponseMilliseconds || config.interactionMaxResponseMilliseconds,
      softNavigationMaxLinks: input.softNavigationMaxLinks || config.softNavigationMaxLinks || 0,
      softNavigationMaxDurationMilliseconds: input.softNavigationMaxDurationMilliseconds || config.softNavigationMaxDurationMilliseconds,
      softNavigationMaxHeapGrowthMB: input.softNavigationMaxHeapGrowthMB || config.softNavigationMaxHeapGrowthMB,
      jsBudgetBytes: input.jsBudgetBytes ?? config.jsBudgetBytes,
      imageBudgetBytes: input.imageBudgetBytes ?? config.imageBudgetBytes,
      linkCheckEnabled: input.linkCheckEnabled ?? config.linkCheckEnabled,
      linkCheckMaxLinks: input.linkCheckMaxLinks ?? config.linkCheckMaxLinks,
      linkCheckTimeoutMilliseconds: input.linkCheckTimeoutMilliseconds ?? config.linkCheckTimeoutMilliseconds,
      linkCheckExternal: input.linkCheckExternal ?? config.linkCheckExternal,
    })
    const remainingTimeoutMilliseconds = Math.max(
      1_000,
      config.timeoutMilliseconds - (Date.now() - startedAt),
    )
    const renderedDocument = await captureRenderedDocumentAuditResult({
      debuggingPort: chrome.port,
      input,
      timeoutMilliseconds: remainingTimeoutMilliseconds,
      settleMilliseconds: Math.max(
        input.headingSettleMilliseconds || config.headingSettleMilliseconds || 0,
        input.structuredDataSettleMilliseconds || config.structuredDataSettleMilliseconds || 0,
      ),
      userAgent: String(result.lhr.configSettings?.emulatedUserAgent || '').trim(),
    })
    return normalizeLighthouseReport(
      result.lhr,
      renderedDocument.renderedHeadings,
      renderedDocument.renderedStructuredData,
      renderedDocument.renderedLinks,
      renderedDocument.interactionAudit,
      renderedDocument.softNavigationAudit,
    )
  } catch (error) {
    if (error instanceof LighthouseExecutionError) {
      throw error
    }
    throw new LighthouseExecutionError(`Lighthouse run failed: ${error instanceof Error ? error.message : String(error)}`)
  } finally {
    if (timeoutHandle) {
      clearTimeout(timeoutHandle)
    }
    try {
      await chrome.kill()
    } catch {
      // A timed-out or failed Chrome process may already be gone.
    }
  }
}

function summarizeLighthouseSamples(results) {
  return (Array.isArray(results) ? results : [])
    .filter((item) => item?.lhr)
    .map((item) => ({
      sampleIndex: item.sampleIndex,
      performanceScore: finiteOrNull(item.lhr.categories?.performance?.score),
      largestContentfulPaintMilliseconds: finiteOrNull(item.lhr.audits?.['largest-contentful-paint']?.numericValue),
    }))
}

function finiteOrNull(value) {
  const numeric = Number(value)
  return Number.isFinite(numeric) ? numeric : null
}

export async function runLighthouseSamples(url, flags, runCount, lighthouseRunner = lighthouse) {
  const results = []
  const failures = []
  for (let index = 0; index < runCount; index++) {
    try {
      const result = await lighthouseRunner(url, flags)
      if (result?.lhr) {
        result.sampleIndex = index
        results.push(result)
        continue
      }
      failures.push({ sampleIndex: index, error: 'Lighthouse returned no report' })
    } catch (error) {
      failures.push({ sampleIndex: index, error: normalizeLighthouseSampleError(error) })
    }
  }
  if (results.length === 0) {
    const details = failures
      .map((failure) => `sample ${failure.sampleIndex + 1}: ${failure.error}`)
      .join('; ')
    throw new LighthouseExecutionError(`All Lighthouse samples failed${details ? `: ${details}` : ''}`)
  }
  return results
}

function normalizeLighthouseSampleError(error) {
  const message = error instanceof Error ? error.message : String(error)
  return message.replace(/\s+/g, ' ').trim().slice(0, 500) || 'Lighthouse sample failed'
}

export function selectMedianLighthouseResult(results) {
  const samples = Array.isArray(results) ? results.filter((item) => item?.lhr) : []
  if (samples.length === 0) {
    return null
  }
  const scored = samples.map((sample) => ({
    sample,
    score: lighthouseResultMedianRankScore(sample.lhr),
  }))
  const rankedScores = scored.map((item) => item.score).sort((a, b) => a - b)
  const index = Math.floor(rankedScores.length / 2)
  const medianScore = rankedScores.length % 2 === 0
    ? (rankedScores[index - 1] + rankedScores[index]) / 2
    : rankedScores[index]
  scored.sort((a, b) => {
    const distanceA = Math.abs(a.score - medianScore)
    const distanceB = Math.abs(b.score - medianScore)
    if (distanceA !== distanceB) {
      return distanceA - distanceB
    }
    if (a.score !== b.score) {
      return a.score - b.score
    }
    return (a.sample.sampleIndex || 0) - (b.sample.sampleIndex || 0)
  })
  return scored[0].sample
}

export function calculateRenderedAuditBudgetMilliseconds(input = {}, config = {}) {
  const totalTimeoutMilliseconds = Math.max(0, Number(config?.timeoutMilliseconds || 0))
  const maxReservableMilliseconds = Math.max(
    5_000,
    totalTimeoutMilliseconds - minimumLighthouseTimeoutMilliseconds,
  )
  const proportionalBudgetMilliseconds = Math.round(totalTimeoutMilliseconds * renderedAuditBudgetRatio)
  const balancedMaximumBudgetMilliseconds = Math.max(
    minimumRenderedAuditBudgetMilliseconds,
    Math.round(totalTimeoutMilliseconds * maxRenderedAuditBudgetRatio),
  )
  const desiredBudgetMilliseconds = Math.max(
    minimumRenderedAuditBudgetMilliseconds,
    proportionalBudgetMilliseconds,
    estimateRenderedAuditFeatureBudgetMilliseconds(input, config),
  )
  return Math.min(maxReservableMilliseconds, balancedMaximumBudgetMilliseconds, desiredBudgetMilliseconds)
}

function estimateRenderedAuditFeatureBudgetMilliseconds(input = {}, config = {}) {
  const settleMilliseconds = Math.max(
    positiveNumber(input.headingSettleMilliseconds ?? config.headingSettleMilliseconds, 1_500),
    positiveNumber(input.structuredDataSettleMilliseconds ?? config.structuredDataSettleMilliseconds, 1_500),
  )
  const renderWaitMilliseconds = String(input.renderWaitSelector || config.renderWaitSelector || '').trim()
    ? positiveNumber(input.renderWaitTimeoutMilliseconds ?? config.renderWaitTimeoutMilliseconds, 3_000)
    : 0
  const baseBudgetMilliseconds = Math.max(10_000, 7_000 + settleMilliseconds + renderWaitMilliseconds)

  return baseBudgetMilliseconds +
    estimateRenderedLinkAuditBudgetMilliseconds(input, config) +
    estimateInteractionAuditBudgetMilliseconds(input, config) +
    estimateSoftNavigationAuditBudgetMilliseconds(input, config)
}

function estimateRenderedLinkAuditBudgetMilliseconds(input = {}, config = {}) {
  const linkCheckEnabled = input.linkCheckEnabled ?? config.linkCheckEnabled
  const linkCount = boundedInteger(input.linkCheckMaxLinks ?? config.linkCheckMaxLinks, 80, 0, 250)
  if (!linkCheckEnabled || linkCount <= 0) {
    return 0
  }
  const linkTimeoutMilliseconds = boundedInteger(
    input.linkCheckTimeoutMilliseconds ?? config.linkCheckTimeoutMilliseconds,
    2_500,
    500,
    15_000,
  )
  const batches = Math.ceil(linkCount / renderedAuditLinkProbeConcurrency)
  const estimatedBatchMilliseconds = Math.min(1_000, linkTimeoutMilliseconds)
  return Math.min(
    maxRenderedAuditLinkBudgetMilliseconds,
    Math.max(5_000, batches * estimatedBatchMilliseconds),
  )
}

function estimateInteractionAuditBudgetMilliseconds(input = {}, config = {}) {
  const probes = Array.isArray(input.interactionProbes)
    ? input.interactionProbes
    : Array.isArray(config.interactionProbes)
      ? config.interactionProbes
      : []
  if (probes.length === 0) {
    return 0
  }
  const probeCount = Math.min(probes.length, 8)
  const thresholdMilliseconds = boundedInteger(
    input.interactionMaxResponseMilliseconds ?? config.interactionMaxResponseMilliseconds,
    200,
    50,
    2_000,
  )
  return Math.min(
    maxRenderedAuditInteractionBudgetMilliseconds,
    probeCount * Math.max(1_000, thresholdMilliseconds + 500),
  )
}

function estimateSoftNavigationAuditBudgetMilliseconds(input = {}, config = {}) {
  const selectors = Array.isArray(input.softNavigationSelectors)
    ? input.softNavigationSelectors
    : Array.isArray(config.softNavigationSelectors)
      ? config.softNavigationSelectors
      : []
  const targetCount = boundedInteger(
    input.softNavigationMaxLinks ?? config.softNavigationMaxLinks,
    selectors.length > 0 ? selectors.length : 0,
    0,
    8,
  )
  if (targetCount <= 0 && selectors.length === 0) {
    return 0
  }
  const thresholdMilliseconds = boundedInteger(
    input.softNavigationMaxDurationMilliseconds ?? config.softNavigationMaxDurationMilliseconds,
    2_000,
    250,
    10_000,
  )
  const checkedTargetCount = Math.max(targetCount, selectors.length)
  const estimatedTargetMilliseconds = Math.min(4_000, thresholdMilliseconds + 1_500)
  return Math.min(
    maxRenderedAuditSoftNavigationBudgetMilliseconds,
    Math.max(5_000, checkedTargetCount * estimatedTargetMilliseconds),
  )
}

function boundedInteger(value, fallback, minimum, maximum) {
  const parsed = Number.parseInt(String(value ?? ''), 10)
  if (!Number.isInteger(parsed)) {
    return fallback
  }
  return Math.max(minimum, Math.min(parsed, maximum))
}

function positiveNumber(value, fallback) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return parsed
}

function lighthouseResultMedianRankScore(lhr) {
  const performanceScore = Number(lhr?.categories?.performance?.score)
  if (Number.isFinite(performanceScore)) {
    return performanceScore
  }
  const lcp = Number(lhr?.audits?.['largest-contentful-paint']?.numericValue)
  if (Number.isFinite(lcp)) {
    return -lcp
  }
  return 0
}

function annotateLighthouseConfigSettings(lhr, settings) {
  if (!lhr || typeof lhr !== 'object') {
    return
  }
  const configSettings = lhr.configSettings && typeof lhr.configSettings === 'object'
    ? lhr.configSettings
    : {}
  configSettings.siteQuality = {
    lighthouseRunCount: settings.lighthouseRunCount,
    selectedSampleIndex: settings.selectedSampleIndex,
    sampleSummaries: Array.isArray(settings.sampleSummaries) ? settings.sampleSummaries : [],
    throttlingMethod: settings.throttlingMethod,
    renderWaitSelector: String(settings.renderWaitSelector || ''),
    renderWaitTimeoutMilliseconds: settings.renderWaitTimeoutMilliseconds,
    headingSettleMilliseconds: settings.headingSettleMilliseconds,
    structuredDataSettleMilliseconds: settings.structuredDataSettleMilliseconds,
    interactionProbeCount: settings.interactionProbeCount,
    interactionMaxResponseMilliseconds: settings.interactionMaxResponseMilliseconds,
    softNavigationMaxLinks: settings.softNavigationMaxLinks,
    softNavigationMaxDurationMilliseconds: settings.softNavigationMaxDurationMilliseconds,
    softNavigationMaxHeapGrowthMB: settings.softNavigationMaxHeapGrowthMB,
    jsBudgetBytes: settings.jsBudgetBytes,
    imageBudgetBytes: settings.imageBudgetBytes,
    linkCheckEnabled: settings.linkCheckEnabled,
    linkCheckMaxLinks: settings.linkCheckMaxLinks,
    linkCheckTimeoutMilliseconds: settings.linkCheckTimeoutMilliseconds,
    linkCheckExternal: settings.linkCheckExternal,
  }
  lhr.configSettings = configSettings
}

export async function cleanupOrphanLighthouseBrowsers() {
  const entries = await readdir('/proc', { withFileTypes: true }).catch(() => [])
  const stalePIDs = []

  await Promise.all(entries.map(async (entry) => {
    if (!entry.isDirectory() || !/^\d+$/.test(entry.name)) return

    const pid = Number.parseInt(entry.name, 10)
    if (!Number.isInteger(pid) || pid <= 1 || pid === process.pid) return

    const command = await readFile(`/proc/${pid}/cmdline`, 'utf8').catch(() => '')
    if (!isLighthouseBrowserCommand(command)) return

    if (signalProcess(pid, 'SIGTERM')) {
      stalePIDs.push(pid)
    }
  }))

  if (stalePIDs.length === 0) {
    return 0
  }

  await delay(250)
  stalePIDs.forEach((pid) => {
    signalProcess(pid, 'SIGKILL')
  })

  return stalePIDs.length
}

function normalizeWorkerHeapLimitMB(value) {
  const parsed = Number.parseInt(String(value || ''), 10)
  if (!Number.isInteger(parsed)) {
    return defaultWorkerHeapLimitMB
  }
  return Math.min(Math.max(parsed, 384), 1536)
}

function normalizeLighthouseRunCount(value) {
  const parsed = Number.parseInt(String(value || ''), 10)
  if (!Number.isInteger(parsed)) {
    return 1
  }
  return Math.min(Math.max(parsed, 1), 5)
}

function isLighthouseBrowserCommand(command) {
  const normalized = String(command || '').replace(/\0/g, ' ')
  if (!normalized) {
    return false
  }
  return (
    normalized.includes('/usr/lib/chromium/chromium') ||
    normalized.includes('/usr/lib/chromium/chrome_crashpad_handler')
  )
}

function signalProcess(pid, signal) {
  try {
    process.kill(pid, signal)
    return true
  } catch (error) {
    return error?.code === 'ESRCH'
  }
}

async function captureRenderedDocumentAuditResult(options) {
  try {
    return await captureRenderedDocumentAudit(options)
  } catch (error) {
    const normalized = normalizeRenderedDocumentAuditError(error)
    const input = options?.input || {}
    return {
      renderedHeadings: {
        status: 'failed',
        source: 'chrome-rendered-dom',
        error: normalized,
      },
      renderedStructuredData: {
        status: 'failed',
        source: 'chrome-rendered-dom',
        error: normalized,
      },
      renderedLinks: {
        status: 'failed',
        source: 'chrome-rendered-dom',
        error: normalized,
        configured: Boolean(input.linkCheckEnabled),
        links: [],
      },
      interactionAudit: {
        status: 'failed',
        source: 'chrome-rendered-dom',
        error: normalized,
        configured: Array.isArray(input.interactionProbes) && input.interactionProbes.length > 0,
        interactions: [],
      },
      softNavigationAudit: {
        status: 'failed',
        source: 'chrome-rendered-dom',
        error: normalized,
        configured: (Array.isArray(input.softNavigationSelectors) && input.softNavigationSelectors.length > 0) ||
          Number(input.softNavigationMaxLinks || 0) > 0,
        navigations: [],
      },
    }
  }
}

function normalizeRenderedDocumentAuditError(error) {
  const message = error instanceof Error ? error.message : String(error)
  return message.replace(/\s+/g, ' ').trim().slice(0, 500) || 'rendered document audit failed'
}

function normalizeLighthouseReport(
  lhr,
  renderedHeadings,
  renderedStructuredData,
  renderedLinks,
  interactionAudit,
  softNavigationAudit,
) {
  return {
    lighthouseResult: {
      finalUrl: String(lhr.finalUrl || ''),
      lighthouseVersion: String(lhr.lighthouseVersion || ''),
      configSettings: lhr.configSettings || {},
      categories: lhr.categories || {},
      audits: lhr.audits || {},
      renderedHeadings,
      renderedStructuredData,
      renderedLinks,
      interactionAudit,
      softNavigationAudit,
    },
  }
}

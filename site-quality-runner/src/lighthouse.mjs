import { fork } from 'node:child_process'
import { fileURLToPath } from 'node:url'

import { launch } from 'chrome-launcher'
import lighthouse from 'lighthouse'

import { captureRenderedDocumentAudit } from './rendered_document.mjs'

const workerScript = fileURLToPath(new URL('./lighthouse_worker.mjs', import.meta.url))
const workerHeapLimitMB = 768
const workerGraceMilliseconds = 5_000

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
    const renderedAuditBudgetMilliseconds = Math.min(
      15_000,
      Math.max(5_000, Math.round(config.timeoutMilliseconds * 0.2)),
    )
    const lighthouseTimeoutMilliseconds = Math.max(
      15_000,
      config.timeoutMilliseconds - renderedAuditBudgetMilliseconds,
    )
    const flags = {
      port: chrome.port,
      output: 'json',
      logLevel: 'error',
      onlyCategories: ['performance', 'accessibility', 'best-practices', 'seo'],
      formFactor: input.strategy,
      screenEmulation: input.strategy === 'mobile'
        ? { mobile: true, width: 412, height: 823, deviceScaleFactor: 1.75, disabled: false }
        : { mobile: false, width: 1350, height: 940, deviceScaleFactor: 1, disabled: false },
      throttlingMethod: 'simulate',
    }
    const result = await Promise.race([
      lighthouse(input.url, flags),
      new Promise((_, reject) => {
        timeoutHandle = setTimeout(() => {
          void chrome.kill()
          reject(new LighthouseExecutionError('Lighthouse run timed out', 504))
        }, lighthouseTimeoutMilliseconds)
      }),
    ])
    if (!result?.lhr) {
      throw new LighthouseExecutionError('Lighthouse returned no report')
    }
    const remainingTimeoutMilliseconds = Math.max(
      1_000,
      config.timeoutMilliseconds - (Date.now() - startedAt),
    )
    const renderedDocument = await captureRenderedDocumentAuditResult({
      debuggingPort: chrome.port,
      input,
      timeoutMilliseconds: remainingTimeoutMilliseconds,
      settleMilliseconds: config.headingSettleMilliseconds,
      userAgent: String(result.lhr.configSettings?.emulatedUserAgent || '').trim(),
    })
    return normalizeLighthouseReport(
      result.lhr,
      renderedDocument.renderedHeadings,
      renderedDocument.renderedStructuredData,
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

async function captureRenderedDocumentAuditResult(options) {
  try {
    return await captureRenderedDocumentAudit(options)
  } catch (error) {
    const normalized = normalizeRenderedDocumentAuditError(error)
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
    }
  }
}

function normalizeRenderedDocumentAuditError(error) {
  const message = error instanceof Error ? error.message : String(error)
  return message.replace(/\s+/g, ' ').trim().slice(0, 500) || 'rendered document audit failed'
}

function normalizeLighthouseReport(lhr, renderedHeadings, renderedStructuredData) {
  return {
    lighthouseResult: {
      finalUrl: String(lhr.finalUrl || ''),
      lighthouseVersion: String(lhr.lighthouseVersion || ''),
      configSettings: lhr.configSettings || {},
      categories: lhr.categories || {},
      audits: lhr.audits || {},
      renderedHeadings,
      renderedStructuredData,
    },
  }
}

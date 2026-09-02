import { connect } from 'puppeteer-core'

import {
  headingViewport,
  settleRenderedDocument,
  snapshotRenderedHeadings,
} from './headings.mjs'
import {
  captureInteractionAudit,
  captureRenderedLinkAudit,
  captureSoftNavigationAudit,
} from './rendered_runtime_audits.mjs'
import { snapshotRenderedStructuredData } from './structured_data.mjs'

export async function captureRenderedDocumentAudit({
  debuggingPort,
  input,
  timeoutMilliseconds,
  settleMilliseconds,
  userAgent,
}) {
  const browser = await connect({
    browserURL: `http://127.0.0.1:${debuggingPort}`,
    defaultViewport: null,
  })
  let page
  try {
    page = await browser.newPage()
    await page.setViewport(headingViewport(input.strategy))
    if (userAgent) {
      await page.setUserAgent(userAgent)
    }

    const startedAt = Date.now()
    const navigationTimeout = Math.max(3_000, Math.min(timeoutMilliseconds, 25_000))
    await page.goto(input.url, {
      waitUntil: 'domcontentloaded',
      timeout: navigationTimeout,
    })

    const remaining = Math.max(0, timeoutMilliseconds - (Date.now() - startedAt))
    if (remaining < 1_000) {
      throw new Error('rendered document capture timed out before the document settled')
    }
    await settleRenderedDocument(page, Math.min(remaining, settleMilliseconds), {
      waitSelector: input.renderWaitSelector,
      waitTimeoutMilliseconds: input.renderWaitTimeoutMilliseconds,
      timeoutMilliseconds: remaining,
    })
    const [headings, structuredData] = await Promise.all([
      page.evaluate(snapshotRenderedHeadings),
      page.evaluate(snapshotRenderedStructuredData),
    ])
    const finalUrl = page.url()
    const renderedLinks = await captureRenderedLinkAudit(page, input)
    const interactionAudit = await captureInteractionAudit(page, input)
    if (shouldResetBeforeSoftNavigation(input)) {
      await page.goto(input.url, {
        waitUntil: 'domcontentloaded',
        timeout: navigationTimeout,
      })
      await settleRenderedDocument(page, Math.min(remaining, settleMilliseconds), {
        waitSelector: input.renderWaitSelector,
        waitTimeoutMilliseconds: input.renderWaitTimeoutMilliseconds,
        timeoutMilliseconds: remaining,
      }).catch(() => {})
    }
    const softNavigationAudit = await captureSoftNavigationAudit(page, input)

    return {
      renderedHeadings: {
        status: 'complete',
        source: 'chrome-rendered-dom',
        finalUrl,
        headings,
      },
      renderedStructuredData: {
        status: 'complete',
        source: 'chrome-rendered-dom',
        finalUrl,
        ...structuredData,
      },
      renderedLinks,
      interactionAudit,
      softNavigationAudit,
    }
  } finally {
    if (page) {
      await page.close().catch(() => {})
    }
    browser.disconnect()
  }
}

function shouldResetBeforeSoftNavigation(input) {
  const hasSoftNavigationTargets = (Array.isArray(input.softNavigationSelectors) && input.softNavigationSelectors.length > 0) ||
    Number(input.softNavigationMaxLinks || 0) > 0
  const hasInteractionProbes = Array.isArray(input.interactionProbes) && input.interactionProbes.length > 0
  return hasSoftNavigationTargets && hasInteractionProbes
}

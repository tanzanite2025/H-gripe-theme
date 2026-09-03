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
  runtimeAuditBudgetRemaining,
  runtimeAuditDeadlineReached,
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
    const deadlineAt = startedAt + Math.max(0, Number(timeoutMilliseconds || 0))
    const remainingTimeoutMilliseconds = () => Math.max(0, deadlineAt - Date.now())
    const initialNavigationBudget = remainingTimeoutMilliseconds()
    if (initialNavigationBudget < 500) {
      throw new Error('rendered document capture timed out before navigation started')
    }
    const navigationTimeout = Math.min(initialNavigationBudget, 25_000)
    await page.goto(input.url, {
      waitUntil: 'domcontentloaded',
      timeout: navigationTimeout,
    })

    const remaining = remainingTimeoutMilliseconds()
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
    const runtimeAuditOptions = { deadlineAt }
    const renderedLinks = await captureRenderedLinkAudit(page, input, runtimeAuditOptions)
    const interactionAudit = await captureInteractionAudit(page, input, runtimeAuditOptions)
    if (shouldResetBeforeSoftNavigation(input)) {
      if (!runtimeAuditDeadlineReached(runtimeAuditOptions)) {
        const resetTimeout = Math.min(navigationTimeout, runtimeAuditBudgetRemaining(runtimeAuditOptions))
        if (resetTimeout >= 500) {
          await page.goto(input.url, {
            waitUntil: 'domcontentloaded',
            timeout: resetTimeout,
          }).catch(() => {})
          const settleBudget = runtimeAuditBudgetRemaining(runtimeAuditOptions)
          if (settleBudget > 0) {
            await settleRenderedDocument(page, Math.min(settleBudget, settleMilliseconds), {
              waitSelector: input.renderWaitSelector,
              waitTimeoutMilliseconds: input.renderWaitTimeoutMilliseconds,
              timeoutMilliseconds: settleBudget,
            }).catch(() => {})
          }
        }
      }
    }
    const softNavigationAudit = await captureSoftNavigationAudit(page, input, runtimeAuditOptions)

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

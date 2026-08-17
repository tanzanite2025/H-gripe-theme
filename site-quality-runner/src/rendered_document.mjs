import { connect } from 'puppeteer-core'

import {
  headingViewport,
  settleRenderedDocument,
  snapshotRenderedHeadings,
} from './headings.mjs'
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
    await settleRenderedDocument(page, Math.min(remaining, settleMilliseconds))
    const [headings, structuredData] = await Promise.all([
      page.evaluate(snapshotRenderedHeadings),
      page.evaluate(snapshotRenderedStructuredData),
    ])
    const finalUrl = page.url()

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
    }
  } finally {
    if (page) {
      await page.close().catch(() => {})
    }
    browser.disconnect()
  }
}

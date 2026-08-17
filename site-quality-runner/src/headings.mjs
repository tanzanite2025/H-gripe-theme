import { connect } from 'puppeteer-core'

const maxHeadingRecords = 200
const maxHeadingTextLength = 180
const maxSelectorDepth = 12

export function headingViewport(strategy) {
  return strategy === 'mobile'
    ? { width: 412, height: 823, deviceScaleFactor: 1.75, isMobile: true, hasTouch: true }
    : { width: 1350, height: 940, deviceScaleFactor: 1, isMobile: false, hasTouch: false }
}

export function normalizeHeadingText(value) {
  const normalized = String(value || '').replace(/\s+/g, ' ').trim()
  if (normalized.length <= maxHeadingTextLength) {
    return normalized
  }
  return `${normalized.slice(0, maxHeadingTextLength - 3)}...`
}

export async function captureRenderedHeadingAudit({
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
      throw new Error('rendered heading capture timed out before the document settled')
    }
    await settleRenderedDocument(page, Math.min(remaining, settleMilliseconds))
    const headings = await page.evaluate(snapshotRenderedHeadings)

    return {
      status: 'complete',
      source: 'chrome-rendered-dom',
      finalUrl: page.url(),
      headings,
    }
  } finally {
    if (page) {
      await page.close().catch(() => {})
    }
    browser.disconnect()
  }
}

export async function settleRenderedDocument(page, settleMilliseconds) {
  const idleBudget = Math.max(500, Math.min(settleMilliseconds, 3_000))
  await page.waitForNetworkIdle({
    idleTime: Math.min(750, idleBudget),
    concurrency: 2,
    timeout: idleBudget,
  }).catch(() => {})

  await page.evaluate(async (delay) => {
    if (document.fonts?.ready) {
      await document.fonts.ready.catch(() => {})
    }
    await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve)))
    if (delay > 0) {
      await new Promise((resolve) => window.setTimeout(resolve, delay))
    }
  }, Math.max(0, settleMilliseconds - idleBudget))
}

export function snapshotRenderedHeadings() {
  const maxHeadingRecords = 200
  const maxHeadingTextLength = 180
  const maxSelectorDepth = 12
  const headings = []
  const seen = new Set()

  const pushHeading = (element) => {
    if (headings.length >= maxHeadingRecords || !isRenderedHeading(element)) {
      return
    }
    const level = Number.parseInt(element.tagName.slice(1), 10)
    const text = normalizeText(element.textContent)
    if (!Number.isInteger(level) || level < 1 || level > 6 || !text) {
      return
    }
    const selector = headingSelector(element)
    const key = `${level}\u0000${selector}\u0000${text}`
    if (seen.has(key)) {
      return
    }
    seen.add(key)
    headings.push({
      level,
      text,
      snippet: headingSnippet(element, text),
      selector,
    })
  }

  const walk = (root) => {
    for (const element of root.children) {
      const tagName = element.tagName.toLowerCase()
      if (tagName === 'script' || tagName === 'style' || tagName === 'noscript' || tagName === 'template') {
        continue
      }
      if (/^h[1-6]$/.test(tagName)) {
        pushHeading(element)
      }
      if (element.shadowRoot?.mode === 'open') {
        walk(element.shadowRoot)
      }
      walk(element)
    }
  }

  walk(document.body || document.documentElement)
  return headings

  function isRenderedHeading(element) {
    for (let current = element; current; current = headingParent(current)) {
      if (current.hidden || current.inert || current.getAttribute('aria-hidden') === 'true') {
        return false
      }
      const style = window.getComputedStyle(current)
      if (
        style.display === 'none' ||
        style.visibility === 'hidden' ||
        style.visibility === 'collapse' ||
        style.contentVisibility === 'hidden'
      ) {
        return false
      }
      if (current.tagName === 'DETAILS' && !current.open && !insideDetailsSummary(element, current)) {
        return false
      }
    }
    return true
  }

  function headingParent(element) {
    if (element.parentElement) {
      return element.parentElement
    }
    const root = element.getRootNode()
    return root instanceof ShadowRoot ? root.host : null
  }

  function insideDetailsSummary(element, details) {
    const summary = Array.from(details.children).find((child) => child.tagName === 'SUMMARY')
    return Boolean(summary?.contains(element))
  }

  function normalizeText(value) {
    const normalized = String(value || '').replace(/\s+/g, ' ').trim()
    if (normalized.length <= maxHeadingTextLength) {
      return normalized
    }
    return `${normalized.slice(0, maxHeadingTextLength - 3)}...`
  }

  function headingSnippet(element, text) {
    const tagName = element.tagName.toLowerCase()
    const attributes = []
    if (element.id) {
      attributes.push(`id="${escapeHTMLAttribute(element.id)}"`)
    }
    if (element.classList.length > 0) {
      attributes.push(`class="${escapeHTMLAttribute(Array.from(element.classList).join(' '))}"`)
    }
    const openTag = attributes.length > 0
      ? `<${tagName} ${attributes.join(' ')}>`
      : `<${tagName}>`
    return `${openTag}${escapeHTMLText(text)}</${tagName}>`
  }

  function headingSelector(element) {
    const parts = []
    let current = element
    while (current && parts.length < maxSelectorDepth) {
      const root = current.getRootNode()
      parts.unshift(selectorPart(current))
      current = headingParent(current)
      if (root instanceof ShadowRoot) {
        parts.unshift('#shadow-root')
      }
    }
    return parts.join(' > ')
  }

  function selectorPart(element) {
    const tagName = element.tagName.toLowerCase()
    if (element.id) {
      return `${tagName}#${cssIdentifier(element.id)}`
    }
    let index = 1
    let sibling = element.previousElementSibling
    while (sibling) {
      if (sibling.tagName === element.tagName) {
        index++
      }
      sibling = sibling.previousElementSibling
    }
    return `${tagName}:nth-of-type(${index})`
  }

  function cssIdentifier(value) {
    return String(value || '').trim().replace(/[^a-zA-Z0-9_-]/g, '-') || 'unknown'
  }

  function escapeHTMLText(value) {
    return String(value)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
  }

  function escapeHTMLAttribute(value) {
    return escapeHTMLText(value).replace(/"/g, '&quot;')
  }
}

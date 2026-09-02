import { setTimeout as delay } from 'node:timers/promises'

import { settleRenderedDocument } from './headings.mjs'

export const resourceBudgetAuditID = 'site-resource-budget'

const linkProbeUserAgent = 'TANZANITE-SiteQualityLinks/1.0'
const maxRenderedLinks = 250
const maxInteractionProbes = 8
const maxSoftNavigationTargets = 8

export function applyResourceBudgetAudit(lhr, options = {}) {
  if (!lhr || typeof lhr !== 'object') {
    return null
  }
  const items = resourceBudgetViolationsFromLighthouse(lhr, options)
  if (items.length === 0) {
    return null
  }
  const totalOverBudgetBytes = items.reduce((sum, item) => sum + Math.max(0, item.overBudgetBytes || 0), 0)
  lhr.audits = lhr.audits && typeof lhr.audits === 'object' ? lhr.audits : {}
  lhr.audits[resourceBudgetAuditID] = {
    id: resourceBudgetAuditID,
    title: 'Resource exceeds configured performance budget',
    description: 'Static resource budgets catch oversized JavaScript and image payloads before they appear only as lower Lighthouse scores.',
    score: 0,
    scoreDisplayMode: 'numeric',
    displayValue: `${items.length} resource${items.length === 1 ? '' : 's'} over budget`,
    numericValue: items.length,
    details: {
      type: 'table',
      overallSavingsBytes: totalOverBudgetBytes,
      headings: [
        { key: 'url', itemType: 'url', text: 'URL' },
        { key: 'resourceType', itemType: 'text', text: 'Type' },
        { key: 'totalBytes', itemType: 'bytes', text: 'Size' },
        { key: 'budgetBytes', itemType: 'bytes', text: 'Budget' },
        { key: 'overBudgetBytes', itemType: 'bytes', text: 'Over budget' },
      ],
      items,
    },
  }
  return lhr.audits[resourceBudgetAuditID]
}

export function resourceBudgetViolationsFromLighthouse(lhr, options = {}) {
  const jsBudgetBytes = normalizeBudgetBytes(options.jsBudgetBytes)
  const imageBudgetBytes = normalizeBudgetBytes(options.imageBudgetBytes)
  const audits = lhr?.audits && typeof lhr.audits === 'object' ? lhr.audits : {}
  const requests = Array.isArray(audits['network-requests']?.details?.items)
    ? audits['network-requests'].details.items
    : []
  const violations = []
  for (const request of requests) {
    const url = String(request?.url || '').trim()
    if (!url || url.startsWith('data:') || url.startsWith('blob:')) {
      continue
    }
    const totalBytes = firstFiniteNumber(request?.resourceSize, request?.transferSize, request?.encodedDataLength)
    if (!Number.isFinite(totalBytes) || totalBytes <= 0) {
      continue
    }
    const kind = classifyBudgetResource(request)
    const budgetBytes = kind === 'script'
      ? jsBudgetBytes
      : kind === 'image'
        ? imageBudgetBytes
        : 0
    if (budgetBytes <= 0 || totalBytes <= budgetBytes) {
      continue
    }
    violations.push({
      url,
      resourceType: kind,
      totalBytes: Math.round(totalBytes),
      budgetBytes,
      overBudgetBytes: Math.round(totalBytes - budgetBytes),
    })
  }
  violations.sort((a, b) => b.overBudgetBytes - a.overBudgetBytes || a.url.localeCompare(b.url))
  return violations.slice(0, 25)
}

export async function captureRenderedLinkAudit(page, input = {}) {
  if (!input.linkCheckEnabled || normalizeCount(input.linkCheckMaxLinks, 0, maxRenderedLinks) <= 0) {
    return {
      status: 'skipped',
      source: 'chrome-rendered-dom',
      configured: false,
      reason: 'link checking is disabled',
      links: [],
    }
  }
  try {
    const links = await page.evaluate(snapshotRenderedLinks, {
      baseURL: input.url,
      includeExternal: Boolean(input.linkCheckExternal),
      maxLinks: normalizeCount(input.linkCheckMaxLinks, 80, maxRenderedLinks),
    })
    const checked = await mapLimited(
      links,
      4,
      (link) => probeRenderedLink(link, {
        allowedOrigin: new URL(input.url).origin,
        allowExternal: Boolean(input.linkCheckExternal),
        timeoutMilliseconds: input.linkCheckTimeoutMilliseconds,
        maxRedirects: input.linkCheckMaxRedirects,
      }),
    )
    return {
      status: 'complete',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      links: checked,
    }
  } catch (error) {
    return {
      status: 'failed',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      error: normalizeAuditError(error),
      links: [],
    }
  }
}

export async function captureInteractionAudit(page, input = {}) {
  const probes = Array.isArray(input.interactionProbes) ? input.interactionProbes.slice(0, maxInteractionProbes) : []
  if (probes.length === 0) {
    return {
      status: 'skipped',
      source: 'chrome-rendered-dom',
      configured: false,
      reason: 'no interaction probes configured',
      interactions: [],
    }
  }
  try {
    await installInteractionObserver(page)
    const interactions = []
    for (const probe of probes) {
      interactions.push(await runInteractionProbe(page, probe, {
        defaultThresholdMilliseconds: input.interactionMaxResponseMilliseconds,
      }))
    }
    return {
      status: 'complete',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      interactions,
    }
  } catch (error) {
    return {
      status: 'failed',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      error: normalizeAuditError(error),
      interactions: [],
    }
  }
}

export async function captureSoftNavigationAudit(page, input = {}) {
  const selectors = Array.isArray(input.softNavigationSelectors) ? input.softNavigationSelectors : []
  const maxLinks = normalizeCount(input.softNavigationMaxLinks, selectors.length > 0 ? selectors.length : 0, maxSoftNavigationTargets)
  if (maxLinks <= 0 && selectors.length === 0) {
    return {
      status: 'skipped',
      source: 'chrome-rendered-dom',
      configured: false,
      reason: 'no soft navigation targets configured',
      navigations: [],
    }
  }
  try {
    const originURL = page.url()
    const targets = await page.evaluate(snapshotSoftNavigationTargets, {
      selectors,
      maxLinks: Math.max(maxLinks, selectors.length),
    })
    const navigations = []
    for (const target of targets.slice(0, maxSoftNavigationTargets)) {
      if (page.url() !== originURL) {
        await page.goto(originURL, { waitUntil: 'domcontentloaded', timeout: 10_000 })
        await settleRenderedDocument(page, 1_000).catch(() => {})
      }
      navigations.push(await runSoftNavigationProbe(page, target, input))
    }
    return {
      status: 'complete',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      navigations,
    }
  } catch (error) {
    return {
      status: 'failed',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      error: normalizeAuditError(error),
      navigations: [],
    }
  }
}

async function probeRenderedLink(link, options) {
  const probe = await fetchWithRedirects(link.href, {
    method: 'HEAD',
    allowedOrigin: options.allowedOrigin,
    allowExternal: options.allowExternal,
    timeoutMilliseconds: options.timeoutMilliseconds,
    maxRedirects: options.maxRedirects,
  })
  const fallbackNeeded = probe.statusCode === 405 || probe.statusCode === 501
  const result = fallbackNeeded
    ? await fetchWithRedirects(link.href, {
      method: 'GET',
      allowedOrigin: options.allowedOrigin,
      allowExternal: options.allowExternal,
      timeoutMilliseconds: options.timeoutMilliseconds,
      maxRedirects: options.maxRedirects,
      rangeRequest: true,
    })
    : probe
  return {
    ...link,
    ...result,
    ok: Boolean(result.ok),
  }
}

async function fetchWithRedirects(url, options) {
  const maxRedirects = normalizeCount(options.maxRedirects, 5, 10)
  const timeoutMilliseconds = normalizeCount(options.timeoutMilliseconds, 2_500, 15_000)
  let currentURL = String(url || '').trim()
  let redirectCount = 0
  try {
    for (;;) {
      const controller = new AbortController()
      const timeoutHandle = setTimeout(() => controller.abort(), timeoutMilliseconds)
      let response
      try {
        response = await fetch(currentURL, {
          method: options.method,
          redirect: 'manual',
          signal: controller.signal,
          headers: {
            Accept: 'text/html,application/xhtml+xml,application/json,*/*;q=0.8',
            'User-Agent': linkProbeUserAgent,
            ...(options.rangeRequest ? { Range: 'bytes=0-0' } : {}),
          },
        })
      } finally {
        clearTimeout(timeoutHandle)
      }
      const location = response.headers.get('location')
      if (response.status >= 300 && response.status < 400 && location) {
        redirectCount++
        const nextURL = new URL(location, currentURL)
        if (!options.allowExternal && options.allowedOrigin && nextURL.origin !== options.allowedOrigin) {
          return {
            statusCode: response.status,
            finalUrl: nextURL.toString(),
            redirected: true,
            redirectCount,
            ok: false,
            error: 'redirected outside the configured storefront origin',
          }
        }
        if (redirectCount > maxRedirects) {
          return {
            statusCode: response.status,
            finalUrl: currentURL,
            redirected: true,
            redirectCount,
            ok: false,
            error: `redirect chain exceeded ${maxRedirects}`,
          }
        }
        currentURL = nextURL.toString()
        continue
      }
      return {
        statusCode: response.status,
        finalUrl: currentURL,
        redirected: redirectCount > 0,
        redirectCount,
        ok: response.status < 400,
      }
    }
  } catch (error) {
    return {
      statusCode: 0,
      finalUrl: currentURL,
      redirected: redirectCount > 0,
      redirectCount,
      ok: false,
      error: normalizeAuditError(error),
    }
  }
}

async function installInteractionObserver(page) {
  await page.evaluate(() => {
    window.__siteQualityInteractionEvents = []
    if (window.__siteQualityInteractionObserver || typeof PerformanceObserver !== 'function') {
      return
    }
    try {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          window.__siteQualityInteractionEvents.push({
            name: entry.name,
            startTime: entry.startTime,
            duration: entry.duration,
            processingStart: entry.processingStart,
            processingEnd: entry.processingEnd,
            interactionId: entry.interactionId || 0,
          })
        }
      })
      observer.observe({ type: 'event', buffered: true, durationThreshold: 16 })
      window.__siteQualityInteractionObserver = observer
    } catch {
      window.__siteQualityInteractionObserver = null
    }
  })
}

async function runInteractionProbe(page, probe, options) {
  const selector = String(probe?.selector || '').trim()
  const name = String(probe?.name || selector || 'interaction').trim().slice(0, 80)
  const action = normalizeInteractionAction(probe?.action)
  const thresholdMilliseconds = normalizeThreshold(
    probe?.maxResponseMilliseconds,
    options.defaultThresholdMilliseconds,
    200,
  )
  if (!selector) {
    return {
      name,
      selector,
      action,
      status: 'failed',
      thresholdMilliseconds,
      error: 'interaction selector is required',
    }
  }
  const handle = await page.$(selector)
  if (!handle) {
    return {
      name,
      selector,
      action,
      status: 'failed',
      thresholdMilliseconds,
      error: 'interaction selector was not found',
    }
  }
  try {
    await handle.evaluate((element) => {
      element.scrollIntoView({ block: 'center', inline: 'center', behavior: 'instant' })
    })
    const marker = await page.evaluate(() => performance.now())
    const startedAt = Date.now()
    await performInteractionAction(page, handle, selector, probe, action)
    await settleAfterInteraction(page)
    const elapsedMilliseconds = Date.now() - startedAt
    const eventTiming = await page.evaluate((startTime) => {
      const events = (window.__siteQualityInteractionEvents || [])
        .filter((entry) => Number(entry.startTime) >= startTime - 50)
        .sort((a, b) => Number(b.duration || 0) - Number(a.duration || 0))
      return events[0] || null
    }, marker)
    const eventDuration = Number(eventTiming?.duration)
    const responseMilliseconds = Number.isFinite(eventDuration) && eventDuration > 0
      ? eventDuration
      : elapsedMilliseconds
    return {
      name,
      selector,
      action,
      status: 'complete',
      responseMilliseconds: Math.round(responseMilliseconds),
      thresholdMilliseconds,
      metricSource: Number.isFinite(eventDuration) && eventDuration > 0 ? 'event-timing' : 'raf-latency',
      eventName: String(eventTiming?.name || ''),
      interactionId: Number(eventTiming?.interactionId || 0),
      exceeded: responseMilliseconds > thresholdMilliseconds,
    }
  } catch (error) {
    return {
      name,
      selector,
      action,
      status: 'failed',
      thresholdMilliseconds,
      error: normalizeAuditError(error),
    }
  } finally {
    await handle.dispose().catch(() => {})
  }
}

async function performInteractionAction(page, handle, selector, probe, action) {
  if (action === 'hover') {
    await handle.hover()
    return
  }
  if (action === 'type') {
    await handle.click({ delay: 20 })
    await page.keyboard.type(String(probe?.text || 'test'), { delay: 10 })
    return
  }
  if (action === 'press') {
    await handle.click({ delay: 20 })
    await page.keyboard.press(String(probe?.key || 'Enter'))
    return
  }
  if (action === 'drag') {
    const box = await handle.boundingBox()
    if (!box) {
      throw new Error('interaction target is not visible')
    }
    const startX = box.x + box.width / 2
    const startY = box.y + box.height / 2
    const deltaX = Number.isFinite(Number(probe?.deltaX)) ? Number(probe.deltaX) : Math.min(80, Math.max(20, box.width / 3))
    const deltaY = Number.isFinite(Number(probe?.deltaY)) ? Number(probe.deltaY) : 0
    await page.mouse.move(startX, startY)
    await page.mouse.down()
    await page.mouse.move(startX + deltaX, startY + deltaY, { steps: 8 })
    await page.mouse.up()
    return
  }
  await page.click(selector, { delay: 20 })
}

async function settleAfterInteraction(page) {
  await page.evaluate(() => new Promise((resolve) => {
    requestAnimationFrame(() => requestAnimationFrame(() => window.setTimeout(resolve, 0)))
  })).catch(() => {})
  await delay(100)
}

async function runSoftNavigationProbe(page, target, input) {
  const thresholdMilliseconds = normalizeThreshold(input.softNavigationMaxDurationMilliseconds, undefined, 2_000)
  const heapThresholdMB = normalizeThreshold(input.softNavigationMaxHeapGrowthMB, undefined, 32)
  const startedURL = page.url()
  const marker = `site-quality-soft-nav-${Date.now()}-${Math.random()}`
  try {
    await page.evaluate((value) => {
      window.__siteQualitySoftNavigationMarker = value
    }, marker)
    const beforeHeap = await page.evaluate(() => performance.memory?.usedJSHeapSize || 0).catch(() => 0)
    const startedAt = Date.now()
    await page.click(target.selector, { delay: 20 })
    await Promise.race([
      page.waitForFunction((previousURL) => location.href !== previousURL, { timeout: Math.max(500, thresholdMilliseconds + 2_000) }, startedURL),
      delay(thresholdMilliseconds + 2_000),
    ])
    await settleRenderedDocument(page, 1_000).catch(() => {})
    const durationMilliseconds = Date.now() - startedAt
    const afterHeap = await page.evaluate(() => performance.memory?.usedJSHeapSize || 0).catch(() => 0)
    const markerStillPresent = await page.evaluate((value) => window.__siteQualitySoftNavigationMarker === value, marker).catch(() => false)
    const finalURL = page.url()
    const heapDeltaBytes = beforeHeap > 0 && afterHeap > 0 ? afterHeap - beforeHeap : 0
    return {
      fromUrl: startedURL,
      toUrl: finalURL,
      expectedUrl: target.href,
      selector: target.selector,
      text: target.text,
      status: finalURL !== startedURL ? 'complete' : 'failed',
      mode: markerStillPresent ? 'soft-navigation' : 'hard-navigation',
      durationMilliseconds,
      thresholdMilliseconds,
      jsHeapDeltaBytes: heapDeltaBytes,
      jsHeapDeltaThresholdBytes: Math.round(heapThresholdMB * 1024 * 1024),
      exceeded: durationMilliseconds > thresholdMilliseconds ||
        (heapDeltaBytes > 0 && heapDeltaBytes > heapThresholdMB * 1024 * 1024),
      error: finalURL !== startedURL ? '' : 'route did not change after clicking navigation target',
    }
  } catch (error) {
    return {
      fromUrl: startedURL,
      toUrl: page.url(),
      expectedUrl: target.href,
      selector: target.selector,
      text: target.text,
      status: 'failed',
      thresholdMilliseconds,
      error: normalizeAuditError(error),
    }
  }
}

function snapshotRenderedLinks(options) {
  const base = new URL(options.baseURL || location.href, location.href)
  const seen = new Set()
  const links = []
  for (const anchor of document.querySelectorAll('a[href]')) {
    if (links.length >= options.maxLinks) {
      break
    }
    const raw = anchor.getAttribute('href') || ''
    if (!raw || raw.trim().startsWith('#')) {
      continue
    }
    let href
    try {
      href = new URL(raw, location.href)
    } catch {
      continue
    }
    if (!['http:', 'https:'].includes(href.protocol)) {
      continue
    }
    href.hash = ''
    if (!options.includeExternal && href.origin !== base.origin) {
      continue
    }
    const key = href.toString()
    if (seen.has(key)) {
      continue
    }
    seen.add(key)
    links.push({
      href: key,
      text: normalizeRenderedText(anchor.textContent || anchor.getAttribute('aria-label') || anchor.title || ''),
      textLang: anchor.lang || '',
      selector: elementSelector(anchor),
    })
  }
  return links
}

function snapshotSoftNavigationTargets(options) {
  const selectors = Array.isArray(options.selectors) ? options.selectors.map((item) => String(item || '').trim()).filter(Boolean) : []
  const maxLinks = Math.max(0, Number(options.maxLinks || 0))
  const anchors = []
  if (selectors.length > 0) {
    for (const selector of selectors) {
      for (const element of document.querySelectorAll(selector)) {
        const anchor = element.matches('a[href]') ? element : element.closest('a[href]')
        if (anchor) {
          anchors.push(anchor)
        }
      }
    }
  } else if (maxLinks > 0) {
    anchors.push(...document.querySelectorAll('nav a[href], main a[href], a[href]'))
  }

  const seen = new Set()
  const targets = []
  for (const anchor of anchors) {
    if (targets.length >= Math.max(maxLinks, selectors.length)) {
      break
    }
    if (anchor.target && anchor.target !== '_self') {
      continue
    }
    if (anchor.hasAttribute('download')) {
      continue
    }
    let href
    try {
      href = new URL(anchor.getAttribute('href') || '', location.href)
    } catch {
      continue
    }
    href.hash = ''
    if (!['http:', 'https:'].includes(href.protocol) || href.origin !== location.origin) {
      continue
    }
    if (href.href === location.href) {
      continue
    }
    const key = href.href
    if (seen.has(key)) {
      continue
    }
    seen.add(key)
    targets.push({
      href: key,
      text: normalizeRenderedText(anchor.textContent || anchor.getAttribute('aria-label') || anchor.title || ''),
      selector: elementSelector(anchor),
    })
  }
  return targets
}

function normalizeRenderedText(value) {
  const normalized = String(value || '').replace(/\s+/g, ' ').trim()
  return normalized.length > 120 ? `${normalized.slice(0, 117)}...` : normalized
}

function elementSelector(element) {
  const parts = []
  let current = element
  while (current && current.nodeType === Node.ELEMENT_NODE && parts.length < 8) {
    const tag = current.tagName.toLowerCase()
    if (current.id) {
      parts.unshift(`${tag}#${cssIdentifier(current.id)}`)
      break
    }
    let index = 1
    let sibling = current.previousElementSibling
    while (sibling) {
      if (sibling.tagName === current.tagName) {
        index++
      }
      sibling = sibling.previousElementSibling
    }
    parts.unshift(`${tag}:nth-of-type(${index})`)
    current = current.parentElement
  }
  return parts.join(' > ')
}

function cssIdentifier(value) {
  return String(value || '').trim().replace(/[^a-zA-Z0-9_-]/g, '-') || 'unknown'
}

async function mapLimited(items, limit, mapper) {
  const results = new Array(items.length)
  let next = 0
  const workers = Array.from({ length: Math.max(1, Math.min(limit, items.length)) }, async () => {
    for (;;) {
      const index = next++
      if (index >= items.length) {
        return
      }
      results[index] = await mapper(items[index], index)
    }
  })
  await Promise.all(workers)
  return results
}

function classifyBudgetResource(request) {
  const resourceType = String(request?.resourceType || request?.resourceTypeDisplay || '').toLowerCase()
  const mimeType = String(request?.mimeType || '').toLowerCase()
  const url = String(request?.url || '').toLowerCase()
  if (resourceType.includes('script') || mimeType.includes('javascript') || /\.m?js(\?|$)/.test(url)) {
    return 'script'
  }
  if (resourceType.includes('image') || mimeType.startsWith('image/') || /\.(avif|gif|jpe?g|png|svg|webp)(\?|$)/.test(url)) {
    return 'image'
  }
  return ''
}

function firstFiniteNumber(...values) {
  for (const value of values) {
    const numeric = Number(value)
    if (Number.isFinite(numeric)) {
      return numeric
    }
  }
  return null
}

function normalizeBudgetBytes(value) {
  const parsed = Number.parseInt(String(value || ''), 10)
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return 0
  }
  return parsed
}

function normalizeCount(value, fallback, maximum) {
  const parsed = Number.parseInt(String(value || ''), 10)
  if (!Number.isInteger(parsed)) {
    return fallback
  }
  return Math.max(0, Math.min(parsed, maximum))
}

function normalizeThreshold(value, fallback, defaultValue) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return Number.isFinite(Number(fallback)) && Number(fallback) > 0 ? Number(fallback) : defaultValue
  }
  return parsed
}

function normalizeInteractionAction(value) {
  const action = String(value || '').trim().toLowerCase()
  return ['click', 'hover', 'type', 'press', 'drag'].includes(action) ? action : 'click'
}

function normalizeAuditError(error) {
  const message = error instanceof Error ? error.message : String(error)
  return message.replace(/\s+/g, ' ').trim().slice(0, 500) || 'runtime audit failed'
}

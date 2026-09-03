import { setTimeout as delay } from 'node:timers/promises'

import { settleRenderedDocument } from './headings.mjs'

export const resourceBudgetAuditID = 'site-resource-budget'

const linkProbeUserAgent = 'TANZANITE-SiteQualityLinks/1.0'
const maxRenderedLinks = 250
const maxInteractionProbes = 8
const maxSoftNavigationTargets = 8
export const runtimeAuditDeadlineReserveMilliseconds = 5_000

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
    const totalBytes = firstFiniteNumber(request?.transferSize, request?.encodedDataLength, request?.resourceSize)
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

export async function captureRenderedLinkAudit(page, input = {}, options = {}) {
  if (!input.linkCheckEnabled || normalizeCount(input.linkCheckMaxLinks, 0, maxRenderedLinks) <= 0) {
    return {
      status: 'skipped',
      source: 'chrome-rendered-dom',
      configured: false,
      reason: 'link checking is disabled',
      links: [],
    }
  }
  if (runtimeAuditDeadlineReached(options)) {
    return timeoutSkippedAudit(page, {
      configured: true,
      collectionKey: 'links',
      error: 'link checking skipped because the rendered audit deadline was reached',
    })
  }
  try {
    const links = await page.evaluate(snapshotRenderedLinks, {
      baseURL: input.url,
      includeExternal: Boolean(input.linkCheckExternal),
      maxLinks: normalizeCount(input.linkCheckMaxLinks, 80, maxRenderedLinks),
    })
    if (runtimeAuditDeadlineReached(options)) {
      return timeoutSkippedAudit(page, {
        configured: true,
        collectionKey: 'links',
        error: 'link checking skipped after DOM link discovery because the rendered audit deadline was reached',
        linkCount: links.length,
        checkedLinkCount: 0,
        skippedLinkCount: links.length,
      })
    }
    const checked = await mapLimited(
      links,
      4,
      (link) => probeRenderedLink(link, {
        allowedOrigin: new URL(input.url).origin,
        allowExternal: Boolean(input.linkCheckExternal),
        timeoutMilliseconds: input.linkCheckTimeoutMilliseconds,
        maxRedirects: input.linkCheckMaxRedirects,
        deadlineAt: options.deadlineAt,
      }),
      options,
    )
    const timeoutSkipped = checked.timedOut || checked.results.some((link) => link?.timeoutSkipped)
    return {
      status: timeoutSkipped ? 'timeout_skipped' : 'complete',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      ...(timeoutSkipped ? { error: 'link checking stopped early because the rendered audit deadline was reached' } : {}),
      linkCount: links.length,
      checkedLinkCount: checked.results.length,
      skippedLinkCount: checked.skippedCount,
      links: checked.results,
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

export async function captureInteractionAudit(page, input = {}, options = {}) {
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
  if (runtimeAuditDeadlineReached(options)) {
    return timeoutSkippedAudit(page, {
      configured: true,
      collectionKey: 'interactions',
      error: 'interaction audit skipped because the rendered audit deadline was reached',
    })
  }
  try {
    await installInteractionObserver(page)
    const interactions = []
    let timeoutSkipped = runtimeAuditDeadlineReached(options)
    for (const probe of probes) {
      if (runtimeAuditDeadlineReached(options)) {
        timeoutSkipped = true
        break
      }
      const result = await runInteractionProbe(page, probe, {
        defaultThresholdMilliseconds: input.interactionMaxResponseMilliseconds,
        deadlineAt: options.deadlineAt,
      })
      interactions.push(result)
      if (result.status === 'timeout_skipped') {
        timeoutSkipped = true
        break
      }
    }
    return {
      status: timeoutSkipped ? 'timeout_skipped' : 'complete',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      ...(timeoutSkipped ? { error: 'interaction audit stopped early because the rendered audit deadline was reached' } : {}),
      probeCount: probes.length,
      checkedProbeCount: interactions.length,
      skippedProbeCount: Math.max(0, probes.length - interactions.length),
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

export async function captureSoftNavigationAudit(page, input = {}, options = {}) {
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
  if (runtimeAuditDeadlineReached(options)) {
    return timeoutSkippedAudit(page, {
      configured: true,
      collectionKey: 'navigations',
      error: 'soft navigation audit skipped because the rendered audit deadline was reached',
    })
  }
  try {
    const originURL = page.url()
    const targets = await page.evaluate(snapshotSoftNavigationTargets, {
      selectors,
      maxLinks: Math.max(maxLinks, selectors.length),
    })
    const navigations = []
    let timeoutSkipped = runtimeAuditDeadlineReached(options)
    for (const target of targets.slice(0, maxSoftNavigationTargets)) {
      if (runtimeAuditDeadlineReached(options)) {
        timeoutSkipped = true
        break
      }
      if (page.url() !== originURL) {
        const resetTimeout = Math.min(10_000, runtimeAuditBudgetRemaining(options))
        if (resetTimeout < 500) {
          timeoutSkipped = true
          break
        }
        await page.goto(originURL, { waitUntil: 'domcontentloaded', timeout: resetTimeout })
          .catch(() => {})
        const settleBudget = runtimeAuditBudgetRemaining(options)
        if (settleBudget <= 0) {
          timeoutSkipped = true
          break
        }
        await settleRenderedDocument(page, Math.min(1_000, settleBudget), {
          timeoutMilliseconds: settleBudget,
        }).catch(() => {})
      }
      const result = await runSoftNavigationProbe(page, target, input, options)
      navigations.push(result)
      if (result.status === 'timeout_skipped') {
        timeoutSkipped = true
        break
      }
    }
    return {
      status: timeoutSkipped ? 'timeout_skipped' : 'complete',
      source: 'chrome-rendered-dom',
      configured: true,
      finalUrl: page.url(),
      ...(timeoutSkipped ? { error: 'soft navigation audit stopped early because the rendered audit deadline was reached' } : {}),
      targetCount: targets.length,
      checkedTargetCount: navigations.length,
      skippedTargetCount: Math.max(0, Math.min(targets.length, maxSoftNavigationTargets) - navigations.length),
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
  if (runtimeAuditDeadlineReached(options)) {
    return timeoutSkippedLinkResult(link, 'link probe skipped because the rendered audit deadline was reached')
  }
  const probe = await fetchWithRedirects(link.href, {
    method: 'HEAD',
    allowedOrigin: options.allowedOrigin,
    allowExternal: options.allowExternal,
    timeoutMilliseconds: options.timeoutMilliseconds,
    maxRedirects: options.maxRedirects,
    deadlineAt: options.deadlineAt,
  })
  const fallbackNeeded = probe.statusCode === 405 || probe.statusCode === 501
  if (fallbackNeeded && runtimeAuditDeadlineReached(options)) {
    return {
      ...link,
      ...probe,
      ok: false,
      timeoutSkipped: true,
      error: 'GET fallback skipped because the rendered audit deadline was reached',
    }
  }
  const result = fallbackNeeded
    ? await fetchWithRedirects(link.href, {
      method: 'GET',
      allowedOrigin: options.allowedOrigin,
      allowExternal: options.allowExternal,
      timeoutMilliseconds: options.timeoutMilliseconds,
      maxRedirects: options.maxRedirects,
      rangeRequest: true,
      deadlineAt: options.deadlineAt,
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
      const availableTimeoutMilliseconds = runtimeAuditBudgetRemaining(options)
      if (availableTimeoutMilliseconds <= 0) {
        return {
          statusCode: 0,
          finalUrl: currentURL,
          redirected: redirectCount > 0,
          redirectCount,
          ok: false,
          timeoutSkipped: true,
          error: 'link probe stopped because the rendered audit deadline was reached',
        }
      }
      const controller = new AbortController()
      const requestTimeoutMilliseconds = Math.max(1, Math.min(timeoutMilliseconds, availableTimeoutMilliseconds))
      const timeoutHandle = setTimeout(() => controller.abort(), requestTimeoutMilliseconds)
      let response
      try {
        try {
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
          } catch (error) {
            if (isAbortError(error) && (requestTimeoutMilliseconds < timeoutMilliseconds || runtimeAuditDeadlineReached(options))) {
              return linkDeadlineSkippedResult(currentURL, redirectCount)
            }
            throw error
          }
        } finally {
          clearTimeout(timeoutHandle)
        }
        const statusCode = response.status
        const location = response.headers.get('location')
        if (statusCode >= 300 && statusCode < 400 && location) {
          redirectCount++
          const nextURL = new URL(location, currentURL)
          if (!options.allowExternal && options.allowedOrigin && nextURL.origin !== options.allowedOrigin) {
            return {
              statusCode,
              finalUrl: nextURL.toString(),
              redirected: true,
              redirectCount,
              ok: false,
              error: 'redirected outside the configured storefront origin',
            }
          }
          if (redirectCount > maxRedirects) {
            return {
              statusCode,
              finalUrl: currentURL,
              redirected: true,
              redirectCount,
              ok: false,
              error: `redirect chain exceeded ${maxRedirects}`,
            }
          }
          if (runtimeAuditDeadlineReached(options)) {
            return {
              statusCode,
              finalUrl: currentURL,
              redirected: true,
              redirectCount,
              ok: false,
              timeoutSkipped: true,
              error: 'redirect follow-up skipped because the rendered audit deadline was reached',
            }
          }
          currentURL = nextURL.toString()
          continue
        }
        return {
          statusCode,
          finalUrl: currentURL,
          redirected: redirectCount > 0,
          redirectCount,
          ok: statusCode < 400,
        }
      } finally {
        await discardResponseBody(response)
      }
    }
  } catch (error) {
    if (isAbortError(error) && runtimeAuditDeadlineReached(options)) {
      return linkDeadlineSkippedResult(currentURL, redirectCount)
    }
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

function linkDeadlineSkippedResult(currentURL, redirectCount) {
  return {
    statusCode: 0,
    finalUrl: currentURL,
    redirected: redirectCount > 0,
    redirectCount,
    ok: false,
    timeoutSkipped: true,
    error: 'link probe stopped because the rendered audit deadline was reached',
  }
}

function isAbortError(error) {
  return error?.name === 'AbortError'
}

async function discardResponseBody(response) {
  if (!response?.body || typeof response.body.cancel !== 'function') {
    return
  }
  try {
    await response.body.cancel()
  } catch {
    // Link probes only need status and headers, so body cleanup must not mask the probe result.
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
  if (runtimeAuditDeadlineReached(options)) {
    return timeoutSkippedInteractionResult({
      name,
      selector,
      action,
      thresholdMilliseconds,
      error: 'interaction probe skipped because the rendered audit deadline was reached',
    })
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
    await settleAfterInteraction(page, options)
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

async function settleAfterInteraction(page, options = {}) {
  if (runtimeAuditDeadlineReached(options)) {
    return
  }
  await page.evaluate(() => new Promise((resolve) => {
    requestAnimationFrame(() => requestAnimationFrame(() => window.setTimeout(resolve, 0)))
  })).catch(() => {})
  const remaining = runtimeAuditBudgetRemaining(options)
  if (remaining > 0) {
    await delay(Math.min(100, remaining))
  }
}

async function runSoftNavigationProbe(page, target, input, options = {}) {
  const thresholdMilliseconds = normalizeThreshold(input.softNavigationMaxDurationMilliseconds, undefined, 2_000)
  const heapThresholdMB = normalizeThreshold(input.softNavigationMaxHeapGrowthMB, undefined, 32)
  const startedURL = page.url()
  const marker = `site-quality-soft-nav-${Date.now()}-${Math.random()}`
  if (runtimeAuditDeadlineReached(options)) {
    return timeoutSkippedSoftNavigationResult({
      target,
      fromUrl: startedURL,
      toUrl: startedURL,
      thresholdMilliseconds,
      error: 'soft navigation probe skipped because the rendered audit deadline was reached',
    })
  }
  try {
    await page.evaluate((value) => {
      window.__siteQualitySoftNavigationMarker = value
    }, marker)
    const beforeHeap = await page.evaluate(() => performance.memory?.usedJSHeapSize || 0).catch(() => 0)
    const startedAt = Date.now()
    await page.click(target.selector, { delay: 20 })
    const waitBudgetMilliseconds = runtimeAuditBudgetRemaining(options)
    if (waitBudgetMilliseconds <= 0) {
      return timeoutSkippedSoftNavigationResult({
        target,
        fromUrl: startedURL,
        toUrl: page.url(),
        thresholdMilliseconds,
        error: 'soft navigation wait skipped because the rendered audit deadline was reached',
      })
    }
    await Promise.race([
      page.waitForFunction(
        (previousURL) => location.href !== previousURL,
        { timeout: Math.max(500, Math.min(thresholdMilliseconds + 2_000, waitBudgetMilliseconds)) },
        startedURL,
      )
        .catch(() => null),
      delay(Math.min(thresholdMilliseconds + 2_000, waitBudgetMilliseconds)),
    ])
    const settleBudget = runtimeAuditBudgetRemaining(options)
    if (settleBudget > 0) {
      await settleRenderedDocument(page, Math.min(1_000, settleBudget), {
        timeoutMilliseconds: settleBudget,
      }).catch(() => {})
    }
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

export async function mapLimited(items, limit, mapper, options = {}) {
  const results = new Array(items.length)
  let next = 0
  let timedOut = false
  const workers = Array.from({ length: Math.max(1, Math.min(limit, items.length)) }, async () => {
    for (;;) {
      if (runtimeAuditDeadlineReached(options)) {
        timedOut = true
        return
      }
      const index = next++
      if (index >= items.length) {
        return
      }
      if (runtimeAuditDeadlineReached(options)) {
        timedOut = true
        return
      }
      results[index] = await mapper(items[index], index)
    }
  })
  await Promise.all(workers)
  const completed = results.filter((item) => item !== undefined)
  return {
    results: completed,
    skippedCount: Math.max(0, items.length - completed.length),
    timedOut,
  }
}

export function runtimeAuditBudgetRemaining(options = {}, reserveMilliseconds = runtimeAuditDeadlineReserveMilliseconds) {
  const deadlineAt = Number(options?.deadlineAt || 0)
  if (!Number.isFinite(deadlineAt) || deadlineAt <= 0) {
    return Number.POSITIVE_INFINITY
  }
  const reserve = Math.max(0, Number(reserveMilliseconds || 0))
  return Math.max(0, Math.floor(deadlineAt - Date.now() - reserve))
}

export function runtimeAuditDeadlineReached(options = {}, reserveMilliseconds = runtimeAuditDeadlineReserveMilliseconds) {
  return runtimeAuditBudgetRemaining(options, reserveMilliseconds) <= 0
}

function timeoutSkippedAudit(page, options) {
  const collectionKey = options.collectionKey
  return {
    status: 'timeout_skipped',
    source: 'chrome-rendered-dom',
    configured: Boolean(options.configured),
    finalUrl: page.url(),
    error: options.error || 'runtime audit skipped because the rendered audit deadline was reached',
    ...(collectionKey ? { [collectionKey]: [] } : {}),
    ...Object.fromEntries(
      Object.entries(options).filter(([key]) => !['collectionKey', 'configured', 'error'].includes(key)),
    ),
  }
}

function timeoutSkippedLinkResult(link, error) {
  return {
    ...link,
    statusCode: 0,
    finalUrl: String(link?.href || ''),
    redirected: false,
    redirectCount: 0,
    ok: false,
    timeoutSkipped: true,
    error,
  }
}

function timeoutSkippedInteractionResult({ name, selector, action, thresholdMilliseconds, error }) {
  return {
    name,
    selector,
    action,
    status: 'timeout_skipped',
    thresholdMilliseconds,
    error,
  }
}

function timeoutSkippedSoftNavigationResult({ target, fromUrl, toUrl, thresholdMilliseconds, error }) {
  return {
    fromUrl,
    toUrl,
    expectedUrl: target.href,
    selector: target.selector,
    text: target.text,
    status: 'timeout_skipped',
    thresholdMilliseconds,
    error,
  }
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

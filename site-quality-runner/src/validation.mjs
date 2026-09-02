import dns from 'node:dns/promises'
import net from 'node:net'

const maxURLLength = 2048
const defaultWorkerHeapLimitMB = 1024
const validThrottlingMethods = new Set(['simulate', 'devtools'])
const defaultJSBudgetBytes = 250 * 1024
const defaultImageBudgetBytes = 300 * 1024

export class RunnerInputError extends Error {
  constructor(message) {
    super(message)
    this.name = 'RunnerInputError'
  }
}

export function loadRunnerConfig(environment = process.env) {
  const allowedOrigin = parseAllowedOrigin(environment.SITE_QUALITY_ALLOWED_ORIGIN)
  const token = String(environment.SITE_QUALITY_RUNNER_TOKEN || '').trim()
  const timeoutSeconds = readBoundedInt(environment.SITE_QUALITY_RUNNER_TIMEOUT_SECONDS, 90, 15, 300)
  const maxConcurrency = readBoundedInt(environment.SITE_QUALITY_RUNNER_MAX_CONCURRENCY, 1, 1, 4)
  const workerHeapLimitMB = readBoundedInt(
    environment.SITE_QUALITY_RUNNER_WORKER_HEAP_LIMIT_MB,
    defaultWorkerHeapLimitMB,
    384,
    1536,
  )
  const headingSettleMilliseconds = readBoundedInt(environment.SITE_QUALITY_HEADING_SETTLE_MS, 1_500, 250, 10_000)
  const structuredDataSettleMilliseconds = readBoundedInt(
    environment.SITE_QUALITY_STRUCTURED_DATA_SETTLE_MS,
    1_500,
    250,
    10_000,
  )
  const renderWaitSelector = normalizeOptionalString(environment.SITE_QUALITY_RENDER_WAIT_SELECTOR, 256)
  const renderWaitTimeoutMilliseconds = readBoundedInt(
    environment.SITE_QUALITY_RENDER_WAIT_TIMEOUT_MS,
    3_000,
    500,
    25_000,
  )
  const throttlingMethod = parseThrottlingMethod(environment.SITE_QUALITY_THROTTLING_METHOD)
  const lighthouseRunCount = readBoundedInt(environment.SITE_QUALITY_LIGHTHOUSE_RUN_COUNT, 1, 1, 5)
  const interactionProbes = normalizeInteractionProbes(
    environment.SITE_QUALITY_INTERACTION_PROBES,
    [],
    (message) => new Error(`SITE_QUALITY_INTERACTION_PROBES ${message}`),
  )
  const interactionMaxResponseMilliseconds = readBoundedInt(
    environment.SITE_QUALITY_INTERACTION_MAX_RESPONSE_MS,
    200,
    50,
    2_000,
  )
  const softNavigationSelectors = normalizeOptionalStringList(
    environment.SITE_QUALITY_SOFT_NAVIGATION_SELECTORS,
    [],
    8,
    256,
    (message) => new Error(`SITE_QUALITY_SOFT_NAVIGATION_SELECTORS ${message}`),
  )
  const softNavigationMaxLinks = readBoundedInt(
    environment.SITE_QUALITY_SOFT_NAVIGATION_MAX_LINKS,
    softNavigationSelectors.length > 0 ? softNavigationSelectors.length : 0,
    0,
    8,
  )
  const softNavigationMaxDurationMilliseconds = readBoundedInt(
    environment.SITE_QUALITY_SOFT_NAVIGATION_MAX_DURATION_MS,
    2_000,
    250,
    10_000,
  )
  const softNavigationMaxHeapGrowthMB = readBoundedInt(
    environment.SITE_QUALITY_SOFT_NAVIGATION_MAX_HEAP_GROWTH_MB,
    32,
    1,
    256,
  )
  const jsBudgetBytes = readBoundedInt(environment.SITE_QUALITY_JS_BUDGET_BYTES, defaultJSBudgetBytes, 0, 5 * 1024 * 1024)
  const imageBudgetBytes = readBoundedInt(
    environment.SITE_QUALITY_IMAGE_BUDGET_BYTES,
    defaultImageBudgetBytes,
    0,
    10 * 1024 * 1024,
  )
  const linkCheckEnabled = normalizeBooleanDefault(environment.SITE_QUALITY_LINK_CHECK_ENABLED, true)
  const linkCheckMaxLinks = readBoundedInt(environment.SITE_QUALITY_LINK_CHECK_MAX_LINKS, 80, 0, 250)
  const linkCheckTimeoutMilliseconds = readBoundedInt(
    environment.SITE_QUALITY_LINK_CHECK_TIMEOUT_MS,
    2_500,
    500,
    15_000,
  )
  const linkCheckExternal = normalizeBoolean(environment.SITE_QUALITY_LINK_CHECK_EXTERNAL)
  const linkCheckMaxRedirects = readBoundedInt(environment.SITE_QUALITY_LINK_CHECK_MAX_REDIRECTS, 5, 0, 10)
  const allowPrivateTarget = normalizeBoolean(environment.SITE_QUALITY_RUNNER_ALLOW_PRIVATE_TARGETS)

  if (token.length < 32) {
    throw new Error('SITE_QUALITY_RUNNER_TOKEN must be at least 32 characters')
  }

  return {
    allowedOrigin,
    token,
    timeoutMilliseconds: timeoutSeconds * 1000,
    maxConcurrency,
    workerHeapLimitMB,
    headingSettleMilliseconds,
    structuredDataSettleMilliseconds,
    renderWaitSelector,
    renderWaitTimeoutMilliseconds,
    throttlingMethod,
    lighthouseRunCount,
    interactionProbes,
    interactionMaxResponseMilliseconds,
    softNavigationSelectors,
    softNavigationMaxLinks,
    softNavigationMaxDurationMilliseconds,
    softNavigationMaxHeapGrowthMB,
    jsBudgetBytes,
    imageBudgetBytes,
    linkCheckEnabled,
    linkCheckMaxLinks,
    linkCheckTimeoutMilliseconds,
    linkCheckExternal,
    linkCheckMaxRedirects,
    allowPrivateTarget,
  }
}

export async function normalizeRunInput(input, config, lookup = dns.lookup) {
  if (!input || typeof input !== 'object') {
    throw new RunnerInputError('request body must be an object')
  }

  const rawURL = String(input.url || '').trim()
  if (!rawURL) {
    throw new RunnerInputError('url is required')
  }
  if (rawURL.length > maxURLLength) {
    throw new RunnerInputError('url is too long')
  }

  let target
  try {
    target = new URL(rawURL)
  } catch {
    throw new RunnerInputError('url must be absolute')
  }

  if (!['http:', 'https:'].includes(target.protocol)) {
    throw new RunnerInputError('url must use HTTP or HTTPS')
  }
  if (target.username || target.password) {
    throw new RunnerInputError('url must not include credentials')
  }
  if (target.origin !== config.allowedOrigin.origin) {
    throw new RunnerInputError('url must belong to the configured storefront origin')
  }

  target.hash = ''
  const strategy = String(input.strategy || 'mobile').trim().toLowerCase()
  if (strategy !== 'mobile' && strategy !== 'desktop') {
    throw new RunnerInputError('strategy must be mobile or desktop')
  }
  const throttlingMethod = normalizeOptionalThrottlingMethod(input.throttling_method, config?.throttlingMethod)

  const records = await lookup(target.hostname, { all: true, verbatim: true })
  if (!Array.isArray(records) || records.length === 0) {
    throw new RunnerInputError('storefront host did not resolve')
  }
  if (!config.allowPrivateTarget && records.some((record) => isPrivateAddress(record.address))) {
    throw new RunnerInputError('storefront host resolved to a private address')
  }

  return {
    url: target.toString(),
    strategy,
    releaseID: typeof input.release_id === 'string' ? input.release_id.trim().slice(0, 128) : '',
    renderWaitSelector: normalizeOptionalString(input.render_wait_selector, 256) || config?.renderWaitSelector || '',
    renderWaitTimeoutMilliseconds: normalizeOptionalInt(
      input.render_wait_timeout_ms,
      config?.renderWaitTimeoutMilliseconds,
      500,
      25_000,
    ),
    headingSettleMilliseconds: normalizeOptionalInt(
      input.heading_settle_ms,
      config?.headingSettleMilliseconds,
      250,
      10_000,
    ),
    structuredDataSettleMilliseconds: normalizeOptionalInt(
      input.structured_data_settle_ms,
      config?.structuredDataSettleMilliseconds,
      250,
      10_000,
    ),
    throttlingMethod,
    lighthouseRunCount: normalizeOptionalInt(
      input.lighthouse_run_count,
      config?.lighthouseRunCount,
      1,
      5,
    ),
    interactionProbes: normalizeInteractionProbes(
      input.interaction_probes,
      config?.interactionProbes || [],
      (message) => new RunnerInputError(`interaction_probes ${message}`),
    ),
    interactionMaxResponseMilliseconds: normalizeOptionalInt(
      input.interaction_max_response_ms,
      config?.interactionMaxResponseMilliseconds,
      50,
      2_000,
    ),
    softNavigationSelectors: normalizeOptionalStringList(
      input.soft_navigation_selectors,
      config?.softNavigationSelectors || [],
      8,
      256,
      (message) => new RunnerInputError(`soft_navigation_selectors ${message}`),
    ),
    softNavigationMaxLinks: normalizeOptionalInt(
      input.soft_navigation_max_links,
      config?.softNavigationMaxLinks,
      0,
      8,
    ),
    softNavigationMaxDurationMilliseconds: normalizeOptionalInt(
      input.soft_navigation_max_duration_ms,
      config?.softNavigationMaxDurationMilliseconds,
      250,
      10_000,
    ),
    softNavigationMaxHeapGrowthMB: normalizeOptionalInt(
      input.soft_navigation_max_heap_growth_mb,
      config?.softNavigationMaxHeapGrowthMB,
      1,
      256,
    ),
    jsBudgetBytes: normalizeOptionalInt(
      input.js_budget_bytes,
      config?.jsBudgetBytes,
      0,
      5 * 1024 * 1024,
    ),
    imageBudgetBytes: normalizeOptionalInt(
      input.image_budget_bytes,
      config?.imageBudgetBytes,
      0,
      10 * 1024 * 1024,
    ),
    linkCheckEnabled: normalizeOptionalBoolean(input.link_check_enabled, config?.linkCheckEnabled ?? true),
    linkCheckMaxLinks: normalizeOptionalInt(
      input.link_check_max_links,
      config?.linkCheckMaxLinks,
      0,
      250,
    ),
    linkCheckTimeoutMilliseconds: normalizeOptionalInt(
      input.link_check_timeout_ms,
      config?.linkCheckTimeoutMilliseconds,
      500,
      15_000,
    ),
    linkCheckExternal: normalizeOptionalBoolean(input.link_check_external, config?.linkCheckExternal || false),
    linkCheckMaxRedirects: normalizeOptionalInt(
      input.link_check_max_redirects,
      config?.linkCheckMaxRedirects,
      0,
      10,
    ),
  }
}

export function isPrivateAddress(address) {
  const family = net.isIP(address)
  if (family === 4) {
    const octets = address.split('.').map(Number)
    const [a, b] = octets
    return a === 0 ||
      a === 10 ||
      a === 127 ||
      a >= 224 ||
      (a === 100 && b >= 64 && b <= 127) ||
      (a === 169 && b === 254) ||
      (a === 172 && b >= 16 && b <= 31) ||
      (a === 192 && b === 0) ||
      (a === 192 && b === 168) ||
      (a === 198 && (b === 18 || b === 19))
  }
  if (family === 6) {
    const normalized = address.toLowerCase()
    return normalized === '::' ||
      normalized === '::1' ||
      normalized.startsWith('fe80:') ||
      normalized.startsWith('fc') ||
      normalized.startsWith('fd') ||
      normalized.startsWith('::ffff:127.') ||
      normalized.startsWith('::ffff:10.') ||
      normalized.startsWith('::ffff:192.168.') ||
      normalized.startsWith('::ffff:169.254.')
  }
  return true
}

function parseAllowedOrigin(value) {
  const raw = String(value || '').trim()
  if (!raw) {
    throw new Error('SITE_QUALITY_ALLOWED_ORIGIN is required')
  }
  let parsed
  try {
    parsed = new URL(raw)
  } catch {
    throw new Error('SITE_QUALITY_ALLOWED_ORIGIN must be an absolute URL')
  }
  if (!['http:', 'https:'].includes(parsed.protocol) || parsed.username || parsed.password || parsed.pathname !== '/' || parsed.search || parsed.hash) {
    throw new Error('SITE_QUALITY_ALLOWED_ORIGIN must contain only an HTTP(S) origin')
  }
  return parsed
}

function normalizeBoolean(value) {
  return ['1', 'true', 'yes', 'on'].includes(String(value || '').trim().toLowerCase())
}

function normalizeBooleanDefault(value, fallback) {
  const raw = String(value ?? '').trim().toLowerCase()
  if (!raw) {
    return fallback
  }
  return ['1', 'true', 'yes', 'on'].includes(raw)
}

function readBoundedInt(value, fallback, min, max) {
  const parsed = Number.parseInt(String(value || ''), 10)
  if (!Number.isInteger(parsed)) {
    return fallback
  }
  return Math.min(Math.max(parsed, min), max)
}

function normalizeOptionalString(value, maxLength) {
  const raw = String(value || '').trim()
  if (!raw) {
    return ''
  }
  return raw.slice(0, maxLength)
}

function normalizeOptionalStringList(value, fallback, maxItems, maxLength, createError) {
  if (value === undefined || value === null || value === '') {
    return Array.isArray(fallback) ? fallback : []
  }
  let rawItems
  if (Array.isArray(value)) {
    rawItems = value
  } else {
    const raw = String(value || '').trim()
    if (!raw) {
      return Array.isArray(fallback) ? fallback : []
    }
    if (raw.startsWith('[')) {
      try {
        rawItems = JSON.parse(raw)
      } catch {
        throw createError('must be a JSON array or semicolon-delimited selector list')
      }
      if (!Array.isArray(rawItems)) {
        throw createError('must be a JSON array')
      }
    } else {
      rawItems = raw.split(';')
    }
  }
  const items = []
  const seen = new Set()
  for (const item of rawItems) {
    const normalized = normalizeOptionalString(item, maxLength)
    if (!normalized || seen.has(normalized)) {
      continue
    }
    seen.add(normalized)
    items.push(normalized)
    if (items.length >= maxItems) {
      break
    }
  }
  return items
}

function normalizeOptionalInt(value, fallback, min, max) {
  if (value === undefined || value === null || value === '') {
    return fallback
  }
  const parsed = Number.parseInt(String(value), 10)
  if (!Number.isInteger(parsed)) {
    return fallback
  }
  return Math.min(Math.max(parsed, min), max)
}

function normalizeOptionalBoolean(value, fallback) {
  if (value === undefined || value === null || value === '') {
    return Boolean(fallback)
  }
  if (typeof value === 'boolean') {
    return value
  }
  const raw = String(value).trim().toLowerCase()
  if (['1', 'true', 'yes', 'on'].includes(raw)) {
    return true
  }
  if (['0', 'false', 'no', 'off'].includes(raw)) {
    return false
  }
  return Boolean(fallback)
}

function normalizeOptionalThrottlingMethod(value, fallback) {
  const raw = String(value || '').trim().toLowerCase()
  if (!raw) {
    return fallback
  }
  if (!validThrottlingMethods.has(raw)) {
    throw new RunnerInputError('throttling_method must be simulate or devtools')
  }
  return raw
}

function normalizeInteractionProbes(value, fallback, createError) {
  if (value === undefined || value === null || value === '') {
    return Array.isArray(fallback) ? fallback : []
  }
  let rawItems
  if (Array.isArray(value)) {
    rawItems = value
  } else {
    try {
      rawItems = JSON.parse(String(value || '').trim())
    } catch {
      throw createError('must be a JSON array')
    }
  }
  if (!Array.isArray(rawItems)) {
    throw createError('must be a JSON array')
  }
  const probes = []
  for (const item of rawItems) {
    if (!item || typeof item !== 'object') {
      continue
    }
    const selector = normalizeOptionalString(item.selector, 256)
    if (!selector) {
      continue
    }
    const action = normalizeInteractionAction(item.action)
    const probe = {
      name: normalizeOptionalString(item.name || selector, 80),
      selector,
      action,
      maxResponseMilliseconds: normalizeOptionalInt(item.max_response_ms ?? item.maxResponseMilliseconds, 0, 0, 2_000),
    }
    if (action === 'type') {
      probe.text = normalizeOptionalString(item.text, 120)
    }
    if (action === 'press') {
      probe.key = normalizeOptionalString(item.key || 'Enter', 32)
    }
    if (action === 'drag') {
      probe.deltaX = normalizeOptionalNumber(item.delta_x ?? item.deltaX, 80, -1_000, 1_000)
      probe.deltaY = normalizeOptionalNumber(item.delta_y ?? item.deltaY, 0, -1_000, 1_000)
    }
    probes.push(probe)
    if (probes.length >= 8) {
      break
    }
  }
  return probes
}

function normalizeOptionalNumber(value, fallback, min, max) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) {
    return fallback
  }
  return Math.min(Math.max(parsed, min), max)
}

function normalizeInteractionAction(value) {
  const action = String(value || '').trim().toLowerCase()
  return ['click', 'hover', 'type', 'press', 'drag'].includes(action) ? action : 'click'
}

function parseThrottlingMethod(value) {
  const raw = String(value || '').trim().toLowerCase()
  if (!raw) {
    return 'simulate'
  }
  if (!validThrottlingMethods.has(raw)) {
    throw new Error('SITE_QUALITY_THROTTLING_METHOD must be simulate or devtools')
  }
  return raw
}

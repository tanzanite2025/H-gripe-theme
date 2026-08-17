import dns from 'node:dns/promises'
import net from 'node:net'

const maxURLLength = 2048

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
  const headingSettleMilliseconds = readBoundedInt(environment.SITE_QUALITY_HEADING_SETTLE_MS, 1_500, 250, 10_000)
  const allowPrivateTarget = normalizeBoolean(environment.SITE_QUALITY_RUNNER_ALLOW_PRIVATE_TARGETS)

  if (token.length < 32) {
    throw new Error('SITE_QUALITY_RUNNER_TOKEN must be at least 32 characters')
  }

  return {
    allowedOrigin,
    token,
    timeoutMilliseconds: timeoutSeconds * 1000,
    maxConcurrency,
    headingSettleMilliseconds,
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

function readBoundedInt(value, fallback, min, max) {
  const parsed = Number.parseInt(String(value || ''), 10)
  if (!Number.isInteger(parsed)) {
    return fallback
  }
  return Math.min(Math.max(parsed, min), max)
}

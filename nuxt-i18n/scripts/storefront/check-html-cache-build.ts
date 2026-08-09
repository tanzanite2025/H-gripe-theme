import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import locales from '../../app/i18n/locales.manifest.ts'

interface RouteCacheRule {
  base?: string
  maxAge?: number
  staleMaxAge?: number
  swr?: boolean
}

interface RouteRule {
  headers?: Record<string, string | number | boolean | undefined>
  cache?: RouteCacheRule
}

type RouteRules = Record<string, RouteRule | undefined>

interface RuntimeConfig {
  nitro?: {
    routeRules?: RouteRules
  }
}

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const nitroChunkPath = resolve(projectRoot, '.output/server/chunks/_/nitro.mjs')
const purgeRoutePath = resolve(projectRoot, '.output/server/chunks/routes/_internal/html-cache/purge.post.mjs')

const failures: string[] = []

const fail = (message: string): void => {
  failures.push(message)
}

const assert = (condition: unknown, message: string): void => {
  if (!condition) fail(message)
}

const readRequiredFile = (filePath: string): string => {
  if (!existsSync(filePath)) {
    fail(`Missing build artifact: ${filePath}`)
    return ''
  }
  return readFileSync(filePath, 'utf8')
}

const extractInlineRuntimeConfig = (source: string): RuntimeConfig | null => {
  const marker = 'const _inlineRuntimeConfig = '
  const markerIndex = source.indexOf(marker)
  if (markerIndex === -1) {
    fail('Cannot find _inlineRuntimeConfig in Nitro build chunk')
    return null
  }

  const objectStart = source.indexOf('{', markerIndex + marker.length)
  if (objectStart === -1) {
    fail('Cannot find _inlineRuntimeConfig object start')
    return null
  }

  let depth = 0
  let inString = false
  let escaped = false

  for (let index = objectStart; index < source.length; index += 1) {
    const char = source[index]

    if (inString) {
      if (escaped) {
        escaped = false
      } else if (char === '\\') {
        escaped = true
      } else if (char === '"') {
        inString = false
      }
      continue
    }

    if (char === '"') {
      inString = true
      continue
    }

    if (char === '{') {
      depth += 1
    } else if (char === '}') {
      depth -= 1
      if (depth === 0) {
        const rawObject = source.slice(objectStart, index + 1)
        try {
          return JSON.parse(rawObject) as RuntimeConfig
        } catch (error: unknown) {
          fail(`Cannot parse _inlineRuntimeConfig JSON: ${error instanceof Error ? error.message : String(error)}`)
          return null
        }
      }
    }
  }

  fail('Cannot find _inlineRuntimeConfig object end')
  return null
}

const getCacheControl = (routeRules: RouteRules, routePath: string): string => {
  return String(routeRules[routePath]?.headers?.['cache-control'] || '')
}

const assertNoStore = (routeRules: RouteRules, routePath: string): void => {
  assert(
    getCacheControl(routeRules, routePath).includes('no-store'),
    `${routePath} must be no-store`,
  )
}

const assertHtmlCache = (
  routeRules: RouteRules,
  routePath: string,
  expectedMaxAge: number,
  expectedStaleMaxAge: number,
): void => {
  const cache = routeRules[routePath]?.cache
  assert(cache, `${routePath} must have Nitro route cache`)
  assert(cache?.base === '/cache/html', `${routePath} cache base must be /cache/html`)
  assert(cache?.maxAge === expectedMaxAge, `${routePath} maxAge must be ${expectedMaxAge}`)
  assert(cache?.staleMaxAge === expectedStaleMaxAge, `${routePath} staleMaxAge must be ${expectedStaleMaxAge}`)
  assert(cache?.swr === true, `${routePath} must enable swr`)
}

const nitroSource = readRequiredFile(nitroChunkPath)
const purgeRouteSource = readRequiredFile(purgeRoutePath)
const runtimeConfig = nitroSource ? extractInlineRuntimeConfig(nitroSource) : null
const routeRules: RouteRules = runtimeConfig?.nitro?.routeRules || {}

if (runtimeConfig) {
  assertNoStore(routeRules, '/')
  assertNoStore(routeRules, '/api/**')
  assertNoStore(routeRules, '/_internal/**')
  assertNoStore(routeRules, '/shop')

  assertHtmlCache(routeRules, '/shop/**', 300, 3600)

  for (const locale of locales) {
    const code = String(locale.code || '')
    if (!code || code === 'en') continue
    assertHtmlCache(routeRules, `/${code}/shop/**`, 300, 3600)
  }

  assertHtmlCache(routeRules, '/support/shipping', 86400, 604800)
  assertHtmlCache(routeRules, '/blog/**', 3600, 86400)
}

if (purgeRouteSource) {
  assert(purgeRouteSource.includes('timingSafeEqual'), 'purge route must use timingSafeEqual token comparison')
  assert(purgeRouteSource.includes('getKeys(htmlRouteCacheStorageBase)'), 'purge route must list keys by /cache/html base')
  assert(purgeRouteSource.includes('removeItem'), 'purge route must remove cache keys explicitly')
  assert(purgeRouteSource.includes('purgedKeys'), 'purge route must return purgedKeys')
  assert(purgeRouteSource.includes('cache-control'), 'purge route must set cache-control no-store')
}

if (nitroSource) {
  assert(nitroSource.includes('lazyConnect: true'), 'Redis cache driver must lazy-connect before the readiness probe')
  assert(nitroSource.includes('await redisClient.connect()'), 'Redis cache driver client must connect before mounting cache storage')
  assert(nitroSource.includes('await redisClient.ping()'), 'Redis cache driver client must be pinged before mounting cache storage')
  assert(nitroSource.includes('storage.mount("cache", cacheDriver)'), 'Nitro cache storage must mount the pre-warmed Redis driver')
}

if (failures.length > 0) {
  console.error('[html-cache-check] FAILED')
  for (const message of failures) {
    console.error(`- ${message}`)
  }
  process.exit(1)
}

console.log('[html-cache-check] OK: Nitro HTML cache routeRules and purge artifact are aligned.')

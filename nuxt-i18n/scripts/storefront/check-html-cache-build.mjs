import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import locales from '../../app/i18n/locales.manifest.js'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const nitroChunkPath = resolve(projectRoot, '.output/server/chunks/_/nitro.mjs')
const purgeRoutePath = resolve(projectRoot, '.output/server/chunks/routes/_internal/html-cache/purge.post.mjs')

const failures = []

const fail = (message) => {
  failures.push(message)
}

const assert = (condition, message) => {
  if (!condition) fail(message)
}

const readRequiredFile = (filePath) => {
  if (!existsSync(filePath)) {
    fail(`Missing build artifact: ${filePath}`)
    return ''
  }
  return readFileSync(filePath, 'utf8')
}

const extractInlineRuntimeConfig = (source) => {
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
          return JSON.parse(rawObject)
        } catch (error) {
          fail(`Cannot parse _inlineRuntimeConfig JSON: ${error.message}`)
          return null
        }
      }
    }
  }

  fail('Cannot find _inlineRuntimeConfig object end')
  return null
}

const getCacheControl = (routeRules, path) => {
  return String(routeRules[path]?.headers?.['cache-control'] || '')
}

const assertNoStore = (routeRules, path) => {
  assert(
    getCacheControl(routeRules, path).includes('no-store'),
    `${path} must be no-store`,
  )
}

const assertHtmlCache = (routeRules, path, expectedMaxAge, expectedStaleMaxAge) => {
  const cache = routeRules[path]?.cache
  assert(cache, `${path} must have Nitro route cache`)
  assert(cache?.base === '/cache/html', `${path} cache base must be /cache/html`)
  assert(cache?.maxAge === expectedMaxAge, `${path} maxAge must be ${expectedMaxAge}`)
  assert(cache?.staleMaxAge === expectedStaleMaxAge, `${path} staleMaxAge must be ${expectedStaleMaxAge}`)
  assert(cache?.swr === true, `${path} must enable swr`)
}

const nitroSource = readRequiredFile(nitroChunkPath)
const purgeRouteSource = readRequiredFile(purgeRoutePath)
const runtimeConfig = nitroSource ? extractInlineRuntimeConfig(nitroSource) : null
const routeRules = runtimeConfig?.nitro?.routeRules || {}

if (runtimeConfig) {
  assertNoStore(routeRules, '/')
  assertNoStore(routeRules, '/api/**')
  assertNoStore(routeRules, '/_internal/**')
  assertNoStore(routeRules, '/shop')

  assertHtmlCache(routeRules, '/shop/**', 300, 3600)

  for (const locale of locales) {
    const code = String(locale?.code || '')
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

if (failures.length > 0) {
  console.error('[html-cache-check] FAILED')
  for (const message of failures) {
    console.error(`- ${message}`)
  }
  process.exit(1)
}

console.log('[html-cache-check] OK: Nitro HTML cache routeRules and purge artifact are aligned.')

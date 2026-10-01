import { existsSync } from 'node:fs'
import { readFile } from 'node:fs/promises'
import { join, relative, sep } from 'node:path'
import { fileURLToPath } from 'node:url'
import locales from '../../app/i18n/locales.manifest'

type ManifestRoute = {
  key?: string
  path?: string
  is_alias?: boolean
  sitemap_locales?: string[]
}

type RouteManifest = {
  version?: string
  routes?: ManifestRoute[]
}

const projectRoot = fileURLToPath(new URL('../../', import.meta.url))
const pagesRoot = join(projectRoot, 'app', 'pages')
const manifestPath = join(projectRoot, 'public', 'storefront-route-manifest.json')

// Some pages deliberately live outside their public URL directory so they do
// not become children of a parent page that has no <NuxtPage /> outlet. Keep
// those explicit Nuxt route overrides in the manifest checker as well.
const customPagePathOverrides = new Map<string, string>([
  [
    '/guides/tireguides/schwalbe-tire-selector',
    join(pagesRoot, 'guides', 'schwalbe-tire-selector.vue'),
  ],
  [
    '/guides/tireguides/tire-frame-clearance',
    join(pagesRoot, 'guides', 'tireguides.vue'),
  ],
  [
    '/guides/tireguides/schwalbe-tire-circumference',
    join(pagesRoot, 'guides', 'tireguides.vue'),
  ],
])

const normalizePath = (value: string) => {
  if (value === '/') return '/'
  return `/${value.replace(/^\/+|\/+$/g, '')}`
}

const isDynamicSegment = (segment: string) => segment.startsWith('[') && segment.endsWith(']')

const pageFileForPath = (path: string) => {
  const normalized = normalizePath(path)
  if (normalized.includes(':') || normalized.split('/').some(isDynamicSegment)) return null

  const customPageFile = customPagePathOverrides.get(normalized)
  if (customPageFile && existsSync(customPageFile)) return customPageFile

  const segments = normalized === '/' ? [] : normalized.slice(1).split('/')
  const directory = join(pagesRoot, ...segments)
  const indexFile = join(directory, 'index.vue')
  if (existsSync(indexFile)) return indexFile

  const pageFile = join(pagesRoot, ...segments) + '.vue'
  return existsSync(pageFile) ? pageFile : null
}

const manifest = JSON.parse(await readFile(manifestPath, 'utf8')) as RouteManifest
const routes = Array.isArray(manifest.routes) ? manifest.routes : []
const enabledLocaleCodes = new Set(locales.map(locale => locale.code.toLowerCase()))
const errors: string[] = []
const seenKeys = new Set<string>()
const seenPaths = new Set<string>()

if (!manifest.version?.trim()) errors.push('manifest version is missing')
if (routes.length === 0) errors.push('manifest has no routes')

for (const route of routes) {
  const key = String(route.key || '').trim()
  const path = normalizePath(String(route.path || '').trim())

  if (!key) errors.push('route key is missing')
  if (!route.path?.trim()) errors.push(`route ${key || '<unknown>'} path is missing`)
  if (key && seenKeys.has(key)) errors.push(`duplicate route key: ${key}`)
  if (key) seenKeys.add(key)

  if (route.sitemap_locales !== undefined) {
    if (!Array.isArray(route.sitemap_locales) || route.sitemap_locales.length === 0) {
      errors.push(`route ${key || path} sitemap_locales must be a non-empty array`)
    } else if (route.sitemap_locales.some(locale => typeof locale !== 'string' || !locale.trim())) {
      errors.push(`route ${key || path} sitemap_locales contains an empty locale`)
    } else {
      const normalizedLocales = route.sitemap_locales.map(locale => locale.trim().toLowerCase().replace(/-/g, '_'))
      const duplicateLocales = normalizedLocales.filter((locale, index) => normalizedLocales.indexOf(locale) !== index)
      if (duplicateLocales.length > 0) {
        errors.push(`route ${key || path} sitemap_locales contains duplicates: ${[...new Set(duplicateLocales)].join(', ')}`)
      }
      for (const locale of normalizedLocales) {
        if (!enabledLocaleCodes.has(locale)) {
          errors.push(`route ${key || path} sitemap_locales contains unsupported locale: ${locale}`)
        }
      }
    }
  }

  if (seenPaths.has(path) && !route.is_alias) errors.push(`duplicate canonical route path: ${path}`)
  if (!route.is_alias) seenPaths.add(path)

  if (route.is_alias) continue

  const pageFile = pageFileForPath(path)
  if (!pageFile) {
    errors.push(`manifest route ${key || path} has no matching static Nuxt page: ${path}`)
    continue
  }

  const relativePage = relative(projectRoot, pageFile).split(sep).join('/')
  if (!relativePage.endsWith('.vue')) {
    errors.push(`manifest route ${key || path} resolved to a non-Vue page: ${relativePage}`)
  }
}

if (errors.length > 0) {
  console.error(`Storefront route manifest check failed (${errors.length} issue${errors.length === 1 ? '' : 's'}):`)
  for (const error of errors) console.error(`- ${error}`)
  process.exitCode = 1
} else {
  console.log(`Storefront route manifest check passed: ${routes.length} declarations match Nuxt pages.`)
}

import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'
import manifestLocales from '../app/i18n/locales.manifest.ts'
import type { LocaleManifestEntry } from '../app/i18n/locales.manifest.ts'
import { STOREFRONT_SUPPORTED_LANGUAGES as adminFallbackLocales } from '../../go-backend/web/admin/src/lib/languages.ts'

interface StorefrontLocaleRegistryEntry {
  code: string
  iso: string
  name: string
  native_name: string
  file: string
  dir?: 'ltr' | 'rtl'
  enabled: boolean
}

interface BackendLocale {
  code: string
  name?: string
  native_name?: string
  enabled?: boolean
}

interface BackendLocaleResponse {
  languages: BackendLocale[]
}

interface LocaleCheckFailure {
  source: string
  message: string
}

const apiBase =
  process.env.NUXT_PUBLIC_API_BASE ||
  process.env.GO_API_BASE ||
  process.env.API_BASE ||
  ''
const localesUrl = process.env.GO_LOCALES_URL || process.env.LOCALES_URL
const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(scriptDir, '..', '..')
const registryPath = process.env.STOREFRONT_LOCALES_FILE
  ? path.resolve(process.env.STOREFRONT_LOCALES_FILE)
  : path.resolve(repoRoot, 'shared/storefront-locales.json')
const defaultGoLanguagesPath = path.resolve(repoRoot, 'go-backend/internal/pkg/locales/locales.go')
const goLanguagesPath = process.env.GO_I18N_LANGUAGES_FILE
  ? path.resolve(process.env.GO_I18N_LANGUAGES_FILE)
  : defaultGoLanguagesPath
const goConfigLocalePaths = [
  path.resolve(repoRoot, 'go-backend/config/config.example.yaml'),
  path.resolve(repoRoot, 'go-backend/config/config.production.yaml'),
]
const nuxtActiveSourceRoot = path.resolve(repoRoot, 'nuxt-i18n/app')
const nuxtLocaleDefinitionFiles = new Set([
  path.resolve(repoRoot, 'nuxt-i18n/app/i18n/locales.manifest.ts'),
  path.resolve(repoRoot, 'nuxt-i18n/app/utils/storefrontLocales.ts'),
])
const staleLocaleDocumentationTargets = [
  path.resolve(repoRoot, 'go-backend/API.md'),
  path.resolve(repoRoot, 'go-backend/frontend-examples/nuxt3'),
  path.resolve(repoRoot, 'nuxt-i18n/docs/notes/I18N-CURRENT-STATUS.md'),
]
const legacyUnsupportedStorefrontLocaleLiterals = new Set([
  'zh',
  'zh-cn',
  'zh-hans',
  'zh-sg',
  'zh-tw',
  'pl',
  'vi',
  'bn',
  'ta',
  'te',
  'mr',
  'ur',
  'fa',
  'he',
  'cs',
  'hu',
  'ro',
])
const localeLiteralPattern = /(['"])([a-z]{2}(?:[_-][A-Za-z]{2,4})?)\1/g
const hardcodedLocaleListThreshold = 6
const staleLocaleDocumentationPatterns: Array<{ label: string; pattern: RegExp }> = [
  { label: '34-language wording', pattern: /Supported Languages \(34|34\s+(?:configured\s+)?locales|34\s+total|"total"\s*:\s*34/i },
  { label: 'legacy Chinese route example', pattern: /\/zh\/|sitemap-zh\.xml/ },
  { label: 'legacy zh locale code example', pattern: /(?:code|locale)\s*:\s*['"]zh['"]|"code"\s*:\s*"zh"|"locale"\s*:\s*"zh"/ },
  { label: 'legacy translation locale map', pattern: /['"]zh-TW['"]\s*:|flagMap\s*:/ },
  { label: 'legacy grouped language list', pattern: /Asian:\s*zh\b/i },
]

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function isBackendLocale(value: unknown): value is BackendLocale {
  return isRecord(value) && typeof value.code === 'string'
}

function isBackendLocaleResponse(value: unknown): value is BackendLocaleResponse {
  return isRecord(value) && Array.isArray(value.languages) && value.languages.every(isBackendLocale)
}

function isRegistryLocale(value: unknown): value is StorefrontLocaleRegistryEntry {
  return isRecord(value) &&
    typeof value.code === 'string' &&
    typeof value.iso === 'string' &&
    typeof value.name === 'string' &&
    typeof value.native_name === 'string' &&
    typeof value.file === 'string' &&
    typeof value.enabled === 'boolean' &&
    (value.dir === undefined || value.dir === 'ltr' || value.dir === 'rtl')
}

function loadRegistryLocales(): StorefrontLocaleRegistryEntry[] {
  if (!fs.existsSync(registryPath)) {
    throw new Error(`Storefront locale registry not found: ${registryPath}`)
  }

  const parsed = JSON.parse(fs.readFileSync(registryPath, 'utf8')) as unknown
  if (!Array.isArray(parsed) || !parsed.every(isRegistryLocale)) {
    throw new Error('Storefront locale registry must be an array of valid locale entries')
  }
  return parsed
}

function loadManifestLocales(): LocaleManifestEntry[] {
  return manifestLocales
}

function resolveEndpoint(): string {
  if (localesUrl) return localesUrl
  const base = apiBase.replace(/\/$/, '')
  if (base.endsWith('/api/v1')) return `${base}/i18n/languages`
  return `${base}/api/v1/i18n/languages`
}

async function fetchBackendLocales(): Promise<BackendLocale[]> {
  const url = resolveEndpoint()
  const res = await fetch(url)
  if (!res.ok) {
    if (res.status === 404) {
      throw new Error(`Fetch backend locales failed: 404 Not Found. Check that the Go backend exposes ${url}. Override with GO_LOCALES_URL or LOCALES_URL if needed.`)
    }
    throw new Error(`Fetch backend locales failed: ${res.status} ${res.statusText} (url: ${url})`)
  }
  const data = await res.json() as unknown
  if (!isBackendLocaleResponse(data)) {
    throw new Error('Unexpected backend locales response shape')
  }
  return data.languages
}

function unquoteGoString(value: unknown): string {
  const trimmed = String(value || '').trim()
  if (trimmed.startsWith('`') && trimmed.endsWith('`')) {
    return trimmed.slice(1, -1)
  }

  try {
    return JSON.parse(trimmed) as string
  } catch {
    return trimmed.replace(/^"|"$/g, '')
  }
}

function readGoStringField(entry: string, fieldName: string): string {
  const quotedStringPattern = '"(?:\\\\.|[^"\\\\])*"|`[^`]*`'
  const fieldPattern = new RegExp(`${fieldName}\\s*:\\s*(${quotedStringPattern})`)
  const match = entry.match(fieldPattern)
  return match ? unquoteGoString(match[1]) : ''
}

function parseGoSupportedLanguages(source: string): BackendLocale[] {
  const blockMatch = source.match(/var\s+SupportedLanguages\s*=\s*\[\]Language\s*\{([\s\S]*?)\n\}/)
  if (!blockMatch) {
    throw new Error('SupportedLanguages declaration not found in Go locales source')
  }

  const languages: BackendLocale[] = []
  for (const match of blockMatch[1].matchAll(/\{([^{}]*)\}/g)) {
    const entry = match[1]
    const code = readGoStringField(entry, 'Code')
    if (!code) continue

    const enabledMatch = entry.match(/Enabled\s*:\s*(true|false)/)
    languages.push({
      code,
      name: readGoStringField(entry, 'Name'),
      native_name: readGoStringField(entry, 'NativeName'),
      enabled: enabledMatch ? enabledMatch[1] === 'true' : false,
    })
  }

  if (!languages.length) {
    throw new Error('No languages parsed from Go locales source')
  }

  return languages
}

function loadLocalBackendLocales(): BackendLocale[] {
  if (!fs.existsSync(goLanguagesPath)) {
    throw new Error(`Go locales source not found: ${goLanguagesPath}`)
  }

  return parseGoSupportedLanguages(fs.readFileSync(goLanguagesPath, 'utf8'))
}

function stripYamlScalar(value: string): string {
  return value.trim().replace(/^['"]|['"]$/g, '')
}

function normalizeLocaleLiteral(value: string): string {
  return value.trim().replace(/_/g, '-').toLowerCase()
}

function toDisplayPath(filePath: string): string {
  return path.relative(repoRoot, filePath).replace(/\\/g, '/')
}

function lineNumberAt(source: string, index: number): number {
  return source.slice(0, index).split(/\r?\n/).length
}

function listNuxtActiveSourceFiles(dir: string): string[] {
  if (!fs.existsSync(dir)) return []

  const files: string[] = []
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      if (['.nuxt', '.output', 'node_modules', 'locales', 'messages'].includes(entry.name)) continue
      files.push(...listNuxtActiveSourceFiles(fullPath))
      continue
    }

    if (entry.isFile() && ['.ts', '.vue'].includes(path.extname(entry.name))) {
      files.push(fullPath)
    }
  }
  return files
}

function listLocaleDocumentationFiles(target: string): string[] {
  if (!fs.existsSync(target)) return []
  const stats = fs.statSync(target)
  if (stats.isFile()) return [target]

  const files: string[] = []
  for (const entry of fs.readdirSync(target, { withFileTypes: true })) {
    const fullPath = path.join(target, entry.name)
    if (entry.isDirectory()) {
      if (['archive', 'node_modules', 'dist', '.nuxt', '.output'].includes(entry.name)) continue
      files.push(...listLocaleDocumentationFiles(fullPath))
      continue
    }
    if (entry.isFile() && ['.md', '.ts', '.vue'].includes(path.extname(entry.name))) {
      files.push(fullPath)
    }
  }
  return files
}

function parseGoConfigSupportedLocales(source: string, sourcePath: string): string[] {
  const locales: string[] = []
  let inI18n = false
  let inSupportedLocales = false
  let i18nIndent = -1
  let supportedLocalesIndent = -1

  for (const rawLine of source.split(/\r?\n/)) {
    const line = rawLine.replace(/\s+#.*$/, '')
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue

    const indent = line.length - line.trimStart().length
    if (trimmed === 'i18n:') {
      inI18n = true
      inSupportedLocales = false
      i18nIndent = indent
      continue
    }

    if (inI18n && indent <= i18nIndent && !trimmed.startsWith('-')) {
      inI18n = false
      inSupportedLocales = false
    }
    if (!inI18n) continue

    if (/^supported_locales\s*:\s*$/.test(trimmed)) {
      inSupportedLocales = true
      supportedLocalesIndent = indent
      continue
    }

    if (!inSupportedLocales) continue
    if (indent <= supportedLocalesIndent && !trimmed.startsWith('-')) {
      inSupportedLocales = false
      continue
    }
    if (trimmed.startsWith('-')) {
      locales.push(stripYamlScalar(trimmed.slice(1)))
    }
  }

  if (locales.length === 0) {
    throw new Error(`i18n.supported_locales not found in ${sourcePath}`)
  }
  return locales
}

function loadGoConfigLocaleLists(): Array<{ source: string; locales: Array<{ code: string }> }> {
  return goConfigLocalePaths.map((configPath) => {
    if (!fs.existsSync(configPath)) {
      throw new Error(`Go config not found: ${configPath}`)
    }
    return {
      source: `go config ${path.relative(repoRoot, configPath).replace(/\\/g, '/')}`,
      locales: parseGoConfigSupportedLocales(fs.readFileSync(configPath, 'utf8'), configPath)
        .map((code) => ({ code })),
    }
  })
}

async function loadBackendLocales(): Promise<BackendLocale[]> {
  if (localesUrl || apiBase) {
    return fetchBackendLocales()
  }

  return loadLocalBackendLocales()
}

function codes(list: readonly { code: string }[]): string[] {
  return list.map((item) => item.code)
}

function toMap<T extends { code?: unknown }>(list: readonly T[]): Map<string, T> {
  const map = new Map<string, T>()
  for (const item of list) {
    if (!item || !item.code) continue
    map.set(String(item.code), item)
  }
  return map
}

function addCodeFailures(
  failures: LocaleCheckFailure[],
  registry: readonly StorefrontLocaleRegistryEntry[],
  target: readonly { code: string }[],
  source: string
): void {
  const registryCodes = codes(registry)
  const targetCodes = codes(target)

  const missing = registryCodes.filter((code) => !targetCodes.includes(code))
  const extra = targetCodes.filter((code) => !registryCodes.includes(code))
  if (missing.length) failures.push({ source, message: `Missing locale codes: ${missing.join(', ')}` })
  if (extra.length) failures.push({ source, message: `Extra locale codes: ${extra.join(', ')}` })

  if (!missing.length && !extra.length && registryCodes.join('|') !== targetCodes.join('|')) {
    failures.push({
      source,
      message: `Locale order mismatch. Expected ${registryCodes.join(', ')}; got ${targetCodes.join(', ')}`,
    })
  }
}

function addFieldFailure(
  failures: LocaleCheckFailure[],
  source: string,
  code: string,
  field: string,
  expected: unknown,
  actual: unknown
): void {
  if (expected !== actual) {
    failures.push({
      source,
      message: `${code}.${field} mismatch: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`,
    })
  }
}

function compareManifest(
  registry: readonly StorefrontLocaleRegistryEntry[],
  manifest: readonly LocaleManifestEntry[],
  failures: LocaleCheckFailure[]
): void {
  const source = 'nuxt manifest'
  addCodeFailures(failures, registry, manifest, source)

  const manifestMap = toMap(manifest)
  for (const entry of registry) {
    const actual = manifestMap.get(entry.code)
    if (!actual) continue
    addFieldFailure(failures, source, entry.code, 'iso', entry.iso, actual.iso)
    addFieldFailure(failures, source, entry.code, 'name', entry.native_name, actual.name)
    addFieldFailure(failures, source, entry.code, 'file', entry.file, actual.file)
    addFieldFailure(failures, source, entry.code, 'dir', entry.dir, actual.dir)
  }
}

function compareLanguageList(
  registry: readonly StorefrontLocaleRegistryEntry[],
  list: readonly BackendLocale[],
  source: string,
  failures: LocaleCheckFailure[]
): void {
  addCodeFailures(failures, registry, list.map((item) => ({ code: item.code })), source)

  const languageMap = toMap(list)
  for (const entry of registry) {
    const actual = languageMap.get(entry.code)
    if (!actual) continue
    addFieldFailure(failures, source, entry.code, 'name', entry.name, actual.name)
    addFieldFailure(failures, source, entry.code, 'native_name', entry.native_name, actual.native_name)
    addFieldFailure(failures, source, entry.code, 'enabled', entry.enabled, actual.enabled)
  }
}

function addNuxtActiveSourceLocaleFailures(
  registry: readonly StorefrontLocaleRegistryEntry[],
  failures: LocaleCheckFailure[]
): void {
  const supportedCodes = new Set(codes(registry))

  for (const filePath of listNuxtActiveSourceFiles(nuxtActiveSourceRoot)) {
    const resolvedPath = path.resolve(filePath)
    if (nuxtLocaleDefinitionFiles.has(resolvedPath)) continue

    const source = fs.readFileSync(filePath, 'utf8')
    const supportedLiterals = new Set<string>()

    for (const match of source.matchAll(localeLiteralPattern)) {
      const literal = match[2]
      const normalizedLiteral = normalizeLocaleLiteral(literal)
      const line = lineNumberAt(source, match.index || 0)

      if (
        !supportedCodes.has(literal) &&
        legacyUnsupportedStorefrontLocaleLiterals.has(normalizedLiteral)
      ) {
        failures.push({
          source: 'nuxt active source',
          message: `${toDisplayPath(filePath)}:${line} uses legacy/unsupported storefront locale literal ${JSON.stringify(literal)}. Use normalizeStorefrontLocaleCode() and the fixed manifest registry.`,
        })
      }

      if (supportedCodes.has(literal)) {
        supportedLiterals.add(literal)
      }
    }

    if (supportedLiterals.size >= hardcodedLocaleListThreshold) {
      failures.push({
        source: 'nuxt active source',
        message: `${toDisplayPath(filePath)} hardcodes ${supportedLiterals.size} storefront locale codes (${Array.from(supportedLiterals).join(', ')}). Use app/i18n/locales.manifest.ts or ~/utils/storefrontLocales instead.`,
      })
    }
  }
}

function addStaleLocaleDocumentationFailures(failures: LocaleCheckFailure[]): void {
  const files = staleLocaleDocumentationTargets.flatMap(listLocaleDocumentationFiles)

  for (const filePath of files) {
    const source = fs.readFileSync(filePath, 'utf8')
    for (const check of staleLocaleDocumentationPatterns) {
      const match = source.match(check.pattern)
      if (!match || match.index === undefined) continue

      failures.push({
        source: 'locale docs/examples',
        message: `${toDisplayPath(filePath)}:${lineNumberAt(source, match.index)} contains ${check.label}. Keep docs/examples on the fixed 20 storefront locale registry.`,
      })
    }
  }
}

async function main(): Promise<void> {
  try {
    const [registry, manifest, backendLocales, goConfigLocaleLists] = await Promise.all([
      Promise.resolve(loadRegistryLocales()),
      Promise.resolve(loadManifestLocales()),
      loadBackendLocales(),
      Promise.resolve(loadGoConfigLocaleLists()),
    ])

    const failures: LocaleCheckFailure[] = []
    compareManifest(registry, manifest, failures)
    compareLanguageList(registry, backendLocales, 'go backend locales', failures)
    compareLanguageList(registry, adminFallbackLocales, 'admin fallback locales', failures)
    for (const configLocaleList of goConfigLocaleLists) {
      addCodeFailures(failures, registry, configLocaleList.locales, configLocaleList.source)
    }
    addNuxtActiveSourceLocaleFailures(registry, failures)
    addStaleLocaleDocumentationFailures(failures)

    if (failures.length === 0) {
      console.log(`Storefront locale registry is aligned across ${registry.length} locales.`)
      process.exit(0)
    }

    for (const failure of failures) {
      console.error(`[${failure.source}] ${failure.message}`)
    }
    process.exit(1)
  } catch (error: unknown) {
    console.error('check-locales failed:', error instanceof Error ? error.message : error)
    process.exit(1)
  }
}

main()

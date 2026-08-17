import fs from 'node:fs'
import path from 'node:path'
import locales, { type LocaleManifestEntry } from '../app/i18n/locales.manifest.ts'

const baseFontStack = ['StorefrontSystemLatin', 'StorefrontSystem']

export const normalizeStorefrontLocaleCode = (locale: string): string => (
  locale.trim().replace(/_/g, '-').toLowerCase()
)

const collectJSONFiles = (directory: string): string[] => {
  if (!fs.existsSync(directory)) return []

  const files: string[] = []
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const entryPath = path.join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...collectJSONFiles(entryPath))
    } else if (entry.isFile() && entry.name.endsWith('.json')) {
      files.push(entryPath)
    }
  }
  return files
}

export const collectStorefrontLocaleSources = (projectDir: string): Map<string, string[]> => {
  const localeRoot = path.join(projectDir, 'app', 'i18n', 'locales')
  const messagesRoot = path.join(projectDir, 'app', 'i18n', 'messages')
  const sources = new Map<string, string[]>()

  for (const sourcePath of collectJSONFiles(localeRoot)) {
    const locale = normalizeStorefrontLocaleCode(path.basename(sourcePath, '.json'))
    sources.set(locale, [...(sources.get(locale) || []), sourcePath])
  }

  for (const sourcePath of collectJSONFiles(messagesRoot)) {
    const relativePath = path.relative(messagesRoot, sourcePath)
    const locale = normalizeStorefrontLocaleCode(relativePath.split(path.sep)[0] || '')
    if (!locale) continue
    sources.set(locale, [...(sources.get(locale) || []), sourcePath])
  }

  return sources
}

export const fontStackForStorefrontLocale = (
  locale: string,
  manifestEntries: LocaleManifestEntry[] = locales,
): string[] => {
  const entry = manifestEntries.find(candidate => (
    normalizeStorefrontLocaleCode(candidate.code) === normalizeStorefrontLocaleCode(locale)
  ))

  switch (entry?.fontFamily) {
    case 'arabic':
      return ['StorefrontSystemArabic', ...baseFontStack]
    case 'devanagari':
      return ['StorefrontSystemDevanagari', ...baseFontStack]
    case 'thai':
      return ['StorefrontSystemThai', ...baseFontStack]
    case 'latin-accent':
      return ['StorefrontSystemLatinAccents', ...baseFontStack]
    default:
      return [...baseFontStack]
  }
}

export const validateStorefrontLocaleSources = (
  projectDir: string,
  sources: Map<string, string[]>,
  manifestEntries: LocaleManifestEntry[] = locales,
): string[] => {
  const violations: string[] = []
  const localeRoot = path.join(projectDir, 'app', 'i18n', 'locales')
  const configuredLocales = new Map(
    manifestEntries.map(entry => [normalizeStorefrontLocaleCode(entry.code), entry]),
  )
  const configuredLocaleOccurrences = new Map<string, string[]>()

  for (const entry of manifestEntries) {
    const localeCode = normalizeStorefrontLocaleCode(entry.code)
    configuredLocaleOccurrences.set(localeCode, [...(configuredLocaleOccurrences.get(localeCode) || []), entry.code])
  }

  if (configuredLocales.size === 0) {
    violations.push('No enabled locales are declared in app/i18n/locales.manifest.ts.')
  }

  for (const [localeCode, entries] of configuredLocaleOccurrences) {
    if (entries.length > 1) {
      violations.push(`Locale ${localeCode} is declared more than once in app/i18n/locales.manifest.ts: ${entries.join(', ')}.`)
    }
  }

  for (const [localeCode, entry] of configuredLocales) {
    const localeFile = path.join(localeRoot, entry.file)
    if (!fs.existsSync(localeFile)) {
      violations.push(`Configured locale ${entry.code} is missing ${path.relative(projectDir, localeFile).replace(/\\/g, '/')}.`)
    }
    if ((sources.get(localeCode) || []).length === 0) {
      violations.push(`Configured locale ${entry.code} has no locale or message JSON resources.`)
    }
  }

  for (const localeCode of sources.keys()) {
    if (!configuredLocales.has(localeCode)) {
      violations.push(`Locale resources exist for ${localeCode}, but it is not declared in app/i18n/locales.manifest.ts.`)
    }
  }

  return violations
}

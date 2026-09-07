import fs from 'node:fs'
import path from 'node:path'
import {
  collectJsonFilesDeep,
  findDuplicateJsonKeys,
  flattenKeys,
  listJsonFiles,
  loadManifestLocales,
  messagesDir,
  pageMessagesDir,
  readJson,
  isPlainObject,
  type JsonObject,
} from './lib.js'

const baseLocale = 'en'
const limitArg = process.argv.find((arg) => arg.startsWith('--limit='))
const detailLimit = Number(limitArg?.split('=')[1] || 40)

function buildLocaleFromMessages(localeCode: string): JsonObject {
  const localeDir = path.resolve(messagesDir, localeCode)
  if (!fs.existsSync(localeDir)) {
    throw new Error(`Missing split locale directory: ${path.relative(process.cwd(), localeDir)}`)
  }

  const result: JsonObject = {}
  const domains = new Map<string, string>()
  for (const filePath of listJsonFiles(localeDir)) {
    const domain = path.basename(filePath, '.json')
    const normalized = domain.toLowerCase()
    if (domains.has(normalized)) {
      throw new Error(`Duplicate domain file "${domain}" in ${path.relative(process.cwd(), localeDir)}`)
    }
    domains.set(normalized, filePath)
    result[domain] = readJson(filePath)
  }
  return result
}

function normalizeLocale(locale: string): string {
  return locale.trim().replace(/_/g, '-').toLowerCase()
}

function listPageMessageNamespaces(): string[] {
  if (!fs.existsSync(pageMessagesDir)) return []

  return fs.readdirSync(pageMessagesDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort((a, b) => a.localeCompare(b))
}

function buildPageMessages(): Map<string, Map<string, JsonObject>> {
  const namespaces = new Map<string, Map<string, JsonObject>>()

  for (const namespace of listPageMessageNamespaces()) {
    const namespaceDir = path.resolve(pageMessagesDir, namespace)
    const localeMessages = new Map<string, JsonObject>()

    for (const filePath of listJsonFiles(namespaceDir)) {
      const locale = normalizeLocale(path.basename(filePath, '.json'))
      if (!locale) continue
      if (localeMessages.has(locale)) {
        throw new Error(`Duplicate page message locale "${locale}" in ${namespaceDir}`)
      }

      const value = readJson(filePath)
      if (!isPlainObject(value)) {
        throw new Error(`Page message resource must be an object: ${path.relative(process.cwd(), filePath)}`)
      }
      localeMessages.set(locale, value)
    }

    namespaces.set(namespace, localeMessages)
  }

  return namespaces
}

function pageKeySet(namespace: string, messages: JsonObject): Set<string> {
  return new Set(
    [...flattenKeys(messages)].map((key) => `${namespace}.${key}`),
  )
}

function walkAppFiles(dir: string, out: string[] = []): string[] {
  if (!fs.existsSync(dir)) return out
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      if (!['node_modules', '.nuxt', '.output', 'dist'].includes(entry.name)) {
        walkAppFiles(fullPath, out)
      }
    } else if (['.vue', '.ts', '.js'].includes(path.extname(entry.name))) {
      out.push(fullPath)
    }
  }
  return out
}

function collectStaticTranslationRefs(): Map<string, string[]> {
  const refs = new Map<string, string[]>()
  const appDir = path.resolve(process.cwd(), 'app')
  const callPattern = /(?:\$t|\bt|\btm)\(\s*(['"`])([^'"`]+)\1/g
  const routeMetaPattern = /\bfooterLabelKey\s*:\s*(['"`])([^'"`]+)\1/g

  const addRef = (key: string, filePath: string, text: string, offset: number) => {
    if (!/^[A-Za-z0-9_.-]+$/.test(key)) return
    if (!refs.has(key)) refs.set(key, [])
    const line = text.slice(0, offset).split(/\r?\n/).length
    refs.get(key)?.push(`${path.relative(process.cwd(), filePath)}:${line}`)
  }

  for (const filePath of walkAppFiles(appDir)) {
    const text = fs.readFileSync(filePath, 'utf8')
    let match: RegExpExecArray | null
    while ((match = callPattern.exec(text))) {
      addRef(match[2], filePath, text, match.index)
    }
    while ((match = routeMetaPattern.exec(text))) {
      addRef(match[2], filePath, text, match.index)
    }
  }

  return refs
}

function printList<T>(
  title: string,
  values: T[],
  formatter: (value: T) => string = (value) => String(value),
): void {
  if (!values.length) return
  console.error(title)
  for (const value of values.slice(0, detailLimit)) {
    console.error(`  ${formatter(value)}`)
  }
  if (values.length > detailLimit) {
    console.error(`  ...and ${values.length - detailLimit} more`)
  }
}

interface LocaleReport {
  code: string
  missingFromBase: string[]
  extraVsBase: string[]
  missingUsed: string[]
}

interface PageLocaleReport {
  code: string
  namespace: string
  missingFromBase: string[]
  extraVsBase: string[]
  missingUsed: string[]
}

async function main(): Promise<void> {
  const locales = await loadManifestLocales()
  const localeCodes = locales.map((locale) => locale.code)
  const duplicateKeys = []
  const pageMessages = buildPageMessages()

  for (const filePath of collectJsonFilesDeep(messagesDir)) {
    for (const duplicate of findDuplicateJsonKeys(filePath)) {
      duplicateKeys.push({
        file: path.relative(process.cwd(), filePath),
        key: duplicate.key,
        offset: duplicate.duplicateOffset,
      })
    }
  }
  for (const filePath of collectJsonFilesDeep(pageMessagesDir)) {
    for (const duplicate of findDuplicateJsonKeys(filePath)) {
      duplicateKeys.push({
        file: path.relative(process.cwd(), filePath),
        key: duplicate.key,
        offset: duplicate.duplicateOffset,
      })
    }
  }

  const builtLocales = new Map<string, JsonObject>()
  for (const code of localeCodes) {
    builtLocales.set(code, buildLocaleFromMessages(code))
  }

  const baseMessages = builtLocales.get(baseLocale)
  if (!baseMessages) {
    throw new Error(`Base locale "${baseLocale}" is missing from manifest`)
  }

  const pageKeys = new Set<string>()
  const pageBaseKeys = new Set<string>()
  const pageNamespacesMissingBase: string[] = []
  for (const [namespace, localeMessages] of pageMessages) {
    const basePageMessage = localeMessages.get(normalizeLocale(baseLocale))
    if (!basePageMessage) {
      pageNamespacesMissingBase.push(namespace)
      continue
    }
    const keys = pageKeySet(namespace, basePageMessage)
    keys.forEach((key) => pageBaseKeys.add(key))
    keys.forEach((key) => pageKeys.add(key))
    for (const localeMessage of localeMessages.values()) {
      pageKeySet(namespace, localeMessage).forEach((key) => pageKeys.add(key))
    }
  }

  const baseKeys = flattenKeys(baseMessages)
  const staticRefs = collectStaticTranslationRefs()
  const usedKeys = [...staticRefs.keys()].sort()
  const knownPageNamespaces = [...pageMessages.keys()]
  const isPageKey = (key: string) => pageKeys.has(key) || (
    knownPageNamespaces.some(
      (namespace) => key === namespace || key.startsWith(`${namespace}.`),
    ) && !baseKeys.has(key)
  )
  const globalUsedKeys = usedKeys.filter((key) => !isPageKey(key))
  const pageUsedKeys = usedKeys.filter((key) => isPageKey(key))
  const baseMissingUsedKeys = globalUsedKeys.filter((key) => !baseKeys.has(key))
  const pageMissingUsedKeys = pageUsedKeys.filter((key) => !pageBaseKeys.has(key))
  const pageKeysInGlobal = [...baseKeys].filter((key) => isPageKey(key))

  const localeReports: LocaleReport[] = []
  for (const code of localeCodes) {
    if (code === baseLocale) continue
    const localeMessages = builtLocales.get(code)
    if (!localeMessages) {
      throw new Error(`Locale "${code}" is missing from built messages`)
    }
    const keys = flattenKeys(localeMessages)
    localeReports.push({
      code,
      missingFromBase: [...baseKeys].filter((key) => !keys.has(key)).sort(),
      extraVsBase: [...keys].filter((key) => !baseKeys.has(key)).sort(),
      missingUsed: globalUsedKeys.filter((key) => !keys.has(key)).sort(),
    })
  }

  const pageLocaleReports: PageLocaleReport[] = []
  for (const [namespace, localeMessages] of pageMessages) {
    const basePageMessage = localeMessages.get(normalizeLocale(baseLocale))
    if (!basePageMessage) continue
    const basePageKeys = pageKeySet(namespace, basePageMessage)

    for (const locale of locales) {
      if (locale.code === baseLocale) continue
      const localeMessage = localeMessages.get(normalizeLocale(locale.code))
      if (!localeMessage) {
        continue
      }

      const keys = pageKeySet(namespace, localeMessage)
      pageLocaleReports.push({
        code: locale.code,
        namespace,
        missingFromBase: [...basePageKeys].filter((key) => !keys.has(key)).sort(),
        extraVsBase: [...keys].filter((key) => !basePageKeys.has(key)).sort(),
        missingUsed: pageUsedKeys.filter((key) => (
          key.startsWith(`${namespace}.`) && !keys.has(key)
        )).sort(),
      })
    }
  }

  console.log(`Locales checked: ${localeCodes.length}`)
  console.log(`Base keys (${baseLocale}): ${baseKeys.size}`)
  console.log(`Static translation refs: ${usedKeys.length}`)
  console.log(`Page message namespaces: ${pageMessages.size}`)
  console.log(`Page message keys (${baseLocale}): ${pageBaseKeys.size}`)
  console.log(`Page message locale files: ${
    [...pageMessages.values()].reduce((total, localeMessages) => total + localeMessages.size, 0)
  }`)
  console.log(`Duplicate JSON keys: ${duplicateKeys.length}`)
  console.log(`Global used keys missing in ${baseLocale}: ${baseMissingUsedKeys.length}`)
  console.log(`Page used keys missing in ${baseLocale}: ${pageMissingUsedKeys.length}`)

  printList(
    '\nDuplicate JSON keys:',
    duplicateKeys,
    (item) => `${item.file}: "${item.key}" near offset ${item.offset}`,
  )

  printList(
    `\nGlobal used keys missing in ${baseLocale}:`,
    baseMissingUsedKeys,
    (key) => `${key} <- ${staticRefs.get(key)?.slice(0, 3).join(', ') || 'no static ref'}`,
  )

  printList(
    `\nPage used keys missing in ${baseLocale}:`,
    pageMissingUsedKeys,
    (key) => `${key} <- ${staticRefs.get(key)?.slice(0, 3).join(', ') || 'no static ref'}`,
  )

  printList(
    '\nPage namespaces missing English base resources:',
    pageNamespacesMissingBase,
  )

  printList(
    '\nPage keys incorrectly included in global locale:',
    pageKeysInGlobal,
  )

  for (const report of localeReports) {
    if (!report.missingFromBase.length && !report.extraVsBase.length && !report.missingUsed.length) {
      continue
    }
    console.error(
      `\n${report.code}: missing ${report.missingFromBase.length} base keys, ` +
      `extra ${report.extraVsBase.length}, missing ${report.missingUsed.length} used keys`,
    )
    printList('  Missing base keys:', report.missingFromBase)
    printList('  Extra keys:', report.extraVsBase)
    printList(
      '  Missing used keys:',
      report.missingUsed,
      (key) => `${key} <- ${staticRefs.get(key)?.slice(0, 2).join(', ') || 'no static ref'}`,
    )
  }

  for (const report of pageLocaleReports) {
    if (
      !report.missingFromBase.length &&
      !report.extraVsBase.length &&
      !report.missingUsed.length
    ) {
      continue
    }
    console.error(
      `\n${report.namespace}/${report.code}: missing ${report.missingFromBase.length} page keys, ` +
      `extra ${report.extraVsBase.length}, missing ${report.missingUsed.length} used keys`,
    )
    printList('  Missing page keys:', report.missingFromBase)
    printList('  Extra page keys:', report.extraVsBase)
    printList(
      '  Missing used page keys:',
      report.missingUsed,
      (key) => `${key} <- ${staticRefs.get(key)?.slice(0, 2).join(', ') || 'no static ref'}`,
    )
  }

  const hasLocaleFailures = localeReports.some(
    (report) => report.missingFromBase.length || report.missingUsed.length,
  )
  const hasPageFailures = pageNamespacesMissingBase.length ||
    pageMissingUsedKeys.length ||
    pageKeysInGlobal.length

  if (duplicateKeys.length || baseMissingUsedKeys.length || hasLocaleFailures || hasPageFailures) {
    process.exit(1)
  }

  if (
    localeReports.some((report) => report.extraVsBase.length) ||
    pageLocaleReports.some((report) => report.extraVsBase.length)
  ) {
    console.log('Extra locale or page message keys are reported as cleanup warnings only.')
  }

  console.log('i18n messages are complete and aligned.')
}

main().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error)
  process.exit(1)
})

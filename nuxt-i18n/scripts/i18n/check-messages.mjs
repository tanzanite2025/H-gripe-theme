import fs from 'node:fs'
import path from 'node:path'
import {
  collectJsonFilesDeep,
  findDuplicateJsonKeys,
  flattenKeys,
  listJsonFiles,
  loadManifestLocales,
  messagesDir,
  readJson,
} from './lib.mjs'

const baseLocale = 'en'
const limitArg = process.argv.find((arg) => arg.startsWith('--limit='))
const detailLimit = Number(limitArg?.split('=')[1] || 40)

function buildLocaleFromMessages(localeCode) {
  const localeDir = path.resolve(messagesDir, localeCode)
  if (!fs.existsSync(localeDir)) {
    throw new Error(`Missing split locale directory: ${path.relative(process.cwd(), localeDir)}`)
  }

  const result = {}
  const domains = new Map()
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

function walkAppFiles(dir, out = []) {
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

function collectStaticTranslationRefs() {
  const refs = new Map()
  const appDir = path.resolve(process.cwd(), 'app')
  const callPattern = /(?:\$t|\bt|\btm)\(\s*(['"`])([^'"`]+)\1/g

  for (const filePath of walkAppFiles(appDir)) {
    const text = fs.readFileSync(filePath, 'utf8')
    let match
    while ((match = callPattern.exec(text))) {
      const key = match[2]
      if (!/^[A-Za-z0-9_.-]+$/.test(key)) continue
      if (!refs.has(key)) refs.set(key, [])
      const line = text.slice(0, match.index).split(/\r?\n/).length
      refs.get(key).push(`${path.relative(process.cwd(), filePath)}:${line}`)
    }
  }

  return refs
}

function printList(title, values, formatter = (value) => value) {
  if (!values.length) return
  console.error(title)
  for (const value of values.slice(0, detailLimit)) {
    console.error(`  ${formatter(value)}`)
  }
  if (values.length > detailLimit) {
    console.error(`  ...and ${values.length - detailLimit} more`)
  }
}

async function main() {
  const locales = await loadManifestLocales()
  const localeCodes = locales.map((locale) => locale.code)
  const duplicateKeys = []

  for (const filePath of collectJsonFilesDeep(messagesDir)) {
    for (const duplicate of findDuplicateJsonKeys(filePath)) {
      duplicateKeys.push({
        file: path.relative(process.cwd(), filePath),
        key: duplicate.key,
        offset: duplicate.duplicateOffset,
      })
    }
  }

  const builtLocales = new Map()
  for (const code of localeCodes) {
    builtLocales.set(code, buildLocaleFromMessages(code))
  }

  const baseMessages = builtLocales.get(baseLocale)
  if (!baseMessages) {
    throw new Error(`Base locale "${baseLocale}" is missing from manifest`)
  }

  const baseKeys = flattenKeys(baseMessages)
  const staticRefs = collectStaticTranslationRefs()
  const usedKeys = [...staticRefs.keys()].sort()
  const baseMissingUsedKeys = usedKeys.filter((key) => !baseKeys.has(key))

  const localeReports = []
  for (const code of localeCodes) {
    if (code === baseLocale) continue
    const keys = flattenKeys(builtLocales.get(code))
    localeReports.push({
      code,
      missingFromBase: [...baseKeys].filter((key) => !keys.has(key)).sort(),
      extraVsBase: [...keys].filter((key) => !baseKeys.has(key)).sort(),
      missingUsed: usedKeys.filter((key) => !keys.has(key)).sort(),
    })
  }

  console.log(`Locales checked: ${localeCodes.length}`)
  console.log(`Base keys (${baseLocale}): ${baseKeys.size}`)
  console.log(`Static translation refs: ${usedKeys.length}`)
  console.log(`Duplicate JSON keys: ${duplicateKeys.length}`)
  console.log(`Used keys missing in ${baseLocale}: ${baseMissingUsedKeys.length}`)

  printList(
    `\nDuplicate JSON keys:`,
    duplicateKeys,
    (item) => `${item.file}: "${item.key}" near offset ${item.offset}`,
  )

  printList(
    `\nUsed keys missing in ${baseLocale}:`,
    baseMissingUsedKeys,
    (key) => `${key} <- ${staticRefs.get(key).slice(0, 3).join(', ')}`,
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

  const hasLocaleFailures = localeReports.some(
    (report) => report.missingFromBase.length || report.missingUsed.length,
  )

  if (duplicateKeys.length || baseMissingUsedKeys.length || hasLocaleFailures) {
    process.exit(1)
  }

  if (localeReports.some((report) => report.extraVsBase.length)) {
    console.log('Extra locale keys are reported as cleanup warnings only.')
  }

  console.log('i18n messages are complete and aligned.')
}

main().catch((error) => {
  console.error(error.message || error)
  process.exit(1)
})

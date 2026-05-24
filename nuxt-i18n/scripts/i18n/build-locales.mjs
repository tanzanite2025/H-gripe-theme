import fs from 'node:fs'
import path from 'node:path'
import {
  deepEqual,
  listJsonFiles,
  loadManifestLocales,
  localeFilePath,
  messagesDir,
  readJson,
  writeJson,
} from './lib.mjs'

const checkOnly = process.argv.includes('--check')

function buildLocaleObject(localeCode) {
  const localeDir = path.resolve(messagesDir, localeCode)
  if (!fs.existsSync(localeDir)) {
    throw new Error(`Missing split locale directory: ${path.relative(process.cwd(), localeDir)}`)
  }

  const domains = {}
  const seen = new Map()
  for (const filePath of listJsonFiles(localeDir)) {
    const domain = path.basename(filePath, '.json')
    const normalized = domain.toLowerCase()
    if (seen.has(normalized)) {
      throw new Error(`Duplicate domain files for "${domain}" in ${localeDir}`)
    }
    seen.set(normalized, filePath)
    domains[domain] = readJson(filePath)
  }
  return domains
}

async function main() {
  const locales = await loadManifestLocales()
  const changed = []

  for (const locale of locales) {
    const nextValue = buildLocaleObject(locale.code)
    const targetPath = localeFilePath(locale)

    if (checkOnly) {
      const currentValue = fs.existsSync(targetPath) ? readJson(targetPath) : null
      if (!deepEqual(currentValue, nextValue)) {
        changed.push(path.relative(process.cwd(), targetPath))
      }
      continue
    }

    writeJson(targetPath, nextValue)
    changed.push(path.relative(process.cwd(), targetPath))
  }

  if (checkOnly) {
    if (changed.length) {
      console.error('Aggregated locale files are out of sync with i18n/messages:')
      for (const file of changed) console.error(`  ${file}`)
      process.exit(1)
    }
    console.log('Aggregated locale files are in sync with i18n/messages.')
    return
  }

  console.log(`Built ${changed.length} aggregated locale files from i18n/messages.`)
}

main().catch((error) => {
  console.error(error.message || error)
  process.exit(1)
})

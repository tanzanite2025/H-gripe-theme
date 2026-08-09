import fs from 'node:fs'
import path from 'node:path'
import {
  assertInside,
  isPlainObject,
  loadManifestLocales,
  localeFilePath,
  messagesDir,
  readJson,
  writeJson,
} from './lib.js'

const force = process.argv.includes('--force')

async function main(): Promise<void> {
  if (fs.existsSync(messagesDir) && !force) {
    throw new Error('i18n/messages already exists. Use --force to regenerate split files from i18n/locales.')
  }

  const locales = await loadManifestLocales()
  let written = 0

  for (const locale of locales) {
    const sourcePath = localeFilePath(locale)
    const source = readJson(sourcePath)
    if (!isPlainObject(source)) {
      throw new Error(`Locale "${locale.code}" must contain a JSON object at ${sourcePath}`)
    }
    const localeDir = path.resolve(messagesDir, locale.code)
    assertInside(messagesDir, localeDir)

    for (const [domain, value] of Object.entries(source)) {
      const targetPath = path.resolve(localeDir, `${domain}.json`)
      assertInside(localeDir, targetPath)
      writeJson(targetPath, value)
      written += 1
    }
  }

  console.log(`Split ${locales.length} locale files into ${written} domain files under i18n/messages.`)
}

main().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error)
  process.exit(1)
})

import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import type { LocaleManifestEntry } from '../app/i18n/locales.manifest.ts'
import {
  collectStorefrontLocaleSources,
  fontStackForStorefrontLocale,
  validateStorefrontLocaleSources,
} from './font-locale-contract.ts'

const manifest: LocaleManifestEntry[] = [
  { code: 'en', iso: 'en-US', name: 'English', file: 'en.json' },
  { code: 'ar', iso: 'ar-SA', name: 'Arabic', file: 'ar.json', fontFamily: 'arabic' },
]
const projectDir = fs.mkdtempSync(path.join(os.tmpdir(), 'storefront-font-locale-contract-'))

try {
  const localeDir = path.join(projectDir, 'app', 'i18n', 'locales')
  const messagesDir = path.join(projectDir, 'app', 'i18n', 'messages')
  fs.mkdirSync(path.join(messagesDir, 'en'), { recursive: true })
  fs.mkdirSync(path.join(messagesDir, 'ar'), { recursive: true })
  fs.mkdirSync(localeDir, { recursive: true })
  fs.writeFileSync(path.join(localeDir, 'en.json'), '{"title":"English"}')
  fs.writeFileSync(path.join(localeDir, 'ar.json'), '{"title":"العربية"}')
  fs.writeFileSync(path.join(messagesDir, 'en', 'checkout.json'), '{"button":"Pay"}')

  const validSources = collectStorefrontLocaleSources(projectDir)
  assert.equal(validSources.get('en')?.length, 2)
  assert.equal(validSources.get('ar')?.length, 1)
  assert.deepEqual(validateStorefrontLocaleSources(projectDir, validSources, manifest), [])
  assert.deepEqual(
    fontStackForStorefrontLocale('ar', manifest),
    ['StorefrontSystemArabic', 'StorefrontSystemLatin', 'StorefrontSystem'],
  )

  fs.rmSync(path.join(localeDir, 'ar.json'))
  const missingFileViolations = validateStorefrontLocaleSources(
    projectDir,
    collectStorefrontLocaleSources(projectDir),
    manifest,
  )
  assert.ok(missingFileViolations.some(violation => violation.includes('Configured locale ar is missing')))

  fs.mkdirSync(path.join(messagesDir, 'orphan'), { recursive: true })
  fs.writeFileSync(path.join(messagesDir, 'orphan', 'copy.json'), '{"title":"Orphan"}')
  const orphanViolations = validateStorefrontLocaleSources(
    projectDir,
    collectStorefrontLocaleSources(projectDir),
    manifest,
  )
  assert.ok(orphanViolations.some(violation => violation.includes('orphan, but it is not declared')))

  const duplicateViolations = validateStorefrontLocaleSources(
    projectDir,
    collectStorefrontLocaleSources(projectDir),
    [...manifest, { ...manifest[0], code: 'EN' }],
  )
  assert.ok(duplicateViolations.some(violation => violation.includes('en is declared more than once')))
} finally {
  fs.rmSync(projectDir, { recursive: true, force: true })
}

console.log('Font locale edge-case contract passed.')

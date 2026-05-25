import fs from 'node:fs'
import path from 'node:path'
import {
  flattenKeys,
  isPlainObject,
  loadManifestLocales,
  messagesDir,
  readJson,
  writeJson,
} from './lib.mjs'

const baseLocale = 'en'

const manualFallbacks = {
  'feedbackForm.messages.loginRequired': 'Log in to attach files.',
  'feedbackForm.messages.membersOnly': 'Attachments are available for eligible members.',
  'feedbackForm.messages.submitError': 'We could not submit your feedback. Please try again.',
  'policyTabs.cookie': 'Cookie',
  'policyTabs.privacy': 'Privacy',
  'policyTabs.refundReturn': 'Refund & Return',
  'policyTabs.terms': 'Terms',
  'warranty.errors.check_tips.0': 'Check that the product code was entered exactly as printed.',
  'warranty.errors.check_tips.1': 'Contact support if your product code still cannot be found.',
}

function clone(value) {
  return value === undefined ? undefined : JSON.parse(JSON.stringify(value))
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

function decodeStringLiteral(raw) {
  return raw
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\r')
    .replace(/\\t/g, '\t')
    .replace(/\\'/g, "'")
    .replace(/\\"/g, '"')
    .replace(/\\\\/g, '\\')
}

function collectStaticFallbacks() {
  const refs = new Map()
  const appDir = path.resolve(process.cwd(), 'app')
  const callPattern = /(?:\$t|\bt|\btm)\(\s*(['"`])([^'"`]+)\1\s*(?:,\s*(['"`])((?:\\.|(?!\3).)*)\3)?/g

  for (const filePath of walkAppFiles(appDir)) {
    const text = fs.readFileSync(filePath, 'utf8')
    let match
    while ((match = callPattern.exec(text))) {
      const key = match[2]
      if (!/^[A-Za-z0-9_.-]+$/.test(key)) continue

      const line = text.slice(0, match.index).split(/\r?\n/).length
      const fallback = match[4] === undefined ? undefined : decodeStringLiteral(match[4])
      if (!refs.has(key)) refs.set(key, { fallback, locations: [] })
      const entry = refs.get(key)
      if (!entry.fallback && fallback) entry.fallback = fallback
      entry.locations.push(`${path.relative(process.cwd(), filePath)}:${line}`)
    }
  }

  return refs
}

function readDomain(localeCode, domain) {
  const filePath = path.resolve(messagesDir, localeCode, `${domain}.json`)
  if (!fs.existsSync(filePath)) return {}
  const value = readJson(filePath)
  return isPlainObject(value) ? value : {}
}

function writeDomain(localeCode, domain, value) {
  const filePath = path.resolve(messagesDir, localeCode, `${domain}.json`)
  writeJson(filePath, value)
}

function getByPath(root, keyPath) {
  let cursor = root
  for (const part of keyPath.split('.')) {
    if (!isPlainObject(cursor) || !Object.prototype.hasOwnProperty.call(cursor, part)) {
      return undefined
    }
    cursor = cursor[part]
  }
  return cursor
}

function setByPath(root, keyPath, value) {
  const parts = keyPath.split('.')
  let cursor = root
  for (const part of parts.slice(0, -1)) {
    if (!isPlainObject(cursor[part])) cursor[part] = {}
    cursor = cursor[part]
  }
  const leaf = parts[parts.length - 1]
  if (!Object.prototype.hasOwnProperty.call(cursor, leaf)) {
    cursor[leaf] = value
    return true
  }
  return false
}

function humanizeKey(key) {
  return key
    .split('.')
    .filter((part) => !/^\d+$/.test(part))
    .slice(-1)[0]
    ?.replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase()) || key
}

function buildLocaleObject(localeCode) {
  const localeDir = path.resolve(messagesDir, localeCode)
  const result = {}
  if (!fs.existsSync(localeDir)) return result
  for (const fileName of fs.readdirSync(localeDir).filter((name) => name.endsWith('.json')).sort()) {
    const domain = path.basename(fileName, '.json')
    result[domain] = readJson(path.join(localeDir, fileName))
  }
  return result
}

function setFullKey(localeCode, fullKey, value) {
  const [domain, ...rest] = fullKey.split('.')
  if (!domain || rest.length === 0) return false
  const domainValue = readDomain(localeCode, domain)
  const changed = setByPath(domainValue, rest.join('.'), clone(value))
  if (changed) writeDomain(localeCode, domain, domainValue)
  return changed
}

async function main() {
  const locales = await loadManifestLocales()
  const fallbackRefs = collectStaticFallbacks()
  let baseAdded = 0
  let localeAdded = 0
  const unresolved = []

  let baseObject = buildLocaleObject(baseLocale)
  let baseKeys = flattenKeys(baseObject)

  for (const [key, value] of Object.entries(manualFallbacks)) {
    if (baseKeys.has(key)) continue
    if (setFullKey(baseLocale, key, value)) {
      baseAdded += 1
    }
  }

  baseObject = buildLocaleObject(baseLocale)
  baseKeys = flattenKeys(baseObject)

  for (const [key, ref] of fallbackRefs.entries()) {
    if (baseKeys.has(key)) continue
    const value = ref.fallback || manualFallbacks[key] || humanizeKey(key)
    if (!value) {
      unresolved.push(`${key} <- ${ref.locations.slice(0, 3).join(', ')}`)
      continue
    }
    if (setFullKey(baseLocale, key, value)) {
      baseAdded += 1
    }
  }

  baseObject = buildLocaleObject(baseLocale)
  baseKeys = flattenKeys(baseObject)

  for (const locale of locales) {
    if (locale.code === baseLocale) continue
    const localeObject = buildLocaleObject(locale.code)
    const localeKeys = flattenKeys(localeObject)
    for (const key of baseKeys) {
      if (localeKeys.has(key)) continue
      const value = getByPath(baseObject, key)
      if (value === undefined) continue
      if (setFullKey(locale.code, key, value)) {
        localeAdded += 1
      }
    }
  }

  console.log(`Added ${baseAdded} missing static keys to ${baseLocale}.`)
  console.log(`Added ${localeAdded} missing locale keys using ${baseLocale} fallback values.`)
  if (unresolved.length) {
    console.error('Unresolved keys:')
    for (const item of unresolved) console.error(`  ${item}`)
    process.exit(1)
  }
}

main().catch((error) => {
  console.error(error.message || error)
  process.exit(1)
})

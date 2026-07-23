import fs from 'node:fs'
import path from 'node:path'

export const rootDir = process.cwd()
const i18nRoot = fs.existsSync(path.resolve(rootDir, 'app/i18n'))
  ? path.resolve(rootDir, 'app/i18n')
  : path.resolve(rootDir, 'i18n')

export const manifestPath = path.resolve(i18nRoot, 'locales.manifest.js')
export const aggregateLocalesDir = path.resolve(i18nRoot, 'locales')
export const messagesDir = path.resolve(i18nRoot, 'messages')

export function stripBom(text) {
  return text.charCodeAt(0) === 0xfeff ? text.slice(1) : text
}

export function readJson(filePath) {
  return JSON.parse(stripBom(fs.readFileSync(filePath, 'utf8')))
}

export function writeJson(filePath, value) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true })
  fs.writeFileSync(filePath, `${JSON.stringify(value, null, 2)}\n`, 'utf8')
}

export function isPlainObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

export function flattenKeys(value, prefix = '', out = new Set()) {
  if (isPlainObject(value)) {
    for (const [key, child] of Object.entries(value)) {
      flattenKeys(child, prefix ? `${prefix}.${key}` : key, out)
    }
    return out
  }

  out.add(prefix)
  return out
}

export function deepEqual(a, b) {
  if (a === b) return true
  if (Array.isArray(a) || Array.isArray(b)) {
    if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) return false
    return a.every((item, index) => deepEqual(item, b[index]))
  }
  if (!isPlainObject(a) || !isPlainObject(b)) return false
  const aKeys = Object.keys(a)
  const bKeys = Object.keys(b)
  if (aKeys.length !== bKeys.length) return false
  return aKeys.every((key) => Object.prototype.hasOwnProperty.call(b, key) && deepEqual(a[key], b[key]))
}

export async function loadManifestLocales() {
  if (!fs.existsSync(manifestPath)) {
    throw new Error(`Manifest not found: ${manifestPath}`)
  }

  const moduleUrl = `file://${manifestPath.replace(/\\/g, '/')}`
  const mod = await import(`${moduleUrl}?t=${Date.now()}`)
  const locales = mod.default || mod.locales || []
  if (!Array.isArray(locales)) {
    throw new Error('Locale manifest must export an array')
  }
  return locales
}

export function localeFilePath(locale) {
  if (!locale?.file) {
    throw new Error(`Locale "${locale?.code || '(missing code)'}" has no file in manifest`)
  }
  return path.resolve(aggregateLocalesDir, locale.file)
}

export function assertInside(parentDir, childPath) {
  const relative = path.relative(parentDir, childPath)
  if (relative.startsWith('..') || path.isAbsolute(relative)) {
    throw new Error(`Refusing to write outside ${parentDir}: ${childPath}`)
  }
}

export function listJsonFiles(dir) {
  if (!fs.existsSync(dir)) return []
  return fs.readdirSync(dir)
    .filter((name) => name.endsWith('.json'))
    .sort((a, b) => a.localeCompare(b))
    .map((name) => path.join(dir, name))
}

export function findDuplicateJsonKeys(filePath) {
  const text = stripBom(fs.readFileSync(filePath, 'utf8'))
  const duplicates = []
  const stack = []
  let index = 0

  const currentObject = () => {
    for (let i = stack.length - 1; i >= 0; i -= 1) {
      if (stack[i].type === 'object') return stack[i]
      if (stack[i].type === 'array') return null
    }
    return null
  }

  const readString = () => {
    const start = index
    index += 1
    let value = ''
    while (index < text.length) {
      const char = text[index]
      if (char === '\\') {
        value += char
        index += 1
        if (index < text.length) {
          value += text[index]
          index += 1
        }
        continue
      }
      if (char === '"') {
        index += 1
        return { value, start, end: index }
      }
      value += char
      index += 1
    }
    return { value, start, end: index }
  }

  const skipWhitespace = (from) => {
    let cursor = from
    while (cursor < text.length && /\s/.test(text[cursor])) cursor += 1
    return cursor
  }

  while (index < text.length) {
    const char = text[index]
    if (char === '{') {
      stack.push({ type: 'object', keys: new Map() })
      index += 1
      continue
    }
    if (char === '[') {
      stack.push({ type: 'array' })
      index += 1
      continue
    }
    if (char === '}' || char === ']') {
      stack.pop()
      index += 1
      continue
    }
    if (char === '"') {
      const token = readString()
      const next = skipWhitespace(index)
      const object = currentObject()
      if (object && text[next] === ':') {
        const previous = object.keys.get(token.value)
        if (previous !== undefined) {
          duplicates.push({ key: token.value, firstOffset: previous, duplicateOffset: token.start })
        } else {
          object.keys.set(token.value, token.start)
        }
      }
      continue
    }
    index += 1
  }

  return duplicates
}

export function collectJsonFilesDeep(dir) {
  if (!fs.existsSync(dir)) return []
  const result = []
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      result.push(...collectJsonFilesDeep(fullPath))
    } else if (entry.isFile() && entry.name.endsWith('.json')) {
      result.push(fullPath)
    }
  }
  return result.sort((a, b) => a.localeCompare(b))
}

import fs from 'node:fs'
import path from 'node:path'
import type { LocaleManifestEntry } from '../../app/i18n/locales.manifest.ts'

export type JsonPrimitive = string | number | boolean | null
export type JsonValue = JsonPrimitive | JsonObject | JsonValue[]
export interface JsonObject {
  [key: string]: JsonValue
}

export interface DuplicateJsonKey {
  key: string
  firstOffset: number
  duplicateOffset: number
}

type ObjectFrame = { type: 'object'; keys: Map<string, number> }
type ArrayFrame = { type: 'array' }
type JsonFrame = ObjectFrame | ArrayFrame

export const rootDir = process.cwd()
const i18nRoot = fs.existsSync(path.resolve(rootDir, 'app/i18n'))
  ? path.resolve(rootDir, 'app/i18n')
  : path.resolve(rootDir, 'i18n')

export const manifestPath = path.resolve(i18nRoot, 'locales.manifest.ts')
export const aggregateLocalesDir = path.resolve(i18nRoot, 'locales')
export const messagesDir = path.resolve(i18nRoot, 'messages')

export function stripBom(text: string): string {
  return text.charCodeAt(0) === 0xfeff ? text.slice(1) : text
}

export function readJson(filePath: string): JsonValue {
  return JSON.parse(stripBom(fs.readFileSync(filePath, 'utf8'))) as JsonValue
}

export function writeJson(filePath: string, value: JsonValue): void {
  fs.mkdirSync(path.dirname(filePath), { recursive: true })
  fs.writeFileSync(filePath, `${JSON.stringify(value, null, 2)}\n`, 'utf8')
}

export function isPlainObject(value: unknown): value is JsonObject {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

export function flattenKeys(
  value: JsonValue,
  prefix = '',
  out = new Set<string>(),
): Set<string> {
  if (isPlainObject(value)) {
    for (const [key, child] of Object.entries(value)) {
      flattenKeys(child, prefix ? `${prefix}.${key}` : key, out)
    }
    return out
  }

  out.add(prefix)
  return out
}

export function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true
  if (Array.isArray(a) || Array.isArray(b)) {
    if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) return false
    return a.every((item, index) => deepEqual(item, b[index]))
  }
  if (!isPlainObject(a) || !isPlainObject(b)) return false
  const aKeys = Object.keys(a)
  const bKeys = Object.keys(b)
  if (aKeys.length !== bKeys.length) return false
  return aKeys.every(
    (key) => Object.prototype.hasOwnProperty.call(b, key) && deepEqual(a[key], b[key]),
  )
}

export async function loadManifestLocales(): Promise<LocaleManifestEntry[]> {
  if (!fs.existsSync(manifestPath)) {
    throw new Error(`Manifest not found: ${manifestPath}`)
  }

  const moduleUrl = `file://${manifestPath.replace(/\\/g, '/')}`
  const mod = (await import(`${moduleUrl}?t=${Date.now()}`)) as {
    default?: unknown
    locales?: unknown
  }
  const locales = mod.default || mod.locales || []
  if (!Array.isArray(locales)) {
    throw new Error('Locale manifest must export an array')
  }
  return locales as LocaleManifestEntry[]
}

export function localeFilePath(locale: LocaleManifestEntry): string {
  if (!locale.file) {
    throw new Error(`Locale "${locale.code || '(missing code)'}" has no file in manifest`)
  }
  return path.resolve(aggregateLocalesDir, locale.file)
}

export function assertInside(parentDir: string, childPath: string): void {
  const relative = path.relative(parentDir, childPath)
  if (relative.startsWith('..') || path.isAbsolute(relative)) {
    throw new Error(`Refusing to write outside ${parentDir}: ${childPath}`)
  }
}

export function listJsonFiles(dir: string): string[] {
  if (!fs.existsSync(dir)) return []
  return fs.readdirSync(dir)
    .filter((name) => name.endsWith('.json'))
    .sort((a, b) => a.localeCompare(b))
    .map((name) => path.join(dir, name))
}

export function findDuplicateJsonKeys(filePath: string): DuplicateJsonKey[] {
  const text = stripBom(fs.readFileSync(filePath, 'utf8'))
  const duplicates: DuplicateJsonKey[] = []
  const stack: JsonFrame[] = []
  let index = 0

  const currentObject = (): ObjectFrame | null => {
    for (let i = stack.length - 1; i >= 0; i -= 1) {
      const frame = stack[i]
      if (frame.type === 'object') return frame
      return null
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

  const skipWhitespace = (from: number): number => {
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
          duplicates.push({
            key: token.value,
            firstOffset: previous,
            duplicateOffset: token.start,
          })
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

export function collectJsonFilesDeep(dir: string): string[] {
  if (!fs.existsSync(dir)) return []
  const result: string[] = []
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

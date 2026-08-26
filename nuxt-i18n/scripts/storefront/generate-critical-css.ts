import { spawn } from 'node:child_process'
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from 'node:fs'
import { basename, dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { gzipSync } from 'node:zlib'
import { parse } from 'parse5'

interface ParsedHtmlNode {
  tagName?: string
  value?: string
  attrs?: Array<{
    name: string
    value: string
  }>
  childNodes?: ParsedHtmlNode[]
}

interface CriticalInventory {
  classes: Set<string>
  ids: Set<string>
  tags: Set<string>
}

interface CssRule {
  prelude: string
  content: string
  full: string
}

interface ExtractionStats {
  fontFaceRules: number
  baseRules: number
  matchedRules: number
  keyframeRules: number
  wrapperRules: number
}

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const publicNuxtDirectory = resolve(projectRoot, '.output/public/_nuxt')
const serverEntry = resolve(projectRoot, '.output/server/index.mjs')
const serverLauncher = resolve(projectRoot, 'scripts/storefront/run-production-server.mjs')
const outputDirectory = resolve(projectRoot, '.output/server/critical-css')
const outputCssPath = resolve(outputDirectory, 'home-entry.css')
const outputManifestPath = resolve(outputDirectory, 'manifest.json')
const port = Number.parseInt(process.env.CRITICAL_CSS_GENERATE_PORT || '4023', 10)
const targetPath = process.env.CRITICAL_CSS_TARGET_PATH || '/'
const origin = `http://127.0.0.1:${port}`
const maxCriticalCssGzipBytes = Number.parseInt(
  process.env.CRITICAL_CSS_MAX_GZIP_BYTES || String(16 * 1024),
  10,
)
const requiredCriticalClasses = [
  'tz-light-theme',
  'tz-text-primary',
  'tz-text-secondary',
  'site-header-layout-spacer',
  'page-content-shell',
  'premium-button',
  'premium-button--active',
  'home-hero',
  'min-h-[100dvh]',
]
const requiredGlobalCriticalClasses = [
  'site-header-layout-spacer',
  'page-content-shell',
  'premium-button',
  'premium-button--active',
  'min-h-[100dvh]',
]
const rootOnlyClasses = new Set([
  'layout',
  'layout-main',
  'site-header-layout-spacer',
  'tz-light-theme',
  'tz-text-primary',
])
const subtreeRootClasses = new Set([
  'home-hero',
  'site-header-root',
])

const timeout = (ms: number): Promise<void> => new Promise(resolveTimeout => setTimeout(resolveTimeout, ms))

const fail = (message: string): never => {
  console.error(`[critical-css-generate] FAILED: ${message}`)
  process.exit(1)
  throw new Error(message)
}

const requiredFile = (filePath: string): void => {
  if (!existsSync(filePath)) fail(`Missing ${filePath}. Run npm run build first.`)
}

const attributeValue = (node: ParsedHtmlNode, attributeName: string): string => (
  node.attrs?.find(attribute => attribute.name.toLowerCase() === attributeName)?.value || ''
)

const classList = (node: ParsedHtmlNode): string[] => (
  attributeValue(node, 'class')
    .split(/\s+/)
    .map(value => value.trim())
    .filter(Boolean)
)

const addNodeToInventory = (node: ParsedHtmlNode, inventory: CriticalInventory): void => {
  const tagName = String(node.tagName || '').toLowerCase()
  if (tagName) inventory.tags.add(tagName)

  const id = attributeValue(node, 'id')
  if (id) inventory.ids.add(id)

  for (const className of classList(node)) {
    inventory.classes.add(className)
  }
}

const addSubtreeToInventory = (node: ParsedHtmlNode, inventory: CriticalInventory): void => {
  addNodeToInventory(node, inventory)
  for (const child of node.childNodes || []) {
    addSubtreeToInventory(child, inventory)
  }
}

const collectCriticalInventory = (html: string): CriticalInventory => {
  const document = parse(html) as unknown as ParsedHtmlNode
  const inventory: CriticalInventory = {
    classes: new Set(requiredCriticalClasses),
    ids: new Set(),
    tags: new Set(['html', 'body']),
  }

  const walk = (node: ParsedHtmlNode): void => {
    const classes = classList(node)
    const tagName = String(node.tagName || '').toLowerCase()

    if (tagName === 'html' || tagName === 'body' || classes.some(className => rootOnlyClasses.has(className))) {
      addNodeToInventory(node, inventory)
    }

    if (classes.some(className => subtreeRootClasses.has(className))) {
      addSubtreeToInventory(node, inventory)
      return
    }

    for (const child of node.childNodes || []) {
      walk(child)
    }
  }

  walk(document)
  return inventory
}

const waitForReady = async (): Promise<void> => {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await fetch(`${origin}${targetPath}`)
      if (response.status < 500) return
    } catch {
      // The production server is still starting.
    }
    await timeout(250)
  }

  fail(`Nuxt preview did not become ready on ${origin}`)
}

const fetchHomepageHtml = async (): Promise<string> => {
  requiredFile(serverEntry)
  requiredFile(serverLauncher)

  const child = spawn(process.execPath, [serverLauncher], {
    cwd: projectRoot,
    env: {
      ...process.env,
      NODE_ENV: 'production',
      HOST: '127.0.0.1',
      PORT: String(port),
      NITRO_PORT: String(port),
      NUXT_HTML_CACHE_ENABLED: 'false',
    },
    stdio: ['ignore', 'pipe', 'pipe'],
    windowsHide: true,
  })

  const logs: string[] = []
  const capture = (chunk: Buffer): void => {
    logs.push(String(chunk))
    if (logs.length > 20) logs.shift()
  }
  child.stdout.on('data', capture)
  child.stderr.on('data', capture)

  let homepageHtml = ''

  try {
    await waitForReady()
    const response = await fetch(`${origin}${targetPath}`)
    const html = await response.text()

    if (response.status !== 200) {
      fail(`${targetPath} returned HTTP ${response.status}`)
    }
    if (!html.includes('home-hero') || !html.includes('site-header-root')) {
      fail(`${targetPath} does not look like the storefront homepage`)
    }

    homepageHtml = html
  } catch (error: unknown) {
    if (logs.length > 0) {
      console.error('[critical-css-generate] Recent server output:')
      console.error(logs.join('').trim())
    }
    fail(error instanceof Error ? error.message : String(error))
  } finally {
    child.kill('SIGTERM')
    setTimeout(() => {
      if (!child.killed) child.kill('SIGKILL')
    }, 1000).unref()
  }

  if (!homepageHtml) fail('Unable to fetch homepage HTML')
  return homepageHtml
}

const findEntryCssAsset = (): { name: string; path: string } => {
  if (!existsSync(publicNuxtDirectory)) {
    fail(`Missing ${publicNuxtDirectory}. Run npm run build first.`)
  }

  const candidates = readdirSync(publicNuxtDirectory)
    .filter(name => /^entry\.[\w-]+\.css$/i.test(name))
    .map(name => ({
      name,
      path: resolve(publicNuxtDirectory, name),
      size: statSync(resolve(publicNuxtDirectory, name)).size,
    }))
    .sort((a, b) => b.size - a.size)

  if (candidates.length === 0) {
    fail('Unable to find .output/public/_nuxt/entry.*.css')
  }

  return {
    name: candidates[0]!.name,
    path: candidates[0]!.path,
  }
}

const stripCssComments = (css: string): string => css.replace(/\/\*[\s\S]*?\*\//g, '')

const findBlockEnd = (css: string, openIndex: number): number => {
  let depth = 0
  let quote = ''
  let escaped = false

  for (let index = openIndex; index < css.length; index += 1) {
    const char = css[index]!

    if (quote) {
      if (escaped) {
        escaped = false
      } else if (char === '\\') {
        escaped = true
      } else if (char === quote) {
        quote = ''
      }
      continue
    }

    if (char === '"' || char === "'") {
      quote = char
      continue
    }
    if (char === '{') {
      depth += 1
    } else if (char === '}') {
      depth -= 1
      if (depth === 0) return index
    }
  }

  return -1
}

const findNextRuleBoundary = (css: string, start: number): { type: 'block' | 'semicolon'; index: number } | null => {
  let quote = ''
  let escaped = false

  for (let index = start; index < css.length; index += 1) {
    const char = css[index]!

    if (quote) {
      if (escaped) {
        escaped = false
      } else if (char === '\\') {
        escaped = true
      } else if (char === quote) {
        quote = ''
      }
      continue
    }

    if (char === '"' || char === "'") {
      quote = char
      continue
    }
    if (char === '{') return { type: 'block', index }
    if (char === ';') return { type: 'semicolon', index }
  }

  return null
}

const splitTopLevelRules = (css: string): CssRule[] => {
  const rules: CssRule[] = []
  let cursor = 0

  while (cursor < css.length) {
    while (cursor < css.length && /\s/.test(css[cursor]!)) cursor += 1
    if (cursor >= css.length) break

    const boundary = findNextRuleBoundary(css, cursor)
    if (!boundary) break

    if (boundary.type === 'semicolon') {
      cursor = boundary.index + 1
      continue
    }

    const end = findBlockEnd(css, boundary.index)
    if (end === -1) break

    const prelude = css.slice(cursor, boundary.index).trim()
    const content = css.slice(boundary.index + 1, end)
    if (prelude) {
      rules.push({
        prelude,
        content,
        full: `${prelude}{${content}}`,
      })
    }
    cursor = end + 1
  }

  return rules
}

const extractFontFaceRules = (css: string): string[] => {
  const rules: string[] = []

  for (const match of css.matchAll(/@font-face\s*\{/gi)) {
    const start = match.index ?? -1
    const openIndex = start >= 0 ? css.indexOf('{', start) : -1
    if (openIndex < 0) continue

    const end = findBlockEnd(css, openIndex)
    if (end < 0) continue

    rules.push(css.slice(start, end + 1))
  }

  return rules
}

const decodeCssIdentifier = (value: string): string => (
  value
    .replace(/\\([0-9a-fA-F]{1,6})\s?/g, (_, hex: string) => String.fromCodePoint(Number.parseInt(hex, 16)))
    .replace(/\\(.)/g, '$1')
)

const extractCssClasses = (selector: string): string[] => {
  const classes: string[] = []
  const pattern = /(^|[^\\])\.((?:\\[0-9a-fA-F]{1,6}\s?|\\.|[-_a-zA-Z0-9])+)/g

  for (const match of selector.matchAll(pattern)) {
    classes.push(decodeCssIdentifier(match[2] || ''))
  }

  return classes
}

const extractCssIds = (selector: string): string[] => {
  const ids: string[] = []
  const pattern = /(^|[^\\])#((?:\\[0-9a-fA-F]{1,6}\s?|\\.|[-_a-zA-Z0-9])+)/g

  for (const match of selector.matchAll(pattern)) {
    ids.push(decodeCssIdentifier(match[2] || ''))
  }

  return ids
}

const isScopedComponentSelector = (selector: string): boolean => /\[data-v-[a-f0-9]+\]/i.test(selector)

const isBaseSelector = (selector: string): boolean => {
  if (isScopedComponentSelector(selector)) return false

  const classes = extractCssClasses(selector)
  const ids = extractCssIds(selector)
  if (classes.length > 0 || ids.length > 0) return false

  return true
}

const shouldKeepStyleRule = (
  selector: string,
  inventory: CriticalInventory,
): 'base' | 'matched' | false => {
  if (isBaseSelector(selector)) return 'base'
  if (isScopedComponentSelector(selector)) return false

  const classes = extractCssClasses(selector)
  if (classes.some(className => inventory.classes.has(className))) return 'matched'

  const ids = extractCssIds(selector)
  if (ids.some(id => inventory.ids.has(id))) return 'matched'

  return false
}

const collectAnimationNames = (css: string): Set<string> => {
  const names = new Set<string>()

  for (const match of css.matchAll(/animation(?:-name)?\s*:\s*([^;}]+)/gi)) {
    const value = match[1] || ''
    for (const part of value.split(',')) {
      const tokens = part.trim().split(/\s+/)
      const candidate = tokens.find(token => (
        token &&
        !/^(?:none|linear|ease|ease-in|ease-out|ease-in-out|infinite|alternate|normal|forwards|backwards|both|running|paused)$/i.test(token) &&
        !/^-?\d*\.?\d+(?:ms|s)$/.test(token) &&
        !/^-?\d*\.?\d+$/.test(token) &&
        !/^(?:cubic-bezier|steps)\(/i.test(token)
      ))
      if (candidate) names.add(candidate.replace(/^['"]|['"]$/g, ''))
    }
  }

  return names
}

const keyframeName = (prelude: string): string => (
  prelude.replace(/^@(?:-\w+-)?keyframes\s+/i, '').trim().replace(/^['"]|['"]$/g, '')
)

const extractCriticalCss = (
  css: string,
  inventory: CriticalInventory,
): { css: string; stats: ExtractionStats } => {
  const normalizedCss = stripCssComments(css)
  const fontFaceRules = extractFontFaceRules(normalizedCss)
  const keyframes = new Map<string, string>()
  const stats: ExtractionStats = {
    fontFaceRules: fontFaceRules.length,
    baseRules: 0,
    matchedRules: 0,
    keyframeRules: 0,
    wrapperRules: 0,
  }

  const extractFromBlock = (source: string): string => {
    const parts: string[] = []

    for (const rule of splitTopLevelRules(source)) {
      const prelude = rule.prelude

      if (/^@font-face\b/i.test(prelude)) {
        continue
      }

      if (/^@(?:-\w+-)?keyframes\b/i.test(prelude)) {
        keyframes.set(keyframeName(prelude), rule.full)
        continue
      }

      if (/^@(?:media|supports|container|layer)\b/i.test(prelude)) {
        const innerCss = extractFromBlock(rule.content)
        if (innerCss) {
          stats.wrapperRules += 1
          parts.push(`${prelude}{${innerCss}}`)
        }
        continue
      }

      if (prelude.startsWith('@')) continue

      const keepReason = shouldKeepStyleRule(prelude, inventory)
      if (keepReason === 'base') {
        stats.baseRules += 1
        parts.push(rule.full)
      } else if (keepReason === 'matched') {
        stats.matchedRules += 1
        parts.push(rule.full)
      }
    }

    return parts.join('')
  }

  let extractedCss = fontFaceRules.join('') + extractFromBlock(normalizedCss)
  const animationNames = collectAnimationNames(extractedCss)
  for (const name of animationNames) {
    const keyframe = keyframes.get(name)
    if (!keyframe) continue
    stats.keyframeRules += 1
    extractedCss += keyframe
  }

  return {
    css: extractedCss,
    stats,
  }
}

const assertRequiredCriticalSelectors = (criticalCss: string): void => {
  const missing = requiredGlobalCriticalClasses.filter((className) => {
    const escapedClassFragment = className
      .replace(/\\/g, '\\\\')
      .replace(/([!"#$%&'()*+,./:;<=>?@[\\\]^`{|}~])/g, '\\$1')

    return !criticalCss.includes(`.${escapedClassFragment}`)
  })

  if (!criticalCss.includes(':root{')) missing.push(':root')

  if (missing.length > 0) {
    fail(`Critical CSS is missing required selectors: ${missing.join(', ')}`)
  }
}

const html = await fetchHomepageHtml()
const entryCssAsset = findEntryCssAsset()
const sourceCss = readFileSync(entryCssAsset.path, 'utf8')
const inventory = collectCriticalInventory(html)
const extracted = extractCriticalCss(sourceCss, inventory)
const criticalCssGzipBytes = gzipSync(extracted.css).byteLength

if (!extracted.css.trim()) {
  fail('Critical CSS extraction produced an empty stylesheet')
}

assertRequiredCriticalSelectors(extracted.css)

if (criticalCssGzipBytes > maxCriticalCssGzipBytes) {
  fail(
    `Critical CSS is ${(criticalCssGzipBytes / 1024).toFixed(2)} KiB gzip; `
    + `budget is ${(maxCriticalCssGzipBytes / 1024).toFixed(0)} KiB gzip`,
  )
}

mkdirSync(outputDirectory, { recursive: true })
writeFileSync(outputCssPath, extracted.css)
writeFileSync(outputManifestPath, `${JSON.stringify({
  source: `/_nuxt/${basename(entryCssAsset.name)}`,
  output: 'home-entry.css',
  targetPath,
  generatedAt: new Date().toISOString(),
  bytes: Buffer.byteLength(extracted.css, 'utf8'),
  gzipBytes: criticalCssGzipBytes,
  inventory: {
    classes: inventory.classes.size,
    ids: inventory.ids.size,
    tags: inventory.tags.size,
  },
  rules: extracted.stats,
}, null, 2)}\n`)

console.log(
  `[critical-css-generate] OK: ${basename(entryCssAsset.path)} -> `
  + `${(criticalCssGzipBytes / 1024).toFixed(2)} KiB gzip `
  + `(${Buffer.byteLength(extracted.css, 'utf8')} bytes)`,
)

import fs from 'node:fs'
import path from 'node:path'
import { createHash } from 'node:crypto'
import { fileURLToPath } from 'node:url'
import { create, type Font } from 'fontkitten'

const projectRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const maximumStorefrontCjkFontShardCodePoints = 512
const maximumStorefrontCjkFontShardBytes = 128 * 1024

export interface StorefrontCjkCharacterShard {
  fontFamily: string
  filename: string
  script: string
  role: string
  unicodeRange: string
  bytes: number
  sha256: string
  codePointCount: number
}

interface StorefrontCjkCharacterShardManifest {
  schemaVersion: number
  sourceFontPath: string
  sourceSha256: string
  sourceBytes: number
  sourceCodePointCount: number
  latinExcludedCodePointCount: number
  coveredCodePointCount: number
  maximumFontBytesPerShard: number
  generatorDependencies: Record<string, string>
  shards: StorefrontCjkCharacterShard[]
}

export const storefrontCjkCharacterShardManifest = JSON.parse(fs.readFileSync(
  path.join(projectRoot, 'config/storefront/storefront-cjk-and-extended-font-character-shards.manifest.json'),
  'utf8',
)) as StorefrontCjkCharacterShardManifest

export const expectedStorefrontFontFaceFiles = [
  { fontFamily: 'MapleUILatin', filename: 'MapleUI-Latin.00af3fec5b34.woff2', script: 'Latin', sourceFace: 'Maple UI', unicodeRange: '', role: 'Maple UI 拉丁核心分片' },
  ...storefrontCjkCharacterShardManifest.shards.map(shard => ({
    fontFamily: shard.fontFamily, filename: shard.filename, script: shard.script,
    sourceFace: 'Maple UI', unicodeRange: shard.unicodeRange, role: shard.role,
  })),
  { fontFamily: 'MapleUICoverageNotoSansDevanagari', filename: 'MapleUI-Coverage-NotoSans-Devanagari.3b3cae4d2600.woff2', script: 'Devanagari', sourceFace: 'Noto Sans Devanagari', unicodeRange: 'U+0900-097F, U+1CD0-1CF9, U+200C-200D, U+20A8, U+20B9, U+20F0, U+25CC, U+A830-A839, U+A8E0-A8FF, U+11B00-11B09', role: 'Maple UI 字体系统的 Devanagari 覆盖分片' },
  { fontFamily: 'MapleUICoverageNotoSansLatinAccents', filename: 'MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2', script: 'Latin accents', sourceFace: 'Noto Sans', unicodeRange: 'U+00C0-00C1, U+00C4-00C5, U+00C8-00C9, U+00CB, U+00CC-00CD, U+00CF, U+00D2-00D3, U+00D6, U+00D9-00DA, U+00DC, U+0116, U+0130, U+016E, U+01D7, U+01DB, U+01FA, U+0226, U+022E, U+0300-0301, U+0307-0308, U+030A, U+1E2E, U+FE00', role: 'Maple UI 字体系统的扩展拉丁覆盖分片' },
  { fontFamily: 'MapleUICoverageNotoSansArabic', filename: 'MapleUI-Coverage-NotoSans-Arabic.ce85091f0209.woff2', script: 'Arabic', sourceFace: 'Noto Sans Arabic', unicodeRange: 'U+0600-06FF, U+0750-077F, U+0870-088E, U+0890-0891, U+0897-08E1, U+08E3-08FF, U+200C-200E, U+2010-2011, U+204F, U+2E41, U+FB50-FDFF, U+FE70-FE74, U+FE76-FEFC, U+102E0-102FB, U+10E60-10E7E, U+10EC2-10EC4, U+10EFC-10EFF, U+1EE00-1EE03, U+1EE05-1EE1F, U+1EE21-1EE22, U+1EE24, U+1EE27, U+1EE29-1EE32, U+1EE34-1EE37, U+1EE39, U+1EE3B, U+1EE42, U+1EE47, U+1EE49, U+1EE4B, U+1EE4D-1EE4F, U+1EE51-1EE52, U+1EE54, U+1EE57, U+1EE59, U+1EE5B, U+1EE5D, U+1EE5F, U+1EE61-1EE62, U+1EE64, U+1EE67-1EE6A, U+1EE6C-1EE72, U+1EE74-1EE77, U+1EE79-1EE7C, U+1EE7E, U+1EE80-1EE89, U+1EE8B-1EE9B, U+1EEA1-1EEA3, U+1EEA5-1EEA9, U+1EEAB-1EEBB, U+1EEF0-1EEF1', role: 'Maple UI 字体系统的 Arabic 覆盖分片' },
  { fontFamily: 'MapleUICoverageNotoSansThai', filename: 'MapleUI-Coverage-NotoSans-Thai.1f5a173641bb.woff2', script: 'Thai', sourceFace: 'Noto Sans Thai', unicodeRange: 'U+02D7, U+0303, U+0331, U+0E01-0E5B, U+200C-200D, U+25CC', role: 'Maple UI 字体系统的 Thai 覆盖分片' },
]

export const originalStorefrontCjkFontSourcePath = path.join(projectRoot, storefrontCjkCharacterShardManifest.sourceFontPath)

const primaryStorefrontInterfaceMessageNamespaces = new Set([
  'header', 'common', 'dockMenu', 'cookieConsent', 'home', 'footer',
  'nav', 'menu', 'shopCategoryMenu', 'productCategoryNavigationCards',
  'storefrontDataNotice', 'faq',
])

function collectStorefrontMessageCodePoints(value: unknown, codePoints: Set<number>): void {
  if (typeof value === 'string') {
    for (const character of value) {
      const codePoint = character.codePointAt(0)
      if (codePoint !== undefined) codePoints.add(codePoint)
    }
    return
  }

  if (Array.isArray(value)) {
    for (const entry of value) collectStorefrontMessageCodePoints(entry, codePoints)
    return
  }

  if (value && typeof value === 'object') {
    for (const entry of Object.values(value)) collectStorefrontMessageCodePoints(entry, codePoints)
  }
}

function collectPrimaryStorefrontInterfaceMessageCodePoints(): Set<number> {
  const messagesDirectory = path.join(projectRoot, 'app/i18n/messages')
  const codePoints = new Set<number>()

  for (const localeEntry of fs.readdirSync(messagesDirectory, { withFileTypes: true })) {
    if (!localeEntry.isDirectory()) continue
    const localeDirectory = path.join(messagesDirectory, localeEntry.name)

    for (const messageEntry of fs.readdirSync(localeDirectory, { withFileTypes: true })) {
      if (!messageEntry.isFile() || !messageEntry.name.endsWith('.json')) continue
      const namespace = path.basename(messageEntry.name, '.json')
      if (!primaryStorefrontInterfaceMessageNamespaces.has(namespace)) continue

      const messages = JSON.parse(fs.readFileSync(path.join(localeDirectory, messageEntry.name), 'utf8')) as unknown
      collectStorefrontMessageCodePoints(messages, codePoints)
    }
  }

  return codePoints
}

export function findPrimaryStorefrontInterfaceFontShardDrift(
  primaryMessageCodePoints: Set<number>,
  originalFontCodePoints: Set<number>,
  latinFontCodePoints: Set<number>,
  commonShardCodePoints: Set<number>,
): number[] {
  return [...primaryMessageCodePoints]
    .filter(codePoint => originalFontCodePoints.has(codePoint)
      && !latinFontCodePoints.has(codePoint)
      && !commonShardCodePoints.has(codePoint))
    .sort((first, second) => first - second)
}

export function hasUnsupportedStorefrontFontUnicodeVariationSequences(
  cmapTables: Array<{ platformID: number; encodingID: number; table: { version: number } }>,
): boolean {
  return cmapTables.some(table => table.platformID === 0 && table.encodingID === 5 && table.table.version === 14)
}

export function parseStorefrontFontUnicodeRangeCodePoints(unicodeRange: string): Set<number> {
  const codePoints = new Set<number>()
  for (const declaration of unicodeRange.split(',')) {
    const match = declaration.trim().match(/^U\+([A-F\d?]{1,6})(?:-([A-F\d]{1,6}))?$/i)
    if (!match) throw new Error(`Invalid font unicode-range: ${declaration}`)
    const start = Number.parseInt(match[1].replace(/\?/g, '0'), 16)
    const end = Number.parseInt(match[2] || match[1].replace(/\?/g, 'F'), 16)
    if (start > end || end > 0x10FFFF) throw new Error(`Invalid font Unicode interval: ${declaration}`)
    for (let codePoint = start; codePoint <= end; codePoint += 1) codePoints.add(codePoint)
  }
  return codePoints
}

function readStorefrontFontForGlyphComparison(fontBytes: Buffer): Font {
  const font = create(fontBytes)
  if (font.isCollection) throw new Error('Storefront font shards must be individual font faces.')
  return font
}

export function verifyStorefrontFontCharacterShardCoverageAndOriginalGlyphParity(): string[] {
  const violations: string[] = []
  const manifest = storefrontCjkCharacterShardManifest
  const sourceBytes = fs.readFileSync(originalStorefrontCjkFontSourcePath)
  if (manifest.schemaVersion !== 2 || manifest.shards.length < 2) violations.push('The CJK character shard inventory is invalid.')
  if (manifest.maximumFontBytesPerShard !== maximumStorefrontCjkFontShardBytes) {
    violations.push(`Font shard byte budget must remain ${maximumStorefrontCjkFontShardBytes} bytes.`)
  }
  const generatorRequirementsPath = path.join(projectRoot, 'scripts/requirements-storefront-font-shard-generation.txt')
  const pinnedGeneratorDependencies = fs.readFileSync(generatorRequirementsPath, 'utf8')
    .split(/\r?\n/)
    .map(line => line.trim())
    .filter(line => line && !line.startsWith('#'))
    .map((line) => {
      const match = line.match(/^([A-Za-z0-9_-]+)==([A-Za-z0-9_.+-]+)$/)
      if (!match) throw new Error(`Invalid pinned font shard generator requirement: ${line}`)
      return [match[1].toLowerCase(), match[2]] as const
    })
    .sort(([first], [second]) => first.localeCompare(second))
  const recordedGeneratorDependencies = Object.entries(manifest.generatorDependencies)
    .sort(([first], [second]) => first.localeCompare(second))
  if (JSON.stringify(recordedGeneratorDependencies) !== JSON.stringify(pinnedGeneratorDependencies)) {
    violations.push('Font shard generator dependency versions differ from the pinned requirements file.')
  }
  if (sourceBytes.length !== manifest.sourceBytes) violations.push('Original Maple UI source byte count differs from the shard inventory.')
  if (createHash('sha256').update(sourceBytes).digest('hex') !== manifest.sourceSha256) violations.push('Original Maple UI source fingerprint differs from the shard inventory.')
  const originalFont = readStorefrontFontForGlyphComparison(sourceBytes)
  const originalCmapTables = (originalFont as unknown as {
    cmap: { tables: Array<{ platformID: number; encodingID: number; table: { version: number } }> }
  }).cmap.tables
  if (hasUnsupportedStorefrontFontUnicodeVariationSequences(originalCmapTables)) {
    violations.push('Original Maple UI source contains cmap format 14 Unicode variation sequences; extend the shard generator and parity checks before regenerating.')
  }
  const latinFont = readStorefrontFontForGlyphComparison(fs.readFileSync(path.join(projectRoot, 'public/fonts/MapleUI-Latin.00af3fec5b34.woff2')))
  const sourceCodePoints = new Set(originalFont.characterSet.filter(codePoint => codePoint !== 0xFFFF))
  const latinCodePoints = new Set(latinFont.characterSet.filter(codePoint => codePoint !== 0xFFFF))
  const expectedCodePoints = new Set([...sourceCodePoints].filter(codePoint => !latinCodePoints.has(codePoint)))
  if (sourceCodePoints.size !== manifest.sourceCodePointCount) violations.push('Original Maple UI cmap character count differs from the shard inventory.')
  if ([...sourceCodePoints].filter(codePoint => latinCodePoints.has(codePoint)).length !== manifest.latinExcludedCodePointCount) violations.push('Latin exclusion count differs from the shard inventory.')
  const coveredCodePoints = new Set<number>()
  const matchingMetrics = ['unitsPerEm', 'ascent', 'descent', 'lineGap', 'capHeight', 'xHeight'] as const
  for (const shard of manifest.shards) {
    const fontBytes = fs.readFileSync(path.join(projectRoot, 'public/fonts', shard.filename))
    const digest = createHash('sha256').update(fontBytes).digest('hex')
    if (digest !== shard.sha256 || !shard.filename.endsWith(`.${digest.slice(0, 12)}.woff2`)) violations.push(`Font shard fingerprint mismatch: ${shard.filename}`)
    if (fontBytes.length !== shard.bytes || fontBytes.length > manifest.maximumFontBytesPerShard) violations.push(`Font shard exceeds its recorded size or byte budget: ${shard.filename}`)
    if (shard.codePointCount > maximumStorefrontCjkFontShardCodePoints) {
      violations.push(`Font shard exceeds the ${maximumStorefrontCjkFontShardCodePoints}-character request-granularity budget: ${shard.filename}`)
    }
    const shardFont = readStorefrontFontForGlyphComparison(fontBytes)
    for (const metric of matchingMetrics) {
      if (shardFont[metric] !== originalFont[metric] || shardFont[metric] !== latinFont[metric]) violations.push(`${shard.filename} changed ${metric}.`)
    }
    const shardCodePoints = parseStorefrontFontUnicodeRangeCodePoints(shard.unicodeRange)
    if (shardCodePoints.size !== shard.codePointCount) violations.push(`Font shard character count mismatch: ${shard.filename}`)
    const actualShardCodePoints = new Set(shardFont.characterSet.filter(codePoint => codePoint !== 0xFFFF))
    if (actualShardCodePoints.size !== shardCodePoints.size || [...actualShardCodePoints].some(codePoint => !shardCodePoints.has(codePoint))) {
      violations.push(`Font shard cmap differs from its declared unicode-range: ${shard.filename}`)
    }
    for (const codePoint of shardCodePoints) {
      const label = `U+${codePoint.toString(16).toUpperCase()}`
      if (coveredCodePoints.has(codePoint)) violations.push(`Overlapping CJK character shards: ${label}`)
      coveredCodePoints.add(codePoint)
      if (!expectedCodePoints.has(codePoint) || !shardFont.hasGlyphForCodePoint(codePoint)) {
        violations.push(`Missing or unexpected shard glyph ${label}: ${shard.filename}`)
        continue
      }
      const originalGlyph = originalFont.glyphForCodePoint(codePoint)
      const shardGlyph = shardFont.glyphForCodePoint(codePoint)
      if (originalGlyph.advanceWidth !== shardGlyph.advanceWidth || originalGlyph.advanceHeight !== shardGlyph.advanceHeight || originalGlyph.path.toSVG() !== shardGlyph.path.toSVG()) violations.push(`Shard glyph changed relative to original Maple UI: ${label} in ${shard.filename}`)
    }
  }
  for (const codePoint of expectedCodePoints) {
    if (!coveredCodePoints.has(codePoint)) violations.push(`CJK character sharding dropped original code point U+${codePoint.toString(16).toUpperCase()}.`)
  }
  if (expectedCodePoints.size !== manifest.coveredCodePointCount) violations.push('Original CJK character count differs from the shard inventory.')

  const commonInterfaceCodePoints = new Set<number>()
  for (const shard of manifest.shards) {
    if (shard.role !== 'Common-Storefront-Interface') continue
    for (const codePoint of parseStorefrontFontUnicodeRangeCodePoints(shard.unicodeRange)) {
      commonInterfaceCodePoints.add(codePoint)
    }
  }
  for (const codePoint of findPrimaryStorefrontInterfaceFontShardDrift(
    collectPrimaryStorefrontInterfaceMessageCodePoints(),
    sourceCodePoints,
    latinCodePoints,
    commonInterfaceCodePoints,
  )) {
    violations.push(`Primary storefront interface messages use U+${codePoint.toString(16).toUpperCase()}, which is outside the common font shards; regenerate the character shards.`)
  }

  // Preserve the original Latin/core parity check even though the complete face is no longer public.
  for (const codePoint of latinFont.characterSet) {
    if (codePoint === 0xFFFF) continue
    const latinGlyph = latinFont.glyphForCodePoint(codePoint)
    const originalGlyph = originalFont.glyphForCodePoint(codePoint)
    if (latinGlyph.advanceWidth !== originalGlyph.advanceWidth || latinGlyph.advanceHeight !== originalGlyph.advanceHeight || latinGlyph.path.toSVG() !== originalGlyph.path.toSVG()) violations.push(`Latin core glyph differs from original Maple UI: U+${codePoint.toString(16).toUpperCase()}`)
  }
  return violations
}

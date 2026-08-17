import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { create } from 'fontkitten'
import {
  collectStorefrontLocaleSources,
  fontStackForStorefrontLocale,
  validateStorefrontLocaleSources,
} from './font-locale-contract.ts'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const appDir = path.join(projectDir, 'app')
const fontCssPath = path.join(appDir, 'assets', 'css', 'tailwind.css')
const stripeFontCssPath = path.join(projectDir, 'public', 'fonts', 'storefront-system.css')
const baseFontFamily = 'StorefrontSystem'
const latinFontFamily = 'StorefrontSystemLatin'
const allowedFontFamilies = new Set([
  latinFontFamily,
  baseFontFamily,
  'StorefrontSystemArabic',
  'StorefrontSystemDevanagari',
  'StorefrontSystemLatinAccents',
  'StorefrontSystemThai',
])
const expectedFontFamilyByFilename = new Map<string, string>([
  ['StorefrontSystem-Latin.00af3fec5b34.woff2', latinFontFamily],
  ['StorefrontSystem-CJK.f8ce6d72e8cb.woff2', baseFontFamily],
  ['StorefrontSystem-Arabic.ce85091f0209.woff2', 'StorefrontSystemArabic'],
  ['StorefrontSystem-Devanagari.3b3cae4d2600.woff2', 'StorefrontSystemDevanagari'],
  ['StorefrontSystem-Latin-Accents.e645edc952b6.woff2', 'StorefrontSystemLatinAccents'],
  ['StorefrontSystem-Thai.1f5a173641bb.woff2', 'StorefrontSystemThai'],
])

interface UnicodeRange {
  start: number
  end: number
}

interface StorefrontFontFace {
  fontFamily: string
  fontPath: string
  fontWeight: string
  fontStyle: string
  fontDisplay: string
  unicodeRanges: UnicodeRange[] | null
}

function collectLocaleCharacters(value: unknown, characters: Set<number>): void {
  if (typeof value === 'string') {
    for (const character of value) {
      const codePoint = character.codePointAt(0)

      if (codePoint !== undefined && !/\p{Cc}/u.test(character)) {
        characters.add(codePoint)
      }
    }

    return
  }

  if (Array.isArray(value)) {
    for (const entry of value) {
      collectLocaleCharacters(entry, characters)
    }

    return
  }

  if (value && typeof value === 'object') {
    for (const entry of Object.values(value)) {
      collectLocaleCharacters(entry, characters)
    }
  }
}

function parseUnicodeRanges(declaration: string): UnicodeRange[] | null {
  const unicodeRangeDeclaration = declaration.match(/unicode-range\s*:\s*([^;]+);?/i)
  if (!unicodeRangeDeclaration) return null

  return unicodeRangeDeclaration[1]
    .split(',')
    .map(value => value.trim().toUpperCase())
    .filter(Boolean)
    .map((value) => {
      const range = value.match(/^U\+([0-9A-F?]{1,6})(?:-([0-9A-F]{1,6}))?$/)
      if (!range) {
        throw new Error(`Unsupported unicode-range value in storefront fonts: ${value}`)
      }

      const start = Number.parseInt(range[1].replace(/\?/g, '0'), 16)
      const end = range[2]
        ? Number.parseInt(range[2], 16)
        : Number.parseInt(range[1].replace(/\?/g, 'F'), 16)

      return { start, end }
    })
}

function findStorefrontFontFaces(cssPath: string): StorefrontFontFace[] {
  const css = fs.readFileSync(cssPath, 'utf8')
  const fontFaces: StorefrontFontFace[] = []
  const fontFacePattern = /@font-face\s*\{([\s\S]*?)\}/g

  for (const match of css.matchAll(fontFacePattern)) {
    const declaration = match[1]
    const fontFamily = readCssProperty(declaration, 'font-family').replace(/^['"]|['"]$/g, '')

    if (!allowedFontFamilies.has(fontFamily)) {
      continue
    }

    for (const source of declaration.matchAll(/url\(\s*['"]?([^'")]+)['"]?\s*\)/gi)) {
      const sourcePath = source[1]
      const fontPath = sourcePath.startsWith('/')
        ? path.join(projectDir, 'public', sourcePath.slice(1))
        : path.resolve(path.dirname(cssPath), sourcePath)

      if (!fs.existsSync(fontPath)) {
        throw new Error(`Configured ${fontFamily} file does not exist: ${fontPath}`)
      }

      const expectedFontFamily = expectedFontFamilyByFilename.get(path.basename(fontPath))
      if (expectedFontFamily && expectedFontFamily !== fontFamily) {
        throw new Error(
          `${path.basename(fontPath)} must be declared as ${expectedFontFamily}, not ${fontFamily}.`,
        )
      }

      fontFaces.push({
        fontFamily,
        fontPath,
        fontWeight: readCssProperty(declaration, 'font-weight'),
        fontStyle: readCssProperty(declaration, 'font-style'),
        fontDisplay: readCssProperty(declaration, 'font-display'),
        unicodeRanges: parseUnicodeRanges(declaration),
      })
    }
  }

  if (fontFaces.length === 0) {
    throw new Error('No self-hosted storefront @font-face declarations found.')
  }

  return fontFaces
}

function readCssProperty(declaration: string, property: string): string {
  const value = declaration.match(new RegExp(`${property}\\s*:\\s*([^;]+);?`, 'i'))?.[1]?.trim()

  if (!value) {
    throw new Error(`Missing ${property} in storefront @font-face declaration.`)
  }

  return value
}

function compareFontFaces(first: StorefrontFontFace[], second: StorefrontFontFace[]): boolean {
  return first.length === second.length && first.every((fontFace, index) => {
    const other = second[index]
    return (
      fontFace.fontFamily === other.fontFamily &&
      fontFace.fontPath === other.fontPath &&
      fontFace.fontWeight === other.fontWeight &&
      fontFace.fontStyle === other.fontStyle &&
      fontFace.fontDisplay === other.fontDisplay &&
      JSON.stringify(fontFace.unicodeRanges) === JSON.stringify(other.unicodeRanges)
    )
  })
}

function supportsCodePoint(fontFace: StorefrontFontFace, characterSet: Set<number>, codePoint: number): boolean {
  if (!characterSet.has(codePoint)) return false
  return fontFace.unicodeRanges === null || fontFace.unicodeRanges.some(range => (
    codePoint >= range.start && codePoint <= range.end
  ))
}

function formatCharacters(codePoints: number[]): string {
  return codePoints
    .slice(0, 12)
    .map(codePoint => `'${String.fromCodePoint(codePoint)}' (U+${codePoint.toString(16).toUpperCase().padStart(4, '0')})`)
    .join(', ')
}

const storefrontFontFaces = findStorefrontFontFaces(fontCssPath)
const stripeFontFaces = findStorefrontFontFaces(stripeFontCssPath)

if (!compareFontFaces(storefrontFontFaces, stripeFontFaces)) {
  throw new Error(
    `Storefront and Stripe font declarations differ. Keep ${fontCssPath} and ${stripeFontCssPath} aligned.`,
  )
}

const configuredFontPaths = new Set(
  [...storefrontFontFaces, ...stripeFontFaces].map(fontFace => path.resolve(fontFace.fontPath)),
)
const productionFontDir = path.join(projectDir, 'public', 'fonts')
const productionFontExtensions = new Set(['.otf', '.ttf', '.woff', '.woff2'])
const unconfiguredProductionFonts = fs.readdirSync(productionFontDir)
  .filter(file => productionFontExtensions.has(path.extname(file).toLowerCase()))
  .map(file => path.join(productionFontDir, file))
  .filter(file => !configuredFontPaths.has(path.resolve(file)))

if (unconfiguredProductionFonts.length > 0) {
  throw new Error(
    `Unconfigured production font file(s): ${unconfiguredProductionFonts
      .map(file => path.relative(projectDir, file).replace(/\\/g, '/'))
      .join(', ')}`,
  )
}

const fontCharacterSets = new Map<string, Set<number>>()

for (const { fontPath } of storefrontFontFaces) {
  if (fontCharacterSets.has(fontPath)) continue

  const font = create(fs.readFileSync(fontPath))
  if (font.isCollection) {
    throw new Error(`Font collections are not supported by this check: ${fontPath}`)
  }

  fontCharacterSets.set(fontPath, new Set(font.characterSet))
}

const missingByLocale: Array<{ locale: string; missing: number[] }> = []
const localeSources = collectStorefrontLocaleSources(projectDir)
const localeContractViolations = validateStorefrontLocaleSources(projectDir, localeSources)

if (localeContractViolations.length > 0) {
  throw new Error(`Storefront locale contract violations:\n${localeContractViolations.map(violation => `- ${violation}`).join('\n')}`)
}

for (const [localeCode, sourcePaths] of [...localeSources.entries()].sort(([first], [second]) => first.localeCompare(second))) {
  const characters = new Set<number>()
  for (const sourcePath of sourcePaths) {
    const locale = JSON.parse(fs.readFileSync(sourcePath, 'utf8')) as unknown
    collectLocaleCharacters(locale, characters)
  }

  const allowedFamilies = new Set(fontStackForStorefrontLocale(localeCode))
  const localeFontFaces = storefrontFontFaces.filter(fontFace => allowedFamilies.has(fontFace.fontFamily))
  const missing = [...characters].filter(codePoint => !localeFontFaces.some(fontFace => (
    supportsCodePoint(fontFace, fontCharacterSets.get(fontFace.fontPath)!, codePoint)
  )))

  if (missing.length > 0) {
    missingByLocale.push({ locale: localeCode, missing })
  }
}

if (missingByLocale.length > 0) {
  console.error('Storefront font stacks do not cover all enabled locale messages:')

  for (const { locale, missing } of missingByLocale) {
    console.error(`- ${locale}: ${missing.length} missing characters; ${formatCharacters(missing)}`)
  }

  process.exit(1)
}

console.log(`Storefront locale font stacks cover all locale messages with ${storefrontFontFaces.length} self-hosted font file(s).`)

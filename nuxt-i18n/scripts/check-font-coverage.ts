import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { create } from 'fontkitten'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const appDir = path.join(projectDir, 'app')
const localeDir = path.join(appDir, 'i18n', 'locales')
const fontCssPath = path.join(appDir, 'assets', 'css', 'tailwind.css')
const stripeFontCssPath = path.join(projectDir, 'public', 'fonts', 'storefront-system.css')
const fontFamily = 'StorefrontSystem'

interface UnicodeRange {
  start: number
  end: number
}

interface StorefrontFontFace {
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
        throw new Error(`Unsupported unicode-range value in ${fontFamily}: ${value}`)
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

    if (!new RegExp(`font-family\\s*:\\s*['"]${fontFamily}['"]`, 'i').test(declaration)) {
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

      fontFaces.push({
        fontPath,
        fontWeight: readCssProperty(declaration, 'font-weight'),
        fontStyle: readCssProperty(declaration, 'font-style'),
        fontDisplay: readCssProperty(declaration, 'font-display'),
        unicodeRanges: parseUnicodeRanges(declaration),
      })
    }
  }

  if (fontFaces.length === 0) {
    throw new Error(`No self-hosted @font-face declarations found for ${fontFamily}.`)
  }

  return fontFaces
}

function readCssProperty(declaration: string, property: string): string {
  const value = declaration.match(new RegExp(`${property}\\s*:\\s*([^;]+);?`, 'i'))?.[1]?.trim()

  if (!value) {
    throw new Error(`Missing ${property} in ${fontFamily} @font-face declaration.`)
  }

  return value
}

function compareFontFaces(first: StorefrontFontFace[], second: StorefrontFontFace[]): boolean {
  return first.length === second.length && first.every((fontFace, index) => {
    const other = second[index]
    return (
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
    `Storefront and Stripe ${fontFamily} declarations differ. Keep ${fontCssPath} and ${stripeFontCssPath} aligned.`,
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

for (const localeFile of fs.readdirSync(localeDir).filter(file => file.endsWith('.json')).sort()) {
  const locale = JSON.parse(fs.readFileSync(path.join(localeDir, localeFile), 'utf8')) as unknown
  const characters = new Set<number>()
  collectLocaleCharacters(locale, characters)

  const missing = [...characters].filter(codePoint => !storefrontFontFaces.some(fontFace => (
    supportsCodePoint(fontFace, fontCharacterSets.get(fontFace.fontPath)!, codePoint)
  )))

  if (missing.length > 0) {
    missingByLocale.push({ locale: localeFile, missing })
  }
}

if (missingByLocale.length > 0) {
  console.error(`${fontFamily} does not cover all enabled locale messages:`)

  for (const { locale, missing } of missingByLocale) {
    console.error(`- ${locale}: ${missing.length} missing characters; ${formatCharacters(missing)}`)
  }

  process.exit(1)
}

console.log(`${fontFamily} covers all locale messages with ${storefrontFontFaces.length} self-hosted font file(s).`)

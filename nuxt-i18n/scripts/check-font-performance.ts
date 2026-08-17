import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { create, type Font } from 'fontkitten'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const fontDir = path.join(projectDir, 'public', 'fonts')
const fontCssPath = path.join(projectDir, 'app', 'assets', 'css', 'tailwind.css')
const stripeFontCssPath = path.join(projectDir, 'public', 'fonts', 'storefront-system.css')
const tailwindConfigPath = path.join(projectDir, 'tailwind.config.ts')
const latinFontFamily = 'StorefrontSystemLatin'
const cjkFontFamily = 'StorefrontSystem'
const fontDisplay = 'swap'
const maximumLatinFontBytes = 160 * 1024
const versionedFontFilenamePattern = /^StorefrontSystem-[A-Za-z-]+\.[a-f0-9]{12}\.woff2$/
const expectedFontFaces = [
  { fontFamily: latinFontFamily, filename: 'StorefrontSystem-Latin.00af3fec5b34.woff2', unicodeRange: false },
  { fontFamily: cjkFontFamily, filename: 'StorefrontSystem-CJK.f8ce6d72e8cb.woff2', unicodeRange: true },
  { fontFamily: 'StorefrontSystemDevanagari', filename: 'StorefrontSystem-Devanagari.3b3cae4d2600.woff2', unicodeRange: true },
  { fontFamily: 'StorefrontSystemLatinAccents', filename: 'StorefrontSystem-Latin-Accents.e645edc952b6.woff2', unicodeRange: true },
  { fontFamily: 'StorefrontSystemArabic', filename: 'StorefrontSystem-Arabic.ce85091f0209.woff2', unicodeRange: true },
  { fontFamily: 'StorefrontSystemThai', filename: 'StorefrontSystem-Thai.1f5a173641bb.woff2', unicodeRange: true },
] as const
const matchingFontMetrics = ['unitsPerEm', 'ascent', 'descent', 'lineGap', 'capHeight', 'xHeight'] as const

interface FontFaceSource {
  fontFamily: string
  sourcePath: string
  fontDisplay: string | null
  unicodeRange: string | null
}

function findFontFaceSources(css: string): FontFaceSource[] {
  const fontFaces: FontFaceSource[] = []
  const fontFacePattern = /@font-face\s*\{([\s\S]*?)\}/g

  for (const match of css.matchAll(fontFacePattern)) {
    const declaration = match[1]
    const fontFamily = declaration.match(/font-family\s*:\s*['"]?([^;'"]+)['"]?\s*;/i)?.[1]?.trim()
    const sourcePath = declaration.match(/src\s*:\s*url\(\s*['"]?([^'")]+)['"]?\s*\)/i)?.[1]
    const display = declaration.match(/font-display\s*:\s*([^;]+);?/i)?.[1]?.trim() || null
    const unicodeRange = declaration.match(/unicode-range\s*:\s*([^;]+);?/i)?.[1]?.trim() || null

    if (fontFamily && sourcePath) {
      fontFaces.push({ fontFamily, sourcePath, fontDisplay: display, unicodeRange })
    }
  }

  return fontFaces
}

function compareFontFaceContracts(first: FontFaceSource[], second: FontFaceSource[]): boolean {
  return first.length === second.length && first.every((fontFace, index) => {
    const other = second[index]

    return (
      fontFace.fontFamily === other.fontFamily &&
      path.basename(fontFace.sourcePath) === path.basename(other.sourcePath) &&
      fontFace.fontDisplay === other.fontDisplay &&
      fontFace.unicodeRange === other.unicodeRange
    )
  })
}

function loadFont(fontPath: string): Font {
  const font = create(fs.readFileSync(fontPath))

  if (font.isCollection) {
    throw new Error(`Font collections are not supported by this check: ${fontPath}`)
  }

  return font
}

const violations: string[] = []
const fontCss = fs.readFileSync(fontCssPath, 'utf8')
const tailwindConfig = fs.readFileSync(tailwindConfigPath, 'utf8')
const fontFaces = findFontFaceSources(fontCss)
const stripeFontFaces = findFontFaceSources(fs.readFileSync(stripeFontCssPath, 'utf8'))
const latinFontFace = fontFaces.find(fontFace => fontFace.fontFamily === latinFontFamily)
const cjkFontFace = fontFaces.find(fontFace => fontFace.fontFamily === cjkFontFamily)

if (!fontCss.includes("--tz-font-base: 'StorefrontSystemLatin', 'StorefrontSystem';")) {
  violations.push('The default storefront font stack must prefer StorefrontSystemLatin before StorefrontSystem.')
}

for (const fontUtility of ['sans', 'mono']) {
  if (!tailwindConfig.includes(`${fontUtility}: ['StorefrontSystemLatin', 'StorefrontSystem']`)) {
    violations.push(`Tailwind font-${fontUtility} must prefer StorefrontSystemLatin before StorefrontSystem.`)
  }
}

if (!latinFontFace) {
  violations.push('StorefrontSystemLatin must be declared as a self-hosted @font-face.')
} else {
  const latinFontPath = path.join(projectDir, 'public', latinFontFace.sourcePath.replace(/^\//, ''))

  if (!fs.existsSync(latinFontPath)) {
    violations.push(`The default Latin font file does not exist: ${latinFontFace.sourcePath}`)
  } else if (fs.statSync(latinFontPath).size > maximumLatinFontBytes) {
    violations.push(
      `The default Latin font is ${fs.statSync(latinFontPath).size} bytes; limit is ${maximumLatinFontBytes} bytes.`,
    )
  }
}

if (!cjkFontFace?.unicodeRange) {
  violations.push('StorefrontSystem must declare unicode-range so unsupported Latin symbols and emoji do not download the CJK fallback.')
}

if (!compareFontFaceContracts(fontFaces, stripeFontFaces)) {
  violations.push('Storefront and Stripe font declarations must stay identical, including font-display and unicode-range.')
}

if (fontFaces.length !== expectedFontFaces.length) {
  violations.push(`Expected ${expectedFontFaces.length} storefront @font-face declarations, found ${fontFaces.length}.`)
}

for (const expectedFontFace of expectedFontFaces) {
  const fontFace = fontFaces.find(face => face.fontFamily === expectedFontFace.fontFamily)

  if (!fontFace) {
    violations.push(`Missing required @font-face declaration: ${expectedFontFace.fontFamily}.`)
    continue
  }

  if (path.basename(fontFace.sourcePath) !== expectedFontFace.filename) {
    violations.push(
      `${expectedFontFace.fontFamily} must use ${expectedFontFace.filename}, not ${path.basename(fontFace.sourcePath)}.`,
    )
  }

  if (fontFace.fontDisplay !== fontDisplay) {
    violations.push(`${expectedFontFace.fontFamily} must use font-display: ${fontDisplay}.`)
  }

  if (Boolean(fontFace.unicodeRange) !== expectedFontFace.unicodeRange) {
    const requirement = expectedFontFace.unicodeRange ? 'declare' : 'not declare'
    violations.push(`${expectedFontFace.fontFamily} must ${requirement} unicode-range.`)
  }
}

if (latinFontFace && cjkFontFace) {
  const latinFontPath = path.join(fontDir, path.basename(latinFontFace.sourcePath))
  const cjkFontPath = path.join(fontDir, path.basename(cjkFontFace.sourcePath))

  if (fs.existsSync(latinFontPath) && fs.existsSync(cjkFontPath)) {
    const latinFont = loadFont(latinFontPath)
    const cjkFont = loadFont(cjkFontPath)

    for (const metric of matchingFontMetrics) {
      if (latinFont[metric] !== cjkFont[metric]) {
        violations.push(
          `Latin subset and complete Maple UI must share ${metric}: ${latinFont[metric]} !== ${cjkFont[metric]}.`,
        )
      }
    }

    for (const codePoint of latinFont.characterSet) {
      // Font metadata uses U+FFFF as the terminal cmap sentinel, not a glyph.
      if (codePoint === 0xFFFF) continue

      if (!cjkFont.hasGlyphForCodePoint(codePoint)) {
        violations.push(
          `Complete Maple UI is missing Latin subset glyph U+${codePoint.toString(16).toUpperCase().padStart(4, '0')}.`,
        )
        continue
      }

      const latinGlyph = latinFont.glyphForCodePoint(codePoint)
      const cjkGlyph = cjkFont.glyphForCodePoint(codePoint)

      if (
        latinGlyph.advanceWidth !== cjkGlyph.advanceWidth
        || latinGlyph.advanceHeight !== cjkGlyph.advanceHeight
        || latinGlyph.path.toSVG() !== cjkGlyph.path.toSVG()
      ) {
        violations.push(
          `Latin subset glyph U+${codePoint.toString(16).toUpperCase().padStart(4, '0')} differs from complete Maple UI.`,
        )
      }
    }
  }
}

const productionFonts = fs.readdirSync(fontDir)
  .filter(file => path.extname(file).toLowerCase() === '.woff2')

for (const fontFile of productionFonts) {
  if (!versionedFontFilenamePattern.test(fontFile)) {
    violations.push(`Production font filename must be content-addressed: public/fonts/${fontFile}`)
  }
}

if (violations.length > 0) {
  console.error('Font performance policy violations:')
  for (const violation of violations) {
    console.error(`- ${violation}`)
  }
  process.exit(1)
}

console.log(
  `Font performance and layout contract passed for ${productionFonts.length} production font file(s).`,
)

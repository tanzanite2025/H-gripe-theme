import { basename } from 'node:path'

const unversionedStorefrontFontReferencePattern = /\bStorefrontSystem-[A-Za-z-]+\.woff2\b/g
const expectedStorefrontFontFaces: ReadonlyMap<string, { filename: string; unicodeRange: boolean }> = new Map([
  ['StorefrontSystemLatin', { filename: 'StorefrontSystem-Latin.00af3fec5b34.woff2', unicodeRange: false }],
  ['StorefrontSystem', { filename: 'StorefrontSystem-CJK.f8ce6d72e8cb.woff2', unicodeRange: true }],
  ['StorefrontSystemDevanagari', { filename: 'StorefrontSystem-Devanagari.3b3cae4d2600.woff2', unicodeRange: true }],
  ['StorefrontSystemLatinAccents', { filename: 'StorefrontSystem-Latin-Accents.e645edc952b6.woff2', unicodeRange: true }],
  ['StorefrontSystemArabic', { filename: 'StorefrontSystem-Arabic.ce85091f0209.woff2', unicodeRange: true }],
  ['StorefrontSystemThai', { filename: 'StorefrontSystem-Thai.1f5a173641bb.woff2', unicodeRange: true }],
] as const)

interface BuiltFontFace {
  line: number
  fontFamily: string
  sourcePath: string
  fontDisplay: string | null
  unicodeRange: string | null
}

export interface StorefrontFontArtifactFinding {
  line: number
  violation: string
}

const lineForIndex = (source: string, index: number): number => (
  source.slice(0, index).split(/\r?\n/).length
)

const declarationValue = (declaration: string, property: string): string | null => {
  const match = declaration.match(new RegExp(`${property}\\s*:\\s*([^;}]+)`, 'i'))
  return match?.[1]?.trim().replace(/^['"]|['"]$/g, '') || null
}

const fontFaceSources = (css: string): BuiltFontFace[] => {
  const fontFaces: BuiltFontFace[] = []
  const fontFacePattern = /@font-face\s*\{([\s\S]*?)\}/gi

  for (const match of css.matchAll(fontFacePattern)) {
    const declaration = match[1]
    const fontFamily = declarationValue(declaration, 'font-family')
    const sourcePath = declaration.match(/src\s*:\s*url\(\s*['"]?([^'")]+)['"]?\s*\)/i)?.[1]?.trim()

    if (!fontFamily || !sourcePath) continue

    fontFaces.push({
      line: lineForIndex(css, match.index ?? 0),
      fontFamily,
      sourcePath,
      fontDisplay: declarationValue(declaration, 'font-display'),
      unicodeRange: declarationValue(declaration, 'unicode-range'),
    })
  }

  return fontFaces
}

export const storefrontFontArtifactViolations = (css: string): StorefrontFontArtifactFinding[] => {
  const violations: StorefrontFontArtifactFinding[] = []

  for (const match of css.matchAll(unversionedStorefrontFontReferencePattern)) {
    violations.push({
      line: lineForIndex(css, match.index ?? 0),
      violation: `storefront font file reference must be content-addressed: ${match[0]}`,
    })
  }

  for (const fontFace of fontFaceSources(css)) {
    if (!fontFace.fontFamily.startsWith('StorefrontSystem')) continue

    const expectedFontFace = expectedStorefrontFontFaces.get(fontFace.fontFamily)
    const filename = basename(fontFace.sourcePath)

    if (!expectedFontFace) {
      violations.push({
        line: fontFace.line,
        violation: `unapproved storefront @font-face family "${fontFace.fontFamily}"`,
      })
      continue
    }

    if (filename !== expectedFontFace.filename) {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must use ${expectedFontFace.filename}, not ${filename}`,
      })
    }

    if (fontFace.fontDisplay !== 'swap') {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must use font-display: swap`,
      })
    }

    if (expectedFontFace.unicodeRange && !fontFace.unicodeRange) {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must declare unicode-range in built CSS`,
      })
    }

    if (!expectedFontFace.unicodeRange && fontFace.unicodeRange) {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must not declare unicode-range in built CSS`,
      })
    }
  }

  return violations
}

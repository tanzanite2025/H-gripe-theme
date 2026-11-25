import { basename } from 'node:path'
import { expectedStorefrontFontFaceFiles, parseStorefrontFontUnicodeRangeCodePoints } from '../storefront-font-character-shard-contract.ts'

const unversionedStorefrontFontReferencePattern = /\bMapleUI-[A-Za-z0-9-]+\.woff2\b/g
const expectedStorefrontFontFacesByFilename = new Map(expectedStorefrontFontFaceFiles.map(face => [face.filename, face]))
const approvedStorefrontFontFamilies = new Set(expectedStorefrontFontFaceFiles.map(face => face.fontFamily))

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
      violation: `Maple UI font file reference must be content-addressed: ${match[0]}`,
    })
  }

  for (const fontFace of fontFaceSources(css)) {
    if (!fontFace.fontFamily.startsWith('MapleUI')) continue

    const filename = basename(fontFace.sourcePath)
    const expectedFontFace = expectedStorefrontFontFacesByFilename.get(filename)

    if (!approvedStorefrontFontFamilies.has(fontFace.fontFamily)) {
      violations.push({
        line: fontFace.line,
        violation: `unapproved Maple UI @font-face family "${fontFace.fontFamily}"`,
      })
      continue
    }

    if (!expectedFontFace || expectedFontFace.fontFamily !== fontFace.fontFamily) {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must use an approved character-shard inventory file, not ${filename}`,
      })
    }

    if (fontFace.fontDisplay !== 'block') {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must use font-display: block`,
      })
    }

    if (fontFace.fontFamily !== 'MapleUILatin' && !fontFace.unicodeRange) {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must declare unicode-range in built CSS`,
      })
    }

    if (fontFace.fontFamily === 'MapleUILatin' && fontFace.unicodeRange) {
      violations.push({
        line: fontFace.line,
        violation: `${fontFace.fontFamily} must not declare unicode-range in built CSS`,
      })
    }
    if (expectedFontFace?.unicodeRange && fontFace.unicodeRange) {
      try {
        const expectedCodePoints = parseStorefrontFontUnicodeRangeCodePoints(expectedFontFace.unicodeRange)
        const actualCodePoints = parseStorefrontFontUnicodeRangeCodePoints(fontFace.unicodeRange)
        if (actualCodePoints.size !== expectedCodePoints.size || [...actualCodePoints].some(codePoint => !expectedCodePoints.has(codePoint))) {
          violations.push({ line: fontFace.line, violation: `${filename} changed its approved character coverage in built CSS` })
        }
      } catch {
        violations.push({ line: fontFace.line, violation: `${filename} has invalid unicode-range in built CSS` })
      }
    }
  }

  return violations
}

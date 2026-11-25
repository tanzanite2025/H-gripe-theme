import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { expectedStorefrontFontFaceFiles, verifyStorefrontFontCharacterShardCoverageAndOriginalGlyphParity } from './storefront-font-character-shard-contract.ts'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const fontDir = path.join(projectDir, 'public', 'fonts')
const fontCssPath = path.join(projectDir, 'app', 'assets', 'css', 'tailwind.css')
const stripeFontCssPath = path.join(projectDir, 'public', 'fonts', 'maple-ui.css')
const tailwindConfigPath = path.join(projectDir, 'tailwind.config.ts')
const nuxtConfigPath = path.join(projectDir, 'nuxt.config.ts')
const fontPreloadPriorityPluginPath = path.join(projectDir, 'server', 'plugins', '05-font-preload-priority.server.ts')
const fontPreloadPriorityUtilityPath = path.join(projectDir, 'server', 'utils', 'fontPreloadPriority.ts')
const latinFontFamily = 'MapleUILatin'
const cjkFontFamily = 'MapleUICJK'
const fontDisplay = 'block'
const maximumLatinFontBytes = 160 * 1024
const latinFontFilename = 'MapleUI-Latin.00af3fec5b34.woff2'
const versionedFontFilenamePattern = /^MapleUI-[A-Za-z0-9-]+\.[a-f0-9]{12}\.woff2$/
const expectedFontFaces = expectedStorefrontFontFaceFiles

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

const violations: string[] = []
const fontCss = fs.readFileSync(fontCssPath, 'utf8')
const tailwindConfig = fs.readFileSync(tailwindConfigPath, 'utf8')
const nuxtConfig = fs.readFileSync(nuxtConfigPath, 'utf8')
const fontPreloadPriorityPlugin = fs.existsSync(fontPreloadPriorityPluginPath)
  ? fs.readFileSync(fontPreloadPriorityPluginPath, 'utf8')
  : ''
const fontPreloadPriorityUtility = fs.existsSync(fontPreloadPriorityUtilityPath)
  ? fs.readFileSync(fontPreloadPriorityUtilityPath, 'utf8')
  : ''
const fontPreloadPrioritySource = `${fontPreloadPriorityPlugin}\n${fontPreloadPriorityUtility}`
const fontFaces = findFontFaceSources(fontCss)
const stripeFontFaces = findFontFaceSources(fs.readFileSync(stripeFontCssPath, 'utf8'))
const latinFontFace = fontFaces.find(fontFace => fontFace.fontFamily === latinFontFamily)
const cjkFontFace = fontFaces.find(fontFace => fontFace.fontFamily === cjkFontFamily)

if (!fontCss.includes("--tz-font-base: 'MapleUILatin', 'MapleUICJK';")) {
  violations.push('The default storefront font stack must prefer MapleUILatin before MapleUICJK.')
}

for (const fontUtility of ['sans', 'mono']) {
  if (!tailwindConfig.includes(`${fontUtility}: ['MapleUILatin', 'MapleUICJK']`)) {
    violations.push(`Tailwind font-${fontUtility} must prefer MapleUILatin before MapleUICJK.`)
  }
}

if (!latinFontFace) {
  violations.push('MapleUILatin must be declared as a self-hosted @font-face.')
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

const hasStaticLatinPreload = [
  "key: 'storefront-font-preload-latin'",
  "rel: 'preload'",
  `href: '/fonts/${latinFontFilename}'`,
  "as: 'font'",
  "type: 'font/woff2'",
  "crossorigin: 'anonymous'",
].every(snippet => nuxtConfig.includes(snippet))

if (!hasStaticLatinPreload) {
  violations.push('Nuxt app.head must statically declare the MapleUILatin first-paint preload.')
}

const hasLatinPreloadPriorityPlugin = [
  `LATIN_FONT_PRELOAD_HREF = '/fonts/${latinFontFilename}'`,
  'LATIN_FONT_PRELOAD_PATTERN',
  'prioritizeLatinFontPreload',
  "'render:response'",
].every(snippet => fontPreloadPrioritySource.includes(snippet))

if (!hasLatinPreloadPriorityPlugin) {
  violations.push('Nitro must prioritize the MapleUILatin preload at the start of SSR <head> output.')
}

if (!cjkFontFace?.unicodeRange) {
  violations.push('MapleUICJK must declare unicode-range so Latin first paint does not download the CJK/extended built-in shard.')
}

if (!compareFontFaceContracts(fontFaces, stripeFontFaces)) {
  violations.push('Storefront and Stripe font declarations must stay identical, including font-display and unicode-range.')
}

if (fontFaces.length !== expectedFontFaces.length) {
  violations.push(`Expected ${expectedFontFaces.length} storefront @font-face declarations, found ${fontFaces.length}.`)
}

for (const expectedFontFace of expectedFontFaces) {
  const fontFace = fontFaces.find(face => path.basename(face.sourcePath) === expectedFontFace.filename)

  if (!fontFace) {
    violations.push(`Missing required @font-face declaration: ${expectedFontFace.fontFamily}.`)
    continue
  }

  if (path.basename(fontFace.sourcePath) !== expectedFontFace.filename) {
    violations.push(
      `${expectedFontFace.fontFamily} must use ${expectedFontFace.filename}, not ${path.basename(fontFace.sourcePath)}.`,
    )
  }
  if (fontFace.fontFamily !== expectedFontFace.fontFamily) {
    violations.push(`${expectedFontFace.filename} must belong to ${expectedFontFace.fontFamily}.`)
  }

  if (fontFace.fontDisplay !== fontDisplay) {
    violations.push(`${expectedFontFace.fontFamily} must use font-display: ${fontDisplay}.`)
  }

  if (Boolean(fontFace.unicodeRange) !== Boolean(expectedFontFace.unicodeRange)) {
    const requirement = expectedFontFace.unicodeRange ? 'declare' : 'not declare'
    violations.push(`${expectedFontFace.fontFamily} must ${requirement} unicode-range.`)
  }
  if ((fontFace.unicodeRange || '') !== expectedFontFace.unicodeRange) {
    violations.push(`${expectedFontFace.filename} must retain its approved character coverage.`)
  }
}

violations.push(...verifyStorefrontFontCharacterShardCoverageAndOriginalGlyphParity())

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

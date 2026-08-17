import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { create, type Font } from 'fontkitten'
import {
  collectStorefrontLocaleSources,
  fontStackForStorefrontLocale,
  validateStorefrontLocaleSources,
} from './font-locale-contract.ts'
import { collectFontPolicyViolations } from './font-policy-utils.ts'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const appDir = path.join(projectDir, 'app')
const fontDir = path.join(projectDir, 'public', 'fonts')
const fontCssPath = path.join(appDir, 'assets', 'css', 'tailwind.css')
const stripeFontCssPath = path.join(fontDir, 'storefront-system.css')
const tailwindConfigPath = path.join(projectDir, 'tailwind.config.ts')
const manifestPath = path.join(projectDir, 'public', '_internal', 'font-preflight.json')

const latinFontFamily = 'StorefrontSystemLatin'
const cjkFontFamily = 'StorefrontSystem'
const fontDisplay = 'swap'
const maximumLatinFontBytes = 160 * 1024
const versionedFontFilenamePattern = /^StorefrontSystem-[A-Za-z-]+\.[a-f0-9]{12}\.woff2$/
const productionFontExtensions = new Set(['.otf', '.ttf', '.woff', '.woff2'])
const baseFontStack = [latinFontFamily, cjkFontFamily]
const expectedFontFaces = [
  { fontFamily: latinFontFamily, filename: 'StorefrontSystem-Latin.00af3fec5b34.woff2', script: 'Latin', unicodeRange: '', role: '默认首屏子集' },
  { fontFamily: cjkFontFamily, filename: 'StorefrontSystem-CJK.f8ce6d72e8cb.woff2', script: 'CJK / complete Maple UI', unicodeRange: 'U+0250-02FF, U+0370-052F, U+0530-058F, U+1F00-1FFF, U+2070-209F, U+2150-218F, U+2C00-2DFF, U+2E80-2EFF, U+3000-303F, U+3040-30FF, U+3100-31FF, U+3200-33FF, U+3400-4DBF, U+4E00-9FFF, U+AC00-D7AF, U+F900-FAFF, U+FE30-FE4F, U+FF00-FFEF, U+20000-2FA1F', role: 'CJK 与完整字形覆盖' },
  { fontFamily: 'StorefrontSystemDevanagari', filename: 'StorefrontSystem-Devanagari.3b3cae4d2600.woff2', script: 'Devanagari', unicodeRange: 'U+0900-097F, U+1CD0-1CF9, U+200C-200D, U+20A8, U+20B9, U+20F0, U+25CC, U+A830-A839, U+A8E0-A8FF, U+11B00-11B09', role: '印地语分片' },
  { fontFamily: 'StorefrontSystemLatinAccents', filename: 'StorefrontSystem-Latin-Accents.e645edc952b6.woff2', script: 'Latin accents', unicodeRange: 'U+00C0-00C1, U+00C4-00C5, U+00C9, U+00D6, U+00DC, U+0130', role: '扩展拉丁语种分片' },
  { fontFamily: 'StorefrontSystemArabic', filename: 'StorefrontSystem-Arabic.ce85091f0209.woff2', script: 'Arabic', unicodeRange: 'U+0600-06FF, U+0750-077F, U+0870-088E, U+0890-0891, U+0897-08E1, U+08E3-08FF, U+200C-200E, U+2010-2011, U+204F, U+2E41, U+FB50-FDFF, U+FE70-FE74, U+FE76-FEFC, U+102E0-102FB, U+10E60-10E7E, U+10EC2-10EC4, U+10EFC-10EFF, U+1EE00-1EE03, U+1EE05-1EE1F, U+1EE21-1EE22, U+1EE24, U+1EE27, U+1EE29-1EE32, U+1EE34-1EE37, U+1EE39, U+1EE3B, U+1EE42, U+1EE47, U+1EE49, U+1EE4B, U+1EE4D-1EE4F, U+1EE51-1EE52, U+1EE54, U+1EE57, U+1EE59, U+1EE5B, U+1EE5D, U+1EE5F, U+1EE61-1EE62, U+1EE64, U+1EE67-1EE6A, U+1EE6C-1EE72, U+1EE74-1EE77, U+1EE79-1EE7C, U+1EE7E, U+1EE80-1EE89, U+1EE8B-1EE9B, U+1EEA1-1EEA3, U+1EEA5-1EEA9, U+1EEAB-1EEBB, U+1EEF0-1EEF1', role: '阿拉伯语分片' },
  { fontFamily: 'StorefrontSystemThai', filename: 'StorefrontSystem-Thai.1f5a173641bb.woff2', script: 'Thai', unicodeRange: 'U+02D7, U+0303, U+0331, U+0E01-0E5B, U+200C-200D, U+25CC', role: '泰语分片' },
] as const
const matchingFontMetrics = ['unitsPerEm', 'ascent', 'descent', 'lineGap', 'capHeight', 'xHeight'] as const
const allowedFontFamilies = new Set<string>(expectedFontFaces.map(face => face.fontFamily))
type CheckStatus = 'pass' | 'block'

interface UnicodeRange {
  start: number
  end: number
}

interface FontFaceSource {
  fontFamily: string
  sourcePath: string
  sourceCount: number
  sourceFormats: string[]
  hasLocalSource: boolean
  fontDisplay: string | null
  unicodeRange: string | null
}

interface StorefrontFontFace extends FontFaceSource {
  fontPath: string
  fontWeight: string
  fontStyle: string
  unicodeRanges: UnicodeRange[] | null
}

interface FontPreflightCheck {
  key: string
  label: string
  status: CheckStatus
  message: string
  details: string[]
}

const readCssProperty = (declaration: string, property: string): string => {
  const value = declaration.match(new RegExp(`${property}\\s*:\\s*([^;]+);?`, 'i'))?.[1]?.trim()
  if (!value) throw new Error(`Missing ${property} in storefront @font-face declaration.`)
  return value
}

const parseUnicodeRanges = (declaration: string): UnicodeRange[] | null => {
  const unicodeRangeDeclaration = declaration.match(/unicode-range\s*:\s*([^;]+);?/i)
  if (!unicodeRangeDeclaration) return null

  return unicodeRangeDeclaration[1]
    .split(',')
    .map(value => value.trim().toUpperCase())
    .filter(Boolean)
    .map((value) => {
      const range = value.match(/^U\+([0-9A-F?]{1,6})(?:-([0-9A-F]{1,6}))?$/)
      if (!range) throw new Error(`Unsupported unicode-range value in storefront fonts: ${value}`)

      const parsedRange = {
        start: Number.parseInt(range[1].replace(/\?/g, '0'), 16),
        end: range[2] ? Number.parseInt(range[2], 16) : Number.parseInt(range[1].replace(/\?/g, 'F'), 16),
      }
      if (parsedRange.start > parsedRange.end) throw new Error(`Invalid unicode-range in storefront fonts: ${value}`)
      return parsedRange
    })
}

const findFontSources = (declaration: string): Array<{ sourcePath: string; format: string }> => (
  [...declaration.matchAll(/url\(\s*['"]?([^'")]+)['"]?\s*\)(?:\s*format\(\s*['"]?([^'")]+)['"]?\s*\))?/gi)]
    .map(match => ({
      sourcePath: match[1],
      format: (match[2] || '').trim().toLowerCase(),
    }))
)

const findFontFaceSources = (css: string): FontFaceSource[] => {
  const fontFaces: FontFaceSource[] = []
  const fontFacePattern = /@font-face\s*\{([\s\S]*?)\}/g

  for (const match of css.matchAll(fontFacePattern)) {
    const declaration = match[1]
    const fontFamily = declaration.match(/font-family\s*:\s*['"]?([^;'"]+)['"]?\s*;/i)?.[1]?.trim()
    const sources = findFontSources(declaration)
    const sourcePath = sources[0]?.sourcePath
    const declaredDisplay = declaration.match(/font-display\s*:\s*([^;]+);?/i)?.[1]?.trim() || null
    const unicodeRange = declaration.match(/unicode-range\s*:\s*([^;]+);?/i)?.[1]?.trim() || null

    if (fontFamily && sourcePath) {
      fontFaces.push({
        fontFamily,
        sourcePath,
        sourceCount: sources.length,
        sourceFormats: sources.map(source => source.format),
        hasLocalSource: /\blocal\s*\(/i.test(declaration),
        fontDisplay: declaredDisplay,
        unicodeRange,
      })
    }
  }

  return fontFaces
}

const findStorefrontFontFaces = (cssPath: string): StorefrontFontFace[] => {
  const css = fs.readFileSync(cssPath, 'utf8')
  const fontFaces: StorefrontFontFace[] = []
  const fontFacePattern = /@font-face\s*\{([\s\S]*?)\}/g

  for (const match of css.matchAll(fontFacePattern)) {
    const declaration = match[1]
    const fontFamily = readCssProperty(declaration, 'font-family').replace(/^['"]|['"]$/g, '')
    if (!allowedFontFamilies.has(fontFamily)) continue

    const sources = findFontSources(declaration)
    const source = sources[0]?.sourcePath
    if (!source) throw new Error(`Missing font source for ${fontFamily}.`)

    const fontPath = source.startsWith('/')
      ? path.join(projectDir, 'public', source.slice(1))
      : path.resolve(path.dirname(cssPath), source)

    fontFaces.push({
      fontFamily,
      sourcePath: source,
      sourceCount: sources.length,
      sourceFormats: sources.map(source => source.format),
      hasLocalSource: /\blocal\s*\(/i.test(declaration),
      fontDisplay: readCssProperty(declaration, 'font-display'),
      unicodeRange: declaration.match(/unicode-range\s*:\s*([^;]+);?/i)?.[1]?.trim() || null,
      fontPath,
      fontWeight: readCssProperty(declaration, 'font-weight'),
      fontStyle: readCssProperty(declaration, 'font-style'),
      unicodeRanges: parseUnicodeRanges(declaration),
    })
  }

  return fontFaces
}

const compareFontFaceContracts = (first: FontFaceSource[], second: FontFaceSource[]): boolean => (
  first.length === second.length && first.every((fontFace, index) => {
    const other = second[index]
    return fontFace.fontFamily === other.fontFamily
      && path.basename(fontFace.sourcePath) === path.basename(other.sourcePath)
      && fontFace.sourceCount === other.sourceCount
      && fontFace.sourceFormats.join(',') === other.sourceFormats.join(',')
      && fontFace.fontDisplay === other.fontDisplay
      && fontFace.unicodeRange === other.unicodeRange
  })
)

const collectLocaleCharacters = (value: unknown, characters: Set<number>): void => {
  if (typeof value === 'string') {
    for (const character of value) {
      const codePoint = character.codePointAt(0)
      if (codePoint !== undefined && !/\p{Cc}/u.test(character)) characters.add(codePoint)
    }
    return
  }

  if (Array.isArray(value)) {
    for (const entry of value) collectLocaleCharacters(entry, characters)
    return
  }

  if (value && typeof value === 'object') {
    for (const entry of Object.values(value)) collectLocaleCharacters(entry, characters)
  }
}

const supportsCodePoint = (fontFace: StorefrontFontFace, characterSet: Set<number>, codePoint: number): boolean => (
  characterSet.has(codePoint)
  && (fontFace.unicodeRanges === null || fontFace.unicodeRanges.some(range => codePoint >= range.start && codePoint <= range.end))
)

const formatCodePoint = (codePoint: number): string => `U+${codePoint.toString(16).toUpperCase().padStart(4, '0')}`

const loadFont = (fontPath: string): Font => {
  const font = create(fs.readFileSync(fontPath))
  if (font.isCollection) throw new Error(`Font collections are not supported by this check: ${fontPath}`)
  return font
}

const check = (
  key: string,
  label: string,
  details: string[],
  successMessage: string,
): FontPreflightCheck => ({
  key,
  label,
  status: details.length === 0 ? 'pass' : 'block',
  message: details.length === 0 ? successMessage : `${details.length} 项不满足上线字体基线`,
  details,
})

const fontFaceRole = (fontFamily: string): string => (
  expectedFontFaces.find(face => face.fontFamily === fontFamily)?.role || '未识别字体'
)

const main = (): void => {
  const fontCss = fs.readFileSync(fontCssPath, 'utf8')
  const stripeCss = fs.readFileSync(stripeFontCssPath, 'utf8')
  const tailwindConfig = fs.readFileSync(tailwindConfigPath, 'utf8')
  const fontFaces = findFontFaceSources(fontCss)
  const stripeFontFaces = findFontFaceSources(stripeCss)
  const storefrontFontFaces = findStorefrontFontFaces(fontCssPath)
  const policyViolations = collectFontPolicyViolations(projectDir)
  const faceViolations: string[] = []
  const splitViolations: string[] = []
  const layoutViolations: string[] = []
  const coverageViolations: string[] = []

  if (fontFaces.length !== expectedFontFaces.length) {
    faceViolations.push(`Expected ${expectedFontFaces.length} @font-face declarations, found ${fontFaces.length}.`)
  }

  if (!compareFontFaceContracts(fontFaces, stripeFontFaces)) {
    faceViolations.push('Storefront and checkout font declarations differ.')
  }

  if (!fontCss.includes("--tz-font-base: 'StorefrontSystemLatin', 'StorefrontSystem';")) {
    splitViolations.push('The default font stack must prefer StorefrontSystemLatin before StorefrontSystem.')
  }

  for (const fontUtility of ['sans', 'mono']) {
    if (!tailwindConfig.includes(`${fontUtility}: ['StorefrontSystemLatin', 'StorefrontSystem']`)) {
      splitViolations.push(`Tailwind font-${fontUtility} must prefer the Latin subset before complete Maple UI.`)
    }
  }

  const fontFaceByFamily = new Map(fontFaces.map(fontFace => [fontFace.fontFamily, fontFace]))
  for (const expectedFontFace of expectedFontFaces) {
    const face = fontFaceByFamily.get(expectedFontFace.fontFamily)
    if (!face) {
      faceViolations.push(`Missing ${expectedFontFace.fontFamily}.`)
      continue
    }

    if (path.basename(face.sourcePath) !== expectedFontFace.filename) {
      faceViolations.push(`${expectedFontFace.fontFamily} must use ${expectedFontFace.filename}.`)
    }
    if (face.hasLocalSource) {
      faceViolations.push(`${expectedFontFace.fontFamily} must not declare local() font sources.`)
    }
    if (face.sourceCount !== 1) {
      faceViolations.push(`${expectedFontFace.fontFamily} must declare exactly one font source.`)
    }
    if (face.sourceFormats.length !== 1 || face.sourceFormats[0] !== 'woff2') {
      faceViolations.push(`${expectedFontFace.fontFamily} must declare format('woff2').`)
    }
    if (path.extname(face.sourcePath).toLowerCase() !== '.woff2') {
      faceViolations.push(`${expectedFontFace.fontFamily} must point to a WOFF2 font file.`)
    }
    if (/^(?:https?:)?\/\//i.test(face.sourcePath) || /^data:/i.test(face.sourcePath)) {
      faceViolations.push(`${expectedFontFace.fontFamily} must not use an external or inline font source.`)
    }
    if (face.fontDisplay !== fontDisplay) {
      faceViolations.push(`${expectedFontFace.fontFamily} must use font-display: swap.`)
    }
    if ((face.unicodeRange || '') !== expectedFontFace.unicodeRange) {
      const requirement = expectedFontFace.unicodeRange
        ? `declare the baseline unicode-range ${expectedFontFace.unicodeRange}.`
        : 'not declare unicode-range.'
      faceViolations.push(`${expectedFontFace.fontFamily} must ${requirement}`)
    }
    if (!face.sourcePath.startsWith('/fonts/')) {
      faceViolations.push(`${expectedFontFace.fontFamily} must be served from this project's /fonts/ directory.`)
    }
  }

  const configuredFontPaths = new Set(storefrontFontFaces.map(face => path.resolve(face.fontPath)))
  const productionFontFiles = fs.readdirSync(fontDir)
    .filter(file => productionFontExtensions.has(path.extname(file).toLowerCase()))
    .sort()
  for (const fontFile of productionFontFiles) {
    const fontPath = path.join(fontDir, fontFile)
    const extension = path.extname(fontFile).toLowerCase()
    if (extension !== '.woff2') {
      faceViolations.push(`Production font file must be WOFF2 only: public/fonts/${fontFile}`)
    } else if (!versionedFontFilenamePattern.test(fontFile)) {
      faceViolations.push(`Font filename is not content-addressed: public/fonts/${fontFile}`)
    }
    if (!configuredFontPaths.has(path.resolve(fontPath))) {
      faceViolations.push(`Production font file is not declared by the storefront font contract: public/fonts/${fontFile}`)
    }
  }

  const latinFace = fontFaceByFamily.get(latinFontFamily)
  const cjkFace = fontFaceByFamily.get(cjkFontFamily)
  const latinFontPath = latinFace ? path.join(projectDir, 'public', latinFace.sourcePath.replace(/^\//, '')) : ''
  const cjkFontPath = cjkFace ? path.join(projectDir, 'public', cjkFace.sourcePath.replace(/^\//, '')) : ''

  if (!cjkFace?.unicodeRange) {
    splitViolations.push('The complete Maple UI face must have unicode-range so Latin first paint does not download it.')
  }

  if (!latinFace || !fs.existsSync(latinFontPath)) {
    splitViolations.push('The Latin first-paint subset is missing.')
  } else if (fs.statSync(latinFontPath).size > maximumLatinFontBytes) {
    splitViolations.push(`Latin subset is ${fs.statSync(latinFontPath).size} bytes, over the ${maximumLatinFontBytes} byte budget.`)
  }

  if (!cjkFace || !fs.existsSync(cjkFontPath)) {
    layoutViolations.push('The complete Maple UI face is missing.')
  } else if (latinFace && fs.existsSync(latinFontPath)) {
    const latinFont = loadFont(latinFontPath)
    const cjkFont = loadFont(cjkFontPath)

    for (const metric of matchingFontMetrics) {
      if (latinFont[metric] !== cjkFont[metric]) {
        layoutViolations.push(`Latin subset and complete Maple UI differ in ${metric}.`)
      }
    }

    for (const codePoint of latinFont.characterSet) {
      if (codePoint === 0xFFFF) continue
      if (!cjkFont.hasGlyphForCodePoint(codePoint)) {
        layoutViolations.push(`Complete Maple UI is missing ${formatCodePoint(codePoint)} from the Latin subset.`)
        continue
      }
      const latinGlyph = latinFont.glyphForCodePoint(codePoint)
      const cjkGlyph = cjkFont.glyphForCodePoint(codePoint)
      if (
        latinGlyph.advanceWidth !== cjkGlyph.advanceWidth
        || latinGlyph.advanceHeight !== cjkGlyph.advanceHeight
        || latinGlyph.path.toSVG() !== cjkGlyph.path.toSVG()
      ) {
        layoutViolations.push(`${formatCodePoint(codePoint)} differs between the Latin subset and complete Maple UI.`)
      }
    }
  }

  const fontCharacterSets = new Map<string, Set<number>>()
  for (const face of storefrontFontFaces) {
    if (!fs.existsSync(face.fontPath)) {
      coverageViolations.push(`Configured font file does not exist: ${path.relative(projectDir, face.fontPath)}.`)
      continue
    }
    if (!fontCharacterSets.has(face.fontPath)) {
      fontCharacterSets.set(face.fontPath, new Set(loadFont(face.fontPath).characterSet))
    }
  }

  const localeSources = collectStorefrontLocaleSources(projectDir)
  coverageViolations.push(...validateStorefrontLocaleSources(projectDir, localeSources))
  if (localeSources.size === 0) {
    coverageViolations.push('No locale or message JSON resources were found.')
  }

  const localeCoverage = [...localeSources.entries()]
    .sort(([first], [second]) => first.localeCompare(second))
    .map(([locale, sourcePaths]) => {
      const characters = new Set<number>()
      for (const sourcePath of sourcePaths) {
        collectLocaleCharacters(JSON.parse(fs.readFileSync(sourcePath, 'utf8')) as unknown, characters)
      }

      const fontStack = fontStackForStorefrontLocale(locale)
      const localeFaces = storefrontFontFaces.filter(face => fontStack.includes(face.fontFamily))
      const missing = [...characters].filter(codePoint => !localeFaces.some(face => (
        supportsCodePoint(face, fontCharacterSets.get(face.fontPath) || new Set(), codePoint)
      )))

      if (missing.length > 0) {
        coverageViolations.push(`${locale}: ${missing.length} missing character(s), including ${missing.slice(0, 8).map(formatCodePoint).join(', ')}.`)
      }

      return {
        locale,
        source_files: sourcePaths.length,
        checked_characters: characters.size,
        missing_characters: missing.length,
        missing_sample: missing.slice(0, 12).map(formatCodePoint),
        font_stack: fontStack,
        status: missing.length === 0 ? 'pass' : 'block' as CheckStatus,
      }
    })

  const checks = [
    check(
      'no-external-fallback',
      '禁止外部 / 系统 / 通用字体回退',
      policyViolations,
      '未发现系统、通用、CDN 或未批准字体回退。',
    ),
    check(
      'font-face-contract',
      '自托管字体合同',
      faceViolations,
      `${expectedFontFaces.length} 个受控字体分片均为内容寻址的本项目资源，并保持 checkout 声明一致。`,
    ),
    check(
      'multilingual-split',
      '多语言分片与首屏策略',
      splitViolations,
      'Latin 子集首屏加载；CJK 与其他脚本按 unicode-range 分片，完整 Maple UI 不参与 Latin 首屏下载。',
    ),
    check(
      'layout-parity',
      'Latin 子集布局一致性',
      layoutViolations,
      'Latin 子集与完整 Maple UI 的字体度量、字宽和轮廓一致，不会因字体切换造成布局漂移。',
    ),
    check(
      'subset-completeness',
      '语言资源子集完备性',
      coverageViolations,
      `${localeCoverage.length} 个项目语言的所有 locale 与 message JSON 均由其受控字体栈覆盖。`,
    ),
  ]
  const overallStatus: CheckStatus = checks.some(result => result.status === 'block') ? 'block' : 'pass'

  const manifest = {
    schema_version: 1,
    project: 'tanzanite-theme storefront',
    generated_at: new Date().toISOString(),
    overall_status: overallStatus,
    baseline: {
      id: 'storefront-self-hosted-no-fallback-v1',
      label: '自托管字体，禁止非预期回退',
      font_display: fontDisplay,
      rules: [
        '禁止 system-ui、sans-serif、serif、monospace、平台系统字体、CDN 字体和未批准字体族。',
        'StorefrontSystem 是完整 Maple UI 的自托管脚本覆盖面，不是系统 fallback。',
        'font-display 必须保持 swap；性能策略不能通过改成 block 来规避。',
      ],
    },
    checks,
    strategy: {
      status: splitViolations.length === 0 && layoutViolations.length === 0 ? 'pass' : 'block',
      label: 'Latin 首屏子集 + unicode-range 脚本分片',
      default_stack: baseFontStack,
      latin_bytes: latinFace && fs.existsSync(latinFontPath) ? fs.statSync(latinFontPath).size : 0,
      latin_budget_bytes: maximumLatinFontBytes,
      complete_maple_ui_family: cjkFontFamily,
      cjk_unicode_range: cjkFace?.unicodeRange || '',
      layout_parity_verified: layoutViolations.length === 0,
      rationale: '默认 Latin 使用更小的同度量子集；CJK 和非 Latin 语言仅在实际字符命中时加载自托管脚本字体。',
    },
    faces: expectedFontFaces.map(expected => {
      const face = fontFaceByFamily.get(expected.fontFamily)
      const fontPath = face ? path.join(projectDir, 'public', face.sourcePath.replace(/^\//, '')) : ''
      return {
        family: expected.fontFamily,
        role: fontFaceRole(expected.fontFamily),
        script: expected.script,
        filename: face ? path.basename(face.sourcePath) : expected.filename,
        bytes: fontPath && fs.existsSync(fontPath) ? fs.statSync(fontPath).size : 0,
        font_display: face?.fontDisplay || '',
        unicode_range: face?.unicodeRange || '',
        self_hosted: Boolean(face?.sourcePath.startsWith('/fonts/')),
      }
    }),
    coverage: {
      locale_count: localeCoverage.length,
      source_file_count: [...localeSources.values()].reduce((total, files) => total + files.length, 0),
      checked_characters: localeCoverage.reduce((total, locale) => total + locale.checked_characters, 0),
      missing_characters: localeCoverage.reduce((total, locale) => total + locale.missing_characters, 0),
      locales: localeCoverage,
    },
  }

  fs.mkdirSync(path.dirname(manifestPath), { recursive: true })
  fs.writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`, 'utf8')

  if (overallStatus === 'block') {
    console.error(`Font preflight manifest generated with blocking violations: ${manifestPath}`)
    for (const result of checks.filter(result => result.status === 'block')) {
      console.error(`- ${result.label}: ${result.details.join(' ')}`)
    }
    process.exit(1)
  }

  console.log(`Font preflight manifest generated for ${localeCoverage.length} locales: ${path.relative(projectDir, manifestPath).replace(/\\/g, '/')}`)
}

main()

import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

interface Violation {
  file: string
  line?: number
  message: string
}

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const adminRoot = path.resolve(scriptDir, '..')
const srcRoot = path.join(adminRoot, 'src')
const adminCssPath = path.join(srcRoot, 'styles', 'admin.css')
const dashboardPresentationPath = path.join(srcRoot, 'lib', 'dashboardPresentation.ts')
const fontRoot = path.join(srcRoot, 'assets', 'fonts')
const expectedFontPath = path.join(fontRoot, 'maple-ui', 'MapleUI-CJK.f8ce6d72e8cb.woff2')
const expectedFontHash = 'f8ce6d72e8cbda22971300b1367fe49ba98f5f798e0a1a297f4dea90c7fb8be9'
const distRoot = path.join(adminRoot, 'dist')
const sourceExtensions = new Set(['.css', '.js', '.ts', '.vue'])
const inspectableOutputExtensions = new Set(['.css', '.html', '.js', '.mjs', '.json'])
const outputFontExtensions = new Set(['.otf', '.ttf', '.woff', '.woff2'])
const forbiddenFontReference = /(?:\b(?:AdminSystem|system-ui|ui-sans-serif|sans-serif|serif|monospace|ui-monospace|-apple-system|BlinkMacSystemFont|Segoe UI|Arial|Helvetica|Noto Sans|SFMono-Regular|Menlo|Monaco|Consolas|Liberation Mono|Courier New)\b|fonts\.(?:googleapis|gstatic)\.com|use\.typekit\.net|data:font\/)/i
const forbiddenOutputReference = /(?:\b(?:AdminSystem|StorefrontSystem|storefront-system|nuxt-fonts(?:-global)?)\b|@nuxt\/fonts|\/_fonts\/|fonts\.(?:googleapis|gstatic)\.com|use\.typekit\.net|data:font\/)/i
const forbiddenStylesheetOutputReference = /(?:\b(?:AdminSystem|StorefrontSystem|storefront-system|nuxt-fonts(?:-global)?)\b|@nuxt\/fonts|\/_fonts\/|fonts\.(?:googleapis|gstatic)\.com|use\.typekit\.net|data:font\/|local\s*\()/i
const fontFamilyAttributePattern = /(?<![\w-])font-family\s*=\s*(['"])(.*?)\1/gi
const fontFamilyDeclarationPattern = /\bfont-family\s*:\s*([^;}\]]+)/gi
const fontFamilyAssignmentPattern = /\bfontFamily\s*:\s*(['"`])([\s\S]*?)\1/gi
const tailwindFontSerifUtilityPattern = /(?:^|[\s"'`])((?:[\w-]+:)*font-serif)(?=$|[\s"'`])/gi
const tailwindArbitraryFontUtilityPattern = /(?:^|[\s"'`])((?:[\w-]+:)*font-\[([^\]\r\n]+)\])(?=$|[\s"'`])/gi
const systemFallbackPattern = /\b(?:-apple-system|arial|blinkmacsystemfont|calibri|cambria|consolas|courier(?: new)?|cursive|fangsong|fantasy|georgia|helvetica(?: neue)?|liberation(?: mono| sans| serif)?|math|menlo|monaco|monospace|sans-serif|segoe ui|serif|sf mono|sf pro|sfmono-regular|system-ui|tahoma|times(?: new roman)?|trebuchet ms|ui-monospace|ui-rounded|ui-sans-serif|ui-serif|verdana)\b/i
const cssFontFamilyKeywords = new Set(['inherit', 'initial', 'unset', 'revert', 'revert-layer'])
const cssFontWeightKeywords = new Set(['bold', 'bolder', 'lighter', 'normal'])
const approvedFontFamilies = new Set(['MapleUICJK'])
const approvedFontVariables = new Set(['--tz-font-admin-ui', '--font-sans', '--font-mono', '--font-heading'])

const violations: Violation[] = []
const checkDist = process.argv.includes('--dist')

function relativePath(filePath: string): string {
  return path.relative(adminRoot, filePath).replace(/\\/g, '/')
}

function collectSourceFiles(directory: string): string[] {
  const files: string[] = []
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    if (entry.name === 'dist' || entry.name === 'node_modules') continue
    const entryPath = path.join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...collectSourceFiles(entryPath))
      continue
    }
    if (entry.isFile() && sourceExtensions.has(path.extname(entry.name))) {
      files.push(entryPath)
    }
  }
  return files
}

function collectAllFiles(directory: string): string[] {
  const files: string[] = []
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    if (entry.name === 'dist' || entry.name === 'node_modules') continue
    const entryPath = path.join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...collectAllFiles(entryPath))
      continue
    }
    if (entry.isFile()) files.push(entryPath)
  }
  return files
}

function collectOutputFiles(directory: string): string[] {
  return collectAllFiles(directory).filter((file) => {
    const extension = path.extname(file).toLowerCase()
    return inspectableOutputExtensions.has(extension) || outputFontExtensions.has(extension)
  })
}

function addFileViolation(file: string, message: string, line?: number): void {
  violations.push({ file: relativePath(file), line, message })
}

function sha256(filePath: string): string {
  return crypto.createHash('sha256').update(fs.readFileSync(filePath)).digest('hex')
}

function assertIncludes(file: string, source: string, expected: string, message: string): void {
  if (!source.includes(expected)) addFileViolation(file, message)
}

function normalizeFontFamily(value: string): string {
  return value
    .trim()
    .replace(/^['"`]|['"`]$/g, '')
    .trim()
}

function splitFontFamilyList(value: string): string[] {
  const families: string[] = []
  let current = ''
  let quote = ''
  let parenDepth = 0

  for (let index = 0; index < value.length; index += 1) {
    const character = value[index]
    const previous = index > 0 ? value[index - 1] : ''

    if (quote) {
      current += character
      if (character === quote && previous !== '\\') quote = ''
      continue
    }

    if (character === '"' || character === "'" || character === '`') {
      quote = character
      current += character
      continue
    }

    if (character === '(') {
      parenDepth += 1
      current += character
      continue
    }

    if (character === ')') {
      parenDepth = Math.max(0, parenDepth - 1)
      current += character
      continue
    }

    if (character === ',' && parenDepth === 0) {
      families.push(current)
      current = ''
      continue
    }

    current += character
  }

  families.push(current)
  return families
}

function collectFontVariableReferences(value: string): string[] {
  return [...value.matchAll(/var\(\s*(--[\w-]+)/gi)].map(match => match[1])
}

function collectFontVariableFallbacks(value: string): string[] {
  return [...value.matchAll(/var\(([^)]*)\)/gi)]
    .map((match) => {
      const [, ...fallbackParts] = splitFontFamilyList(match[1])
      return fallbackParts.join(',').trim()
    })
    .filter(Boolean)
}

function fontValueViolations(value: string): string[] {
  const violations: string[] = []
  const cleanedValue = value.trim()
  if (!cleanedValue) return violations

  const variableReferences = collectFontVariableReferences(cleanedValue)
  for (const variableReference of variableReferences) {
    if (!approvedFontVariables.has(variableReference)) {
      violations.push(`unapproved admin font variable "${variableReference}"`)
    }
  }

  if (variableReferences.length > 0) {
    for (const fallbackValue of collectFontVariableFallbacks(cleanedValue)) {
      violations.push(...fontValueViolations(fallbackValue))
    }
    return [...new Set(violations)]
  }

  const fontFamilies = splitFontFamilyList(cleanedValue)
    .map(normalizeFontFamily)
    .filter(Boolean)

  if (systemFallbackPattern.test(cleanedValue) || fontFamilies.some(fontFamily => cssFontFamilyKeywords.has(fontFamily.toLowerCase()))) {
    violations.push('forbidden system or generic font family')
  }

  for (const fontFamily of fontFamilies) {
    if (!systemFallbackPattern.test(fontFamily) && !cssFontFamilyKeywords.has(fontFamily.toLowerCase()) && !approvedFontFamilies.has(fontFamily)) {
      violations.push(`unapproved admin font family "${fontFamily}"`)
    }
  }

  return [...new Set(violations)]
}

function tailwindFontUtilityViolations(line: string): string[] {
  const violations: string[] = []

  for (const match of line.matchAll(tailwindFontSerifUtilityPattern)) {
    violations.push(`forbidden Tailwind font family utility "${match[1]}"`)
  }

  for (const match of line.matchAll(tailwindArbitraryFontUtilityPattern)) {
    const rawValue = match[2].trim()
    const normalizedValue = rawValue
      .replace(/\\_/g, '\u0000')
      .replace(/_/g, ' ')
      .replace(/\u0000/g, '_')
      .replace(/\\2c\s*/gi, ',')
      .replace(/\\,/g, ',')
      .trim()
    const typeHint = normalizedValue.match(/^([a-z-]+)\s*:\s*(.+)$/i)
    if (typeHint && !['family-name', 'font-family'].includes(typeHint[1].toLowerCase())) continue

    const value = (typeHint?.[2] || normalizedValue).trim()

    if (!value) continue
    if (/^var\(/i.test(value)) continue
    if (/^\d+(?:\.\d+)?$/.test(value) || cssFontWeightKeywords.has(value.toLowerCase())) continue
    if (!(/['",]/.test(rawValue) || /^(?:family-name|font-family)\s*:/i.test(rawValue) || systemFallbackPattern.test(value) || approvedFontFamilies.has(value) || /[A-Za-z]/.test(value))) continue

    violations.push(...fontValueViolations(value))
  }

  return [...new Set(violations)]
}

for (const file of collectSourceFiles(srcRoot)) {
  const source = fs.readFileSync(file, 'utf8')
  source.split(/\r?\n/).forEach((line, index) => {
    const match = line.match(forbiddenFontReference)
    if (match) {
      addFileViolation(file, `forbidden font reference "${match[0]}"`, index + 1)
    }

    for (const fontFamily of line.matchAll(fontFamilyAttributePattern)) {
      for (const violation of fontValueViolations(fontFamily[2])) {
        addFileViolation(file, violation, index + 1)
      }
    }

    for (const fontFamily of line.matchAll(fontFamilyDeclarationPattern)) {
      for (const violation of fontValueViolations(fontFamily[1])) {
        addFileViolation(file, violation, index + 1)
      }
    }

    for (const fontFamily of line.matchAll(fontFamilyAssignmentPattern)) {
      for (const violation of fontValueViolations(fontFamily[2])) {
        addFileViolation(file, violation, index + 1)
      }
    }

    for (const violation of tailwindFontUtilityViolations(line)) {
      addFileViolation(file, violation, index + 1)
    }
  })
}

if (!fs.existsSync(expectedFontPath)) {
  addFileViolation(expectedFontPath, 'required Maple UI admin font file is missing')
} else if (sha256(expectedFontPath) !== expectedFontHash) {
  addFileViolation(expectedFontPath, 'Maple UI admin font file hash does not match the approved CJK shard')
}

const fontFiles = fs.existsSync(fontRoot)
  ? collectAllFiles(fontRoot).filter(file => /\.(?:otf|ttf|woff|woff2)$/i.test(file))
  : []
for (const fontFile of fontFiles) {
  if (path.resolve(fontFile) !== path.resolve(expectedFontPath)) {
    addFileViolation(fontFile, 'unapproved admin font file')
  }
}

const legacyFontPath = path.join(fontRoot, 'system', 'AdminSystem-CN-Latin.woff2')
if (fs.existsSync(legacyFontPath)) {
  addFileViolation(legacyFontPath, 'obsolete AdminSystem font file still exists')
}

const adminCss = fs.readFileSync(adminCssPath, 'utf8')
if (/\blocal\s*\(/i.test(adminCss)) {
  addFileViolation(adminCssPath, 'admin @font-face must not use local() font sources')
}
assertIncludes(adminCssPath, adminCss, "font-family: 'MapleUICJK';", 'admin @font-face must use MapleUICJK')
assertIncludes(
  adminCssPath,
  adminCss,
  "src: url('../assets/fonts/maple-ui/MapleUI-CJK.f8ce6d72e8cb.woff2') format('woff2');",
  'admin @font-face must load the approved Maple UI CJK file',
)
assertIncludes(adminCssPath, adminCss, 'font-display: block;', 'admin @font-face must use font-display: block')
assertIncludes(adminCssPath, adminCss, "--tz-font-admin-ui: 'MapleUICJK';", 'admin font stack must not include system fallbacks')
assertIncludes(adminCssPath, adminCss, '--font-sans: var(--tz-font-admin-ui);', 'Tailwind font-sans must use the admin Maple UI stack')
assertIncludes(adminCssPath, adminCss, '--font-mono: var(--tz-font-admin-ui);', 'Tailwind font-mono must use the admin Maple UI stack')

const dashboardPresentation = fs.readFileSync(dashboardPresentationPath, 'utf8')
assertIncludes(
  dashboardPresentationPath,
  dashboardPresentation,
  "export const adminChartFontFamily = 'MapleUICJK'",
  'admin ECharts font family must be centralized',
)
assertIncludes(
  dashboardPresentationPath,
  dashboardPresentation,
  'fontFamily: adminChartFontFamily',
  'admin ECharts options must pass MapleUICJK to canvas text rendering',
)

if (checkDist) {
  if (!fs.existsSync(distRoot)) {
    addFileViolation(distRoot, 'admin dist directory is missing')
  } else {
    const distFiles = collectOutputFiles(distRoot)

    for (const file of distFiles) {
      const extension = path.extname(file).toLowerCase()
      const outputPath = relativePath(file)

      if (outputFontExtensions.has(extension)) {
        const filename = path.basename(file)
        if (!/^MapleUI-CJK\.f8ce6d72e8cb-[A-Za-z0-9_-]+\.woff2$/.test(filename)) {
          addFileViolation(file, 'unapproved admin production font file')
        }
        continue
      }

      const source = fs.readFileSync(file, 'utf8')
      const outputReferencePattern = extension === '.css' || extension === '.html'
        ? forbiddenStylesheetOutputReference
        : forbiddenOutputReference
      if (outputReferencePattern.test(source)) {
        addFileViolation(file, `forbidden admin production font reference in ${outputPath}`)
      }

      if (extension === '.css' || extension === '.html') {
        source.split(/\r?\n/).forEach((line, index) => {
          const match = line.match(forbiddenFontReference)
          if (match) {
            addFileViolation(file, `forbidden admin production font reference "${match[0]}"`, index + 1)
          }
        })
      }
    }

    const outputFontFiles = distFiles.filter((file) => outputFontExtensions.has(path.extname(file).toLowerCase()))
    if (outputFontFiles.length !== 1) {
      addFileViolation(distRoot, `expected exactly one admin production font file, found ${outputFontFiles.length}`)
    }
  }
}

if (violations.length > 0) {
  console.error('Admin font policy violations:')
  for (const violation of violations) {
    const location = violation.line ? `${violation.file}:${violation.line}` : violation.file
    console.error(`- ${location}: ${violation.message}`)
  }
  process.exit(1)
}

console.log(`Admin font policy passed with ${relativePath(expectedFontPath)}.`)

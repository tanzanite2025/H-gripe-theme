import fs from 'node:fs'
import path from 'node:path'

const sourceExtensions = new Set(['.css', '.html', '.json', '.less', '.sass', '.scss', '.ts', '.vue'])
const ignoredDirectoryNames = new Set(['.git', '.nuxt', '.output', '.playwright-cli', '.vs', 'dist', 'node_modules'])
const systemFallbackPattern = /\b(?:-apple-system|arial|blinkmacsystemfont|calibri|cambria|consolas|courier(?: new)?|cursive|fangsong|fantasy|georgia|helvetica(?: neue)?|liberation(?: mono| sans| serif)?|math|menlo|monaco|monospace|sans-serif|segoe ui|serif|sf mono|sf pro|sfmono-regular|system-ui|tahoma|times(?: new roman)?|trebuchet ms|ui-monospace|ui-rounded|ui-sans-serif|ui-serif|verdana)\b/i
const fontFamilyDeclarationPattern = /(?:\bfont-family\s*:|\bfontFamily\s*:|\[font-family:)/i
const fontShorthandDeclarationPattern = /(?:\bfont\s*:|\[font:)/i
const storefrontFontVariableDeclarationPattern = /--tz-font-[\w-]+\s*:/i
const legacyFontPattern = /(?:\b(?:aerialfaster|storefrontsystem|storefront-system|nuxt-fonts(?:-global)?)\b|@nuxt\/fonts)/i
const localeFontSelectorPattern = /\bfontFamily\s*:\s*['"](?:latin|latin-accent|maple-ui|arabic|devanagari|thai)['"]/i
const localFontSourcePattern = /\blocal\s*\(/i
const externalFontSourcePattern = /(?:@import\s+(?:url\(\s*)?['"]?(?:https?:)?\/\/[^'")\s]+|url\(\s*['"]?(?:(?:https?:)?\/\/|data:font\/)[^'")\s]+|fonts\.(?:googleapis|gstatic)\.com|use\.typekit\.net)/i
const fontFamilyAttributePattern = /(?<![\w-])font-family\s*=\s*(['"])(.*?)\1/gi
const tailwindFontSerifUtilityPattern = /(?:^|[\s"'`])((?:[\w-]+:)*font-serif)(?=$|[\s"'`])/gi
const tailwindArbitraryFontUtilityPattern = /(?:^|[\s"'`])((?:[\w-]+:)*font-\[([^\]\r\n]+)\])(?=$|[\s"'`])/gi
const cssFontSystemKeywords = new Set(['caption', 'icon', 'menu', 'message-box', 'small-caption', 'status-bar'])
const cssFontFamilyKeywords = new Set(['inherit', 'initial', 'unset', 'revert', 'revert-layer'])
const cssFontWeightKeywords = new Set(['bold', 'bolder', 'lighter', 'normal'])
const approvedStorefrontFontVariables = new Set([
  '--tz-font-base',
  '--tz-font-ui',
  '--tz-font-latin',
  '--tz-font-maple-ui',
  '--tz-font-latin-accents',
  '--tz-font-arabic',
  '--tz-font-devanagari',
  '--tz-font-thai',
])
const approvedStorefrontFontFamilies = new Set([
  'MapleUICJK',
  'MapleUILatin',
  'MapleUICoverageNotoSansArabic',
  'MapleUICoverageNotoSansDevanagari',
  'MapleUICoverageNotoSansLatinAccents',
  'MapleUICoverageNotoSansThai',
])

const declarationValue = (line: string, pattern: RegExp): string => {
  const declarationIndex = line.search(pattern)
  if (declarationIndex < 0) return ''

  const colonIndex = line.indexOf(':', declarationIndex)
  if (colonIndex < 0) return ''

  return line.slice(colonIndex + 1).trim()
}

const valueBeforeTerminator = (value: string, stopAtTopLevelComma = false): string => {
  let quote = ''
  let parenDepth = 0

  for (let index = 0; index < value.length; index += 1) {
    const character = value[index]
    const previous = index > 0 ? value[index - 1] : ''

    if (quote) {
      if (character === quote && previous !== '\\') quote = ''
      continue
    }

    if (character === '"' || character === "'" || character === '`') {
      quote = character
      continue
    }
    if (character === '(') {
      parenDepth += 1
      continue
    }
    if (character === ')') {
      parenDepth = Math.max(0, parenDepth - 1)
      continue
    }
    if (parenDepth === 0 && (
      character === ';'
      || character === ']'
      || character === '}'
      || (stopAtTopLevelComma && character === ',')
    )) {
      return value.slice(0, index)
    }
  }

  return value
}

const splitFontFamilyList = (value: string): string[] => {
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

const normalizeFontFamily = (value: string): string => (
  value
    .trim()
    .replace(/!\s*important\b/i, '')
    .replace(/^['"`]|['"`]$/g, '')
    .trim()
)

const isDynamicFontFamily = (fontFamily: string): boolean => (
  fontFamily.includes('${')
  || /\b(?:var|theme|storefrontFontFamilyForLocale|computed|ref)\s*\(/i.test(fontFamily)
)

const collectFontVariableReferences = (value: string): string[] => (
  [...value.matchAll(/var\(\s*(--[\w-]+)/gi)].map(match => match[1])
)

const collectFontVariableFallbacks = (value: string): string[] => (
  [...value.matchAll(/var\(([^)]*)\)/gi)]
    .map((match) => {
      const [, ...fallbackParts] = splitFontFamilyList(match[1])
      return fallbackParts.join(',').trim()
    })
    .filter(Boolean)
)

const staticFontFamiliesFromList = (value: string): string[] => (
  splitFontFamilyList(value)
    .flatMap((fontFamily) => {
      const normalized = normalizeFontFamily(fontFamily)
      if (!normalized || cssFontFamilyKeywords.has(normalized.toLowerCase()) || isDynamicFontFamily(normalized)) {
        return []
      }
      if (/^[^'"`]+,[^'"`]+$/.test(normalized)) {
        return staticFontFamiliesFromList(normalized)
      }
      return [normalized]
    })
)

const fontValueViolations = (value: string): string[] => {
  const violations: string[] = []
  const fontFamilies = [
    ...staticFontFamiliesFromList(value),
    ...collectFontVariableFallbacks(value).flatMap(staticFontFamiliesFromList),
  ]

  for (const variableName of collectFontVariableReferences(value)) {
    if (!approvedStorefrontFontVariables.has(variableName)) {
      violations.push(`unapproved storefront font variable "${variableName}"`)
    }
  }

  if (systemFallbackPattern.test(value) || fontFamilies.some(fontFamily => cssFontSystemKeywords.has(fontFamily.toLowerCase()))) {
    violations.push('forbidden system or generic font family')
  }

  for (const fontFamily of fontFamilies) {
    if (!systemFallbackPattern.test(fontFamily) && !cssFontSystemKeywords.has(fontFamily.toLowerCase()) && !approvedStorefrontFontFamilies.has(fontFamily)) {
      violations.push(`unapproved Maple UI font family "${fontFamily}"`)
    }
  }

  return [...new Set(violations)]
}

const fontFamilyDeclarationFamilies = (line: string): string[] => {
  const value = valueBeforeTerminator(declarationValue(line, fontFamilyDeclarationPattern))
  const normalizedValue = value.trim()
  if (!normalizedValue || normalizedValue.startsWith('{')) return []
  if (/\bfontFamily\s*:/.test(line) && /^[A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)*$/.test(normalizedValue)) {
    return []
  }
  return staticFontFamiliesFromList(value)
}

const fontFamilyDeclarationValue = (line: string): string => {
  const value = valueBeforeTerminator(declarationValue(line, fontFamilyDeclarationPattern))
  const normalizedValue = value.trim()
  if (!normalizedValue || normalizedValue.startsWith('{')) return ''
  if (/\bfontFamily\s*:/.test(line) && /^[A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)*$/.test(normalizedValue)) {
    return ''
  }
  return value
}

const fontShorthandFamilyValue = (line: string): string => {
  const rawValue = valueBeforeTerminator(declarationValue(line, fontShorthandDeclarationPattern))
    .replace(/_/g, ' ')
    .trim()
  if (!rawValue || isDynamicFontFamily(rawValue)) return ''

  const lowerValue = normalizeFontFamily(rawValue).toLowerCase()
  if (cssFontFamilyKeywords.has(lowerValue)) return ''
  if (cssFontSystemKeywords.has(lowerValue)) return rawValue

  const fontSizePattern = /(?:^|\s)(?:xx-small|x-small|small|medium|large|x-large|xx-large|larger|smaller|(?:clamp|calc|min|max)\([^)]*\)|-?\d*\.?\d+(?:px|r?em|%|vh|vw|vmin|vmax|pt|pc|in|cm|mm|q|ch|ex))(?:\s*\/\s*(?:normal|-?\d*\.?\d+(?:px|r?em|%|vh|vw|vmin|vmax|pt|pc|in|cm|mm|q|ch|ex)?|(?:clamp|calc|min|max)\([^)]*\)))?\s+(.+)$/i
  const match = rawValue.match(fontSizePattern)
  return match?.[1]?.trim() || ''
}

const fontShorthandFamilies = (line: string): string[] => {
  const value = fontShorthandFamilyValue(line)
  if (!value) return []
  return cssFontSystemKeywords.has(normalizeFontFamily(value).toLowerCase())
    ? [value]
    : staticFontFamiliesFromList(value)
}

const basicFontPolicyLineViolations = (line: string): string[] => {
  const violations: string[] = []

  if (legacyFontPattern.test(line)) {
    violations.push('obsolete or secondary font authority reference')
  }

  if (localFontSourcePattern.test(line)) {
    violations.push('local font source')
  }

  if (externalFontSourcePattern.test(line)) {
    violations.push('external font source')
  }

  return violations
}

const fontPolicyDeclarationViolations = (line: string): string[] => {
  const violations: string[] = []

  if (!fontFamilyDeclarationPattern.test(line) && !fontShorthandDeclarationPattern.test(line)) {
    return violations
  }

  if (localeFontSelectorPattern.test(line)) {
    return violations
  }

  violations.push(...fontValueViolations(fontFamilyDeclarationValue(line)))
  violations.push(...fontValueViolations(fontShorthandFamilies(line).join(', ')))

  return [...new Set(violations)]
}

const fontFamilyAttributeViolations = (line: string): string[] => {
  const violations: string[] = []

  for (const match of line.matchAll(fontFamilyAttributePattern)) {
    violations.push(...fontValueViolations(match[2]))
  }

  return [...new Set(violations)]
}

const normalizeTailwindArbitraryFontFamily = (value: string): string => {
  const decodedValue = value
    .trim()
    .replace(/\\_/g, '\u0000')
    .replace(/_/g, ' ')
    .replace(/\u0000/g, '_')
    .replace(/\\2c\s*/gi, ',')
    .replace(/\\,/g, ',')
  const typeHint = decodedValue.match(/^([a-z-]+)\s*:\s*(.+)$/i)

  if (typeHint && !['family-name', 'font-family'].includes(typeHint[1].toLowerCase())) {
    return ''
  }

  return (typeHint?.[2] || decodedValue).trim()
}

const isTailwindArbitraryFontFamily = (rawValue: string, normalizedValue: string): boolean => {
  if (/^([a-z-]+)\s*:/i.test(rawValue) && !/^(?:family-name|font-family)\s*:/i.test(rawValue)) return false
  if (/^(?:family-name|font-family)\s*:/i.test(rawValue)) return true
  if (/^var\(/i.test(normalizedValue)) return false
  if (/^\d+(?:\.\d+)?$/.test(normalizedValue) || cssFontWeightKeywords.has(normalizedValue.toLowerCase())) return false
  return true
}

const tailwindFontUtilityViolations = (line: string): string[] => {
  const violations: string[] = []

  for (const match of line.matchAll(tailwindFontSerifUtilityPattern)) {
    violations.push(`forbidden Tailwind font family utility "${match[1]}"`)
  }

  for (const match of line.matchAll(tailwindArbitraryFontUtilityPattern)) {
    const normalizedValue = normalizeTailwindArbitraryFontFamily(match[2])
    if (!normalizedValue) continue
    if (!isTailwindArbitraryFontFamily(match[2], normalizedValue)) continue
    violations.push(...fontValueViolations(normalizedValue))
  }

  return [...new Set(violations)]
}

export const fontPolicyLineViolations = (line: string): string[] => (
  [...new Set([
    ...basicFontPolicyLineViolations(line),
    ...fontPolicyDeclarationViolations(line),
    ...fontFamilyAttributeViolations(line),
    ...tailwindFontUtilityViolations(line),
  ])]
)

export const fontPolicySourceViolations = (source: string): Array<{ line: number; violation: string }> => {
  const violations: Array<{ line: number; violation: string }> = []
  const lines = source.split(/\r?\n/)

  for (const [index, line] of lines.entries()) {
    for (const violation of basicFontPolicyLineViolations(line)) {
      violations.push({ line: index + 1, violation })
    }
    for (const violation of fontFamilyAttributeViolations(line)) {
      violations.push({ line: index + 1, violation })
    }
    for (const violation of tailwindFontUtilityViolations(line)) {
      violations.push({ line: index + 1, violation })
    }
  }

  const declarationStartPattern = /(?:\bfont-family\s*:|\bfontFamily\s*:|\[font-family:|\bfont\s*:|\[font:|--tz-font-[\w-]+\s*:)/gi
  for (const match of source.matchAll(declarationStartPattern)) {
    const declarationStart = match.index ?? 0
    const isJavaScriptFontFamily = /\bfontFamily\s*:$/.test(match[0])
    const declaration = `${match[0]}${valueBeforeTerminator(
      source.slice(declarationStart + match[0].length),
      isJavaScriptFontFamily,
    )}`
    const line = source.slice(0, declarationStart).split(/\r?\n/).length

    const declarationViolations = storefrontFontVariableDeclarationPattern.test(match[0])
      ? fontValueViolations(valueBeforeTerminator(source.slice(declarationStart + match[0].length)))
      : fontPolicyDeclarationViolations(declaration)

    for (const violation of declarationViolations) {
      violations.push({ line, violation })
    }
  }

  const staticFontFamilyAssignmentPattern = /(?:const|let|var)\s+[A-Za-z_$][\w$]*fontFamily[A-Za-z_$\w]*\s*=\s*(['"`])([\s\S]*?)\1/gi
  for (const match of source.matchAll(staticFontFamilyAssignmentPattern)) {
    const assignmentStart = match.index ?? 0
    const line = source.slice(0, assignmentStart).split(/\r?\n/).length
    const quote = match[1]
    const value = quote === '`'
      ? match[2].replace(/\$\{[^}]*\}/g, ',')
      : match[2]

    for (const violation of fontValueViolations(value)) {
      violations.push({ line, violation })
    }
  }

  const seen = new Set<string>()
  return violations.filter((finding) => {
    const key = `${finding.line}:${finding.violation}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

export const collectFontPolicySourceFiles = (projectDir: string): string[] => {
  const sourceRoots = [
    path.join(projectDir, 'app'),
    path.join(projectDir, 'config'),
    path.join(projectDir, 'docs'),
    path.join(projectDir, 'layers'),
    path.join(projectDir, 'public'),
    path.join(projectDir, 'server'),
    path.join(projectDir, 'tests'),
    path.join(projectDir, 'types'),
  ]
  const standaloneSources = [
    path.join(projectDir, 'nuxt.config.ts'),
    path.join(projectDir, 'package.json'),
    path.join(projectDir, 'package-lock.json'),
    path.join(projectDir, 'tailwind.config.ts'),
    path.join(projectDir, 'public', 'fonts', 'maple-ui.css'),
  ]

  const collectSourceFiles = (directory: string): string[] => {
    if (!fs.existsSync(directory)) return []

    const files: string[] = []
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      if (entry.isDirectory()) {
        if (!ignoredDirectoryNames.has(entry.name)) {
          files.push(...collectSourceFiles(path.join(directory, entry.name)))
        }
        continue
      }

      if (entry.isFile() && sourceExtensions.has(path.extname(entry.name))) {
        files.push(path.join(directory, entry.name))
      }
    }

    return files
  }

  return [...new Set([
    ...sourceRoots.flatMap(collectSourceFiles),
    ...standaloneSources.filter(fs.existsSync),
  ])]
}

export const collectFontPolicyViolations = (projectDir: string): string[] => {
  const violations: string[] = []

  for (const sourceFile of collectFontPolicySourceFiles(projectDir)) {
    const source = fs.readFileSync(sourceFile, 'utf8')
    for (const finding of fontPolicySourceViolations(source)) {
      const location = `${path.relative(projectDir, sourceFile).replace(/\\/g, '/')}:${finding.line}`
      violations.push(`${location}: ${finding.violation}`)
    }
  }

  const obsoleteFontPaths = [
    path.join(projectDir, 'app', 'assets', 'fonts', 'AerialFasterRegular.woff'),
    path.join(projectDir, 'app', 'assets', 'fonts', 'AerialFasterRegular.woff2'),
    path.join(projectDir, 'public', 'fonts', 'AerialFasterRegular.woff'),
    path.join(projectDir, 'public', 'fonts', 'AerialFasterRegular.woff2'),
    path.join(projectDir, 'public', 'fonts', 'storefront-system.css'),
    path.join(projectDir, 'public', 'fonts', 'StorefrontSystem-Latin.00af3fec5b34.woff2'),
    path.join(projectDir, 'public', 'fonts', 'StorefrontSystem-CJK.f8ce6d72e8cb.woff2'),
    path.join(projectDir, 'public', 'fonts', 'StorefrontSystem-Latin-Accents.e645edc952b6.woff2'),
    path.join(projectDir, 'public', 'fonts', 'StorefrontSystem-Arabic.ce85091f0209.woff2'),
    path.join(projectDir, 'public', 'fonts', 'StorefrontSystem-Devanagari.3b3cae4d2600.woff2'),
    path.join(projectDir, 'public', 'fonts', 'StorefrontSystem-Thai.1f5a173641bb.woff2'),
  ]

  for (const obsoleteFontPath of obsoleteFontPaths) {
    if (fs.existsSync(obsoleteFontPath)) {
      violations.push(`${path.relative(projectDir, obsoleteFontPath).replace(/\\/g, '/')}: obsolete font file exists`)
    }
  }

  return violations
}

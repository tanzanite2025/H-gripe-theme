import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const sourceRoots = [
  path.join(projectDir, 'app'),
  path.join(projectDir, 'config'),
  path.join(projectDir, 'docs'),
  path.join(projectDir, 'server'),
  path.join(projectDir, 'tests'),
  path.join(projectDir, 'types'),
]
const standaloneSources = [
  path.join(projectDir, 'nuxt.config.ts'),
  path.join(projectDir, 'tailwind.config.ts'),
  path.join(projectDir, 'public', 'fonts', 'storefront-system.css'),
]
const sourceExtensions = new Set(['.css', '.html', '.less', '.sass', '.scss', '.ts', '.vue'])
const ignoredDirectoryNames = new Set(['.git', '.nuxt', '.output', '.playwright-cli', '.vs', 'dist', 'node_modules'])
const systemFallbackPattern = /\b(?:-apple-system|blinkmacsystemfont|consolas|courier new|inter|liberation mono|menlo|monaco|monospace|sans-serif|serif|segoe ui|sfmono-regular|system-ui|ui-monospace|ui-sans-serif)\b/i
const fontDeclarationPattern = /(?:font-family|fontFamily|\[font-family:)/i
const legacyFontPattern = /\baerialfaster\b/i

function collectSourceFiles(directory: string): string[] {
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

function formatPath(filePath: string, lineNumber: number): string {
  return `${path.relative(projectDir, filePath).replace(/\\/g, '/')}:${lineNumber}`
}

const violations: string[] = []
const sourceFiles = [
  ...sourceRoots.flatMap(collectSourceFiles),
  ...standaloneSources.filter(fs.existsSync),
]

for (const sourceFile of sourceFiles) {
  const lines = fs.readFileSync(sourceFile, 'utf8').split(/\r?\n/)

  for (const [index, line] of lines.entries()) {
    const location = formatPath(sourceFile, index + 1)

    if (legacyFontPattern.test(line)) {
      violations.push(`${location}: obsolete AerialFaster reference`)
    }

    if (fontDeclarationPattern.test(line) && systemFallbackPattern.test(line)) {
      violations.push(`${location}: system or generic font fallback`)
    }
  }
}

const obsoleteFontPaths = [
  path.join(projectDir, 'app', 'assets', 'fonts', 'AerialFasterRegular.woff'),
  path.join(projectDir, 'app', 'assets', 'fonts', 'AerialFasterRegular.woff2'),
]

for (const obsoleteFontPath of obsoleteFontPaths) {
  if (fs.existsSync(obsoleteFontPath)) {
    violations.push(`${path.relative(projectDir, obsoleteFontPath).replace(/\\/g, '/')}: obsolete font file exists`)
  }
}

if (violations.length > 0) {
  console.error('Font policy violations:')
  for (const violation of violations) {
    console.error(`- ${violation}`)
  }
  process.exit(1)
}

console.log(`Font policy passed for ${sourceFiles.length} source files.`)

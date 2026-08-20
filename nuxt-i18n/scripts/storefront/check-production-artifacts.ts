import { readdir, readFile } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { basename, extname, join, relative } from 'node:path'
import { fontPolicySourceViolations } from '../font-policy-utils.ts'
import { storefrontFontArtifactViolations } from './font-artifact-contract.ts'

const outputDirectory = fileURLToPath(new URL('../../.output/', import.meta.url))
const sourceMapReference = /\/\/[#@]\s*sourceMappingURL=/i
const inspectableExtensions = new Set(['.css', '.html', '.js', '.json', '.mjs'])
const forbiddenOutputReferencePattern = /(?:\b(?:StorefrontSystem|storefront-system|nuxt-fonts(?:-global)?)\b|@nuxt\/fonts|\/_fonts\/|fonts\.(?:googleapis|gstatic)\.com|use\.typekit\.net|data:font\/|local\s*\()/i

async function walk(directory: string): Promise<string[]> {
  const entries = await readdir(directory, { withFileTypes: true })
  const files: string[] = []

  for (const entry of entries) {
    const filePath = join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...await walk(filePath))
      continue
    }
    if (entry.isFile()) files.push(filePath)
  }

  return files
}

if (!existsSync(outputDirectory)) {
  console.error('Production output directory is missing. Run `npm run build` before checking production artifacts.')
  process.exit(1)
}

const files = await walk(outputDirectory)
const sourceMaps = files.filter((filePath) => basename(filePath).endsWith('.map'))
const sourceMapReferences: string[] = []
const fontPolicyViolations: string[] = []
const forbiddenOutputReferences: string[] = []

for (const filePath of files) {
  if (!inspectableExtensions.has(extname(filePath))) continue
  const outputPath = relative(outputDirectory, filePath).replaceAll('\\', '/')
  const outputFilename = basename(filePath)

  // Nitro may externalize third-party server dependencies into .output.
  // They are not browser-deliverable storefront assets, so only enforce
  // source-map reference removal for the application's own bundles.
  if (outputPath.includes('/node_modules/')) continue

  const content = await readFile(filePath, 'utf8')
  if (sourceMapReference.test(content)) {
    sourceMapReferences.push(outputPath)
  }
  const forbiddenReference = content.match(forbiddenOutputReferencePattern)
  if (forbiddenReference) {
    forbiddenOutputReferences.push(`${outputPath}: forbidden font artifact reference "${forbiddenReference[0]}"`)
  }
  const isBuiltStyleModule = outputPath.startsWith('server/chunks/build/')
    && /(?:^|[-_])styles?(?:[.-]|$)/i.test(outputFilename)

  if (extname(filePath) === '.css' || extname(filePath) === '.html' || isBuiltStyleModule) {
    for (const finding of fontPolicySourceViolations(content)) {
      fontPolicyViolations.push(`${outputPath}:${finding.line}: ${finding.violation}`)
    }
    for (const finding of storefrontFontArtifactViolations(content)) {
      fontPolicyViolations.push(`${outputPath}:${finding.line}: ${finding.violation}`)
    }
  }
}

const obsoleteFontArtifacts = files
  .map(filePath => relative(outputDirectory, filePath).replaceAll('\\', '/'))
  .filter(outputPath => /(?:^|\/)(?:StorefrontSystem-[^/]+\.woff2|storefront-system\.css)$/.test(outputPath))

if (
  sourceMaps.length > 0
  || sourceMapReferences.length > 0
  || fontPolicyViolations.length > 0
  || forbiddenOutputReferences.length > 0
  || obsoleteFontArtifacts.length > 0
) {
  if (sourceMaps.length > 0) {
    console.error('Production output includes source-map files:')
    for (const filePath of sourceMaps) console.error(`- ${relative(outputDirectory, filePath)}`)
  }
  if (sourceMapReferences.length > 0) {
    console.error('Production output includes source-map references:')
    for (const filePath of sourceMapReferences) console.error(`- ${filePath}`)
  }
  if (fontPolicyViolations.length > 0) {
    console.error('Production output includes storefront font policy violations:')
    for (const violation of fontPolicyViolations) console.error(`- ${violation}`)
  }
  if (forbiddenOutputReferences.length > 0) {
    console.error('Production output includes forbidden font authority references:')
    for (const violation of forbiddenOutputReferences) console.error(`- ${violation}`)
  }
  if (obsoleteFontArtifacts.length > 0) {
    console.error('Production output includes obsolete StorefrontSystem font artifacts:')
    for (const filePath of obsoleteFontArtifacts) console.error(`- ${filePath}`)
  }
  process.exitCode = 1
} else {
  console.log('Production artifact check passed: source maps are absent and built output keeps the Maple UI font baseline.')
}

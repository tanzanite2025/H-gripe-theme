import { readdir, readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { basename, extname, join, relative } from 'node:path'

const outputDirectory = fileURLToPath(new URL('../../.output/', import.meta.url))
const sourceMapReference = /\/\/[#@]\s*sourceMappingURL=/i
const inspectableExtensions = new Set(['.css', '.js', '.mjs'])

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

const files = await walk(outputDirectory)
const sourceMaps = files.filter((filePath) => basename(filePath).endsWith('.map'))
const sourceMapReferences: string[] = []

for (const filePath of files) {
  if (!inspectableExtensions.has(extname(filePath))) continue
  const outputPath = relative(outputDirectory, filePath).replaceAll('\\', '/')

  // Nitro may externalize third-party server dependencies into .output.
  // They are not browser-deliverable storefront assets, so only enforce
  // source-map reference removal for the application's own bundles.
  if (outputPath.includes('/node_modules/')) continue

  const content = await readFile(filePath, 'utf8')
  if (sourceMapReference.test(content)) {
    sourceMapReferences.push(outputPath)
  }
}

if (sourceMaps.length > 0 || sourceMapReferences.length > 0) {
  if (sourceMaps.length > 0) {
    console.error('Production output includes source-map files:')
    for (const filePath of sourceMaps) console.error(`- ${relative(outputDirectory, filePath)}`)
  }
  if (sourceMapReferences.length > 0) {
    console.error('Production output includes source-map references:')
    for (const filePath of sourceMapReferences) console.error(`- ${filePath}`)
  }
  process.exitCode = 1
} else {
  console.log('Production artifact check passed: source maps are absent.')
}

import { readdir, readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { basename, extname, join, relative } from 'node:path'

const outputDirectory = fileURLToPath(new URL('../../.output/', import.meta.url))
const sourceMapReference = /\/\/[#@]\s*sourceMappingURL=/i
const inspectableExtensions = new Set(['.css', '.js', '.mjs'])

async function walk(directory) {
  const entries = await readdir(directory, { withFileTypes: true })
  const files = []

  for (const entry of entries) {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...await walk(path))
      continue
    }
    if (entry.isFile()) files.push(path)
  }

  return files
}

const files = await walk(outputDirectory)
const sourceMaps = files.filter(path => basename(path).endsWith('.map'))
const sourceMapReferences = []

for (const path of files) {
  if (!inspectableExtensions.has(extname(path))) continue
  const outputPath = relative(outputDirectory, path).replaceAll('\\', '/')

  // Nitro may externalize third-party server dependencies into .output.
  // They are not browser-deliverable storefront assets, so only enforce
  // source-map reference removal for the application's own bundles.
  if (outputPath.includes('/node_modules/')) continue

  const content = await readFile(path, 'utf8')
  if (sourceMapReference.test(content)) {
    sourceMapReferences.push(outputPath)
  }
}

if (sourceMaps.length > 0 || sourceMapReferences.length > 0) {
  if (sourceMaps.length > 0) {
    console.error('Production output includes source-map files:')
    for (const path of sourceMaps) console.error(`- ${relative(outputDirectory, path)}`)
  }
  if (sourceMapReferences.length > 0) {
    console.error('Production output includes source-map references:')
    for (const path of sourceMapReferences) console.error(`- ${path}`)
  }
  process.exitCode = 1
} else {
  console.log('Production artifact check passed: source maps are absent.')
}

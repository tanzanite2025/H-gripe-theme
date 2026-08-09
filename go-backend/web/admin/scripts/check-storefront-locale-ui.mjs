import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const adminRoot = path.resolve(scriptDir, '..')
const srcRoot = path.resolve(adminRoot, 'src')

const forbiddenBindings = [
  {
    label: 'native select locale binding',
    pattern: /<select\b[^>]*\bv-model(?:\.[^=]+)?\s*=\s*["'][^"']*locale[^"']*["'][^>]*>/g,
  },
  {
    label: 'free text locale binding',
    pattern: /<(?:Input|Textarea|input|textarea)\b[^>]*\bv-model(?:\.[^=]+)?\s*=\s*["'][^"']*locale[^"']*["'][^>]*>/g,
  },
]

function collectSourceFiles(dir) {
  const result = []
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === 'node_modules' || entry.name === 'dist' || entry.name === 'output') continue
    const fullPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      result.push(...collectSourceFiles(fullPath))
      continue
    }
    if (/\.(vue|ts)$/.test(entry.name)) result.push(fullPath)
  }
  return result
}

function lineNumber(source, index) {
  return source.slice(0, index).split(/\r\n|\r|\n/).length
}

const violations = []
for (const file of collectSourceFiles(srcRoot)) {
  const source = fs.readFileSync(file, 'utf8')
  for (const binding of forbiddenBindings) {
    for (const match of source.matchAll(binding.pattern)) {
      violations.push({
        file: path.relative(adminRoot, file).replace(/\\/g, '/'),
        line: lineNumber(source, match.index || 0),
        label: binding.label,
        snippet: match[0].replace(/\s+/g, ' ').trim(),
      })
    }
  }
}

if (violations.length > 0) {
  console.error('Storefront locale UI check failed.')
  console.error('Use StorefrontLocaleSelect or fixed locale option sets for storefront-facing locale fields.')
  for (const violation of violations) {
    console.error(`- ${violation.file}:${violation.line} ${violation.label}: ${violation.snippet}`)
  }
  process.exit(1)
}

console.log('Storefront locale UI bindings are controlled.')

import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { collectFontPolicySourceFiles, collectFontPolicyViolations } from './font-policy-utils.ts'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..')
const sourceFiles = collectFontPolicySourceFiles(projectDir)
const violations = collectFontPolicyViolations(projectDir)

if (violations.length > 0) {
  console.error('Font policy violations:')
  for (const violation of violations) {
    console.error(`- ${violation}`)
  }
  process.exit(1)
}

console.log(`Font policy passed for ${sourceFiles.length} source files.`)

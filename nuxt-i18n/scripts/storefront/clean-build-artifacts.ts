import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectDir = path.resolve(scriptDir, '..', '..')
const allowedTargets = ['.nuxt', '.output', 'dist', path.join('node_modules', '.cache')]

const assertSafeTarget = (target: string): string => {
  const absoluteTarget = path.resolve(projectDir, target)
  const relativeTarget = path.relative(projectDir, absoluteTarget)

  if (
    relativeTarget.startsWith('..')
    || path.isAbsolute(relativeTarget)
    || !allowedTargets.includes(relativeTarget)
  ) {
    throw new Error(`Refusing to remove unsafe build artifact target: ${absoluteTarget}`)
  }

  return absoluteTarget
}

for (const target of allowedTargets) {
  const absoluteTarget = assertSafeTarget(target)
  fs.rmSync(absoluteTarget, { recursive: true, force: true })
  console.log(`Removed ${path.relative(projectDir, absoluteTarget).replace(/\\/g, '/')}`)
}

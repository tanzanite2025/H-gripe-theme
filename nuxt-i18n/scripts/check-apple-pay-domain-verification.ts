import { existsSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

const projectRoot = process.cwd()
const relativePath = join('public', '.well-known', 'apple-developer-merchantid-domain-association')
const filePath = join(projectRoot, relativePath)

if (!existsSync(filePath)) {
  if (process.env.NODE_ENV === 'production' || process.env.REQUIRE_APPLE_PAY_DOMAIN_VERIFICATION === '1') {
    throw new Error(`Apple Pay domain verification file is missing: ${relativePath}. Download it from Stripe Dashboard and place it at this path.`)
  }
  console.warn(`[apple-pay] domain verification file is not present: ${relativePath}`)
  process.exit(0)
}

const stats = statSync(filePath)
const contents = readFileSync(filePath, 'utf8').trim()
if (!stats.isFile() || !contents || contents.includes('PLACEHOLDER') || contents.includes('REPLACE_ME')) {
  throw new Error(`Apple Pay domain verification file is empty or contains placeholder content: ${relativePath}`)
}

console.log(`[apple-pay] verified domain association asset: ${relativePath}`)

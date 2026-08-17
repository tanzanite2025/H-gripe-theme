import { spawn } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { gzipSync } from 'node:zlib'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const serverEntry = resolve(projectRoot, '.output/server/index.mjs')
const publicDirectory = resolve(projectRoot, '.output/public')
const port = Number.parseInt(process.env.CRITICAL_CSS_CHECK_PORT || '4021', 10)
const origin = `http://127.0.0.1:${port}`
const maxCriticalCssGzipBytes = 35 * 1024
const forbiddenStylesheetNames = [
  'AuthModal',
  'BrowsingHistoryDark',
  'CartDrawer',
  'CheckoutModal',
  'HeaderMegaMenu',
  'LeverAndPoint',
  'GlobalProductDetailBottomSheet',
  'ProductCategoryNavigationCards',
  'QuickBuyEntryRouterPopover',
  'ShopProductDisplayCard',
  'ShopProductQuickSearchForm',
  'ShopSearchSheet',
  'WhatsAppChatModal',
  'guide-sections',
  'whatsapp-mobile-drawer',
]

const timeout = (ms: number): Promise<void> => new Promise(resolveTimeout => setTimeout(resolveTimeout, ms))

const findStylesheetHrefs = (html: string): string[] => {
  const linkTags = html.match(/<link\b[^>]*>/gi) || []

  return linkTags
    .filter(tag => /\brel=(['"])stylesheet\1/i.test(tag))
    .map((tag) => tag.match(/\bhref=(['"])(.*?)\1/i)?.[2] || '')
    .filter(Boolean)
}

const waitForReady = async (): Promise<void> => {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await fetch(`${origin}/`)
      if (response.status < 500) return
    } catch {
      // The preview server is still starting.
    }
    await timeout(250)
  }

  throw new Error(`Nuxt preview did not become ready on ${origin}`)
}

const getAssetPath = (href: string): string => {
  const pathname = new URL(href, origin).pathname
  if (!pathname.startsWith('/_nuxt/')) {
    throw new Error(`Unexpected critical stylesheet path: ${href}`)
  }

  return resolve(publicDirectory, `.${pathname}`)
}

if (!existsSync(serverEntry)) {
  console.error(`[critical-css] FAILED: Missing ${serverEntry}. Run npm run build first.`)
  process.exit(1)
}

const child = spawn(process.execPath, [serverEntry], {
  cwd: projectRoot,
  env: {
    ...process.env,
    NODE_ENV: 'production',
    HOST: '127.0.0.1',
    PORT: String(port),
    NITRO_PORT: String(port),
    // Use the pass-through image provider for this CSS-only preview check so
    // the server does not need native sharp/libvips runtime binaries in CI.
    NUXT_IMAGE_PROVIDER: 'none',
    NUXT_HTML_CACHE_ENABLED: 'false',
  },
  stdio: ['ignore', 'pipe', 'pipe'],
  windowsHide: true,
})

const logs: string[] = []
const capture = (chunk: Buffer): void => {
  logs.push(String(chunk))
  if (logs.length > 20) logs.shift()
}
child.stdout.on('data', capture)
child.stderr.on('data', capture)

try {
  await waitForReady()

  const response = await fetch(`${origin}/`)
  const html = await response.text()
  if (response.status !== 200) {
    throw new Error(`Homepage returned HTTP ${response.status}`)
  }

  const stylesheets = findStylesheetHrefs(html)
  const forbiddenStylesheets = stylesheets.filter(href =>
    forbiddenStylesheetNames.some(name => href.includes(name)),
  )
  if (forbiddenStylesheets.length > 0) {
    throw new Error(`Deferred UI CSS was rendered as blocking: ${forbiddenStylesheets.join(', ')}`)
  }

  const criticalCssGzipBytes = stylesheets.reduce((total, href) => {
    const assetPath = getAssetPath(href)
    if (!existsSync(assetPath)) {
      throw new Error(`Missing critical stylesheet asset: ${href}`)
    }
    return total + gzipSync(readFileSync(assetPath)).byteLength
  }, 0)

  if (criticalCssGzipBytes > maxCriticalCssGzipBytes) {
    throw new Error(
      `Homepage critical CSS is ${(criticalCssGzipBytes / 1024).toFixed(2)} KiB gzip; `
      + `budget is ${(maxCriticalCssGzipBytes / 1024).toFixed(0)} KiB`,
    )
  }

  console.log(
    `[critical-css] OK: ${stylesheets.length} blocking stylesheet(s), `
    + `${(criticalCssGzipBytes / 1024).toFixed(2)} KiB gzip`,
  )
} catch (error: unknown) {
  console.error(`[critical-css] FAILED: ${error instanceof Error ? error.message : String(error)}`)
  if (logs.length > 0) {
    console.error('[critical-css] Recent server output:')
    console.error(logs.join('').trim())
  }
  process.exitCode = 1
} finally {
  child.kill('SIGTERM')
  setTimeout(() => {
    if (!child.killed) child.kill('SIGKILL')
  }, 1000).unref()
}

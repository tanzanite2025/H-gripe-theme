import { spawn } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { gzipSync } from 'node:zlib'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const serverEntry = resolve(projectRoot, '.output/server/index.mjs')
const serverLauncher = resolve(projectRoot, 'scripts/storefront/run-production-server.mjs')
const publicDirectory = resolve(projectRoot, '.output/public')
const port = Number.parseInt(process.env.CRITICAL_CSS_CHECK_PORT || '4021', 10)
const origin = `http://127.0.0.1:${port}`
const maxBlockingCssGzipBytes = 20 * 1024
const maxInlineCriticalCssGzipBytes = 16 * 1024
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

const stripNoscriptContent = (html: string): string => html.replace(/<noscript\b[^>]*>[\s\S]*?<\/noscript>/gi, '')

const findStylesheetHrefs = (html: string): string[] => {
  const linkTags = stripNoscriptContent(html).match(/<link\b[^>]*>/gi) || []

  return linkTags
    .filter(tag => /\brel=(['"])stylesheet\1/i.test(tag))
    .map((tag) => tag.match(/\bhref=(['"])(.*?)\1/i)?.[2] || '')
    .filter(Boolean)
}

const findInlineCriticalCss = (html: string): string => (
  html.match(/<style\b[^>]*\bdata-storefront-critical-css=(['"])entry\1[^>]*>([\s\S]*?)<\/style>/i)?.[2] || ''
)

const hasEntryCssPreload = (html: string): boolean => (
  /<link\b(?=[^>]*\brel=(['"])preload\1)(?=[^>]*\bas=(['"])style\2)(?=[^>]*\bhref=(['"])\/_nuxt\/entry\.[^"']+\.css\3)[^>]*>/i.test(html)
)

const hasEntryCssLoader = (html: string): boolean => (
  /<script\b[^>]*\bdata-storefront-critical-css-loader=(['"])entry\1[^>]*>/i.test(html)
)

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
if (!existsSync(serverLauncher)) {
  console.error(`[critical-css] FAILED: Missing ${serverLauncher}.`)
  process.exit(1)
}

const child = spawn(process.execPath, [serverLauncher], {
  cwd: projectRoot,
  env: {
    ...process.env,
    NODE_ENV: 'production',
    HOST: '127.0.0.1',
    PORT: String(port),
    NITRO_PORT: String(port),
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
  const blockingEntryStylesheets = stylesheets.filter(href => new URL(href, origin).pathname.startsWith('/_nuxt/entry.'))
  if (blockingEntryStylesheets.length > 0) {
    throw new Error(`Homepage entry CSS is still blocking: ${blockingEntryStylesheets.join(', ')}`)
  }

  const inlineCriticalCss = findInlineCriticalCss(html)
  if (!inlineCriticalCss) {
    throw new Error('Homepage entry critical CSS was not inlined')
  }
  if (!/@font-face\b/i.test(inlineCriticalCss)) {
    throw new Error('Homepage inline critical CSS is missing font-face declarations')
  }
  if (!hasEntryCssPreload(html)) {
    throw new Error('Homepage entry CSS preload was not rendered')
  }
  if (!hasEntryCssLoader(html)) {
    throw new Error('Homepage entry CSS async loader was not rendered')
  }

  const inlineCriticalCssGzipBytes = gzipSync(inlineCriticalCss).byteLength
  if (inlineCriticalCssGzipBytes > maxInlineCriticalCssGzipBytes) {
    throw new Error(
      `Homepage inline critical CSS is ${(inlineCriticalCssGzipBytes / 1024).toFixed(2)} KiB gzip; `
      + `budget is ${(maxInlineCriticalCssGzipBytes / 1024).toFixed(0)} KiB`,
    )
  }

  const forbiddenStylesheets = stylesheets.filter(href =>
    forbiddenStylesheetNames.some(name => href.includes(name)),
  )
  if (forbiddenStylesheets.length > 0) {
    throw new Error(`Deferred UI CSS was rendered as blocking: ${forbiddenStylesheets.join(', ')}`)
  }

  const blockingCssGzipBytes = stylesheets.reduce((total, href) => {
    const assetPath = getAssetPath(href)
    if (!existsSync(assetPath)) {
      throw new Error(`Missing critical stylesheet asset: ${href}`)
    }
    return total + gzipSync(readFileSync(assetPath)).byteLength
  }, 0)

  if (blockingCssGzipBytes > maxBlockingCssGzipBytes) {
    throw new Error(
      `Homepage blocking CSS is ${(blockingCssGzipBytes / 1024).toFixed(2)} KiB gzip; `
      + `budget is ${(maxBlockingCssGzipBytes / 1024).toFixed(0)} KiB`,
    )
  }

  console.log(
    `[critical-css] OK: ${stylesheets.length} blocking stylesheet(s), `
    + `${(blockingCssGzipBytes / 1024).toFixed(2)} KiB blocking gzip, `
    + `${(inlineCriticalCssGzipBytes / 1024).toFixed(2)} KiB inline critical gzip`,
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

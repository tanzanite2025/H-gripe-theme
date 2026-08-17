import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { h } from 'vue'
import { renderToString } from 'vue/server-renderer'
import {
  renderRichText,
  safeRichTextMediaOriginsFromRuntimeConfig,
} from '../../app/utils/security/safeRichText'
import {
  collectInlineContentHashes,
  collectResourceOrigins,
  createContentSecurityPolicy,
} from '../../server/security/content-security-policy'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '../../app/utils/storefrontMedia'

const projectRoot = resolve(import.meta.dirname, '../..')
const appRoot = resolve(projectRoot, 'app')
const failures: string[] = []

const assert = (condition: unknown, message: string): void => {
  if (!condition) failures.push(message)
}

const mediaContext = createStorefrontMediaContext({
  apiInternalOrigin: 'http://localhost:9200',
  public: {
    siteUrl: 'https://storefront.example.test',
  },
})

assert(
  normalizeStorefrontMediaUrl('http://localhost:9200/uploads/a.png', mediaContext) === '/uploads/a.png',
  'Internal upload origin was not normalized to a first-party relative path',
)
assert(
  normalizeStorefrontMediaUrl('https://storefront.example.test/uploads/a.png', mediaContext) === '/uploads/a.png',
  'Known first-party upload origin was not normalized to a relative path',
)
assert(
  normalizeStorefrontMediaUrl('https://unreviewed-csp-origin.invalid/uploads/a.png', mediaContext)
    === 'https://unreviewed-csp-origin.invalid/uploads/a.png',
  'Unknown third-party upload origin must remain unchanged',
)

const vueFiles = (directory: string): string[] => {
  const files: string[] = []
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = resolve(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...vueFiles(path))
    } else if (entry.isFile() && entry.name.endsWith('.vue')) {
      files.push(path)
    }
  }
  return files
}

for (const filePath of vueFiles(appRoot)) {
  const source = readFileSync(filePath, 'utf8')
  assert(!/\bv-html\s*=/.test(source), `v-html is not allowed: ${filePath}`)
}

const html = [
  '<!doctype html>',
  '<html><head>',
  '<style>body { color: #111; }</style>',
  '</head><body>',
  '<script>window.__securityCheck = true</script>',
  '<script type="application/ld+json">{"@context":"https://schema.org"}</script>',
  '<img src="https://unreviewed-csp-origin.invalid/asset.png" alt="">',
  '</body></html>',
].join('')
const policy = createContentSecurityPolicy(html)
const hashes = collectInlineContentHashes(html)
const resourceOrigins = collectResourceOrigins(html)

assert(policy.includes("require-trusted-types-for 'script'"), 'Trusted Types enforcement is missing')
assert(policy.includes('trusted-types vue tanzanite-script-url'), 'Trusted Types policy allowlist is missing')
assert(policy.includes("script-src-attr 'none'"), 'Inline event handlers are not blocked')
assert(!policy.includes("script-src 'self' 'unsafe-inline'"), 'Inline scripts must not be allowed')
assert(policy.includes("style-src-attr 'unsafe-inline'"), 'Vue-owned dynamic style attributes must remain explicitly scoped')

for (const hash of [...hashes.script, ...hashes.style]) {
  assert(policy.includes(hash), `Policy does not contain expected inline content hash: ${hash}`)
}
assert(
  resourceOrigins.image.includes('https://unreviewed-csp-origin.invalid'),
  'Final HTML resource-origin audit did not detect the unreviewed image origin',
)
assert(
  !policy.includes('https://unreviewed-csp-origin.invalid'),
  'Final HTML resource origins must not automatically widen the CSP',
)

assert(existsSync(resolve(projectRoot, 'app/utils/security/trustedScriptUrl.ts')), 'Trusted script URL policy helper is missing')

const richTextOrigins = safeRichTextMediaOriginsFromRuntimeConfig({
  public: {
    apiBase: 'https://api.example.test/api/v1',
    siteUrl: 'https://storefront.example.test/shop',
  },
})

assert(richTextOrigins.includes('https://api.example.test'), 'Rich-text media origin policy missed apiBase')
assert(richTextOrigins.includes('https://storefront.example.test'), 'Rich-text media origin policy missed siteUrl')

const richTextHtml = await renderToString(h('div', {}, renderRichText([
  '<p onclick="alert(1)">Hello <strong>world</strong></p>',
  '<script>alert(1)</script>',
  '<img src="/uploads/good.png" alt="good" style="width:999px">',
  '<img src="data:image/png;base64,AAAA" alt="inline">',
  '<img src="data:image/svg+xml;base64,AAAA" alt="svg">',
  '<img src="https://storefront.example.test/good.png" alt="storefront">',
  '<img src="https://evil.example.test/bad.png" alt="bad">',
  '<video src="https://api.example.test/good.mp4" poster="https://evil.example.test/poster.jpg" controls></video>',
  '<a href="javascript:alert(1)" target="_blank">bad link</a>',
  '<a href="https://external.example.test/page" target="_blank">external link</a>',
].join(''), {
  mediaOrigins: richTextOrigins,
})))

assert(richTextHtml.includes('<strong>world</strong>'), 'Safe rich text removed allowed formatting')
assert(richTextHtml.includes('src="/uploads/good.png"'), 'Safe rich text removed a relative upload image')
assert(richTextHtml.includes('src="data:image/png;base64,AAAA"'), 'Safe rich text removed a raster data image')
assert(richTextHtml.includes('src="https://storefront.example.test/good.png"'), 'Safe rich text removed an approved storefront image')
assert(richTextHtml.includes('src="https://api.example.test/good.mp4"'), 'Safe rich text removed an approved API media URL')
assert(richTextHtml.includes('href="https://external.example.test/page"'), 'Safe rich text removed an allowed external link')
assert(richTextHtml.includes('rel="noopener noreferrer"'), 'Safe rich text did not protect target=_blank links')
assert(!richTextHtml.includes('onclick'), 'Safe rich text preserved an inline event handler')
assert(!richTextHtml.includes('<script'), 'Safe rich text preserved a script element')
assert(!richTextHtml.includes('javascript:'), 'Safe rich text preserved a javascript URL')
assert(!richTextHtml.includes('data:image/svg+xml'), 'Safe rich text allowed an SVG data image')
assert(!richTextHtml.includes('https://evil.example.test'), 'Safe rich text allowed an unapproved media origin')

if (failures.length > 0) {
  console.error('[csp-check] FAILED')
  for (const failure of failures) {
    console.error(`- ${failure}`)
  }
  process.exit(1)
}

console.log('[csp-check] OK: CSP hashes, Trusted Types, and safe rich-text boundaries are enforced.')

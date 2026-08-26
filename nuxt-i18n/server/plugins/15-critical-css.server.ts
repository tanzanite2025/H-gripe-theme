import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineNitroPlugin } from 'nitropack/runtime'

interface RenderResponse {
  body?: unknown
  headers?: Record<string, string>
}

const CRITICAL_CSS_FILENAME = 'home-entry.css'
const ENTRY_STYLESHEET_PATTERN = /<link\b(?=[^>]*\brel=(["'])stylesheet\1)(?=[^>]*\bhref=(["'])(\/_nuxt\/entry\.[^"']+\.css)\2)[^>]*>/i

const isTruthyEnv = (value: string | undefined): boolean => ['1', 'true', 'yes', 'on'].includes(
  String(value || '').trim().toLowerCase(),
)

const escapeHtmlAttribute = (value: string): string => (
  value
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
)

const escapeStyleText = (value: string): string => value.replace(/<\/style/gi, '<\\/style')
const escapeScriptText = (value: string): string => value.replace(/<\/script/gi, '<\\/script')

const loadCriticalCss = (): string => {
  if (isTruthyEnv(process.env.NUXT_CRITICAL_CSS_DISABLED)) return ''

  const currentDirectory = dirname(fileURLToPath(import.meta.url))
  const candidates = [
    process.env.NUXT_CRITICAL_CSS_HOME_PATH || '',
    resolve(process.cwd(), '.output/server/critical-css', CRITICAL_CSS_FILENAME),
    resolve(process.cwd(), 'server/critical-css', CRITICAL_CSS_FILENAME),
    resolve(currentDirectory, '../../critical-css', CRITICAL_CSS_FILENAME),
    resolve(currentDirectory, '../critical-css', CRITICAL_CSS_FILENAME),
  ].filter(Boolean)

  for (const candidate of candidates) {
    if (!existsSync(candidate)) continue
    return readFileSync(candidate, 'utf8').trim()
  }

  return ''
}

const shouldTransformHtml = (html: string): boolean => (
  html.includes('<html') &&
  html.includes('home-hero') &&
  html.includes('site-header-root') &&
  !html.includes('data-storefront-critical-css="entry"') &&
  ENTRY_STYLESHEET_PATTERN.test(html)
)

const buildAsyncEntryCssMarkup = (linkTag: string, href: string, criticalCss: string): string => {
  const escapedHref = escapeHtmlAttribute(href)
  const crossorigin = /\bcrossorigin(?:=(["']).*?\1)?/i.test(linkTag) ? ' crossorigin=""' : ''
  const integrity = linkTag.match(/\bintegrity=(["'])(.*?)\1/i)?.[2] || ''
  const integrityAttribute = integrity ? ` integrity="${escapeHtmlAttribute(integrity)}"` : ''
  const stylesheetAttributes = `href="${escapedHref}"${crossorigin}${integrityAttribute}`
  const loaderScript = `(() => {
  const href = ${JSON.stringify(href)};
  const absoluteHref = new URL(href, document.baseURI).href;
  const links = document.getElementsByTagName('link');
  for (let index = 0; index < links.length; index += 1) {
    const link = links[index];
    if (link.rel === 'stylesheet' && link.href === absoluteHref) return;
  }
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = href;
  ${crossorigin ? "link.crossOrigin = '';" : ''}
  link.setAttribute('data-storefront-full-css', 'entry');
  document.head.appendChild(link);
})();`

  return [
    `<style data-storefront-critical-css="entry">${escapeStyleText(criticalCss)}</style>`,
    `<link rel="preload" as="style" ${stylesheetAttributes} data-storefront-entry-css="preload">`,
    `<noscript><link rel="stylesheet" ${stylesheetAttributes}></noscript>`,
    `<script data-storefront-critical-css-loader="entry">${escapeScriptText(loaderScript)}</script>`,
  ].join('')
}

export default defineNitroPlugin((nitroApp) => {
  const criticalCss = loadCriticalCss()
  if (!criticalCss) return

  nitroApp.hooks.hook('render:response', (response: RenderResponse) => {
    if (typeof response.body !== 'string' || !shouldTransformHtml(response.body)) return

    response.body = response.body.replace(ENTRY_STYLESHEET_PATTERN, (linkTag: string, _relQuote: string, _hrefQuote: string, href: string) => (
      buildAsyncEntryCssMarkup(linkTag, href, criticalCss)
    ))
  })
})

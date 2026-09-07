import assert from 'node:assert/strict'
import { prioritizeLatinFontPreload } from '../../server/utils/fontPreloadPriority.ts'

const latinPreload = '<link rel="preload" href="/fonts/MapleUI-Latin.00af3fec5b34.woff2" as="font" type="font/woff2" crossorigin="anonymous" data-hid="storefront-font-preload-latin">'
const stylesheet = '<link rel="stylesheet" href="/_nuxt/entry.css">'

const html = [
  '<!doctype html><html><head>',
  '<meta charset="utf-8">',
  stylesheet,
  '<link rel="preload" as="font" type="font/woff2" href="/fonts/MapleUI-Latin.00af3fec5b34.woff2" crossorigin="anonymous">',
  '</head><body><main>Home</main></body></html>',
].join('')

const prioritized = prioritizeLatinFontPreload(html)
const head = prioritized.split(/<head>|<\/head>/)[1]

assert.ok(head.startsWith(latinPreload), 'Latin preload should be the first SSR head child.')
assert.ok(head.indexOf(latinPreload) < head.indexOf(stylesheet), 'Latin preload should appear before generated CSS.')
assert.equal([...head.matchAll(/MapleUI-Latin\.00af3fec5b34\.woff2/g)].length, 1, 'Latin preload should not be duplicated.')

const nonHtml = '{"status":"ok"}'
assert.equal(prioritizeLatinFontPreload(nonHtml), nonHtml)

console.log('Font preload priority edge-case checks passed.')

const LATIN_FONT_PRELOAD_HREF = '/fonts/MapleUI-Latin.00af3fec5b34.woff2'
const LATIN_FONT_PRELOAD_LINK = `<link rel="preload" href="${LATIN_FONT_PRELOAD_HREF}" as="font" type="font/woff2" crossorigin="anonymous" data-hid="storefront-font-preload-latin">`
const LATIN_FONT_PRELOAD_PATTERN = /<link\b(?=[^>]*\brel=(["'])preload\1)(?=[^>]*\bas=(["'])font\2)(?=[^>]*\bhref=(["'])\/fonts\/MapleUI-Latin\.00af3fec5b34\.woff2\3)[^>]*>/gi
const HEAD_OPEN_PATTERN = /<head[^>]*>/i

export const prioritizeLatinFontPreload = (html: string): string => {
  if (!html.includes('<html') || !html.includes('</head>')) return html

  const withoutExistingPreload = html.replace(LATIN_FONT_PRELOAD_PATTERN, '')
  const headMatch = withoutExistingPreload.match(HEAD_OPEN_PATTERN)
  if (!headMatch || headMatch.index === undefined) return html

  const insertionIndex = headMatch.index + headMatch[0].length
  return `${withoutExistingPreload.slice(0, insertionIndex)}${LATIN_FONT_PRELOAD_LINK}${withoutExistingPreload.slice(insertionIndex)}`
}

import { createError, defineEventHandler, getRequestURL, setHeader } from 'h3'

const exactProbePaths = new Set([
  '/products.json',
  '/products.js',
  '/products/count.json',
  '/collections.json',
  '/search/suggest.json',
  '/recommendations/products.json',
  '/cart.js',
  '/cart/add.js',
  '/cart/update.js',
  '/cart/change.js',
  '/cart/clear.js',
])

const probePathPatterns = [
  /^\/collections\/[^/]+\.json$/i,
  /^\/collections\/[^/]+\/products\.json$/i,
  /^\/wc-ajax(?:\/|$)/i,
  /^\/wp-json\/(?:wc(?:\/store)?|wp\/v2)(?:\/|$)/i,
]

const isCommerceProbePath = (pathname: string) => {
  const normalized = pathname.replace(/\/+$/, '') || '/'
  if (exactProbePaths.has(normalized.toLowerCase())) return true
  return probePathPatterns.some(pattern => pattern.test(normalized))
}

export default defineEventHandler((event) => {
  const pathname = getRequestURL(event).pathname
  if (!isCommerceProbePath(pathname)) return

  setHeader(event, 'Cache-Control', 'no-store, max-age=0')
  setHeader(event, 'X-Robots-Tag', 'noindex, nofollow, noarchive')
  throw createError({
    statusCode: 404,
    statusMessage: 'Not Found',
  })
})

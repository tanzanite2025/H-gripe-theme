import { defineNitroPlugin } from 'nitropack/runtime'
import { getResponseHeader, setResponseHeader } from 'h3'
import { secureHtmlWithContentSecurityPolicy } from '../security/content-security-policy'
import { appendNoTransformDirectiveToHtmlCacheControl } from '../security/html-cache-control-proxy-transformation-protection'

interface RenderResponse {
  body?: unknown
  headers?: Record<string, string>
}

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:response', (response: RenderResponse) => {
    if (typeof response.body !== 'string' || !response.body.includes('<html')) return

    const secured = secureHtmlWithContentSecurityPolicy(response.body)
    response.headers = response.headers || {}
    response.body = secured.body
    response.headers['content-security-policy'] = secured.contentSecurityPolicy

    // Cloudflare JavaScript Detection rewrites HTML responses by injecting an
    // iframe bootstrap that assigns a raw string to innerHTML. That browser
    // sink is intentionally blocked by our Trusted Types policy. Cloudflare
    // documents that `no-transform` prevents this HTML rewrite, so preserve
    // the existing cache directives and add the narrowly scoped directive to
    // keep the response compatible with the site's CSP.
    response.headers['cache-control'] = appendNoTransformDirectiveToHtmlCacheControl(response.headers['cache-control'])
  })

  nitroApp.hooks.hook('beforeResponse', (event, response) => {
    const contentType = String(getResponseHeader(event, 'content-type') || '').toLowerCase()
    const hasHtmlBody = typeof response.body === 'string' && response.body.includes('<html')
    if (!contentType.includes('text/html') && !hasHtmlBody) return

    // Nitro's route-cache wrapper overwrites Cache-Control after render:response.
    // Apply the directive to the final headers for both cold renders and cache
    // hits, without changing s-maxage, stale-while-revalidate or no-store.
    setResponseHeader(event, 'cache-control', appendNoTransformDirectiveToHtmlCacheControl(
      getResponseHeader(event, 'cache-control'),
    ))
  })
})

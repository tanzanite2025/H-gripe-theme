import { defineNitroPlugin } from 'nitropack/runtime'
import { secureHtmlWithContentSecurityPolicy } from '../security/content-security-policy'

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
    const cacheControl = response.headers['cache-control']
    response.headers['cache-control'] = cacheControl
      ? `${cacheControl}, no-transform`
      : 'no-transform'
  })
})

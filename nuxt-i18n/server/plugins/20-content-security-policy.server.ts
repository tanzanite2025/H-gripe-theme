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
  })
})

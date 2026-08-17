import { defineNitroPlugin } from 'nitropack/runtime'
import { createContentSecurityPolicy } from '../security/content-security-policy'

interface RenderResponse {
  body?: unknown
  headers?: Record<string, string>
}

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:response', (response: RenderResponse) => {
    if (typeof response.body !== 'string' || !response.body.includes('<html')) return

    response.headers = response.headers || {}
    response.headers['content-security-policy'] = createContentSecurityPolicy(response.body)
  })
})

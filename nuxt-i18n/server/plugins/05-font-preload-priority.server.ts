import { defineNitroPlugin } from 'nitropack/runtime'
import { prioritizeLatinFontPreload } from '../utils/fontPreloadPriority'

interface RenderResponse {
  body?: unknown
}

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:response', (response: RenderResponse) => {
    if (typeof response.body !== 'string') return
    response.body = prioritizeLatinFontPreload(response.body)
  })
})

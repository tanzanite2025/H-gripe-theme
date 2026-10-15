import { defineNitroPlugin } from 'nitropack/runtime'
import { storefrontHeaderCriticalLayoutStyles } from '../utils/storefrontHeaderCriticalLayoutStyles'

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:html', (html) => {
    if (!html.body.join('').includes('site-header-root')) return
    if (html.head.some(markup => markup.includes('data-storefront-header-critical-layout'))) return

    // Low-specificity rules precede component CSS. The CSP response hook hashes
    // this style so the header's initial layout also works in production.
    html.head.unshift(`<style data-storefront-header-critical-layout>${storefrontHeaderCriticalLayoutStyles}</style>`)
  })
})

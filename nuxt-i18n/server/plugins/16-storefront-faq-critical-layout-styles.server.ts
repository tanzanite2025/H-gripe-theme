import { defineNitroPlugin } from 'nitropack/runtime'
import { storefrontFaqCriticalLayoutStyles } from '../utils/storefrontFaqCriticalLayoutStyles'

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:html', (html) => {
    const renderedBody = html.body.join('')
    if (!/class="[^"]*\b(?:page-faq|home-faq|faqs-page|storefront-faq-placeholder)\b/.test(renderedBody)) return
    if (html.head.some(markup => markup.includes('data-storefront-faq-critical-layout'))) return
    // Put the fallback before external styles. CSP hashes it in render:response;
    // it works in DEV, in cached production HTML, and without client JavaScript.
    html.head.unshift(`<style data-storefront-faq-critical-layout>${storefrontFaqCriticalLayoutStyles}</style>`)
  })
})

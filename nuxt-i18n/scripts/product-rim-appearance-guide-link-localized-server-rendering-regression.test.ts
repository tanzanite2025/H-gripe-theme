import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import vuePlugin from '@vitejs/plugin-vue'
import { renderToString } from '@vue/server-renderer'
import { createServer } from 'vite'
import { createSSRApp, defineComponent, h } from 'vue'
import { createI18n } from 'vue-i18n'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const compilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    {
      name: 'provide-locale-routing-for-product-rim-appearance-guide-link-rendering',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0product-rim-appearance-guide-link-nuxt-imports'
      },
      load(moduleId) {
        if (moduleId !== '\0product-rim-appearance-guide-link-nuxt-imports') return
        return `
          import { useI18n } from 'vue-i18n';
          export { useI18n };
          export const useLocalePath = () => {
            const { locale } = useI18n();
            return path => locale.value === 'en' ? path : '/' + locale.value + path;
          };
        `
      },
    },
    vuePlugin(),
  ],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

try {
  const { default: appearanceGuideLink } = await compilationServer.ssrLoadModule(
    '/app/components/shop/product-detail/ProductCarbonRimAppearanceGuideLink.vue',
  )
  for (const locale of ['en', 'zh_cn']) {
    const products = JSON.parse(await readFile(new URL(`../app/i18n/messages/${locale}/products.json`, import.meta.url), 'utf8'))
    for (const template of ['rim', 'carbon_rim', 'wheelset', 'spoke', 'hub', 'tire', '']) {
      const app = createSSRApp({ render: () => h(appearanceGuideLink, { productSpecificationTemplateSlug: template }) })
      app.use(createI18n({ legacy: false, locale, messages: { [locale]: { products } } }))
      app.component('NuxtLink', defineComponent({
        props: ['to'],
        setup(props, { slots }) { return () => h('a', { href: props.to }, slots.default?.()) },
      }))
      const markup = await renderToString(app)
      if (['rim', 'carbon_rim', 'wheelset'].includes(template)) {
        assert.ok(markup.includes(`href="${locale === 'en' ? '' : '/zh_cn'}/company/about/appearance"`))
        assert.ok(markup.includes(products.detail.rimAppearanceGuide.replaceAll('&', '&amp;')))
      } else {
        assert.ok(!markup.includes('/company/about/appearance'))
      }
    }
  }
  console.log('PASS localized appearance guide links for rims and wheelsets; unrelated product templates have no link')
} finally {
  await compilationServer.close()
}

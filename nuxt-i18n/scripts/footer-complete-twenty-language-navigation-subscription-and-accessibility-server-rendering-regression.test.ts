import assert from 'node:assert/strict'
import { readFile, readdir } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import vuePlugin from '@vitejs/plugin-vue'
import { renderToString } from '@vue/server-renderer'
import { createServer } from 'vite'
import { createSSRApp, defineComponent, h, type Component } from 'vue'
import { createI18n } from 'vue-i18n'
import type { RouteRecordNormalized } from 'vue-router'
import storefrontLocales from '../app/i18n/locales.manifest.js'
import { createFooterMenusFromRoutes, type FooterNavigationItem } from '../app/utils/footerMenus.js'

type TranslationResource = { [key: string]: string | TranslationResource }
const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const pageDirectory = path.join(frontendDirectory, 'app/pages')
const routes = [{ path: '/shop', name: 'shop', meta: {} }] as unknown as RouteRecordNormalized[]

for (const pageFile of await readdir(pageDirectory, { recursive: true })) {
  if (!pageFile.endsWith('.vue')) continue
  const source = await readFile(path.join(pageDirectory, pageFile), 'utf8')
  const footerLabelKey = source.match(/footerLabelKey:\s*['"]([^'"]+)['"]/)?.[1]
  const footerLabelFallback = source.match(/footerLabelFallback:\s*['"]([^'"]+)['"]/)?.[1]
  if (!footerLabelKey && !footerLabelFallback) continue
  const canonicalPath = source.match(/\bpath:\s*['"]([^'"]+)['"]/)?.[1]
    || '/' + pageFile.replaceAll('\\', '/').replace(/\.vue$/, '').replace(/\/index$/, '')
  routes.push({
    path: canonicalPath,
    name: canonicalPath,
    meta: { footerLabelKey, footerLabelFallback },
  } as unknown as RouteRecordNormalized)
}

const footerRenderingTestContext = {
  routes,
  nuxtApp: { $i18n: { mergeLocaleMessage: (_locale: string, _resource: TranslationResource): void => {} } },
}
const globalFooterRenderingTestContext = globalThis as typeof globalThis & {
  footerRenderingTestContext?: typeof footerRenderingTestContext
}
globalFooterRenderingTestContext.footerRenderingTestContext = footerRenderingTestContext

const compilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    vuePlugin(),
    {
      name: 'provide-isolated-storefront-context-for-complete-footer-localization-tests',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0complete-footer-test-imports'
        const modulePath = source.replaceAll('\\', '/').replace(/\.ts$/, '')
        if (modulePath.endsWith('/composables/useApiRequest')) return '\0complete-footer-test-api'
        if (modulePath.endsWith('/composables/useProductCategories')) return '\0complete-footer-test-categories'
        if (modulePath.endsWith('/composables/usePublicSettings')) return '\0complete-footer-test-settings'
        if (modulePath.endsWith('/composables/useSocialLinks')) return '\0complete-footer-test-social-links'
      },
      load(moduleId) {
        switch (moduleId) {
          case '\0complete-footer-test-imports': return `
            export { useI18n } from 'vue-i18n';
            export const useLocalePath = () => to => to;
            export const useRouter = () => ({ getRoutes: () => globalThis.footerRenderingTestContext.routes });
            export const useNuxtApp = () => globalThis.footerRenderingTestContext.nuxtApp;
            export const useRuntimeConfig = () => ({ public: {} });
          `
          case '\0complete-footer-test-api': return `
            export const useApiRequest = () => ({ request: () => { throw new Error('Footer SSR must not submit a subscription'); } });
          `
          case '\0complete-footer-test-categories': return `
            import { ref } from 'vue';
            export const useProductCategories = () => ({
              tree: ref([
                { name: 'Wheelsets', slug: 'wheelset' },
                { name: 'Wheel Components', slug: 'wheel-components', children: [
                  { name: 'Tires', slug: 'tire' },
                  { name: 'Spoke Repair Kits', slug: 'spoke-repair-kits' },
                  { name: 'Carbon Rim', slug: 'carbon-rim' },
                ] },
              ]),
              loadCategories: async () => [],
            });
          `
          case '\0complete-footer-test-settings': return `
            import { ref } from 'vue';
            export const useSiteSettings = () => ({ siteSettings: ref({}) });
          `
          case '\0complete-footer-test-social-links': return `
            import { ref } from 'vue';
            export const useSocialLinks = () => ({ socialLinks: ref([]) });
          `
        }
      },
    },
  ],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

function flattenFooterNavigationTranslationKeys(items: readonly FooterNavigationItem[]): string[] {
  return items.flatMap(item => [
    ...(item.labelKey ? [item.labelKey] : []),
    ...flattenFooterNavigationTranslationKeys(item.children || []),
  ])
}

function flattenFooterTranslationResource(resource: TranslationResource, prefix = ''): Record<string, string> {
  return Object.fromEntries(Object.entries(resource).flatMap(([key, value]) => {
    const fullKey = prefix ? `${prefix}.${key}` : key
    return typeof value === 'string'
      ? [[fullKey, value]]
      : Object.entries(flattenFooterTranslationResource(value, fullKey))
  }))
}

function escapeFooterTranslationForHtml(value: string): string {
  return value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;').replaceAll("'", '&#39;')
}

try {
  const footerComponent = (await compilationServer.ssrLoadModule('/app/components/AppFooter.vue')).default as Component
  const englishMessages = JSON.parse(await readFile(path.join(frontendDirectory, 'app/i18n/locales/en.json'), 'utf8')) as TranslationResource
  const englishLabels = flattenFooterTranslationResource(englishMessages)
  const menuSections = createFooterMenusFromRoutes(routes)
  assert.equal(menuSections.length, 9, 'Every footer section, including the site overview, must be covered')
  const navigationKeys = menuSections.flatMap(section => [section.titleKey, ...flattenFooterNavigationTranslationKeys(section.links)])
  const sharedBlogLocales = new Set(['fr', 'de', 'es', 'it', 'nl', 'tr', 'id', 'da', 'ms'])
  const renderedFooterKeys = [
    'footer.subscription.label', 'footer.subscription.placeholder', 'footer.subscription.emailLabel',
    'footer.subscription.button', 'footer.securityNotice', 'footer.accessibility.navigation',
    'footer.accessibility.paymentMethods', 'footer.accessibility.sslSecure', 'footer.accessibility.cardPayment',
    ...['wheelset', 'wheelComponents', 'tire', 'spokeRepairKits', 'carbonRim'].map(key => `footer.categories.${key}`),
  ]

  for (const { code: locale } of storefrontLocales) {
    const messages = JSON.parse(await readFile(path.join(frontendDirectory, `app/i18n/locales/${locale}.json`), 'utf8')) as TranslationResource
    const labels = flattenFooterTranslationResource(messages)
    assert.deepEqual(Object.keys(flattenFooterTranslationResource(messages.footer as TranslationResource)).sort(),
      Object.keys(flattenFooterTranslationResource(englishMessages.footer as TranslationResource)).sort(),
      `${locale}: footer keys must be complete`)
    for (const [key, label] of Object.entries(labels)) {
      if (!key.startsWith('footer.') && !navigationKeys.includes(key)) continue
      assert.ok(label.trim(), `${locale}: ${key} must not be empty`)
      if (locale === 'en' || key.startsWith('footer.brand.')) continue
      const isEstablishedSharedTerm = label === 'Blog' && sharedBlogLocales.has(locale)
        || label === 'Shop' && locale === 'de'
        || label === 'Guides' && locale === 'fr'
      if (!isEstablishedSharedTerm) {
        assert.notEqual(label, englishLabels[key], `${locale}: ${key} must not contain an English placeholder`)
      }
    }

    const missingKeys: string[] = []
    const i18n = createI18n({
      legacy: false, locale, fallbackLocale: false, messages: { [locale]: messages },
      missing: (_locale, key) => { missingKeys.push(key) },
    })
    footerRenderingTestContext.nuxtApp.$i18n.mergeLocaleMessage = (requestedLocale, resource) => {
      assert.equal(requestedLocale, locale)
      i18n.global.mergeLocaleMessage(requestedLocale, resource)
    }
    const application = createSSRApp({ render: () => h(footerComponent) })
    application.use(i18n)
    application.component('NuxtLink', defineComponent({
      props: ['to'], setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
    }))
    application.component('Icon', defineComponent({ setup: () => () => h('svg', { 'aria-hidden': 'true' }) }))
    application.component('ClientOnly', defineComponent({ setup: () => () => null }))
    const renderErrors: unknown[] = []
    application.config.errorHandler = error => { renderErrors.push(error) }
    const markup = await renderToString(application)
    assert.deepEqual(renderErrors, [], `${locale}: complete footer must render successfully`)
    assert.deepEqual(missingKeys, [], `${locale}: complete footer must not use missing translation keys`)
    for (const key of [...navigationKeys, ...renderedFooterKeys]) {
      const label = i18n.global.t(key)
      assert.notEqual(label, key, `${locale}: ${key} must resolve`)
      assert.ok(markup.includes(escapeFooterTranslationForHtml(label)), `${locale}: initial footer HTML must contain ${key}`)
    }
    if (locale !== 'en') {
      assert.ok(!markup.includes('Subscribe for new products &amp; blog updates'), `${locale}: subscription copy must use the selected language`)
    }
  }
  console.log('Complete footer localization passed for all 20 languages: menu and guide links, system categories, subscription fields, security copy, and accessible labels render without missing keys or English placeholders.')
} finally {
  await compilationServer.close()
  delete globalFooterRenderingTestContext.footerRenderingTestContext
}

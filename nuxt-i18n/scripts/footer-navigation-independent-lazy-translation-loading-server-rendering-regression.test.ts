import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import type { LocaleMessage } from '@intlify/core-base'
import vuePlugin from '@vitejs/plugin-vue'
import { renderToString } from '@vue/server-renderer'
import { createServer } from 'vite'
import { createSSRApp, defineComponent, h, type Component } from 'vue'
import { createI18n } from 'vue-i18n'
import type { FooterNavigationItem, FooterSection } from '../app/utils/footerMenus.js'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const footerTranslationTestContext = {
  nuxtApp: {
    $i18n: {
      mergeLocaleMessage: (_locale: string, _messages: Record<string, unknown>): void => {},
    },
  },
}
const globalFooterTranslationTestContext = globalThis as typeof globalThis & {
  footerTranslationTestContext?: typeof footerTranslationTestContext
}
globalFooterTranslationTestContext.footerTranslationTestContext = footerTranslationTestContext

const componentCompilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    vuePlugin(),
    {
      name: 'provide-isolated-nuxt-context-for-footer-translation-tests',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0footer-translation-test-nuxt-imports'
      },
      load(moduleId) {
        if (moduleId === '\0footer-translation-test-nuxt-imports') {
          return `
            export { useI18n } from 'vue-i18n';
            export const useLocalePath = () => to => to;
            export const useRouter = () => ({ getRoutes: () => [] });
            export const useNuxtApp = () => globalThis.footerTranslationTestContext.nuxtApp;
          `
        }
      },
    },
  ],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

const flattenFooterNavigationItemsForTranslationVerification = (
  items: readonly FooterNavigationItem[],
): FooterNavigationItem[] => items.flatMap(item => [
  item,
  ...flattenFooterNavigationItemsForTranslationVerification(item.children || []),
])

try {
  const footerMenusComponent = (await componentCompilationServer.ssrLoadModule(
    '/app/components/FooterMenus.vue',
  )).default as Component
  const footerNavigationItemComponent = (await componentCompilationServer.ssrLoadModule(
    '/app/components/FooterMenuNavigationItem.vue',
  )).default as Component
  const { createFooterMenusFromRoutes } = await componentCompilationServer.ssrLoadModule(
    '/app/utils/footerMenus.ts',
  )
  const menuSections: FooterSection[] = createFooterMenusFromRoutes([
    { path: '/guides/tireguides', name: 'tireguides', meta: {} },
    { path: '/guides/wheelset-buyers', name: 'wheelset-buyers', meta: {} },
  ])
  const guideItems = flattenFooterNavigationItemsForTranslationVerification(
    menuSections.flatMap(section => section.links),
  ).filter(item => item.labelKey?.startsWith('guides'))
  assert.equal(guideItems.length, 7, 'All tire and wheelset guide child links must be exercised')

  for (const locale of ['zh_cn', 'en', 'fr']) {
    const baseMessages = JSON.parse(await readFile(
      path.join(frontendDirectory, `app/i18n/locales/${locale}.json`), 'utf8',
    )) as LocaleMessage<string>
    const missingTranslationKeys: string[] = []
    const applicationTranslations = createI18n({
      legacy: false,
      locale,
      fallbackLocale: false,
      messages: { [locale]: baseMessages },
      missing: (_locale, key) => { missingTranslationKeys.push(key) },
    })
    const loadedNamespaces: string[] = []
    footerTranslationTestContext.nuxtApp.$i18n.mergeLocaleMessage = (requestedLocale, messages) => {
      assert.equal(requestedLocale, locale)
      loadedNamespaces.push(...Object.keys(messages))
      applicationTranslations.global.mergeLocaleMessage(requestedLocale, messages)
    }
    assert.equal(applicationTranslations.global.te(guideItems[0]!.labelKey!), false)

    // Render the recursive item before any page translations have loaded. This
    // is also the state seen briefly when an already open menu changes locale.
    const fallbackApplication = createSSRApp({
      render: () => h(footerNavigationItemComponent, { item: guideItems[0] }),
    })
    fallbackApplication.use(applicationTranslations)
    fallbackApplication.component('NuxtLink', defineComponent({
      props: ['to'],
      setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
    }))
    const fallbackMarkup = await renderToString(fallbackApplication)
    assert.ok(fallbackMarkup.includes(guideItems[0]!.fallback!))
    assert.deepEqual(missingTranslationKeys, [], 'Pending item translations must use the fallback without warnings')

    // A fresh footer on an unrelated page has only global messages. Use the
    // real lazy JSON loader so no sibling header or page can supply its copy.
    const application = createSSRApp({
      render: () => h(footerMenusComponent, { menus: menuSections }),
    })
    application.use(applicationTranslations)
    application.component('NuxtLink', defineComponent({
      props: ['to'],
      setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
    }))
    const renderErrors: unknown[] = []
    application.config.errorHandler = error => { renderErrors.push(error) }
    const markup = await renderToString(application)
    assert.deepEqual(renderErrors, [])
    assert.deepEqual(loadedNamespaces.sort(), ['guidesTireguides', 'guidesWheelsetBuyers'])
    assert.deepEqual(missingTranslationKeys, [], 'Fresh footer SSR must not produce missing-key warnings')
    for (const item of guideItems) {
      const translatedLabel = applicationTranslations.global.t(item.labelKey!)
      assert.notEqual(translatedLabel, item.labelKey)
      const escapedLabel = translatedLabel.replaceAll('&', '&amp;').replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;').replaceAll('"', '&quot;').replaceAll("'", '&#39;')
      assert.ok(markup.includes(escapedLabel), `Initial footer HTML must contain ${translatedLabel}`)
      assert.ok(markup.includes(`href="${item.to}"`), 'Translated child links must keep their canonical routes')
    }
    assert.ok(markup.includes(locale === 'zh_cn' ? '购买概览' : 'Buying overview'))
  }
  console.log('Footer independent translation loading passed: Chinese, English, English page-copy fallback, and unloaded recursive item labels render without missing-key events.')
} finally {
  await componentCompilationServer.close()
  delete globalFooterTranslationTestContext.footerTranslationTestContext
}

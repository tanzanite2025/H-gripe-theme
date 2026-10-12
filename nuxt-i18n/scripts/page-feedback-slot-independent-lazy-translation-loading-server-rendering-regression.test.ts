import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import vuePlugin from '@vitejs/plugin-vue'
import { renderToString } from '@vue/server-renderer'
import { createServer } from 'vite'
import { createSSRApp, h, type Component } from 'vue'
import { createI18n } from 'vue-i18n'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const pageFeedbackTranslationTestContext = {
  route: {
    path: '/zh_cn/tools/tire-rim-width-compatibility-engine',
    meta: {
      feedbackThreadKey: 'guides-tireguides',
      feedbackTitleKey: 'guidesTireguides.feedbackTitle',
      feedbackSubtitleKey: 'feedback.defaultSubtitle',
      pageTitleKey: 'guidesTireguides.tabs.choose.label',
      pageTitle: 'Tire and Rim Width Calculator',
    },
  },
  loadedNamespaces: [] as string[],
  loadMessages: async (_namespace: string, _locale: string): Promise<void> => {},
}
const globalTranslationTestContext = globalThis as typeof globalThis & {
  pageFeedbackTranslationTestContext?: typeof pageFeedbackTranslationTestContext
}
globalTranslationTestContext.pageFeedbackTranslationTestContext = pageFeedbackTranslationTestContext

const componentCompilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    vuePlugin(),
    {
      name: 'provide-isolated-nuxt-context-for-page-feedback-translation-tests',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0page-feedback-test-nuxt-imports'
        const normalizedSource = source.replaceAll('\\', '/')
        if (normalizedSource === '~/composables/usePageMessages'
          || /\/app\/composables\/usePageMessages(?:\.ts)?$/.test(normalizedSource)) {
          return '\0page-feedback-test-delayed-message-loader'
        }
        if (normalizedSource === '~/components/UserFeedbackThread.vue'
          || /\/app\/components\/UserFeedbackThread\.vue$/.test(normalizedSource)) {
          return '\0page-feedback-test-visible-feedback-text'
        }
      },
      load(moduleId) {
        if (moduleId === '\0page-feedback-test-nuxt-imports') {
          return 'export { useI18n } from "vue-i18n"; export const useRoute = () => globalThis.pageFeedbackTranslationTestContext.route'
        }
        if (moduleId === '\0page-feedback-test-delayed-message-loader') {
          return `
            export const pageMessageNamespaces = ['guidesTireguides'];
            export const usePageMessages = namespace => ({
              loadPageMessages: locale => globalThis.pageFeedbackTranslationTestContext.loadMessages(namespace, locale),
            });
          `
        }
        if (moduleId === '\0page-feedback-test-visible-feedback-text') {
          return `
            import { defineComponent, h } from 'vue';
            export default defineComponent({
              props: ['title', 'subtitle'],
              setup: props => () => h('article', [h('h2', props.title), h('p', props.subtitle)]),
            });
          `
        }
      },
    },
  ],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

try {
  const feedbackSlotComponent = (await componentCompilationServer.ssrLoadModule(
    '/app/components/PageFeedbackSlot.vue',
  )).default as Component

  for (const locale of ['zh_cn', 'en']) {
    // A fresh app deliberately has no page messages. This reproduces a layout
    // slot hydrating before its async sibling page has loaded translations.
    const pageMessages = JSON.parse(await readFile(
      path.join(frontendDirectory, `app/i18n/page-messages/guidesTireguides/${locale}.json`),
      'utf8',
    ))
    const applicationTranslations = createI18n({
      legacy: false, locale, messages: {}, missingWarn: false, fallbackWarn: false,
    })
    pageFeedbackTranslationTestContext.route.path = `${locale === 'en' ? '' : '/zh_cn'}/tools/tire-rim-width-compatibility-engine`
    pageFeedbackTranslationTestContext.loadedNamespaces = []
    pageFeedbackTranslationTestContext.loadMessages = async (namespace, requestedLocale) => {
      await new Promise(resolveDelay => setTimeout(resolveDelay, 25))
      assert.equal(requestedLocale, locale)
      applicationTranslations.global.mergeLocaleMessage(locale, { [namespace]: pageMessages })
      pageFeedbackTranslationTestContext.loadedNamespaces.push(namespace)
    }
    const application = createSSRApp({ render: () => h(feedbackSlotComponent) })
    application.use(applicationTranslations)
    const renderErrors: unknown[] = []
    application.config.errorHandler = error => { renderErrors.push(error) }
    const markup = await renderToString(application)
    assert.deepEqual(renderErrors, [])
    assert.deepEqual(pageFeedbackTranslationTestContext.loadedNamespaces, ['guidesTireguides'])
    assert.ok(markup.includes(pageMessages.feedbackTitle), 'Initial slot render must use translated feedback copy')
    assert.ok(!markup.includes('<h2>Tire and Rim Width Calculator</h2>'), 'Initial slot render must not use the page-title fallback')
  }
  console.log('Page feedback independent translation loading passed: fresh English and Chinese slot renders await their own lazy messages.')
} finally {
  await componentCompilationServer.close()
  delete globalTranslationTestContext.pageFeedbackTranslationTestContext
}

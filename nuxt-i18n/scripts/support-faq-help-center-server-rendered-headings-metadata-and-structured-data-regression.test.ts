import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import vuePlugin from '@vitejs/plugin-vue'
import { createHead } from '@unhead/vue/server'
import { renderToString } from '@vue/server-renderer'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import { renderSSRHead } from 'unhead/server'
import { createServer } from 'vite'
import { createSSRApp, defineComponent, h, type Component } from 'vue'
import { createI18n } from 'vue-i18n'
import type { ResolvedPageFaqData } from '../app/data/faq/helpers.ts'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const faqHelpCenterRenderingContext = {
  route: { path: '/support/faqs', query: {} as Record<string, string> },
  pages: [] as ResolvedPageFaqData[],
  loadPageMessages: async (_namespace: string, _locale: string): Promise<void> => {},
}
const renderingGlobals = globalThis as typeof globalThis & {
  faqHelpCenterRenderingContext?: typeof faqHelpCenterRenderingContext
}
renderingGlobals.faqHelpCenterRenderingContext = faqHelpCenterRenderingContext

// Compile the real page, grouping/filtering composables, panels, and rich-text
// renderer. Only Nuxt's route, lazy translations, and backend catalog are supplied.
const compilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    {
      name: 'provide-isolated-nuxt-context-for-faq-help-center-server-rendering',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0faq-help-center-test-nuxt-imports'
        const normalizedSource = source.replaceAll('\\', '/').replace(/\.ts$/, '')
        if (normalizedSource === '~/data/faq' || normalizedSource.endsWith('/app/data/faq')
          || normalizedSource.endsWith('/app/data/faq/index')) {
          return '\0faq-help-center-test-catalog-data-exports'
        }
        for (const composable of ['useFaqCatalog', 'usePageMessages', 'seo/useStorefrontSeoLinks']) {
          if (normalizedSource === `~/composables/${composable}`
            || normalizedSource.endsWith(`/app/composables/${composable}`)) {
            return `\0faq-help-center-test-${composable}`
          }
        }
      },
      load(moduleId) {
        if (moduleId === '\0faq-help-center-test-catalog-data-exports') {
          return "export * from '~/data/faq/global'"
        }
        if (moduleId === '\0faq-help-center-test-nuxt-imports') return `
          export { useI18n } from 'vue-i18n';
          export { useHead } from '@unhead/vue';
          export const definePageMeta = () => {};
          export const useRoute = () => globalThis.faqHelpCenterRenderingContext.route;
          export const useRuntimeConfig = () => ({ public: { siteUrl: 'https://example.com' } });
        `
        if (moduleId === '\0faq-help-center-test-useFaqCatalog') return `
          import { computed, ref } from 'vue';
          export async function useFaqCatalog() {
            await Promise.resolve();
            return {
              allPages: computed(() => globalThis.faqHelpCenterRenderingContext.pages),
              pending: ref(false),
              refreshFaqCatalog: async () => {},
            };
          }
        `
        if (moduleId === '\0faq-help-center-test-usePageMessages') return `
          export const usePageMessages = namespace => ({
            loadPageMessages: locale => globalThis.faqHelpCenterRenderingContext.loadPageMessages(namespace, locale),
          });
        `
        if (moduleId === '\0faq-help-center-test-seo/useStorefrontSeoLinks') return `
          import { computed } from 'vue';
          export const useStorefrontSeoLinks = () => ({
            canonicalUrl: computed(() => new URL(globalThis.faqHelpCenterRenderingContext.route.path, 'https://example.com').href),
          });
        `
      },
      transform(source, moduleId) {
        const normalizedModuleId = moduleId.replaceAll('\\', '/')
        const additionalImport = normalizedModuleId.endsWith('/app/pages/support/faqs.vue')
          ? "import { definePageMeta } from '#imports'"
          : normalizedModuleId.endsWith('/app/components/FaqAnswerContent.vue')
            ? "import { useI18n } from '#imports'"
            : ''
        if (additionalImport) return source.replace('<script setup lang="ts">', `<script setup lang="ts">\n${additionalImport}`)
      },
    },
    vuePlugin(),
  ],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

type FaqRenderingHtmlNode = DefaultTreeAdapterMap['node']
type FaqRenderingHtmlElement = DefaultTreeAdapterMap['element']

function collectFaqHelpCenterRenderedElements(node: FaqRenderingHtmlNode): FaqRenderingHtmlElement[] {
  return [
    ...('tagName' in node ? [node] : []),
    ...('childNodes' in node ? node.childNodes.flatMap(collectFaqHelpCenterRenderedElements) : []),
  ]
}

function readFaqHelpCenterRenderedAttribute(element: FaqRenderingHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readFaqHelpCenterRenderedText(node: FaqRenderingHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readFaqHelpCenterRenderedText).join('') : ''
}

try {
  const page = (await compilationServer.ssrLoadModule('/app/pages/support/faqs.vue')).default as Component
  const safeRichText = (await compilationServer.ssrLoadModule('/app/components/SafeRichText.vue')).default as Component

  for (const locale of ['en', 'zh_cn']) {
    const messages = JSON.parse(await readFile(
      path.join(frontendDirectory, `app/i18n/page-messages/supportFaqs/${locale}.json`), 'utf8',
    ))
    const pages: ResolvedPageFaqData[] = Array.from({ length: 4 }, (_, index) => ({
      pageId: `support-topic-${index + 1}`,
      title: `Support topic ${index + 1}`,
      items: [{
        id: `question-${index + 1}`,
        question: locale === 'en' ? `Question ${index + 1}: care & compatibility?` : `问题 ${index + 1}：保养与兼容性？`,
        answer: `<p>${locale === 'en' ? 'Answer' : '解答'} ${index + 1}: <strong>care</strong>&nbsp;&amp; compatibility.</p>`,
      }],
    }))

    for (const scenario of ['initial', 'filtered-page', 'empty-catalog'] as const) {
      faqHelpCenterRenderingContext.route = {
        path: `${locale === 'en' ? '' : '/zh_cn'}/support/faqs`,
        query: scenario === 'filtered-page' ? { page: 'support-topic-4', utm_source: 'regression' } : {},
      }
      faqHelpCenterRenderingContext.pages = scenario === 'empty-catalog' ? [] : pages
      const translations = createI18n({ legacy: false, locale, messages: {}, missingWarn: false, fallbackWarn: false })
      faqHelpCenterRenderingContext.loadPageMessages = async (namespace, requestedLocale) => {
        assert.equal(namespace, 'supportFaqs')
        assert.equal(requestedLocale, locale)
        await Promise.resolve()
        translations.global.mergeLocaleMessage(locale, { supportFaqs: messages })
      }
      const application = createSSRApp({ render: () => h(page) })
      const head = createHead()
      application.use(head)
      application.use(translations)
      application.component('SafeRichText', safeRichText)
      application.component('StorefrontImage', defineComponent({ render: () => h('img') }))
      application.component('Icon', defineComponent({ render: () => h('span') }))
      const renderingErrors: unknown[] = []
      application.config.errorHandler = error => { renderingErrors.push(error) }
      const bodyMarkup = await renderToString(application)
      assert.deepEqual(renderingErrors, [])
      const { headTags } = await renderSSRHead(head)
      const elements = collectFaqHelpCenterRenderedElements(parse(`<html><head>${headTags}</head><body>${bodyMarkup}</body></html>`))
      const headingElements = elements.filter(element => element.tagName === 'h1')
      assert.equal(headingElements.length, 1)
      assert.equal(readFaqHelpCenterRenderedText(headingElements[0]!), messages.title)
      const titleElements = elements.filter(element => element.tagName === 'title')
      assert.equal(titleElements.length, 1)
      assert.equal(readFaqHelpCenterRenderedText(titleElements[0]!), messages.ui.allFaqsMetaTitle)
      const descriptions = elements.filter(element => element.tagName === 'meta'
        && readFaqHelpCenterRenderedAttribute(element, 'name') === 'description')
      assert.equal(descriptions.length, 1)
      assert.equal(readFaqHelpCenterRenderedAttribute(descriptions[0]!, 'content'), messages.metaDescription)

      const structuredDataScripts = elements.filter(element => element.tagName === 'script'
        && readFaqHelpCenterRenderedAttribute(element, 'type') === 'application/ld+json')
      const expectedPages = scenario === 'empty-catalog' ? [] : scenario === 'initial' ? pages.slice(0, 3) : pages.slice(3)
      assert.equal(structuredDataScripts.length, expectedPages.length ? 1 : 0)
      if (expectedPages.length) {
        const schema = JSON.parse(readFaqHelpCenterRenderedText(structuredDataScripts[0]!))
        assert.equal(schema['@context'], 'https://schema.org')
        assert.equal(schema['@type'], 'FAQPage')
        assert.equal(schema.inLanguage, locale === 'en' ? 'en-US' : 'zh-CN')
        assert.equal(schema.url, `https://example.com${faqHelpCenterRenderingContext.route.path}`)
        assert.equal(schema['@id'], `${schema.url}#faq-support-faqs`)
        assert.deepEqual(schema.mainEntity.map((question: { name: string }) => question.name), expectedPages.map(topic => topic.items[0]!.question))
        const renderedAnswers = elements.filter(element => readFaqHelpCenterRenderedAttribute(element, 'class')
          .split(/\s+/).includes('global-all-faqs-mobile-grouped-search-results-panel__answer'))
        assert.equal(renderedAnswers.length, expectedPages.length, 'Collapsed answers must be present in server-rendered HTML')
        schema.mainEntity.forEach((question: { '@type': string; acceptedAnswer: { '@type': string; text: string } }, index: number) => {
          assert.equal(question['@type'], 'Question')
          assert.equal(question.acceptedAnswer['@type'], 'Answer')
          assert.equal(question.acceptedAnswer.text, readFaqHelpCenterRenderedText(renderedAnswers[index]!).replace(/\s+/g, ' ').trim())
        })
      }
      console.log(`FAQ help center SSR passed: ${locale}, ${scenario}`)
    }
  }
} finally {
  await compilationServer.close()
  delete renderingGlobals.faqHelpCenterRenderingContext
}

import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import vuePlugin from '@vitejs/plugin-vue'
import { createHead, renderSSRHead } from '@unhead/vue/server'
import { renderToString } from '@vue/server-renderer'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import { createServer } from 'vite'
import { createSSRApp, defineComponent, h, type Component, type Ref } from 'vue'
import { createI18n } from 'vue-i18n'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const shopRootRenderingContext = {
  route: { path: '/shop', fullPath: '/shop', query: {} as Record<string, string> },
  states: new Map<string, Ref>(),
  translations: {} as unknown,
  requestCategoryCatalog: async (_route: string, _options: { headers?: Record<string, string> }): Promise<unknown> => ({}),
}
const renderingGlobals = globalThis as typeof globalThis & { shopRootRenderingContext?: typeof shopRootRenderingContext }
renderingGlobals.shopRootRenderingContext = shopRootRenderingContext

const isolatedShopRootRuntimeModules: Record<string, string> = {
  '#imports': `
    import { ref } from 'vue';
    export { useI18n } from 'vue-i18n';
    export { useHead } from '@unhead/vue';
    export const definePageMeta = () => {};
    export const useRoute = () => globalThis.shopRootRenderingContext.route;
    export const useRouter = () => ({ replace: async () => {} });
    export const useLocalePath = () => path => {
      const prefix = globalThis.shopRootRenderingContext.route.path.slice(0, -'/shop'.length);
      return path.startsWith(prefix + '/shop') ? path : prefix + path;
    };
    export const useRuntimeConfig = () => ({ public: { siteUrl: 'https://example.com', apiBase: 'https://example.com/api/v1' } });
    export const useRequestURL = () => new URL(globalThis.shopRootRenderingContext.route.fullPath, 'https://example.com');
    export const useSwitchLocalePath = () => () => globalThis.shopRootRenderingContext.route.path;
    export const useNuxtApp = () => ({ $i18n: globalThis.shopRootRenderingContext.translations });
    export function useState(key, initialize) {
      const states = globalThis.shopRootRenderingContext.states;
      if (!states.has(key)) states.set(key, ref(initialize()));
      return states.get(key);
    }
    export async function useAsyncData(_key, load) {
      const data = ref(await load());
      return { data, pending: ref(false), error: ref(null), refresh: async () => { data.value = await load(); } };
    }
  `,
  'useApiRequest': 'export const useApiRequest = () => ({ request: (route, options) => globalThis.shopRootRenderingContext.requestCategoryCatalog(route, options) });',
  'useShopProducts': 'export const useShopProducts = () => ({ fetchShopProducts: async () => ({ items: [], raw: {} }) });',
  'useStorefrontContext': "import { ref } from 'vue'; export const useStorefrontContext = () => ({ displayCurrency: ref('USD') });",
  'useWishlist': 'export const useWishlist = () => ({ addToWishlist: async () => {} });',
  'useShopSearchSheet': "import { ref } from 'vue'; export const useShopSearchSheet = () => ({ pendingSearch: ref(null) });",
  'useOverlayBackStack': 'export const useOverlayBackStack = () => ({ open: () => {}, close: async () => {} });',
}

// Keep the real page, category loader/normalizer, category navigation, lazy
// translations, canonical helper, and head manager. Supply only external APIs
// and unrelated product, filter, recommendation, and feedback UI dependencies.
const compilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    {
      name: 'provide-isolated-storefront-services-for-shop-root-collection-ssr-tests',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0shop-root-test-#imports'
        const normalizedSource = source.replaceAll('\\', '/').replace(/\.ts$/, '')
        for (const name of Object.keys(isolatedShopRootRuntimeModules).filter(name => name !== '#imports')) {
          if (normalizedSource === `~/composables/${name}` || normalizedSource.endsWith(`/app/composables/${name}`)) {
            return `\0shop-root-test-${name}`
          }
        }
        if (/\/(?:UserFeedbackThread|ShopProductQuickSearchForm|ShopCatalogFilterDrawer|ShopProductDisplayCard|ProductRecommendations)\.vue$/.test(normalizedSource)) {
          return '\0shop-root-test-unrelated-ui-component'
        }
      },
      load(moduleId) {
        if (moduleId === '\0shop-root-test-unrelated-ui-component') {
          return "import { defineComponent, h } from 'vue'; export default defineComponent({ render: () => h('div') });"
        }
        if (moduleId.startsWith('\0shop-root-test-')) return isolatedShopRootRuntimeModules[moduleId.slice('\0shop-root-test-'.length)]
      },
      transform(source, moduleId) {
        const normalizedModuleId = moduleId.replaceAll('\\', '/')
        const additionalImport = normalizedModuleId.endsWith('/app/pages/shop/index.vue')
          ? "import { definePageMeta } from '#imports'"
          : normalizedModuleId.endsWith('/app/components/shop/ShopCategoryVerticalMenu.vue')
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

type ShopRootRenderingHtmlNode = DefaultTreeAdapterMap['node']
type ShopRootRenderingHtmlElement = DefaultTreeAdapterMap['element']

function collectShopRootServerRenderedElements(node: ShopRootRenderingHtmlNode): ShopRootRenderingHtmlElement[] {
  return [...('tagName' in node ? [node] : []), ...('childNodes' in node ? node.childNodes.flatMap(collectShopRootServerRenderedElements) : [])]
}

function readShopRootServerRenderedAttribute(element: ShopRootRenderingHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readShopRootServerRenderedText(node: ShopRootRenderingHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readShopRootServerRenderedText).join('') : ''
}

try {
  const page = (await compilationServer.ssrLoadModule('/app/pages/shop/index.vue')).default as Component
  for (const locale of ['en', 'zh_cn', 'fr']) {
    const translationLocale = locale === 'zh_cn' ? 'zh_cn' : 'en'
    const messages = JSON.parse(await readFile(path.join(frontendDirectory,
      `app/i18n/page-messages/shopRootCollectionPageSeo/${translationLocale}.json`), 'utf8'))
    const prefix = locale === 'en' ? '' : `/${locale}`
    const categoryTree = [
      { id: 1, name: locale === 'zh_cn' ? '轮组' : 'Wheelsets', slug: 'wheelset', route_path: `${prefix}/shop/wheelset`, children: [
        { id: 4, name: 'Road wheelsets', slug: 'road', depth: 2, route_path: `${prefix}/shop/wheelset/road` },
      ] },
      { id: 2, name: locale === 'zh_cn' ? '碳纤维轮圈' : 'Carbon rims', slug: 'rim', route_path: `${prefix}/shop/rim` },
      { id: 3, name: 'Spokes & nipples </script><script>untrusted</script>', slug: 'spokes-and-nipples' },
    ]

    for (const scenario of ['published-categories', 'empty-catalog'] as const) {
      shopRootRenderingContext.route = { path: `${prefix}/shop`, fullPath: `${prefix}/shop?currency=EUR&page=2&utm_source=regression`, query: { currency: 'EUR', page: '2', utm_source: 'regression' } }
      shopRootRenderingContext.states = new Map()
      const translations = createI18n({ legacy: false, locale, messages: {}, missingWarn: false, fallbackWarn: false })
      shopRootRenderingContext.translations = translations.global
      let categoryCatalogRequests = 0
      shopRootRenderingContext.requestCategoryCatalog = async (route, options) => {
        assert.equal(route, '/products/categories')
        assert.equal(options.headers?.['Accept-Language'], locale)
        categoryCatalogRequests += 1
        // Finish after the product request to catch category data omitted from SSR.
        await new Promise(resolveCatalogDelay => setTimeout(resolveCatalogDelay, 25))
        return { data: { tree: scenario === 'published-categories' ? categoryTree : [] } }
      }
      const application = createSSRApp({ render: () => h(page) })
      const head = createHead()
      application.use(head)
      application.use(translations)
      application.component('Icon', defineComponent({ render: () => h('span') }))
      const renderingErrors: unknown[] = []
      application.config.errorHandler = error => { renderingErrors.push(error) }
      const bodyMarkup = await renderToString(application)
      assert.deepEqual(renderingErrors, [])
      assert.equal(categoryCatalogRequests, 1, 'The root page must await published categories during SSR')
      const { headTags } = await renderSSRHead(head)
      const elements = collectShopRootServerRenderedElements(parse(`<html><head>${headTags}</head><body>${bodyMarkup}</body></html>`))
      const titleElements = elements.filter(element => element.tagName === 'title')
      assert.equal(titleElements.length, 1)
      assert.equal(readShopRootServerRenderedText(titleElements[0]!), messages.title)
      for (const [identifier, expected] of [
        ['description', messages.description], ['og:title', messages.title], ['og:description', messages.description],
        ['og:url', `https://example.com${prefix}/shop`], ['twitter:title', messages.title], ['twitter:description', messages.description],
      ]) {
        const meta = elements.filter(element => element.tagName === 'meta' && (
          readShopRootServerRenderedAttribute(element, 'name') === identifier || readShopRootServerRenderedAttribute(element, 'property') === identifier))
        assert.equal(meta.length, 1, `${identifier} must occur once`)
        assert.equal(readShopRootServerRenderedAttribute(meta[0]!, 'content'), expected)
      }
      const scripts = elements.filter(element => element.tagName === 'script')
      assert.equal(scripts.length, 1, 'Catalog names must not terminate the JSON-LD script')
      assert.equal(readShopRootServerRenderedAttribute(scripts[0]!, 'type'), 'application/ld+json')
      const schema = JSON.parse(readShopRootServerRenderedText(scripts[0]!))
      assert.equal(schema['@context'], 'https://schema.org')
      assert.equal(schema['@type'], 'CollectionPage')
      assert.equal(schema.url, `https://example.com${prefix}/shop`)
      assert.equal(schema['@id'], `${schema.url}#shop-collection`)
      assert.equal(schema.name, messages.title)
      assert.equal(schema.description, messages.description)
      assert.equal(schema.inLanguage, locale === 'zh_cn' ? 'zh-CN' : locale === 'fr' ? 'fr-FR' : 'en-US')
      if (scenario === 'published-categories') {
        assert.equal(schema.mainEntity['@type'], 'ItemList')
        assert.equal(schema.mainEntity.numberOfItems, categoryTree.length)
        assert.deepEqual(schema.mainEntity.itemListElement, categoryTree.map((category, index) => ({
          '@type': 'ListItem', position: index + 1, name: category.name,
          url: `https://example.com${category.route_path || `${prefix}/shop/${category.slug}`}`,
        })))
        const renderedCategoryNames = elements.filter(element => readShopRootServerRenderedAttribute(element, 'class')
          .split(/\s+/).includes('shop-category-menu__label-line')).map(readShopRootServerRenderedText)
        for (const category of categoryTree) assert.ok(renderedCategoryNames.includes(category.name), 'Schema category names must also appear in SSR navigation')
      } else {
        assert.equal('mainEntity' in schema, false, 'An empty catalog must not invent categories')
      }
      console.log(`Shop root CollectionPage SSR passed: ${locale}, ${scenario}`)
    }
  }
} finally {
  await compilationServer.close()
  delete renderingGlobals.shopRootRenderingContext
}

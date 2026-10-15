import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import localeManifest from '../app/i18n/locales.manifest.ts'
import { wheelsetBuyerTabs } from '../app/utils/pageSubNavigationData.ts'

const storefrontOrigin = process.env.SHOP_AND_WHEELSET_GUIDE_SSR_ORIGIN || 'http://127.0.0.1:10240'
const backendOrigin = process.env.SHOP_AND_WHEELSET_GUIDE_BACKEND_ORIGIN || 'http://127.0.0.1:10241'
type ServerRenderedNode = DefaultTreeAdapterMap['node']
type ServerRenderedElement = DefaultTreeAdapterMap['element']

function collectServerRenderedShopAndWheelsetGuideElements(node: ServerRenderedNode): ServerRenderedElement[] {
  return [
    ...('tagName' in node ? [node] : []),
    ...('childNodes' in node ? node.childNodes.flatMap(collectServerRenderedShopAndWheelsetGuideElements) : []),
  ]
}

function readServerRenderedShopAndWheelsetGuideAttribute(element: ServerRenderedElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readServerRenderedShopAndWheelsetGuideText(node: ServerRenderedNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readServerRenderedShopAndWheelsetGuideText).join('') : ''
}

interface WheelsetGuidePageMessages {
  title: string
  seo: {
    description: string
    sections: Record<string, { description: string }>
  }
  tabs: Record<string, { label: string }>
}

interface BackendProductCategory {
  name: string
  route_path: string
}

let verifiedRouteCount = 0
for (const locale of ['en', 'zh_cn', 'fr']) {
  const localePrefix = locale === 'en' ? '' : `/${locale}`
  const pageContentLocale = locale === 'zh_cn' ? 'zh_cn' : 'en'
  const guideMessages: WheelsetGuidePageMessages = JSON.parse(await readFile(
    new URL(`../app/i18n/page-messages/guidesWheelsetBuyers/${pageContentLocale}.json`, import.meta.url), 'utf8',
  ))
  const overviewMessages: { title: string } = JSON.parse(await readFile(
    new URL(`../app/i18n/page-messages/guidesWheelsetBuyersOverview/${pageContentLocale}.json`, import.meta.url), 'utf8',
  ))
  const shopMessages: { heading: string; title: string; description: string } = JSON.parse(await readFile(
    new URL(`../app/i18n/page-messages/shopRootCollectionPageSeo/${pageContentLocale}.json`, import.meta.url), 'utf8',
  ))
  const categoryResponse = await fetch(new URL('/api/v1/products/categories', backendOrigin), {
    headers: { 'Accept-Language': locale },
  })
  assert.equal(categoryResponse.status, 200)
  const categoryPayload = await categoryResponse.json() as { data: { flat: BackendProductCategory[] } }
  const checkedCategoryRoutes = ['/shop/wheelset', '/shop/wheel-components', '/shop/wheel-components/tire']
  const checkedCategories = checkedCategoryRoutes.map(routePath => {
    const category = categoryPayload.data.flat.find(category => category.route_path === `${localePrefix}${routePath}`)
    assert.ok(category, `Missing backend category fixture: ${routePath}`)
    return { path: routePath, name: category.name }
  })

  const routeCases = [
    { path: '/shop', heading: shopMessages.heading, title: shopMessages.title, description: shopMessages.description },
    ...checkedCategories.map(category => ({
      path: category.path, heading: category.name, title: '', description: '',
    })),
    {
      path: '/guides/wheelset-buyers', heading: guideMessages.title,
      title: guideMessages.title, description: guideMessages.seo.description,
    },
    ...wheelsetBuyerTabs.map(tab => {
      const tabMessageKey = tab.labelKey.split('.').at(-2)!
      const heading = tab.id === 'overview' ? overviewMessages.title : guideMessages.tabs[tabMessageKey]!.label
      return {
        path: `/guides/wheelset-buyers/${tab.id}`, heading,
        title: `${heading} | ${guideMessages.title}`,
        description: guideMessages.seo.sections[tab.id]!.description,
      }
    }),
  ]

  for (const routeCase of routeCases) {
    const routePath = `${localePrefix}${routeCase.path}`
    const response = await fetch(new URL(routePath, storefrontOrigin))
    assert.equal(response.status, 200, routePath)
    const elements = collectServerRenderedShopAndWheelsetGuideElements(parse(await response.text()))
    const headings = elements.filter(element => element.tagName === 'h1')
    assert.equal(headings.length, 1, `${routePath}: one H1 in initial HTML without JavaScript`)
    assert.equal(readServerRenderedShopAndWheelsetGuideText(headings[0]!).trim(), routeCase.heading, routePath)
    assert.doesNotMatch(readServerRenderedShopAndWheelsetGuideAttribute(headings[0]!, 'class'), /sr-only|\bhidden\b/)
    const titles = elements.filter(element => element.tagName === 'title')
    assert.equal(titles.length, 1, `${routePath}: one title`)
    const title = readServerRenderedShopAndWheelsetGuideText(titles[0]!).trim()
    assert.ok(title.length > 0)
    if (routeCase.title) assert.ok(title.startsWith(routeCase.title), `${routePath}: localized page title`)

    for (const metadataKey of ['description', 'og:description', 'twitter:description', 'og:title', 'twitter:title']) {
      const metadata = elements.filter(element => element.tagName === 'meta'
        && (readServerRenderedShopAndWheelsetGuideAttribute(element, 'name') === metadataKey
          || readServerRenderedShopAndWheelsetGuideAttribute(element, 'property') === metadataKey))
      assert.equal(metadata.length, 1, `${routePath}: exactly one ${metadataKey}`)
      const content = readServerRenderedShopAndWheelsetGuideAttribute(metadata[0]!, 'content')
      assert.ok(content.trim().length > 0, `${routePath}: nonempty ${metadataKey}`)
      assert.doesNotMatch(content, /guidesWheelsetBuyers\.|shopRootCollectionPageSeo\./)
      const expectedContent = metadataKey.endsWith('title') ? routeCase.title : routeCase.description
      if (expectedContent) assert.equal(content, expectedContent, `${routePath}: ${metadataKey}`)
    }

    const canonicalLinks = elements.filter(element => element.tagName === 'link'
      && readServerRenderedShopAndWheelsetGuideAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalLinks.length, 1, `${routePath}: one canonical`)
    assert.equal(new URL(readServerRenderedShopAndWheelsetGuideAttribute(canonicalLinks[0]!, 'href')).pathname, routePath)
    const alternateLinks = elements.filter(element => element.tagName === 'link'
      && readServerRenderedShopAndWheelsetGuideAttribute(element, 'rel') === 'alternate'
      && readServerRenderedShopAndWheelsetGuideAttribute(element, 'hreflang'))
    assert.equal(alternateLinks.length, localeManifest.length + 1, `${routePath}: language links and x-default`)
    assert.equal(new Set(alternateLinks.map(element => readServerRenderedShopAndWheelsetGuideAttribute(element, 'hreflang'))).size,
      alternateLinks.length, `${routePath}: unique language links`)

    if (routeCase.path.startsWith('/guides/wheelset-buyers/')) {
      const activeSectionId = routeCase.path.split('/').at(-1)!
      for (const tab of wheelsetBuyerTabs) {
        assert.equal(elements.filter(element => element.tagName === 'section'
          && readServerRenderedShopAndWheelsetGuideAttribute(element, 'id') === tab.id).length,
        tab.id === activeSectionId ? 1 : 0, `${routePath}: only the active guide section is rendered`)
      }
      assert.ok(elements.some(element => element.tagName === 'h2'), `${routePath}: H2 subsections below the H1`)
    }
    if (routeCase.path.startsWith('/shop')) {
      const headingParent = headings[0]!.parentNode
      assert.ok(headingParent && 'tagName' in headingParent, `${routePath}: page heading has an element parent`)
      assert.ok(readServerRenderedShopAndWheelsetGuideAttribute(headingParent, 'class').split(/\s+/).includes('shop-page'),
        `${routePath}: H1 belongs to the page, outside the search card`)
      const structuredData = elements.filter(element => element.tagName === 'script'
        && readServerRenderedShopAndWheelsetGuideAttribute(element, 'type') === 'application/ld+json')
      assert.ok(structuredData.some(element => JSON.parse(readServerRenderedShopAndWheelsetGuideText(element))['@type'] === 'CollectionPage'),
        `${routePath}: collection structured data remains available`)
    }
    verifiedRouteCount += 1
    console.log(`PASS ${routePath}: one visible H1, localized search/social metadata, canonical and language links`)
  }
}
console.log(`Verified ${verifiedRouteCount} shop and wheelset guide server-rendered routes.`)

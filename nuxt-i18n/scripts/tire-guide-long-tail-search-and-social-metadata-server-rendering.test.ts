import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import sharp from 'sharp'

type TireGuideMetadataHtmlNode = DefaultTreeAdapterMap['node']
type TireGuideMetadataHtmlElement = DefaultTreeAdapterMap['element']
type TireGuideTranslatedSearchMetadata = { title: string; description: string; imageAlt?: string }

const storefrontBaseUrl = process.env.SEO_STOREFRONT_BASE_URL
if (!storefrontBaseUrl) throw new Error('Set SEO_STOREFRONT_BASE_URL to a running storefront origin.')

function collectTireGuideMetadataHtmlElements(node: TireGuideMetadataHtmlNode): TireGuideMetadataHtmlElement[] {
  const elements: TireGuideMetadataHtmlElement[] = 'tagName' in node ? [node] : []
  if ('childNodes' in node) {
    for (const child of node.childNodes) elements.push(...collectTireGuideMetadataHtmlElements(child))
  }
  return elements
}

function readTireGuideMetadataHtmlAttribute(element: TireGuideMetadataHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value ?? ''
}

function readTireGuideMetadataHtmlText(node: TireGuideMetadataHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readTireGuideMetadataHtmlText).join('') : ''
}

function readUniqueTireGuideSearchOrSocialMetaContent(elements: TireGuideMetadataHtmlElement[], identifier: string): string {
  const matches = elements.filter(element => element.tagName === 'meta'
    && (readTireGuideMetadataHtmlAttribute(element, 'name') === identifier
      || readTireGuideMetadataHtmlAttribute(element, 'property') === identifier))
  assert.equal(matches.length, 1, `Expected exactly one ${identifier} tag`)
  const content = readTireGuideMetadataHtmlAttribute(matches[0]!, 'content')
  assert.ok(content.trim(), `${identifier} must not be empty`)
  assert.doesNotMatch(content, /guidesTireguides\./, 'Translation keys must not enter metadata')
  return content
}

const tireGuideRouteMetadataChecks = [
  { tabId: 'schwalbe-tire-circumference', messageKey: 'schwalbeCircumference', expectedSearchTerms: /Garmin.*Wahoo/, expectedImagePath: '/public/tiresizecharts/schwalbe-tire-circumference/exact-circumference-of-tire.webp' },
  { tabId: 'tire-size-markings', routePath: '/tools/tire-size-marking-converter', messageKey: 'tireSizeMarkings', expectedSearchTerms: /ETRTO.*700C|700C.*ETRTO/, expectedImagePath: '/public/tiresizecharts/tiresize/schwalbe-tiresize.webp' },
  { tabId: 'tubeless', messageKey: 'tubeless', expectedSearchTerms: /sealant|补胎液/, expectedImagePath: '/public/tiresizecharts/installation/3/Tubelesstire-installation-diagram-2.webp' },
  { tabId: 'tire-pressure', messageKey: 'tirePressure', expectedSearchTerms: /rider weight|骑手体重/, expectedImagePath: null },
]
const verifiedSocialImages = new Set<string>()

for (const localeCode of ['en', 'zh_cn', 'fr'] as const) {
  const messageLocaleCode = localeCode === 'zh_cn' ? 'zh_cn' : 'en'
  const messages = JSON.parse(readFileSync(new URL(`../app/i18n/page-messages/guidesTireguides/${messageLocaleCode}.json`, import.meta.url), 'utf8')) as {
    seo: Record<string, TireGuideTranslatedSearchMetadata>
  }
  const descriptionValues = new Set<string>()
  for (const routeCheck of tireGuideRouteMetadataChecks) {
    const routeBasePath = routeCheck.routePath || (routeCheck.tabId === 'tire-pressure'
      ? '/tools/dynamic-tire-pressure-calculator'
      : `/guides/tireguides/${routeCheck.tabId}`)
    const routePath = `${localeCode === 'en' ? '' : `/${localeCode}`}${routeBasePath}`
    // The inner-tube calculator selector and tracking must not affect other guides.
    const response: Response = await fetch(new URL(`${routePath}?tab=calculator&utm_source=forum`, storefrontBaseUrl), { signal: AbortSignal.timeout(60_000) })
    assert.equal(response.status, 200, `${routePath} must return HTTP 200`)
    const elements = collectTireGuideMetadataHtmlElements(parse(await response.text()))
    const expectedMetadata = messages.seo[routeCheck.messageKey]!
    const titles = elements.filter(element => element.tagName === 'title')
    assert.equal(titles.length, 1)
    const title = readTireGuideMetadataHtmlText(titles[0]!)
    assert.ok(title === expectedMetadata.title || title.startsWith(`${expectedMetadata.title} | `), 'Expected the guide title with an optional site brand')
    const description = readUniqueTireGuideSearchOrSocialMetaContent(elements, 'description')
    assert.equal(description, expectedMetadata.description)
    assert.match(title + description, routeCheck.expectedSearchTerms)
    descriptionValues.add(description)
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:title'), title)
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'twitter:title'), title)
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:description'), description)
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'twitter:description'), description)
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:type'), 'article')
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:locale'), { en: 'en_US', zh_cn: 'zh_CN', fr: 'fr_FR' }[localeCode])

    const canonicals = elements.filter(element => element.tagName === 'link' && readTireGuideMetadataHtmlAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicals.length, 1)
    const canonicalUrl = new URL(readTireGuideMetadataHtmlAttribute(canonicals[0]!, 'href'))
    assert.equal(canonicalUrl.pathname, routePath)
    assert.equal(canonicalUrl.search + canonicalUrl.hash, '')
    assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:url'), canonicalUrl.toString())
    const robots = elements.filter(element => element.tagName === 'meta' && readTireGuideMetadataHtmlAttribute(element, 'name') === 'robots')
    assert.equal(robots.length, 0, 'The inner-tube fallback-locale noindex policy must not leak to another guide')

    const imageTags = elements.filter(element => element.tagName === 'meta' && (
      readTireGuideMetadataHtmlAttribute(element, 'property').startsWith('og:image')
        || readTireGuideMetadataHtmlAttribute(element, 'name').startsWith('twitter:image')
    ))
    if (routeCheck.expectedImagePath) {
      const imageUrl = new URL(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:image'))
      assert.equal(imageUrl.origin, canonicalUrl.origin)
      assert.equal(imageUrl.pathname, routeCheck.expectedImagePath)
      assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:image:alt'), expectedMetadata.imageAlt)
      assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'twitter:image:alt'), expectedMetadata.imageAlt)
      assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'twitter:image'), imageUrl.toString())
      assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'twitter:card'), 'summary_large_image')
      if (!verifiedSocialImages.has(imageUrl.pathname)) {
        const imageResponse: Response = await fetch(new URL(imageUrl.pathname, storefrontBaseUrl))
        assert.equal(imageResponse.status, 200, 'The guide image must be publicly available')
        const imageMetadata = await sharp(Buffer.from(await imageResponse.arrayBuffer())).metadata()
        assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:image:type'), `image/${imageMetadata.format}`)
        assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:image:width'), String(imageMetadata.width))
        assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'og:image:height'), String(imageMetadata.height))
        verifiedSocialImages.add(imageUrl.pathname)
      }
    } else {
      assert.equal(imageTags.length, 0, 'A guide without an image must not inherit another guide image')
      assert.equal(readUniqueTireGuideSearchOrSocialMetaContent(elements, 'twitter:card'), 'summary')
    }
    console.log(`Tire guide search and social SSR metadata passed: ${routePath}`)
  }
  assert.equal(descriptionValues.size, 4, 'Every guide must have a distinct description')
}

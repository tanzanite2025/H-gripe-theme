import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import { bicycleInnerTubeGuideManufacturerReferenceUrls } from '../../app/utils/seo/createBicycleInnerTubeGuideArticleAndBreadcrumbStructuredData.ts'

type InnerTubeContentHtmlNode = DefaultTreeAdapterMap['node']
type InnerTubeContentHtmlElement = DefaultTreeAdapterMap['element']
const storefrontBaseUrl = process.env.SEO_STOREFRONT_BASE_URL
if (!storefrontBaseUrl) throw new Error('Set SEO_STOREFRONT_BASE_URL to a running storefront origin.')

function collectInnerTubeContentHtmlElements(node: InnerTubeContentHtmlNode): InnerTubeContentHtmlElement[] {
  return [ ...('tagName' in node ? [node] : []), ...('childNodes' in node ? node.childNodes.flatMap(collectInnerTubeContentHtmlElements) : []) ]
}

function readInnerTubeContentHtmlAttribute(element: InnerTubeContentHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readInnerTubeContentHtmlText(node: InnerTubeContentHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readInnerTubeContentHtmlText).join('') : ''
}

function normalizeInnerTubeContentVisibleText(value: string): string {
  return value.replace(/\s+/g, ' ').trim()
}

// Untrusted script content is parsed as unknown before structural assertions.
function readInnerTubeContentSchemaRecord(value: unknown): Record<string, unknown> {
  assert.ok(value && typeof value === 'object' && !Array.isArray(value))
  return value as Record<string, unknown>
}

for (const localeCode of ['en', 'zh_cn', 'fr']) {
  const translationCode = localeCode === 'zh_cn' ? 'zh_cn' : 'en'
  const messages = JSON.parse(readFileSync(new URL(`../../app/i18n/page-messages/guidesTireInnerTube/${translationCode}.json`, import.meta.url), 'utf8'))
  const languageTag = localeCode === 'zh_cn' ? 'zh-CN' : 'en-US'
  const routePath = `${localeCode === 'en' ? '' : `/${localeCode}`}/tools/valve-length-fitment-calculator`
  for (const suffix of ['', '?rim_depth=80&fitment_mode=manual&valve_length=80&extender_length=30&pump_head_grip_depth=22&rim_depth_uncertainty=4&utm_source=forum', '?tab=calculator&rim_depth=80&utm_source=forum']) {
    const response: Response = await fetch(new URL(routePath + suffix, storefrontBaseUrl), { signal: AbortSignal.timeout(60_000) })
    assert.equal(response.status, 200)
    const elements = collectInnerTubeContentHtmlElements(parse(await response.text()))
    const canonical = elements.filter(element => element.tagName === 'link' && readInnerTubeContentHtmlAttribute(element, 'rel') === 'canonical')
    assert.equal(canonical.length, 1)
    const canonicalUrl = readInnerTubeContentHtmlAttribute(canonical[0]!, 'href')
    assert.equal(new URL(canonicalUrl).search, suffix.includes('tab=calculator') ? '?tab=calculator' : '')
    assert.equal(new URL(canonicalUrl).hash, '')
    const allHeadingOneElements = elements.filter(element => element.tagName === 'h1')
    assert.equal(allHeadingOneElements.length, 1, 'The page must have one H1')
    assert.equal(normalizeInnerTubeContentVisibleText(readInnerTubeContentHtmlText(allHeadingOneElements[0]!)), messages.page.title)
    assert.doesNotMatch(readInnerTubeContentHtmlAttribute(allHeadingOneElements[0]!, 'class'), /sr-only/)
    const guide = elements.filter(element => readInnerTubeContentHtmlAttribute(element, 'id') === 'inner-tube-guide')
    assert.equal(guide.length, 1)
    assert.equal(guide[0]!.tagName, 'article')
    assert.equal(readInnerTubeContentHtmlAttribute(guide[0]!, 'lang'), languageTag)
    const guideElements = collectInnerTubeContentHtmlElements(guide[0]!)
    assert.equal(guideElements.some(element => element.tagName === 'nav' && readInnerTubeContentHtmlAttribute(element, 'aria-label') === messages.page.breadcrumbLabel), false, 'Do not add a second breadcrumb inside the guide')
    assert.ok(elements.some(element => element.tagName === 'nav' && readInnerTubeContentHtmlAttribute(element, 'class').split(/\s+/).includes('site-header-breadcrumb-row')), 'Keep the existing site-header breadcrumb')
    const guideVisibleText = normalizeInnerTubeContentVisibleText(readInnerTubeContentHtmlText(guide[0]!))
    assert.ok(guideVisibleText.includes(messages.page.summary))
    assert.ok(guideVisibleText.includes(messages.steps.size.etrtoExample))
    assert.ok(guideVisibleText.includes(messages.steps.size.widthRangeExample))
    let previousHeadingLevel = 0
    for (const heading of guideElements.filter(element => /^h[1-6]$/.test(element.tagName))) {
      const level = Number(heading.tagName.slice(1))
      assert.ok(level <= previousHeadingLevel + 1, 'Guide headings must not skip a level')
      previousHeadingLevel = level
    }
    const idElements = new Map<string, InnerTubeContentHtmlElement>()
    for (const element of guideElements) {
      const id = readInnerTubeContentHtmlAttribute(element, 'id')
      if (!id) continue
      assert.equal(idElements.has(id), false, `Duplicate guide ID: ${id}`)
      idElements.set(id, element)
    }
    for (const link of guideElements.filter(element => element.tagName === 'a')) {
      const href = readInnerTubeContentHtmlAttribute(link, 'href')
      const navigationUrl = new URL(href, canonicalUrl)
      if (navigationUrl.pathname === routePath && navigationUrl.hash) {
        assert.ok(idElements.has(navigationUrl.hash.slice(1)), `Missing navigation target: ${href}`)
      }
      assert.equal(navigationUrl.origin, new URL(canonicalUrl).origin, 'Guide navigation must remain on the storefront')
    }
    const scripts = elements.filter(element => element.tagName === 'script' && readInnerTubeContentHtmlAttribute(element, 'type') === 'application/ld+json')
    const entities = scripts.flatMap(element => {
      const schema = readInnerTubeContentSchemaRecord(JSON.parse(readInnerTubeContentHtmlText(element)))
      return Array.isArray(schema['@graph']) ? schema['@graph'].map(readInnerTubeContentSchemaRecord) : [schema]
    })
    const articleEntities = entities.filter(entity => entity['@id'] === `${canonicalUrl}#inner-tube-guide`)
    assert.equal(articleEntities.length, 1)
    const article = articleEntities[0]!
    assert.equal(article['@type'], 'Article')
    assert.equal(article.headline, messages.page.title)
    assert.equal(article.description, messages.page.summary)
    assert.equal(article.inLanguage, languageTag)
    assert.equal(article.url, canonicalUrl)
    assert.equal(article.mainEntityOfPage, canonicalUrl)
    if (article.publisher) {
      const publisher = readInnerTubeContentSchemaRecord(article.publisher)
      assert.equal(publisher['@type'], 'Organization')
      assert.ok(guideVisibleText.includes(String(publisher.name)))
    }
    const articleImage = readInnerTubeContentSchemaRecord(article.image)
    assert.equal(articleImage.width, 600)
    assert.equal(articleImage.height, 300)
    assert.equal(articleImage.caption, messages.valves.imageOverviewAlt)
    assert.ok(guideElements.some(element => element.tagName === 'img' && readInnerTubeContentHtmlAttribute(element, 'alt') === articleImage.caption))
    assert.deepEqual(article.citation, [...bicycleInnerTubeGuideManufacturerReferenceUrls])
    const referencesSection = idElements.get('inner-tube-guide-references')
    assert.ok(referencesSection)
    const referencesText = normalizeInnerTubeContentVisibleText(readInnerTubeContentHtmlText(referencesSection))
    for (const source of [messages.references.source1, messages.references.source2]) {
      assert.ok(referencesText.includes(source.title))
      assert.ok(referencesText.includes(source.description))
    }
    assert.equal(collectInnerTubeContentHtmlElements(referencesSection).some(element => element.tagName === 'a'), false, 'Manufacturer references must be plain text without outbound links')
    for (const property of ['author', 'datePublished', 'dateModified', 'aggregateRating', 'review']) assert.equal(property in article, false, `Do not invent ${property}`)
    assert.ok(Array.isArray(article.hasPart))
    for (const part of article.hasPart.map(readInnerTubeContentSchemaRecord)) {
      const entityId = String(part['@id'])
      assert.equal(entities.filter(entity => entity['@id'] === entityId).length, 1)
      assert.ok(idElements.has(new URL(entityId).hash.slice(1)))
    }
    assert.equal(entities.some(entity => entity['@type'] === 'TechArticle'), false)
    const breadcrumbs = entities.filter(entity => entity['@id'] === `${canonicalUrl}#inner-tube-guide-breadcrumbs`)
    assert.equal(breadcrumbs.length, 1)
    const breadcrumb = breadcrumbs[0]!
    assert.equal(breadcrumb['@type'], 'BreadcrumbList')
    assert.ok(Array.isArray(breadcrumb.itemListElement))
    const breadcrumbItems = breadcrumb.itemListElement.map(readInnerTubeContentSchemaRecord)
    const localePrefix = localeCode === 'en' ? '' : `/${localeCode}`
    const expectedBreadcrumbPaths = [localePrefix ? `${localePrefix}` : '/', `${localePrefix}/tools`, routePath]
    assert.equal(breadcrumbItems.length, 3)
    breadcrumbItems.forEach((item, index) => {
      assert.equal(item.position, index + 1)
      const url = new URL(String(item.item))
      assert.equal(url.pathname.replace(/\/$/, '') || '/', expectedBreadcrumbPaths[index])
      assert.equal(url.origin, new URL(canonicalUrl).origin)
      assert.equal(url.search + url.hash, '')
      assert.ok(typeof item.name === 'string' && item.name.trim(), 'Breadcrumb schema entries must have names')
    })
    const howTo = entities.find(entity => entity['@id'] === `${canonicalUrl}#inner-tube-fitment-how-to`)!
    assert.equal(howTo['@type'], 'HowTo')
    assert.ok(Array.isArray(howTo.step))
    assert.equal(howTo.step.length, 6)
    for (const [index, step] of howTo.step.map(readInnerTubeContentSchemaRecord).entries()) {
      assert.equal(step.position, index + 1)
      const stepElement = idElements.get(new URL(String(step.url)).hash.slice(1))
      assert.ok(stepElement)
      const stepText = normalizeInnerTubeContentVisibleText(readInnerTubeContentHtmlText(stepElement))
      assert.ok(stepText.includes(String(step.name)))
      assert.ok(stepText.includes(String(step.text)))
    }
    for (const entity of entities.filter(entity => String(entity['@id'] || '').includes('#inner-tube-fitment-'))) {
      assert.equal(entity.inLanguage, languageTag)
      assert.deepEqual(entity.isPartOf, { '@id': article['@id'] })
      assert.equal('aggregateRating' in entity || 'review' in entity, false)
    }
    const faqs = entities.filter(entity => entity['@type'] === 'FAQPage')
    assert.ok(faqs.length <= 1, 'Do not duplicate the backend-owned FAQ')
    if (process.env.INNER_TUBE_GUIDE_REQUIRE_FAQ_OUTPUT === 'true') assert.equal(faqs.length, 1)
    if (faqs.length) {
      assert.ok(elements.some(element => element.tagName === 'h2' && readInnerTubeContentHtmlAttribute(element, 'class').includes('page-faq__title')))
      const questions = faqs[0]!.mainEntity
      assert.ok(Array.isArray(questions))
      for (const question of questions.map(readInnerTubeContentSchemaRecord)) assert.ok(elements.some(element => readInnerTubeContentHtmlAttribute(element, 'class').includes('page-faq__question') && readInnerTubeContentHtmlText(element).includes(String(question.name))))
    }
    console.log(`Inner-tube guide content/schema passed: ${localeCode}, ${suffix || 'plain'} URL`)
  }
}

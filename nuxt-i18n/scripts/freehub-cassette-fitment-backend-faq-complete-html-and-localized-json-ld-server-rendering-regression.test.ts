import assert from 'node:assert/strict'
import { parse, parseFragment, type DefaultTreeAdapterMap } from 'parse5'
import storefrontLocales from '../app/i18n/locales.manifest.ts'

const freehubFaqStorefrontOrigin = process.env.FREEHUB_CASSETTE_FITMENT_BASE_URL || 'http://127.0.0.1:10240'
const freehubFaqBackendOrigin = process.env.FREEHUB_CASSETTE_FITMENT_BACKEND_ORIGIN || 'http://127.0.0.1:10241'
const freehubFaqRoutePath = '/tools/freehub-cassette-fitment-engine'

type FreehubFaqRenderedNode = DefaultTreeAdapterMap['node']
type FreehubFaqRenderedElement = DefaultTreeAdapterMap['element']

interface FreehubBackendConfiguredFaqPage {
  pageId: string
  title: string
  subtitle: string
  items: { question: string; answer: string }[]
}

function collectFreehubFaqServerRenderedElements(node: FreehubFaqRenderedNode): FreehubFaqRenderedElement[] {
  return [
    ...('tagName' in node ? [node] : []),
    ...('childNodes' in node ? node.childNodes.flatMap(collectFreehubFaqServerRenderedElements) : []),
  ]
}

function readFreehubFaqServerRenderedAttribute(element: FreehubFaqRenderedElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readFreehubFaqServerRenderedText(node: FreehubFaqRenderedNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  const text = 'childNodes' in node ? node.childNodes.map(readFreehubFaqServerRenderedText).join('') : ''
  return 'tagName' in node && /^(?:br|p|li|div|h[1-6]|tr|blockquote)$/.test(node.tagName) ? `${text} ` : text
}

function normalizeFreehubFaqServerRenderedText(node: FreehubFaqRenderedNode): string {
  return readFreehubFaqServerRenderedText(node).replace(/\s+/g, ' ').trim()
}

function freehubFaqServerRenderedElementHasClass(element: FreehubFaqRenderedElement, className: string): boolean {
  return readFreehubFaqServerRenderedAttribute(element, 'class').split(/\s+/).includes(className)
}

for (const locale of ['en', 'zh_cn']) {
  const backendFaqUrl = new URL('/api/v1/content/faq-pages/by-route', freehubFaqBackendOrigin)
  backendFaqUrl.searchParams.set('route_path', freehubFaqRoutePath)
  backendFaqUrl.searchParams.set('locale', locale)
  const backendResponse = await fetch(backendFaqUrl, { signal: AbortSignal.timeout(30_000) })
  assert.equal(backendResponse.status, 200, backendFaqUrl.toString())
  const backendPage: FreehubBackendConfiguredFaqPage = (await backendResponse.json()).page
  assert.ok(backendPage.items.length > 0, 'Apply the freehub FAQ migration or publish backend questions before running this regression')

  const localizedRoutePath = `${locale === 'en' ? '' : `/${locale}`}${freehubFaqRoutePath}`
  for (const query of ['', '?brand=SRAM&cassette_spec=sram_12s_10_52t']) {
    const storefrontResponse = await fetch(new URL(`${localizedRoutePath}${query}`, freehubFaqStorefrontOrigin), {
      signal: AbortSignal.timeout(60_000),
    })
    assert.equal(storefrontResponse.status, 200, `${localizedRoutePath}${query}`)
    const elements = collectFreehubFaqServerRenderedElements(parse(await storefrontResponse.text()))
    assert.equal(elements.filter(element => element.tagName === 'h1').length, 1, 'The tool retains one main heading')
    const canonicalLinks = elements.filter(element => element.tagName === 'link'
      && readFreehubFaqServerRenderedAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalLinks.length, 1)
    const canonicalUrl = new URL(readFreehubFaqServerRenderedAttribute(canonicalLinks[0]!, 'href'))
    assert.equal(canonicalUrl.pathname, localizedRoutePath)
    assert.equal(canonicalUrl.search, '', 'Cassette selection parameters must not create duplicate canonical URLs')

    const faqSections = elements.filter(element => freehubFaqServerRenderedElementHasClass(element, 'page-faq'))
    assert.equal(faqSections.length, 1, 'The layout inserts one backend-configured FAQ module')
    const faqElements = collectFreehubFaqServerRenderedElements(faqSections[0]!)
    const faqHeadings = faqElements.filter(element => freehubFaqServerRenderedElementHasClass(element, 'page-faq__title'))
    assert.equal(faqHeadings.length, 1)
    assert.equal(faqHeadings[0]!.tagName, 'h2', 'FAQ is a top-level section below the tool H1')
    assert.equal(normalizeFreehubFaqServerRenderedText(faqHeadings[0]!), backendPage.title)
    const faqSubtitles = faqElements.filter(element => freehubFaqServerRenderedElementHasClass(element, 'page-faq__subtitle'))
    if (backendPage.subtitle) {
      assert.equal(faqSubtitles.length, 1)
      assert.equal(normalizeFreehubFaqServerRenderedText(faqSubtitles[0]!), backendPage.subtitle)
    }
    const questionButtons = faqElements.filter(element => freehubFaqServerRenderedElementHasClass(element, 'page-faq__question'))
    const renderedAnswers = faqElements.filter(element => freehubFaqServerRenderedElementHasClass(element, 'page-faq__answer'))
    assert.equal(questionButtons.length, backendPage.items.length)
    assert.equal(renderedAnswers.length, backendPage.items.length, 'Every collapsed answer is present in HTML before JavaScript runs')

    const structuredData = elements.filter(element => element.tagName === 'script'
      && readFreehubFaqServerRenderedAttribute(element, 'type') === 'application/ld+json')
      .flatMap(element => {
        const schema = JSON.parse(readFreehubFaqServerRenderedText(element))
        return schema['@graph'] || [schema]
      })
    const faqStructuredData = structuredData.filter(schema => schema['@type'] === 'FAQPage')
    assert.equal(faqStructuredData.length, 1, 'Emit a single FAQPage schema from the published backend data')
    const faqSchema = faqStructuredData[0]
    assert.equal(faqSchema.url, canonicalUrl.toString())
    assert.equal(faqSchema['@id'], `${canonicalUrl}#faq-${encodeURIComponent(backendPage.pageId)}`)
    assert.equal(faqSchema.inLanguage, storefrontLocales.find(entry => entry.code === locale)!.iso)
    assert.equal(faqSchema.mainEntity.length, backendPage.items.length)
    for (const [index, item] of backendPage.items.entries()) {
      const answerText = normalizeFreehubFaqServerRenderedText(parseFragment(item.answer))
      assert.equal(normalizeFreehubFaqServerRenderedText(questionButtons[index]!), item.question)
      assert.equal(normalizeFreehubFaqServerRenderedText(renderedAnswers[index]!), answerText)
      assert.equal(faqSchema.mainEntity[index].name, item.question.trim())
      assert.equal(faqSchema.mainEntity[index].acceptedAnswer.text.replace(/\s+/g, ''), answerText.replace(/\s+/g, ''))
    }
    console.log(`PASS ${localizedRoutePath}${query}: ${backendPage.items.length} backend FAQ answers in initial HTML and matching localized JSON-LD`)
  }
}

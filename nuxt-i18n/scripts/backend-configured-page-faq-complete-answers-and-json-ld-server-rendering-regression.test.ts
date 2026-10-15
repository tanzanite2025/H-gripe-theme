import assert from 'node:assert/strict'
import { parse, parseFragment, type DefaultTreeAdapterMap } from 'parse5'

const faqStorefrontOrigin = process.env.COMPANY_ABOUT_SSR_ORIGIN || 'http://127.0.0.1:10240'
const faqBackendOrigin = process.env.COMPANY_ABOUT_FAQ_BACKEND_ORIGIN || 'http://127.0.0.1:10241'

type BackendConfiguredFaqRenderedNode = DefaultTreeAdapterMap['node']
type BackendConfiguredFaqRenderedElement = DefaultTreeAdapterMap['element']

function collectBackendConfiguredFaqRenderedElements(node: BackendConfiguredFaqRenderedNode): BackendConfiguredFaqRenderedElement[] {
  return [
    ...('tagName' in node ? [node] : []),
    ...('childNodes' in node ? node.childNodes.flatMap(collectBackendConfiguredFaqRenderedElements) : []),
  ]
}

function readBackendConfiguredFaqRenderedAttribute(element: BackendConfiguredFaqRenderedElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readBackendConfiguredFaqRenderedText(node: BackendConfiguredFaqRenderedNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  const text = 'childNodes' in node ? node.childNodes.map(readBackendConfiguredFaqRenderedText).join('') : ''
  return 'tagName' in node && /^(?:br|p|li|div|h[1-6]|tr|blockquote)$/.test(node.tagName) ? `${text} ` : text
}

function normalizeBackendConfiguredFaqRichAnswerText(answer: string): string {
  return readBackendConfiguredFaqRenderedText(parseFragment(answer))
    .replace(/\s+/g, ' ').trim()
}

interface BackendConfiguredFaqPage {
  title: string
  items: { question: string; answer: string }[]
}

for (const locale of ['en', 'zh_cn']) {
  for (const routePath of [
    '/company/about/appearance',
    '/company/about/hole-patterns',
    '/tools/dynamic-tire-pressure-calculator',
    '/tools/tire-rim-width-compatibility-engine',
    '/tools/schwalbe-tire-selector',
  ]) {
    const backendFaqUrl = new URL('/api/v1/content/faq-pages/by-route', faqBackendOrigin)
    backendFaqUrl.searchParams.set('route_path', routePath)
    backendFaqUrl.searchParams.set('locale', locale)
    const localizedRoutePath = `${locale === 'en' ? '' : `/${locale}`}${routePath}`
    const [backendResponse, storefrontResponse] = await Promise.all([
      fetch(backendFaqUrl), fetch(new URL(localizedRoutePath, faqStorefrontOrigin)),
    ])
    assert.equal(backendResponse.status, 200, localizedRoutePath)
    assert.equal(storefrontResponse.status, 200, localizedRoutePath)
    const backendPage: BackendConfiguredFaqPage = (await backendResponse.json()).page
    assert.ok(backendPage.items.length > 0, 'Publish backend FAQ content before running this regression')
    const elements = collectBackendConfiguredFaqRenderedElements(parse(await storefrontResponse.text()))
    const faqSections = elements.filter(element => readBackendConfiguredFaqRenderedAttribute(element, 'class').split(/\s+/).includes('page-faq'))
    assert.equal(faqSections.length, 1, `${localizedRoutePath}: one backend-configured FAQ module`)
    if (routePath === '/company/about/hole-patterns') {
      const feedbackSections = elements.filter(element => readBackendConfiguredFaqRenderedAttribute(element, 'class').split(/\s+/).includes('share-your-feedback-card'))
      assert.equal(feedbackSections.length, 1, `${localizedRoutePath}: one feedback section`)
      assert.ok(elements.indexOf(faqSections[0]!) < elements.indexOf(feedbackSections[0]!), 'The backend FAQ appears before the feedback section')
      const faqTitle = collectBackendConfiguredFaqRenderedElements(faqSections[0]!)
        .find(element => readBackendConfiguredFaqRenderedAttribute(element, 'class').split(/\s+/).includes('page-faq__title'))
      assert.equal(faqTitle?.tagName, 'h2')
      assert.equal(readBackendConfiguredFaqRenderedText(faqTitle!).trim(), backendPage.title)
    }
    const faqSectionElements = collectBackendConfiguredFaqRenderedElements(faqSections[0]!)
    const serverRenderedAnswers = faqSectionElements.filter(element => readBackendConfiguredFaqRenderedAttribute(element, 'class').split(/\s+/).includes('page-faq__answer'))
    assert.equal(serverRenderedAnswers.length, backendPage.items.length, `${localizedRoutePath}: every collapsed answer exists in initial HTML`)
    const faqStructuredData = elements.filter(element => element.tagName === 'script'
      && readBackendConfiguredFaqRenderedAttribute(element, 'type') === 'application/ld+json')
      .flatMap(element => {
        const data = JSON.parse(readBackendConfiguredFaqRenderedText(element))
        return data['@graph'] || [data]
      }).filter(data => data['@type'] === 'FAQPage')
    assert.equal(faqStructuredData.length, 1, `${localizedRoutePath}: no duplicate FAQPage schema`)
    assert.equal(faqStructuredData[0].mainEntity.length, backendPage.items.length)
    const questionButtons = faqSectionElements.filter(element => readBackendConfiguredFaqRenderedAttribute(element, 'class').split(/\s+/).includes('page-faq__question'))
    assert.equal(questionButtons.length, backendPage.items.length)
    for (const [index, item] of backendPage.items.entries()) {
      assert.equal(readBackendConfiguredFaqRenderedText(questionButtons[index]!).trim(), item.question)
      assert.equal(readBackendConfiguredFaqRenderedText(serverRenderedAnswers[index]!).replace(/\s+/g, ' ').trim(), normalizeBackendConfiguredFaqRichAnswerText(item.answer))
      assert.equal(faqStructuredData[0].mainEntity[index].name, item.question.trim())
      assert.equal(faqStructuredData[0].mainEntity[index].acceptedAnswer.text.replace(/\s+/g, ''), normalizeBackendConfiguredFaqRichAnswerText(item.answer).replace(/\s+/g, ''), 'JSON-LD retains the complete backend answer despite rich-text spacing')
    }
    console.log(`PASS ${localizedRoutePath}: ${backendPage.items.length} backend questions and complete answers in initial HTML and JSON-LD`)
  }
}

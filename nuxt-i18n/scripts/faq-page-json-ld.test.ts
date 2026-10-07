import assert from 'node:assert/strict'
import { createFaqPageJsonLd, createSeoJsonLdScript } from '../app/utils/seo/jsonLd.ts'

const faqPageJsonLd = createFaqPageJsonLd(
  {
    canonicalUrl: 'https://example.com/guides/tireguides/choose-inner-tube',
    pageId: 'guides-tireguides-choose-inner-tube',
    inLanguage: 'en',
    visibleQuestionsAndAnswers: [
    {
      question: '  How do I check a rim valve hole? ',
      answer: '<p>Measure the <strong>narrowest point</strong>&nbsp;with calipers.</p><ul><li>Do not use the outer countersink.</li></ul>',
    },
    { question: 'Empty answer', answer: '<p> </p>' },
    ],
  },
)

assert.ok(faqPageJsonLd)
assert.equal(faqPageJsonLd['@type'], 'FAQPage')
assert.equal(faqPageJsonLd.inLanguage, 'en')
assert.equal(
  faqPageJsonLd['@id'],
  'https://example.com/guides/tireguides/choose-inner-tube#faq-guides-tireguides-choose-inner-tube',
)
assert.deepEqual(faqPageJsonLd.mainEntity, [
  {
    '@type': 'Question',
    name: 'How do I check a rim valve hole?',
    acceptedAnswer: {
      '@type': 'Answer',
      text: 'Measure the narrowest point with calipers. Do not use the outer countersink.',
    },
  },
])
assert.equal(createFaqPageJsonLd({
  canonicalUrl: 'https://example.com/guide',
  pageId: 'empty-page',
  inLanguage: 'en',
  visibleQuestionsAndAnswers: [],
}), null)

const serializedFaqJsonLd = String(createSeoJsonLdScript(faqPageJsonLd).textContent)
assert.ok(serializedFaqJsonLd.includes('FAQPage'))
assert.ok(!serializedFaqJsonLd.includes('<p>'))

console.log('FAQ page JSON-LD tests passed.')

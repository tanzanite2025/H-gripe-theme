import assert from 'node:assert/strict'
import test from 'node:test'

import {
  normalizeStructuredDataType,
  normalizeStructuredDataTypes,
  normalizeStructuredDataURL,
  snapshotRenderedStructuredData,
} from '../src/structured_data.mjs'

const originalDocument = globalThis.document
const originalWindow = globalThis.window

test('counts explicit FAQ questions and question-like labels', () => {
  const summary = createElement({
    tagName: 'summary',
    text: 'How does the return process work?',
    kind: 'summary',
    parent: createElement({ tagName: 'details', className: 'faq-group' }),
  })
  const button = createElement({
    tagName: 'button',
    text: 'What is the shipping lead time?',
    kind: 'faq-button',
    attrs: { 'data-faq-question': 'true' },
  })
  const questionName = createElement({
    tagName: 'span',
    text: 'Can I change my order after payment?',
    kind: 'microdata-question-name',
    attrs: { itemprop: 'name' },
    parent: createElement({
      tagName: 'div',
      attrs: { itemscope: '', itemtype: 'https://schema.org/Question' },
      className: 'faq-item',
    }),
  })
  const unrelated = createElement({
    tagName: 'h2',
    text: 'About our warranty',
    kind: 'heading',
  })

  setDocument([summary, button, questionName, unrelated])

  const snapshot = snapshotRenderedStructuredData()

  assert.equal(snapshot.page.faqQuestionCount, 3)
})

test('ignores generic headings without FAQ signals', () => {
  const heading = createElement({
    tagName: 'h3',
    text: 'Shipping policy',
    kind: 'heading',
  })
  const paragraph = createElement({
    tagName: 'p',
    text: 'Helpful support text',
    kind: 'paragraph',
  })

  setDocument([heading, paragraph])

  const snapshot = snapshotRenderedStructuredData()

  assert.equal(snapshot.page.faqQuestionCount, 0)
})

test('normalizes structured data types', () => {
  assert.equal(normalizeStructuredDataType('https://schema.org/Product/'), 'Product')
  assert.equal(normalizeStructuredDataType('https://schema.org/Thing#'), 'Thing')
  assert.deepEqual(
    normalizeStructuredDataTypes([
      'https://schema.org/Product/',
      'product',
      'HTTP://schema.org/Thing#',
      'thing',
    ]),
    ['Product', 'Thing'],
  )
})

test('normalizes structured data urls', () => {
  assert.equal(
    normalizeStructuredDataURL(
      'HTTPS://Example.com:443/products/carbon-rim/?utm_source=nav#hero',
      'https://example.com/shop/',
    ),
    'https://example.com/products/carbon-rim',
  )
  assert.equal(
    normalizeStructuredDataURL('/products/carbon-rim/?utm_source=nav#hero', 'https://example.com/shop/'),
    'https://example.com/products/carbon-rim',
  )
})

function setDocument(candidates) {
  globalThis.document = {
    title: 'FAQ snapshot',
    querySelectorAll(selector) {
      if (String(selector).includes('script[type]')) return []
      if (String(selector).includes('[itemscope][itemtype]')) return []
      if (String(selector).includes('[typeof]')) return []
      return candidates
    },
    querySelector(selector) {
      if (String(selector).includes('link[rel~="canonical"]')) return null
      return null
    },
  }
  globalThis.window = {
    getComputedStyle() {
      return {
        display: 'block',
        visibility: 'visible',
      }
    },
  }
}

function createElement({
  tagName,
  text,
  kind,
  attrs = {},
  className = '',
  parent = null,
}) {
  return {
    tagName: String(tagName).toUpperCase(),
    textContent: text,
    className,
    parentElement: parent,
    hidden: false,
    getAttribute(name) {
      if (name === 'class') return className
      if (name === 'id') return attrs.id || ''
      return Object.prototype.hasOwnProperty.call(attrs, name) ? attrs[name] : null
    },
    querySelector() {
      return null
    },
    matches(selector) {
      if (kind === 'summary' && String(selector).includes('details > summary')) {
        return true
      }
      if (kind === 'faq-button' && String(selector).includes('[data-faq-question]')) {
        return true
      }
      if (kind === 'microdata-question-name' && String(selector).includes('[itemprop~="name"]')) {
        return true
      }
      return false
    },
  }
}

test.after(() => {
  globalThis.document = originalDocument
  globalThis.window = originalWindow
})

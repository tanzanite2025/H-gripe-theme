import assert from 'node:assert/strict'
import test from 'node:test'

import { snapshotRenderedHeadings } from '../src/headings.mjs'

const originalDocument = globalThis.document
const originalWindow = globalThis.window
const originalShadowRoot = globalThis.ShadowRoot

test('uses aria-label text for image-only headings', () => {
  setDocument([
    createElement({
      tagName: 'h1',
      attrs: { 'aria-label': 'Tanzanite' },
      children: [
        createElement({ tagName: 'svg' }),
      ],
    }),
  ])

  const headings = snapshotRenderedHeadings()

  assert.equal(headings.length, 1)
  assert.equal(headings[0].level, 1)
  assert.equal(headings[0].text, 'Tanzanite')
})

test('uses nested image alt text for logo headings', () => {
  setDocument([
    createElement({
      tagName: 'h1',
      children: [
        createElement({ tagName: 'img', attrs: { alt: 'Tanzanite Wheels' } }),
      ],
    }),
  ])

  const headings = snapshotRenderedHeadings()

  assert.equal(headings.length, 1)
  assert.equal(headings[0].level, 1)
  assert.equal(headings[0].text, 'Tanzanite Wheels')
})

function setDocument(children) {
  const body = createElement({ tagName: 'body', children })
  globalThis.ShadowRoot = class ShadowRoot {}
  globalThis.document = {
    body,
    documentElement: body,
  }
  globalThis.window = {
    getComputedStyle() {
      return {
        display: 'block',
        visibility: 'visible',
        contentVisibility: 'visible',
      }
    },
  }
}

function createElement({
  tagName,
  text = '',
  attrs = {},
  children = [],
  parent = null,
}) {
  const element = {
    tagName: String(tagName).toUpperCase(),
    textContent: text,
    id: attrs.id || '',
    classList: [],
    children,
    parentElement: parent,
    previousElementSibling: null,
    hidden: false,
    inert: false,
    open: false,
    getAttribute(name) {
      if (name === 'id') return this.id
      return Object.prototype.hasOwnProperty.call(attrs, name) ? attrs[name] : null
    },
    querySelector(selector) {
      if (selector !== 'img[alt]') return null
      return findFirst(this, child => child.tagName === 'IMG' && child.getAttribute('alt') !== null)
    },
    getRootNode() {
      return globalThis.document
    },
  }

  children.forEach((child, index) => {
    child.parentElement = element
    child.previousElementSibling = children[index - 1] || null
  })

  return element
}

function findFirst(element, predicate) {
  for (const child of element.children || []) {
    if (predicate(child)) return child
    const nested = findFirst(child, predicate)
    if (nested) return nested
  }
  return null
}

test.after(() => {
  globalThis.document = originalDocument
  globalThis.window = originalWindow
  globalThis.ShadowRoot = originalShadowRoot
})

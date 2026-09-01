export function normalizeStructuredDataType(value) {
  const raw = String(value || '').replace(/\s+/g, ' ').trim()
  if (!raw) {
    return ''
  }
  const trimmed = raw.replace(/[\/#]+$/g, '')
  const hashIndex = trimmed.lastIndexOf('#')
  const slashIndex = trimmed.lastIndexOf('/')
  const start = Math.max(hashIndex, slashIndex)
  return start >= 0 ? trimmed.slice(start + 1) : trimmed
}

export function normalizeStructuredDataTypes(value) {
  const values = Array.isArray(value)
    ? value
    : String(value || '')
      .split(/\s+/)
      .filter(Boolean)
  const result = []
  const seen = new Set()
  for (const item of values) {
    const normalized = normalizeStructuredDataType(item)
    if (!normalized) {
      continue
    }
    const key = normalized.toLowerCase()
    if (seen.has(key)) {
      continue
    }
    seen.add(key)
    result.push(normalized)
  }
  return result
}

export function normalizeStructuredDataURL(value, baseURL = typeof document !== 'undefined' ? document.baseURI : '') {
  const raw = String(value || '').trim()
  if (!raw) {
    return ''
  }
  try {
    const resolved = baseURL ? new URL(raw, baseURL) : new URL(raw)
    if (resolved.protocol !== 'http:' && resolved.protocol !== 'https:') {
      return raw
    }
    resolved.hash = ''
    resolved.search = ''
    if (!resolved.pathname) {
      resolved.pathname = '/'
    } else if (resolved.pathname !== '/') {
      resolved.pathname = resolved.pathname.replace(/\/+$/, '') || '/'
    }
    return resolved.href
  } catch {
    return raw
  }
}

export function snapshotRenderedStructuredData() {
  const maxJsonLdScripts = 80
  const maxJsonLdNodes = 240
  const maxSurfaceRecords = 120
  const maxSnippetLength = 900
  const maxRawLength = 6_000
  const maxSelectorDepth = 12

  const jsonLd = []
  const scripts = Array.from(document.querySelectorAll('script[type]'))
    .filter((script) => isJsonLdScriptType(script.getAttribute('type') || script.type))
    .slice(0, maxJsonLdScripts)

  scripts.forEach((script, index) => {
    const rawText = String(script.textContent || '').trim()
    const entry = {
      index,
      selector: domSelector(script),
      raw: truncate(rawText, maxRawLength),
      parseError: '',
      nodes: [],
    }

    if (!rawText) {
      entry.parseError = 'JSON-LD script is empty.'
      jsonLd.push(entry)
      return
    }

    try {
      const parsed = JSON.parse(rawText)
      const seen = new WeakSet()
      collectJsonLdNodes(parsed, `script[${index}]`, entry.nodes, seen)
    } catch (error) {
      entry.parseError = error instanceof Error
        ? normalizeSpace(error.message)
        : 'JSON-LD script could not be parsed.'
    }

    jsonLd.push(entry)
  })

  return {
    page: {
      title: normalizeSpace(document.title || ''),
      canonicalUrl: canonicalURL(),
      faqQuestionCount: visibleFAQQuestionCount(),
      productSignal: Boolean(document.querySelector('.product-page, [data-product-id], [itemtype*="Product"]')),
    },
    jsonLd,
    microdata: snapshotMicrodata(),
    rdfa: snapshotRDFa(),
  }

  function isJsonLdScriptType(value) {
    return String(value || '')
      .trim()
      .toLowerCase()
      .split(';')[0]
      .trim() === 'application/ld+json'
  }

  function collectJsonLdNodes(value, path, nodes, seen) {
    if (nodes.length >= maxJsonLdNodes || value == null) {
      return
    }
    if (Array.isArray(value)) {
      value.forEach((item, index) => collectJsonLdNodes(item, `${path}[${index}]`, nodes, seen))
      return
    }
    if (typeof value !== 'object') {
      return
    }
    if (seen.has(value)) {
      return
    }
    seen.add(value)

    const types = normalizeStructuredDataTypes(value['@type'])
    if (types.length > 0) {
      nodes.push({
        types,
        type: types[0],
        id: stringValue(value['@id']),
        name: stringValue(value.name),
        url: normalizeStructuredDataURL(value.url),
        graphPath: path,
        data: compactJSONValue(value, 0),
      })
    }

    if (Array.isArray(value['@graph'])) {
      value['@graph'].forEach((item, index) => {
        collectJsonLdNodes(item, `${path}.@graph[${index}]`, nodes, seen)
      })
    }

    for (const key of ['mainEntity', 'itemListElement', 'hasVariant', 'item', 'offers', 'brand', 'aggregateRating', 'acceptedAnswer', 'address', 'geo']) {
      if (Object.prototype.hasOwnProperty.call(value, key)) {
        collectJsonLdNodes(value[key], `${path}.${key}`, nodes, seen)
      }
    }
  }

  function snapshotMicrodata() {
    return Array.from(document.querySelectorAll('[itemscope][itemtype]'))
      .filter((element) => !element.parentElement?.closest('[itemscope]'))
      .slice(0, maxSurfaceRecords)
      .map((element) => {
        const types = normalizeStructuredDataTypes(element.getAttribute('itemtype'))
        return {
          format: 'microdata',
          types,
          type: types[0] || '',
          id: element.getAttribute('itemid') || element.id || '',
          name: itemPropText(element, 'name'),
          url: normalizeStructuredDataURL(itemPropText(element, 'url') || element.getAttribute('itemid') || ''),
          selector: domSelector(element),
          snippet: elementSnippet(element),
        }
      })
      .filter((item) => item.type || item.name || item.url)
  }

  function snapshotRDFa() {
    return Array.from(document.querySelectorAll('[typeof]'))
      .slice(0, maxSurfaceRecords)
      .map((element) => {
        const types = normalizeStructuredDataTypes(element.getAttribute('typeof'))
        return {
          format: 'rdfa',
          types,
          type: types[0] || '',
          id: element.getAttribute('about') || element.id || '',
          name: propertyText(element, 'name'),
          url: normalizeStructuredDataURL(element.getAttribute('resource') || element.getAttribute('href') || ''),
          selector: domSelector(element),
          snippet: elementSnippet(element),
        }
      })
      .filter((item) => item.type || item.name || item.url)
  }

  function itemPropText(root, propName) {
    const element = root.querySelector(`[itemprop~="${cssString(propName)}"]`)
    if (!element) {
      return ''
    }
    return normalizeSpace(element.getAttribute('content') || element.getAttribute('href') || element.getAttribute('src') || element.textContent || '')
  }

  function propertyText(root, propName) {
    const element = root.querySelector(`[property~="${cssString(propName)}"]`)
    if (!element) {
      return ''
    }
    return normalizeSpace(element.getAttribute('content') || element.getAttribute('href') || element.textContent || '')
  }

  function visibleFAQQuestionCount() {
    const candidates = Array.from(document.querySelectorAll([
      'details > summary',
      '[data-faq-question]',
      '[aria-controls*="faq" i]',
      '[class*="faq" i] [class*="question" i]',
      '[class*="faq" i] button[aria-expanded]',
      '[class*="faq" i] [role="button"][aria-expanded]',
      '[itemscope][itemtype*="Question" i] [itemprop~="name"]',
      '[itemtype*="Question" i] [itemprop~="name"]',
      '[aria-expanded]',
      'h2',
      'h3',
    ].join(', ')))
    const seen = new Set()
    let count = 0
    for (const element of candidates) {
      const text = normalizeSpace(element.textContent || '')
      if (!text || text.length < 8 || !isElementVisible(element)) {
        continue
      }
      const explicitFAQQuestion = isExplicitFAQQuestion(element)
      if (!explicitFAQQuestion && !looksLikeQuestionText(text)) {
        continue
      }
      const key = `${text}\u0000${element.getAttribute('aria-controls') || ''}`
      if (seen.has(key)) {
        continue
      }
      seen.add(key)
      count++
    }
    return Math.min(count, 200)
  }

  function isExplicitFAQQuestion(element) {
    if (
      element.matches('details > summary, [data-faq-question], [aria-controls*="faq" i], [itemscope][itemtype*="Question" i] [itemprop~="name"], [itemtype*="Question" i] [itemprop~="name"]')
    ) {
      return true
    }
    for (let current = element; current; current = current.parentElement) {
      const marker = [
        current.id,
        current.className,
        current.getAttribute('data-testid'),
        current.getAttribute('data-component'),
      ].map(value => String(value || '')).join(' ')
      if (/\bfaq\b|faq[-_ ]|[-_ ]faq|question[-_ ]|[-_ ]question/i.test(marker)) {
        return true
      }
    }
    return false
  }

  function looksLikeQuestionText(text) {
    return /\?$|？$|^(how|what|when|where|why|can|do|does|is|are|will|which)\b/i.test(text)
  }

  function canonicalURL() {
    const link = document.querySelector('link[rel~="canonical"]')
    return normalizeStructuredDataURL(link?.href || '')
  }

  function normalizeTypes(value) {
    return normalizeStructuredDataTypes(value)
  }

  function normalizeSchemaType(value) {
    return normalizeStructuredDataType(value)
  }

  function compactJSONValue(value, depth) {
    if (value == null || typeof value === 'number' || typeof value === 'boolean') {
      return value
    }
    if (typeof value === 'string') {
      return truncate(normalizeSpace(value), 700)
    }
    if (Array.isArray(value)) {
      if (depth >= 6) {
        return value.length ? [`[${value.length} items]`] : []
      }
      return value.slice(0, 40).map((item) => compactJSONValue(item, depth + 1))
    }
    if (typeof value === 'object') {
      if (depth >= 6) {
        return { summary: 'object depth limit reached' }
      }
      const result = {}
      for (const key of Object.keys(value).slice(0, 80)) {
        result[key] = compactJSONValue(value[key], depth + 1)
      }
      return result
    }
    return String(value)
  }

  function isElementVisible(element) {
    for (let current = element; current; current = current.parentElement) {
      if (current.hidden || current.getAttribute('aria-hidden') === 'true') {
        return false
      }
      const style = window.getComputedStyle(current)
      if (style.display === 'none' || style.visibility === 'hidden' || style.visibility === 'collapse') {
        return false
      }
    }
    return true
  }

  function domSelector(element) {
    const parts = []
    let current = element
    while (current && current.nodeType === Node.ELEMENT_NODE && parts.length < maxSelectorDepth) {
      parts.unshift(selectorPart(current))
      current = current.parentElement
    }
    return parts.join(' > ')
  }

  function selectorPart(element) {
    const tagName = element.tagName.toLowerCase()
    if (element.id) {
      return `${tagName}#${cssIdentifier(element.id)}`
    }
    let index = 1
    let sibling = element.previousElementSibling
    while (sibling) {
      if (sibling.tagName === element.tagName) {
        index++
      }
      sibling = sibling.previousElementSibling
    }
    return `${tagName}:nth-of-type(${index})`
  }

  function elementSnippet(element) {
    const tagName = element.tagName.toLowerCase()
    const attrs = []
    for (const name of ['id', 'class', 'itemtype', 'typeof']) {
      const value = element.getAttribute(name)
      if (value) {
        attrs.push(`${name}="${escapeHTMLAttribute(value)}"`)
      }
    }
    const openTag = attrs.length ? `<${tagName} ${attrs.join(' ')}>` : `<${tagName}>`
    return truncate(`${openTag}${escapeHTMLText(normalizeSpace(element.textContent || ''))}</${tagName}>`, maxSnippetLength)
  }

  function stringValue(value) {
    if (typeof value === 'string') {
      return normalizeSpace(value)
    }
    if (typeof value === 'number' || typeof value === 'boolean') {
      return String(value)
    }
    return ''
  }

  function truncate(value, maxLength) {
    const normalized = String(value || '')
    if (normalized.length <= maxLength) {
      return normalized
    }
    return `${normalized.slice(0, maxLength - 3)}...`
  }

  function normalizeSpace(value) {
    return String(value || '').replace(/\s+/g, ' ').trim()
  }

  function cssIdentifier(value) {
    return String(value || '').trim().replace(/[^a-zA-Z0-9_-]/g, '-') || 'unknown'
  }

  function cssString(value) {
    return String(value || '').replace(/\\/g, '\\\\').replace(/"/g, '\\"')
  }

  function escapeHTMLText(value) {
    return String(value)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
  }

  function escapeHTMLAttribute(value) {
    return escapeHTMLText(value).replace(/"/g, '&quot;')
  }
}

import { h, type VNode, type VNodeChild } from 'vue'
import { parseFragment } from 'parse5'

interface ParsedHtmlNode {
  nodeName?: string
  tagName?: string
  value?: string
  attrs?: Array<{
    name: string
    value: string
  }>
  childNodes?: ParsedHtmlNode[]
}

export interface SafeRichTextOptions {
  mediaOrigins?: Iterable<string | null | undefined>
}

interface SafeRichTextRuntimeConfig {
  public?: {
    apiBase?: unknown
    siteUrl?: unknown
  }
}

interface RenderContext {
  mediaOrigins: ReadonlySet<string>
}

const allowedTags = new Set([
  'a',
  'address',
  'article',
  'aside',
  'b',
  'blockquote',
  'br',
  'code',
  'del',
  'details',
  'dd',
  'div',
  'dl',
  'dt',
  'em',
  'footer',
  'figcaption',
  'figure',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'header',
  'hr',
  'i',
  'img',
  'li',
  'main',
  'mark',
  'ol',
  'p',
  'pre',
  'q',
  's',
  'small',
  'source',
  'span',
  'strong',
  'sub',
  'summary',
  'sup',
  'table',
  'tbody',
  'td',
  'tfoot',
  'th',
  'thead',
  'tr',
  'u',
  'ul',
  'video',
])

const discardedContentTags = new Set([
  'base',
  'embed',
  'iframe',
  'link',
  'meta',
  'object',
  'script',
  'style',
  'template',
])

const globalAttributes = new Set([
  'dir',
  'lang',
  'role',
  'title',
])

const tagAttributes: Record<string, Set<string>> = {
  a: new Set(['href', 'rel', 'target']),
  blockquote: new Set(['cite']),
  img: new Set(['alt', 'decoding', 'height', 'loading', 'src', 'width']),
  li: new Set(['value']),
  ol: new Set(['reversed', 'start']),
  q: new Set(['cite']),
  source: new Set(['media', 'src', 'type']),
  td: new Set(['colspan', 'rowspan']),
  th: new Set(['colspan', 'rowspan', 'scope']),
  video: new Set([
    'autoplay',
    'controls',
    'height',
    'loop',
    'muted',
    'playsinline',
    'poster',
    'preload',
    'src',
    'width',
  ]),
}

const booleanAttributes = new Set([
  'autoplay',
  'controls',
  'loop',
  'muted',
  'playsinline',
  'reversed',
])

const allowedLinkProtocols = new Set(['http:', 'https:', 'mailto:', 'tel:'])
const allowedMediaProtocols = new Set(['http:', 'https:'])
const urlBaseOrigin = 'https://storefront.invalid'
const urlBase = `${urlBaseOrigin}/`
const rasterImageDataUrlPattern = /^data:image\/(?:avif|gif|jpe?g|png|webp);base64,[a-z0-9+/]+={0,2}$/i

const uniqueOrigins = (values: Array<string | null>): string[] => (
  [...new Set(values.filter((value): value is string => Boolean(value)))]
)

const normalizedTrustedOrigin = (value: unknown): string | null => {
  const candidate = String(value || '').trim()
  if (!candidate) return null

  try {
    const url = new URL(candidate)
    if (!['http:', 'https:'].includes(url.protocol)) return null
    if (url.username || url.password || url.hostname.includes('*')) return null
    return url.origin
  } catch {
    return null
  }
}

export const safeRichTextMediaOriginsFromRuntimeConfig = (
  runtimeConfig: SafeRichTextRuntimeConfig,
  browserOrigin = '',
): string[] => {
  const publicConfig = runtimeConfig.public || {}

  return uniqueOrigins([
    normalizedTrustedOrigin(publicConfig.siteUrl),
    normalizedTrustedOrigin(publicConfig.apiBase),
    normalizedTrustedOrigin(browserOrigin),
  ])
}

const renderContext = (options: SafeRichTextOptions = {}): RenderContext => {
  const mediaOrigins = [...(options.mediaOrigins || [])]
    .map(normalizedTrustedOrigin)
    .filter((origin): origin is string => Boolean(origin))

  return {
    mediaOrigins: new Set(mediaOrigins),
  }
}

const sanitizedClassName = (value: string): string | null => {
  const classes = value
    .split(/\s+/)
    .filter(className => /^[A-Za-z0-9_-]{1,64}$/.test(className))

  return classes.length > 0 ? classes.join(' ') : null
}

const sanitizedUrl = (
  value: string,
  allowedProtocols: ReadonlySet<string>,
  context: RenderContext,
  options: {
    allowImageDataUrl?: boolean
    requireTrustedOrigin?: boolean
  } = {},
): string | null => {
  const candidate = value.trim()
  if (!candidate || /[\u0000-\u001F\u007F]/.test(candidate)) return null
  if (options.allowImageDataUrl && rasterImageDataUrlPattern.test(candidate)) return candidate

  try {
    const url = new URL(candidate, urlBase)
    if (!allowedProtocols.has(url.protocol)) return null
    if (url.username || url.password) return null

    if (url.origin === urlBaseOrigin) {
      return `${url.pathname}${url.search}${url.hash}`
    }

    if (options.requireTrustedOrigin && !context.mediaOrigins.has(url.origin)) {
      return null
    }

    return url.href
  } catch {
    return null
  }
}

const sanitizedAttributeValue = (
  tagName: string,
  name: string,
  value: string,
  context: RenderContext,
): string | boolean | null => {
  const normalized = value.trim()

  if (booleanAttributes.has(name)) return true
  if (name === 'class') return sanitizedClassName(value)
  if (name === 'href' || name === 'cite') return sanitizedUrl(normalized, allowedLinkProtocols, context)
  if (name === 'src' || name === 'poster') {
    return sanitizedUrl(
      normalized,
      allowedMediaProtocols,
      context,
      {
        allowImageDataUrl: tagName === 'img' && name === 'src',
        requireTrustedOrigin: true,
      },
    )
  }
  if (name === 'target') return normalized === '_blank' ? '_blank' : null
  if (name === 'rel') return null
  if (name === 'alt') return normalized.slice(0, 500) || ''
  if (name === 'loading') return normalized === 'lazy' || normalized === 'eager' ? normalized : null
  if (name === 'decoding') return ['async', 'auto', 'sync'].includes(normalized) ? normalized : null
  if (name === 'preload') return ['auto', 'metadata', 'none'].includes(normalized) ? normalized : null
  if (name === 'scope') return ['col', 'colgroup', 'row', 'rowgroup'].includes(normalized) ? normalized : null
  if (name === 'type' || name === 'media') return normalized.slice(0, 160) || null
  if (name === 'width' || name === 'height' || name === 'colspan' || name === 'rowspan' || name === 'start' || name === 'value') {
    return /^\d{1,4}$/.test(normalized) ? normalized : null
  }
  if (name.startsWith('aria-') || globalAttributes.has(name)) return normalized.slice(0, 500) || null

  return null
}

const attributesForNode = (
  node: ParsedHtmlNode,
  tagName: string,
  context: RenderContext,
): Record<string, string | boolean> => {
  const attributes: Record<string, string | boolean> = {}
  const allowedAttributes = tagAttributes[tagName] || new Set<string>()
  let opensNewWindow = false

  for (const attribute of node.attrs || []) {
    const name = attribute.name.toLowerCase()
    const isAllowed = name === 'class'
      || name.startsWith('aria-')
      || globalAttributes.has(name)
      || allowedAttributes.has(name)
    if (!isAllowed) continue

    const sanitized = sanitizedAttributeValue(tagName, name, attribute.value, context)
    if (sanitized === null) continue

    attributes[name] = sanitized
    if (tagName === 'a' && name === 'target' && sanitized === '_blank') {
      opensNewWindow = true
    }
  }

  if (opensNewWindow) attributes.rel = 'noopener noreferrer'
  return attributes
}

const renderNode = (node: ParsedHtmlNode, context: RenderContext): VNodeChild | null => {
  if (node.nodeName === '#text') return node.value || ''

  const tagName = String(node.tagName || '').toLowerCase()
  if (discardedContentTags.has(tagName)) return null
  if (!allowedTags.has(tagName)) {
    const children = (node.childNodes || [])
      .map(child => renderNode(child, context))
      .filter((child): child is VNodeChild => child !== null)
    return children.length > 0 ? h('span', {}, children) : null
  }

  const children = (node.childNodes || [])
    .map(child => renderNode(child, context))
    .filter((child): child is VNodeChild => child !== null)
  return h(tagName, attributesForNode(node, tagName, context), children) as VNode
}

export const renderRichText = (html: string, options: SafeRichTextOptions = {}): VNodeChild[] => {
  const fragment = parseFragment(html) as unknown as ParsedHtmlNode
  const context = renderContext(options)
  return (fragment.childNodes || [])
    .map(node => renderNode(node, context))
    .filter((node): node is VNodeChild => node !== null)
}

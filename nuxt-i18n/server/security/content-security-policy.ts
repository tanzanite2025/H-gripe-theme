import { createHash, randomBytes } from 'node:crypto'
import { parse, serialize } from 'parse5'

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

interface InlineContentHashes {
  script: string[]
  style: string[]
}

interface ResourceOrigins {
  font: string[]
  frame: string[]
  image: string[]
  media: string[]
  script: string[]
  style: string[]
}

interface ContentSecurityPolicyOptions {
  scriptNonce?: string
}

interface SecuredHtmlResponse {
  body: string
  contentSecurityPolicy: string
}

const CSP_CACHE_LIMIT = 512
const cspByDocumentHash = new Map<string, string>()
const SCRIPT_NONCE_BYTES = 18

const executableScriptTypes = new Set([
  '',
  'application/ecmascript',
  'application/javascript',
  'application/x-ecmascript',
  'application/x-javascript',
  'module',
  'text/ecmascript',
  'text/javascript',
  'text/jsx',
])

const staticScriptSources = [
  'https://accounts.google.com',
  'https://challenges.cloudflare.com',
  'https://*.js.stripe.com',
  'https://js.stripe.com',
  'https://www.googletagmanager.com',
]

const staticConnectSources = [
  'https://accounts.google.com',
  'https://api.stripe.com',
  'https://challenges.cloudflare.com',
  'https://www.google-analytics.com',
  'https://www.googletagmanager.com',
]

const staticFrameSources = [
  'https://accounts.google.com',
  'https://challenges.cloudflare.com',
  'https://hooks.stripe.com',
  'https://*.js.stripe.com',
  'https://js.stripe.com',
  'https://maps.google.com',
  'https://www.google.com',
]

const environment = process.env

const normalizeOrigin = (value: string | undefined): string | null => {
  const candidate = String(value || '').trim()
  if (!candidate) return null

  try {
    const url = new URL(candidate)
    if (!['http:', 'https:'].includes(url.protocol)) return null
    return url.origin
  } catch {
    return null
  }
}

const normalizeConfiguredSource = (
  value: string,
  allowedSpecialSources: ReadonlySet<string>,
): string | null => {
  const candidate = value.trim()
  if (!candidate) return null
  if (candidate === "'self'" || allowedSpecialSources.has(candidate)) return candidate

  try {
    const url = new URL(candidate)
    if (!['http:', 'https:'].includes(url.protocol)) return null
    if (url.username || url.password || url.hostname.includes('*') || url.pathname !== '/' || url.search || url.hash) return null
    return url.origin
  } catch {
    return null
  }
}

const configuredSources = (
  variableName: string,
  allowedSpecialSources = new Set<string>(),
): string[] => (
  String(environment[variableName] || '')
    .split(',')
    .map(value => normalizeConfiguredSource(value, allowedSpecialSources))
    .filter((value): value is string => Boolean(value))
)

const publicSiteOrigins = [
  normalizeOrigin(environment.NUXT_SITE_URL),
  normalizeOrigin(environment.NUXT_PUBLIC_SITE_URL),
].filter((value): value is string => Boolean(value))

const publicApiOrigins = [
  normalizeOrigin(environment.NUXT_PUBLIC_API_BASE),
  normalizeOrigin(environment.GO_API_BASE),
  normalizeOrigin(environment.API_BASE),
].filter((value): value is string => Boolean(value))

const cdnOrigins = [
  normalizeOrigin(environment.CDN_URL),
].filter((value): value is string => Boolean(value))

const uniqueSources = (...sources: Array<string | null | undefined | string[]>): string[] => {
  const flattened = sources.flatMap(source => Array.isArray(source) ? source : [source])
  return [...new Set(flattened.filter((source): source is string => Boolean(source)))]
}

const attributeValue = (node: ParsedHtmlNode, attributeName: string): string => (
  node.attrs?.find(attribute => attribute.name.toLowerCase() === attributeName)?.value || ''
)

const setAttributeValue = (node: ParsedHtmlNode, attributeName: string, value: string): void => {
  const attrs = node.attrs || []
  const existing = attrs.find(attribute => attribute.name.toLowerCase() === attributeName)
  if (existing) {
    existing.value = value
  } else {
    attrs.push({ name: attributeName, value })
  }
  node.attrs = attrs
}

const textContent = (node: ParsedHtmlNode): string => (
  node.childNodes?.map(child => child.value || textContent(child)).join('') || ''
)

const sha256Source = (content: string): string => (
  `'sha256-${createHash('sha256').update(content, 'utf8').digest('base64')}'`
)

const originFromUrl = (value: string): string | null => {
  const candidate = value.trim()
  if (!candidate) return null

  try {
    const url = new URL(candidate)
    if (!['http:', 'https:'].includes(url.protocol)) return null
    return url.origin
  } catch {
    return null
  }
}

const addResourceOrigin = (origins: Set<string>, value: string): void => {
  const origin = originFromUrl(value)
  if (origin) origins.add(origin)
}

const hasRel = (node: ParsedHtmlNode, relation: string): boolean => (
  attributeValue(node, 'rel')
    .split(/\s+/)
    .some(value => value.toLowerCase() === relation)
)

const walkNodes = (
  nodes: ParsedHtmlNode[] | undefined,
  hashes: {
    script: Set<string>
    style: Set<string>
  },
  resourceOrigins: {
    font: Set<string>
    frame: Set<string>
    image: Set<string>
    media: Set<string>
    script: Set<string>
    style: Set<string>
  },
  options: {
    scriptNonce?: string
  } = {},
): void => {
  for (const node of nodes || []) {
    const tagName = String(node.tagName || '').toLowerCase()

    if (tagName === 'style') {
      const content = textContent(node)
      if (content) hashes.style.add(sha256Source(content))
    } else if (tagName === 'script') {
      if (options.scriptNonce) setAttributeValue(node, 'nonce', options.scriptNonce)

      const source = attributeValue(node, 'src')
      if (source) {
        addResourceOrigin(resourceOrigins.script, source)
      } else {
        const type = attributeValue(node, 'type').trim().toLowerCase()
        if (executableScriptTypes.has(type)) {
          const content = textContent(node)
          if (content) hashes.script.add(sha256Source(content))
        }
      }
    } else if (tagName === 'img') {
      addResourceOrigin(resourceOrigins.image, attributeValue(node, 'src'))
    } else if (tagName === 'video' || tagName === 'audio') {
      addResourceOrigin(resourceOrigins.media, attributeValue(node, 'src'))
      addResourceOrigin(resourceOrigins.image, attributeValue(node, 'poster'))
    } else if (tagName === 'source' || tagName === 'track') {
      addResourceOrigin(resourceOrigins.media, attributeValue(node, 'src'))
    } else if (tagName === 'iframe') {
      addResourceOrigin(resourceOrigins.frame, attributeValue(node, 'src'))
    } else if (tagName === 'link') {
      const href = attributeValue(node, 'href')
      if (hasRel(node, 'stylesheet')) addResourceOrigin(resourceOrigins.style, href)
      if (hasRel(node, 'icon')) addResourceOrigin(resourceOrigins.image, href)
      if (hasRel(node, 'preload')) {
        const as = attributeValue(node, 'as').toLowerCase()
        if (as === 'font') addResourceOrigin(resourceOrigins.font, href)
        if (as === 'image') addResourceOrigin(resourceOrigins.image, href)
        if (as === 'script') addResourceOrigin(resourceOrigins.script, href)
        if (as === 'style') addResourceOrigin(resourceOrigins.style, href)
      }
    }

    walkNodes(node.childNodes, hashes, resourceOrigins, options)
  }
}

export const collectInlineContentHashes = (html: string): InlineContentHashes => {
  const document = parse(html) as unknown as ParsedHtmlNode
  const hashes = {
    script: new Set<string>(),
    style: new Set<string>(),
  }

  walkNodes(document.childNodes, hashes, {
    font: new Set<string>(),
    frame: new Set<string>(),
    image: new Set<string>(),
    media: new Set<string>(),
    script: new Set<string>(),
    style: new Set<string>(),
  })

  return {
    script: [...hashes.script].sort(),
    style: [...hashes.style].sort(),
  }
}

export const collectResourceOrigins = (html: string): ResourceOrigins => {
  const document = parse(html) as unknown as ParsedHtmlNode
  const hashes = {
    script: new Set<string>(),
    style: new Set<string>(),
  }
  const origins = {
    font: new Set<string>(),
    frame: new Set<string>(),
    image: new Set<string>(),
    media: new Set<string>(),
    script: new Set<string>(),
    style: new Set<string>(),
  }

  walkNodes(document.childNodes, hashes, origins)

  return {
    font: [...origins.font].sort(),
    frame: [...origins.frame].sort(),
    image: [...origins.image].sort(),
    media: [...origins.media].sort(),
    script: [...origins.script].sort(),
    style: [...origins.style].sort(),
  }
}

const directive = (name: string, sources: string[]): string => `${name} ${sources.join(' ')}`

export const createScriptNonce = (): string => randomBytes(SCRIPT_NONCE_BYTES).toString('base64')

export const applyScriptNonce = (html: string, scriptNonce: string): string => {
  const document = parse(html) as unknown as ParsedHtmlNode
  const hashes = {
    script: new Set<string>(),
    style: new Set<string>(),
  }
  const origins = {
    font: new Set<string>(),
    frame: new Set<string>(),
    image: new Set<string>(),
    media: new Set<string>(),
    script: new Set<string>(),
    style: new Set<string>(),
  }

  walkNodes(document.childNodes, hashes, origins, { scriptNonce })
  return serialize(document as any)
}

export const createContentSecurityPolicy = (
  html: string,
  options: ContentSecurityPolicyOptions = {},
): string => {
  const documentHash = createHash('sha256').update(html, 'utf8').digest('base64')
  const cached = options.scriptNonce ? undefined : cspByDocumentHash.get(documentHash)
  if (cached) return cached

  const hashes = collectInlineContentHashes(html)
  const assetSources = uniqueSources(publicSiteOrigins, publicApiOrigins, cdnOrigins, configuredSources('NUXT_CSP_ASSET_SRC'))
  const imageSources = uniqueSources(assetSources, configuredSources('NUXT_CSP_IMG_SRC', new Set(['data:', 'blob:'])))
  const mediaSources = uniqueSources(assetSources, configuredSources('NUXT_CSP_MEDIA_SRC', new Set(['blob:'])))
  const connectSources = uniqueSources(
    staticConnectSources,
    publicApiOrigins,
    publicSiteOrigins,
    configuredSources('NUXT_CSP_CONNECT_SRC'),
  )
  const fontSources = uniqueSources(assetSources, configuredSources('NUXT_CSP_FONT_SRC', new Set(['data:'])))
  const formActionSources = uniqueSources(
    ['https://checkout.stripe.com'],
    configuredSources('NUXT_CSP_FORM_ACTION_SRC'),
  )
  const scriptTrustSources = uniqueSources(
    hashes.script,
    options.scriptNonce ? [`'nonce-${options.scriptNonce}'`, "'strict-dynamic'"] : [],
  )
  const scriptFallbackSources = uniqueSources(staticScriptSources, cdnOrigins, configuredSources('NUXT_CSP_SCRIPT_SRC'))
  const scriptSources = uniqueSources(scriptTrustSources, "'self'", scriptFallbackSources)
  const styleSources = uniqueSources(cdnOrigins, configuredSources('NUXT_CSP_STYLE_SRC'), hashes.style)
  const frameSources = uniqueSources(staticFrameSources, configuredSources('NUXT_CSP_FRAME_SRC'))
  // Vite injects component-scoped styles through runtime <style> elements in
  // development. Keep production hash-only, but allow that dev-only mechanism
  // so layout CSS is not silently dropped during local visual verification.
  const developmentStyleSources = environment.NODE_ENV === 'development'
    ? ["'unsafe-inline'"]
    : []

  const policy = [
    directive('default-src', ["'self'"]),
    directive('base-uri', ["'self'"]),
    directive('object-src', ["'none'"]),
    directive('frame-ancestors', ["'self'"]),
    directive('form-action', ["'self'", ...formActionSources]),
    directive('script-src', scriptSources),
    directive('script-src-elem', scriptSources),
    directive('script-src-attr', ["'none'"]),
    directive('style-src', ["'self'", ...developmentStyleSources, ...styleSources]),
    directive('style-src-elem', ["'self'", ...developmentStyleSources, ...styleSources]),
    directive('style-src-attr', ["'unsafe-inline'"]),
    directive('img-src', ["'self'", 'data:', 'blob:', ...imageSources]),
    directive('font-src', ["'self'", 'data:', ...fontSources]),
    directive('connect-src', ["'self'", ...connectSources]),
    directive('media-src', ["'self'", 'blob:', ...mediaSources]),
    directive('frame-src', frameSources),
    directive('worker-src', ["'self'", 'blob:']),
    directive('manifest-src', ["'self'"]),
    directive('trusted-types', ['default', 'vue', 'tanzanite-script-url']),
    directive('require-trusted-types-for', ["'script'"]),
  ]

  if (environment.NODE_ENV === 'production') {
    policy.push('upgrade-insecure-requests')
  }

  const value = policy.join('; ')
  if (!options.scriptNonce) {
    if (cspByDocumentHash.size >= CSP_CACHE_LIMIT) {
      const oldestDocumentHash = cspByDocumentHash.keys().next().value
      if (oldestDocumentHash) cspByDocumentHash.delete(oldestDocumentHash)
    }
    cspByDocumentHash.set(documentHash, value)
  }
  return value
}

export const secureHtmlWithContentSecurityPolicy = (html: string): SecuredHtmlResponse => {
  const scriptNonce = createScriptNonce()
  const body = applyScriptNonce(html, scriptNonce)

  return {
    body,
    contentSecurityPolicy: createContentSecurityPolicy(body, { scriptNonce }),
  }
}

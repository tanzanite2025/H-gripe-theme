import { useRuntimeConfig } from 'nuxt/app'
import { attachDeviceFingerprintHeader } from '~/utils/deviceFingerprint'

export type MaybeJson = Record<string, unknown> | string | null
export type ApiRequestQueryPrimitive = string | number | boolean | Date
export type ApiRequestQueryValue = ApiRequestQueryPrimitive | ApiRequestQueryPrimitive[] | Record<string, unknown> | null | undefined
export type ApiRequestQuery = Record<string, ApiRequestQueryValue> | URLSearchParams

export interface ApiRequestInit extends RequestInit {
  params?: ApiRequestQuery
  query?: ApiRequestQuery
}

export type ApiRequestFunction = <T = MaybeJson>(path: string, init?: ApiRequestInit, fallbackMessage?: string) => Promise<T>

export class ApiRequestError extends Error {
  status: number
  code?: string
  details?: unknown
  payload: MaybeJson

  constructor(message: string, options: { status: number; code?: string; details?: unknown; payload: MaybeJson }) {
    super(message)
    this.name = 'ApiRequestError'
    this.status = options.status
    this.code = options.code
    this.details = options.details
    this.payload = options.payload
  }
}

const defaultCredentials: RequestCredentials = 'include'
const csrfCookieName = 'csrf_token'
const csrfHeaderName = 'X-CSRF-Token'

const isUnsafeMethod = (method?: string) => !['GET', 'HEAD', 'OPTIONS', 'TRACE'].includes((method || 'GET').toUpperCase())

const readCookie = (name: string) => {
  if (typeof document === 'undefined') {
    return ''
  }
  const prefix = `${encodeURIComponent(name)}=`
  const cookie = document.cookie
    .split(';')
    .map((item) => item.trim())
    .find((item) => item.startsWith(prefix))
  if (!cookie) {
    return ''
  }
  try {
    return decodeURIComponent(cookie.slice(prefix.length))
  } catch {
    return ''
  }
}

const firstHeaderValue = (...values: Array<string | undefined>) => {
  for (const value of values) {
    const trimmed = String(value || '').trim()
    if (trimmed) return trimmed
  }
  return ''
}

const clientLocaleHeader = () => {
  if (!import.meta.client) return ''
  return firstHeaderValue(
    readCookie('locale'),
    readCookie('i18n_redirected'),
    typeof navigator !== 'undefined' ? navigator.language : '',
  )
}

const clientTimezoneHeader = () => {
  if (!import.meta.client) return ''
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || ''
  } catch {
    return ''
  }
}

const attachClientContextHeaders = (headers: Headers) => {
  if (!import.meta.client) return headers

  const locale = clientLocaleHeader()
  if (locale) {
    if (!headers.has('X-Locale')) {
      headers.set('X-Locale', locale)
    }
    if (!headers.has('Accept-Language')) {
      headers.set('Accept-Language', locale)
    }
  }

  const timezone = clientTimezoneHeader()
  if (timezone && !headers.has('X-Timezone')) {
    headers.set('X-Timezone', timezone)
  }

  return headers
}

const toBase64Url = (bytes: ArrayBuffer) => {
  const value = String.fromCharCode(...new Uint8Array(bytes))
  return btoa(value).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

const toHex = (bytes: ArrayBuffer) =>
  Array.from(new Uint8Array(bytes), byte => byte.toString(16).padStart(2, '0')).join('')

const normalizeBaseURL = (value?: string) => {
  const normalized = String(value || '').trim().replace(/\/+$/, '')
  if (!normalized) return '/api/v1'
  if (/^https?:\/\//i.test(normalized) || normalized.startsWith('/')) return normalized
  return `/${normalized}`
}

const resolveApiBases = (config: ReturnType<typeof useRuntimeConfig>) => {
  const baseURL = normalizeBaseURL((config.public as { apiBase?: string })?.apiBase)
  const internalOrigin = String((config as { apiInternalOrigin?: string }).apiInternalOrigin || '').trim().replace(/\/+$/, '')
  const requestBaseURL = import.meta.server && internalOrigin && baseURL.startsWith('/')
    ? `${internalOrigin}${baseURL === '/' ? '' : baseURL}`
    : baseURL

  return {
    baseURL,
    requestBaseURL,
  }
}

const apiRequestURL = (baseURL: string, path: string) => {
  const normalizedBase = baseURL.replace(/\/+$/, '')
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return normalizedBase ? `${normalizedBase}${normalizedPath}` : normalizedPath
}

const requestTarget = (baseURL: string, path: string) => {
  const origin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost'
  const url = new URL(apiRequestURL(baseURL || '/', path), origin)
  return `${url.pathname}${url.search}`
}

const appendQueryValue = (params: URLSearchParams, key: string, value: ApiRequestQueryValue) => {
  if (value === undefined || value === null || value === '') {
    return
  }
  if (Array.isArray(value)) {
    value.forEach(item => appendQueryValue(params, key, item))
    return
  }
  if (value instanceof Date) {
    params.append(key, value.toISOString())
    return
  }
  if (typeof value === 'object') {
    params.append(key, JSON.stringify(value))
    return
  }
  params.append(key, String(value))
}

const queryString = (query?: ApiRequestQuery) => {
  if (!query) {
    return ''
  }
  if (query instanceof URLSearchParams) {
    return query.toString()
  }

  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => appendQueryValue(params, key, value))
  return params.toString()
}

const appendQuery = (path: string, ...queries: Array<ApiRequestQuery | undefined>) => {
  const additions = queries.map(queryString).filter(Boolean)
  if (additions.length === 0) {
    return path
  }

  const [pathWithoutHash = '', hash = ''] = path.split('#', 2)
  const separator = pathWithoutHash.includes('?') ? '&' : '?'
  const nextPath = `${pathWithoutHash}${separator}${additions.join('&')}`
  return hash ? `${nextPath}#${hash}` : nextPath
}

const canSignRequestBody = (body: BodyInit | null | undefined) => {
  return body === undefined || body === null || typeof body === 'string' || body instanceof URLSearchParams
}

const signRequest = async (headers: Headers, method: string, baseURL: string, path: string, body: BodyInit | null | undefined, key: string) => {
  if (!import.meta.client || !key || !globalThis.crypto?.subtle) {
    return
  }
  if (!canSignRequestBody(body)) {
    return
  }

  const timestamp = Math.floor(Date.now() / 1000).toString()
  const nonce = globalThis.crypto.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`
  const bodyText = typeof body === 'string' ? body : body instanceof URLSearchParams ? body.toString() : ''
  const encoder = new TextEncoder()
  const bodyDigest = await globalThis.crypto.subtle.digest('SHA-256', encoder.encode(bodyText))
  const canonical = [
    method.toUpperCase(),
    requestTarget(baseURL, path),
    timestamp,
    nonce,
    toHex(bodyDigest),
  ].join('\n')
  const cryptoKey = await globalThis.crypto.subtle.importKey(
    'raw',
    encoder.encode(key),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const signature = await globalThis.crypto.subtle.sign('HMAC', cryptoKey, encoder.encode(canonical))
  headers.set('X-Request-Timestamp', timestamp)
  headers.set('X-Request-Nonce', nonce)
  headers.set('X-Request-Signature', toBase64Url(signature))
}

const readResponse = async (response: Response): Promise<MaybeJson> => {
  const text = await response.text()
  if (!text) {
    return null
  }
  try {
    return JSON.parse(text)
  } catch (_) {
    return text
  }
}

const extractMessage = (payload: MaybeJson, fallback: string) => {
  if (!payload) {
    return fallback
  }
  if (typeof payload === 'string') {
    return payload || fallback
  }
  const message = payload?.message
  return typeof message === 'string' && message.trim().length > 0 ? message : fallback
}

const extractCode = (payload: MaybeJson) => {
  if (!payload || typeof payload === 'string') {
    return undefined
  }
  const code = payload.code
  return typeof code === 'string' && code.trim().length > 0 ? code : undefined
}

const extractDetails = (payload: MaybeJson) => {
  if (!payload || typeof payload === 'string') {
    return undefined
  }
  return payload.details
}

export function useApiRequest() {
  const config = useRuntimeConfig()
  const { baseURL, requestBaseURL } = resolveApiBases(config)
  const requestSigningKey = config.public?.requestSigningKey || ''

  const request: ApiRequestFunction = async <T = MaybeJson>(path: string, init: ApiRequestInit = {}, fallbackMessage = 'Request failed'): Promise<T> => {
    if (!requestBaseURL) {
      throw new Error('Missing runtimeConfig.public.apiBase for API requests')
    }

    const { params, query, ...fetchInit } = init
    const requestPath = appendQuery(path, params, query)
    const headers = new Headers(fetchInit.headers || undefined)
    if (isUnsafeMethod(fetchInit.method)) {
      const csrfToken = readCookie(csrfCookieName)
      if (csrfToken) {
        headers.set(csrfHeaderName, csrfToken)
      }
    }
    const requestHeaders = await attachDeviceFingerprintHeader(attachClientContextHeaders(headers))
    await signRequest(requestHeaders, (fetchInit.method || 'GET').toUpperCase(), requestBaseURL, requestPath, fetchInit.body, requestSigningKey)
    const requestURL = apiRequestURL(requestBaseURL, requestPath)

    const finalInit: RequestInit = {
      credentials: defaultCredentials,
      ...fetchInit,
      headers: requestHeaders,
    }

    const response = await fetch(requestURL, finalInit)
    const payload = await readResponse(response)

    if (!response.ok) {
      throw new ApiRequestError(extractMessage(payload, fallbackMessage), {
        status: response.status,
        code: extractCode(payload),
        details: extractDetails(payload),
        payload,
      })
    }

    return payload as T
  }

  return {
    baseURL,
    requestBaseURL,
    request,
  }
}

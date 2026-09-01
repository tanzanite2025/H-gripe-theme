import { useRuntimeConfig } from 'nuxt/app'
import { attachDeviceFingerprintHeader } from '~/utils/deviceFingerprint'

export type MaybeJson = Record<string, unknown> | string | null

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
  return cookie ? decodeURIComponent(cookie.slice(prefix.length)) : ''
}

const toBase64Url = (bytes: ArrayBuffer) => {
  const value = String.fromCharCode(...new Uint8Array(bytes))
  return btoa(value).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

const toHex = (bytes: ArrayBuffer) =>
  Array.from(new Uint8Array(bytes), byte => byte.toString(16).padStart(2, '0')).join('')

const apiRequestURL = (baseURL: string, path: string) => {
  const normalizedBase = baseURL.replace(/\/+$/, '')
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

const requestTarget = (baseURL: string, path: string) => {
  const origin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost'
  const url = new URL(apiRequestURL(baseURL || '/', path), origin)
  return `${url.pathname}${url.search}`
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
  const baseURL = config.public?.apiBase || '/api/v1'
  const requestSigningKey = config.public?.requestSigningKey || ''

  const request = async <T = MaybeJson>(path: string, init: RequestInit = {}, fallbackMessage = 'Request failed'): Promise<T> => {
    if (!baseURL) {
      throw new Error('Missing runtimeConfig.public.apiBase for API requests')
    }

    const headers = new Headers(init.headers || undefined)
    if (isUnsafeMethod(init.method)) {
      const csrfToken = readCookie(csrfCookieName)
      if (csrfToken) {
        headers.set(csrfHeaderName, csrfToken)
      }
    }
    const requestHeaders = await attachDeviceFingerprintHeader(headers)
    await signRequest(requestHeaders, (init.method || 'GET').toUpperCase(), baseURL, path, init.body, requestSigningKey)
    const requestURL = apiRequestURL(baseURL, path)

    const finalInit: RequestInit = {
      credentials: defaultCredentials,
      ...init,
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
    request,
  }
}

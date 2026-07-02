import { useRuntimeConfig } from 'nuxt/app'

export type MaybeJson = Record<string, unknown> | string | null

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

export function useApiRequest() {
  const config = useRuntimeConfig()
  const baseURL = config.public?.apiBase || '/api/v1'

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

    const finalInit: RequestInit = {
      credentials: defaultCredentials,
      ...init,
      headers,
    }

    const response = await fetch(`${baseURL}${path}`, finalInit)
    const payload = await readResponse(response)

    if (!response.ok) {
      throw new Error(extractMessage(payload, fallbackMessage))
    }

    return payload as T
  }

  return {
    baseURL,
    request,
  }
}


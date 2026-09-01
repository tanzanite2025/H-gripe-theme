import axios, { AxiosHeaders } from 'axios'
import type { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { toast } from 'vue-sonner'
import { deviceFingerprintHeaderName, resolveDeviceFingerprint } from './deviceFingerprint'

declare module 'axios' {
  export interface AxiosRequestConfig {
    suppressGlobalErrorToast?: boolean
  }
}

interface ApiErrorPayload {
  error?: string
  message?: string
}

interface RetryableRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
}

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  withCredentials: true
})

const LOGIN_PATH = '/login'
const AUTH_STORAGE_KEYS = ['admin_user', 'admin_permissions']
const CSRF_COOKIE_NAME = 'csrf_token'
const CSRF_HEADER_NAME = 'X-CSRF-Token'

let refreshPromise: Promise<boolean> | null = null
let sessionExpiredHandled = false
let pendingRequests = 0

const isLoginEndpoint = (url = ''): boolean => (
  url.includes('/api/admin/auth/login') || url.includes('/api/admin/auth/google-login')
)
const isRefreshEndpoint = (url = ''): boolean => url.includes('/api/admin/auth/refresh')
const isUnsafeMethod = (method = 'get'): boolean => !['get', 'head', 'options', 'trace'].includes(method.toLowerCase())

const readCookie = (name: string): string => {
  if (typeof document === 'undefined') return ''
  const prefix = `${encodeURIComponent(name)}=`
  const cookie = document.cookie
    .split(';')
    .map((item) => item.trim())
    .find((item) => item.startsWith(prefix))
  return cookie ? decodeURIComponent(cookie.slice(prefix.length)) : ''
}

const attachCsrfHeader = (headers?: InternalAxiosRequestConfig['headers']): AxiosHeaders => {
  const nextHeaders = AxiosHeaders.from(headers || {})
  const token = readCookie(CSRF_COOKIE_NAME)
  if (!token) return nextHeaders

  nextHeaders.set(CSRF_HEADER_NAME, token)
  return nextHeaders
}

const attachDeviceFingerprintHeader = async (
  headers?: InternalAxiosRequestConfig['headers'],
): Promise<AxiosHeaders> => {
  const nextHeaders = AxiosHeaders.from(headers || {})
  const fingerprint = await resolveDeviceFingerprint()
  if (fingerprint && !nextHeaders.has(deviceFingerprintHeaderName)) {
    nextHeaders.set(deviceFingerprintHeaderName, fingerprint)
  }
  return nextHeaders
}

const clearAdminAuth = (): void => {
  AUTH_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key))
}

const redirectToLoginOnce = (): void => {
  if (sessionExpiredHandled) return
  sessionExpiredHandled = true

  clearAdminAuth()

  if (window.location.pathname === LOGIN_PATH) return

  toast.error('登录已过期，请重新登录', { id: 'admin-session-expired' })
  window.location.assign(LOGIN_PATH)
}

const refreshAdminToken = async (): Promise<boolean> => {
  if (!refreshPromise) {
    refreshPromise = axios.post('/api/admin/auth/refresh', null, {
      baseURL: instance.defaults.baseURL,
      timeout: instance.defaults.timeout,
      withCredentials: true,
      headers: await attachDeviceFingerprintHeader(attachCsrfHeader())
    }).then(() => {
      sessionExpiredHandled = false
      return true
    }).finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

const silenceAuthFailure = (): Promise<never> => new Promise(() => {})

const emitLoading = (): void => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent('admin-api-loading', { detail: { loading: pendingRequests > 0 } }))
}

const beginRequest = (): void => {
  pendingRequests += 1
  emitLoading()
}

const endRequest = (): void => {
  pendingRequests = Math.max(0, pendingRequests - 1)
  emitLoading()
}

instance.interceptors.request.use(
  async (config) => {
    beginRequest()
    try {
      let headers = AxiosHeaders.from(config.headers || {})
      if (isUnsafeMethod(config.method)) {
        headers = attachCsrfHeader(headers)
      }
      config.headers = await attachDeviceFingerprintHeader(headers)
      return config
    } catch (error) {
      endRequest()
      return Promise.reject(error)
    }
  },
  (error) => {
    return Promise.reject(error)
  }
)

instance.interceptors.response.use(
  (response) => {
    endRequest()
    if (isLoginEndpoint(response.config?.url)) {
      sessionExpiredHandled = false
    }
    return response
  },
  async (error: AxiosError<ApiErrorPayload>) => {
    endRequest()
    const requestConfig = error.config as RetryableRequestConfig | undefined
    const suppressGlobalErrorToast = Boolean(requestConfig?.suppressGlobalErrorToast)

    if (error.response) {
      const { status, data } = error.response

      if (isLoginEndpoint(requestConfig?.url)) {
        return Promise.reject(error)
      }

      switch (status) {
        case 401:
          if (requestConfig && !isRefreshEndpoint(requestConfig.url) && !requestConfig._retry) {
            try {
              await refreshAdminToken()
              const retryConfig: RetryableRequestConfig = {
                ...requestConfig,
                _retry: true,
                headers: AxiosHeaders.from(requestConfig.headers || {})
              }
              return instance(retryConfig)
            } catch {
              redirectToLoginOnce()
              return silenceAuthFailure()
            }
          }

          redirectToLoginOnce()
          return silenceAuthFailure()
        case 403:
          if (!suppressGlobalErrorToast) {
            toast.error(data.message || '没有权限访问', { id: 'api-forbidden' })
          }
          break
        case 404:
          if (!suppressGlobalErrorToast) {
            toast.error('请求的资源不存在', { id: 'api-not-found' })
          }
          break
        case 500:
          if (!suppressGlobalErrorToast) {
            toast.error('服务器错误', { id: 'api-server-error' })
          }
          break
        default:
          if (!suppressGlobalErrorToast) {
            toast.error(data.error || data.message || '请求失败', { id: 'api-request-error' })
          }
      }
    } else if (error.request) {
      if (!suppressGlobalErrorToast) {
        toast.error('网络错误，请检查网络连接', { id: 'api-network-error' })
      }
    } else {
      if (!suppressGlobalErrorToast) {
        toast.error('请求配置错误', { id: 'api-config-error' })
      }
    }

    return Promise.reject(error)
  }
)

export default instance

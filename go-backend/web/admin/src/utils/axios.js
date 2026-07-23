import axios from 'axios'
import { toast } from 'vue-sonner'

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  withCredentials: true
})

const LOGIN_PATH = '/login'
const AUTH_STORAGE_KEYS = ['admin_user', 'admin_permissions']
const CSRF_COOKIE_NAME = 'csrf_token'
const CSRF_HEADER_NAME = 'X-CSRF-Token'

let refreshPromise = null
let sessionExpiredHandled = false
let pendingRequests = 0

const isLoginEndpoint = (url = '') => url.includes('/api/admin/auth/login')
const isRefreshEndpoint = (url = '') => url.includes('/api/admin/auth/refresh')
const isUnsafeMethod = (method = 'get') => !['get', 'head', 'options', 'trace'].includes(method.toLowerCase())

const readCookie = (name) => {
  if (typeof document === 'undefined') return ''
  const prefix = `${encodeURIComponent(name)}=`
  const cookie = document.cookie
    .split(';')
    .map((item) => item.trim())
    .find((item) => item.startsWith(prefix))
  return cookie ? decodeURIComponent(cookie.slice(prefix.length)) : ''
}

const attachCsrfHeader = (headers = {}) => {
  const token = readCookie(CSRF_COOKIE_NAME)
  if (token) {
    headers[CSRF_HEADER_NAME] = token
  }
  return headers
}

const clearAdminAuth = () => {
  AUTH_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key))
}

const redirectToLoginOnce = () => {
  if (sessionExpiredHandled) return
  sessionExpiredHandled = true

  toast.error('登录已过期，请重新登录', { id: 'admin-session-expired' })
  clearAdminAuth()

  if (window.location.pathname !== LOGIN_PATH) {
    window.location.assign(LOGIN_PATH)
  }
}

const refreshAdminToken = async () => {
  if (!refreshPromise) {
    refreshPromise = axios.post('/api/admin/auth/refresh', null, {
      baseURL: instance.defaults.baseURL,
      timeout: instance.defaults.timeout,
      withCredentials: true,
      headers: attachCsrfHeader({})
    }).then(() => {
      sessionExpiredHandled = false
      return true
    }).finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

const silenceAuthFailure = () => new Promise(() => {})

const emitLoading = () => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent('admin-api-loading', { detail: { loading: pendingRequests > 0 } }))
}

const beginRequest = () => {
  pendingRequests += 1
  emitLoading()
}

const endRequest = () => {
  pendingRequests = Math.max(0, pendingRequests - 1)
  emitLoading()
}

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    beginRequest()
    if (isUnsafeMethod(config.method)) {
      attachCsrfHeader(config.headers)
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
instance.interceptors.response.use(
  (response) => {
    endRequest()
    if (isLoginEndpoint(response.config?.url)) {
      sessionExpiredHandled = false
    }
    return response
  },
  async (error) => {
    endRequest()
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          if (isLoginEndpoint(error.config?.url)) {
            return Promise.reject(error)
          }

          if (!isRefreshEndpoint(error.config?.url) && !error.config?._retry) {
            try {
              await refreshAdminToken()
              const retryConfig = {
                ...error.config,
                _retry: true,
                headers: { ...error.config.headers }
              }
              return instance(retryConfig)
            } catch (refreshError) {
              redirectToLoginOnce()
              return silenceAuthFailure()
            }
          }

          redirectToLoginOnce()
          return silenceAuthFailure()
        case 403:
          toast.error(data.message || '没有权限访问', { id: 'api-forbidden' })
          break
        case 404:
          toast.error('请求的资源不存在', { id: 'api-not-found' })
          break
        case 500:
          toast.error('服务器错误', { id: 'api-server-error' })
          break
        default:
          toast.error(data.error || data.message || '请求失败', { id: 'api-request-error' })
      }
    } else if (error.request) {
      toast.error('网络错误，请检查网络连接', { id: 'api-network-error' })
    } else {
      toast.error('请求配置错误', { id: 'api-config-error' })
    }

    return Promise.reject(error)
  }
)

export default instance

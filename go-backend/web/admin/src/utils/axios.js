import axios from 'axios'
import { ElMessage } from 'element-plus'

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  withCredentials: true
})

const LOGIN_PATH = '/login'
const AUTH_STORAGE_KEYS = ['admin_token', 'admin_user', 'admin_permissions']
const CSRF_COOKIE_NAME = 'csrf_token'
const CSRF_HEADER_NAME = 'X-CSRF-Token'

let refreshPromise = null
let sessionExpiredHandled = false

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
  delete instance.defaults.headers.common.Authorization
  delete axios.defaults.headers.common.Authorization
}

const storeToken = (token) => {
  if (!token) return
  sessionExpiredHandled = false
  localStorage.setItem('admin_token', token)
  instance.defaults.headers.common.Authorization = `Bearer ${token}`
  axios.defaults.headers.common.Authorization = `Bearer ${token}`
}

const redirectToLoginOnce = () => {
  if (sessionExpiredHandled) return
  sessionExpiredHandled = true

  ElMessage.error('登录已过期，请重新登录')
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
      headers: attachCsrfHeader({
        Authorization: localStorage.getItem('admin_token')
          ? `Bearer ${localStorage.getItem('admin_token')}`
          : undefined
      })
    }).then((response) => {
      const newToken = response.data?.token
      if (newToken) {
        storeToken(newToken)
      }
      return newToken || localStorage.getItem('admin_token')
    }).finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

const silenceAuthFailure = () => new Promise(() => {})

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
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
    if (isLoginEndpoint(response.config?.url)) {
      sessionExpiredHandled = false
    }
    return response
  },
  async (error) => {
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          if (isLoginEndpoint(error.config?.url)) {
            return Promise.reject(error)
          }

          if (!isRefreshEndpoint(error.config?.url) && !error.config?._retry) {
            try {
              const token = await refreshAdminToken()
              const retryConfig = {
                ...error.config,
                _retry: true,
                headers: {
                  ...error.config.headers,
                  Authorization: token ? `Bearer ${token}` : error.config.headers?.Authorization
                }
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
          ElMessage.error(data.message || '没有权限访问')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器错误')
          break
        default:
          ElMessage.error(data.error || data.message || '请求失败')
      }
    } else if (error.request) {
      ElMessage.error('网络错误，请检查网络连接')
    } else {
      ElMessage.error('请求配置错误')
    }

    return Promise.reject(error)
  }
)

export default instance

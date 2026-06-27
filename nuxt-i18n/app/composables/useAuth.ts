import { useRuntimeConfig, useState } from 'nuxt/app'
import { computed } from 'vue'

interface LoginPayload {
  username: string
  password: string
  remember?: boolean
}

interface RegisterProfile {
  fullName?: string
  phone?: string
  country?: string
  company?: string
  marketingOptIn?: boolean
  notes?: string
}

interface RegisterPayload {
  username: string
  email: string
  password: string
  profile?: RegisterProfile
}

interface AuthUser {
  id?: number
  username?: string
  email?: string
  display_name?: string
  avatar?: string
  roles?: string[]
  is_agent?: boolean
  agent_id?: string | null
  profile?: RegisterProfile
  loyalty?: Record<string, unknown>
  [key: string]: unknown
}

// API 响应格式
interface ApiResponse<T> {
  ok: boolean
  data: T
  message?: string
}

type MaybeJson = Record<string, unknown> | string | null

const defaultCredentials: RequestCredentials = 'include'

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

const unwrapData = <T>(payload: T | { data?: T } | null | undefined): T | null => {
  if (!payload || typeof payload !== 'object') {
    return (payload as T) || null
  }
  if ('data' in payload && payload.data !== undefined) {
    return payload.data as T
  }
  return payload as T
}

export function useAuth() {
  const config = useRuntimeConfig()
  // 抛弃 wpApiBase，直连 Go 后端
  const baseURL = config.public?.apiBase || '/api/v1'

  const user = useState<AuthUser | null>('auth-user', () => null)
  const loading = useState<boolean>('auth-loading', () => false)
  const error = useState<string | null>('auth-error', () => null)
  const initialized = useState<boolean>('auth-initialized', () => false)

  const request = async <T = MaybeJson>(path: string, init: RequestInit = {}, fallbackMessage = 'Request failed'): Promise<T> => {
    if (!baseURL) {
      throw new Error('Missing runtimeConfig.public.wpApiBase for authentication requests')
    }

    const headers = new Headers(init.headers || undefined)

    const finalInit: RequestInit = {
      credentials: defaultCredentials,
      ...init,
      headers
    }

    const response = await fetch(`${baseURL}${path}`, finalInit)
    const payload = await readResponse(response)

    if (!response.ok) {
      throw new Error(extractMessage(payload, fallbackMessage))
    }

    return payload as T
  }

  const ensureSession = async (force = false) => {
    if (!baseURL) {
      initialized.value = true
      return null
    }
    if (initialized.value && !force) {
      return user.value
    }

    initialized.value = true
    try {
      const response = await request<AuthUser | { data?: AuthUser }>('/auth/profile', { headers: { 'Accept': 'application/json' } }, 'Unable to fetch session')
      const data = unwrapData<AuthUser>(response)
      user.value = data
      error.value = null
      return data
    } catch (_) {
      user.value = null
      return null
    }
  }

  const login = async (payload: LoginPayload) => {
    loading.value = true
    error.value = null

    try {
      const response = await request<{ token?: string, user?: AuthUser } | { data?: { token?: string, user?: AuthUser } }>(
        '/auth/login',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify(payload)
        },
        'Login failed'
      )
      
      const payload = unwrapData<{ token?: string, user?: AuthUser }>(response)
      const data = payload?.user || null
      user.value = data
      return data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed'
      error.value = message
      throw new Error(message)
    } finally {
      loading.value = false
    }
  }

  const register = async (payload: RegisterPayload) => {
    loading.value = true
    error.value = null

    try {
      const response = await request<{ message?: string, user?: AuthUser } | { data?: { message?: string, user?: AuthUser } }>(
        '/auth/register',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify(payload)
        },
        'Registration failed'
      )
      // 注册接口目前未返回token，需要在注册后让用户去登录，或后端改进返回token
      const payload = unwrapData<{ message?: string, user?: AuthUser }>(response)
      const data = payload?.user || null
      user.value = data
      return data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Registration failed'
      error.value = message
      throw new Error(message)
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    if (!baseURL) {
      user.value = null
      return
    }

    try {
      await request('/auth/logout', { method: 'POST' }, 'Logout failed')
    } catch (err) {
      console.warn('Logout request failed:', err)
    } finally {
      user.value = null
    }
  }

  /**
   * 使用 Google ID Token 登录
   * @param idToken - Google Identity Services 返回的 JWT token
   */
  const loginWithGoogle = async (idToken: string) => {
    loading.value = true
    error.value = null

    try {
      const response = await request<{ token?: string, user?: AuthUser } | { data?: { token?: string, user?: AuthUser } }>(
        '/auth/google-login',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify({ id_token: idToken })
        },
        'Google login failed'
      )
      const payload = unwrapData<{ token?: string, user?: AuthUser }>(response)
      const data = payload?.user || null
      user.value = data
      return data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Google login failed'
      error.value = message
      throw new Error(message)
    } finally {
      loading.value = false
    }
  }

  // 计算属性：是否是客服
  const isAgent = computed(() => !!user.value?.is_agent)

  // 计算属性：客服 ID
  const agentId = computed(() => user.value?.agent_id || null)

  return {
    user,
    loading,
    error,
    initialized,
    isAgent,
    agentId,
    ensureSession,
    login,
    loginWithGoogle,
    register,
    logout,
    request // 暴露 request 方法以便其他 composable (如 useMembership) 调用
  }
}

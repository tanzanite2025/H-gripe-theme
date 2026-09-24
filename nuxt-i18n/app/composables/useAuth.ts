import { useNuxtApp, useState } from 'nuxt/app'
import { computed } from 'vue'
import { ApiRequestError, type ApiRequestInit } from '~/composables/useApiRequest'
import { hasBrowserCookie } from '~/utils/browserCookies'

interface LoginPayload {
  username: string
  password: string
  remember?: boolean
  corporate_website?: string
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
  corporate_website?: string
  profile?: RegisterProfile
  referralCode?: string
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

const sessionRequests = new WeakMap<object, Promise<AuthUser | null>>()

// API 响应格式
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
  const nuxtApp = useNuxtApp()
  const { baseURL, request } = useApiRequest()

  const user = useState<AuthUser | null>('auth-user', () => null)
  const loading = useState<boolean>('auth-loading', () => false)
  const error = useState<string | null>('auth-error', () => null)
  const referralBindingError = useState<string | null>('auth-referral-binding-error', () => null)
  // Keep a manually entered code across the registration -> login hand-off.
  // The register endpoint intentionally does not create a session, so the
  // code is bound after the user completes their first login.
  const pendingReferralCode = useState<string>('auth-pending-referral-code', () => '')
  const initialized = useState<boolean>('auth-initialized', () => false)
  const sessionVersion = useState<number>('auth-session-version', () => 0)
  const logoutInFlight = useState<boolean>('auth-logout-in-flight', () => false)
  const isAuthenticated = computed(() => !!user.value)

  const bindPendingReferral = async () => {
    referralBindingError.value = null
    const manualCode = pendingReferralCode.value.trim()
    try {
      const init: ApiRequestInit = {
        method: 'POST',
        headers: { 'Accept': 'application/json' }
      }
      if (manualCode) {
        init.headers = { 'Accept': 'application/json', 'Content-Type': 'application/json' }
        init.body = JSON.stringify({ referral_code: manualCode })
      }
      await request('/customer/referral/bind', init, 'Unable to apply referral attribution')
      pendingReferralCode.value = ''
    } catch (err) {
      if (err instanceof ApiRequestError && err.code === 'referral_attribution_missing') {
        return
      }
      pendingReferralCode.value = ''
      referralBindingError.value = err instanceof Error
        ? err.message
        : 'Unable to apply referral attribution'
    }
  }

  const ensureSession = async (force = false) => {
    if (logoutInFlight.value) {
      user.value = null
      initialized.value = true
      return null
    }
    if (!baseURL) {
      initialized.value = true
      return null
    }

    const pendingRequest = sessionRequests.get(nuxtApp)
    if (pendingRequest) {
      return pendingRequest
    }

    if (initialized.value && !force) {
      return user.value
    }

    if (import.meta.client && !force && !hasBrowserCookie('storefront_csrf_token')) {
      user.value = null
      initialized.value = true
      return null
    }

    const requestVersion = sessionVersion.value
    const sessionRequest = (async () => {
      try {
        const response = await request<AuthUser | { data?: AuthUser }>('/auth/profile', { headers: { 'Accept': 'application/json' } }, 'Unable to fetch session')
        const data = unwrapData<AuthUser>(response)
        if (sessionVersion.value !== requestVersion) {
          return null
        }
        user.value = data
        error.value = null
        return data
      } catch (_) {
        if (sessionVersion.value === requestVersion) {
          user.value = null
        }
        return null
      } finally {
        if (sessionVersion.value === requestVersion) {
          initialized.value = true
        }
      }
    })()

    sessionRequests.set(nuxtApp, sessionRequest)

    try {
      return await sessionRequest
    } finally {
      if (sessionRequests.get(nuxtApp) === sessionRequest) {
        sessionRequests.delete(nuxtApp)
      }
    }
  }

  const login = async (credentials: LoginPayload) => {
    sessionVersion.value += 1
    loading.value = true
    error.value = null

    try {
      const response = await request<{ token?: string, user?: AuthUser } | { data?: { token?: string, user?: AuthUser } }>(
        '/auth/login',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify(credentials)
        },
        'Login failed'
      )
      
      const responsePayload = unwrapData<{ token?: string, user?: AuthUser }>(response)
      const data = responsePayload?.user || null
      user.value = data
      await bindPendingReferral()
      return data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed'
      error.value = message
      throw new Error(message)
    } finally {
      loading.value = false
    }
  }

  const register = async (registration: RegisterPayload) => {
    sessionVersion.value += 1
    loading.value = true
    error.value = null
    pendingReferralCode.value = String(registration.referralCode || '').trim()

    try {
      const response = await request<{ message?: string, user?: AuthUser } | { data?: { message?: string, user?: AuthUser } }>(
        '/auth/register',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify(registration)
        },
        'Registration failed'
      )
      // 注册接口目前未返回token，需要在注册后让用户去登录，或后端改进返回token
      const responsePayload = unwrapData<{ message?: string, user?: AuthUser }>(response)
      const data = responsePayload?.user || null
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
    // Invalidate in-flight session checks and clear local identity immediately.
    const logoutVersion = sessionVersion.value + 1
    sessionVersion.value = logoutVersion
    logoutInFlight.value = true
    sessionRequests.delete(nuxtApp)
    user.value = null
    initialized.value = false

    if (!baseURL) {
      logoutInFlight.value = false
      return
    }

    try {
      await request('/auth/logout', { method: 'POST' }, 'Logout failed')
    } catch (err) {
      console.warn('Logout request failed:', err)
    } finally {
      logoutInFlight.value = false
      sessionRequests.delete(nuxtApp)
      if (sessionVersion.value === logoutVersion) {
        user.value = null
        initialized.value = false
      }
      sessionVersion.value += 1
    }
  }

  /**
   * 使用 Google ID Token 登录
   * @param idToken - Google Identity Services 返回的 JWT token
   */
  const loginWithGoogle = async (idToken: string) => {
    sessionVersion.value += 1
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
      await bindPendingReferral()
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
    baseURL,
    user,
    loading,
    error,
    referralBindingError,
    initialized,
    isAuthenticated,
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

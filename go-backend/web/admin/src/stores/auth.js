import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import axios from '@/utils/axios'

const readJSON = (key, fallback) => {
  try {
    return JSON.parse(localStorage.getItem(key) || JSON.stringify(fallback))
  } catch {
    return fallback
  }
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref(readJSON('admin_user', null))
  const permissions = ref(readJSON('admin_permissions', []))
  const initialized = ref(false)

  const isAuthenticated = computed(() => !!user.value)

  const persistSession = (newUser) => {
    if (!newUser || !Array.isArray(newUser.permissions)) {
      throw new Error('[CRITICAL] Missing permissions array in auth response')
    }

    user.value = newUser
    permissions.value = newUser.permissions
    localStorage.setItem('admin_user', JSON.stringify(newUser))
    localStorage.setItem('admin_permissions', JSON.stringify(newUser.permissions))
  }

  const clearSession = () => {
    user.value = null
    permissions.value = []
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_permissions')
  }

  const hasPermission = (permission) => {
    if (!permission) return true
    return permissions.value.includes(permission)
  }

  const hasRole = (role) => {
    return user.value?.role === role
  }

  const login = async (email, password) => {
    try {
      const response = await axios.post('/api/admin/auth/login', {
        email,
        password,
      })

      persistSession(response.data?.user)
      initialized.value = true
      return true
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  const logout = async () => {
    try {
      await axios.post('/api/admin/auth/logout')
    } catch (error) {
      console.warn('Logout request failed:', error)
    } finally {
      clearSession()
      initialized.value = true
    }
  }

  const refreshToken = async () => {
    try {
      await axios.post('/api/admin/auth/refresh')
      return true
    } catch (error) {
      console.error('Token refresh failed:', error)
      clearSession()
      return false
    }
  }

  const verifyPermissions = async () => {
    try {
      const response = await axios.get('/api/admin/auth/profile')
      const serverUser = response.data?.user

      if (!serverUser || !Array.isArray(serverUser.permissions)) {
        throw new Error('[CRITICAL] Missing permissions array in profile response')
      }

      const localPerms = JSON.stringify([...permissions.value].sort())
      const serverPerms = JSON.stringify([...serverUser.permissions].sort())
      persistSession(serverUser)

      return {
        valid: true,
        updated: localPerms !== serverPerms,
      }
    } catch (error) {
      console.error('[Auth] Permission verification failed', error)

      if (error.response?.status === 401) {
        clearSession()
        return { valid: false, reason: 'Session expired' }
      }

      return { valid: !!user.value, updated: false, warning: 'Network error' }
    }
  }

  const initAuth = async () => {
    if (initialized.value) {
      return { authenticated: isAuthenticated.value }
    }

    const result = await verifyPermissions()
    initialized.value = true

    if (!result.valid) {
      return { authenticated: false, reason: result.reason }
    }

    return {
      authenticated: true,
      permissionsUpdated: result.updated,
      warning: result.warning,
    }
  }

  return {
    user,
    permissions,
    initialized,
    isAuthenticated,
    hasPermission,
    hasRole,
    login,
    logout,
    refreshToken,
    verifyPermissions,
    initAuth,
  }
})

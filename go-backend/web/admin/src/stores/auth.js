import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '@/utils/axios'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const user = ref(JSON.parse(localStorage.getItem('admin_user') || 'null'))
  const permissions = ref(JSON.parse(localStorage.getItem('admin_permissions') || '[]'))

  const isAuthenticated = computed(() => !!token.value)

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
        password
      })

      const { token: newToken, user: newUser } = response.data

      if (!newUser || !Array.isArray(newUser.permissions)) {
        throw new Error('[CRITICAL] Missing permissions array in login response')
      }

      token.value = newToken
      user.value = newUser
      permissions.value = newUser.permissions

      localStorage.setItem('admin_token', newToken)
      localStorage.setItem('admin_user', JSON.stringify(newUser))
      localStorage.setItem('admin_permissions', JSON.stringify(newUser.permissions))

      // 设置 axios 默认 header
      axios.defaults.headers.common['Authorization'] = `Bearer ${newToken}`

      return true
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  const logout = () => {
    token.value = ''
    user.value = null
    permissions.value = []

    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_permissions')

    delete axios.defaults.headers.common['Authorization']
  }

  const refreshToken = async () => {
    try {
      const response = await axios.post('/api/admin/auth/refresh')
      const { token: newToken } = response.data

      token.value = newToken
      localStorage.setItem('admin_token', newToken)
      axios.defaults.headers.common['Authorization'] = `Bearer ${newToken}`

      return true
    } catch (error) {
      console.error('Token refresh failed:', error)
      logout()
      return false
    }
  }

  // 验证并刷新权限（从服务器获取最新权限）
  const verifyPermissions = async () => {
    if (!token.value) {
      return { valid: false, reason: 'No token' }
    }

    try {
      const response = await axios.get('/api/admin/auth/permissions')
      const serverPermissions = response.data.permissions || []

      // 对比本地和服务器权限
      const localPerms = JSON.stringify([...permissions.value].sort())
      const serverPerms = JSON.stringify([...serverPermissions].sort())

      if (localPerms !== serverPerms) {
        console.warn('[Auth] Permissions updated from server')
        permissions.value = serverPermissions
        localStorage.setItem('admin_permissions', JSON.stringify(serverPermissions))

        return { valid: true, updated: true }
      }

      return { valid: true, updated: false }
    } catch (error) {
      console.error('[Auth] Permission verification failed', error)

      // 如果是401错误，Token已过期，需要登出
      if (error.response?.status === 401) {
        console.warn('[Auth] Token expired, logging out')
        logout()
        return { valid: false, reason: 'Token expired' }
      }

      // 其他错误（如网络问题），保留当前权限但标记为未验证
      return { valid: true, updated: false, warning: 'Network error' }
    }
  }

  // 初始化认证状态
  const initAuth = async () => {
    if (!token.value) {
      return { authenticated: false }
    }

    // 页面加载时验证权限
    const result = await verifyPermissions()

    if (!result.valid) {
      console.warn('[Auth] Authentication invalid:', result.reason)
      return { authenticated: false, reason: result.reason }
    }

    return {
      authenticated: true,
      permissionsUpdated: result.updated,
      warning: result.warning
    }
  }

  // 初始化时设置 axios header
  if (token.value) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  return {
    token,
    user,
    permissions,
    isAuthenticated,
    hasPermission,
    hasRole,
    login,
    logout,
    refreshToken,
    verifyPermissions,
    initAuth
  }
})

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
    refreshToken
  }
})

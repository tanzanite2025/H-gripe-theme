/**
 * 购物车数据加载器
 * 负责从后端加载税率配置和用户积分
 */
import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'
import type { TaxRate, UserPoints } from './types/cart-calculation-types'

const extractList = <T>(payload: unknown): T[] | null => {
  let current = payload

  for (let depth = 0; depth < 3; depth += 1) {
    if (Array.isArray(current)) return current as T[]
    if (!current || typeof current !== 'object') return null

    const record = current as Record<string, unknown>
    if (Array.isArray(record.items)) return record.items as T[]
    current = record.data
  }

  return null
}

const extractData = <T>(payload: unknown): T => {
  let current = payload

  for (let depth = 0; depth < 2; depth += 1) {
    if (!current || typeof current !== 'object' || Array.isArray(current)) break
    const record = current as Record<string, unknown>
    if (!('data' in record)) break
    current = record.data
  }

  return current as T
}

export const useCartDataLoader = () => {
  const { ensureSession, initialized, isAuthenticated, request } = useAuth()

  // 状态
  const taxRates = ref<TaxRate[]>([])
  const userPoints = ref<UserPoints | null>(null)

  /**
   * 加载税率配置
   */
  const loadTaxRates = async () => {
    try {
      const response = await request<unknown>(
        '/payment/tax-rates',
        { headers: { accept: 'application/json' } }
      )
      const items = extractList<TaxRate>(response)
      if (!items) throw new Error('[CRITICAL] tax rate list missing')
      taxRates.value = items.filter((t: TaxRate) => t.is_active)
    } catch (error) {
      console.error('Failed to load tax rates:', error)
    }
  }

  /**
   * 加载用户积分信息
   */
  const loadUserPoints = async () => {
    if (!initialized.value) {
      await ensureSession()
    }

    if (!isAuthenticated.value) {
      userPoints.value = null
      return
    }

    try {
      const response = await request<unknown>(
        '/marketing/loyalty/points',
        { headers: { accept: 'application/json' } }
      )
      userPoints.value = extractData<UserPoints>(response)
    } catch (error) {
      console.error('Failed to load user points:', error)
      userPoints.value = null
    }
  }

  /**
   * 初始化（加载所有配置）
   */
  const initialize = async () => {
    await Promise.all([
      loadTaxRates(),
      loadUserPoints(),
    ])
  }

  return {
    // 状态
    taxRates,
    userPoints,

    // 方法
    loadTaxRates,
    loadUserPoints,
    initialize,
  }
}

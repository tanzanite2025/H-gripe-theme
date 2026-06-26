/**
 * 购物车数据加载器
 * 负责从后端加载运费模板、税率配置、用户积分等数据
 */
import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'
import type { CartShippingTemplate, TaxRate, UserPoints } from './types/cart-calculation-types'

export const useCartDataLoader = () => {
  const { request } = useAuth()

  // 状态
  const shippingTemplates = ref<CartShippingTemplate[]>([])
  const taxRates = ref<TaxRate[]>([])
  const userPoints = ref<UserPoints | null>(null)

  /**
   * 加载运费模板
   */
  const loadShippingTemplates = async () => {
    try {
      const response = await request<{ items: CartShippingTemplate[] }>(
        '/shipping/templates',
        { headers: { accept: 'application/json' } }
      )
      if (!response.items) throw new Error("[CRITICAL] response.items missing")
      shippingTemplates.value = response.items
    } catch (error) {
      console.error('Failed to load shipping templates:', error)
    }
  }

  /**
   * 加载税率配置
   */
  const loadTaxRates = async () => {
    try {
      const response = await request<{ items: TaxRate[] }>(
        '/payment/tax-rates',
        { headers: { accept: 'application/json' } }
      )
      if (!response.items) throw new Error("[CRITICAL] response.items missing")
      taxRates.value = response.items.filter((t: TaxRate) => t.is_active)
    } catch (error) {
      console.error('Failed to load tax rates:', error)
    }
  }

  /**
   * 加载用户积分信息
   */
  const loadUserPoints = async () => {
    try {
      const response = await request<UserPoints>(
        '/marketing/loyalty/points',
        { headers: { accept: 'application/json' } }
      )
      userPoints.value = response
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
      loadShippingTemplates(),
      loadTaxRates(),
      loadUserPoints(),
    ])
  }

  return {
    // 状态
    shippingTemplates,
    taxRates,
    userPoints,

    // 方法
    loadShippingTemplates,
    loadTaxRates,
    loadUserPoints,
    initialize,
  }
}

/**
 * 购物车税费计算
 * 支持多税率累加和自动税率选择
 */
import { ref, type Ref } from 'vue'
import type { TaxRate } from './types/cart-calculation-types'

export const useCartTax = (
  taxRates: Ref<TaxRate[]>
) => {
  // 状态
  const selectedTaxRates = ref<number[]>([])
  const shippingAddress = ref<{ region?: string } | null>(null)

  /**
   * 计算税费
   */
  const calculateTax = (subtotal: number, shipping: number): number => {
    if (selectedTaxRates.value.length === 0) {
      // Tax depends on destination and backend configuration. Never invent a
      // default rate before a checkout quote has been returned.
      return 0
    }

    // 累加所有选中的税率
    const totalTaxRate = selectedTaxRates.value.reduce((sum, taxId) => {
      const tax = (taxRates.value || []).find(t => t.id === taxId)
      return sum + (tax?.rate || 0)
    }, 0)

    return (subtotal + shipping) * (totalTaxRate / 100)
  }

  /**
   * 根据地区自动选择税率
   */
  const autoSelectTaxRates = () => {
    if (!shippingAddress.value?.region) {
      return
    }

    const region = shippingAddress.value.region
    const matchedTaxes = (taxRates.value || []).filter(
      t => t.region && t.region.toLowerCase() === region.toLowerCase()
    )

    selectedTaxRates.value = matchedTaxes.map(t => t.id)
  }

  return {
    // 状态
    selectedTaxRates,
    shippingAddress,

    // 方法
    calculateTax,
    autoSelectTaxRates,
  }
}

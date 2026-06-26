/**
 * 购物车运费计算
 * 支持基本运费计算和基于地区的运费计算
 */
import { ref } from 'vue'
import { isZipInRanges } from '../useShippingValidation'
import type { 
  CartShippingTemplate, 
  CartItem, 
  ShippingAddressInfo, 
  ShippingCalculationResult 
} from './types/cart-calculation-types'

export const useCartShipping = (
  shippingTemplates: ReturnType<typeof ref<CartShippingTemplate[]>>
) => {
  // 状态
  const selectedShippingTemplate = ref<number | null>(null)

  /**
   * 计算运费
   */
  const calculateShipping = (
    items: CartItem[],
    subtotal: number
  ): number => {
    if (!selectedShippingTemplate.value) {
      // 默认运费规则：满 100 免运费
      return subtotal >= 100 ? 0 : 10
    }

    const template = shippingTemplates.value.find(
      t => t.id === selectedShippingTemplate.value
    )

    if (!template) {
      return 10
    }

    // 检查免运费阈值
    if (template.free_threshold && subtotal >= template.free_threshold) {
      return 0
    }

    // 根据模板类型计算
    let calculationValue = 0

    switch (template.type) {
      case 'weight':
        calculationValue = items.reduce((sum, item) => {
          return sum + (item.weight || 0) * item.quantity
        }, 0)
        break
      case 'quantity':
        calculationValue = items.reduce((sum, item) => sum + item.quantity, 0)
        break
      case 'amount':
        calculationValue = subtotal
        break
      case 'items':
        calculationValue = items.length
        break
      default:
        return template.base_fee
    }

    // 查找匹配的规则
    const matchedRule = template.rules?.find(
      rule => {
        const min = rule.min ?? 0
        const max = rule.max ?? Infinity
        return calculationValue >= min && calculationValue <= max
      }
    )

    return matchedRule ? matchedRule.fee : template.base_fee
  }

  /**
   * 基于地区计算运费（新版）
   * 根据国家、邮编匹配运费规则
   */
  const calculateShippingByRegion = (
    items: CartItem[],
    subtotal: number,
    address: ShippingAddressInfo
  ): ShippingCalculationResult => {
    const { country, zip } = address
    const normalizedCountry = country?.toUpperCase() || ''

    if (!normalizedCountry) {
      return { fee: 0, rule: null, template: null }
    }

    // 计算总重量
    const totalWeight = items.reduce((sum, item) => sum + (item.weight || 0) * item.quantity, 0)
    const totalQuantity = items.reduce((sum, item) => sum + item.quantity, 0)

    // 遍历所有模板查找匹配的规则
    for (const template of shippingTemplates.value) {
      if (template.is_active === false) continue

      if (!template.rules) throw new Error("[CRITICAL] template.rules missing")
      for (const rule of template.rules) {
        // 检查国家匹配
        if (!rule.regions) throw new Error("[CRITICAL] rule.regions missing")
        const regions = rule.regions
        if (regions.length === 0 || !regions.map(r => r.toUpperCase()).includes(normalizedCountry)) {
          continue
        }

        // 检查邮编匹配
        if (!rule.zip_ranges) throw new Error("[CRITICAL] rule.zip_ranges missing")
        const zipRanges = rule.zip_ranges
        if (zipRanges.length > 0 && zip) {
          const zipMatched = isZipInRanges(zip, zipRanges)
          if (!zipMatched) continue
        }

        // 检查重量/数量范围
        const ruleType = rule.type || 'weight'
        let valueToCheck = 0

        switch (ruleType) {
          case 'weight':
            valueToCheck = totalWeight
            break
          case 'quantity':
          case 'items':
            valueToCheck = totalQuantity
            break
          case 'amount':
            valueToCheck = subtotal
            break
          default:
            valueToCheck = totalWeight
        }

        const min = rule.min ?? 0
        const max = rule.max ?? Infinity

        if (valueToCheck >= min && valueToCheck <= max) {
          // 检查免运费门槛
          if (rule.free_over && subtotal >= rule.free_over) {
            return { fee: 0, rule, template }
          }
          return { fee: rule.fee, rule, template }
        }
      }
    }

    // 没有匹配的规则，返回默认运费
    return { fee: subtotal >= 100 ? 0 : 10, rule: null, template: null }
  }

  return {
    // 状态
    selectedShippingTemplate,

    // 方法
    calculateShipping,
    calculateShippingByRegion,
  }
}

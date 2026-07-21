import { ref } from 'vue'
import { isZipInRanges } from '../useShippingValidation'
import type {
  CartShippingTemplate,
  CartItem,
  ShippingAddressInfo,
  ShippingCalculationResult,
} from './types/cart-calculation-types'

export const useCartShipping = (
  shippingTemplates: ReturnType<typeof ref<CartShippingTemplate[]>>
) => {
  const selectedShippingTemplate = ref<number | null>(null)

  const templateBaseFee = (template: CartShippingTemplate): number => {
    return Number(template.default_fee ?? template.base_fee ?? 0)
  }

  const isTemplateActive = (template: CartShippingTemplate): boolean => {
    return template.enabled !== false && template.is_active !== false
  }

  const itemWeightKg = (item: CartItem): number => {
    if (typeof item.weight === 'number') return item.weight
    if (typeof item.weight_grams === 'number') return item.weight_grams / 1000
    return 0
  }

  const ruleMin = (rule: CartShippingTemplate['rules'][0]): number => {
    return Number(rule.min_value ?? rule.min ?? 0)
  }

  const ruleMax = (rule: CartShippingTemplate['rules'][0]): number => {
    const value = Number(rule.max_value ?? rule.max ?? 0)
    return value > 0 ? value : Infinity
  }

  const splitRegions = (region: string): string[] => {
    const trimmed = region.trim()
    if (!trimmed) return []

    try {
      const parsed = JSON.parse(trimmed)
      if (Array.isArray(parsed)) {
        return parsed.map(item => String(item).trim().toUpperCase()).filter(Boolean)
      }
    } catch {
      // Use delimiter parsing below.
    }

    return trimmed
      .split(/[,;|\s]+/)
      .map(item => item.trim().toUpperCase())
      .filter(Boolean)
  }

  const ruleRegions = (rule: CartShippingTemplate['rules'][0]): string[] => {
    if (Array.isArray(rule.regions)) {
      return rule.regions.map(region => region.toUpperCase())
    }
    return splitRegions(rule.region || '')
  }

  const calculateShipping = (
    items: CartItem[],
    subtotal: number
  ): number => {
    if (!selectedShippingTemplate.value) {
      return 0
    }

    const template = shippingTemplates.value.find(
      item => item.id === selectedShippingTemplate.value
    )

    if (!template || !isTemplateActive(template)) {
      return 0
    }

    if (template.free_shipping && template.free_threshold && subtotal >= template.free_threshold) {
      return 0
    }

    let calculationValue = 0

    switch (template.type) {
      case 'weight':
        calculationValue = items.reduce((sum, item) => {
          return sum + itemWeightKg(item) * item.quantity
        }, 0)
        break
      case 'quantity':
      case 'items':
        calculationValue = items.reduce((sum, item) => sum + item.quantity, 0)
        break
      case 'amount':
      case 'price':
        calculationValue = subtotal
        break
      default:
        return templateBaseFee(template)
    }

    const matchedRule = template.rules?.find(rule => {
      const min = ruleMin(rule)
      const max = ruleMax(rule)
      return calculationValue >= min && calculationValue <= max
    })

    return matchedRule ? Number(matchedRule.fee || 0) : templateBaseFee(template)
  }

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

    const totalWeight = items.reduce((sum, item) => sum + itemWeightKg(item) * item.quantity, 0)
    const totalQuantity = items.reduce((sum, item) => sum + item.quantity, 0)

    for (const template of shippingTemplates.value) {
      if (!isTemplateActive(template)) continue

      if (!template.rules?.length) {
        if (template.free_shipping && template.free_threshold && subtotal >= template.free_threshold) {
          return { fee: 0, rule: null, template }
        }
        return { fee: templateBaseFee(template), rule: null, template }
      }

      for (const rule of template.rules || []) {
        const regions = ruleRegions(rule)
        if (regions.length > 0 && !regions.includes(normalizedCountry) && !regions.some(region => ['*', 'ALL', 'GLOBAL', 'WORLDWIDE'].includes(region))) {
          continue
        }

        const zipRanges = rule.zip_ranges || []
        if (zipRanges.length > 0 && zip) {
          const zipMatched = isZipInRanges(zip, zipRanges)
          if (!zipMatched) continue
        }

        const ruleType = rule.type || template.type || 'weight'
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
          case 'price':
            valueToCheck = subtotal
            break
          default:
            valueToCheck = totalWeight
        }

        const min = ruleMin(rule)
        const max = ruleMax(rule)

        if (valueToCheck >= min && valueToCheck <= max) {
          if (
            (rule.free_over && subtotal >= rule.free_over) ||
            (template.free_shipping && template.free_threshold && subtotal >= template.free_threshold)
          ) {
            return { fee: 0, rule, template }
          }
          return { fee: Number(rule.fee || 0), rule, template }
        }
      }
    }

    return { fee: 0, rule: null, template: null }
  }

  return {
    selectedShippingTemplate,
    calculateShipping,
    calculateShippingByRegion,
  }
}

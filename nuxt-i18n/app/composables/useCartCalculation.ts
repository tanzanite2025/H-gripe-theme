import { ref, computed } from 'vue'
import { isZipInRanges } from './useShippingValidation'
import { useAuth } from '~/composables/useAuth'

/**
 * 购物车计算系统 - 集成 Tanzanite Setting 配置
 * 
 * 功能：
 * - 从后端获取运费模板
 * - 从后端获取税率配置
 * - 计算会员等级折扣
 * - 计算积分抵扣
 * - 支持礼品卡/优惠券
 */

// 会员等级配置（与后端保持一致）
export interface MemberTier {
  name: string
  min: number
  max: number | null
  discount: number // 折扣百分比
}

export const MEMBER_TIERS: Record<string, MemberTier> = {
  ordinary: { name: 'Ordinary', min: 0, max: 499, discount: 0 },
  bronze: { name: 'Bronze', min: 500, max: 1999, discount: 5 },
  silver: { name: 'Silver', min: 2000, max: 4999, discount: 10 },
  gold: { name: 'Gold', min: 5000, max: 9999, discount: 15 },
  platinum: { name: 'Platinum', min: 10000, max: null, discount: 20 },
}

// 运费模板（扩展版）
export interface CartShippingTemplate {
  id: number
  name: string
  template_name?: string
  type: 'weight' | 'quantity' | 'volume' | 'amount' | 'items'
  base_fee: number
  free_threshold?: number
  is_active?: boolean
  rules: Array<{
    type?: string
    min: number | null
    max: number | null
    fee: number
    free_over?: number | null
    regions?: string[]
    zip_ranges?: string[]
    eta_min_days?: number | null
    eta_max_days?: number | null
    service?: string
    service_label?: string
  }>
}

// 配送地址
export interface ShippingAddressInfo {
  country: string
  zip?: string
  city?: string
}

// 税率配置
export interface TaxRate {
  id: number
  name: string
  rate: number // 百分比
  region?: string
  is_active: boolean
}

// 用户积分信息
export interface UserPoints {
  total: number
  available: number
  tier: string
}

// 优惠券/礼品卡
export interface Coupon {
  code: string
  type: 'percentage' | 'fixed' | 'points'
  value: number
  min_amount?: number
}

export const useCartCalculation = () => {
  // 状态
  const shippingTemplates = ref<CartShippingTemplate[]>([])
  const taxRates = ref<TaxRate[]>([])
  const userPoints = ref<UserPoints | null>(null)
  const appliedCoupon = ref<Coupon | null>(null)
  const usePointsDiscount = ref(false)
  const pointsToUse = ref(0)
  const selectedShippingTemplate = ref<number | null>(null)
  const selectedTaxRates = ref<number[]>([])
  const shippingAddress = ref<{ region?: string } | null>(null)

  const { request } = useAuth()

  /**
   * 加载运费模板
   */
  const loadShippingTemplates = async () => {
    try {
      const response = await request<{ items: CartShippingTemplate[] }>(
        '/shipping/templates',
        { headers: { accept: 'application/json' } }
      )
      if (!response.items) throw new Error("[CRITICAL] response.items missing");
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
      if (!response.items) throw new Error("[CRITICAL] response.items missing");
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
   * 根据用户积分获取会员等级
   */
  const getUserTier = computed((): MemberTier => {
    const defaultTier: MemberTier = { name: 'Ordinary', min: 0, max: 499, discount: 0 }
    
    if (!userPoints.value) {
      return MEMBER_TIERS.ordinary ?? defaultTier
    }

    const points = userPoints.value.total

    for (const [key, tier] of Object.entries(MEMBER_TIERS)) {
      if (tier.max === null) {
        if (points >= tier.min) return tier
      } else {
        if (points >= tier.min && points <= tier.max) return tier
      }
    }

    return MEMBER_TIERS.ordinary ?? defaultTier
  })

  /**
   * 计算商品小计
   */
  const calculateSubtotal = (items: Array<{ price: number; quantity: number }>) => {
    return items.reduce((sum, item) => sum + item.price * item.quantity, 0)
  }

  /**
   * 计算会员折扣
   */
  const calculateMemberDiscount = (subtotal: number): number => {
    const tier = getUserTier.value
    return subtotal * (tier.discount / 100)
  }

  /**
   * 计算积分抵扣
   * 规则：1 积分 = 0.01 元，最多抵扣订单金额的 50%
   */
  const calculatePointsDiscount = (subtotal: number): number => {
    if (!usePointsDiscount.value || !userPoints.value) {
      return 0
    }

    const maxDiscount = subtotal * 0.5 // 最多抵扣 50%
    const pointsValue = pointsToUse.value * 0.01 // 1 积分 = 0.01 元
    const availablePoints = userPoints.value.available * 0.01

    return Math.min(pointsValue, availablePoints, maxDiscount)
  }

  /**
   * 计算优惠券折扣
   */
  const calculateCouponDiscount = (subtotal: number): number => {
    if (!appliedCoupon.value) {
      return 0
    }

    const coupon = appliedCoupon.value

    // 检查最低消费
    if (coupon.min_amount && subtotal < coupon.min_amount) {
      return 0
    }

    switch (coupon.type) {
      case 'percentage':
        return subtotal * (coupon.value / 100)
      case 'fixed':
        return Math.min(coupon.value, subtotal)
      case 'points':
        // 积分券：直接抵扣
        return Math.min(coupon.value * 0.01, subtotal)
      default:
        return 0
    }
  }

  /**
   * 计算运费
   */
  const calculateShipping = (
    items: Array<{ weight?: number; quantity: number; price: number }>,
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
    items: Array<{ weight?: number; quantity: number; price: number }>,
    subtotal: number,
    address: ShippingAddressInfo
  ): { fee: number; rule: CartShippingTemplate['rules'][0] | null; template: CartShippingTemplate | null } => {
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

      if (!template.rules) throw new Error("[CRITICAL] template.rules missing");
      for (const rule of template.rules) {
        // 检查国家匹配
        if (!rule.regions) throw new Error("[CRITICAL] rule.regions missing");
        const regions = rule.regions
        if (regions.length === 0 || !regions.map(r => r.toUpperCase()).includes(normalizedCountry)) {
          continue
        }

        // 检查邮编匹配
        if (!rule.zip_ranges) throw new Error("[CRITICAL] rule.zip_ranges missing");
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

  /**
   * 计算税费
   */
  const calculateTax = (subtotal: number, shipping: number): number => {
    if (selectedTaxRates.value.length === 0) {
      // 默认税率：10%
      return (subtotal + shipping) * 0.1
    }

    // 累加所有选中的税率
    const totalTaxRate = selectedTaxRates.value.reduce((sum, taxId) => {
      const tax = taxRates.value.find(t => t.id === taxId)
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
    const matchedTaxes = taxRates.value.filter(
      t => t.region && t.region.toLowerCase() === region.toLowerCase()
    )

    selectedTaxRates.value = matchedTaxes.map(t => t.id)
  }

  /**
   * 应用优惠券
   */
  const applyCoupon = async (code: string): Promise<{ success: boolean; message: string }> => {
    try {
      const response = await request<Coupon>(
        '/marketing/coupons/validate',
        {
          method: 'POST',
          body: JSON.stringify({ code }),
          headers: { 'Content-Type': 'application/json', accept: 'application/json' }
        }
      )

      appliedCoupon.value = response
      return { success: true, message: 'Coupon applied successfully' }
    } catch (error) {
      console.error('Failed to apply coupon:', error)
      return { success: false, message: 'Invalid coupon code' }
    }
  }

  /**
   * 移除优惠券
   */
  const removeCoupon = () => {
    appliedCoupon.value = null
  }

  /**
   * 设置使用积分
   */
  const setPointsUsage = (points: number) => {
    if (!userPoints.value) {
      pointsToUse.value = 0
      return
    }

    pointsToUse.value = Math.min(points, userPoints.value.available)
  }

  /**
   * 完整计算购物车总价
   */
  const calculateTotal = (items: Array<{ price: number; quantity: number; weight?: number }>) => {
    // 1. 商品小计
    const subtotal = calculateSubtotal(items)

    // 2. 会员折扣
    const memberDiscount = calculateMemberDiscount(subtotal)

    // 3. 优惠券折扣
    const couponDiscount = calculateCouponDiscount(subtotal)

    // 4. 积分抵扣
    const pointsDiscount = calculatePointsDiscount(subtotal)

    // 5. 折扣后的小计
    const discountedSubtotal = Math.max(
      0,
      subtotal - memberDiscount - couponDiscount - pointsDiscount
    )

    // 6. 运费
    const shipping = calculateShipping(items, discountedSubtotal)

    // 7. 税费（基于折扣后的小计 + 运费）
    const tax = calculateTax(discountedSubtotal, shipping)

    // 8. 最终总计
    const total = discountedSubtotal + shipping + tax

    return {
      subtotal,
      memberDiscount,
      memberTier: getUserTier.value,
      couponDiscount,
      pointsDiscount,
      discountedSubtotal,
      shipping,
      tax,
      total,
      breakdown: {
        originalSubtotal: subtotal,
        totalDiscount: memberDiscount + couponDiscount + pointsDiscount,
        shippingFee: shipping,
        taxFee: tax,
        finalTotal: total,
      }
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
    appliedCoupon,
    usePointsDiscount,
    pointsToUse,
    selectedShippingTemplate,
    selectedTaxRates,
    shippingAddress,
    
    // 计算属性
    getUserTier,
    
    // 方法
    loadShippingTemplates,
    loadTaxRates,
    loadUserPoints,
    calculateSubtotal,
    calculateMemberDiscount,
    calculatePointsDiscount,
    calculateCouponDiscount,
    calculateShipping,
    calculateShippingByRegion,
    calculateTax,
    calculateTotal,
    autoSelectTaxRates,
    applyCoupon,
    removeCoupon,
    setPointsUsage,
    initialize,
  }
}

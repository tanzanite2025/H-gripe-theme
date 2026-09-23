/**
 * 购物车计算系统 - 主入口
 * 集成 site settings 配置
 * 
 * 功能：
 * - 从后端获取税率配置
 * - 计算会员等级折扣
 * - 计算积分抵扣
 * - 支持优惠券
 */

// 导出类型定义
export type {
  MemberTier,
  CartShippingTemplate,
  ShippingAddressInfo,
  TaxRate,
  UserPoints,
  Coupon,
  CartItem,
  ShippingCalculationResult,
  TotalCalculationResult,
} from './cart/types/cart-calculation-types'

// 导出会员等级配置
export { MEMBER_TIERS } from './cart/config/member-tiers'

// 导入模块化composables
import { useCartDataLoader } from './cart/useCartDataLoader'
import { useCartDiscount } from './cart/useCartDiscount'
import { useCartTax } from './cart/useCartTax'
import type { CartItem, TotalCalculationResult } from './cart/types/cart-calculation-types'

export const useCartCalculation = () => {
  // 1. 数据加载模块
  const dataLoader = useCartDataLoader()
  const {
    taxRates,
    userPoints,
    loadTaxRates,
    loadUserPoints,
  } = dataLoader

  // 2. 折扣计算模块
  const discount = useCartDiscount(userPoints)
  const {
    appliedCoupon,
    usePointsDiscount,
    pointsToUse,
    loyaltyPointRedemptionEnabled,
    getUserTier,
    calculateMemberDiscount,
    calculatePointsDiscount,
    calculateCouponDiscount,
    applyCoupon,
    removeCoupon,
    setPointsUsage,
    loadLoyaltyProgramConfig,
  } = discount

  // 3. 税费计算模块
  const tax = useCartTax(taxRates)
  const {
    selectedTaxRates,
    shippingAddress,
    calculateTax,
    autoSelectTaxRates,
  } = tax

  /**
   * 计算商品小计
   */
  const calculateSubtotal = (items: CartItem[]) => {
    return items.reduce((sum, item) => sum + item.price_minor * item.quantity, 0)
  }

  /**
   * 完整计算购物车总价
   */
  const calculateTotal = (items: CartItem[], shippingFeeMinor = 0): TotalCalculationResult => {
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
    const shippingFee = Math.max(0, Number(shippingFeeMinor) || 0)

    // 7. 税费（基于折扣后的小计 + 运费）
    const taxFee = calculateTax(discountedSubtotal, shippingFee)

    // 8. 最终总计
    const total = discountedSubtotal + shippingFee + taxFee

    return {
      subtotal_minor: subtotal,
      member_discount_minor: memberDiscount,
      memberTier: getUserTier.value,
      coupon_discount_minor: couponDiscount,
      points_discount_minor: pointsDiscount,
      discounted_subtotal_minor: discountedSubtotal,
      shipping_minor: shippingFee,
      tax_minor: taxFee,
      total_minor: total,
      breakdown: {
        original_subtotal_minor: subtotal,
        total_discount_minor: memberDiscount + couponDiscount + pointsDiscount,
        shipping_fee_minor: shippingFee,
        tax_fee_minor: taxFee,
        final_total_minor: total,
      }
    }
  }

  /**
   * 初始化（加载所有配置）
   */
  const initialize = async () => {
    await Promise.allSettled([
      dataLoader.initialize(),
      loadLoyaltyProgramConfig(),
    ])
  }

  return {
    // 状态
    taxRates,
    userPoints,
    appliedCoupon,
    usePointsDiscount,
    pointsToUse,
    loyaltyPointRedemptionEnabled,
    selectedTaxRates,
    shippingAddress,
    
    // 计算属性
    getUserTier,
    
    // 方法
    loadTaxRates,
    loadUserPoints,
    calculateSubtotal,
    calculateMemberDiscount,
    calculatePointsDiscount,
    calculateCouponDiscount,
    calculateTax,
    calculateTotal,
    autoSelectTaxRates,
    applyCoupon,
    removeCoupon,
    setPointsUsage,
    loadLoyaltyProgramConfig,
    initialize,
  }
}

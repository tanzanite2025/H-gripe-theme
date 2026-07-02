/**
 * 购物车折扣计算
 * 包含会员折扣、积分抵扣、优惠券折扣
 */
import { ref, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'
import { getTierByPoints } from './config/member-tiers'
import type { UserPoints, Coupon, CouponValidationResponse, MemberTier } from './types/cart-calculation-types'

export const useCartDiscount = (userPoints: ReturnType<typeof ref<UserPoints | null>>) => {
  const { request } = useAuth()

  // 状态
  const appliedCoupon = ref<Coupon | null>(null)
  const usePointsDiscount = ref(false)
  const pointsToUse = ref(0)

  /**
   * 根据用户积分获取会员等级
   */
  const getUserTier = computed((): MemberTier => {
    const defaultTier: MemberTier = { name: 'Ordinary', min: 0, max: 499, discount: 0 }
    
    if (!userPoints.value) {
      return defaultTier
    }

    return getTierByPoints(userPoints.value.total)
  })

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
   * 应用优惠券
   */
  const applyCoupon = async (code: string, amount: number): Promise<{ success: boolean; message: string }> => {
    const normalizedCode = code.trim()
    if (!normalizedCode) {
      return { success: false, message: 'Coupon code is required' }
    }
    if (!Number.isFinite(amount) || amount <= 0) {
      return { success: false, message: 'Cart subtotal must be greater than zero' }
    }

    try {
      const response = await request<CouponValidationResponse>(
        '/marketing/coupons/validate',
        {
          method: 'POST',
          body: JSON.stringify({ code: normalizedCode, amount }),
          headers: { 'Content-Type': 'application/json', accept: 'application/json' }
        }
      )

      appliedCoupon.value = response.coupon
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

  return {
    // 状态
    appliedCoupon,
    usePointsDiscount,
    pointsToUse,

    // 计算属性
    getUserTier,

    // 方法
    calculateMemberDiscount,
    calculatePointsDiscount,
    calculateCouponDiscount,
    applyCoupon,
    removeCoupon,
    setPointsUsage,
  }
}

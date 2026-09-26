import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'
import type { Coupon, CouponValidationResponse } from './types/cart-calculation-types'

export const useCartDiscount = () => {
  const { request } = useAuth()
  const appliedCoupon = ref<Coupon | null>(null)

  const applyCoupon = async (code: string, amountMinor: number): Promise<{ success: boolean; message: string }> => {
    const normalizedCode = code.trim()
    if (!normalizedCode) return { success: false, message: 'Coupon code is required' }
    if (!Number.isSafeInteger(amountMinor) || amountMinor <= 0) {
      return { success: false, message: 'Cart subtotal must be greater than zero' }
    }

    try {
      const response = await request<CouponValidationResponse>('/marketing/coupons/validate', {
        method: 'POST',
        body: JSON.stringify({ code: normalizedCode, amount_minor: amountMinor }),
        headers: { 'Content-Type': 'application/json', accept: 'application/json' },
      })
      if (!response.coupon || !['percentage', 'fixed'].includes(response.coupon.type)) {
        return { success: false, message: 'Invalid coupon code' }
      }
      const discountMinor = Number(response.discount_minor)
      if (!Number.isSafeInteger(discountMinor) || discountMinor < 0) {
        return { success: false, message: 'Invalid coupon response' }
      }
      appliedCoupon.value = { ...response.coupon, discount_minor: discountMinor }
      return { success: true, message: 'Coupon applied successfully' }
    } catch {
      return { success: false, message: 'Invalid coupon code' }
    }
  }

  const removeCoupon = () => {
    appliedCoupon.value = null
  }

  return {
    appliedCoupon,
    applyCoupon,
    removeCoupon,
  }
}

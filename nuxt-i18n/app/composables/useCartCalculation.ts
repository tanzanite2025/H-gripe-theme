/**
 * Cart facts that can be derived safely on the client.
 *
 * Taxes, shipping, membership discounts, coupon discounts and the final
 * payable amount are owned by POST /checkout/quote. This composable only keeps
 * the line-item subtotal and coupon input state needed to build that request.
 */
import { useCartDiscount } from './cart/useCartDiscount'
import type { CartItem } from './cart/types/cart-calculation-types'

export type {
  CartItem,
  Coupon,
  CouponValidationResponse,
} from './cart/types/cart-calculation-types'

export const useCartCalculation = () => {
  const discount = useCartDiscount()

  const calculateSubtotal = (items: CartItem[]) => {
    return items.reduce((sum, item) => {
      const priceMinor = Number(item.price_minor)
      const quantity = Math.max(0, Math.floor(Number(item.quantity)))
      if (!Number.isSafeInteger(priceMinor) || !Number.isSafeInteger(quantity)) return sum
      const lineSubtotal = priceMinor * quantity
      if (!Number.isSafeInteger(lineSubtotal)) return sum
      const nextSubtotal = sum + lineSubtotal
      return Number.isSafeInteger(nextSubtotal) ? nextSubtotal : sum
    }, 0)
  }

  return {
    appliedCoupon: discount.appliedCoupon,
    calculateSubtotal,
    applyCoupon: discount.applyCoupon,
    removeCoupon: discount.removeCoupon,
  }
}

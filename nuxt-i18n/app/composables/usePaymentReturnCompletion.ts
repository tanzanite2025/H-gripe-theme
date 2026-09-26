import { navigateTo, useLocalePath } from '#imports'
import { useCart } from '~/composables/useCart'

export const usePaymentReturnCompletion = () => {
  const localePath = useLocalePath()
  const { clearCart, reloadCartFromBackend } = useCart()

  const completePaymentReturn = async (orderNumber: string) => {
    await clearCart()
    await reloadCartFromBackend()
    await navigateTo({
      path: localePath('/checkout/success'),
      query: { order_number: orderNumber },
    }, { replace: true })
  }

  return { completePaymentReturn }
}

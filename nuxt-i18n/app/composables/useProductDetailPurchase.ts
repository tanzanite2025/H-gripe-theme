import {
  computed,
  onMounted,
  ref,
  toValue,
  watch,
  type MaybeRefOrGetter,
} from 'vue'
import type {
  StripeExpressCheckoutElementConfirmEvent,
  StripeExpressCheckoutElementShippingAddressChangeEvent,
  StripeExpressCheckoutElementShippingRateChangeEvent,
} from '@stripe/stripe-js'
import { useI18n } from '#imports'
import { COUNTRIES } from '~/data/countries'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { usePaymentMethods } from '~/composables/usePaymentMethods'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import { useStripeExpressCheckoutOrder } from '~/composables/useStripeExpressCheckoutOrder'
import {
  convertMajorAmountToStripeMinorAmount,
  type StripeExpressCheckoutAvailablePaymentMethods,
} from '~/composables/useStripeExpressCheckout'
import type { ProductDetailExpressCheckoutExposed } from '~/components/shop/product-detail/ProductDetailExpressCheckout.vue'
import type {
  CheckoutPaymentOption,
} from '~/types/payment'
import {
  isPaymentOptionAvailable,
  paymentMethodFromOption,
  storefrontPaymentMethodOrder,
  type StorefrontPaymentMethod,
} from '~/utils/paymentPresentation'
import {
  normalizeProductCurrencyCode,
} from '~/utils/productDetail'
import type {
  GoProduct,
  ProductVariant,
} from '~/types/productDetail'

export interface ProductDetailPurchaseOptions {
  product: MaybeRefOrGetter<GoProduct | null | undefined>
  shopProduct: MaybeRefOrGetter<ShopProduct | null | undefined>
  selectedVariant: MaybeRefOrGetter<ProductVariant | null>
  selectedVariantWeight: MaybeRefOrGetter<number | null>
  selectedCartTitle: MaybeRefOrGetter<string>
  effectivePrice: MaybeRefOrGetter<number>
  currentCurrency: MaybeRefOrGetter<string>
  selectedAvailability: MaybeRefOrGetter<'in_stock' | 'out_of_stock'>
  primaryMediaThumbnail: MaybeRefOrGetter<string>
}

export function useProductDetailPurchase(options: ProductDetailPurchaseOptions) {
  const { t } = useI18n()
  const auth = useAuth()
  const {
    addToCart,
    openCart,
    openCheckout,
    cartItems,
    cartCurrency,
    clearCart,
  } = useCart()
  const { toCartItem } = useShopProducts()
  const { countryCode } = useStorefrontContext()
  const {
    paymentMethodOptions,
    paymentMethodsLoading,
    paymentMethodsError,
    loadPaymentMethods,
  } = usePaymentMethods()
  const {
    loadStripeExpressCheckoutPublishableKey,
    createStripeExpressCheckoutOrderAndPaymentSession,
  } = useStripeExpressCheckoutOrder()

  const maxProductQuantity = 99
  const selectedQuantity = ref(1)
  const selectedProductPaymentMethod = ref<StorefrontPaymentMethod>('card')

  const product = computed(() => toValue(options.product) || null)
  const shopProduct = computed(() => toValue(options.shopProduct) || null)
  const selectedVariant = computed(() => toValue(options.selectedVariant) || null)
  const selectedVariantWeight = computed(() => toValue(options.selectedVariantWeight) || null)
  const selectedCartTitle = computed(() => toValue(options.selectedCartTitle) || '')
  const effectivePrice = computed(() => Number(toValue(options.effectivePrice) || 0))
  const currentCurrency = computed(() => toValue(options.currentCurrency) || 'USD')
  const selectedAvailability = computed(() => toValue(options.selectedAvailability) || 'out_of_stock')
  const primaryMediaThumbnail = computed(() => toValue(options.primaryMediaThumbnail) || '')

  const canAddToCart = computed(() => Boolean(
    product.value
    && effectivePrice.value > 0
    && selectedAvailability.value === 'in_stock',
  ))

  const normalizeSelectedQuantity = (value: unknown) => {
    const numeric = Math.floor(Number(value))
    if (!Number.isFinite(numeric)) return 1
    return Math.min(maxProductQuantity, Math.max(1, numeric))
  }

  const setSelectedQuantity = (value: unknown) => {
    selectedQuantity.value = normalizeSelectedQuantity(value)
  }

  const decreaseSelectedQuantity = () => {
    setSelectedQuantity(selectedQuantity.value - 1)
  }

  const increaseSelectedQuantity = () => {
    setSelectedQuantity(selectedQuantity.value + 1)
  }

  const fallbackProductPaymentOptions = computed<CheckoutPaymentOption[]>(() => [
    {
      id: 'card',
      code: 'card',
      provider: 'stripe',
      title: 'Credit / Debit cards',
      subtitle: '',
      description: '',
      enabled: true,
      available: false,
      unavailableReason: 'gateway_not_configured',
    },
    {
      id: 'paypal',
      code: 'paypal',
      provider: 'paypal',
      title: 'PayPal',
      subtitle: '',
      description: '',
      enabled: true,
      available: false,
      unavailableReason: 'gateway_not_configured',
    },
    {
      id: 'alipay',
      code: 'alipay',
      provider: 'alipay',
      title: 'Alipay',
      subtitle: '',
      description: '',
      enabled: true,
      available: false,
      unavailableReason: 'gateway_not_configured',
    },
    {
      id: 'wechat',
      code: 'wechat',
      provider: 'wechat',
      title: 'WeChat Pay',
      subtitle: '',
      description: '',
      enabled: true,
      available: false,
      unavailableReason: 'gateway_not_configured',
    },
  ])

  const productPaymentMethod = (option: CheckoutPaymentOption) => paymentMethodFromOption(option)
  const productPaymentOptions = computed(() => {
    const optionsByMethod = new Map<string, CheckoutPaymentOption>(
      fallbackProductPaymentOptions.value.map(option => [productPaymentMethod(option), option]),
    )

    paymentMethodOptions.value.forEach(option => {
      const method = productPaymentMethod(option)
      if (method) optionsByMethod.set(method, option)
    })

    return storefrontPaymentMethodOrder
      .map(method => optionsByMethod.get(method))
      .filter((option): option is CheckoutPaymentOption => Boolean(option))
  })

  const selectedProductPaymentOption = computed(() => {
    return productPaymentOptions.value.find(option => (
      productPaymentMethod(option) === selectedProductPaymentMethod.value
    )) || null
  })

  const canBuyNow = computed(() => Boolean(
    canAddToCart.value
    && selectedProductPaymentOption.value
    && isPaymentOptionAvailable(selectedProductPaymentOption.value),
  ))

  const selectProductPaymentMethod = (method: StorefrontPaymentMethod | '') => {
    if (method) selectedProductPaymentMethod.value = method
  }

  const stripeExpressCheckoutPublishableKey = ref('')
  const stripeExpressCheckoutError = ref('')
  const isStripeExpressCheckoutProcessing = ref(false)
  const isStripeExpressCheckoutWalletAvailable = ref(true)
  const stripeExpressCheckoutElementRef = ref<ProductDetailExpressCheckoutExposed | null>(null)
  let stripeExpressCheckoutPreparationPromise: Promise<void> | null = null

  const selectedExpressCheckoutCartItem = computed(() => {
    if (!product.value || !shopProduct.value || !canAddToCart.value) return null

    return {
      ...toCartItem(shopProduct.value, {
        variantId: selectedVariant.value?.id || null,
        price: effectivePrice.value,
        salePrice: selectedVariant.value?.sale_price ?? product.value.sale_price ?? null,
        sku: selectedVariant.value?.sku || product.value.sku || '',
        currency: currentCurrency.value,
        title: selectedCartTitle.value,
        thumbnail: primaryMediaThumbnail.value || undefined,
        weightGrams: selectedVariantWeight.value,
      }),
      quantity: selectedQuantity.value,
    }
  })

  const stripeExpressCheckoutCartItems = computed(() => {
    const items = cartItems.value.map(item => ({ ...item }))
    const selectedItem = selectedExpressCheckoutCartItem.value
    if (!selectedItem) return items

    const existingItem = items.find(item => (
      Number(item.product_id || item.id) === Number(selectedItem.product_id)
      && Number(item.variant_id || 0) === Number(selectedItem.variant_id || 0)
    ))
    if (existingItem) {
      existingItem.quantity += selectedItem.quantity
      return items
    }

    items.push({
      ...selectedItem,
      id: Number(selectedItem.variant_id || selectedItem.product_id || 0),
    })
    return items
  })

  const stripeExpressCheckoutAmount = computed(() => (
    stripeExpressCheckoutCartItems.value.reduce(
      (total, item) => total + Number(item.price || 0) * Math.max(1, Number(item.quantity || 1)),
      0,
    )
  ))

  const stripeExpressCheckoutLineItems = computed(() => (
    stripeExpressCheckoutCartItems.value.map(item => ({
      name: item.title,
      amount: convertMajorAmountToStripeMinorAmount(
        Number(item.price || 0) * Math.max(1, Number(item.quantity || 1)),
        cartCurrency.value,
      ),
    }))
  ))

  const stripeExpressCheckoutAllowedShippingCountries = computed(() => (
    COUNTRIES.map(country => country.code)
  ))

  const stripeCardPaymentAvailable = computed(() => (
    paymentMethodOptions.value.some((option) => {
      const provider = String(option.provider || '').trim().toLowerCase()
      const code = String(option.code || option.id || '').trim().toLowerCase()
      return (
        provider === 'stripe'
        || ['card', 'credit_card', 'credit-card', 'stripe'].includes(code)
      )
        && option.enabled !== false
        && option.available === true
    })
  ))

  const shouldPrepareStripeExpressCheckout = computed(() => Boolean(
    auth.isAuthenticated.value
    && stripeCardPaymentAvailable.value
    && stripeExpressCheckoutPublishableKey.value
    && stripeExpressCheckoutAmount.value > 0
    && canAddToCart.value
    && isStripeExpressCheckoutWalletAvailable.value,
  ))

  const productBuyNowUnavailableLabel = computed(() => {
    if (!canAddToCart.value) return t('products.detail.outOfStock', 'Out of stock')
    if (paymentMethodsLoading.value) return t('common.loading', 'Loading...')
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  })

  const addSelectedProductToCart = () => {
    if (!product.value || !shopProduct.value || !canAddToCart.value) return

    return addToCart(toCartItem(shopProduct.value, {
      variantId: selectedVariant.value?.id || null,
      price: effectivePrice.value,
      salePrice: selectedVariant.value?.sale_price ?? product.value.sale_price ?? null,
      sku: selectedVariant.value?.sku || product.value.sku || '',
      currency: normalizeProductCurrencyCode(currentCurrency.value) || 'USD',
      title: selectedCartTitle.value,
      thumbnail: primaryMediaThumbnail.value || undefined,
      weightGrams: selectedVariantWeight.value,
    }), selectedQuantity.value)
  }

  const addSelectedToCart = () => {
    const result = addSelectedProductToCart()
    if (result?.success) openCart()
  }

  const checkoutSelectedWithPayment = () => {
    const result = addSelectedProductToCart()
    if (result?.success && selectedProductPaymentMethod.value) {
      openCheckout(selectedProductPaymentMethod.value)
    }
  }

  const loadProductPaymentMethods = async () => {
    const marketCountry = countryCode.value && countryCode.value !== 'ZZ'
      ? countryCode.value
      : undefined
    await loadPaymentMethods(marketCountry)
  }

  const prepareStripeExpressCheckout = async () => {
    if (
      stripeExpressCheckoutPublishableKey.value
      || stripeExpressCheckoutPreparationPromise
      || !canAddToCart.value
    ) {
      return
    }

    const preparation = (async () => {
      const session = await auth.ensureSession()
      if (!session || !canAddToCart.value) return

      const marketCountry = countryCode.value && countryCode.value !== 'ZZ'
        ? countryCode.value
        : undefined
      await loadPaymentMethods(marketCountry)
      if (!stripeCardPaymentAvailable.value) return

      stripeExpressCheckoutPublishableKey.value = await loadStripeExpressCheckoutPublishableKey()
    })()

    stripeExpressCheckoutPreparationPromise = preparation
    try {
      await preparation
    } catch (error) {
      stripeExpressCheckoutPublishableKey.value = ''
      stripeExpressCheckoutError.value = error instanceof Error
        ? error.message
        : 'Express Checkout is unavailable'
    } finally {
      if (stripeExpressCheckoutPreparationPromise === preparation) {
        stripeExpressCheckoutPreparationPromise = null
      }
    }
  }

  const handleStripeExpressCheckoutAvailability = (
    availablePaymentMethods: StripeExpressCheckoutAvailablePaymentMethods,
  ) => {
    isStripeExpressCheckoutWalletAvailable.value = (
      availablePaymentMethods.applePay || availablePaymentMethods.googlePay
    )
  }

  const handleStripeExpressCheckoutShippingAddressChange = (
    shippingEvent: StripeExpressCheckoutElementShippingAddressChangeEvent,
  ) => {
    shippingEvent.resolve({
      lineItems: stripeExpressCheckoutLineItems.value,
    })
  }

  const handleStripeExpressCheckoutShippingRateChange = (
    shippingEvent: StripeExpressCheckoutElementShippingRateChangeEvent,
  ) => {
    shippingEvent.resolve({
      lineItems: stripeExpressCheckoutLineItems.value,
    })
  }

  const handleStripeExpressCheckoutConfirm = async (
    confirmationEvent: StripeExpressCheckoutElementConfirmEvent,
  ) => {
    if (isStripeExpressCheckoutProcessing.value) {
      confirmationEvent.paymentFailed({
        reason: 'fail',
        message: 'Another payment is already being processed',
      })
      return
    }

    const expressCheckoutElement = stripeExpressCheckoutElementRef.value
    if (!expressCheckoutElement) {
      confirmationEvent.paymentFailed({
        reason: 'fail',
        message: 'Express Checkout is not ready',
      })
      return
    }

    isStripeExpressCheckoutProcessing.value = true
    stripeExpressCheckoutError.value = ''
    try {
      await expressCheckoutElement.submitExpressCheckoutPayment()

      const addResult = addSelectedProductToCart()
      if (!addResult?.success) {
        throw new Error('The selected product could not be added to the order')
      }
      await addResult.syncPromise

      const session = await createStripeExpressCheckoutOrderAndPaymentSession(
        confirmationEvent,
        cartItems.value,
      )
      const returnUrl = new URL(window.location.href)
      returnUrl.searchParams.set('order_number', session.orderNumber)
      const result = await expressCheckoutElement.confirmExpressCheckoutPayment(
        session.clientSecret,
        returnUrl.toString(),
      )
      if (['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
        clearCart()
        return
      }

      throw new Error(t(
        'checkout.modal.messages.paymentPending',
        'Payment is not complete yet. Please check your order status later.',
      ))
    } catch (error) {
      expressCheckoutElement.resetExpressCheckoutPaymentState()
      const message = error instanceof Error
        ? error.message
        : t('checkout.modal.messages.orderFailed', 'Order submission failed, please try again')
      stripeExpressCheckoutError.value = message
      confirmationEvent.paymentFailed({ reason: 'fail', message })
    } finally {
      isStripeExpressCheckoutProcessing.value = false
    }
  }

  const handleStripeExpressCheckoutError = (message: string) => {
    stripeExpressCheckoutError.value = message
  }

  watch(productPaymentOptions, (options) => {
    const selectedExists = options.some(option => (
      productPaymentMethod(option) === selectedProductPaymentMethod.value
    ))
    if (selectedExists) return

    const firstAvailable = options.find(isPaymentOptionAvailable) || options[0]
    const method = firstAvailable ? productPaymentMethod(firstAvailable) : ''
    if (method) selectedProductPaymentMethod.value = method
  }, { immediate: true })

  watch(
    [canAddToCart, () => auth.isAuthenticated.value, paymentMethodsLoading],
    ([canAdd, isAuthenticated, loading]) => {
      if (canAdd && isAuthenticated && !loading) {
        void prepareStripeExpressCheckout()
      }
    },
    { flush: 'post' },
  )

  onMounted(() => {
    void loadProductPaymentMethods()
  })

  watch(countryCode, (value, previousValue) => {
    if (!import.meta.client || value === previousValue) return
    void loadProductPaymentMethods()
  })

  return {
    maxProductQuantity,
    selectedQuantity,
    setSelectedQuantity,
    decreaseSelectedQuantity,
    increaseSelectedQuantity,
    canAddToCart,
    canBuyNow,
    productBuyNowUnavailableLabel,
    selectedProductPaymentMethod,
    productPaymentOptions,
    paymentMethodsLoading,
    paymentMethodsError,
    selectProductPaymentMethod,
    stripeExpressCheckoutPublishableKey,
    stripeExpressCheckoutError,
    isStripeExpressCheckoutProcessing,
    stripeExpressCheckoutElementRef,
    stripeExpressCheckoutAmount,
    stripeExpressCheckoutLineItems,
    stripeExpressCheckoutAllowedShippingCountries,
    shouldPrepareStripeExpressCheckout,
    addSelectedToCart,
    checkoutSelectedWithPayment,
    prepareStripeExpressCheckout,
    handleStripeExpressCheckoutAvailability,
    handleStripeExpressCheckoutShippingAddressChange,
    handleStripeExpressCheckoutShippingRateChange,
    handleStripeExpressCheckoutConfirm,
    handleStripeExpressCheckoutError,
    cartCurrency,
  }
}

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
import { navigateTo, useI18n, useLocalePath, useRequestURL } from '#imports'
import { COUNTRIES } from '~/data/countries'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { usePaymentMethods } from '~/composables/usePaymentMethods'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import { useStripeExpressCheckoutOrder } from '~/composables/useStripeExpressCheckoutOrder'
import {
  type StripeExpressCheckoutAvailablePaymentMethods,
} from '~/composables/useStripeExpressCheckout'
import {
  STRIPE_RETURN_PATH,
  clearStripeReturnSession,
  saveStripeReturnSession,
} from '~/utils/stripeReturn'
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
import { majorToMinor, minorToMajor } from '~/utils/money'
import type {
  GoProduct,
  ProductAvailability,
  ProductVariant,
} from '~/types/productDetail'

export interface ProductDetailPurchaseOptions {
  product: MaybeRefOrGetter<GoProduct | null | undefined>
  shopProduct: MaybeRefOrGetter<ShopProduct | null | undefined>
  selectedVariant: MaybeRefOrGetter<ProductVariant | null>
  selectedVariantWeight: MaybeRefOrGetter<number | null>
  selectedCartTitle: MaybeRefOrGetter<string>
  effectivePriceMinor?: MaybeRefOrGetter<number>
  effectivePrice: MaybeRefOrGetter<number>
  currentCurrency: MaybeRefOrGetter<string>
  selectedAvailability: MaybeRefOrGetter<ProductAvailability>
  primaryMediaThumbnail: MaybeRefOrGetter<string>
  selectedOptions?: MaybeRefOrGetter<Array<{ group_slug: string; value_keys: string[] }>>
  customOptionsValid?: MaybeRefOrGetter<boolean>
}

export function useProductDetailPurchase(options: ProductDetailPurchaseOptions) {
  const { t } = useI18n()
  const localePath = useLocalePath()
  const requestUrl = useRequestURL()
  const auth = useAuth()
  const {
    addToCart,
    openCart,
    openCheckout,
    cartItems,
    cartCurrency,
    clearCart,
    reloadCartFromBackend,
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
  const effectivePriceMinor = computed(() => {
    const provided = options.effectivePriceMinor == null
      ? null
      : Number(toValue(options.effectivePriceMinor))
    return provided != null && Number.isFinite(provided)
      ? Math.round(provided)
      : majorToMinor(effectivePrice.value, currentCurrency.value)
  })
  const selectedAvailability = computed<ProductAvailability>(() => toValue(options.selectedAvailability) || 'out_of_stock')
  const primaryMediaThumbnail = computed(() => toValue(options.primaryMediaThumbnail) || '')
  const selectedOptions = computed(() => toValue(options.selectedOptions) || [])
  const customOptionsValid = computed(() => options.customOptionsValid == null
    ? true
    : Boolean(toValue(options.customOptionsValid)))

  const canAddToCart = computed(() => Boolean(
    product.value
    && effectivePrice.value > 0
    && ['in_stock', 'made_to_order'].includes(selectedAvailability.value)
    && customOptionsValid.value,
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
        priceMinor: effectivePriceMinor.value,
        sku: selectedVariant.value?.sku || product.value.sku || '',
        currency: currentCurrency.value,
        title: selectedCartTitle.value,
        thumbnail: primaryMediaThumbnail.value || undefined,
        weightGrams: selectedVariantWeight.value,
        fulfillmentMode: shopProduct.value.fulfillmentMode,
        selectedOptions: selectedOptions.value,
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
    minorToMajor(stripeExpressCheckoutCartItems.value.reduce(
      (total, item) => total + Number(item.price_minor || 0) * Math.max(1, Number(item.quantity || 1)),
      0,
    ), cartCurrency.value)
  ))

  const stripeExpressCheckoutLineItems = computed(() => (
    stripeExpressCheckoutCartItems.value.map(item => ({
      name: item.title,
      amount: Number(item.price_minor || 0) * Math.max(1, Number(item.quantity || 1)),
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
      priceMinor: effectivePriceMinor.value,
      sku: selectedVariant.value?.sku || product.value.sku || '',
      currency: normalizeProductCurrencyCode(currentCurrency.value) || 'USD',
      title: selectedCartTitle.value,
      thumbnail: primaryMediaThumbnail.value || undefined,
      weightGrams: selectedVariantWeight.value,
      fulfillmentMode: shopProduct.value.fulfillmentMode,
      selectedOptions: selectedOptions.value,
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
    await loadPaymentMethods(marketCountry, cartCurrency.value)
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
      await loadPaymentMethods(marketCountry, cartCurrency.value)
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
      await expressCheckoutElement.updateExpressCheckoutPaymentAmount(session.amountMinor, session.currency)
      saveStripeReturnSession({
        orderNumber: session.orderNumber,
        clientSecret: session.clientSecret,
        publishableKey: session.publishableKey,
      })
      const returnUrl = new URL(localePath(STRIPE_RETURN_PATH), requestUrl.origin)
      returnUrl.searchParams.set('order_number', session.orderNumber)
      const result = await expressCheckoutElement.confirmExpressCheckoutPayment(
        session.clientSecret,
        returnUrl.toString(),
      )
      if (['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
        clearStripeReturnSession(session.orderNumber)
        await clearCart()
        await reloadCartFromBackend()
        await navigateTo({
          path: localePath('/checkout/success'),
          query: { order_number: session.orderNumber },
        })
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

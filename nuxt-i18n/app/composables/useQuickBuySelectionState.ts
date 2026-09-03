import { computed, ref } from 'vue'
import { useQuickBuySession } from '~/composables/useQuickBuySession'
import type { ShopProduct } from '~/composables/useShopProducts'
import type { QuickBuyStep, QuickBuySessionItem } from '~/utils/quickBuy/types'
import type {
  QuickBuySelectedProduct,
  QuickBuySelectedProductStepSlot,
} from '~/utils/quickBuy/selection'

const QUICK_BUY_SELECTED_PRODUCT_SLOT_COUNT = 5

export function useQuickBuySelectionState(
  steps: Readonly<{ value: QuickBuyStep[] }>,
  hasConfiguredFlow: Readonly<{ value: boolean }>,
  quickBuySession = useQuickBuySession('dock'),
) {
  const selectedProducts = ref<QuickBuySelectedProduct[]>([])
  const selectionMutationVersions = new Map<string, number>()
  const confirmedStepSelections = new Map<string, QuickBuySelectedProduct[]>()
  let selectionSaveQueue = Promise.resolve(true)
  let hydratedSessionToken = ''

  const {
    session,
    createSession,
    fetchStepCandidates,
    updateSessionSelections,
    error: quickBuySessionError,
  } = quickBuySession

  const snapshotValue = (snapshot: Record<string, unknown> | undefined, ...keys: string[]) => {
    if (!snapshot) return undefined
    for (const key of keys) {
      if (snapshot[key] !== undefined && snapshot[key] !== null) return snapshot[key]
    }
    return undefined
  }

  const snapshotString = (snapshot: Record<string, unknown> | undefined, ...keys: string[]) => {
    const value = snapshotValue(snapshot, ...keys)
    return value === undefined ? '' : String(value).trim()
  }

  const snapshotNumber = (snapshot: Record<string, unknown> | undefined, ...keys: string[]) => {
    const value = snapshotValue(snapshot, ...keys)
    const numberValue = Number(value)
    return Number.isFinite(numberValue) ? numberValue : 0
  }

  const selectedProductFromSessionItem = (item: QuickBuySessionItem): QuickBuySelectedProduct => {
    const productSnapshot = item.productSnapshot
    const variantSnapshot = item.variantSnapshot
    const productId = Number(item.productId)
    const variantId = item.variantId ? Number(item.variantId) : null
    const unitPrice = Number(item.unitPriceSnapshot)
      || snapshotNumber(variantSnapshot, 'sale_price', 'price')
      || snapshotNumber(productSnapshot, 'sale_price', 'price')

    return {
      productId,
      stepKey: item.stepKey,
      variantId,
      title: snapshotString(productSnapshot, 'name', 'title')
        || snapshotString(variantSnapshot, 'title')
        || `#${productId}`,
      slug: snapshotString(productSnapshot, 'slug') || String(productId),
      sku: snapshotString(variantSnapshot, 'sku') || snapshotString(productSnapshot, 'sku') || undefined,
      thumbnail: snapshotString(productSnapshot, 'thumbnail', 'featured_image') || '',
      quantity: Number(item.quantity) > 0 ? Number(item.quantity) : 1,
      weightGrams: Number(item.weightSnapshotG) || snapshotNumber(variantSnapshot, 'weight_grams'),
      unitPrice,
      currency: item.currencySnapshot
        || snapshotString(variantSnapshot, 'currency')
        || snapshotString(productSnapshot, 'currency')
        || 'USD',
    }
  }

  const hydrateSelectedProductsFromSession = (
    activeSession: { sessionToken: string, items: QuickBuySessionItem[] } | null,
  ) => {
    if (!activeSession?.sessionToken || hydratedSessionToken === activeSession.sessionToken) return

    const hydratedSelections = activeSession.items
      .map(selectedProductFromSessionItem)
      .filter(item => item.productId > 0 && item.stepKey)

    selectedProducts.value = hydratedSelections
    confirmedStepSelections.clear()
    for (const item of hydratedSelections) {
      const stepSelections = confirmedStepSelections.get(item.stepKey) || []
      confirmedStepSelections.set(item.stepKey, [...stepSelections, item])
    }
    hydratedSessionToken = activeSession.sessionToken
  }

  const getStepSelections = (stepKey: string) =>
    selectedProducts.value.filter(item => item.stepKey === stepKey)

  const stepLabelFor = (stepKey: string) => {
    const stepIndex = steps.value.findIndex((step, index) =>
      (step.stepKey || step.slug || `step-${index + 1}`) === stepKey,
    )
    if (stepIndex < 0) return ''
    const step = steps.value[stepIndex]
    return step?.name || `Step ${stepIndex + 1}`
  }

  const selectedProductSlots = computed<QuickBuySelectedProductStepSlot[]>(() =>
    Array.from({ length: QUICK_BUY_SELECTED_PRODUCT_SLOT_COUNT }, (_, index) => {
      const step = steps.value[index]
      const stepKey = step
        ? step.stepKey || step.slug || `step-${index + 1}`
        : `step-${index + 1}`
      const stepSelections = step ? getStepSelections(stepKey) : []

      return {
        slotKey: `quick-buy-selected-step-slot-${index + 1}`,
        index: index + 1,
        stepKey,
        stepLabel: step?.name || '',
        item: stepSelections[0] || null,
        additionalItemCount: Math.max(0, stepSelections.length - 1),
      }
    }),
  )

  const totalQty = computed(() =>
    selectedProducts.value.reduce((sum, item) => sum + (Number(item.quantity) || 0), 0),
  )

  const totalWeightG = computed(() =>
    selectedProducts.value.reduce(
      (sum, item) => sum + (Number(item.weightGrams) || 0) * (Number(item.quantity) || 0),
      0,
    ),
  )

  const totalPrice = computed(() =>
    selectedProducts.value.reduce(
      (sum, item) => sum + (Number(item.unitPrice) || 0) * (Number(item.quantity) || 0),
      0,
    ),
  )

  const createSelectedProduct = (product: ShopProduct, stepKey: string): QuickBuySelectedProduct => {
    const selectedVariant = product.defaultVariantId
      ? product.variants.find(variant => Number(variant.id) === Number(product.defaultVariantId)) || null
      : product.variants.find(variant => variant.isDefault) || product.variants[0] || null
    const selectedVariantId = selectedVariant?.id
      ? Number(selectedVariant.id)
      : Number(product.defaultVariantId) || null

    return {
      productId: product.id,
      stepKey,
      variantId: selectedVariantId,
      title: product.title,
      slug: product.slug,
      sku: selectedVariant?.sku || product.sku,
      thumbnail: selectedVariant?.thumbnail || selectedVariant?.image || product.thumbnail || '',
      quantity: 1,
      weightGrams: selectedVariant?.weightGrams || 0,
      // QUICK selections are later submitted as cart lines. Keep the
      // catalog/source amount here; display snapshots are presentation only.
      unitPrice: selectedVariant?.priceNumber ?? product.priceNumber,
      currency: selectedVariant?.currency || product.currency,
    }
  }

  const isSameSelectedProduct = (
    left: Pick<QuickBuySelectedProduct, 'productId' | 'variantId'>,
    right: Pick<QuickBuySelectedProduct, 'productId' | 'variantId'>,
  ) =>
    left.productId === right.productId
    && Number(left.variantId || 0) === Number(right.variantId || 0)

  const isProductSelectedForStep = (product: ShopProduct, stepKey: string) => {
    const selectedProduct = createSelectedProduct(product, stepKey)
    return getStepSelections(stepKey).some(item => isSameSelectedProduct(item, selectedProduct))
  }

  const buildSessionSelectionInputs = (
    stepKey: string,
    stepSelections: QuickBuySelectedProduct[],
  ) => {
    if (!stepSelections.length) {
      return [{ stepKey, productId: 0, variantId: null, quantity: 0 }]
    }
    return stepSelections.map(item => ({
      stepKey: item.stepKey,
      productId: item.productId,
      variantId: item.variantId,
      quantity: item.quantity,
    }))
  }

  const replaceStepSelectionsLocally = (
    stepKey: string,
    stepSelections: QuickBuySelectedProduct[],
  ) => {
    selectedProducts.value = [
      ...selectedProducts.value.filter(item => item.stepKey !== stepKey),
      ...stepSelections,
    ]
  }

  const saveStepSelections = (
    stepKey: string,
    nextStepSelections: QuickBuySelectedProduct[],
  ) => {
    const mutationVersion = (selectionMutationVersions.get(stepKey) || 0) + 1
    selectionMutationVersions.set(stepKey, mutationVersion)
    const previousStepSelections = getStepSelections(stepKey)
    if (!confirmedStepSelections.has(stepKey)) {
      confirmedStepSelections.set(stepKey, previousStepSelections.slice())
    }
    replaceStepSelectionsLocally(stepKey, nextStepSelections)
    if (!hasConfiguredFlow.value) return Promise.resolve(true)

    selectionSaveQueue = selectionSaveQueue.then(async () => {
      const nextSession = await updateSessionSelections(
        buildSessionSelectionInputs(stepKey, nextStepSelections),
      )
      if (nextSession) {
        confirmedStepSelections.set(stepKey, nextStepSelections.slice())
        return true
      }
      if (!nextSession && selectionMutationVersions.get(stepKey) === mutationVersion) {
        replaceStepSelectionsLocally(stepKey, confirmedStepSelections.get(stepKey) || [])
      }
      return false
    })
    return selectionSaveQueue
  }

  const toggleProductSelection = async (
    product: ShopProduct,
    stepKey: string,
  ) => {
    const selectedProduct = createSelectedProduct(product, stepKey)
    const currentStepSelections = getStepSelections(stepKey)
    const isSelected = currentStepSelections.some(item =>
      isSameSelectedProduct(item, selectedProduct),
    )
    const nextStepSelections = isSelected
      ? currentStepSelections.filter(item => !isSameSelectedProduct(item, selectedProduct))
      : [selectedProduct]

    return saveStepSelections(stepKey, nextStepSelections)
  }

  const removeSelectedProduct = async (selectedProduct: QuickBuySelectedProduct) => {
    const nextStepSelections = getStepSelections(selectedProduct.stepKey)
      .filter(item => !isSameSelectedProduct(item, selectedProduct))
    await saveStepSelections(selectedProduct.stepKey, nextStepSelections)
  }

  const normalizeSelectedQuantity = (value: unknown) => {
    const parsed = Math.floor(Number(value))
    if (!Number.isFinite(parsed)) return 1
    return Math.min(999, Math.max(1, parsed))
  }

  const setSelectedProductQuantity = async (
    selectedProduct: QuickBuySelectedProduct,
    quantity: number,
  ) => {
    const nextQuantity = normalizeSelectedQuantity(quantity)
    if (nextQuantity === selectedProduct.quantity) return

    const nextStepSelections = getStepSelections(selectedProduct.stepKey).map(item =>
      isSameSelectedProduct(item, selectedProduct)
        ? { ...item, quantity: nextQuantity }
        : item,
    )
    await saveStepSelections(selectedProduct.stepKey, nextStepSelections)
  }

  const changeSelectedProductQuantity = async (
    selectedProduct: QuickBuySelectedProduct,
    delta: number,
  ) => {
    await setSelectedProductQuantity(selectedProduct, selectedProduct.quantity + delta)
  }

  const updateSelectedProductQuantity = async (
    selectedProduct: QuickBuySelectedProduct,
    value: unknown,
  ) => {
    await setSelectedProductQuantity(selectedProduct, normalizeSelectedQuantity(value))
  }

  const resetHydration = () => {
    hydratedSessionToken = ''
  }

  return {
    session,
    createSession,
    fetchStepCandidates,
    selectedProducts,
    selectedProductSlots,
    totalQty,
    totalWeightG,
    totalPrice,
    quickBuySessionError,
    hydrateSelectedProductsFromSession,
    getStepSelections,
    stepLabelFor,
    createSelectedProduct,
    isProductSelectedForStep,
    toggleProductSelection,
    removeSelectedProduct,
    changeSelectedProductQuantity,
    updateSelectedProductQuantity,
    resetHydration,
  }
}

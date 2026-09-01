import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import procurementProfitabilityApi, {
  type ProfitabilityBulkUpsertResult,
  type ProfitabilityItemPayload,
  type ProfitabilityProcurementPayload,
} from '@/api/procurementProfitability'
import type { ProductVariantForm } from '@/modules/product/productEditorTypes'

export interface ProcurementProfitDraft {
  productCode: string
  productName: string
  purchasePrice: number | null
  purchasePriceKnown: boolean
  currency: string
  supplierName: string
  supplierContactName: string
  supplierPhone: string
  supplierEmail: string
  leadTimeDays: number
  minimumOrderQuantity: number
  inboundShippingUnitCost: number
  packagingUnitCost: number
  otherUnitCost: number
}

interface DraftSaveResult {
  success: boolean
  result?: ProfitabilityBulkUpsertResult
  error?: unknown
}

interface ProductSnapshot {
  name?: string
  currency?: string
  variants?: ProductVariantForm[]
}

const normalizeCode = (value: unknown): string => String(value || '').trim()
const normalizeCurrency = (value: unknown): string => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : 'USD'
}

const createEmptyDraft = (
  productCode: string,
  productName: string,
  currency: string,
): ProcurementProfitDraft => ({
  productCode,
  productName,
  purchasePrice: null,
  purchasePriceKnown: false,
  currency: normalizeCurrency(currency),
  supplierName: '',
  supplierContactName: '',
  supplierPhone: '',
  supplierEmail: '',
  leadTimeDays: 0,
  minimumOrderQuantity: 1,
  inboundShippingUnitCost: 0,
  packagingUnitCost: 0,
  otherUnitCost: 0,
})

const errorMessage = (error: unknown, fallback: string): string => {
  const value = error as {
    response?: { data?: { error?: string; message?: string } }
    message?: string
  }
  return value?.response?.data?.error || value?.response?.data?.message || value?.message || fallback
}

export const useProcurementProfitDraft = () => {
  const draftsByKey = reactive<Record<string, ProcurementProfitDraft>>({})
  const variantStorageKeys = new WeakMap<object, string>()
  const loading = ref(false)
  const saving = ref(false)
  const pending = ref(false)
  const loadError = ref('')
  const saveError = ref('')
  const lastSavedAt = ref('')
  const pendingItems = ref<ProfitabilityItemPayload[] | null>(null)

  const clearSessionState = () => {
    loading.value = false
    saving.value = false
    pending.value = false
    loadError.value = ''
    saveError.value = ''
    lastSavedAt.value = ''
    pendingItems.value = null
  }

  const reset = () => {
    Object.keys(draftsByKey).forEach((key) => delete draftsByKey[key])
    clearSessionState()
  }

  const variantIdentity = (variant: ProductVariantForm, index: number): string => {
    if (variant.id != null && String(variant.id).trim()) return `id:${variant.id}`
    const object = variant as object
    const existing = variantStorageKeys.get(object)
    if (existing) return existing
    const next = `new:${index}:${Math.random().toString(36).slice(2)}`
    variantStorageKeys.set(object, next)
    return next
  }

  const storageKeyForVariant = (
    variant: ProductVariantForm,
    index: number,
    productName: string,
    currency: string,
  ): string => {
    const identity = variantIdentity(variant, index)
    const code = normalizeCode(variant.sku)
    const previousKey = draftsByKey[identity]
      ? identity
      : Object.keys(draftsByKey).find((key) => draftsByKey[key].productCode === code)

    if (code) {
      if (!draftsByKey[code]) {
        const blankKey = `${identity}:blank`
        if (draftsByKey[blankKey]) {
          draftsByKey[code] = {
            ...draftsByKey[blankKey],
            productCode: code,
            productName,
            currency: normalizeCurrency(currency),
          }
          delete draftsByKey[blankKey]
        } else {
          draftsByKey[code] = createEmptyDraft(code, productName, currency)
        }
      }
      if (previousKey && previousKey !== code && previousKey.endsWith(':blank')) {
        delete draftsByKey[previousKey]
      }
      return code
    }

    const blankKey = `${identity}:blank`
    if (!draftsByKey[blankKey]) draftsByKey[blankKey] = createEmptyDraft('', productName, currency)
    return blankKey
  }

  const syncDrafts = (
    variants: ProductVariantForm[],
    productName: string,
    currency: string,
  ): string[] => variants.map((variant, index) => {
    const key = storageKeyForVariant(variant, index, productName, currency)
    const draft = draftsByKey[key]
    draft.productCode = normalizeCode(variant.sku)
    draft.productName = productName
    if (!draft.currency) draft.currency = normalizeCurrency(currency)
    if (draft.minimumOrderQuantity < 1) draft.minimumOrderQuantity = 1
    return key
  })

  const rowsForVariants = (
    variants: ProductVariantForm[],
    productName: string,
    currency: string,
  ): ProcurementProfitDraft[] => syncDrafts(variants, productName, currency)
    .map((key) => draftsByKey[key])

  const mergeLoadedRecords = (
    procurementRecords: Awaited<ReturnType<typeof procurementProfitabilityApi.listProcurementByCodes>>,
    profitabilityRecords: Awaited<ReturnType<typeof procurementProfitabilityApi.listProfitabilityByCodes>>,
  ) => {
    const profitabilityByCode = new Map(profitabilityRecords.map((record) => [record.product_code, record]))
    procurementRecords.forEach((record) => {
      const draft = draftsByKey[record.product_code] || createEmptyDraft(
        record.product_code,
        record.product_name,
        record.currency,
      )
      Object.assign(draft, {
        productCode: record.product_code,
        productName: record.product_name,
        purchasePrice: Number(record.purchase_price),
        purchasePriceKnown: true,
        currency: normalizeCurrency(record.currency),
        supplierName: record.supplier_name || '',
        supplierContactName: record.supplier_contact_name || '',
        supplierPhone: record.supplier_phone || '',
        supplierEmail: record.supplier_email || '',
        leadTimeDays: Number(record.lead_time_days || 0),
        minimumOrderQuantity: Math.max(1, Number(record.minimum_order_quantity || 1)),
        inboundShippingUnitCost: Number(record.inbound_shipping_unit_cost || 0),
        packagingUnitCost: Number(record.packaging_unit_cost || 0),
        otherUnitCost: Number(record.other_unit_cost || 0),
      })
      draftsByKey[record.product_code] = draft
    })
    profitabilityRecords.forEach((record) => {
      const draft = draftsByKey[record.product_code] || createEmptyDraft(
        record.product_code,
        record.product_name,
        record.currency,
      )
      const profit = profitabilityByCode.get(record.product_code)
      if (!draft.purchasePriceKnown && profit) {
        draft.purchasePrice = Number(profit.purchase_price)
        draft.purchasePriceKnown = true
      }
      draft.productCode = record.product_code
      draft.productName = record.product_name
      draft.currency = normalizeCurrency(record.currency)
      draft.inboundShippingUnitCost = Number(record.inbound_shipping_unit_cost || 0)
      draft.packagingUnitCost = Number(record.packaging_unit_cost || 0)
      draft.otherUnitCost = Number(record.other_unit_cost || 0)
      draftsByKey[record.product_code] = draft
    })
  }

  const loadForProduct = async (product: ProductSnapshot | null): Promise<void> => {
    reset()
    const variants = Array.isArray(product?.variants) ? product.variants : []
    const productName = String(product?.name || '')
    const currency = normalizeCurrency(product?.currency)
    syncDrafts(variants, productName, currency)
    const codes = variants.map((variant) => normalizeCode(variant.sku)).filter(Boolean)
    if (!codes.length) return

    loading.value = true
    try {
      const [procurementRecords, profitabilityRecords] = await Promise.all([
        procurementProfitabilityApi.listProcurementByCodes(codes),
        procurementProfitabilityApi.listProfitabilityByCodes(codes),
      ])
      mergeLoadedRecords(procurementRecords, profitabilityRecords)
    } catch (error) {
      loadError.value = errorMessage(error, '成本与利润资料加载失败')
    } finally {
      loading.value = false
    }
  }

  const hasProcurementInput = (draft: ProcurementProfitDraft): boolean => draft.purchasePriceKnown

  const toProcurementPayload = (draft: ProcurementProfitDraft): ProfitabilityProcurementPayload => ({
    supplier_name: draft.supplierName.trim(),
    supplier_contact_name: draft.supplierContactName.trim(),
    supplier_phone: draft.supplierPhone.trim(),
    supplier_email: draft.supplierEmail.trim(),
    lead_time_days: Math.max(0, Number(draft.leadTimeDays || 0)),
    minimum_order_quantity: Math.max(1, Number(draft.minimumOrderQuantity || 1)),
  })

  const buildItems = (
    variants: ProductVariantForm[],
    productName: string,
    currency: string,
  ): ProfitabilityItemPayload[] => {
    const keys = syncDrafts(variants, productName, currency)
    return variants
      .map((variant, index) => {
        const productCode = normalizeCode(variant.sku)
        if (!productCode) return null
        const draft = draftsByKey[keys[index]]
        const item: ProfitabilityItemPayload = {
          product_code: productCode,
          product_name: productName.trim(),
          currency: normalizeCurrency(currency),
          cost_currency: String(draft.currency || '').trim().toUpperCase() || normalizeCurrency(currency),
          list_price: Number(variant.price || 0),
          sale_price: variant.sale_price == null || variant.sale_price === ''
            ? null
            : Number(variant.sale_price),
          purchase_price: draft.purchasePriceKnown && draft.purchasePrice != null
            ? Number(draft.purchasePrice)
            : null,
          purchase_price_known: draft.purchasePriceKnown && draft.purchasePrice != null,
          inbound_shipping_unit_cost: Number(draft.inboundShippingUnitCost || 0),
          packaging_unit_cost: Number(draft.packagingUnitCost || 0),
          other_unit_cost: Number(draft.otherUnitCost || 0),
        }
        if (hasProcurementInput(draft)) item.procurement = toProcurementPayload(draft)
        return item
      })
      .filter((item): item is ProfitabilityItemPayload => item !== null)
  }

  const saveItems = async (
    items: ProfitabilityItemPayload[],
    requestId = '',
  ): Promise<DraftSaveResult> => {
    if (!items.length) {
      pending.value = false
      pendingItems.value = null
      return { success: true, result: { records: [], skipped: [] } }
    }
    saving.value = true
    saveError.value = ''
    try {
      const result = await procurementProfitabilityApi.bulkUpsert(items, requestId)
      pending.value = false
      pendingItems.value = null
      lastSavedAt.value = new Date().toISOString()
      return { success: true, result }
    } catch (error) {
      pending.value = true
      pendingItems.value = items.map((item) => ({ ...item, procurement: item.procurement ? { ...item.procurement } : undefined }))
      saveError.value = errorMessage(error, '成本与利润资料保存失败')
      return { success: false, error }
    } finally {
      saving.value = false
    }
  }

  const saveForProduct = async (product: ProductSnapshot): Promise<DraftSaveResult> => {
    if (loading.value || loadError.value) {
      return {
        success: false,
        error: new Error(loadError.value || '成本与利润资料尚未加载完成'),
      }
    }
    const variants = Array.isArray(product.variants) ? product.variants : []
    const productName = String(product.name || '')
    const currency = normalizeCurrency(product.currency)
    return saveItems(buildItems(variants, productName, currency), `product-${Date.now()}`)
  }

  const retryPending = async (): Promise<DraftSaveResult> => {
    if (!pendingItems.value) return { success: true, result: { records: [], skipped: [] } }
    return saveItems(pendingItems.value, `retry-${Date.now()}`)
  }

  const showLoadError = () => {
    if (loadError.value) toast.warning(loadError.value)
  }

  return {
    draftsByKey,
    loading,
    saving,
    pending,
    loadError,
    saveError,
    lastSavedAt,
    reset,
    rowsForVariants,
    loadForProduct,
    saveForProduct,
    retryPending,
    showLoadError,
  }
}

export default useProcurementProfitDraft


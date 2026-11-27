import { validateProductDeclarationIdentifiers } from '@/modules/product/productRequiredMerchantManufacturerAndOptionalStandardIdentifierValidation'
import { computed, reactive, ref } from 'vue'
import { createEmptyCarbonRimSpecificationForm, createEmptyCarbonRimVariantSpecificationForm, normalizeCarbonRimSpecificationForSubmission, normalizeCarbonRimVariantSpecificationForSubmission, validateCarbonRimDedicatedProductSubmission } from '@/modules/product/carbonRimDedicatedSpecificationFormValuesAndValidation'
import { createBicycleTireProductFormValuesFromSchwalbeCatalogItem, createEmptyBicycleTireSpecificationForm, normalizeBicycleTireSpecificationForSubmission, validateBicycleTireDedicatedProductSubmission } from '@/modules/product/bicycleTireDedicatedSpecificationFormValuesAndValidation'
import { createEmptyWheelsetSpecificationForm, createWheelsetSpecificationFormFromProduct, normalizeWheelsetSpecificationForSubmission, validateWheelsetDedicatedProductSubmission } from '@/modules/product/wheelsetDedicatedProductSpecificationFormValuesAndValidation'
import { createEmptySpokesProductSpecificationForm, createSpokesProductSpecificationFormFromProduct, normalizeSpokesProductSpecificationForSubmission, validateSpokesDedicatedProductSubmission } from '@/modules/product/spokesDedicatedProductSpecificationFormValuesAndValidation'
import { toast } from 'vue-sonner'
import productApi from '@/api/products'
import productCategoryApi from '@/api/productCategories'
import { requireCarbonRimSystemCategory, buildCarbonRimRequiredSystemCategoryDisplayPath } from '@/modules/product/carbonRimRequiredSystemCategorySelection'
import { requireBicycleTireSystemCategory, buildBicycleTireSystemCategoryDisplayPath } from '@/modules/product/bicycleTireRequiredSystemCategorySelection'
import { requireWheelsetSystemCategory, buildWheelsetSystemCategoryDisplayPath } from '@/modules/product/wheelsetRequiredSystemCategorySelection'
import { requireSpokesSystemCategory, buildSpokesSystemCategoryDisplayPath } from '@/modules/product/spokesRequiredSystemCategorySelection'
import schwalbeTireCatalogApi, { type SchwalbeTireCatalogItem } from '@/api/schwalbeTireCatalog'
import { useProductMediaManager } from '@/composables/product/useProductMediaManager'
import { buildProductMediaFormValues } from '@/lib/productMedia'
import axios from '@/utils/axios'
import type { ProductFormRecord } from '@/modules/product/productEditorTypes'

type ProductEditorMode = 'create' | 'edit'

const createEmptyGenericProductSchema = () => ({
  schema_version: 1 as const,
  model: '',
  material: '',
  color: '',
  size: '',
  weight_g: null as number | null,
})

const normalizeGenericProductSchemaForSubmission = (value: ProductFormRecord['generic_product_schema']) => {
  if (!value) return undefined
  const model = String(value.model || '').trim()
  const material = String(value.material || '').trim()
  const color = String(value.color || '').trim()
  const size = String(value.size || '').trim()
  const weight = value.weight_g == null ? null : Number(value.weight_g)
  if (!model && !material && !color && !size && (!weight || !Number.isFinite(weight) || weight <= 0)) return null
  return {
    schema_version: 1 as const,
    model,
    material,
    color,
    size,
    weight_g: weight && Number.isFinite(weight) && weight > 0 ? weight : null,
  }
}

interface ProductEditorCallbackResult {
  keepDialogOpen?: boolean
}

export const useProductEditor = (options: Record<string, any> = {}) => {
  const refreshProducts = options.refreshProducts || (() => Promise.resolve())
  const resolveDefaultLocale = () => options.defaultLocale?.value || options.defaultLocale || ''
  const afterProductLoaded = options.afterProductLoaded as (
    product: any | null,
    mode: ProductEditorMode,
  ) => Promise<void> | void
  const afterProductSaved = options.afterProductSaved as (
    product: any,
    mode: ProductEditorMode,
  ) => Promise<ProductEditorCallbackResult | void> | ProductEditorCallbackResult | void
  const defaultPrimaryCurrency = 'USD'
  const minorUnitsForCurrency = (value: any): number => {
    const code = normalizeCurrencyCode(value)
    if (['BHD', 'IQD', 'JOD', 'KWD', 'LYD', 'OMR', 'TND'].includes(code)) return 3
    if (['BIF', 'CLP', 'DJF', 'GNF', 'JPY', 'KMF', 'KRW', 'MGA', 'PYG', 'RWF', 'UGX', 'VND', 'VUV', 'XAF', 'XOF', 'XPF'].includes(code)) return 0
    return 2
  }
  const majorFromMinor = (value: any, currency: any): string => {
    const minor = Number(value)
    if (!Number.isFinite(minor)) return ''
    const units = minorUnitsForCurrency(currency)
    return (minor / (10 ** units)).toFixed(units)
  }
  const minorFromMajor = (value: any, currency: any): number => {
    const major = Number(value)
    return Number.isFinite(major) ? Math.round(major * (10 ** minorUnitsForCurrency(currency))) : 0
  }
  const normalizeCurrencyCode = (value: any) => String(value || '').trim().toUpperCase()
  const validCurrencyCodeOrDefault = (value: any) => {
    const code = normalizeCurrencyCode(value)
    return /^[A-Z]{3}$/.test(code) ? code : defaultPrimaryCurrency
  }

  const carbonRimSystemCategoryDisplayPath = ref('')
  const bicycleTireSystemCategoryDisplayPath = ref('')
  const wheelsetSystemCategoryDisplayPath = ref('')
  const spokesSystemCategoryDisplayPath = ref('')
  const loadRequiredCarbonRimSystemCategory = async () => {
    const categories = (await productCategoryApi.list({ include_disabled: true })).flat
    const category = requireCarbonRimSystemCategory(categories)
    carbonRimSystemCategoryDisplayPath.value = buildCarbonRimRequiredSystemCategoryDisplayPath(categories)
    return category
  }
  const loadRequiredBicycleTireSystemCategory = async () => {
    const categories = (await productCategoryApi.list({ include_disabled: true })).flat
    const category = requireBicycleTireSystemCategory(categories)
    bicycleTireSystemCategoryDisplayPath.value = buildBicycleTireSystemCategoryDisplayPath(categories)
    return category
  }
  const loadRequiredWheelsetSystemCategory = async () => {
    const categories = (await productCategoryApi.list({ include_disabled: true })).flat
    const category = requireWheelsetSystemCategory(categories)
    wheelsetSystemCategoryDisplayPath.value = buildWheelsetSystemCategoryDisplayPath(categories)
    return category
  }
  const loadRequiredSpokesSystemCategory = async () => {
    const categories = (await productCategoryApi.list({ include_disabled: true })).flat
    const category = requireSpokesSystemCategory(categories)
    spokesSystemCategoryDisplayPath.value = buildSpokesSystemCategoryDisplayPath(categories)
    return category
  }
  const schwalbeTireCatalogItems = ref<SchwalbeTireCatalogItem[]>([])
  const schwalbeTireCatalogLoading = ref(false)
  const schwalbeTireCatalogError = ref('')
  const primaryCurrency = ref(defaultPrimaryCurrency)
  const currencyPolicyLoaded = ref(false)
  const dialogVisible = ref(false)
  const dialogMode = ref<'create' | 'edit'>('create')
  const submitting = ref(false)
  const formErrors = reactive<Record<string, string>>({})

  const productForm = reactive<ProductFormRecord>({
    id: null,
    product_type: '',
    carbon_rim_specification: null,
    bicycle_tire_specification: null,
    wheelset_specification: null,
    spoke_specification: null,
    generic_product_schema: {
      schema_version: 1,
      model: '',
      material: '',
      color: '',
      size: '',
      weight_g: null,
    },
    parent_id: null,
    product_category_id: null,
    brand_id: null,
    after_sales_template_id: null,
    packaging_template_id: null,
    customs_classification_profile_id: null,
    hs_code: '',
    cn_code: '',
    country_of_origin: '',
    customs_description: '',
    merchant_product_identifier: '',
    manufacturer_product_identifier: '',
    standard_product_identifier: '',
    name: '',
    slug: '',
    description: '',
    short_description: '',
    currency: primaryCurrency.value,
    status: 'active',
    locale: resolveDefaultLocale(),
    featured: false,
    variants: [],
    media: []
  })

  const clearFormErrors = () => Object.keys(formErrors).forEach((key) => delete formErrors[key])
  const clearFieldError = (field: string) => { delete formErrors[field] }

  const {
    uploadingMedia,
    mediaTypeLabel,
    mediaRoleOptions,
    addMediaUrl,
    handleMediaUpload,
    setPrimaryMedia,
    moveMedia,
    removeMedia,
    normalizeFormMedia
  } = useProductMediaManager(productForm, { clearFieldError })

  const selectedSchwalbeTireArticleNo = computed(() => String(
    productForm.product_type === 'bicycle_tire'
      ? productForm.bicycle_tire_specification?.article_no || ''
      : '',
  ))
  const defaultVariantIndex = computed(() => {
    const index = productForm.variants.findIndex((variant: any) => variant.is_default)
    return index >= 0 ? index : 0
  })
  const productCategorySelectValue = computed(() => productForm.product_category_id == null ? '__none__' : String(productForm.product_category_id))
  const brandSelectValue = computed(() => productForm.brand_id == null ? '__none__' : String(productForm.brand_id))
  const afterSalesTemplateSelectValue = computed(() => productForm.after_sales_template_id == null ? '__none__' : String(productForm.after_sales_template_id))
  const packagingTemplateSelectValue = computed(() => productForm.packaging_template_id == null ? '__none__' : String(productForm.packaging_template_id))
  const setProductBrand = (value: string) => {
    productForm.brand_id = value === '__none__' ? null : Number(value)
    clearFieldError('brand_id')
  }
  const setProductCategory = (value: string) => {
    if (productForm.product_type === 'carbon_rim' || productForm.product_type === 'wheelset') return
    productForm.product_category_id = value === '__none__' ? null : Number(value)
    clearFieldError('product_category_id')
  }
  const setProductInformationTemplate = (field: string, value: string) => {
    productForm[field] = value === '__none__' ? null : Number(value)
    clearFieldError(field)
  }
  const applyCustomsClassification = (profile: Record<string, any>) => {
    productForm.customs_classification_profile_id = profile?.id ? Number(profile.id) : null
    productForm.hs_code = String(profile?.hs_code || '')
    productForm.cn_code = String(profile?.cn_code || '')
    productForm.country_of_origin = String(profile?.country_of_origin || '').toUpperCase()
    productForm.customs_description = String(profile?.customs_description || '')
    ;['hs_code', 'cn_code', 'country_of_origin', 'customs_description'].forEach(clearFieldError)
  }
  const clearCustomsClassification = () => {
    productForm.customs_classification_profile_id = null
  }
  const primaryPriceCurrency = () => validCurrencyCodeOrDefault(primaryCurrency.value)

  const fetchPrimaryPricingCurrency = async (force = false) => {
    if (!force && currencyPolicyLoaded.value) return primaryPriceCurrency()
    try {
      const response = await axios.get('/api/admin/settings/currency-policy')
      primaryCurrency.value = validCurrencyCodeOrDefault(response.data?.policy?.primary_currency)
      currencyPolicyLoaded.value = true
    } catch (error) {
      console.error('Failed to fetch primary pricing currency:', error)
      primaryCurrency.value = defaultPrimaryCurrency
    }
    return primaryPriceCurrency()
  }

  const createEmptyVariant = (overrides: Record<string, any> = {}) => ({
    shipping_template_ids: [],
    carbon_rim_specification: productForm.product_type === 'carbon_rim' ? createEmptyCarbonRimVariantSpecificationForm() : undefined,
    id: null,
    sku: '',
    title: '',
    currency: primaryPriceCurrency(),
    price: '0.00',
    stock: 0,
    weight_grams: 0,
    is_default: false,
    is_active: true,
    sort_order: productForm.variants.length * 10,
    ...overrides
  })

  const buildVariantFormValues = (product: any) => {
    const variants = (product.variants || []).map((variant: any, index: number) => createEmptyVariant({
      carbon_rim_specification: variant.carbon_rim_specification ? { ...createEmptyCarbonRimVariantSpecificationForm(), ...variant.carbon_rim_specification } : undefined,
      id: variant.id || null,
      shipping_template_ids: Array.isArray(variant.shipping_template_ids) ? [...variant.shipping_template_ids] : [],
      sku: variant.sku || '',
      title: variant.title || '',
      currency: validCurrencyCodeOrDefault(variant.currency || product.currency),
      price_minor: Number.isFinite(Number(variant.price_minor)) ? Number(variant.price_minor) : 0,
      price: majorFromMinor(Number.isFinite(Number(variant.price_minor)) ? Number(variant.price_minor) : 0, variant.currency || product.currency),
      stock: Number(variant.stock || 0),
      weight_grams: variant.weight_grams ?? variant.weight ?? 0,
      is_default: Boolean(variant.is_default),
      is_active: variant.is_active !== false,
      sort_order: variant.sort_order ?? index * 10
    }))
    if (variants.length === 0) variants.push(createEmptyVariant({ is_default: true }))
    if (!variants.some((variant: any) => variant.is_default)) variants[0].is_default = true
    return variants
  }

  const addVariant = () => {
    if (productForm.product_type === 'carbon_rim' && productForm.parent_id) { toast.warning('请在原语言商品中维护销售规格'); return }
    productForm.variants.push(createEmptyVariant({ is_default: productForm.variants.length === 0 }))
    clearFieldError('variants')
  }

  const removeVariant = (index: number) => {
    if (productForm.product_type === 'carbon_rim' && productForm.variants[index]?.id) { toast.warning('已保存的车圈 SKU 请取消启用，保留销售记录'); return }
    if (productForm.variants.length <= 1) {
      toast.warning('至少保留一个变体')
      return
    }
    const wasDefault = productForm.variants[index]?.is_default
    productForm.variants.splice(index, 1)
    if (wasDefault) setDefaultVariant(0)
  }

  const setDefaultVariant = (index: number) => {
    productForm.variants.forEach((variant: any, currentIndex: number) => { variant.is_default = currentIndex === index })
  }

  const ensureDefaultVariantIsEnabled = () => {
    if (!productForm.variants.length) return
    const activeDefaultIndex = productForm.variants.findIndex((variant: any) => variant.is_default && variant.is_active !== false)
    if (activeDefaultIndex >= 0) return
    const firstActiveIndex = productForm.variants.findIndex((variant: any) => variant.is_active !== false)
    setDefaultVariant(firstActiveIndex >= 0 ? firstActiveIndex : 0)
  }

  const setVariantActive = (index: number, isActive: boolean) => {
    const variant = productForm.variants[index]
    if (!variant) return
    variant.is_active = Boolean(isActive)
    ensureDefaultVariantIsEnabled()
    clearFieldError('variants')
  }

  const normalizeFormVariants = () => {
    if (!productForm.variants.length) return []
    ensureDefaultVariantIsEnabled()
    if (!productForm.variants.some((variant: any) => variant.is_default)) productForm.variants[0].is_default = true
    return productForm.variants.map((variant: any, index: number) => ({
        carbon_rim_specification: productForm.product_type === 'carbon_rim' ? normalizeCarbonRimVariantSpecificationForSubmission(variant.carbon_rim_specification) : undefined,
        id: variant.id || undefined,
        shipping_template_ids: Array.from(new Set((Array.isArray(variant.shipping_template_ids) ? variant.shipping_template_ids : []).map((value: any) => Number(value)).filter((value: number) => Number.isInteger(value) && value > 0))),
        sku: String(variant.sku || '').trim(),
        title: String(variant.title || '').trim(),
        currency: validCurrencyCodeOrDefault(variant.currency || productForm.currency),
        price: String(variant.price ?? ''),
        price_minor: minorFromMajor(variant.price, variant.currency || productForm.currency),
        stock: Number(variant.stock || 0),
        weight_grams: Number(variant.weight_grams || 0),
        is_default: Boolean(variant.is_default),
        is_active: variant.is_active !== false,
        sort_order: Number(variant.sort_order ?? index * 10)
      }))
  }

  const buildProductPayload = () => ({
    product_type: productForm.product_type || undefined,
    carbon_rim_specification: productForm.product_type === 'carbon_rim' ? normalizeCarbonRimSpecificationForSubmission(productForm.carbon_rim_specification) : undefined,
    bicycle_tire_specification: productForm.product_type === 'bicycle_tire' ? normalizeBicycleTireSpecificationForSubmission(productForm.bicycle_tire_specification) : undefined,
    wheelset_specification: productForm.product_type === 'wheelset' ? normalizeWheelsetSpecificationForSubmission(productForm.wheelset_specification) : undefined,
    spoke_specification: productForm.product_type === 'spoke' ? normalizeSpokesProductSpecificationForSubmission(productForm.spoke_specification) : undefined,
    generic_product_schema: productForm.product_type === '' ? normalizeGenericProductSchemaForSubmission(productForm.generic_product_schema) : undefined,
    id: productForm.id,
    product_category_id: productForm.product_category_id,
    brand_id: productForm.brand_id,
    after_sales_template_id: productForm.after_sales_template_id,
    packaging_template_id: productForm.packaging_template_id,
    customs_classification_profile_id: productForm.customs_classification_profile_id || null,
    hs_code: String(productForm.hs_code || '').trim(),
    cn_code: String(productForm.cn_code || '').trim(),
    country_of_origin: String(productForm.country_of_origin || '').trim().toUpperCase(),
    customs_description: String(productForm.customs_description || '').trim(),
    merchant_product_identifier: String(productForm.merchant_product_identifier || '').trim(),
    manufacturer_product_identifier: String(productForm.manufacturer_product_identifier || '').trim(),
    standard_product_identifier: String(productForm.standard_product_identifier || '').trim(),
    name: productForm.name.trim(),
    slug: productForm.slug.trim(),
    description: productForm.description,
    short_description: productForm.short_description,
    currency: validCurrencyCodeOrDefault(productForm.currency),
    status: productForm.status,
    locale: productForm.locale,
    featured: productForm.featured,
    variants: normalizeFormVariants(),
    media: normalizeFormMedia()
  })

  const validateForm = (payload: any) => {
    clearFormErrors()
    Object.assign(formErrors, validateProductDeclarationIdentifiers(payload))
    if (payload.product_type === 'carbon_rim') Object.assign(formErrors, validateCarbonRimDedicatedProductSubmission(payload))
    if (payload.product_type === 'bicycle_tire') Object.assign(formErrors, validateBicycleTireDedicatedProductSubmission(payload))
    if (payload.product_type === 'wheelset') Object.assign(formErrors, validateWheelsetDedicatedProductSubmission(payload))
    if (payload.product_type === 'spoke') Object.assign(formErrors, validateSpokesDedicatedProductSubmission(payload))
    if (!payload.name) formErrors.name = '请输入商品名称'
    if (!payload.slug) formErrors.slug = '请输入 URL slug'
    if (!payload.locale) formErrors.locale = '请选择语言'
    if (!/^[A-Z]{3}$/.test(payload.currency)) formErrors.currency = '请选择商品主基准币种'
    if (payload.hs_code && !/^\d{6}$/.test(payload.hs_code)) formErrors.hs_code = 'HS Code 必须是 6 位数字'
    if (payload.cn_code && !/^\d{8}$/.test(payload.cn_code)) formErrors.cn_code = 'CN Code 必须是 8 位数字'
    if (payload.country_of_origin && !/^[A-Z]{2}$/.test(payload.country_of_origin)) formErrors.country_of_origin = '请输入 2 位国家代码，例如 CN'
    if (payload.customs_description.length > 255) formErrors.customs_description = '英文报关品名不能超过 255 个字符'
    if (!payload.variants.length && !(payload.product_type === 'spoke' && payload.status === 'inactive')) formErrors.variants = '请至少添加一个 SKU 变体'
    else if (payload.variants.some((variant: any) => !variant.sku)) formErrors.variants = '每个变体都必须填写 SKU'
    else if (new Set(payload.variants.map((variant: any) => variant.sku.toLowerCase())).size !== payload.variants.length) formErrors.variants = '变体 SKU 不能重复'
    else if (payload.variants.some((variant: any) => Number(variant.price) <= 0)) formErrors.variants = '每个变体价格必须大于 0'
    else if (payload.variants.some((variant: any) => Number(variant.stock) < 0)) formErrors.variants = '变体库存不能为负数'
    else if (payload.variants.length > 0 && !payload.variants.some((variant: any) => variant.is_active !== false)) formErrors.variants = '请至少启用一个 SKU 变体'
    if (productForm.media.some((item: any) => !String(item.url || '').trim())) formErrors.media = '媒体条目必须填写 URL，空条目请删除'
    else if (payload.media.filter((item: any) => item.media_type === 'image' && item.is_primary).length > 1) formErrors.media = '商品主图只能设置一张'
    if (Object.keys(formErrors).length > 0) {
      toast.error('请检查商品表单中的必填项')
      return false
    }
    return true
  }

  const fetchSchwalbeTireCatalog = async () => {
    schwalbeTireCatalogLoading.value = true
    schwalbeTireCatalogError.value = ''
    try {
      schwalbeTireCatalogItems.value = await schwalbeTireCatalogApi.list()
    } catch (error) {
      console.error('Failed to load Schwalbe tire catalog:', error)
      schwalbeTireCatalogError.value = '读取 Schwalbe 型号目录失败，请重试。'
    } finally {
      schwalbeTireCatalogLoading.value = false
    }
  }

  const setSchwalbeTireCatalogModel = (articleNo: string) => {
    if (!articleNo || articleNo === '__none__') return
    const item = schwalbeTireCatalogItems.value.find((candidate) => candidate.article_no === articleNo)
    if (!item) return
    if (productForm.product_type !== 'bicycle_tire') return
    const values = createBicycleTireProductFormValuesFromSchwalbeCatalogItem(item)
    productForm.bicycle_tire_specification = values.bicycle_tire_specification
    productForm.standard_product_identifier = values.standard_product_identifier
    clearFieldError('bicycle_tire_specification')
    clearFieldError('standard_product_identifier')
  }

  const resetForm = () => {
    Object.assign(productForm, {
      id: null,
      product_type: '',
      carbon_rim_specification: null,
      bicycle_tire_specification: null,
      wheelset_specification: null,
      spoke_specification: null,
      generic_product_schema: createEmptyGenericProductSchema(),
      parent_id: null,
      product_category_id: null,
      brand_id: null,
      after_sales_template_id: null,
      packaging_template_id: null,
      customs_classification_profile_id: null,
      hs_code: '',
      cn_code: '',
      country_of_origin: '',
      customs_description: '',
      merchant_product_identifier: '',
      manufacturer_product_identifier: '',
      standard_product_identifier: '',
      name: '',
      slug: '',
      description: '',
      short_description: '',
      currency: primaryPriceCurrency(),
      status: 'active',
      locale: resolveDefaultLocale(),
      featured: false,
      variants: [],
      media: []
    })
    productForm.variants = [createEmptyVariant({ is_default: true })]
    clearFormErrors()
  }

  const notifyProductLoaded = async (product: any | null, mode: ProductEditorMode) => {
    if (!afterProductLoaded) return
    try {
      await afterProductLoaded(product, mode)
    } catch (error) {
      console.error('Failed to load independent product add-on data:', error)
    }
  }

  const notifyProductSaved = async (product: any, mode: ProductEditorMode): Promise<ProductEditorCallbackResult> => {
    if (!afterProductSaved) return {}
    try {
      return (await afterProductSaved(product, mode)) || {}
    } catch (error) {
      // Product persistence has already succeeded. Keep the dialog open so an
      // independent add-on failure can be retried without submitting again.
      console.error('Failed to save independent product add-on data:', error)
      return { keepDialogOpen: true }
    }
  }

  const showCreateDialog = async (productType = '') => {
    let carbonRimCategoryID: number | null = null
    let bicycleTireCategoryID: number | null = null
    let wheelsetCategoryID: number | null = null
    let spokesCategoryID: number | null = null
    if (productType === 'carbon_rim') {
      try {
        carbonRimCategoryID = (await loadRequiredCarbonRimSystemCategory()).id
      } catch (error) {
        toast.error(error instanceof Error ? error.message : '读取碳圈系统分类失败，请重试')
        return
      }
    }
    if (productType === 'bicycle_tire') {
      try {
        bicycleTireCategoryID = (await loadRequiredBicycleTireSystemCategory()).id
      } catch (error) {
        toast.error(error instanceof Error ? error.message : '读取外胎系统分类失败，请重试')
        return
      }
    }
    if (productType === 'wheelset') {
      try {
        wheelsetCategoryID = (await loadRequiredWheelsetSystemCategory()).id
      } catch (error) {
        toast.error(error instanceof Error ? error.message : '读取 WHEELSET 系统分类失败，请重试')
        return
      }
    }
    if (productType === 'spoke') {
      try {
        spokesCategoryID = (await loadRequiredSpokesSystemCategory()).id
      } catch (error) {
        toast.error(error instanceof Error ? error.message : '读取辐条系统分类失败，请重试')
        return
      }
    }
    await fetchPrimaryPricingCurrency()
    dialogMode.value = 'create'
    resetForm()
    if (productType === 'carbon_rim') {
      productForm.product_type = productType
      productForm.product_category_id = carbonRimCategoryID
      productForm.carbon_rim_specification = createEmptyCarbonRimSpecificationForm()
      productForm.variants = [createEmptyVariant({ is_default: true })]
    }
    if (productType === 'bicycle_tire') {
      productForm.product_type = productType
      productForm.product_category_id = bicycleTireCategoryID
      productForm.bicycle_tire_specification = createEmptyBicycleTireSpecificationForm()
      productForm.variants = [createEmptyVariant({ is_default: true })]
    }
    if (productType === 'wheelset') {
      productForm.product_type = productType
      productForm.product_category_id = wheelsetCategoryID
      productForm.wheelset_specification = createEmptyWheelsetSpecificationForm()
      productForm.variants = [createEmptyVariant({ is_default: true })]
    }
    if (productType === 'spoke') {
      productForm.product_type = productType
      productForm.product_category_id = spokesCategoryID
      productForm.spoke_specification = createEmptySpokesProductSpecificationForm()
      productForm.status = 'inactive'
      productForm.variants = []
    }
    await notifyProductLoaded(null, 'create')
    dialogVisible.value = true
    if (productType === 'bicycle_tire') void fetchSchwalbeTireCatalog()
  }

  const showEditDialog = async (product: any) => {
    dialogMode.value = 'edit'
    let detail = product
    try {
      await fetchPrimaryPricingCurrency()
      detail = await productApi.get(product.id)
      if (detail.product_type === 'carbon_rim') {
        const category = await loadRequiredCarbonRimSystemCategory()
        if (detail.product_category_id !== category.id) throw new Error('碳圈商品分类不正确，请先检查数据库迁移')
      }
      if (detail.product_type === 'bicycle_tire') {
        const category = await loadRequiredBicycleTireSystemCategory()
        if (detail.product_category_id !== category.id) throw new Error('外胎商品分类不正确，请先检查数据库迁移')
      }
      if (detail.product_type === 'wheelset') {
        const category = await loadRequiredWheelsetSystemCategory()
        if (detail.product_category_id !== category.id) throw new Error('WHEELSET 商品分类不正确，请先检查系统分类配置')
      }
      if (detail.product_type === 'spoke') {
        const category = await loadRequiredSpokesSystemCategory()
        if (detail.product_category_id !== category.id) throw new Error('辐条商品分类不正确，请先检查系统分类配置')
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : '获取商品详情失败，请重试后再编辑')
      return
    }
    Object.assign(productForm, {
      product_type: detail.product_type || '',
      carbon_rim_specification: detail.carbon_rim_specification ? { ...createEmptyCarbonRimSpecificationForm(), ...detail.carbon_rim_specification } : null,
      bicycle_tire_specification: detail.bicycle_tire_specification ? { ...createEmptyBicycleTireSpecificationForm(), ...detail.bicycle_tire_specification } : (detail.product_type === 'bicycle_tire' ? createEmptyBicycleTireSpecificationForm() : null),
      wheelset_specification: detail.product_type === 'wheelset' ? createWheelsetSpecificationFormFromProduct(detail.wheelset_specification) : null,
      spoke_specification: detail.product_type === 'spoke' ? createSpokesProductSpecificationFormFromProduct(detail.spoke_specification) : null,
      generic_product_schema: detail.product_type === '' ? { ...createEmptyGenericProductSchema(), ...(detail.generic_product_schema || {}) } : null,
      parent_id: detail.parent_id || null,
      id: detail.id,
      product_category_id: detail.product_category_id ?? detail.product_category?.id ?? null,
      brand_id: detail.brand_id ?? detail.brand?.id ?? null,
      after_sales_template_id: detail.after_sales_template_id ?? detail.after_sales_template?.id ?? null,
      packaging_template_id: detail.packaging_template_id ?? detail.packaging_template?.id ?? null,
      customs_classification_profile_id: detail.customs_classification_profile_id ?? detail.customs_classification_profile?.id ?? null,
      hs_code: String(detail.hs_code || ''),
      cn_code: String(detail.cn_code || ''),
      country_of_origin: String(detail.country_of_origin || ''),
      customs_description: String(detail.customs_description || ''),
      merchant_product_identifier: String(detail.merchant_product_identifier || ''),
      manufacturer_product_identifier: String(detail.manufacturer_product_identifier || ''),
      standard_product_identifier: String(detail.standard_product_identifier || ''),
      name: detail.name || '',
      slug: detail.slug || '',
      description: detail.description || '',
      short_description: detail.short_description || detail.short_desc || '',
      currency: validCurrencyCodeOrDefault(detail.currency),
      status: detail.status || 'active',
      locale: detail.locale || resolveDefaultLocale(),
      featured: Boolean(detail.featured),
      variants: buildVariantFormValues(detail),
      media: buildProductMediaFormValues(detail)
    })
	if (detail.product_type === 'bicycle_tire') {
      await fetchSchwalbeTireCatalog()
    }
    clearFormErrors()
    await notifyProductLoaded(detail, 'edit')
    dialogVisible.value = true
  }

  const submitForm = async () => {
    if (submitting.value || uploadingMedia.value) return
    submitting.value = true
    try {
      const payload = buildProductPayload()
      if (payload.product_type === 'carbon_rim') {
        const category = await loadRequiredCarbonRimSystemCategory()
        if (payload.product_category_id !== category.id) {
          formErrors.product_category_id = '碳圈商品必须属于 Wheel Components → Carbon Rim'
          toast.error(formErrors.product_category_id)
          return
        }
      }
      if (payload.product_type === 'bicycle_tire') {
        const category = await loadRequiredBicycleTireSystemCategory()
        if (payload.product_category_id !== category.id) {
          formErrors.product_category_id = '自行车外胎必须归入 Wheel Components → Tires'
          toast.error(formErrors.product_category_id)
          return
        }
      }
      if (payload.product_type === 'wheelset') {
        const category = await loadRequiredWheelsetSystemCategory()
        if (payload.product_category_id !== category.id) {
          formErrors.product_category_id = 'WHEELSET 商品必须归入 WHEELSET 系统分类'
          toast.error(formErrors.product_category_id)
          return
        }
      }
      if (payload.product_type === 'spoke') {
        const category = await loadRequiredSpokesSystemCategory()
        if (payload.product_category_id !== category.id) {
          formErrors.product_category_id = '辐条商品必须归入辐条系统分类'
          toast.error(formErrors.product_category_id)
          return
        }
      }
      if (!validateForm(payload)) return
      let savedProduct: any
      if (dialogMode.value === 'create') {
        savedProduct = await productApi.create(payload)
        toast.success('商品创建成功')
      } else {
        const { id, ...data } = payload
        savedProduct = await productApi.update(id, data)
        toast.success('商品更新成功')
      }
      const savedMode = dialogMode.value
      productForm.id = savedProduct.id
      dialogMode.value = 'edit'
      productForm.variants.forEach((variant: any) => {
        const saved = savedProduct.variants?.find((item: any) => item.sku === variant.sku)
        if (saved) variant.id = saved.id
      })
      const afterSaveResult = await notifyProductSaved(savedProduct, savedMode)
      if (!afterSaveResult.keepDialogOpen) dialogVisible.value = false
      await refreshProducts()
    } catch (error: any) {
      formErrors.submit = error?.response?.data?.error || error?.message || '商品保存失败，请检查资料后重试'
      toast.error(formErrors.submit)
      console.error('Failed to save product:', error)
    } finally {
      submitting.value = false
    }
  }

  const closeDialog = () => {
    dialogVisible.value = false
  }

  return {
    carbonRimSystemCategoryDisplayPath,
    bicycleTireSystemCategoryDisplayPath,
    wheelsetSystemCategoryDisplayPath,
    spokesSystemCategoryDisplayPath,
    schwalbeTireCatalogItems,
    schwalbeTireCatalogLoading,
    schwalbeTireCatalogError,
    selectedSchwalbeTireArticleNo,
    dialogVisible,
    dialogMode,
    submitting,
    formErrors,
    productForm,
    uploadingMedia,
    defaultVariantIndex,
    productCategorySelectValue,
    brandSelectValue,
    afterSalesTemplateSelectValue,
    packagingTemplateSelectValue,
    setProductBrand,
    setProductCategory,
    setProductInformationTemplate,
    applyCustomsClassification,
    clearCustomsClassification,
    clearFieldError,
    addMediaUrl,
    mediaTypeLabel,
    mediaRoleOptions,
    handleMediaUpload,
    setPrimaryMedia,
    moveMedia,
    removeMedia,
    addVariant,
    removeVariant,
    setDefaultVariant,
    setVariantActive,
    fetchSchwalbeTireCatalog,
    setSchwalbeTireCatalogModel,
    showCreateDialog,
    showEditDialog,
    submitForm,
    closeDialog
  }
}

export default useProductEditor


import { computed, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import productApi from '@/api/products'
import productSpecificationTemplateApi from '@/api/productSpecificationTemplates'
import { useProductMediaManager } from '@/composables/product/useProductMediaManager'
import { buildProductMediaFormValues } from '@/lib/productMedia'
import axios from '@/utils/axios'
import type {
  ProductFormRecord,
  ProductVariantOptionValueForm
} from '@/components/admin/product/productEditorTypes'

export const useProductEditor = (options: Record<string, any> = {}) => {
  const refreshProducts = options.refreshProducts || (() => Promise.resolve())
  const resolveDefaultLocale = () => options.defaultLocale?.value || options.defaultLocale || ''
  const defaultPrimaryCurrency = 'USD'
  const normalizeCurrencyCode = (value: any) => String(value || '').trim().toUpperCase()
  const validCurrencyCodeOrDefault = (value: any) => {
    const code = normalizeCurrencyCode(value)
    return /^[A-Z]{3}$/.test(code) ? code : defaultPrimaryCurrency
  }

  const productSpecTemplates = ref<any[]>([])
  const primaryCurrency = ref(defaultPrimaryCurrency)
  const currencyPolicyLoaded = ref(false)
  const dialogVisible = ref(false)
  const dialogMode = ref<'create' | 'edit'>('create')
  const submitting = ref(false)
  const formErrors = reactive<Record<string, string>>({})

  const productForm = reactive<ProductFormRecord>({
    id: null,
    product_specification_template_id: null,
    product_category_id: null,
    brand_id: null,
    shipping_template_id: null,
    after_sales_template_id: null,
    packaging_template_id: null,
    customs_classification_profile_id: null,
    hs_code: '',
    cn_code: '',
    country_of_origin: '',
    customs_description: '',
    name: '',
    slug: '',
    description: '',
    short_description: '',
    currency: primaryCurrency.value,
    status: 'active',
    locale: resolveDefaultLocale(),
    featured: false,
    specs: {},
    variants: [],
    variant_option_values: [],
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

  const selectedProductSpecTemplate = computed(() => productSpecTemplates.value.find((template) => template.id === productForm.product_specification_template_id) || null)
  const selectedSpecDefinitions = computed(() => (selectedProductSpecTemplate.value?.spec_definitions || []).filter((spec: any) => !spec.is_variant_option))
  const variantSpecDefinitions = computed(() => (selectedProductSpecTemplate.value?.spec_definitions || []).filter((spec: any) => spec.is_variant_option))
  const defaultVariantIndex = computed(() => {
    const index = productForm.variants.findIndex((variant: any) => variant.is_default)
    return index >= 0 ? index : 0
  })
  const productSpecTemplateSelectValue = computed(() => productForm.product_specification_template_id == null ? '__none__' : String(productForm.product_specification_template_id))
  const productCategorySelectValue = computed(() => productForm.product_category_id == null ? '__none__' : String(productForm.product_category_id))
  const brandSelectValue = computed(() => productForm.brand_id == null ? '__none__' : String(productForm.brand_id))
  const shippingTemplateSelectValue = computed(() => productForm.shipping_template_id == null ? '__none__' : String(productForm.shipping_template_id))
  const afterSalesTemplateSelectValue = computed(() => productForm.after_sales_template_id == null ? '__none__' : String(productForm.after_sales_template_id))
  const packagingTemplateSelectValue = computed(() => productForm.packaging_template_id == null ? '__none__' : String(productForm.packaging_template_id))
  const hasMeaningfulTemplateValue = (value: any) => {
    if (value === undefined || value === null || value === '') return false
    if (value === false) return false
    if (Array.isArray(value)) return value.length > 0
    if (typeof value === 'object') return Object.keys(value).length > 0
    return true
  }
  const templateScopedValuesTouched = computed(() => (
    Object.values(productForm.specs || {}).some(hasMeaningfulTemplateValue) ||
    productForm.variants.some((variant: any) => Object.values(variant.option_values || {}).some(hasMeaningfulTemplateValue)) ||
    productForm.variant_option_values.some((item: ProductVariantOptionValueForm) => (
      hasMeaningfulTemplateValue(item.value_key)
      || hasMeaningfulTemplateValue(item.label)
      || hasMeaningfulTemplateValue(item.color_hex)
      || hasMeaningfulTemplateValue(item.swatch_url)
    ))
  ))

  const parseSpecOptions = (spec: any) => {
    if (!spec?.options) return []
    try {
      const parsed = JSON.parse(spec.options)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  const formatSpecOption = (option: unknown) => String(option).replace(/_/g, ' ')
  const getSpecLabel = (spec: any) => spec.unit ? `${spec.name} (${spec.unit})` : spec.name
  const specSelectValue = (value: any) => value === undefined || value === null || value === '' ? '__empty__' : String(value)
  const setSpecSelectValue = (slug: string, value: string) => {
    productForm.specs[slug] = value === '__empty__' ? '' : value
    clearFieldError(`spec:${slug}`)
  }
  const setProductShippingTemplate = (value: string) => {
    productForm.shipping_template_id = value === '__none__' ? null : Number(value)
    clearFieldError('shipping_template_id')
  }
  const setProductBrand = (value: string) => {
    productForm.brand_id = value === '__none__' ? null : Number(value)
    clearFieldError('brand_id')
  }
  const setProductCategory = (value: string) => {
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

  const coerceSpecValueForForm = (definition: any, value: any) => {
    if (!definition) return value
    if (definition.field_type === 'number') {
      const numberValue = Number(value)
      return Number.isFinite(numberValue) ? numberValue : undefined
    }
    if (definition.field_type === 'boolean') return value === true || value === 'true' || value === '1'
    return value
  }

  const buildSpecFormValues = (product: any) => {
    const values: Record<string, any> = {}
    ;(product.spec_values || []).forEach((item: any) => {
      if (item.definition?.slug) values[item.definition.slug] = coerceSpecValueForForm(item.definition, item.value)
    })
    return values
  }

  const parseVariantOptions = (variant: any) => {
    if (!variant?.option_values) return {}
    if (typeof variant.option_values === 'object') return { ...variant.option_values }
    try {
      const parsed = JSON.parse(variant.option_values)
      return parsed && typeof parsed === 'object' ? parsed : {}
    } catch {
      return {}
    }
  }

  const normalizeDisplayPrices = (values: any) => {
    const list = Array.isArray(values) ? values : []
    const seen = new Set<string>()
    return list
      .map((item: any) => {
        const quoteCurrency = normalizeCurrencyCode(item?.quote_currency || item?.currency)
        if (!/^[A-Z]{3}$/.test(quoteCurrency)) return null
        return {
          amount: Number(item?.amount || 0),
          currency: quoteCurrency,
          quote_currency: quoteCurrency,
          rate: Number(item?.rate || 0),
          source: String(item?.source || '').trim(),
          converted: item?.converted !== false
        }
      })
      .filter(Boolean)
      .filter((item: any) => item.amount > 0 && item.currency !== primaryPriceCurrency())
      .filter((item: any) => {
        if (seen.has(item.currency)) return false
        seen.add(item.currency)
        return true
      })
  }

  const createEmptyVariant = (overrides: Record<string, any> = {}) => ({
    id: null,
    shipping_template_id: null,
    sku: '',
    title: '',
    option_values: {},
    currency: primaryPriceCurrency(),
    price: 0,
    sale_price: null,
    display_prices: [],
    stock: 0,
    weight_grams: 0,
    is_default: false,
    is_active: true,
    sort_order: productForm.variants.length * 10,
    ...overrides
  })

  const createEmptyVariantOptionValue = (
    overrides: Partial<ProductVariantOptionValueForm> = {}
  ): ProductVariantOptionValueForm => ({
    id: null,
    spec_definition_id: 0,
    value_key: '',
    label: '',
    color_hex: '',
    swatch_media_asset_id: null,
    swatch_url: '',
    sort_order: 0,
    is_enabled: true,
    ...overrides
  })

  const buildVariantOptionValueFormValues = (product: any): ProductVariantOptionValueForm[] => (
    (product.variant_option_values || []).map((item: any, index: number) => createEmptyVariantOptionValue({
      id: item.id || null,
      spec_definition_id: item.spec_definition_id || 0,
      value_key: String(item.value_key || ''),
      label: String(item.label || ''),
      color_hex: String(item.color_hex || ''),
      swatch_media_asset_id: item.swatch_media_asset_id || null,
      swatch_url: String(item.swatch_url || ''),
      sort_order: Number(item.sort_order ?? index * 10),
      is_enabled: item.is_enabled !== false
    }))
  )

  const normalizeVariantOptionValues = () => {
    const definitionIDs = new Set(
      variantSpecDefinitions.value
        .map((definition: any) => Number(definition.id || 0))
        .filter((id: number) => id > 0)
    )
    return productForm.variant_option_values
      .filter((item: ProductVariantOptionValueForm) => (
        definitionIDs.has(Number(item.spec_definition_id))
        && String(item.value_key || '').trim()
      ))
      .map((item: ProductVariantOptionValueForm, index: number) => ({
        id: item.id || undefined,
        spec_definition_id: Number(item.spec_definition_id),
        value_key: String(item.value_key || '').trim(),
        label: String(item.label || '').trim(),
        color_hex: String(item.color_hex || '').trim(),
        swatch_media_asset_id: item.swatch_media_asset_id ? Number(item.swatch_media_asset_id) : undefined,
        swatch_url: String(item.swatch_url || '').trim(),
        sort_order: Number(item.sort_order ?? index * 10),
        is_enabled: item.is_enabled !== false
      }))
  }

  const buildVariantFormValues = (product: any) => {
    const variants = (product.variants || []).map((variant: any, index: number) => createEmptyVariant({
      id: variant.id || null,
      shipping_template_id: variant.shipping_template_id ?? null,
      sku: variant.sku || '',
      title: variant.title || '',
      option_values: parseVariantOptions(variant),
      currency: primaryPriceCurrency(),
      price: Number(variant.price || 0),
      sale_price: variant.sale_price ?? null,
      display_prices: normalizeDisplayPrices(variant.display_prices),
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
    productForm.variants.push(createEmptyVariant({ is_default: productForm.variants.length === 0 }))
    clearFieldError('variants')
  }

  const removeVariant = (index: number) => {
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
    return productForm.variants.map((variant: any, index: number) => {
      const optionValues: Record<string, any> = {}
      variantSpecDefinitions.value.forEach((spec: any) => {
        const value = variant.option_values?.[spec.slug]
        if (value !== undefined && value !== null && value !== '') optionValues[spec.slug] = value
      })
      return {
        id: variant.id || undefined,
        shipping_template_id: variant.shipping_template_id == null || variant.shipping_template_id === '' ? null : Number(variant.shipping_template_id),
        sku: String(variant.sku || '').trim(),
        title: String(variant.title || '').trim(),
        option_values: optionValues,
        currency: primaryPriceCurrency(),
        price: Number(variant.price || 0),
        sale_price: variant.sale_price === '' || variant.sale_price == null ? null : Number(variant.sale_price),
        display_prices: normalizeDisplayPrices(variant.display_prices),
        stock: Number(variant.stock || 0),
        weight_grams: Number(variant.weight_grams || 0),
        is_default: Boolean(variant.is_default),
        is_active: variant.is_active !== false,
        sort_order: Number(variant.sort_order ?? index * 10)
      }
    })
  }

  const buildProductPayload = () => ({
    id: productForm.id,
    product_specification_template_id: productForm.product_specification_template_id,
    product_category_id: productForm.product_category_id,
    brand_id: productForm.brand_id,
    shipping_template_id: productForm.shipping_template_id,
    after_sales_template_id: productForm.after_sales_template_id,
    packaging_template_id: productForm.packaging_template_id,
    customs_classification_profile_id: productForm.customs_classification_profile_id || null,
    hs_code: String(productForm.hs_code || '').trim(),
    cn_code: String(productForm.cn_code || '').trim(),
    country_of_origin: String(productForm.country_of_origin || '').trim().toUpperCase(),
    customs_description: String(productForm.customs_description || '').trim(),
    name: productForm.name.trim(),
    slug: productForm.slug.trim(),
    description: productForm.description,
    short_description: productForm.short_description,
    currency: primaryPriceCurrency(),
    status: productForm.status,
    locale: productForm.locale,
    featured: productForm.featured,
    specs: { ...productForm.specs },
    variants: normalizeFormVariants(),
    variant_option_values: normalizeVariantOptionValues(),
    media: normalizeFormMedia()
  })

  const validateForm = (payload: any) => {
    clearFormErrors()
    if (!payload.name) formErrors.name = '请输入商品名称'
    if (!payload.slug) formErrors.slug = '请输入 URL slug'
    if (!payload.locale) formErrors.locale = '请选择语言'
    if (!/^[A-Z]{3}$/.test(payload.currency)) formErrors.currency = '请选择商品主基准币种'
    if (payload.hs_code && !/^\d{6}$/.test(payload.hs_code)) formErrors.hs_code = 'HS Code 必须是 6 位数字'
    if (payload.cn_code && !/^\d{8}$/.test(payload.cn_code)) formErrors.cn_code = 'CN Code 必须是 8 位数字'
    if (payload.country_of_origin && !/^[A-Z]{2}$/.test(payload.country_of_origin)) formErrors.country_of_origin = '请输入 2 位国家代码，例如 CN'
    if (payload.customs_description.length > 255) formErrors.customs_description = '英文报关品名不能超过 255 个字符'
    selectedSpecDefinitions.value.forEach((spec: any) => {
      const value = payload.specs[spec.slug]
      if (spec.is_required && (value === undefined || value === null || value === '')) {
        formErrors[`spec:${spec.slug}`] = `请填写${spec.name}`
      }
    })
    if (!payload.variants.length) formErrors.variants = '请至少添加一个 SKU 变体'
    else if (payload.variants.some((variant: any) => !variant.sku)) formErrors.variants = '每个变体都必须填写 SKU'
    else if (new Set(payload.variants.map((variant: any) => variant.sku.toLowerCase())).size !== payload.variants.length) formErrors.variants = '变体 SKU 不能重复'
    else if (payload.variants.some((variant: any) => Number(variant.price) <= 0)) formErrors.variants = '每个变体价格必须大于 0'
    else if (payload.variants.some((variant: any) => Number(variant.stock) < 0)) formErrors.variants = '变体库存不能为负数'
    else if (!payload.variants.some((variant: any) => variant.is_active !== false)) formErrors.variants = '请至少启用一个 SKU 变体'
    if (productForm.media.some((item: any) => !String(item.url || '').trim())) formErrors.media = '媒体条目必须填写 URL，空条目请删除'
    else if (payload.media.filter((item: any) => item.media_type === 'image' && item.is_primary).length > 1) formErrors.media = '商品主图只能设置一张'
    if (Object.keys(formErrors).length > 0) {
      toast.error('请检查商品表单中的必填项')
      return false
    }
    return true
  }

  const handleProductSpecTemplateSelect = (value: string) => {
    const nextProductSpecTemplateID = value === '__none__' ? null : Number(value)
    if (productForm.product_specification_template_id === nextProductSpecTemplateID) return

    const hadTemplateValues = templateScopedValuesTouched.value
    productForm.product_specification_template_id = nextProductSpecTemplateID
    productForm.customs_classification_profile_id = null
    const nextSpecs: Record<string, any> = {}
    selectedSpecDefinitions.value.forEach((spec: any) => {
      if (spec.field_type === 'boolean') nextSpecs[spec.slug] = false
    })
    productForm.specs = nextSpecs
    productForm.variants.forEach((variant: any) => { variant.option_values = {} })
    productForm.variant_option_values = []
    clearFormErrors()
    if (hadTemplateValues) {
      toast.info('已切换商品规格模板，商品参数和 SKU 选项值已按新模板重置；SKU 价格、重量、库存和媒体已保留。')
    }
  }

  const resetForm = () => {
    Object.assign(productForm, {
      id: null,
      product_specification_template_id: null,
      product_category_id: null,
      brand_id: null,
      shipping_template_id: null,
      after_sales_template_id: null,
      packaging_template_id: null,
      customs_classification_profile_id: null,
      hs_code: '',
      cn_code: '',
      country_of_origin: '',
      customs_description: '',
      name: '',
      slug: '',
      description: '',
      short_description: '',
      currency: primaryPriceCurrency(),
      status: 'active',
      locale: resolveDefaultLocale(),
      featured: false,
      specs: {},
      variants: [],
      variant_option_values: [],
      media: []
    })
    productForm.variants = [createEmptyVariant({ is_default: true })]
    clearFormErrors()
  }

  const fetchProductSpecTemplates = async () => {
    try {
      productSpecTemplates.value = await productSpecificationTemplateApi.list()
    } catch (error) {
      console.error('Failed to fetch product specification templates:', error)
    }
  }

  const showCreateDialog = async () => {
    await fetchPrimaryPricingCurrency()
    dialogMode.value = 'create'
    resetForm()
    dialogVisible.value = true
  }

  const showEditDialog = async (product: any) => {
    dialogMode.value = 'edit'
    let detail = product
    try {
      await fetchPrimaryPricingCurrency()
      if (productSpecTemplates.value.length === 0) await fetchProductSpecTemplates()
      detail = await productApi.get(product.id)
      if (detail.product_specification_template && !productSpecTemplates.value.some((template) => template.id === detail.product_specification_template.id)) {
        productSpecTemplates.value.push(detail.product_specification_template)
      }
    } catch (error) {
      toast.warning('获取商品详情失败，已使用列表数据编辑')
    }
    Object.assign(productForm, {
      id: detail.id,
      product_specification_template_id: detail.product_specification_template_id || detail.product_specification_template?.id || null,
      product_category_id: detail.product_category_id ?? detail.product_category?.id ?? null,
      brand_id: detail.brand_id ?? detail.brand?.id ?? null,
      shipping_template_id: detail.shipping_template_id ?? null,
      after_sales_template_id: detail.after_sales_template_id ?? detail.after_sales_template?.id ?? null,
      packaging_template_id: detail.packaging_template_id ?? detail.packaging_template?.id ?? null,
      customs_classification_profile_id: detail.customs_classification_profile_id ?? detail.customs_classification_profile?.id ?? null,
      hs_code: String(detail.hs_code || ''),
      cn_code: String(detail.cn_code || ''),
      country_of_origin: String(detail.country_of_origin || ''),
      customs_description: String(detail.customs_description || ''),
      name: detail.name || '',
      slug: detail.slug || '',
      description: detail.description || '',
      short_description: detail.short_description || detail.short_desc || '',
      currency: primaryPriceCurrency(),
      status: detail.status || 'active',
      locale: detail.locale || resolveDefaultLocale(),
      featured: Boolean(detail.featured),
      specs: buildSpecFormValues(detail),
      variants: buildVariantFormValues(detail),
      variant_option_values: buildVariantOptionValueFormValues(detail),
      media: buildProductMediaFormValues(detail)
    })
    clearFormErrors()
    dialogVisible.value = true
  }

  const submitForm = async () => {
    const payload = buildProductPayload()
    if (!validateForm(payload)) return
    submitting.value = true
    try {
      if (dialogMode.value === 'create') {
        await productApi.create(payload)
        toast.success('商品创建成功')
      } else {
        const { id, ...data } = payload
        await productApi.update(id, data)
        toast.success('商品更新成功')
      }
      dialogVisible.value = false
      await refreshProducts()
    } catch (error) {
      console.error('Failed to save product:', error)
    } finally {
      submitting.value = false
    }
  }

  return {
    productSpecTemplates,
    dialogVisible,
    dialogMode,
    submitting,
    formErrors,
    productForm,
    uploadingMedia,
    selectedProductSpecTemplate,
    selectedSpecDefinitions,
    variantSpecDefinitions,
    defaultVariantIndex,
    productSpecTemplateSelectValue,
    productCategorySelectValue,
    brandSelectValue,
    shippingTemplateSelectValue,
    afterSalesTemplateSelectValue,
    packagingTemplateSelectValue,
    templateScopedValuesTouched,
    parseSpecOptions,
    formatSpecOption,
    getSpecLabel,
    specSelectValue,
    setSpecSelectValue,
    setProductShippingTemplate,
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
    handleProductSpecTemplateSelect,
    fetchProductSpecTemplates,
    showCreateDialog,
    showEditDialog,
    submitForm
  }
}

export default useProductEditor

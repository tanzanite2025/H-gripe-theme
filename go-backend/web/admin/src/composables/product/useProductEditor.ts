import { computed, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import productApi from '@/api/products'
import productTypeApi from '@/api/productTypes'
import { useProductMediaManager } from '@/composables/product/useProductMediaManager'
import { buildProductMediaFormValues } from '@/lib/productMedia'

export const useProductEditor = (options: Record<string, any> = {}) => {
  const refreshProducts = options.refreshProducts || (() => Promise.resolve())
  const resolveDefaultLocale = () => options.defaultLocale?.value || options.defaultLocale || ''

  const productTypes = ref<any[]>([])
  const dialogVisible = ref(false)
  const dialogMode = ref<'create' | 'edit'>('create')
  const submitting = ref(false)
  const formErrors = reactive<Record<string, string>>({})

  const productForm = reactive<Record<string, any>>({
    id: null,
    product_type_id: null,
    shipping_template_id: null,
    name: '',
    slug: '',
    description: '',
    short_description: '',
    status: 'active',
    locale: resolveDefaultLocale(),
    featured: false,
    meta_title: '',
    meta_description: '',
    specs: {},
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

  const selectedProductType = computed(() => productTypes.value.find((type) => type.id === productForm.product_type_id) || null)
  const selectedSpecDefinitions = computed(() => (selectedProductType.value?.spec_definitions || []).filter((spec: any) => !spec.is_variant_option))
  const variantSpecDefinitions = computed(() => (selectedProductType.value?.spec_definitions || []).filter((spec: any) => spec.is_variant_option))
  const defaultVariantIndex = computed(() => {
    const index = productForm.variants.findIndex((variant: any) => variant.is_default)
    return index >= 0 ? index : 0
  })
  const productTypeSelectValue = computed(() => productForm.product_type_id == null ? '__none__' : String(productForm.product_type_id))
  const shippingTemplateSelectValue = computed(() => productForm.shipping_template_id == null ? '__none__' : String(productForm.shipping_template_id))
  const hasMeaningfulTemplateValue = (value: any) => {
    if (value === undefined || value === null || value === '') return false
    if (value === false) return false
    if (Array.isArray(value)) return value.length > 0
    if (typeof value === 'object') return Object.keys(value).length > 0
    return true
  }
  const templateScopedValuesTouched = computed(() => (
    Object.values(productForm.specs || {}).some(hasMeaningfulTemplateValue) ||
    productForm.variants.some((variant: any) => Object.values(variant.option_values || {}).some(hasMeaningfulTemplateValue))
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

  const createEmptyVariant = (overrides: Record<string, any> = {}) => ({
    id: null,
    shipping_template_id: null,
    sku: '',
    title: '',
    option_values: {},
    price: 0,
    sale_price: null,
    stock: 0,
    weight_grams: 0,
    is_default: false,
    is_active: true,
    sort_order: productForm.variants.length * 10,
    ...overrides
  })

  const buildVariantFormValues = (product: any) => {
    const variants = (product.variants || []).map((variant: any, index: number) => createEmptyVariant({
      id: variant.id || null,
      shipping_template_id: variant.shipping_template_id ?? null,
      sku: variant.sku || '',
      title: variant.title || '',
      option_values: parseVariantOptions(variant),
      price: Number(variant.price || 0),
      sale_price: variant.sale_price ?? null,
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

  const normalizeFormVariants = () => {
    if (!productForm.variants.length) return []
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
        price: Number(variant.price || 0),
        sale_price: variant.sale_price === '' || variant.sale_price == null ? null : Number(variant.sale_price),
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
    product_type_id: productForm.product_type_id,
    shipping_template_id: productForm.shipping_template_id,
    name: productForm.name.trim(),
    slug: productForm.slug.trim(),
    description: productForm.description,
    short_description: productForm.short_description,
    status: productForm.status,
    locale: productForm.locale,
    featured: productForm.featured,
    meta_title: productForm.meta_title,
    meta_description: productForm.meta_description,
    specs: { ...productForm.specs },
    variants: normalizeFormVariants(),
    media: normalizeFormMedia()
  })

  const validateForm = (payload: any) => {
    clearFormErrors()
    if (!payload.name) formErrors.name = '请输入商品名称'
    if (!payload.slug) formErrors.slug = '请输入 URL slug'
    if (!payload.locale) formErrors.locale = '请选择语言'
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
    if (productForm.media.some((item: any) => !String(item.url || '').trim())) formErrors.media = '媒体条目必须填写 URL，空条目请删除'
    else if (payload.media.filter((item: any) => item.media_type === 'image' && item.is_primary).length > 1) formErrors.media = '商品主图只能设置一张'
    if (Object.keys(formErrors).length > 0) {
      toast.error('请检查商品表单中的必填项')
      return false
    }
    return true
  }

  const handleProductTypeSelect = (value: string) => {
    const nextProductTypeID = value === '__none__' ? null : Number(value)
    if (productForm.product_type_id === nextProductTypeID) return

    const hadTemplateValues = templateScopedValuesTouched.value
    productForm.product_type_id = nextProductTypeID
    const nextSpecs: Record<string, any> = {}
    selectedSpecDefinitions.value.forEach((spec: any) => {
      if (spec.field_type === 'boolean') nextSpecs[spec.slug] = false
    })
    productForm.specs = nextSpecs
    productForm.variants.forEach((variant: any) => { variant.option_values = {} })
    clearFormErrors()
    if (hadTemplateValues) {
      toast.info('已切换产品模板，商品参数和 SKU 选项值已按新模板重置；SKU 价格、重量、库存和媒体已保留。')
    }
  }

  const resetForm = () => {
    Object.assign(productForm, {
      id: null,
      product_type_id: null,
      shipping_template_id: null,
      name: '',
      slug: '',
      description: '',
      short_description: '',
      status: 'active',
      locale: resolveDefaultLocale(),
      featured: false,
      meta_title: '',
      meta_description: '',
      specs: {},
      variants: [],
      media: []
    })
    productForm.variants = [createEmptyVariant({ is_default: true })]
    clearFormErrors()
  }

  const fetchProductTypes = async () => {
    try {
      productTypes.value = await productTypeApi.list()
    } catch (error) {
      console.error('Failed to fetch product types:', error)
    }
  }

  const showCreateDialog = () => {
    dialogMode.value = 'create'
    resetForm()
    dialogVisible.value = true
  }

  const showEditDialog = async (product: any) => {
    dialogMode.value = 'edit'
    let detail = product
    try {
      if (productTypes.value.length === 0) await fetchProductTypes()
      detail = await productApi.get(product.id)
      if (detail.product_type && !productTypes.value.some((type) => type.id === detail.product_type.id)) {
        productTypes.value.push(detail.product_type)
      }
    } catch (error) {
      toast.warning('获取商品详情失败，已使用列表数据编辑')
    }
    Object.assign(productForm, {
      id: detail.id,
      product_type_id: detail.product_type_id || detail.product_type?.id || null,
      shipping_template_id: detail.shipping_template_id ?? null,
      name: detail.name || '',
      slug: detail.slug || '',
      description: detail.description || '',
      short_description: detail.short_description || detail.short_desc || '',
      status: detail.status || 'active',
      locale: detail.locale || resolveDefaultLocale(),
      featured: Boolean(detail.featured),
      meta_title: detail.meta_title || '',
      meta_description: detail.meta_description || detail.meta_desc || '',
      specs: buildSpecFormValues(detail),
      variants: buildVariantFormValues(detail),
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
    productTypes,
    dialogVisible,
    dialogMode,
    submitting,
    formErrors,
    productForm,
    uploadingMedia,
    selectedProductType,
    selectedSpecDefinitions,
    variantSpecDefinitions,
    defaultVariantIndex,
    productTypeSelectValue,
    shippingTemplateSelectValue,
    templateScopedValuesTouched,
    parseSpecOptions,
    formatSpecOption,
    getSpecLabel,
    specSelectValue,
    setSpecSelectValue,
    setProductShippingTemplate,
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
    handleProductTypeSelect,
    fetchProductTypes,
    showCreateDialog,
    showEditDialog,
    submitForm
  }
}

export default useProductEditor

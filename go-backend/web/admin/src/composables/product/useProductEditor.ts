import { computed, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import productApi from '@/api/products'
import productSpecificationTemplateApi from '@/api/productSpecificationTemplates'
import { useProductMediaManager } from '@/composables/product/useProductMediaManager'
import { buildProductMediaFormValues } from '@/lib/productMedia'
import axios from '@/utils/axios'
import type {
  ProductFormRecord,
  ProductOptionValueRelationForm,
  ProductTemplateSyncDiff,
  ProductVariantOptionValueForm
} from '@/modules/product/productEditorTypes'

type ProductEditorMode = 'create' | 'edit'

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

  const productSpecTemplates = ref<any[]>([])
  const primaryCurrency = ref(defaultPrimaryCurrency)
  const currencyPolicyLoaded = ref(false)
  const dialogVisible = ref(false)
  const dialogMode = ref<'create' | 'edit'>('create')
  const submitting = ref(false)
  const formErrors = reactive<Record<string, string>>({})
  const templateSyncDialogVisible = ref(false)
  const templateSyncDiff = ref<ProductTemplateSyncDiff | null>(null)
  const templateSyncLoading = ref(false)
  const templateSyncApplying = ref(false)

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
    option_value_relations: [],
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
  const definitionRole = (spec: any): string => {
    const role = String(spec?.role || '').trim()
    if (role === 'custom_option' || role === 'variant' || role === 'attribute') return role
    return 'attribute'
  }
  const selectedSpecDefinitions = computed(() => (selectedProductSpecTemplate.value?.spec_definitions || []).filter((spec: any) => definitionRole(spec) === 'attribute'))
  const variantSpecDefinitions = computed(() => (selectedProductSpecTemplate.value?.spec_definitions || []).filter((spec: any) => definitionRole(spec) === 'variant'))
  const customOptionDefinitions = computed(() => (selectedProductSpecTemplate.value?.spec_definitions || []).filter((spec: any) => definitionRole(spec) === 'custom_option'))
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
    const templateValues = (spec?.option_items || [])
      .map((item: any) => String(item?.value_key || '').trim())
      .filter(Boolean)
    if (templateValues.length) return templateValues
    return []
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

  const createEmptyVariant = (overrides: Record<string, any> = {}) => ({
    id: null,
    shipping_template_id: null,
    sku: '',
    title: '',
    option_values: {},
    currency: primaryPriceCurrency(),
    price: '0.00',
    sale_price: '',
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
      template_option_item_id: item.template_option_item_id || null,
      source_template_revision: item.source_template_revision ?? 0,
      value_key: String(item.value_key || ''),
      label: String(item.label || ''),
      color_hex: String(item.color_hex || ''),
      swatch_media_asset_id: item.swatch_media_asset_id || null,
      swatch_url: String(item.swatch_url || ''),
      sort_order: Number(item.sort_order ?? index * 10),
      is_enabled: item.is_enabled !== false,
      price_delta_minor: item.price_delta_minor ?? item.custom_option_policy?.price_delta_minor ?? null,
      is_default: Boolean(item.is_default ?? item.custom_option_policy?.is_default),
      inventory_policy: item.inventory_policy || item.custom_option_policy?.inventory_policy || 'none',
      component_variant_id: item.component_variant_id ?? item.custom_option_policy?.component_variant_id ?? null,
      component_quantity: item.component_quantity ?? item.custom_option_policy?.component_quantity ?? 0,
      weight_delta_grams: item.weight_delta_grams ?? item.custom_option_policy?.weight_delta_grams ?? 0,
      packaging_weight_delta_grams: item.packaging_weight_delta_grams ?? item.custom_option_policy?.packaging_weight_delta_grams ?? 0,
      production_lead_time_days: item.production_lead_time_days ?? item.custom_option_policy?.production_lead_time_days ?? 0,
      requires_production: Boolean(item.requires_production ?? item.custom_option_policy?.requires_production),
      cancellation_policy: item.cancellation_policy ?? item.custom_option_policy?.cancellation_policy ?? '',
      return_policy: item.return_policy ?? item.custom_option_policy?.return_policy ?? 'standard'
    }))
  )

  const buildOptionValueRelationFormValues = (product: any): ProductOptionValueRelationForm[] => (
    (product.option_value_relations || []).map((relation: any) => ({
      id: relation.id || null,
      source_option_value_id: Number(relation.source_option_value_id || 0),
      target_option_value_id: Number(relation.target_option_value_id || 0),
      relation_type: relation.relation_type === 'conflicts' ? 'conflicts' : 'requires'
    }))
  )

  const normalizeVariantOptionValues = () => {
    const definitionIDs = new Set(
      [...variantSpecDefinitions.value, ...customOptionDefinitions.value]
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
        template_option_item_id: item.template_option_item_id == null ? undefined : Number(item.template_option_item_id),
        source_template_revision: Number(item.source_template_revision || 0),
        value_key: String(item.value_key || '').trim(),
        label: String(item.label || '').trim(),
        color_hex: String(item.color_hex || '').trim(),
        swatch_media_asset_id: item.swatch_media_asset_id ? Number(item.swatch_media_asset_id) : undefined,
        swatch_url: String(item.swatch_url || '').trim(),
        sort_order: Number(item.sort_order ?? index * 10),
        is_enabled: item.is_enabled !== false,
        price_delta_minor: item.price_delta_minor == null ? undefined : Number(item.price_delta_minor),
        is_default: Boolean(item.is_default),
        inventory_policy: String(item.inventory_policy || 'none'),
        component_variant_id: item.component_variant_id == null ? undefined : Number(item.component_variant_id),
        component_quantity: Number(item.component_quantity || 0),
        weight_delta_grams: Number(item.weight_delta_grams || 0),
        packaging_weight_delta_grams: Number(item.packaging_weight_delta_grams || 0),
        production_lead_time_days: Number(item.production_lead_time_days || 0),
        requires_production: Boolean(item.requires_production),
        cancellation_policy: String(item.cancellation_policy || ''),
        return_policy: String(item.return_policy || 'standard')
      }))
  }

  const normalizeOptionValueRelations = () => productForm.option_value_relations.map((relation) => ({
    id: relation.id || undefined,
    source_option_value_id: Number(relation.source_option_value_id || 0),
    target_option_value_id: Number(relation.target_option_value_id || 0),
    relation_type: relation.relation_type === 'conflicts' ? 'conflicts' : 'requires'
  }))

  const materializeTemplateOptionValues = (template: any | null): void => {
    if (!template) {
      productForm.variant_option_values = []
      return
    }
    const revision = Number(template.revision || 1)
    const values: ProductVariantOptionValueForm[] = []
    ;(template.spec_definitions || []).forEach((definition: any) => {
      if (definitionRole(definition) !== 'variant' && definitionRole(definition) !== 'custom_option') return
      ;(definition.option_items || []).forEach((item: any, index: number) => {
        const enabled = item.is_enabled_by_default !== false
        values.push(createEmptyVariantOptionValue({
          spec_definition_id: Number(definition.id || 0),
          template_option_item_id: item.id || null,
          source_template_revision: revision,
          value_key: String(item.value_key || '').trim(),
          label: String(item.default_label || item.value_key || '').trim(),
          color_hex: String(item.color_hex || ''),
          swatch_media_asset_id: item.swatch_media_asset_id || null,
          swatch_url: String(item.swatch_url || ''),
          sort_order: Number(item.sort_order ?? index * 10),
          is_enabled: enabled,
          price_delta_minor: definitionRole(definition) === 'custom_option' ? (item.default_price_delta_minor ?? null) : null,
          is_default: definitionRole(definition) === 'custom_option' ? Boolean(item.is_default) : false,
          inventory_policy: definitionRole(definition) === 'custom_option' ? 'none' : undefined,
          component_variant_id: null,
          component_quantity: 0,
          weight_delta_grams: 0,
          packaging_weight_delta_grams: 0,
          production_lead_time_days: 0,
          requires_production: false,
          cancellation_policy: '',
          return_policy: 'standard'
        }))
      })
    })
    productForm.variant_option_values = values
  }

  const buildVariantFormValues = (product: any) => {
    const variants = (product.variants || []).map((variant: any, index: number) => createEmptyVariant({
      id: variant.id || null,
      shipping_template_id: variant.shipping_template_id ?? null,
      sku: variant.sku || '',
      title: variant.title || '',
      option_values: parseVariantOptions(variant),
      currency: validCurrencyCodeOrDefault(variant.currency || product.currency),
      price_minor: Number.isFinite(Number(variant.price_minor)) ? Number(variant.price_minor) : 0,
      price: majorFromMinor(Number.isFinite(Number(variant.price_minor)) ? Number(variant.price_minor) : 0, variant.currency || product.currency),
      sale_price_minor: variant.sale_price_minor == null ? null : Number(variant.sale_price_minor),
      sale_price: variant.sale_price_minor == null ? null : majorFromMinor(variant.sale_price_minor, variant.currency || product.currency),
      stock: Number(variant.stock || 0),
      weight_grams: variant.weight_grams ?? variant.weight ?? 0,
      is_default: Boolean(variant.is_default),
      is_active: variant.is_active !== false,
      sort_order: variant.sort_order ?? index * 10,
      option_group_rules: (variant.option_group_rules || []).map((rule: any) => ({
        id: rule.id || null,
        spec_definition_id: Number(rule.spec_definition_id || 0),
        is_applicable: rule.is_applicable !== false,
        min_selections_override: rule.min_selections_override ?? null,
        max_selections_override: rule.max_selections_override ?? null
      })),
      option_value_rules: (variant.option_value_rules || []).map((rule: any) => ({
        id: rule.id || null,
        product_variant_option_value_id: Number(rule.product_variant_option_value_id || 0),
        is_enabled: rule.is_enabled !== false,
        price_delta_minor_override: rule.price_delta_minor_override ?? null,
        unavailable_reason: String(rule.unavailable_reason || '')
      }))
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
        currency: validCurrencyCodeOrDefault(variant.currency || productForm.currency),
        price: String(variant.price ?? ''),
        price_minor: minorFromMajor(variant.price, variant.currency || productForm.currency),
        sale_price: variant.sale_price === '' || variant.sale_price == null ? null : String(variant.sale_price),
        sale_price_minor: variant.sale_price === '' || variant.sale_price == null ? null : minorFromMajor(variant.sale_price, variant.currency || productForm.currency),
        stock: Number(variant.stock || 0),
        weight_grams: Number(variant.weight_grams || 0),
        is_default: Boolean(variant.is_default),
        is_active: variant.is_active !== false,
        sort_order: Number(variant.sort_order ?? index * 10),
        option_group_rules: (variant.option_group_rules || []).filter((rule: any) => Number(rule.spec_definition_id) > 0).map((rule: any) => ({
          id: rule.id || undefined,
          spec_definition_id: Number(rule.spec_definition_id),
          is_applicable: rule.is_applicable !== false,
          min_selections_override: rule.min_selections_override == null || rule.min_selections_override === '' ? null : Number(rule.min_selections_override),
          max_selections_override: rule.max_selections_override == null || rule.max_selections_override === '' ? null : Number(rule.max_selections_override)
        })),
        option_value_rules: (variant.option_value_rules || []).filter((rule: any) => Number(rule.product_variant_option_value_id) > 0).map((rule: any) => ({
          id: rule.id || undefined,
          product_variant_option_value_id: Number(rule.product_variant_option_value_id),
          is_enabled: rule.is_enabled !== false,
          price_delta_minor_override: rule.price_delta_minor_override == null || rule.price_delta_minor_override === '' ? null : Number(rule.price_delta_minor_override),
          unavailable_reason: String(rule.unavailable_reason || '').trim()
        }))
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
    currency: validCurrencyCodeOrDefault(productForm.currency),
    status: productForm.status,
    locale: productForm.locale,
    featured: productForm.featured,
    specs: { ...productForm.specs },
    variants: normalizeFormVariants(),
    variant_option_values: normalizeVariantOptionValues(),
    option_value_relations: normalizeOptionValueRelations(),
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
    const persistedOptionValueIDs = new Set(
      payload.variant_option_values
        .map((item: any) => Number(item.id || 0))
        .filter((id: number) => id > 0)
    )
    const relationKeys = new Set<string>()
    for (const relation of payload.option_value_relations) {
      const sourceID = Number(relation.source_option_value_id || 0)
      const targetID = Number(relation.target_option_value_id || 0)
      if (!persistedOptionValueIDs.has(sourceID) || !persistedOptionValueIDs.has(targetID)) {
        formErrors.option_value_relations = '选项关系只能使用已保存的选项值'
        break
      }
      if (sourceID === targetID) {
        formErrors.option_value_relations = '关系两端不能选择同一个选项值'
        break
      }
      const relationKey = relation.relation_type === 'conflicts'
        ? `conflicts:${Math.min(sourceID, targetID)}:${Math.max(sourceID, targetID)}`
        : `requires:${sourceID}:${targetID}`
      if (relationKeys.has(relationKey)) {
        formErrors.option_value_relations = '不能重复添加相同的选项关系'
        break
      }
      relationKeys.add(relationKey)
    }
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
    materializeTemplateOptionValues(selectedProductSpecTemplate.value)
    productForm.option_value_relations = []
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
      option_value_relations: [],
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

  const showCreateDialog = async () => {
    await fetchPrimaryPricingCurrency()
    dialogMode.value = 'create'
    resetForm()
    templateSyncDialogVisible.value = false
    templateSyncDiff.value = null
    await notifyProductLoaded(null, 'create')
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
      currency: validCurrencyCodeOrDefault(detail.currency),
      status: detail.status || 'active',
      locale: detail.locale || resolveDefaultLocale(),
      featured: Boolean(detail.featured),
      specs: buildSpecFormValues(detail),
      variants: buildVariantFormValues(detail),
      variant_option_values: buildVariantOptionValueFormValues(detail),
      option_value_relations: buildOptionValueRelationFormValues(detail),
      media: buildProductMediaFormValues(detail)
    })
    clearFormErrors()
    await notifyProductLoaded(detail, 'edit')
    dialogVisible.value = true
  }

  const previewTemplateSync = async () => {
    if (dialogMode.value !== 'edit' || !productForm.id || templateSyncLoading.value) return
    templateSyncLoading.value = true
    templateSyncDiff.value = null
    try {
      templateSyncDiff.value = await productApi.previewTemplateSync(productForm.id) as ProductTemplateSyncDiff
      templateSyncDialogVisible.value = true
    } catch (error: any) {
      console.error('Failed to preview product template sync:', error)
      toast.error(error?.response?.data?.error || '读取模板差异失败')
    } finally {
      templateSyncLoading.value = false
    }
  }

  const confirmTemplateSync = async () => {
    if (!productForm.id || !templateSyncDiff.value || templateSyncApplying.value) return
    templateSyncApplying.value = true
    try {
      const payload = await productApi.syncTemplate(productForm.id, templateSyncDiff.value.template_revision)
      templateSyncDialogVisible.value = false
      templateSyncDiff.value = payload.data as ProductTemplateSyncDiff
      await showEditDialog(payload.product)
      toast.success('商品规格模板已同步')
    } catch (error: any) {
      console.error('Failed to synchronize product template:', error)
      toast.error(error?.response?.data?.error || '同步商品规格模板失败，请重新预览')
    } finally {
      templateSyncApplying.value = false
    }
  }

  const submitForm = async () => {
    const payload = buildProductPayload()
    if (!validateForm(payload)) return
    submitting.value = true
    try {
      let savedProduct: any
      if (dialogMode.value === 'create') {
        savedProduct = await productApi.create(payload)
        toast.success('商品创建成功')
      } else {
        const { id, ...data } = payload
        savedProduct = await productApi.update(id, data)
        toast.success('商品更新成功')
      }
      const afterSaveResult = await notifyProductSaved(savedProduct, dialogMode.value)
      if (!afterSaveResult.keepDialogOpen) dialogVisible.value = false
      await refreshProducts()
    } catch (error) {
      console.error('Failed to save product:', error)
    } finally {
      submitting.value = false
    }
  }

  const closeDialog = () => {
    dialogVisible.value = false
  }

  return {
    productSpecTemplates,
    dialogVisible,
    dialogMode,
    submitting,
    templateSyncDialogVisible,
    templateSyncDiff,
    templateSyncLoading,
    templateSyncApplying,
    formErrors,
    productForm,
    uploadingMedia,
    selectedProductSpecTemplate,
    selectedSpecDefinitions,
    variantSpecDefinitions,
    customOptionDefinitions,
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
    previewTemplateSync,
    confirmTemplateSync,
    submitForm,
    closeDialog
  }
}

export default useProductEditor


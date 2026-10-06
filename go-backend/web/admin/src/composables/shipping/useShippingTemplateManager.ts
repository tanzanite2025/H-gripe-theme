import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import {
  clearErrors,
  defaultShippingTemplateForm,
  resetReactive,
} from '@/lib/shippingForms'

const TEMPLATE_DISPLAY_PRICE_FIELDS = ['default_fee', 'free_threshold']
const RULE_DISPLAY_PRICE_FIELDS = ['min_value', 'max_value', 'fee', 'additional']

export const useShippingTemplateManager = (options: Record<string, any> = {}) => {
  const fetchTemplates = options.fetchTemplates || (() => Promise.resolve())
  const fetchCarrierServices = options.fetchCarrierServices || (() => Promise.resolve())
  const carrierServices = options.carrierServices

  const templateDialogOpen = ref(false)
  const templateDialogMode = ref<'create' | 'edit'>('create')
  const templateSubmitting = ref(false)
  const templateErrors = reactive<Record<string, string>>({})
  const templateForm = reactive(defaultShippingTemplateForm())

  const clearTemplateError = (field: string) => {
    delete templateErrors[field]
  }

  const normalizeCurrencyCode = (value: any) => {
    const code = String(value || '').trim().toUpperCase()
    return /^[A-Z]{3}$/.test(code) ? code : ''
  }

  const normalizeDisplayPrices = (values: any) => {
    const list = Array.isArray(values) ? values : []
    const seen = new Set<string>()
    return list
      .map((item: any) => {
        const quoteCurrency = normalizeCurrencyCode(item?.quote_currency || item?.currency)
        if (!quoteCurrency || item?.fallback_reason) return null
        return {
          amount_decimal: String(item?.amount_decimal ?? item?.amount ?? '0'),
          currency: quoteCurrency,
          quote_currency: quoteCurrency,
          rate: Number(item?.rate || 0),
          source: String(item?.source || '').trim(),
          converted: item?.converted !== false,
        }
      })
      .filter(Boolean)
      .filter((item: any) => Number(item.amount_decimal) > 0 && item.converted !== false)
      .filter((item: any) => {
        if (seen.has(item.currency)) return false
        seen.add(item.currency)
        return true
      })
  }

  const normalizeDisplayPriceSnapshotMap = (value: any, allowedFields: string[]) => {
    if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
    return allowedFields.reduce((result: Record<string, any[]>, field) => {
      const prices = normalizeDisplayPrices(value[field])
      if (prices.length) result[field] = prices
      return result
    }, {})
  }

  const ruleDisplayPriceFieldsForType = (templateType: string) =>
    templateType === 'price' ? RULE_DISPLAY_PRICE_FIELDS : ['fee', 'additional']

  const normalizedTemplateCurrency = () => normalizeCurrencyCode(templateForm.currency) || 'USD'

  const normalizeFreeShippingCountries = (value: any) => {
    const raw = Array.isArray(value) ? value : String(value || '').trim()
    if (Array.isArray(raw)) {
      return JSON.stringify(Array.from(new Set(raw.map((code: any) => String(code || '').trim().toUpperCase()).filter(Boolean))).sort())
    }
    if (!raw) return '[]'
    try {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) return normalizeFreeShippingCountries(parsed)
    } catch {
      // Keep compatibility with older comma-separated country scopes.
    }
    return normalizeFreeShippingCountries(String(raw).split(/[,，;|\s]+/))
  }

  const showCreateTemplateDialog = () => {
    templateDialogMode.value = 'create'
    resetReactive(templateForm, defaultShippingTemplateForm())
    clearErrors(templateErrors)
    templateDialogOpen.value = true
  }

  const normalizeTemplateCarrierServices = (services: any[] = []) => services.map((service: any) => ({
    id: service.id,
    carrier_id: service.carrier_id,
    fpx_channel_id: service.fpx_channel_id ?? null,
    yanwen_published_channel_id: service.yanwen_published_channel_id ?? null,
    provider_code: String(service.provider_code || service.carrier?.code || (String(service.service_code || '').toUpperCase().startsWith('YANWEN:') ? 'YANWEN' : '')).trim().toUpperCase(),
    service_code: String(service.service_code || '').trim().toUpperCase(),
    service_name: String(service.service_name || '').trim(),
    route_name: String(service.route_name || '').trim(),
    countries: String(service.countries || '[]'),
    currency: normalizeCurrencyCode(service.currency) || normalizedTemplateCurrency(),
    billing_mode: service.billing_mode || 'actual_weight',
    first_weight_grams: Number(service.first_weight_grams || 0),
    additional_weight_grams: Number(service.additional_weight_grams || 0),
    min_charge_weight_grams: Number(service.min_charge_weight_grams || 0),
    volumetric_divisor: Number(service.volumetric_divisor || 6000),
    fuel_surcharge_percent_decimal: String(service.fuel_surcharge_percent_decimal ?? '0'),
    remote_surcharge_minor: Number(service.remote_surcharge_minor || 0),
    remote_postal_codes: String(service.remote_postal_codes || '[]'),
    eta_min_days: Number(service.eta_min_days || 0),
    eta_max_days: Number(service.eta_max_days || 0),
    enabled: service.enabled !== false,
    sort_order: Number(service.sort_order || 0),
    description: String(service.description || ''),
  }))

  const showEditTemplateDialog = (template: any) => {
    templateDialogMode.value = 'edit'
    resetReactive(templateForm, {
      ...defaultShippingTemplateForm(),
      ...template,
      template_kind: String(template.template_kind || (template.type === 'free_shipping' ? 'system_free_shipping' : 'carrier')).trim().toLowerCase() || 'carrier',
      is_system_managed: template.is_system_managed === true,
      free_shipping_countries: normalizeFreeShippingCountries(template.free_shipping_countries),
      currency: normalizeCurrencyCode(template.currency) || 'USD',
      free_threshold_minor: Number(template.free_threshold_minor || 0),
      default_fee_minor: Number(template.default_fee_minor || 0),
      display_price_snapshots: normalizeDisplayPriceSnapshotMap(template.display_price_snapshots, TEMPLATE_DISPLAY_PRICE_FIELDS),
      enabled: template.enabled !== false,
      carrier_services: normalizeTemplateCarrierServices(
        Array.isArray(carrierServices?.value)
          ? carrierServices.value.filter((service: any) => Number(service.template_id) === Number(template.id))
          : [],
      ),
      rules: Array.isArray(template.rules) ? template.rules.map((rule: any) => ({
        id: rule.id,
        region: rule.region || '',
        currency: normalizeCurrencyCode(rule.currency) || normalizeCurrencyCode(template.currency) || 'USD',
        min_value: Number(rule.min_value || 0),
        max_value: Number(rule.max_value || 0),
        min_value_minor: Number(rule.min_value_minor || 0),
        max_value_minor: Number(rule.max_value_minor || 0),
        fee_minor: Number(rule.fee_minor || 0),
        additional_minor: Number(rule.additional_minor || 0),
        display_price_snapshots: normalizeDisplayPriceSnapshotMap(rule.display_price_snapshots, ruleDisplayPriceFieldsForType(template.type)),
      })) : [],
    })
    clearErrors(templateErrors)
    templateDialogOpen.value = true
  }

  const normalizeTemplateRules = () => (Array.isArray(templateForm.rules) ? templateForm.rules : [])
    .map((rule: any) => ({
      region: String(rule.region || '').trim().toUpperCase(),
      currency: normalizeCurrencyCode(rule.currency) || normalizedTemplateCurrency(),
      ...(templateForm.type === 'price'
        ? {
            min_value_minor: Number(rule.min_value_minor || 0),
            max_value_minor: Number(rule.max_value_minor || 0),
          }
        : {
            min_value: Number(rule.min_value || 0),
            max_value: Number(rule.max_value || 0),
          }),
      fee_minor: Number(rule.fee_minor || 0),
      additional_minor: Number(rule.additional_minor || 0),
      display_price_snapshots: normalizeDisplayPriceSnapshotMap(rule.display_price_snapshots, ruleDisplayPriceFieldsForType(templateForm.type)),
    }))

  const validateTemplate = () => {
    clearErrors(templateErrors)
    const isSystemFreeShippingTemplate = templateForm.template_kind === 'system_free_shipping' || templateForm.type === 'free_shipping'
    if (!templateForm.name?.trim()) templateErrors.name = '请输入模板名称'
    if (!isSystemFreeShippingTemplate && !['weight', 'quantity', 'price'].includes(templateForm.type)) templateErrors.type = '请选择计费类型'
    if (!normalizeCurrencyCode(templateForm.currency)) templateErrors.currency = '请输入运费录入币种'
    if (Number(templateForm.default_fee_minor) < 0) templateErrors.default_fee_minor = '默认运费不能小于 0'

    if (templateForm.template_kind === 'system_free_shipping' || templateForm.type === 'free_shipping') {
      templateForm.free_shipping_countries = normalizeFreeShippingCountries(templateForm.free_shipping_countries)
      const selectedCountries = JSON.parse(templateForm.free_shipping_countries) as string[]
      if (!selectedCountries.length) templateErrors.free_shipping_countries = '请至少选择一个免邮国家'
      if (!templateForm.free_shipping) templateErrors.free_shipping = '系统免邮模板必须保持开启'
    }

    const invalidRule = normalizeTemplateRules().find((rule: any) => {
      const minValue = templateForm.type === 'price' ? rule.min_value_minor : rule.min_value
      const maxValue = templateForm.type === 'price' ? rule.max_value_minor : rule.max_value
      return !rule.region || minValue < 0 || maxValue < 0 || rule.fee_minor < 0 || rule.additional_minor < 0 || (maxValue > 0 && maxValue < minValue)
    })
    if (invalidRule) {
      toast.error('请检查规则矩阵：Region 必填，数值不能小于 0，最大值不能小于最小值')
      return false
    }

    return Object.keys(templateErrors).length === 0
  }

  const saveTemplate = async () => {
    if (!validateTemplate()) return

    templateSubmitting.value = true
    try {
      const payload = {
        name: templateForm.name.trim(),
        type: templateForm.type,
        template_kind: templateForm.template_kind === 'system_free_shipping' || templateForm.type === 'free_shipping' ? 'system_free_shipping' : 'carrier',
        currency: normalizedTemplateCurrency(),
        free_shipping: templateForm.template_kind === 'system_free_shipping' || templateForm.type === 'free_shipping' ? true : Boolean(templateForm.free_shipping),
        free_threshold_minor: Number(templateForm.free_threshold_minor || 0),
        default_fee_minor: Number(templateForm.default_fee_minor || 0),
        free_shipping_countries: normalizeFreeShippingCountries(templateForm.free_shipping_countries),
        display_price_snapshots: normalizeDisplayPriceSnapshotMap(templateForm.display_price_snapshots, TEMPLATE_DISPLAY_PRICE_FIELDS),
        description: templateForm.description || '',
        enabled: Boolean(templateForm.enabled),
        rules: normalizeTemplateRules(),
        carrier_services: templateForm.template_kind === 'system_free_shipping' || templateForm.type === 'free_shipping'
          ? []
          : normalizeTemplateCarrierServices(templateForm.carrier_services),
      }

      if (templateDialogMode.value === 'create') {
        await shippingApi.createTemplate(payload)
        toast.success('运费模板已创建')
      } else {
        await shippingApi.updateTemplate(templateForm.id, payload)
        toast.success('运费模板已更新')
      }

      templateDialogOpen.value = false
      await fetchTemplates()
      await fetchCarrierServices()
    } catch (error) {
      console.error('Failed to save shipping template:', error)
    } finally {
      templateSubmitting.value = false
    }
  }

  return {
    templateDialogOpen,
    templateDialogMode,
    templateSubmitting,
    templateErrors,
    templateForm,
    clearTemplateError,
    showCreateTemplateDialog,
    showEditTemplateDialog,
    saveTemplate,
  }
}

export default useShippingTemplateManager

import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import {
  clearErrors,
  defaultShippingTemplateForm,
  defaultShippingZoneForm,
  resetReactive,
} from '@/lib/shippingForms'

export const useShippingTemplateManager = (options: Record<string, any> = {}) => {
  const fetchTemplates = options.fetchTemplates || (() => Promise.resolve())
  const fetchZones = options.fetchZones || (() => Promise.resolve())

  const templateDialogOpen = ref(false)
  const templateDialogMode = ref<'create' | 'edit'>('create')
  const templateSubmitting = ref(false)
  const templateErrors = reactive<Record<string, string>>({})
  const templateForm = reactive(defaultShippingTemplateForm())

  const zoneDialogOpen = ref(false)
  const zoneDialogMode = ref<'create' | 'edit'>('create')
  const zoneSubmitting = ref(false)
  const zoneErrors = reactive<Record<string, string>>({})
  const zoneForm = reactive(defaultShippingZoneForm())

  const clearTemplateError = (field: string) => {
    delete templateErrors[field]
  }

  const clearZoneError = (field: string) => {
    delete zoneErrors[field]
  }

  const showCreateTemplateDialog = () => {
    templateDialogMode.value = 'create'
    resetReactive(templateForm, defaultShippingTemplateForm())
    clearErrors(templateErrors)
    templateDialogOpen.value = true
  }

  const showEditTemplateDialog = (template: any) => {
    templateDialogMode.value = 'edit'
    resetReactive(templateForm, {
      ...defaultShippingTemplateForm(),
      ...template,
      free_threshold: Number(template.free_threshold || 0),
      default_fee: Number(template.default_fee || 0),
      enabled: template.enabled !== false,
      rules: Array.isArray(template.rules) ? template.rules.map((rule: any) => ({
        id: rule.id,
        region: rule.region || '',
        min_value: Number(rule.min_value || 0),
        max_value: Number(rule.max_value || 0),
        fee: Number(rule.fee || 0),
        additional: Number(rule.additional || 0),
      })) : [],
    })
    clearErrors(templateErrors)
    templateDialogOpen.value = true
  }

  const normalizeTemplateRules = () => (Array.isArray(templateForm.rules) ? templateForm.rules : [])
    .map((rule: any) => ({
      region: String(rule.region || '').trim().toUpperCase(),
      min_value: Number(rule.min_value || 0),
      max_value: Number(rule.max_value || 0),
      fee: Number(rule.fee || 0),
      additional: Number(rule.additional || 0),
    }))

  const validateTemplate = () => {
    clearErrors(templateErrors)
    if (!templateForm.name?.trim()) templateErrors.name = '请输入模板名称'
    if (!['weight', 'quantity', 'price'].includes(templateForm.type)) templateErrors.type = '请选择计费类型'
    if (Number(templateForm.default_fee) < 0) templateErrors.default_fee = '默认运费不能小于 0'

    const invalidRule = normalizeTemplateRules().find((rule: any) =>
      !rule.region || rule.min_value < 0 || rule.max_value < 0 || rule.fee < 0 || rule.additional < 0 || (rule.max_value > 0 && rule.max_value < rule.min_value)
    )
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
        free_shipping: Boolean(templateForm.free_shipping),
        free_threshold: Number(templateForm.free_threshold || 0),
        default_fee: Number(templateForm.default_fee || 0),
        description: templateForm.description || '',
        enabled: Boolean(templateForm.enabled),
        rules: normalizeTemplateRules(),
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
    } catch (error) {
      console.error('Failed to save shipping template:', error)
    } finally {
      templateSubmitting.value = false
    }
  }

  const showCreateZoneDialog = () => {
    zoneDialogMode.value = 'create'
    resetReactive(zoneForm, defaultShippingZoneForm())
    clearErrors(zoneErrors)
    zoneDialogOpen.value = true
  }

  const showEditZoneDialog = (zone: any) => {
    zoneDialogMode.value = 'edit'
    resetReactive(zoneForm, {
      ...defaultShippingZoneForm(),
      ...zone,
      countries: zone.countries || '[]',
      states: zone.states || '[]',
      postal_codes: zone.postal_codes || '[]',
      enabled: zone.enabled !== false,
    })
    clearErrors(zoneErrors)
    zoneDialogOpen.value = true
  }

  const validateZone = () => {
    clearErrors(zoneErrors)
    if (!zoneForm.name?.trim()) zoneErrors.name = '请输入区域名称'
    if (!zoneForm.countries?.trim() || zoneForm.countries.trim() === '[]') zoneErrors.countries = '请输入至少一个国家/地区代码'
    return Object.keys(zoneErrors).length === 0
  }

  const saveZone = async () => {
    if (!validateZone()) return

    zoneSubmitting.value = true
    try {
      const payload = {
        name: zoneForm.name.trim(),
        countries: zoneForm.countries.trim(),
        states: zoneForm.states?.trim() || '[]',
        postal_codes: zoneForm.postal_codes?.trim() || '[]',
        enabled: Boolean(zoneForm.enabled),
      }

      if (zoneDialogMode.value === 'create') {
        await shippingApi.createZone(payload)
        toast.success('配送区域已创建')
      } else {
        await shippingApi.updateZone(zoneForm.id, payload)
        toast.success('配送区域已更新')
      }

      zoneDialogOpen.value = false
      await fetchZones()
    } catch (error) {
      console.error('Failed to save shipping zone:', error)
    } finally {
      zoneSubmitting.value = false
    }
  }

  return {
    templateDialogOpen,
    templateDialogMode,
    templateSubmitting,
    templateErrors,
    templateForm,
    zoneDialogOpen,
    zoneDialogMode,
    zoneSubmitting,
    zoneErrors,
    zoneForm,
    clearTemplateError,
    clearZoneError,
    showCreateTemplateDialog,
    showEditTemplateDialog,
    saveTemplate,
    showCreateZoneDialog,
    showEditZoneDialog,
    saveZone,
  }
}

export default useShippingTemplateManager

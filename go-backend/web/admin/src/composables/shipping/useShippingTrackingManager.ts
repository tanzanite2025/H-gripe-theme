import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import {
  clearErrors,
  defaultShippingTrackingCarrierMappingForm,
  defaultShippingTrackingProviderForm,
  resetReactive,
} from '@/lib/shippingForms'

const nullablePositiveID = (value: any) => {
  const numericValue = Number(value)
  return Number.isFinite(numericValue) && numericValue > 0 ? numericValue : null
}

export const useShippingTrackingManager = (options: Record<string, any> = {}) => {
  const trackingProviders = options.trackingProviders
  const carriers = options.carriers
  const carrierServices = options.carrierServices
  const fetchTrackingProviders = options.fetchTrackingProviders || (() => Promise.resolve())
  const fetchTrackingCarrierMappings = options.fetchTrackingCarrierMappings || (() => Promise.resolve())

  const trackingProviderDialogOpen = ref(false)
  const trackingProviderDialogMode = ref<'create' | 'edit'>('create')
  const trackingProviderSubmitting = ref(false)
  const trackingProviderErrors = reactive<Record<string, string>>({})
  const trackingProviderForm = reactive(defaultShippingTrackingProviderForm())

  const trackingCarrierMappingDialogOpen = ref(false)
  const trackingCarrierMappingDialogMode = ref<'create' | 'edit'>('create')
  const trackingCarrierMappingSubmitting = ref(false)
  const trackingCarrierMappingErrors = reactive<Record<string, string>>({})
  const trackingCarrierMappingForm = reactive(defaultTrackingCarrierMappingForm())

  function defaultTrackingCarrierMappingForm() {
    return defaultShippingTrackingCarrierMappingForm(
      trackingProviders?.value || [],
      carriers?.value || [],
      carrierServices?.value || []
    )
  }

  const clearTrackingProviderError = (field: string) => {
    delete trackingProviderErrors[field]
  }

  const clearTrackingCarrierMappingError = (field: string) => {
    delete trackingCarrierMappingErrors[field]
  }

  const showCreateTrackingProviderDialog = () => {
    trackingProviderDialogMode.value = 'create'
    resetReactive(trackingProviderForm, defaultShippingTrackingProviderForm())
    clearErrors(trackingProviderErrors)
    trackingProviderDialogOpen.value = true
  }

  const showEditTrackingProviderDialog = (provider: any) => {
    trackingProviderDialogMode.value = 'edit'
    resetReactive(trackingProviderForm, {
      ...defaultShippingTrackingProviderForm(),
      ...provider,
      webhook_enabled: provider.webhook_enabled === true,
      auto_register: provider.auto_register === true,
      polling_enabled: provider.polling_enabled === true,
      enabled: provider.enabled !== false,
      polling_interval_minutes: Number(provider.polling_interval_minutes || 60),
      request_timeout_seconds: Number(provider.request_timeout_seconds || 15),
      sort_order: Number(provider.sort_order || 0),
    })
    clearErrors(trackingProviderErrors)
    trackingProviderDialogOpen.value = true
  }

  const validateTrackingProvider = () => {
    clearErrors(trackingProviderErrors)
    if (!trackingProviderForm.provider_name?.trim()) trackingProviderErrors.provider_name = '请输入 Provider 名称'
    if (!trackingProviderForm.provider_code?.trim()) trackingProviderErrors.provider_code = '请输入 Provider 代码'
    if (!['production', 'sandbox'].includes(trackingProviderForm.environment)) trackingProviderErrors.environment = '请选择 Provider 环境'

    const numericFields = ['polling_interval_minutes', 'request_timeout_seconds', 'sort_order']
    const hasInvalidNumber = numericFields.some((field) => Number(trackingProviderForm[field] || 0) < 0)
    if (hasInvalidNumber) {
      toast.error('追踪配置的数字字段不能小于 0')
      return false
    }

    return Object.keys(trackingProviderErrors).length === 0
  }

  const buildTrackingProviderPayload = () => ({
    provider_code: trackingProviderForm.provider_code?.trim().toUpperCase() || '',
    provider_name: trackingProviderForm.provider_name?.trim() || '',
    environment: trackingProviderForm.environment || 'production',
    base_url: trackingProviderForm.base_url?.trim() || '',
    api_key: trackingProviderForm.api_key?.trim() || '',
    webhook_secret: trackingProviderForm.webhook_secret?.trim() || '',
    webhook_enabled: Boolean(trackingProviderForm.webhook_enabled),
    auto_register: Boolean(trackingProviderForm.auto_register),
    polling_enabled: Boolean(trackingProviderForm.polling_enabled),
    polling_interval_minutes: Number(trackingProviderForm.polling_interval_minutes || 60),
    request_timeout_seconds: Number(trackingProviderForm.request_timeout_seconds || 15),
    enabled: Boolean(trackingProviderForm.enabled),
    sort_order: Number(trackingProviderForm.sort_order || 0),
    description: trackingProviderForm.description || '',
  })

  const saveTrackingProvider = async () => {
    if (!validateTrackingProvider()) return

    trackingProviderSubmitting.value = true
    try {
      const payload = buildTrackingProviderPayload()
      if (trackingProviderDialogMode.value === 'create') {
        await shippingApi.createTrackingProvider(payload)
        toast.success('追踪配置已创建')
      } else {
        await shippingApi.updateTrackingProvider(trackingProviderForm.id, payload)
        toast.success('追踪配置已更新')
      }

      trackingProviderDialogOpen.value = false
      await Promise.all([fetchTrackingProviders(), fetchTrackingCarrierMappings()])
    } catch (error) {
      console.error('Failed to save tracking provider:', error)
    } finally {
      trackingProviderSubmitting.value = false
    }
  }

  const showCreateTrackingCarrierMappingDialog = () => {
    trackingCarrierMappingDialogMode.value = 'create'
    const form = defaultTrackingCarrierMappingForm()
    if (!form.carrier_id && form.carrier_service_id) {
      form.scope = 'carrier_service'
    }
    resetReactive(trackingCarrierMappingForm, form)
    clearErrors(trackingCarrierMappingErrors)
    trackingCarrierMappingDialogOpen.value = true
  }

  const showEditTrackingCarrierMappingDialog = (mapping: any) => {
    trackingCarrierMappingDialogMode.value = 'edit'
    resetReactive(trackingCarrierMappingForm, {
      ...defaultTrackingCarrierMappingForm(),
      ...mapping,
      provider_id: mapping.provider_id ? String(mapping.provider_id) : '',
      carrier_id: mapping.carrier_id ? String(mapping.carrier_id) : '',
      carrier_service_id: mapping.carrier_service_id ? String(mapping.carrier_service_id) : '',
      priority: Number(mapping.priority || 0),
      enabled: mapping.enabled !== false,
    })
    clearErrors(trackingCarrierMappingErrors)
    trackingCarrierMappingDialogOpen.value = true
  }

  const validateTrackingCarrierMapping = () => {
    clearErrors(trackingCarrierMappingErrors)
    if (!nullablePositiveID(trackingCarrierMappingForm.provider_id)) trackingCarrierMappingErrors.provider_id = '请选择追踪 Provider'
    if (!['carrier', 'carrier_service'].includes(trackingCarrierMappingForm.scope)) trackingCarrierMappingErrors.scope = '请选择映射层级'
    if (trackingCarrierMappingForm.scope === 'carrier' && !nullablePositiveID(trackingCarrierMappingForm.carrier_id)) {
      trackingCarrierMappingErrors.carrier_id = '请选择本地承运商'
    }
    if (trackingCarrierMappingForm.scope === 'carrier_service' && !nullablePositiveID(trackingCarrierMappingForm.carrier_service_id)) {
      trackingCarrierMappingErrors.carrier_service_id = '请选择本地线路服务'
    }
    if (!trackingCarrierMappingForm.provider_carrier_code?.trim()) {
      trackingCarrierMappingErrors.provider_carrier_code = '请输入 Provider Carrier Code'
    }
    if (Number(trackingCarrierMappingForm.priority || 0) < 0) {
      toast.error('承运商映射优先级不能小于 0')
      return false
    }
    return Object.keys(trackingCarrierMappingErrors).length === 0
  }

  const buildTrackingCarrierMappingPayload = () => ({
    provider_id: nullablePositiveID(trackingCarrierMappingForm.provider_id),
    scope: trackingCarrierMappingForm.scope || 'carrier',
    carrier_id: trackingCarrierMappingForm.scope === 'carrier' ? nullablePositiveID(trackingCarrierMappingForm.carrier_id) : null,
    carrier_service_id: trackingCarrierMappingForm.scope === 'carrier_service' ? nullablePositiveID(trackingCarrierMappingForm.carrier_service_id) : null,
    provider_carrier_code: trackingCarrierMappingForm.provider_carrier_code?.trim() || '',
    provider_carrier_name: trackingCarrierMappingForm.provider_carrier_name?.trim() || '',
    enabled: Boolean(trackingCarrierMappingForm.enabled),
    priority: Number(trackingCarrierMappingForm.priority || 0),
    description: trackingCarrierMappingForm.description || '',
  })

  const saveTrackingCarrierMapping = async () => {
    if (!validateTrackingCarrierMapping()) return

    trackingCarrierMappingSubmitting.value = true
    try {
      const payload = buildTrackingCarrierMappingPayload()
      if (trackingCarrierMappingDialogMode.value === 'create') {
        await shippingApi.createTrackingCarrierMapping(payload)
        toast.success('承运商映射已创建')
      } else {
        await shippingApi.updateTrackingCarrierMapping(trackingCarrierMappingForm.id, payload)
        toast.success('承运商映射已更新')
      }

      trackingCarrierMappingDialogOpen.value = false
      await fetchTrackingCarrierMappings()
    } catch (error) {
      console.error('Failed to save tracking carrier mapping:', error)
    } finally {
      trackingCarrierMappingSubmitting.value = false
    }
  }

  return {
    trackingProviderDialogOpen,
    trackingProviderDialogMode,
    trackingProviderSubmitting,
    trackingProviderErrors,
    trackingProviderForm,
    trackingCarrierMappingDialogOpen,
    trackingCarrierMappingDialogMode,
    trackingCarrierMappingSubmitting,
    trackingCarrierMappingErrors,
    trackingCarrierMappingForm,
    clearTrackingProviderError,
    clearTrackingCarrierMappingError,
    showCreateTrackingProviderDialog,
    showEditTrackingProviderDialog,
    saveTrackingProvider,
    showCreateTrackingCarrierMappingDialog,
    showEditTrackingCarrierMappingDialog,
    saveTrackingCarrierMapping,
  }
}

export default useShippingTrackingManager

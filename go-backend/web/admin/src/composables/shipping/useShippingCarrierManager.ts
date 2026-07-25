import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import {
  clearErrors,
  defaultShippingCarrierForm,
  defaultShippingCarrierServiceForm,
  resetReactive,
} from '@/lib/shippingForms'

const nullablePositiveID = (value: any) => {
  const numericValue = Number(value)
  return Number.isFinite(numericValue) && numericValue > 0 ? numericValue : null
}

export const useShippingCarrierManager = (options: Record<string, any> = {}) => {
  const carriers = options.carriers
  const fetchCarriers = options.fetchCarriers || (() => Promise.resolve())
  const fetchCarrierServices = options.fetchCarrierServices || (() => Promise.resolve())

  const carrierDialogOpen = ref(false)
  const carrierDialogMode = ref<'create' | 'edit'>('create')
  const carrierSubmitting = ref(false)
  const carrierErrors = reactive<Record<string, string>>({})
  const carrierForm = reactive(defaultShippingCarrierForm())

  const carrierServiceDialogOpen = ref(false)
  const carrierServiceDialogMode = ref<'create' | 'edit'>('create')
  const carrierServiceSubmitting = ref(false)
  const carrierServiceErrors = reactive<Record<string, string>>({})
  const carrierServiceForm = reactive(defaultShippingCarrierServiceForm(carriers?.value || []))

  const defaultCarrierServiceForm = () => defaultShippingCarrierServiceForm(carriers?.value || [])

  const clearCarrierError = (field: string) => {
    delete carrierErrors[field]
  }

  const clearCarrierServiceError = (field: string) => {
    delete carrierServiceErrors[field]
  }

  const showCreateCarrierDialog = () => {
    carrierDialogMode.value = 'create'
    resetReactive(carrierForm, defaultShippingCarrierForm())
    clearErrors(carrierErrors)
    carrierDialogOpen.value = true
  }

  const showEditCarrierDialog = (carrier: any) => {
    carrierDialogMode.value = 'edit'
    resetReactive(carrierForm, {
      ...defaultShippingCarrierForm(),
      ...carrier,
      enabled: carrier.enabled !== false,
      sort_order: Number(carrier.sort_order || 0),
    })
    clearErrors(carrierErrors)
    carrierDialogOpen.value = true
  }

  const validateCarrier = () => {
    clearErrors(carrierErrors)
    if (!carrierForm.name?.trim()) carrierErrors.name = '请输入承运商名称'
    if (!carrierForm.code?.trim()) carrierErrors.code = '请输入承运商代码'
    return Object.keys(carrierErrors).length === 0
  }

  const saveCarrier = async () => {
    if (!validateCarrier()) return

    carrierSubmitting.value = true
    try {
      const payload = {
        name: carrierForm.name.trim(),
        code: carrierForm.code.trim().toUpperCase(),
        tracking_url: carrierForm.tracking_url?.trim() || '',
        contact: carrierForm.contact?.trim() || '',
        phone: carrierForm.phone?.trim() || '',
        email: carrierForm.email?.trim() || '',
        service_area: carrierForm.service_area || '',
        enabled: Boolean(carrierForm.enabled),
        sort_order: Number(carrierForm.sort_order || 0),
      }

      if (carrierDialogMode.value === 'create') {
        await shippingApi.createCarrier(payload)
        toast.success('承运商已创建')
      } else {
        await shippingApi.updateCarrier(carrierForm.id, payload)
        toast.success('承运商已更新')
      }

      carrierDialogOpen.value = false
      await fetchCarriers()
    } catch (error) {
      console.error('Failed to save carrier:', error)
    } finally {
      carrierSubmitting.value = false
    }
  }

  const showCreateCarrierServiceDialog = () => {
    carrierServiceDialogMode.value = 'create'
    resetReactive(carrierServiceForm, defaultCarrierServiceForm())
    clearErrors(carrierServiceErrors)
    carrierServiceDialogOpen.value = true
  }

  const showEditCarrierServiceDialog = (service: any) => {
    carrierServiceDialogMode.value = 'edit'
    resetReactive(carrierServiceForm, {
      ...defaultCarrierServiceForm(),
      ...service,
      carrier_id: service.carrier_id ? String(service.carrier_id) : '',
      template_id: service.template_id ? String(service.template_id) : 'none',
      first_weight_grams: Number(service.first_weight_grams || 0),
      additional_weight_grams: Number(service.additional_weight_grams || 0),
      min_charge_weight_grams: Number(service.min_charge_weight_grams || 0),
      volumetric_divisor: Number(service.volumetric_divisor || 6000),
      fuel_surcharge_percent: Number(service.fuel_surcharge_percent || 0),
      remote_surcharge: Number(service.remote_surcharge || 0),
      eta_min_days: Number(service.eta_min_days || 0),
      eta_max_days: Number(service.eta_max_days || 0),
      enabled: service.enabled !== false,
      sort_order: Number(service.sort_order || 0),
    })
    clearErrors(carrierServiceErrors)
    carrierServiceDialogOpen.value = true
  }

  const validateCarrierService = () => {
    clearErrors(carrierServiceErrors)
    if (!nullablePositiveID(carrierServiceForm.carrier_id)) carrierServiceErrors.carrier_id = '请选择承运商'
    if (!carrierServiceForm.service_code?.trim()) carrierServiceErrors.service_code = '请输入线路代码'
    if (!carrierServiceForm.service_name?.trim()) carrierServiceErrors.service_name = '请输入线路名称'
    if (!['actual_weight', 'volumetric_weight', 'greater_of_actual_and_volumetric'].includes(carrierServiceForm.billing_mode)) {
      carrierServiceErrors.billing_mode = '请选择计费模式'
    }
    if (Number(carrierServiceForm.eta_max_days || 0) > 0 && Number(carrierServiceForm.eta_min_days || 0) > 0 && Number(carrierServiceForm.eta_max_days) < Number(carrierServiceForm.eta_min_days)) {
      carrierServiceErrors.eta_max_days = '最长时效不能小于最短时效'
    }

    const numericFields = [
      'first_weight_grams',
      'additional_weight_grams',
      'min_charge_weight_grams',
      'volumetric_divisor',
      'fuel_surcharge_percent',
      'remote_surcharge',
      'eta_min_days',
      'eta_max_days',
      'sort_order',
    ]
    const hasNegative = numericFields.some((field) => Number(carrierServiceForm[field] || 0) < 0)
    if (hasNegative) {
      toast.error('线路服务的数字字段不能小于 0')
      return false
    }

    return Object.keys(carrierServiceErrors).length === 0
  }

  const buildCarrierServicePayload = () => ({
    carrier_id: nullablePositiveID(carrierServiceForm.carrier_id),
    template_id: carrierServiceForm.template_id === 'none' ? null : nullablePositiveID(carrierServiceForm.template_id),
    service_code: carrierServiceForm.service_code?.trim().toUpperCase() || '',
    service_name: carrierServiceForm.service_name?.trim() || '',
    route_name: carrierServiceForm.route_name?.trim() || '',
    countries: carrierServiceForm.countries?.trim() || '[]',
    currency: carrierServiceForm.currency?.trim().toUpperCase() || 'USD',
    billing_mode: carrierServiceForm.billing_mode || 'actual_weight',
    first_weight_grams: Number(carrierServiceForm.first_weight_grams || 0),
    additional_weight_grams: Number(carrierServiceForm.additional_weight_grams || 0),
    min_charge_weight_grams: Number(carrierServiceForm.min_charge_weight_grams || 0),
    volumetric_divisor: Number(carrierServiceForm.volumetric_divisor || 6000),
    fuel_surcharge_percent: Number(carrierServiceForm.fuel_surcharge_percent || 0),
    remote_surcharge: Number(carrierServiceForm.remote_surcharge || 0),
    eta_min_days: Number(carrierServiceForm.eta_min_days || 0),
    eta_max_days: Number(carrierServiceForm.eta_max_days || 0),
    enabled: Boolean(carrierServiceForm.enabled),
    sort_order: Number(carrierServiceForm.sort_order || 0),
    description: carrierServiceForm.description || '',
  })

  const saveCarrierService = async () => {
    if (!validateCarrierService()) return

    carrierServiceSubmitting.value = true
    try {
      const payload = buildCarrierServicePayload()
      if (carrierServiceDialogMode.value === 'create') {
        await shippingApi.createCarrierService(payload)
        toast.success('线路服务已创建')
      } else {
        await shippingApi.updateCarrierService(carrierServiceForm.id, payload)
        toast.success('线路服务已更新')
      }

      carrierServiceDialogOpen.value = false
      await fetchCarrierServices()
    } catch (error) {
      console.error('Failed to save carrier service:', error)
    } finally {
      carrierServiceSubmitting.value = false
    }
  }

  return {
    carrierDialogOpen,
    carrierDialogMode,
    carrierSubmitting,
    carrierErrors,
    carrierForm,
    carrierServiceDialogOpen,
    carrierServiceDialogMode,
    carrierServiceSubmitting,
    carrierServiceErrors,
    carrierServiceForm,
    clearCarrierError,
    clearCarrierServiceError,
    showCreateCarrierDialog,
    showEditCarrierDialog,
    saveCarrier,
    showCreateCarrierServiceDialog,
    showEditCarrierServiceDialog,
    saveCarrierService,
  }
}

export default useShippingCarrierManager

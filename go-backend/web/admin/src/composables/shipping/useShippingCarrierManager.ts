import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import {
  clearErrors,
  defaultShippingCarrierForm,
  resetReactive,
} from '@/lib/shippingForms'

export const useShippingCarrierManager = (options: Record<string, any> = {}) => {
  const fetchCarriers = options.fetchCarriers || (() => Promise.resolve())

  const carrierDialogOpen = ref(false)
  const carrierDialogMode = ref<'create' | 'edit'>('create')
  const carrierSubmitting = ref(false)
  const carrierErrors = reactive<Record<string, string>>({})
  const carrierForm = reactive(defaultShippingCarrierForm())

  const clearCarrierError = (field: string) => {
    delete carrierErrors[field]
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


  return {
    carrierDialogOpen,
    carrierDialogMode,
    carrierSubmitting,
    carrierErrors,
    carrierForm,
    clearCarrierError,
    showCreateCarrierDialog,
    showEditCarrierDialog,
    saveCarrier,
  }
}

export default useShippingCarrierManager

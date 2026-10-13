import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import type { ShippingRouteAggregateForm } from '@/modules/shipping/shippingTypes'
import {
  clearErrors,
  defaultShippingRouteAggregateForm,
  resetReactive,
} from '@/lib/shippingForms'

const normalizeIDs = (value: unknown) => (
  Array.isArray(value)
    ? Array.from(new Set(value.map((item) => String(item || '').trim()).filter(Boolean)))
    : []
)

export const useShippingRouteAggregateManager = (options: Record<string, any> = {}) => {
  const fetchRouteAggregates = options.fetchRouteAggregates || (() => Promise.resolve())

  const routeAggregateDialogOpen = ref(false)
  const routeAggregateDialogMode = ref<'create' | 'edit'>('create')
  const routeAggregateSubmitting = ref(false)
  const routeAggregateErrors = reactive<Record<string, string>>({})
  const routeAggregateForm = reactive<ShippingRouteAggregateForm>(defaultShippingRouteAggregateForm())

  const clearRouteAggregateError = (field: string) => {
    delete routeAggregateErrors[field]
  }

  const showCreateRouteAggregateDialog = () => {
    routeAggregateDialogMode.value = 'create'
    resetReactive(routeAggregateForm, defaultShippingRouteAggregateForm())
    clearErrors(routeAggregateErrors)
    routeAggregateDialogOpen.value = true
  }

  const showEditRouteAggregateDialog = (aggregate: any) => {
    routeAggregateDialogMode.value = 'edit'
    resetReactive(routeAggregateForm, {
      ...defaultShippingRouteAggregateForm(),
      ...aggregate,
      countries: String(aggregate?.countries || '[]'),
      published_carrier_services: Array.isArray(aggregate?.carrier_services) ? aggregate.carrier_services.filter((item: any) => item.fpx_channel_id || item.yanwen_published_channel_id) : [],
      carrier_service_ids: Array.isArray(aggregate?.carrier_services)
        ? aggregate.carrier_services.filter((item: any) => !item.fpx_channel_id && !item.yanwen_published_channel_id).map((service: any) => String(service.id))
        : normalizeIDs(aggregate?.carrier_service_ids),
      enabled: aggregate?.enabled !== false,
      sort_order: Number(aggregate?.sort_order || 0),
    })
    clearErrors(routeAggregateErrors)
    routeAggregateDialogOpen.value = true
  }

  const validateRouteAggregate = () => {
    clearErrors(routeAggregateErrors)
    if (!routeAggregateForm.name?.trim()) routeAggregateErrors.name = '请输入聚合线路名称'
    routeAggregateForm.carrier_service_ids = normalizeIDs(routeAggregateForm.carrier_service_ids)
    if (!routeAggregateForm.published_carrier_services.length && !routeAggregateForm.carrier_service_ids.length) routeAggregateErrors.published_carrier_services = '请至少选择一条已发布线路'
    return Object.keys(routeAggregateErrors).length === 0
  }

  const saveRouteAggregate = async () => {
    if (!validateRouteAggregate()) return
    routeAggregateSubmitting.value = true
    try {
      const payload = {
        name: routeAggregateForm.name.trim(),
        description: routeAggregateForm.description?.trim() || '',
        enabled: Boolean(routeAggregateForm.enabled),
        sort_order: Number(routeAggregateForm.sort_order || 0),
        carrier_service_ids: routeAggregateForm.carrier_service_ids.map((id) => Number(id)),
        published_carrier_services: routeAggregateForm.published_carrier_services.map((service: any) => ({
          carrier_id: Number(service.carrier_id), fpx_channel_id: service.fpx_channel_id || null,
          yanwen_published_channel_id: service.yanwen_published_channel_id || null, provider_code: service.provider_code,
        })),
      }
      if (routeAggregateDialogMode.value === 'create') {
        await shippingApi.createRouteAggregate(payload)
        toast.success('聚合线路已创建')
      } else {
        await shippingApi.updateRouteAggregate(routeAggregateForm.id, payload)
        toast.success('聚合线路已更新')
      }
      routeAggregateDialogOpen.value = false
      await fetchRouteAggregates()
    } catch (error) {
      console.error('Failed to save shipping route aggregate:', error)
    } finally {
      routeAggregateSubmitting.value = false
    }
  }

  return {
    routeAggregateDialogOpen,
    routeAggregateDialogMode,
    routeAggregateSubmitting,
    routeAggregateErrors,
    routeAggregateForm,
    clearRouteAggregateError,
    showCreateRouteAggregateDialog,
    showEditRouteAggregateDialog,
    saveRouteAggregate,
  }
}

export default useShippingRouteAggregateManager

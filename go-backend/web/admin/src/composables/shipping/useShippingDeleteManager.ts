import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'

const titleByType: Record<string, string> = {
  template: '删除运费模板？',
  zone: '删除配送区域？',
  carrier: '删除承运商？',
  carrierService: '删除线路服务？',
  trackingProvider: '删除追踪配置？',
  trackingCarrierMapping: '删除承运商映射？',
  packaging: '删除包装规则？',
}

const nameOf = (target: any) => (
  target?.name ||
  target?.provider_name ||
  target?.provider_carrier_code ||
  target?.service_name ||
  target?.rule_name ||
  '当前记录'
)

export const useShippingDeleteManager = (refreshers: Record<string, any>) => {
  const deleteDialogOpen = ref(false)
  const deleteTarget = ref<any>(null)
  const deleteType = ref('')
  const deleteLoading = ref(false)

  const deleteDialogTitle = computed(() => titleByType[deleteType.value] || '确认删除？')

  const deleteDialogDescription = computed(() => {
    if (deleteLoading.value) return '正在删除，请稍候。'
    return `将删除「${nameOf(deleteTarget.value)}」。这个操作不可撤销，请确认没有正在使用它的运费规则或订单流程。`
  })

  const requestDelete = (type: string, target: any) => {
    deleteType.value = type
    deleteTarget.value = target
    deleteDialogOpen.value = true
  }

  const confirmDelete = async () => {
    if (!deleteTarget.value || deleteLoading.value) return

    deleteLoading.value = true
    try {
      if (deleteType.value === 'template') {
        await shippingApi.deleteTemplate(deleteTarget.value.id)
        toast.success('运费模板已删除')
        await refreshers.fetchTemplates?.()
      } else if (deleteType.value === 'zone') {
        await shippingApi.deleteZone(deleteTarget.value.id)
        toast.success('配送区域已删除')
        await refreshers.fetchZones?.()
      } else if (deleteType.value === 'carrier') {
        await shippingApi.deleteCarrier(deleteTarget.value.id)
        toast.success('承运商已删除')
        await refreshers.fetchCarriers?.()
      } else if (deleteType.value === 'carrierService') {
        await shippingApi.deleteCarrierService(deleteTarget.value.id)
        toast.success('线路服务已删除')
        await refreshers.fetchCarrierServices?.()
      } else if (deleteType.value === 'trackingProvider') {
        await shippingApi.deleteTrackingProvider(deleteTarget.value.id)
        toast.success('追踪配置已删除')
        await Promise.all([
          refreshers.fetchTrackingProviders?.(),
          refreshers.fetchTrackingCarrierMappings?.(),
        ])
      } else if (deleteType.value === 'trackingCarrierMapping') {
        await shippingApi.deleteTrackingCarrierMapping(deleteTarget.value.id)
        toast.success('承运商映射已删除')
        await refreshers.fetchTrackingCarrierMappings?.()
      } else if (deleteType.value === 'packaging') {
        await shippingApi.deletePackagingRule(deleteTarget.value.id)
        toast.success('包装规则已删除')
        await refreshers.fetchPackagingRules?.()
      }
      deleteDialogOpen.value = false
    } catch (error) {
      console.error('Failed to delete shipping record:', error)
    } finally {
      deleteLoading.value = false
    }
  }

  return {
    deleteDialogOpen,
    deleteTarget,
    deleteType,
    deleteLoading,
    deleteDialogTitle,
    deleteDialogDescription,
    requestDelete,
    confirmDelete,
  }
}

export default useShippingDeleteManager

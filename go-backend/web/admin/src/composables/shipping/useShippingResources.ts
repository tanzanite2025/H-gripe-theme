import { reactive, ref } from 'vue'
import shippingApi from '@/api/shipping'

export const useShippingResources = () => {
  const templates = ref<any[]>([])
  const zones = ref<any[]>([])
  const templateBindings = ref<any[]>([])
  const carriers = ref<any[]>([])
  const carrierServices = ref<any[]>([])
  const trackingProviders = ref<any[]>([])
  const trackingCarrierMappings = ref<any[]>([])
  const trackingShipmentsCount = ref(0)
  const packagingRules = ref<any[]>([])
  const refreshing = ref(false)
  const loading = reactive({
    templates: false,
    zones: false,
    bindings: false,
    carriers: false,
    services: false,
    tracking: false,
    trackingMappings: false,
    packaging: false,
  })

  const handleTrackingShipmentsCountChange = (count: number) => {
    trackingShipmentsCount.value = Number(count || 0)
  }

  const fetchTemplates = async () => {
    loading.templates = true
    try {
      templates.value = await shippingApi.listTemplates()
    } catch (error) {
      console.error('Failed to fetch shipping templates:', error)
    } finally {
      loading.templates = false
    }
  }

  const fetchZones = async () => {
    loading.zones = true
    try {
      zones.value = await shippingApi.listZones()
    } catch (error) {
      console.error('Failed to fetch shipping zones:', error)
    } finally {
      loading.zones = false
    }
  }

  const fetchTemplateBindings = async () => {
    loading.bindings = true
    try {
      templateBindings.value = await shippingApi.listTemplateBindings()
    } catch (error) {
      console.error('Failed to fetch shipping template bindings:', error)
    } finally {
      loading.bindings = false
    }
  }

  const fetchCarriers = async () => {
    loading.carriers = true
    try {
      carriers.value = await shippingApi.listCarriers()
    } catch (error) {
      console.error('Failed to fetch carriers:', error)
    } finally {
      loading.carriers = false
    }
  }

  const fetchCarrierServices = async () => {
    loading.services = true
    try {
      carrierServices.value = await shippingApi.listCarrierServices()
    } catch (error) {
      console.error('Failed to fetch carrier services:', error)
    } finally {
      loading.services = false
    }
  }

  const fetchTrackingProviders = async () => {
    loading.tracking = true
    try {
      trackingProviders.value = await shippingApi.listTrackingProviders()
    } catch (error) {
      console.error('Failed to fetch tracking providers:', error)
    } finally {
      loading.tracking = false
    }
  }

  const fetchTrackingCarrierMappings = async () => {
    loading.trackingMappings = true
    try {
      trackingCarrierMappings.value = await shippingApi.listTrackingCarrierMappings()
    } catch (error) {
      console.error('Failed to fetch tracking carrier mappings:', error)
    } finally {
      loading.trackingMappings = false
    }
  }

  const fetchPackagingRules = async () => {
    loading.packaging = true
    try {
      packagingRules.value = await shippingApi.listPackagingRules()
    } catch (error) {
      console.error('Failed to fetch packaging rules:', error)
    } finally {
      loading.packaging = false
    }
  }

  const refreshCurrentTab = async (activeTab: string, trackingShipmentsPanelRef: any) => {
    refreshing.value = true
    try {
      if (activeTab === 'templates') {
        await fetchTemplates()
      } else if (activeTab === 'zones') {
        await fetchZones()
      } else if (activeTab === 'bindings') {
        await Promise.all([fetchTemplateBindings(), fetchTemplates()])
      } else if (activeTab === 'carriers') {
        await fetchCarriers()
      } else if (activeTab === 'services') {
        await Promise.all([fetchCarrierServices(), fetchCarriers(), fetchTemplates()])
      } else if (activeTab === 'tracking') {
        await Promise.all([fetchTrackingProviders(), fetchTrackingCarrierMappings(), fetchCarriers(), fetchCarrierServices()])
      } else if (activeTab === 'trackingShipments') {
        await Promise.all([
          trackingShipmentsPanelRef.value?.refresh?.(),
          fetchTrackingProviders(),
          fetchCarriers(),
          fetchCarrierServices(),
        ])
      } else if (activeTab === 'packaging') {
        await fetchPackagingRules()
      } else {
        await Promise.all([
          fetchTemplates(),
          fetchZones(),
          fetchTemplateBindings(),
          fetchCarriers(),
          fetchCarrierServices(),
          fetchTrackingProviders(),
          fetchTrackingCarrierMappings(),
          fetchPackagingRules(),
        ])
      }
    } finally {
      refreshing.value = false
    }
  }

  const fetchAllShippingResources = () => Promise.all([
    fetchTemplates(),
    fetchZones(),
    fetchTemplateBindings(),
    fetchCarriers(),
    fetchCarrierServices(),
    fetchTrackingProviders(),
    fetchTrackingCarrierMappings(),
    fetchPackagingRules(),
  ])

  return {
    templates,
    zones,
    templateBindings,
    carriers,
    carrierServices,
    trackingProviders,
    trackingCarrierMappings,
    trackingShipmentsCount,
    packagingRules,
    refreshing,
    loading,
    handleTrackingShipmentsCountChange,
    fetchTemplates,
    fetchZones,
    fetchTemplateBindings,
    fetchCarriers,
    fetchCarrierServices,
    fetchTrackingProviders,
    fetchTrackingCarrierMappings,
    fetchPackagingRules,
    refreshCurrentTab,
    fetchAllShippingResources,
  }
}

export default useShippingResources
